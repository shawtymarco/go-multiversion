// Package v1_26_44 implements the Minecraft Bedrock 1.26.40-1.26.44
// protocol 2168 family.
package v1_26_44

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	// ID is the Minecraft Bedrock 1.26.40-1.26.44 protocol ID.
	ID int32 = 2168
	// Version is the newest Minecraft Bedrock version implemented by this adapter.
	Version = "1.26.44"
)

// Protocol implements minecraft.Protocol for Minecraft Bedrock 1.26.40-1.26.44.
// These releases reuse protocol ID 2168, so outgoing conversions use the
// connection's login GameVersion when a patch-specific layout is required.
type Protocol struct{}

// New creates a Minecraft Bedrock protocol 2168 family adapter.
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
	if legacy, ok := pk.(*setScore); ok {
		return []packet.Packet{legacy.latest()}
	}
	return []packet.Packet{pk}
}

func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	gameVersion := ""
	if conn != nil {
		gameVersion = conn.ClientData().GameVersion
	}
	return convertFromLatestForGameVersion(pk, gameVersion)
}

func convertFromLatestForGameVersion(pk packet.Packet, gameVersion string) []packet.Packet {
	latest, ok := pk.(*packet.SetScore)
	if !ok || !usesV1_26_44SetScoreLayout(gameVersion) {
		return []packet.Packet{pk}
	}
	return []packet.Packet{setScoreFromLatest(latest)}
}

func usesV1_26_44SetScoreLayout(gameVersion string) bool {
	switch gameVersion {
	case "1.26.40", "1.26.41", "1.26.42", "1.26.43":
		return false
	default:
		// Preserve the original 1.26.44 adapter behaviour for nil connections,
		// missing client data, and unrecognised protocol 2168 patch versions.
		return true
	}
}
