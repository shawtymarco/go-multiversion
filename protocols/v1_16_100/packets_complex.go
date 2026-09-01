package v1_16_100

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const legacyPlayerAuthInputFlagCount = 25

func marshalPlayerAuthInput(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAuthInput)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Vec3(&pk.Position)
	io.Vec2(&pk.MoveVector)
	io.Float32(&pk.HeadYaw)
	marshalLegacyInputFlags(io, &pk.InputData)
	io.Varuint32(&pk.InputMode)
	io.Varuint32(&pk.PlayMode)
	if pk.PlayMode == 4 {
		io.Vec3(&pk.CameraOrientation)
	} else if io.reading {
		pk.CameraOrientation = mgl32.Vec3{}
	}
	io.Varuint64(&pk.Tick)
	io.Vec3(&pk.Delta)

	if io.reading {
		pk.ItemInteractionData = protocol.Optional[protocol.UseItemTransactionData]{}
		pk.ItemStackRequest = protocol.Optional[protocol.ItemStackRequest]{}
		pk.BlockActions = protocol.Optional[[]protocol.PlayerBlockAction]{}
		pk.InteractionModel = 0
		pk.InteractPitch, pk.InteractYaw = pk.Pitch, pk.Yaw
		pk.VehicleRotation = protocol.Optional[mgl32.Vec2]{}
		pk.ClientPredictedVehicle = protocol.Optional[int64]{}
		pk.AnalogueMoveVector = pk.MoveVector
		pk.RawMoveVector = pk.MoveVector
	}
}

func marshalLegacyInputFlags(io *wireIO, flags *protocol.InputFlags) {
	var bits uint64
	if !io.reading {
		for index := 0; index < legacyPlayerAuthInputFlagCount && index < flags.Len(); index++ {
			if flags.Load(index) {
				bits |= uint64(1) << index
			}
		}
	}
	io.Varuint64(&bits)
	if io.reading {
		*flags = protocol.NewInputFlags(packet.InputFlagCount)
		for index := 0; index < legacyPlayerAuthInputFlagCount; index++ {
			if bits&(uint64(1)<<index) != 0 {
				flags.Set(index)
			}
		}
	}
}

func marshalPlayerBlockAction(io *wireIO, action *protocol.PlayerBlockAction) {
	io.Varint32(&action.Action)
	switch action.Action {
	case protocol.PlayerActionStartBreak, protocol.PlayerActionAbortBreak, protocol.PlayerActionCrackBreak,
		protocol.PlayerActionPredictDestroyBlock, protocol.PlayerActionContinueDestroyBlock:
		io.BlockPos(&action.BlockPos)
		io.Varint32(&action.Face)
	}
}
