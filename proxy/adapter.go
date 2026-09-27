// Package proxy translates external native servers to historical clients. Each
// Adapter belongs to one connection and is driven by one session event loop.
package proxy

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/itemconv"
	"github.com/shawtymarco/go-multiversion/mapping"
	v475 "github.com/shawtymarco/go-multiversion/protocols/v1_18_0"
	v486 "github.com/shawtymarco/go-multiversion/protocols/v1_18_10"
)

type Direction uint8

const (
	ClientToServer Direction = iota
	ServerToClient
)

// UpstreamSnapshot retains the original bootstrap packets, not a lossy GameData.
type UpstreamSnapshot struct {
	StartGame    *packet.StartGame
	ItemRegistry *packet.ItemRegistry
	Dimensions   []protocol.DimensionDefinition
}
type semanticProtocol interface {
	minecraft.Protocol
	ProxyWireProtocol() minecraft.Protocol
	ProxyToNative(packet.Packet) []packet.Packet
	ProxyFromNative(packet.Packet) []packet.Packet
	ProxyMappings() (*mapping.BlockMapper, *mapping.ItemMapper)
}
type protocolHolder struct{ minecraft.Protocol }

type Adapter struct {
	version     string
	id          int32
	codec       atomic.Pointer[protocolHolder]
	semantic    semanticProtocol
	registry    *NetworkRegistry
	blocks      *mapping.BlockMapper
	items       *mapping.ItemMapper
	dimensions  map[int32]int
	bound       bool
	biomesSent  bool
	lastTick    uint64
	jump, sneak bool
	started     time.Time
	serverTick  int64
	replies     []packet.Packet
	inventories map[uint32]*packet.InventoryContent
	pending     map[int32]struct{}
	dropped     map[uint32]uint64
}

func NewAdapter(version string) (*Adapter, error) {
	a := &Adapter{version: version, dimensions: map[int32]int{0: 24, 1: 8, 2: 16}, started: time.Now(), inventories: map[uint32]*packet.InventoryContent{}, pending: map[int32]struct{}{}, dropped: map[uint32]uint64{}}
	switch version {
	case "1.18.0", "1.18.1", "1.18.2":
		a.id = 475
		a.semantic = v475.New().(semanticProtocol)
	case "1.18.10", "1.18.11", "1.18.12":
		a.id = 486
		a.semantic = v486.New().(semanticProtocol)
	default:
		return nil, fmt.Errorf("unsupported client version %q", version)
	}
	a.codec.Store(&protocolHolder{a.semantic.ProxyWireProtocol()})
	return a, nil
}

func (a *Adapter) WireProtocol() minecraft.Protocol { return &wireProtocol{adapter: a} }
func (a *Adapter) Bound() bool                      { return a.bound }
func (a *Adapter) ClientReplies() []packet.Packet   { out := a.replies; a.replies = nil; return out }

// SelectClientVersion refines the advertised family after validated Login.
func (a *Adapter) SelectClientVersion(version string) error {
	if a.bound {
		return fmt.Errorf("client version is already bound")
	}
	selected, err := NewAdapter(version)
	if err != nil || selected.id != a.id {
		return fmt.Errorf("client version does not match listener family")
	}
	a.version = version
	return nil
}
func (a *Adapter) DropReport() map[uint32]uint64 {
	out := make(map[uint32]uint64, len(a.dropped))
	for id, n := range a.dropped {
		out[id] = n
	}
	return out
}

func (a *Adapter) BindUpstream(snapshot UpstreamSnapshot) error {
	if a.bound {
		return fmt.Errorf("upstream registry already bound; reconnect requires a new adapter")
	}
	if snapshot.StartGame == nil || snapshot.ItemRegistry == nil || len(snapshot.ItemRegistry.Items) == 0 {
		return fmt.Errorf("incomplete upstream bootstrap")
	}
	start := clonePacket(snapshot.StartGame).(*packet.StartGame)
	items := clonePacket(snapshot.ItemRegistry).(*packet.ItemRegistry)
	registry, err := NewNetworkRegistry(start.UseBlockNetworkIDHashes, start.Blocks)
	if err != nil {
		return err
	}
	var p minecraft.Protocol
	if a.id == 475 {
		p, err = v475.NewWithRegistries(registry, items.Items)
	} else {
		p, err = v486.NewWithRegistries(registry, items.Items)
	}
	if err != nil {
		return fmt.Errorf("bind server registry: %w", err)
	}
	a.registry = registry
	a.semantic = p.(semanticProtocol)
	a.blocks, a.items = a.semantic.ProxyMappings()
	for _, d := range snapshot.Dimensions {
		var id int32
		switch strings.TrimPrefix(d.Name, "minecraft:") {
		case "overworld":
			id = 0
		case "nether":
			id = 1
		case "the_end", "end":
			id = 2
		default:
			return fmt.Errorf("custom dimension %q is not supported by 1.18", d.Name)
		}
		if d.HeightRange <= 0 || d.HeightRange%16 != 0 || d.HeightRange > 1024 {
			return fmt.Errorf("invalid dimension range")
		}
		a.dimensions[id] = int(d.HeightRange / 16)
	}
	a.serverTick = start.Time
	a.started = time.Now()
	a.bound = true
	a.codec.Store(&protocolHolder{a.semantic.ProxyWireProtocol()})
	return nil
}

