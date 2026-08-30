package v1_18_0

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetMarshal func(*wireIO, packet.Packet)

type translatedPacket struct {
	inner    packet.Packet
	marshal  packetMarshal
	targetID uint32
}

func (pk *translatedPacket) ID() uint32 {
	if pk.targetID != 0 {
		return pk.targetID
	}
	return pk.inner.ID()
}

func (pk *translatedPacket) Marshal(io protocol.IO) {
	pk.marshal(asWireIO(io), pk.inner)
}

func translatedConstructor(constructor func() packet.Packet, marshal packetMarshal) func() packet.Packet {
	return func() packet.Packet { return &translatedPacket{inner: constructor(), marshal: marshal} }
}

func translatedConstructorID(constructor func() packet.Packet, marshal packetMarshal, targetID uint32) func() packet.Packet {
	return func() packet.Packet {
		return &translatedPacket{inner: constructor(), marshal: marshal, targetID: targetID}
	}
}

func translated(pk packet.Packet, marshal packetMarshal) packet.Packet {
	return &translatedPacket{inner: pk, marshal: marshal}
}

func translatedID(pk packet.Packet, marshal packetMarshal, targetID uint32) packet.Packet {
	return &translatedPacket{inner: pk, marshal: marshal, targetID: targetID}
}
