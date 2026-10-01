package v1_26_45

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_50"
)

type packetMarshal func(*wireIO, packet.Packet)

type translatedPacket struct {
	inner   packet.Packet
	marshal packetMarshal
}

func (pk *translatedPacket) ID() uint32 { return pk.inner.ID() }

func (pk *translatedPacket) Marshal(io protocol.IO) {
	if current, ok := io.(*wireReader); ok {
		pk.marshal(current.wireIO, pk.inner)
		return
	}
	if current, ok := io.(*wireWriter); ok {
		pk.marshal(current.wireIO, pk.inner)
		return
	}
	// Older adapters reuse only these wire deltas. Their own IO and registry
	// conversion remain authoritative, so no intermediate semantic mapping runs.
	_, reading := io.(interface{ SliceLength(uint32, uint32) })
	pk.marshal(newWireIO(io, reading), pk.inner)
}

// WrapWirePacket applies only the layouts shared with older releases. It never
// changes item/block IDs, registries, version strings, or gameplay semantics.
func WrapWirePacket(pk packet.Packet) packet.Packet {
	if marshal, ok := packetMarshals[pk.ID()]; ok {
		return translated(pk, marshal)
	}
	return v1_26_50.WrapWirePacket(pk)
}

// UnwrapWirePacket returns the native-shaped packet decoded by WrapWirePacket.
func UnwrapWirePacket(pk packet.Packet) packet.Packet {
	if translated, ok := pk.(*translatedPacket); ok {
		return translated.inner
	}
	return v1_26_50.UnwrapWirePacket(pk)
}

func translatedConstructor(constructor func() packet.Packet, marshal packetMarshal) func() packet.Packet {
	return func() packet.Packet {
		return &translatedPacket{inner: constructor(), marshal: marshal}
	}
}

func translated(pk packet.Packet, marshal packetMarshal) packet.Packet {
	return &translatedPacket{inner: pk, marshal: marshal}
}