func (a *Adapter) Translate(direction Direction, original packet.Packet) (out []packet.Packet, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out = nil
			err = fmt.Errorf("translate %T: %v", original, recovered)
		}
	}()
	if direction != ClientToServer && direction != ServerToClient {
		return nil, fmt.Errorf("invalid packet direction")
	}
	if original == nil {
		return nil, fmt.Errorf("nil packet")
	}
	if unknown, ok := original.(*packet.Unknown); ok {
		return nil, fmt.Errorf("unsupported packet ID %d in direction %d", unknown.PacketID, direction)
	}
	pk := clonePacket(original)
	if direction == ClientToServer {
		if clock, ok := pk.(interface{ ProxyTickReply(int64) packet.Packet }); ok {
			a.replies = append(a.replies, clock.ProxyTickReply(a.serverTick+int64(time.Since(a.started)/(50*time.Millisecond))))
			return nil, nil
		}
		if _, ok := pk.(*packet.ClientCacheStatus); ok {
			return []packet.Packet{&packet.ClientCacheStatus{Enabled: false}}, nil
		}
	}
	if !a.bound {
		switch current := pk.(type) {
		case *packet.ResourcePackStack:
			current.BaseGameVersion = a.version
		case *packet.PlayStatus, *packet.ResourcePacksInfo, *packet.ResourcePackClientResponse, *packet.ResourcePackDataInfo, *packet.ResourcePackChunkData, *packet.ResourcePackChunkRequest, *packet.Disconnect:
		default:
			return nil, fmt.Errorf("%T arrived before server registries were bound", pk)
		}
		return []packet.Packet{pk}, nil
	}
	if direction == ServerToClient {
		switch current := pk.(type) {
		case *packet.InventoryContent:
			a.inventories[current.WindowID] = clonePacket(current).(*packet.InventoryContent)
		case *packet.InventorySlot:
			if inv := a.inventories[current.WindowID]; inv != nil && int(current.Slot) < len(inv.Content) {
				inv.Content[current.Slot] = current.NewItem
			}
		case *packet.ItemStackResponse:
			for _, response := range current.Responses {
				delete(a.pending, response.RequestID)
			}
		case *packet.ItemRegistry:
			if err := a.items.ValidateNativeEntries(current.Items); err != nil {
				return nil, fmt.Errorf("server changed its item registry without a new session: %w", err)
			}
		case *packet.LevelChunk:
			if current.CacheEnabled {
				return nil, fmt.Errorf("server sent cached chunk despite cache-off negotiation")
			}
			count := int(current.SubChunkCount)
			if _, requestMode := current.SubChunkLimit.Value(); requestMode {
				count = 0
			}
			current.RawPayload, err = rewriteLevelPayload(current.RawPayload, count, a.dimensions[current.Dimension], a.mapBlock, a.mapBiome)
			if err != nil {
				return nil, fmt.Errorf("chunk %v: %w", current.Position, err)
			}
		case *packet.SubChunk:
			if current.CacheEnabled {
				return nil, fmt.Errorf("server sent cached sub-chunk despite cache-off negotiation")
			}
			for i := range current.SubChunkEntries {
				e := &current.SubChunkEntries[i]
				if e.Result == protocol.SubChunkResultSuccessAllAir && a.id == 475 {
					e.Result = protocol.SubChunkResultSuccess
					e.RawPayload = protocol.Option([]byte{9, 0, byte(current.Position[1] + int32(e.Offset[1]))})
				} else if payload, ok := e.RawPayload.Value(); ok && len(payload) > 0 && e.Result == protocol.SubChunkResultSuccess {
					mapped, err := rewriteSubPayload(payload, a.mapBlock)
					if err != nil {
						return nil, err
					}
					e.RawPayload = protocol.Option(mapped)
				}
			}
		case *packet.StartGame:
			current.UseBlockNetworkIDHashes = false
			current.GameVersion = a.version
			current.BaseGameVersion = a.version
		case *packet.BiomeDefinitionList:
			a.biomesSent = true
		}
	}
	// Bootstrap registries and creative/recipe catalogues intentionally include
	// unsupported entries which the historical adapter filters. Live inventory
	// contents must never disappear silently.
	skipItems := false
	switch pk.(type) {
	case *packet.ItemRegistry, *packet.CreativeContent, *packet.CraftingData:
		skipItems = true
	}
	if err := a.walkRegistryFields(pk, direction, skipItems); err != nil {
		if direction == ClientToServer && a.rejectInventory(pk) {
			return nil, nil
		}
		return nil, err
	}
	if direction == ClientToServer {
		if input, ok := pk.(*packet.PlayerAuthInput); ok {
			a.completeInput(input)
		}
		if requests, ok := pk.(*packet.ItemStackRequest); ok {
			for _, request := range requests.Requests {
				if len(a.pending) >= 256 {
					a.rejectInventory(pk)
					return nil, nil
				}
				a.pending[request.RequestID] = struct{}{}
			}
		}
		out = a.semantic.ProxyToNative(pk)
		if len(out) == 0 {
			if a.rejectInventory(pk) {
				return nil, nil
			}
			return nil, fmt.Errorf("unrepresentable serverbound packet %T", pk)
		}
		for _, p := range out {
			if err := a.restoreNetworkFields(p); err != nil {
				return nil, err
			}
		}
		return out, nil
	}
	out = a.semantic.ProxyFromNative(pk)
	filtered := out[:0]
	for _, candidate := range out {
		switch p := candidate.(type) {
		case *packet.StartGame:
			p.GameVersion = a.version
			p.BaseGameVersion = a.version
		case *packet.ResourcePackStack:
			p.BaseGameVersion = a.version
		}
		if len(a.codec.Load().ConvertFromLatest(clonePacket(candidate), nil)) == 0 {
			a.dropped[candidate.ID()]++
			continue
		}
		filtered = append(filtered, candidate)
	}
	out = filtered
	if status, ok := pk.(*packet.PlayStatus); ok && status.Status == packet.PlayStatusPlayerSpawn && !a.biomesSent {
		out = append([]packet.Packet{&packet.BiomeDefinitionList{}}, out...)
		a.biomesSent = true
	}
	return out, nil
}

