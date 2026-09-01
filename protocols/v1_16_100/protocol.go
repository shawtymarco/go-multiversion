// Package v1_16_100 implements Minecraft Bedrock protocol 419 for Minecraft 1.16.100 against the current native model.
package v1_16_100

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
	v975 "github.com/shawtymarco/go-multiversion/protocols/v1_26_20"
)

const (
	ID      int32 = 419
	Version       = "1.16.100"
)

type Protocol struct {
	base    minecraft.Protocol
	runtime *runtimeData
}

type MappingReport struct {
	NativeBlocks, TargetBlocks int
	NativeItems, TargetItems   int
	BlockFallbacks             []mapping.BlockFallback
	ItemFallbacks              []mapping.ItemFallback
	TargetItemFallbacks        []mapping.TargetItemFallback
	CreativeItems              int
}

type fallbackProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type fallbackBlock struct {
	NativeRuntimeID uint32             `json:"native_runtime_id"`
	Name            string             `json:"name"`
	Properties      []fallbackProperty `json:"properties"`
	TargetRuntimeID uint32             `json:"target_runtime_id"`
}

func New() minecraft.Protocol { return &Protocol{base: v975.New()} }

func NewWithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return NewWithRegistries(native, nil)
}

func NewWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (minecraft.Protocol, error) {
	runtime, err := newRuntimeData(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return &Protocol{base: v975.New(), runtime: runtime}, nil
}

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

// LegacyNetworkSettings selects Login-first negotiation with immediately
// active flate batches for RakNet v10 clients.
func (Protocol) LegacyNetworkSettings() packet.Compression { return packet.FlateCompression }

// Encryption selects the CFB8 cipher used before Minecraft 1.16.220.
func (Protocol) Encryption(key [32]byte) packet.Encryption { return packet.NewCFB8Encryption(key) }

// PreSpawnPackets restores the protocol-419 spawn order. The 1.16.100 client
// requires biome definitions before PlayStatusPlayerSpawn and may crash while
// initialising achievements if they have not arrived yet.
func (Protocol) PreSpawnPackets() []packet.Packet {
	return []packet.Packet{&packet.BiomeDefinitionList{}}
}

// ReuseBiomePalettes reports that protocol 419 uses 2D biome bytes rather than paletted biome storages.
func (Protocol) ReuseBiomePalettes() bool { return false }

// NetworkChunkRange returns the fixed target world height represented by protocol 419.
func (Protocol) NetworkChunkRange() (minY, maxY int16) { return 0, 255 }

// NetworkSubChunkVersion returns the target's index-free sub-chunk version.
func (Protocol) NetworkSubChunkVersion() byte { return 8 }

// NetworkBiomes2D reports that protocol 419 carries one biome byte per X/Z column.
func (Protocol) NetworkBiomes2D() bool { return true }

func (p Protocol) MapBlockRuntimeID(runtimeID uint32) (uint32, bool) {
	if p.runtime == nil || p.runtime.blocks == nil {
		return 0, false
	}
	target, valid, _ := p.runtime.blocks.MapNative(runtimeID)
	return target, valid
}

func (p Protocol) MappingReport() MappingReport {
	if p.runtime == nil {
		return MappingReport{}
	}
	report := MappingReport{
		NativeBlocks:   p.runtime.blocks.NativeCount(),
		TargetBlocks:   p.runtime.blocks.TargetCount(),
		BlockFallbacks: p.runtime.blocks.Fallbacks(),
		CreativeItems:  p.runtime.creativeCount,
	}
	if items := p.runtime.currentItemMapper(); items != nil {
		report.NativeItems = items.NativeCount()
		report.TargetItems = items.TargetCount()
		report.ItemFallbacks = items.Fallbacks()
		report.TargetItemFallbacks = items.TargetFallbacks()
	}
	return report
}

func (p Protocol) FallbackReportJSON(nativeDragonflyCommit, targetDragonflyCommit string) ([]byte, error) {
	report := p.MappingReport()
	blocks := make([]fallbackBlock, len(report.BlockFallbacks))
	for index, fallback := range report.BlockFallbacks {
		keys := make([]string, 0, len(fallback.Properties))
		for key := range fallback.Properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		properties := make([]fallbackProperty, len(keys))
		for propertyIndex, key := range keys {
			value := fallback.Properties[key]
			properties[propertyIndex] = fallbackProperty{Name: key, Type: fmt.Sprintf("%T", value), Value: fmt.Sprint(value)}
		}
		blocks[index] = fallbackBlock{NativeRuntimeID: fallback.NativeRuntimeID, Name: fallback.Name, Properties: properties, TargetRuntimeID: fallback.TargetRuntimeID}
	}
	encoded, err := json.MarshalIndent(map[string]any{
		"schema_version":          1,
		"native_dragonfly_commit": nativeDragonflyCommit,
		"target_dragonfly_commit": targetDragonflyCommit,
		"block_policy":            "substitute minecraft:air",
		"item_policy":             "hide clientbound and reject serverbound",
		"block_fallbacks":         blocks,
		"item_fallbacks":          report.ItemFallbacks,
		"target_item_fallbacks":   report.TargetItemFallbacks,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func (p Protocol) Packets(listener bool) packet.Pool {
	pool := p.base.Packets(listener)
	for currentID := range packetMarshals {
		delete(pool, currentID)
		delete(pool, targetPacketID(currentID))
	}
	for id, marshal := range packetMarshals {
		var current packet.Pool
		if listener {
			current = packet.NewClientPool()
		} else {
			current = packet.NewServerPool()
		}
		if constructor, ok := current[id]; ok {
			targetID := targetPacketID(id)
			pool[targetID] = translatedConstructorID(constructor, marshal, targetID)
		}
	}
	for id := range pool {
		if id > maxLegacyPacketID {
			delete(pool, id)
		}
	}
	pool[idTickSync] = func() packet.Packet { return &tickSync{} }
	pool[packet.IDStructureTemplateDataRequest] = func() packet.Packet { return &structureTemplateResponse475{} }
	if listener {
		pool[idPassengerJump] = func() packet.Packet { return &passengerJump{} }
		pool[idCraftingEvent] = func() packet.Packet { return &craftingEvent{} }
		pool[idPlayerInput] = func() packet.Packet { return &playerInput{} }
		pool[idItemFrameDropItem] = func() packet.Packet { return &itemFrameDropItem{} }
	} else {
		pool[idAddEntity] = func() packet.Packet { return &addEntity{} }
		pool[idRemoveEntity] = func() packet.Packet { return &removeEntity{} }
	}
	return pool
}

func (p Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return &wireReader{wireIO: newWireIO(p.base.NewReader(r, shieldID, enableLimits), true, p.runtime)}
}

func (p Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return &wireWriter{wireIO: newWireIO(p.base.NewWriter(w, shieldID), false, p.runtime)}
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if _, legacyOnly := pk.(legacyOnlyPacket); legacyOnly {
		return nil
	}
	if translated, ok := pk.(*translatedPacket); ok {
		converted := p.convertGameplayToLatest(translated.inner, conn)
		if _, startGame := translated.inner.(*packet.StartGame); startGame && p.runtime != nil {
			if items := p.runtime.currentItemMapper(); items != nil {
				converted = append(converted, &packet.ItemRegistry{Items: items.TargetEntries()})
			}
		}
		return converted
	}
	base := p.base.ConvertToLatest(pk, conn)
	converted := make([]packet.Packet, 0, len(base))
	for _, current := range base {
		converted = append(converted, p.convertGameplayToLatest(current, conn)...)
	}
	return converted
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	mapped := p.convertGameplayFromLatest(pk, conn)
	converted := make([]packet.Packet, 0, len(mapped))
	for _, candidate := range mapped {
		for _, current := range splitLegacyPacket(candidate) {
			if targetPacketID(current.ID()) > maxLegacyPacketID {
				continue
			}
			if _, drop := droppedPacketIDs[current.ID()]; drop {
				continue
			}
			if marshal, ok := packetMarshals[current.ID()]; ok {
				converted = append(converted, translatedID(current, marshal, targetPacketID(current.ID())))
				continue
			}
			converted = append(converted, p.base.ConvertFromLatest(current, conn)...)
		}
	}
	return converted
}

func targetPacketID(currentID uint32) uint32 {
	switch currentID {
	case packet.IDMapCreateLockedCopy:
		return 130
	case packet.IDOnScreenTextureAnimation:
		return 131
	default:
		return currentID
	}
}

func splitLegacyPacket(pk packet.Packet) []packet.Packet {
	if subChunks, ok := pk.(*packet.SubChunk); ok && len(subChunks.SubChunkEntries) > 1 {
		packets := make([]packet.Packet, 0, len(subChunks.SubChunkEntries))
		for _, entry := range subChunks.SubChunkEntries {
			cloned := *subChunks
			cloned.Position = protocol.SubChunkPos{
				subChunks.Position[0] + int32(entry.Offset[0]),
				subChunks.Position[1] + int32(entry.Offset[1]),
				subChunks.Position[2] + int32(entry.Offset[2]),
			}
			entry.Offset = protocol.SubChunkOffset{}
			cloned.SubChunkEntries = []protocol.SubChunkEntry{entry}
			packets = append(packets, &cloned)
		}
		return packets
	}
	list, ok := pk.(*packet.PlayerList)
	if !ok {
		return []packet.Packet{pk}
	}
	var additions, removals []protocol.PlayerListEntry
	for _, entry := range list.Entries {
		if entry.ActionType == protocol.PlayerListActionRemove {
			removals = append(removals, entry)
		} else {
			additions = append(additions, entry)
		}
	}
	if len(additions) == 0 || len(removals) == 0 {
		return []packet.Packet{pk}
	}
	return []packet.Packet{
		&packet.PlayerList{Entries: additions},
		&packet.PlayerList{Entries: removals},
	}
}

func isStableGameVersion(version string) bool {
	return version == Version
}
