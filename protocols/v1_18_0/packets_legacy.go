package v1_18_0

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalResourcePackClientResponse(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePackClientResponse)
	response := byte(pk.Response + 1)
	io.Uint8(&response)
	if io.reading {
		if response == 0 {
			io.InvalidValue(response, "resource pack response", "must be between 1 and 4")
			return
		}
		pk.Response = uint32(response - 1)
	}
	funcSliceUint16(io, &pk.PacksToDownload, func(value *string) { io.String(value) })
}

func funcSliceUint16[T any](io *wireIO, values *[]T, marshal func(*T)) {
	count := uint16(len(*values))
	io.Uint16(&count)
	if io.reading {
		*values = make([]T, count)
	}
	for index := range *values {
		marshal(&(*values)[index])
	}
}

func marshalBlockActorData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BlockActorData)
	io.UBlockPos(&pk.Position)
	if io.reading && pk.NBTData == nil {
		pk.NBTData = make(map[string]any)
	}
	io.NBT(&pk.NBTData, nbt.NetworkLittleEndian)
}

func marshalContainerOpen(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ContainerOpen)
	io.Uint8(&pk.WindowID)
	io.Uint8(&pk.ContainerType)
	io.UBlockPos(&pk.ContainerPosition)
	io.Varint64(&pk.ContainerEntityUniqueID)
}

func marshalActorEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ActorEvent)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Uint8(&pk.EventType)
	io.Varint32(&pk.EventData)
	if io.reading {
		pk.FireAtPosition = protocol.Optional[mgl32.Vec3]{}
	}
}

func marshalMobEffect(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MobEffect)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Uint8(&pk.Operation)
	io.Varint32(&pk.EffectType)
	io.Varint32(&pk.Amplifier)
	io.Bool(&pk.Particles)
	io.Varint32(&pk.Duration)
	if io.reading {
		pk.Tick = 0
		pk.Ambient = false
	}
}

func marshalMobEquipment(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MobEquipment)
	io.Varuint64(&pk.EntityRuntimeID)
	io.ItemInstance(&pk.NewItem)
	io.Uint8(&pk.InventorySlot)
	io.Uint8(&pk.HotBarSlot)
	io.Uint8(&pk.WindowID)
}

func marshalMobArmourEquipment(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MobArmourEquipment)
	io.Varuint64(&pk.EntityRuntimeID)
	io.ItemInstance(&pk.Helmet)
	io.ItemInstance(&pk.Chestplate)
	io.ItemInstance(&pk.Leggings)
	io.ItemInstance(&pk.Boots)
	if io.reading {
		pk.Body = protocol.ItemInstance{}
	}
}

func marshalInteract(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Interact)
	io.Uint8(&pk.ActionType)
	io.Varuint64(&pk.TargetEntityRuntimeID)
	if pk.ActionType == packet.InteractActionMouseOverEntity || pk.ActionType == packet.InteractActionLeaveVehicle {
		position, _ := pk.Position.Value()
		io.Vec3(&position)
		pk.Position = protocol.Option(position)
	} else if io.reading {
		pk.Position = protocol.Optional[mgl32.Vec3]{}
	}
}

func marshalSetActorMotion(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetActorMotion)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Vec3(&pk.Velocity)
	if io.reading {
		pk.Tick = 0
	}
}

func marshalBossEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BossEvent)
	io.Varint64(&pk.BossEntityUniqueID)
	eventType := uint32(pk.EventType)
	io.Varuint32(&eventType)
	pk.EventType = uint8(eventType)
	marshalBossEventBody(io, pk, eventType)
	if io.reading {
		pk.FilteredBossBarTitle = ""
	}
}

func marshalBossEventBody(io *wireIO, pk *packet.BossEvent, eventType uint32) {
	marshalAppearance := func(withScreen bool) {
		if withScreen {
			var screenDarkening int16
			io.Int16(&screenDarkening)
		}
		colour, overlay := uint32(pk.Colour), uint32(pk.Overlay)
		io.Varuint32(&colour)
		io.Varuint32(&overlay)
		pk.Colour, pk.Overlay = uint8(colour), uint8(overlay)
	}
	switch eventType {
	case packet.BossEventShow:
		io.String(&pk.BossBarTitle)
		io.Float32(&pk.HealthPercentage)
		marshalAppearance(true)
	case packet.BossEventRegisterPlayer, packet.BossEventUnregisterPlayer:
		io.Varint64(&pk.PlayerUniqueID)
	case packet.BossEventHide:
	case packet.BossEventHealthPercentage:
		io.Float32(&pk.HealthPercentage)
	case packet.BossEventTitle:
		io.String(&pk.BossBarTitle)
	case packet.BossEventAppearanceProperties:
		marshalAppearance(true)
	case packet.BossEventTexture:
		marshalAppearance(false)
	default:
		io.UnknownEnumOption(eventType, "boss event type")
	}
}

func marshalTransfer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Transfer)
	io.String(&pk.Address)
	io.Uint16(&pk.Port)
	if io.reading {
		pk.ReloadWorld = false
		pk.GatheringJoinInfo = protocol.Optional[protocol.GatheringJoinInfo]{}
	}
}

func marshalPlaySound(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlaySound)
	io.String(&pk.SoundName)
	io.SoundPos(&pk.Position)
	io.Float32(&pk.Volume)
	io.Float32(&pk.Pitch)
	if io.reading {
		pk.LoopCount = 0
		pk.Handle = protocol.Optional[uint64]{}
	}
}