func (a *Adapter) mapBlock(network uint32) (uint32, error) {
	id, ok := a.registry.Ordinal(network)
	if !ok {
		return 0, fmt.Errorf("unknown upstream block network ID %d", network)
	}
	target, valid, exact := a.blocks.MapNative(id)
	if !valid || !exact {
		name, _, _ := a.registry.RuntimeIDToState(id)
		return 0, fmt.Errorf("block %s (%d) has no verified 1.18 replacement", name, network)
	}
	return target, nil
}
func (a *Adapter) mapBiome(id uint32) (uint32, error) {
	mapped, ok := mapping.Pre12650Biome(id)
	if !ok {
		return 0, fmt.Errorf("unmapped biome %d", id)
	}
	return mapped, nil
}

func (a *Adapter) walkRegistryFields(pk packet.Packet, d Direction, skipItems bool) error {
	return visitPacket(reflect.ValueOf(pk), func(v reflect.Value, name string) error {
		if v.CanInterface() && !skipItems {
			if stack, ok := v.Interface().(protocol.ItemStack); ok && stack.NetworkID != 0 {
				valid := false
				if d == ServerToClient {
					_, valid = a.items.NativeToTarget(stack.NetworkID)
					if !valid {
						_, valid = itemconv.DowngradeLegacySpawnEgg(stack, a.items)
					}
				} else {
					_, valid = a.items.TargetToNative(stack.NetworkID)
					if !valid {
						_, valid = itemconv.UpgradeLegacySpawnEgg(stack, a.items)
					}
				}
				if !valid {
					return fmt.Errorf("unmapped item network ID %d in %T", stack.NetworkID, pk)
				}
			}
		}
		if d == ServerToClient && v.CanSet() && (name == "BlockRuntimeID" || name == "NewBlockRuntimeID") {
			var network uint32
			switch v.Kind() {
			case reflect.Uint32:
				network = uint32(v.Uint())
			case reflect.Int32:
				network = uint32(v.Int())
			default:
				return nil
			}
			id, ok := a.registry.Ordinal(network)
			if !ok {
				return fmt.Errorf("unknown block network ID %d", network)
			}
			if _, err := a.mapBlock(network); err != nil {
				return err
			}
			if v.Kind() == reflect.Int32 {
				v.SetInt(int64(id))
			} else {
				v.SetUint(uint64(id))
			}
		}
		return nil
	}, "")
}
func (a *Adapter) restoreNetworkFields(pk packet.Packet) error {
	return visitPacket(reflect.ValueOf(pk), func(v reflect.Value, name string) error {
		if v.CanSet() && (name == "BlockRuntimeID" || name == "NewBlockRuntimeID") {
			var id uint32
			switch v.Kind() {
			case reflect.Uint32:
				id = uint32(v.Uint())
			case reflect.Int32:
				id = uint32(v.Int())
			default:
				return nil
			}
			network, ok := a.registry.Network(id)
			if !ok {
				return fmt.Errorf("invalid native block ordinal %d", id)
			}
			if v.Kind() == reflect.Int32 {
				v.SetInt(int64(int32(network)))
			} else {
				v.SetUint(uint64(network))
			}
		}
		return nil
	}, "")
}

