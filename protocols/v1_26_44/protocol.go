// Package v1_26_44 implements Minecraft Bedrock 1.26.44 protocol 2168.
package v1_26_44

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	// ID is the Minecraft Bedrock 1.26.44 protocol ID.
	ID int32 = 2168
	// Version is the Minecraft Bedrock version implemented by this adapter.
	Version = "1.26.44"
)

// Protocol implements minecraft.Protocol for Minecraft Bedrock 1.26.44.
type Protocol struct{}

// New creates a Minecraft Bedrock 1.26.44 protocol adapter.
func New() minecraft.Protocol {
	return Protocol{}
}

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

func (Protocol) Packets(listener bool) packet.Pool {
	if listener {
		return packet.NewClientPool()
	}
	pool := packet.NewServerPool()
	pool[packet.IDSetScore] = func() packet.Packet { return &setScore{} }
	return pool
}

func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return protocol.NewReader(r, shieldID, enableLimits)
}

func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return protocol.NewWriter(w, shieldID)
}

func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch legacy := pk.(type) {
	case *setScore:
		return []packet.Packet{legacy.latest()}
	case *packet.ResourcePackStack:
		latest := *legacy
		latest.BaseGameVersion = protocol.CurrentVersion
		return []packet.Packet{&latest}
	}
	return []packet.Packet{pk}
}

func (Protocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch latest := pk.(type) {
	case *packet.SetScore:
		return []packet.Packet{setScoreFromLatest(latest)}
	case *packet.ResourcePackStack:
		legacy := *latest
		legacy.BaseGameVersion = Version
		return []packet.Packet{&legacy}
	}
	return []packet.Packet{pk}
}
