// Package v1_21_50 implements the Minecraft Bedrock 1.21.5x protocol 766
// family against the current native gophertunnel model.
package v1_21_50

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

const (
	// ID is the Minecraft Bedrock 1.21.5x protocol ID.
	ID int32 = 766
	// Version is the newest stable Minecraft version in the protocol 766 family.
	Version = "1.21.51"
)

// Protocol implements minecraft.Protocol for stable Minecraft 1.21.50 through
// 1.21.51. Registry-aware consumers may advertise it after construction.
type Protocol struct {
	runtime *runtimeData
}

// MappingReport summarises the exact registries and explicit fallbacks used by
// a configured protocol-766 adapter.
type MappingReport struct {
	NativeBlocks, TargetBlocks int
	NativeItems, TargetItems   int
	BlockFallbacks             []mapping.BlockFallback
	ItemFallbacks              []mapping.ItemFallback
	TargetItemFallbacks        []mapping.TargetItemFallback
	CreativeItems              int
	FurnaceRecipes             int
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

// New creates a protocol 766 wire adapter.
func New() minecraft.Protocol { return &Protocol{} }

// NewWithBlockRegistry creates a fully configured protocol-766 adapter using
// the current native block registry for direct semantic mappings.
func NewWithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return NewWithRegistries(native, nil)
}

// NewWithRegistries creates a fully configured adapter and validates both the
// current block and item registries before it may be advertised.
func NewWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (minecraft.Protocol, error) {
	runtime, err := newRuntimeData(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return &Protocol{runtime: runtime}, nil
}

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

// MapBlockRuntimeID maps a current native runtime ID to the protocol-766
// network registry. The method is a generic Dragonfly chunk-encoding hook.
func (p Protocol) MapBlockRuntimeID(runtimeID uint32) (uint32, bool) {
	if p.runtime == nil || p.runtime.blocks == nil {
		return 0, false
	}
	target, valid, _ := p.runtime.blocks.MapNative(runtimeID)
	return target, valid
}

// MappingReport returns an immutable snapshot of configured mapping evidence.
func (p Protocol) MappingReport() MappingReport {
	if p.runtime == nil {
		return MappingReport{}
	}
	report := MappingReport{
		NativeBlocks:   p.runtime.blocks.NativeCount(),
		TargetBlocks:   p.runtime.blocks.TargetCount(),
		BlockFallbacks: p.runtime.blocks.Fallbacks(),
		CreativeItems:  len(p.runtime.creative),
	}
	if items := p.runtime.currentItemMapper(); items != nil {
		report.NativeItems = items.NativeCount()
		report.TargetItems = items.TargetCount()
		report.ItemFallbacks = items.Fallbacks()
		report.TargetItemFallbacks = items.TargetFallbacks()
		report.FurnaceRecipes = len(targetFurnaceRecipes(p.runtime))
	}
	return report
}

// FallbackReportJSON renders every fallback deterministically for audit and
// release manifests.
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
		blocks[index] = fallbackBlock{
			NativeRuntimeID: fallback.NativeRuntimeID,
			Name:            fallback.Name,
			Properties:      properties,
			TargetRuntimeID: fallback.TargetRuntimeID,
		}
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

func (Protocol) Packets(listener bool) packet.Pool {
	base := packet.NewServerPool()
	if listener {
		base = packet.NewClientPool()
	}
	pool := make(packet.Pool, len(base))
	for id, constructor := range base {
		pool[id] = constructor
	}

	// Protocol 766 retains PassengerJump, TickSync, PlayerInput and the legacy
	// item-component slot, and ends at CameraAimAssistPresets.
	delete(pool, packet.IDServerPlayerPostMovePosition)
	if listener {
		delete(pool, packet.IDUpdateBlock)
		delete(pool, packet.IDSimpleEvent)
		for id := uint32(packet.IDClientCameraAimAssist); id <= uint32(packet.IDPartyDestinationCookieResponse); id++ {
			delete(pool, id)
		}
	} else {
		for id := uint32(packet.IDClientCameraAimAssist); id <= uint32(packet.IDPartyDestinationCookieResponse); id++ {
			delete(pool, id)
		}
	}
	pool[idTickSync766] = func() packet.Packet { return &tickSync766{} }
	pool[idPassengerJump766] = func() packet.Packet { return &passengerJump766{} }
	pool[idPlayerInput766] = func() packet.Packet { return &playerInput766{} }
	if listener {
	} else {
		pool[idSetMovementAuthority766] = func() packet.Packet { return &setMovementAuthority766{} }
		pool[idCompressedBiomeDefinitionList766] = func() packet.Packet { return &compressedBiomeDefinitionList766{} }
	}
	for id, marshal := range packetMarshals {
		constructor, ok := pool[id]
		if !ok {
			continue
		}
		pool[id] = translatedConstructor(constructor, marshal)
	}
	return pool
}

func (p Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	base := newWireIO(protocol.NewReader(r, shieldID, enableLimits), true, p.runtime)
	return &wireReader{wireIO: base}
}

func (p Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	base := newWireIO(protocol.NewWriter(w, shieldID), false, p.runtime)
	return &wireWriter{wireIO: base}
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if _, legacyOnly := pk.(legacyOnlyPacket766); legacyOnly {
		return nil
	}
	if translated, ok := pk.(*translatedPacket); ok {
		pk = translated.inner
	}
	return p.convertGameplayToLatest(pk, conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if _, supported := p.Packets(false)[pk.ID()]; !supported {
		return nil
	}
	mapped := p.convertGameplayFromLatest(pk, conn)
	converted := make([]packet.Packet, 0, len(mapped))
	for _, current := range mapped {
		converted = append(converted, downgradePacket(current, conn)...)
	}
	return converted
}
