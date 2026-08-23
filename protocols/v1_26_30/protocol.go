// Package v1_26_30 implements the Minecraft Bedrock 1.26.3x protocol 1001
// family against the current native gophertunnel model.
package v1_26_30

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	// ID is the Minecraft Bedrock 1.26.3x protocol ID.
	ID int32 = 1001
	// Version is the newest stable Minecraft version in the protocol 1001 family.
	Version = "1.26.36"
)

// Protocol implements minecraft.Protocol for stable Minecraft 1.26.30 through
// 1.26.36. It is intentionally not included in multiversion.Protocols until
// registry and chunk conversion are complete.
type Protocol struct{}

// New creates a protocol 1001 wire adapter.
func New() minecraft.Protocol { return Protocol{} }

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

func (Protocol) Packets(listener bool) packet.Pool {
	base := packet.NewServerPool()
	if listener {
		base = packet.NewClientPool()
	}
	pool := make(packet.Pool, len(base))
	for id, constructor := range base {
		pool[id] = constructor
	}

	// Protocol 1001 left packet ID 16 unused. UpdateBlock was not accepted as
	// a client-originating packet by the historical listener pool.
	delete(pool, packet.IDServerPlayerPostMovePosition)
	if listener {
		delete(pool, packet.IDUpdateBlock)
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

func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	base := newWireIO(protocol.NewReader(r, shieldID, enableLimits), true)
	return &wireReader{wireIO: base}
}

func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	base := newWireIO(protocol.NewWriter(w, shieldID), false)
	return &wireWriter{wireIO: base}
}

func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	if translated, ok := pk.(*translatedPacket); ok {
		return []packet.Packet{translated.inner}
	}
	return []packet.Packet{pk}
}

func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if pk.ID() == packet.IDServerPlayerPostMovePosition {
		return nil
	}
	return downgradePacket(pk, conn)
}
