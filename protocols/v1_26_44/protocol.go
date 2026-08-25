// Package v1_26_44 implements the Minecraft Bedrock 1.26.40-1.26.44
// protocol 2168 family.
package v1_26_44

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_45"
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
type Protocol struct {
	base minecraft.Protocol
}

// New creates a Minecraft Bedrock protocol 2168 family adapter.
func New() minecraft.Protocol {
	return Protocol{base: v1_26_45.New()}
}

// NewWithBlockRegistry creates a protocol-2168 adapter using the shared
// 1.26.40-1.26.45 registry baseline against protocol-2192 native data.
func NewWithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	base, err := v1_26_45.NewWithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	return Protocol{base: base}, nil
}

// NewWithRegistries validates the shared 1.26.40-1.26.45 block and item
// snapshots before the protocol-2168 adapter is advertised.
func NewWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (minecraft.Protocol, error) {
	base, err := v1_26_45.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return Protocol{base: base}, nil
}

func (Protocol) ID() int32   { return ID }
func (Protocol) Ver() string { return Version }

func (p Protocol) baseProtocol() minecraft.Protocol {
	if p.base == nil {
		return v1_26_45.New()
	}
	return p.base
}

func (p Protocol) MapBlockRuntimeID(runtimeID uint32) (uint32, bool) {
	mapper, ok := p.baseProtocol().(interface {
		MapBlockRuntimeID(uint32) (uint32, bool)
	})
	if !ok {
		return 0, false
	}
	return mapper.MapBlockRuntimeID(runtimeID)
}

func (p Protocol) Packets(listener bool) packet.Pool {
	pool := p.baseProtocol().Packets(listener)
	if !listener {
		pool[packet.IDSetScore] = func() packet.Packet { return &setScore{} }
	}
	return pool
}

func (p Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return p.baseProtocol().NewReader(r, shieldID, enableLimits)
}

func (p Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return p.baseProtocol().NewWriter(w, shieldID)
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if legacy, ok := pk.(*setScore); ok {
		return []packet.Packet{legacy.latest()}
	}
	return p.baseProtocol().ConvertToLatest(pk, conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	gameVersion := ""
	if conn != nil {
		gameVersion = conn.ClientData().GameVersion
	}
	converted := p.baseProtocol().ConvertFromLatest(pk, conn)
	result := make([]packet.Packet, 0, len(converted))
	for _, current := range converted {
		if start, ok := current.(*packet.StartGame); ok {
			cloned := *start
			cloned.GameVersion, cloned.BaseGameVersion = targetGameVersion(gameVersion), targetGameVersion(gameVersion)
			current = &cloned
		}
		result = append(result, convertFromLatestForGameVersion(current, gameVersion)...)
	}
	return result
}

func targetGameVersion(gameVersion string) string {
	switch gameVersion {
	case "1.26.40", "1.26.41", "1.26.42", "1.26.43", "1.26.44":
		return gameVersion
	default:
		return Version
	}
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
