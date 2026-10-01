// Package v1_26_50 freezes the outgoing native protocol 2193 wire against the
// 1.26.60.29/protocol-2223 native model. Registry mapping is still deferred.
package v1_26_50

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/packetio"
)

const (
	ID      int32 = 2193
	Version       = "1.26.50"
)

type Protocol struct{}

func New() minecraft.Protocol { return Protocol{} }
func (Protocol) ID() int32    { return ID }
func (Protocol) Ver() string  { return Version }

func (Protocol) Packets(listener bool) packet.Pool {
	pool := packet.NewServerPool()
	if listener {
		pool = packet.NewClientPool()
	}
	for id, constructor := range pool {
		if id > packet.IDRecordStarted {
			delete(pool, id)
			continue
		}
		pool[id] = func() packet.Packet { return WrapWirePacket(constructor()) }
	}
	return pool
}

type reader struct{ protocol.IO }
type writer struct{ protocol.IO }

func (r *reader) SliceLength(value, maximum uint32) {
	if limits, ok := r.IO.(interface{ SliceLength(uint32, uint32) }); ok {
		limits.SliceLength(value, maximum)
	}
}
func (r *reader) StackRequestAction(value *protocol.StackRequestAction) {
	packetio.StackRequestAction2193(r, value, true)
}
func (w *writer) StackRequestAction(value *protocol.StackRequestAction) {
	packetio.StackRequestAction2193(w, value, false)
}
func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, limits bool) protocol.IO {
	return &reader{protocol.NewReader(r, shieldID, limits)}
}
func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return &writer{protocol.NewWriter(w, shieldID)}
}

func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	return []packet.Packet{UnwrapWirePacket(pk)}
}
func (Protocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	if pk.ID() > packet.IDRecordStarted {
		return nil
	}
	if start, ok := pk.(*packet.StartGame); ok {
		copy := *start
		copy.GameVersion, copy.BaseGameVersion = Version, Version
		pk = &copy
	}
	return []packet.Packet{WrapWirePacket(pk)}
}

type wireMarshal func(protocol.IO, packet.Packet, bool)
type wirePacket struct {
	inner   packet.Packet
	marshal wireMarshal
}

func (p *wirePacket) ID() uint32 { return p.inner.ID() }
func (p *wirePacket) Marshal(io protocol.IO) {
	_, reading := io.(interface{ SliceLength(uint32, uint32) })
	p.marshal(io, p.inner, reading)
}

// WrapWirePacket applies only the frozen 2193 wire differences. Concrete
// historical wrappers keep their own layout and direct registry conversion.
func WrapWirePacket(pk packet.Packet) packet.Packet {
	var marshal wireMarshal
	switch pk.(type) {
	case *packet.InventoryTransaction:
		marshal = marshalInventoryTransaction
	case *packet.StartGame:
		marshal = marshalStartGame
	case *packet.AddActor:
		marshal = marshalAddActor
	case *packet.AddPlayer:
		marshal = marshalAddPlayer
	case *packet.Animate:
		marshal = marshalAnimate
	case *packet.LevelChunk:
		marshal = marshalLevelChunk
	case *packet.PlayerList:
		marshal = marshalPlayerList
	case *packet.PlayerSkin:
		marshal = marshalPlayerSkin
	case *packet.DimensionData:
		marshal = marshalDimensionData
	case *packet.EducationSettings:
		marshal = marshalEducationSettings
	case *packet.ServerBoundDiagnostics:
		marshal = marshalServerBoundDiagnostics
	case *packet.ClientboundUpdateSoundData:
		marshal = marshalClientboundUpdateSoundData
	case *packet.ClientBoundAttributeLayerSync:
		marshal = marshalClientBoundAttributeLayerSync
	}
	if marshal == nil {
		return pk
	}
	return &wirePacket{pk, marshal}
}
func UnwrapWirePacket(pk packet.Packet) packet.Packet {
	if p, ok := pk.(*wirePacket); ok {
		return p.inner
	}
	return pk
}
