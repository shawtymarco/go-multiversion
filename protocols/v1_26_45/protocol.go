// Package v1_26_45 implements the outgoing Minecraft Bedrock 1.26.45
// protocol 2169 native release against the protocol-2192 native model.
package v1_26_45

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

const (
	ID      int32 = 2169
	Version       = "1.26.45"
)

// Protocol implements the wire boundary for Minecraft 1.26.45. Registry and
// chunk-palette mappings are configured separately before consumer enablement.
type Protocol struct {
	runtime *runtimeData
}

type MappingReport struct {
	NativeBlocks, TargetBlocks int
	NativeItems, TargetItems   int
	BlockFallbacks             []mapping.BlockFallback
	ItemFallbacks              []mapping.ItemFallback
}

func New() minecraft.Protocol { return &Protocol{} }

func NewWithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return NewWithRegistries(native, nil)
}

func NewWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (minecraft.Protocol, error) {
	runtime, err := newRuntimeData(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return &Protocol{runtime: runtime}, nil
}

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

// MapBlockRuntimeID maps a protocol-2192 native runtime ID to the frozen
// protocol-2169 block registry before Dragonfly hashes chunk payloads.
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
	}
	if items := p.runtime.currentItemMapper(); items != nil {
		report.NativeItems = items.NativeCount()
		report.TargetItems = items.TargetCount()
		report.ItemFallbacks = items.Fallbacks()
	}
	return report
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
	delete(pool, packet.IDSetPlayerFurnaceOptions)
	delete(pool, packet.IDRecordStarted)
	for id, marshal := range packetMarshals {
		constructor, ok := pool[id]
		if ok {
			pool[id] = translatedConstructor(constructor, marshal)
		}
	}
	return pool
}

func (p Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return &wireReader{wireIO: newWireIO(protocol.NewReader(r, shieldID, enableLimits), true, p.runtime)}
}

func (p Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return &wireWriter{wireIO: newWireIO(protocol.NewWriter(w, shieldID), false, p.runtime)}
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	return p.convertToLatest(pk, conn)
}

func (p Protocol) convertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if translated, ok := pk.(*translatedPacket); ok {
		pk = translated.inner
	}
	return p.convertGameplayToLatest(pk, conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	mapped := p.convertGameplayFromLatest(pk, conn)
	converted := make([]packet.Packet, 0, len(mapped))
	for _, current := range mapped {
		converted = append(converted, convertWireFromLatest(current)...)
	}
	return converted
}

func convertWireFromLatest(pk packet.Packet) []packet.Packet {
	switch current := pk.(type) {
	case *packet.SetPlayerFurnaceOptions, *packet.RecordStarted:
		return nil
	case *packet.BossEvent:
		switch current.EventType {
		case packet.BossEventRegisterPlayer, packet.BossEventUnregisterPlayer, packet.BossEventRequest:
			// Protocol 2192 removed the target player ID, so these events cannot
			// be represented faithfully for a protocol-2169 client.
			return nil
		}
	case *packet.Disconnect:
		if current.Reason > packet.DisconnectReasonEditorNotAllowed {
			cloned := *current
			cloned.Reason = packet.DisconnectReasonUnknown
			pk = &cloned
		}
	case *packet.StartGame:
		cloned := *current
		cloned.GameVersion, cloned.BaseGameVersion = Version, Version
		pk = &cloned
	}
	marshal, ok := packetMarshals[pk.ID()]
	if !ok {
		return []packet.Packet{pk}
	}
	return []packet.Packet{translated(pk, marshal)}
}

var packetMarshals = map[uint32]packetMarshal{
	packet.IDBossEvent:                     marshalBossEvent,
	packet.IDCameraPresets:                 marshalCameraPresets,
	packet.IDClientBoundAttributeLayerSync: marshalClientBoundAttributeLayerSync,
	packet.IDDimensionData:                 marshalDimensionData,
	packet.IDInventoryTransaction:          marshalInventoryTransaction,
	packet.IDItemStackResponse:             marshalItemStackResponse,
	packet.IDMoveActorDelta:                marshalMoveActorDelta,
	packet.IDPlaySound:                     marshalPlaySound,
	packet.IDPlayerAuthInput:               marshalPlayerAuthInput,
	packet.IDServerBoundDiagnostics:        marshalServerBoundDiagnostics,
	packet.IDSubChunk:                      marshalSubChunk,
}
