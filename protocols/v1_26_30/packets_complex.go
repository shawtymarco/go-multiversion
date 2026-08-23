package v1_26_30

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	moveActorHasX = 1 << iota
	moveActorHasY
	moveActorHasZ
	moveActorHasRotX
	moveActorHasRotY
	moveActorHasRotYHead
	moveActorOnGround
	moveActorTeleport
	moveActorForceMove
)

func marshalMoveActorDelta(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MoveActorDelta)
	io.Varuint64(&pk.EntityRuntimeID)
	flags := uint16(0)
	if !io.reading {
		flags |= optionalFloatFlag(pk.PositionX, moveActorHasX)
		flags |= optionalFloatFlag(pk.PositionY, moveActorHasY)
		flags |= optionalFloatFlag(pk.PositionZ, moveActorHasZ)
		flags |= optionalFloatFlag(pk.RotationX, moveActorHasRotX)
		flags |= optionalFloatFlag(pk.RotationY, moveActorHasRotY)
		flags |= optionalFloatFlag(pk.RotationYHead, moveActorHasRotYHead)
		if pk.OnGround {
			flags |= moveActorOnGround
		}
		if pk.ForceCompletion {
			flags |= moveActorTeleport
		}
		if pk.ForceMove {
			flags |= moveActorForceMove
		}
	}
	io.Uint16(&flags)
	marshalFlaggedFloat(io, &pk.PositionX, flags&moveActorHasX != 0, io.Float32)
	marshalFlaggedFloat(io, &pk.PositionY, flags&moveActorHasY != 0, io.Float32)
	marshalFlaggedFloat(io, &pk.PositionZ, flags&moveActorHasZ != 0, io.Float32)
	marshalFlaggedFloat(io, &pk.RotationX, flags&moveActorHasRotX != 0, io.ByteFloat)
	marshalFlaggedFloat(io, &pk.RotationY, flags&moveActorHasRotY != 0, io.ByteFloat)
	marshalFlaggedFloat(io, &pk.RotationYHead, flags&moveActorHasRotYHead != 0, io.ByteFloat)
	if io.reading {
		pk.OnGround = flags&moveActorOnGround != 0
		pk.ForceCompletion = flags&moveActorTeleport != 0
		pk.ForceMove = flags&moveActorForceMove != 0
		pk.ForceMoveLocalEntity = false
	}
}

func optionalFloatFlag(value protocol.Optional[float32], flag uint16) uint16 {
	if _, ok := value.Value(); ok {
		return flag
	}
	return 0
}

func marshalFlaggedFloat(io *wireIO, value *protocol.Optional[float32], present bool, marshal func(*float32)) {
	if !present {
		if io.reading {
			*value = protocol.Optional[float32]{}
		}
		return
	}
	v, _ := value.Value()
	marshal(&v)
	*value = protocol.Option(v)
}

func marshalMovePlayer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MovePlayer)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Vec3(&pk.Position)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.Uint8(&pk.Mode)
	io.Bool(&pk.OnGround)
	io.Varuint64(&pk.RiddenEntityRuntimeID)
	if pk.Mode == packet.MoveModeTeleport {
		data, _ := pk.TeleportData.Value()
		io.Int32(&data.TeleportCause)
		io.Int32(&data.TeleportSourceEntityType)
		pk.TeleportData = protocol.Option(data)
	} else if io.reading {
		pk.TeleportData = protocol.Optional[protocol.TeleportData]{}
	}
	io.Varuint64(&pk.Tick)
}

func marshalPlayerList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerList)
	action := uint8(protocol.PlayerListActionAdd)
	if !io.reading && len(pk.Entries) != 0 {
		action = pk.Entries[0].ActionType
	}
	io.Uint8(&action)
	switch action {
	case protocol.PlayerListActionAdd:
		protocol.FuncIOSlice(io.directional(), &pk.Entries, marshalPlayerListEntry)
	case protocol.PlayerListActionRemove:
		protocol.FuncIOSlice(io.directional(), &pk.Entries, func(io protocol.IO, entry *protocol.PlayerListEntry) {
			io.UUID(&entry.UUID)
		})
	default:
		io.UnknownEnumOption(action, "player list action type")
		return
	}
	for i := range pk.Entries {
		pk.Entries[i].ActionType = action
	}
	if action == protocol.PlayerListActionAdd {
		for i := range pk.Entries {
			io.Bool(&pk.Entries[i].Skin.Trusted)
		}
	}
}

