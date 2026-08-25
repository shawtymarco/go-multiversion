package v1_26_45

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBossEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BossEvent)
	io.ActorUniqueID(&pk.BossEntityUniqueID)
	var discardedPlayerUniqueID int64
	io.ActorUniqueID(&discardedPlayerUniqueID)
	io.Uint8(&pk.EventType)
	io.String(&pk.BossBarTitle)
	io.String(&pk.FilteredBossBarTitle)
	io.Float32(&pk.HealthPercentage)
	io.Uint8(&pk.Colour)
	io.Uint8(&pk.Overlay)
}

func marshalInventoryTransaction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventoryTransaction)
	io.Varint32(&pk.LegacyRequestID)
	hasLegacy := pk.LegacyRequestID < -1 && pk.LegacyRequestID&1 == 0
	io.Bool(&hasLegacy)
	if hasLegacy {
		protocol.Slice(io.directional(), &pk.LegacySetItemSlots)
	}
	hasType := true
	io.Bool(&hasType)
	if !hasType {
		io.InvalidValue(hasType, "inventory transaction type", "outer optional must be present")
		return
	}
	io.TransactionDataType(&pk.TransactionData)
	hasActions := true
	io.Bool(&hasActions)
	if !hasActions {
		io.InvalidValue(hasActions, "inventory transaction actions", "outer optional must be present")
		return
	}
	protocol.FuncIOSlice(io.directional(), &pk.Actions, func(raw protocol.IO, action *protocol.InventoryAction) {
		marshalInventoryAction(asWireIO(raw), action)
	})
	marshalInventoryTransactionData(io, pk.TransactionData)
}

func marshalMoveActorDelta(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MoveActorDelta)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	protocol.OptionalFunc(io, &pk.PositionX, io.Float32)
	protocol.OptionalFunc(io, &pk.PositionY, io.Float32)
	protocol.OptionalFunc(io, &pk.PositionZ, io.Float32)
	protocol.OptionalFunc(io, &pk.RotationX, io.ByteFloat)
	protocol.OptionalFunc(io, &pk.RotationY, io.ByteFloat)
	protocol.OptionalFunc(io, &pk.RotationYHead, io.ByteFloat)
	io.Bool(&pk.OnGround)
	io.Bool(&pk.ForceMove)
	io.Bool(&pk.ForceMoveLocalEntity)
	io.Bool(&pk.ForceCompletion)
	if io.reading {
		pk.Ticks = 0
	}
}

func marshalPlaySound(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlaySound)
	io.String(&pk.SoundName)
	io.SoundPos(&pk.Position)
	io.Float32(&pk.Volume)
	io.Float32(&pk.Pitch)
	io.Varint32(&pk.LoopCount)
	protocol.OptionalFunc(io, &pk.Handle, io.Uint64)
	if io.reading {
		pk.BypassListenerRangeCheck = false
		pk.PlaybackPositionSeconds = protocol.Optional[float32]{}
	}
}

func marshalPlayerAuthInput(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAuthInput)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Vec3(&pk.Position)
	io.Vec2(&pk.MoveVector)
	io.Float32(&pk.HeadYaw)
	marshalInputFlags(io, &pk.InputData, packet.InputFlagCount)
	io.Varuint32(&pk.InputMode)
	io.Varuint32(&pk.PlayMode)
	io.Varint32(&pk.InteractionModel)
	io.Float32(&pk.InteractPitch)
	io.Float32(&pk.InteractYaw)
	io.Varuint64(&pk.Tick)
	io.Vec3(&pk.Delta)
	doubleOptionalFunc(io, &pk.ItemInteractionData, io.PlayerInventoryAction)
	doubleOptionalFunc(io, &pk.ItemStackRequest, func(request *protocol.ItemStackRequest) {
		request.Marshal(io)
	})
	doubleOptionalFunc(io, &pk.BlockActions, func(actions *[]protocol.PlayerBlockAction) {
		protocol.Slice(io.directional(), actions)
	})
	doubleOptionalFunc(io, &pk.VehicleRotation, io.Vec2)
	doubleOptionalFunc(io, &pk.ClientPredictedVehicle, io.ActorUniqueID)
	io.Vec2(&pk.AnalogueMoveVector)
	io.Vec3(&pk.CameraOrientation)
	io.Vec2(&pk.RawMoveVector)
}

