package v1_18_0

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const legacyPlayerAuthInputFlagCount = 64

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

	if pk.InputData.Load(packet.InputFlagPerformItemInteraction) {
		value, _ := pk.ItemInteractionData.Value()
		marshalPlayerInventoryAction(io, &value)
		pk.ItemInteractionData = protocol.Option(value)
	} else if io.reading {
		pk.ItemInteractionData = protocol.Optional[protocol.UseItemTransactionData]{}
	}
	if pk.InputData.Load(packet.InputFlagPerformItemStackRequest) {
		value, _ := pk.ItemStackRequest.Value()
		marshalItemStackRequest(io, &value)
		pk.ItemStackRequest = protocol.Option(value)
	} else if io.reading {
		pk.ItemStackRequest = protocol.Optional[protocol.ItemStackRequest]{}
	}
	if pk.InputData.Load(packet.InputFlagPerformBlockActions) {
		value, _ := pk.BlockActions.Value()
		count := int32(len(value))
		io.Varint32(&count)
		if io.reading {
			value = make([]protocol.PlayerBlockAction, count)
		}
		for index := range value {
			marshalPlayerBlockAction(io, &value[index])
		}
		pk.BlockActions = protocol.Option(value)
	} else if io.reading {
		pk.BlockActions = protocol.Optional[[]protocol.PlayerBlockAction]{}
	}
	if io.reading {
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