func marshalPlayerLocation(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerLocation)
	io.Int32(&pk.Type)
	io.Varint64(&pk.EntityUniqueID)
	if pk.Type == packet.PlayerLocationTypeCoordinates {
		io.Vec3(&pk.Position)
	} else if pk.Type != packet.PlayerLocationTypeHide {
		io.UnknownEnumOption(pk.Type, "player location type")
	}
}

func marshalPlayerUpdateEntityOverrides(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerUpdateEntityOverrides)
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint32(&pk.PropertyIndex)
	typeID := uint8(pk.Type)
	io.Uint8(&typeID)
	pk.Type = uint32(typeID)
	switch pk.Type {
	case packet.PlayerUpdateEntityOverridesTypeClearAll, packet.PlayerUpdateEntityOverridesTypeRemove:
	case packet.PlayerUpdateEntityOverridesTypeInt:
		io.Int32(&pk.IntValue)
	case packet.PlayerUpdateEntityOverridesTypeFloat:
		io.Float32(&pk.FloatValue)
	default:
		io.UnknownEnumOption(pk.Type, "entity override type")
	}
}

func marshalSetScore(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetScore)
	action := uint8(0)
	if !io.reading && len(pk.Entries) != 0 && pk.Entries[0].IdentityType == protocol.ScoreboardIdentityRemove {
		action = 1
	}
	io.Uint8(&action)
	if action == 1 {
		protocol.FuncIOSlice(io.directional(), &pk.Entries, func(io protocol.IO, entry *protocol.ScoreboardEntry) {
			io.Varint64(&entry.EntryID)
			io.String(&entry.ObjectiveName)
			io.Int32(&entry.Score)
			entry.IdentityType = protocol.ScoreboardIdentityRemove
		})
		return
	}
	if action != 0 {
		io.UnknownEnumOption(action, "set score action type")
		return
	}
	protocol.FuncIOSlice(io.directional(), &pk.Entries, func(io protocol.IO, entry *protocol.ScoreboardEntry) {
		io.Varint64(&entry.EntryID)
		io.String(&entry.ObjectiveName)
		io.Int32(&entry.Score)
		io.Uint8(&entry.IdentityType)
		switch entry.IdentityType {
		case protocol.ScoreboardIdentityPlayer, protocol.ScoreboardIdentityEntity:
			io.Varint64(&entry.EntityUniqueID)
		case protocol.ScoreboardIdentityFakePlayer:
			io.String(&entry.DisplayName)
		default:
			io.UnknownEnumOption(entry.IdentityType, "scoreboard entry identity type")
		}
	})
}

func marshalSetScoreboardIdentity(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetScoreboardIdentity)
	io.Uint8(&pk.ActionType)
	protocol.FuncIOSlice(io.directional(), &pk.Entries, func(io protocol.IO, entry *protocol.ScoreboardIdentityEntry) {
		io.Varint64(&entry.EntryID)
		if pk.ActionType == packet.ScoreboardIdentityActionRegister {
			entityID, _ := entry.EntityUniqueID.Value()
			io.Varint64(&entityID)
			entry.EntityUniqueID = protocol.Option(entityID)
		} else {
			entry.EntityUniqueID = protocol.Optional[int64]{}
		}
	})
}

func marshalLevelChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelChunk)
	io.ChunkPos(&pk.Position)
	io.Varint32(&pk.Dimension)
	subChunkCount := pk.SubChunkCount
	if !io.reading {
		if limit, ok := pk.SubChunkLimit.Value(); ok {
			if limit < 0 {
				subChunkCount = math.MaxUint32
			} else {
				subChunkCount = math.MaxUint32 - 1
			}
		}
	}
	io.Varuint32(&subChunkCount)
	if io.reading {
		pk.SubChunkCount = subChunkCount
	}
	if subChunkCount == math.MaxUint32-1 {
		limit, _ := pk.SubChunkLimit.Value()
		highest := uint16(limit)
		io.Uint16(&highest)
		pk.SubChunkLimit = protocol.Option(int32(highest))
	} else if subChunkCount == math.MaxUint32 {
		pk.SubChunkLimit = protocol.Option(int32(-1))
	} else if io.reading {
		pk.SubChunkLimit = protocol.Optional[int32]{}
	}
	io.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		protocol.FuncSlice(io.directional(), &pk.BlobHashes, io.Uint64)
	} else if io.reading {
		pk.BlobHashes = nil
	}
	io.ByteSlice(&pk.RawPayload)
}

func marshalSubChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunk)
	io.Bool(&pk.CacheEnabled)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Position[0])
	io.Varint32(&pk.Position[1])
	io.Varint32(&pk.Position[2])
	if pk.CacheEnabled {
		funcSliceUint32(io, &pk.SubChunkEntries, func(entry *protocol.SubChunkEntry) {
			marshalSubChunkEntry(io, entry, true)
		})
	} else {
		funcSliceUint32(io, &pk.SubChunkEntries, func(entry *protocol.SubChunkEntry) {
			marshalSubChunkEntry(io, entry, false)
		})
	}
}

func marshalPlayerAuthInput(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAuthInput)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Vec3(&pk.Position)
	io.Vec2(&pk.MoveVector)
	io.Float32(&pk.HeadYaw)
	marshalLegacyInputFlags(io, &pk.InputData, 65)
	io.Varuint32(&pk.InputMode)
	io.Varuint32(&pk.PlayMode)
	interactionModel := uint32(pk.InteractionModel)
	io.Varuint32(&interactionModel)
	pk.InteractionModel = int32(interactionModel)
	io.Float32(&pk.InteractPitch)
	io.Float32(&pk.InteractYaw)
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
		protocol.FuncIOSliceOfLen(io.directional(), uint32(count), &value, marshalPlayerBlockAction)
		pk.BlockActions = protocol.Option(value)
	} else if io.reading {
		pk.BlockActions = protocol.Optional[[]protocol.PlayerBlockAction]{}
	}
	if pk.InputData.Load(packet.InputFlagClientPredictedVehicle) {
		rotation, _ := pk.VehicleRotation.Value()
		io.Vec2(&rotation)
		pk.VehicleRotation = protocol.Option(rotation)
		vehicle, _ := pk.ClientPredictedVehicle.Value()
		io.Varint64(&vehicle)
		pk.ClientPredictedVehicle = protocol.Option(vehicle)
	} else if io.reading {
		pk.VehicleRotation = protocol.Optional[mgl32.Vec2]{}
		pk.ClientPredictedVehicle = protocol.Optional[int64]{}
	}
	io.Vec2(&pk.AnalogueMoveVector)
	io.Vec3(&pk.CameraOrientation)
	io.Vec2(&pk.RawMoveVector)
}

func marshalPlayerBlockAction(io protocol.IO, action *protocol.PlayerBlockAction) {
	io.Varint32(&action.Action)
	switch action.Action {
	case protocol.PlayerActionStartBreak, protocol.PlayerActionAbortBreak, protocol.PlayerActionCrackBreak,
		protocol.PlayerActionPredictDestroyBlock, protocol.PlayerActionContinueDestroyBlock:
		io.BlockPos(&action.BlockPos)
		io.Varint32(&action.Face)
	}
}

func marshalLegacyInputFlags(io *wireIO, flags *protocol.InputFlags, size int) {
	bits := protocol.NewBitset(size)
	if !io.reading {
		for i := 0; i < size && i < flags.Len(); i++ {
			if flags.Load(i) {
				bits.Set(i)
			}
		}
	}
	io.Bitset(&bits, size)
	if io.reading {
		*flags = protocol.NewInputFlags(size)
		for i := 0; i < size; i++ {
			if bits.Load(i) {
				flags.Set(i)
			}
		}
	}
}

func funcSliceUint32[T any](io *wireIO, values *[]T, marshal func(*T)) {
	count := uint32(len(*values))
	io.Uint32(&count)
	if io.reading {
		*values = make([]T, count)
	}
	for i := range *values {
		marshal(&(*values)[i])
	}
}