func marshalInputFlags(io *wireIO, flags *protocol.InputFlags, size int) {
	present := flags.Present()
	io.Bool(&present)
	if !present {
		if io.reading {
			*flags = protocol.NewInputFlags(size)
		}
		return
	}
	protocol.InputFlagList(io.directional(), flags, size)
}

func marshalCameraPresets(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraPresets)
	protocol.FuncIOSlice(io.directional(), &pk.Presets, func(raw protocol.IO, preset *protocol.CameraPreset) {
		marshalCameraPreset(asWireIO(raw), preset)
	})
}

func marshalDimensionData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.DimensionData)
	protocol.FuncIOSlice(io.directional(), &pk.Definitions, func(raw protocol.IO, definition *protocol.DimensionDefinition) {
		marshalDimensionDefinition(asWireIO(raw), definition)
	})
}

func marshalClientBoundAttributeLayerSync(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundAttributeLayerSync)
	io.Varuint32(&pk.PayloadType)
	switch pk.PayloadType {
	case protocol.AttributeLayerPayloadTypeUpdateLayers:
		protocol.FuncIOSlice(io.directional(), &pk.Layers, func(raw protocol.IO, layer *protocol.AttributeLayerData) {
			marshalAttributeLayer(asWireIO(raw), layer)
		})
	case protocol.AttributeLayerPayloadTypeUpdateSettings:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.Single(io.directional(), &pk.Settings)
	case protocol.AttributeLayerPayloadTypeUpdateEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncIOSlice(io.directional(), &pk.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
			marshalEnvironmentAttribute(asWireIO(raw), value)
		})
	case protocol.AttributeLayerPayloadTypeRemoveEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncSlice(io.directional(), &pk.RemoveAttributeNames, io.String)
	default:
		io.UnknownEnumOption(pk.PayloadType, "attribute layer payload type")
	}
}

func marshalServerBoundDiagnostics(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ServerBoundDiagnostics)
	io.Float32(&pk.AverageFramesPerSecond)
	io.Float32(&pk.AverageServerSimTickTime)
	io.Float32(&pk.AverageClientSimTickTime)
	io.Float32(&pk.AverageBeginFrameTime)
	io.Float32(&pk.AverageInputTime)
	io.Float32(&pk.AverageRenderTime)
	io.Float32(&pk.AverageEndFrameTime)
	io.Float32(&pk.AverageRemainderTimePercent)
	io.Float32(&pk.AverageUnaccountedTimePercent)
	protocol.FuncIOSlice(io.directional(), &pk.MemoryCategoryValues, func(raw protocol.IO, value *protocol.MemoryCategoryCounter) {
		marshalMemoryCategory(asWireIO(raw), value)
	})
	protocol.FuncIOSlice(io.directional(), &pk.EntityDiagnostics, func(raw protocol.IO, value *protocol.EntityDiagnosticTimingInfo) {
		marshalEntityDiagnostic(asWireIO(raw), value)
	})
	protocol.Slice(io.directional(), &pk.SystemDiagnostics)
	protocol.Slice(io.directional(), &pk.SystemCategories)
	protocol.Slice(io.directional(), &pk.WhiskerScopes)
}

func marshalSubChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunk)
	io.Bool(&pk.CacheEnabled)
	io.Varint32(&pk.Dimension)
	io.SubChunkPos(&pk.Position)
	protocol.FuncIOSlice(io.directional(), &pk.SubChunkEntries, func(raw protocol.IO, entry *protocol.SubChunkEntry) {
		marshalSubChunkEntry(asWireIO(raw), entry)
	})
}

func marshalItemStackResponse(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ItemStackResponse)
	protocol.FuncIOSlice(io.directional(), &pk.Responses, func(raw protocol.IO, response *protocol.ItemStackResponse) {
		marshalStackResponse(asWireIO(raw), response)
	})
}
