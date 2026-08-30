package v1_21_50

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	idPassengerJump766                 = 20
	idTickSync766                      = 23
	idPlayerInput766                   = 57
	idCompressedBiomeDefinitionList766 = uint32(packet.IDTrimData - 1)
	idSetMovementAuthority766          = uint32(packet.IDCameraAimAssistPresets - 1)
)

type legacyOnlyPacket766 interface{ legacyOnly766() }

type passengerJump766 struct{ JumpStrength int32 }

func (*passengerJump766) ID() uint32                { return idPassengerJump766 }
func (*passengerJump766) legacyOnly766()            {}
func (pk *passengerJump766) Marshal(io protocol.IO) { io.Varint32(&pk.JumpStrength) }

type tickSync766 struct{ ClientRequestTimestamp, ServerReceptionTimestamp int64 }

func (*tickSync766) ID() uint32     { return idTickSync766 }
func (*tickSync766) legacyOnly766() {}
func (pk *tickSync766) Marshal(io protocol.IO) {
	io.Int64(&pk.ClientRequestTimestamp)
	io.Int64(&pk.ServerReceptionTimestamp)
}

type playerInput766 struct {
	Movement          mgl32.Vec2
	Jumping, Sneaking bool
}

func (*playerInput766) ID() uint32     { return idPlayerInput766 }
func (*playerInput766) legacyOnly766() {}
func (pk *playerInput766) Marshal(io protocol.IO) {
	io.Vec2(&pk.Movement)
	io.Bool(&pk.Jumping)
	io.Bool(&pk.Sneaking)
}

type setMovementAuthority766 struct{ MovementType byte }

func (*setMovementAuthority766) ID() uint32                { return idSetMovementAuthority766 }
func (*setMovementAuthority766) legacyOnly766()            {}
func (pk *setMovementAuthority766) Marshal(io protocol.IO) { io.Uint8(&pk.MovementType) }

type compressedBiomeDefinitionList766 struct{ SerialisedBiomeDefinitions []byte }

func (*compressedBiomeDefinitionList766) ID() uint32     { return idCompressedBiomeDefinitionList766 }
func (*compressedBiomeDefinitionList766) legacyOnly766() {}
func (pk *compressedBiomeDefinitionList766) Marshal(io protocol.IO) {
	io.Bytes(&pk.SerialisedBiomeDefinitions)
}
