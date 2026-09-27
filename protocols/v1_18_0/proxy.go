package v1_18_0

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

// ProxyWireProtocol separates codecs from registry conversion for proxy consumers.
// The copy shares only immutable, session-local registries with this Protocol.
func (p Protocol) ProxyWireProtocol() minecraft.Protocol { p.wireOnly = true; return &p }
func (p Protocol) ProxyToNative(pk packet.Packet) []packet.Packet {
	if _, oldOnly := pk.(legacyOnlyPacket); oldOnly {
		return nil
	}
	return p.convertGameplayToLatest(pk, nil)
}
func (p Protocol) ProxyFromNative(pk packet.Packet) []packet.Packet {
	return p.convertGameplayFromLatest(pk, nil)
}
func (p Protocol) ProxyMappings() (*mapping.BlockMapper, *mapping.ItemMapper) {
	if p.runtime == nil {
		return nil, nil
	}
	return p.runtime.blocks, p.runtime.currentItemMapper()
}

// ProxyTickReply answers the target-only clock synchronisation packet locally.
func (pk *tickSync) ProxyTickReply(tick int64) packet.Packet {
	copy := *pk
	copy.ServerReceptionTimestamp = tick
	return &copy
}