func marshalShowStoreOffer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ShowStoreOffer)
	offerID := ""
	if pk.OfferID != uuid.Nil {
		offerID = pk.OfferID.String()
	}
	io.String(&offerID)
	showAll := pk.Type != packet.StoreOfferTypeMarketplace
	io.Bool(&showAll)
	if io.reading {
		if offerID == "" {
			pk.OfferID = uuid.Nil
		} else if parsed, err := uuid.Parse(offerID); err == nil {
			pk.OfferID = parsed
		} else {
			io.InvalidValue(offerID, "store offer UUID", err.Error())
		}
		pk.Type = packet.StoreOfferTypeMarketplace
		if showAll {
			pk.Type = packet.StoreOfferTypeDressingRoom
		}
	}
}

func marshalSpawnParticleEffect(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SpawnParticleEffect)
	io.Uint8(&pk.Dimension)
	io.Varint64(&pk.EntityUniqueID)
	io.Vec3(&pk.Position)
	io.String(&pk.ParticleName)
	if io.reading {
		pk.MoLangVariables = protocol.Optional[[]byte]{}
	}
}

func marshalNetworkChunkPublisherUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.NetworkChunkPublisherUpdate)
	io.BlockPos(&pk.Position)
	io.Varuint32(&pk.Radius)
	if io.reading {
		pk.SavedChunks = nil
	}
}

func marshalLecternUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LecternUpdate)
	io.Uint8(&pk.Page)
	io.Uint8(&pk.PageCount)
	io.BlockPos(&pk.Position)
	var dropBook bool
	io.Bool(&dropBook)
}

func marshalMapInfoRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MapInfoRequest)
	io.Varint64(&pk.MapID)
	if io.reading {
		pk.ClientPixels = nil
	}
}

func marshalMapCreateLockedCopy(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MapCreateLockedCopy)
	io.Varint64(&pk.OriginalMapID)
	io.Varint64(&pk.NewMapID)
}

func marshalOnScreenTextureAnimation(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.OnScreenTextureAnimation)
	io.Uint32(&pk.AnimationType)
}

func marshalNetworkSettings(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.NetworkSettings)
	io.Uint16(&pk.CompressionThreshold)
	if io.reading {
		pk.CompressionAlgorithm = packet.CompressionAlgorithmFlate
		pk.ClientThrottle = false
		pk.ClientThrottleThreshold = 0
		pk.ClientThrottleScalar = 0
	}
}

func marshalUpdatePlayerGameType(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdatePlayerGameType)
	io.Varint32(&pk.GameType)
	io.Varint64(&pk.PlayerUniqueID)
	if io.reading {
		pk.Tick = 0
	}
}

func marshalPositionTrackingDBServerBroadcast(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PositionTrackingDBServerBroadcast)
	io.Uint8(&pk.BroadcastAction)
	io.Varint32(&pk.TrackingID)
	if io.reading {
		var serialised []byte
		io.Bytes(&serialised)
		if len(serialised) != 0 {
			if err := nbt.Unmarshal(serialised, &pk.Payload); err != nil {
				io.InvalidValue(serialised, "position tracking payload", err.Error())
			}
		}
		return
	}
	if pk.Payload != nil {
		serialised, err := nbt.Marshal(pk.Payload)
		if err != nil {
			io.InvalidValue(pk.Payload, "position tracking payload", err.Error())
			return
		}
		io.Bytes(&serialised)
	}
}

func marshalCorrectPlayerMovePrediction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CorrectPlayerMovePrediction)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Delta)
	io.Bool(&pk.OnGround)
	io.Varuint64(&pk.Tick)
	if io.reading {
		pk.PredictionType = packet.PredictionTypePlayer
		pk.Rotation = mgl32.Vec2{}
		pk.VehicleAngularVelocity = protocol.Optional[float32]{}
	}
}

func marshalClientBoundDebugRenderer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundDebugRenderer)
	typeID := int32(pk.Type)
	io.Int32(&typeID)
	pk.Type = uint32(typeID)
	if typeID == 2 {
		io.String(&pk.Text)
		io.Vec3(&pk.Position)
		io.Float32(&pk.Red)
		io.Float32(&pk.Green)
		io.Float32(&pk.Blue)
		io.Float32(&pk.Alpha)
		duration := int64(pk.Duration)
		io.Int64(&duration)
		pk.Duration = uint64(duration)
	}
}

func marshalAddVolumeEntity(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddVolumeEntity)
	runtimeID := uint64(pk.EntityRuntimeID)
	io.Uint64(&runtimeID)
	pk.EntityRuntimeID = uint32(runtimeID)
	io.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	io.String(&pk.EngineVersion)
	if io.reading {
		pk.EncodingIdentifier = ""
		pk.InstanceIdentifier = ""
		pk.Bounds = [2]protocol.BlockPos{}
		pk.Dimension = packet.DimensionOverworld
	}
}

func marshalRemoveVolumeEntity(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.RemoveVolumeEntity)
	runtimeID := uint64(pk.EntityRuntimeID)
	io.Uint64(&runtimeID)
	pk.EntityRuntimeID = uint32(runtimeID)
	if io.reading {
		pk.Dimension = packet.DimensionOverworld
	}
}