func (a *Adapter) completeInput(pk *packet.PlayerAuthInput) {
	pk.InteractionModel = packet.InteractionModelCrosshair
	if pk.InputMode == packet.InputModeTouch {
		pk.InteractionModel = packet.InteractionModelTouch
	}
	pk.InteractPitch, pk.InteractYaw = pk.Pitch, pk.Yaw
	pk.RawMoveVector = pk.MoveVector
	pk.AnalogueMoveVector = pk.MoveVector
	pitch, yaw := float64(pk.Pitch)*math.Pi/180, float64(pk.Yaw)*math.Pi/180
	pk.CameraOrientation[0] = float32(-math.Sin(yaw) * math.Cos(pitch))
	pk.CameraOrientation[1] = float32(-math.Sin(pitch))
	pk.CameraOrientation[2] = float32(math.Cos(yaw) * math.Cos(pitch))
	jump, sneak := pk.InputData.Load(packet.InputFlagJumpDown), pk.InputData.Load(packet.InputFlagSneaking)
	for _, entry := range []struct {
		flag int
		set  bool
	}{{packet.InputFlagJumpCurrentRaw, jump}, {packet.InputFlagJumpPressedRaw, jump && !a.jump}, {packet.InputFlagJumpReleasedRaw, !jump && a.jump}, {packet.InputFlagSneakCurrentRaw, sneak}, {packet.InputFlagSneakPressedRaw, sneak && !a.sneak}, {packet.InputFlagSneakReleasedRaw, !sneak && a.sneak}} {
		if entry.set {
			pk.InputData.Set(entry.flag)
		} else {
			pk.InputData.Unset(entry.flag)
		}
	}
	a.lastTick = pk.Tick
	a.jump, a.sneak = jump, sneak
}

func (a *Adapter) rejectInventory(pk packet.Packet) bool {
	request, ok := pk.(*packet.ItemStackRequest)
	if !ok {
		return false
	}
	response := &packet.ItemStackResponse{}
	for _, r := range request.Requests {
		delete(a.pending, r.RequestID)
		response.Responses = append(response.Responses, protocol.ItemStackResponse{Status: protocol.ItemStackResponseStatusError, RequestID: r.RequestID})
	}
	a.replies = append(a.replies, response)
	for _, inventory := range a.inventories {
		a.replies = append(a.replies, clonePacket(inventory))
	}
	return true
}

type wireProtocol struct{ adapter *Adapter }

func (p *wireProtocol) ID() int32                                 { return p.adapter.id }
func (p *wireProtocol) Ver() string                               { return p.adapter.version }
func (p *wireProtocol) LegacyNetworkSettings() packet.Compression { return packet.FlateCompression }
func (p *wireProtocol) current() minecraft.Protocol               { return p.adapter.codec.Load().Protocol }
func (p *wireProtocol) Packets(listener bool) packet.Pool         { return p.current().Packets(listener) }
func (p *wireProtocol) NewReader(r minecraft.ByteReader, shield int32, limits bool) protocol.IO {
	return p.current().NewReader(r, shield, limits)
}
func (p *wireProtocol) NewWriter(w minecraft.ByteWriter, shield int32) protocol.IO {
	return p.current().NewWriter(w, shield)
}
func (p *wireProtocol) ConvertToLatest(pk packet.Packet, c *minecraft.Conn) []packet.Packet {
	return p.current().ConvertToLatest(pk, c)
}
func (p *wireProtocol) ConvertFromLatest(pk packet.Packet, c *minecraft.Conn) []packet.Packet {
	return p.current().ConvertFromLatest(clonePacket(pk), c)
}
