package v1_21_100

import (
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalResourcePackStack(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePackStack)
	io.Bool(&pk.TexturePackRequired)
	var behaviourPacks []protocol.StackResourcePack
	protocol.Slice(io.directional(), &behaviourPacks)
	protocol.Slice(io.directional(), &pk.TexturePacks)
	io.String(&pk.BaseGameVersion)
	protocol.SliceUint32Length(io.directional(), &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	io.Bool(&pk.IncludeEditorPacks)
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
	io.Varuint64(&pk.Tick)
	if io.reading {
		pk.Ambient = false
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

func marshalAnimate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Animate)
	actionType := int32(pk.ActionType)
	io.Varint32(&actionType)
	pk.ActionType = uint8(actionType)
	io.Varuint64(&pk.EntityRuntimeID)
	if actionType&0x80 != 0 {
		io.Float32(&pk.Data)
	} else if io.reading {
		pk.Data = 0
	}
	if io.reading {
		pk.SwingSource = 0
	}
}

func marshalBossEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BossEvent)
	io.Varint64(&pk.BossEntityUniqueID)
	eventType := uint32(pk.EventType)
	io.Varuint32(&eventType)
	pk.EventType = uint8(eventType)
	var screenDarkening uint16
	switch eventType {
	case packet.BossEventShow:
		io.String(&pk.BossBarTitle)
		io.String(&pk.FilteredBossBarTitle)
		io.Float32(&pk.HealthPercentage)
		io.Uint16(&screenDarkening)
		marshalLegacyBossAppearance(io, pk)
	case packet.BossEventRegisterPlayer, packet.BossEventUnregisterPlayer, packet.BossEventRequest:
		io.Varint64(&pk.PlayerUniqueID)
	case packet.BossEventHide:
	case packet.BossEventHealthPercentage:
		io.Float32(&pk.HealthPercentage)
	case packet.BossEventTitle:
		io.String(&pk.BossBarTitle)
		io.String(&pk.FilteredBossBarTitle)
	case packet.BossEventAppearanceProperties:
		io.Uint16(&screenDarkening)
		marshalLegacyBossAppearance(io, pk)
	case packet.BossEventTexture:
		marshalLegacyBossAppearance(io, pk)
	default:
		io.UnknownEnumOption(eventType, "boss event type")
	}
}

func marshalLegacyBossAppearance(io *wireIO, pk *packet.BossEvent) {
	colour, overlay := uint32(pk.Colour), uint32(pk.Overlay)
	if !io.reading && colour == packet.BossEventColourWhite {
		colour--
	}
	io.Varuint32(&colour)
	io.Varuint32(&overlay)
	if io.reading {
		if colour == 6 {
			colour = packet.BossEventColourWhite
		}
		pk.Colour, pk.Overlay = uint8(colour), uint8(overlay)
	}
}

func marshalCommandRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandRequest)
	io.String(&pk.CommandLine)
	marshalCommandOrigin827(io, &pk.CommandOrigin)
	io.Bool(&pk.Internal)
	version, err := strconv.ParseInt(pk.Version, 10, 32)
	if err != nil {
		version = 0
	}
	legacyVersion := int32(version)
	io.Varint32(&legacyVersion)
	if io.reading {
		pk.Version = strconv.FormatInt(int64(legacyVersion), 10)
	}
}

func marshalCommandOutput(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandOutput)
	marshalCommandOrigin827(io, &pk.CommandOrigin)
	io.Uint8(&pk.OutputType)
	io.Varuint32(&pk.SuccessCount)
	protocol.FuncIOSlice(io.directional(), &pk.OutputMessages, marshalCommandOutputMessage827)
	if pk.OutputType == packet.CommandOutputTypeDataSet {
		data, _ := pk.DataSet.Value()
		io.String(&data)
		pk.DataSet = protocol.Option(data)
	} else if io.reading {
		pk.DataSet = protocol.Optional[string]{}
	}
}

func marshalCommandOrigin827(io *wireIO, origin *protocol.CommandOrigin) {
	io.Varuint32(&origin.Origin)
	io.UUID(&origin.UUID)
	io.String(&origin.RequestID)
	if origin.Origin == protocol.CommandOriginDevConsole || origin.Origin == protocol.CommandOriginTest {
		io.Varint64(&origin.PlayerUniqueID)
	} else if io.reading {
		origin.PlayerUniqueID = 0
	}
}

func marshalCommandOutputMessage827(raw protocol.IO, message *protocol.CommandOutputMessage) {
	io := asWireIO(raw)
	io.Bool(&message.Success)
	io.String(&message.Message)
	protocol.FuncSlice(io.directional(), &message.Parameters, io.String)
}

func marshalShowStoreOffer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ShowStoreOffer)
	offerID := ""
	if !io.reading && pk.OfferID != uuid.Nil {
		offerID = pk.OfferID.String()
	}
	io.String(&offerID)
	if io.reading {
		pk.OfferID = uuid.Nil
		if offerID != "" {
			parsed, err := uuid.Parse(offerID)
			if err != nil {
				io.InvalidValue(offerID, "store offer ID", "invalid UUID")
				return
			}
			pk.OfferID = parsed
		}
	}
	io.Uint8(&pk.Type)
}

func marshalEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Event)
	io.Varint64(&pk.EntityRuntimeID)
	io.EventType(&pk.Event)
	usePlayerID := byte(0)
	if pk.UsePlayerID {
		usePlayerID = 1
	}
	io.Uint8(&usePlayerID)
	pk.UsePlayerID = usePlayerID != 0
	marshalEventData827(io, pk.Event)
}

func marshalEventData827(io *wireIO, event protocol.Event) {
	switch value := event.(type) {
	case *protocol.AchievementAwardedEvent:
		achievementID := int32(value.AchievementID)
		io.Varint32(&achievementID)
		value.AchievementID = uint8(achievementID)
	case *protocol.EntityInteractEvent:
		io.Varint64(&value.InteractedEntityID)
		interactionType := int32(value.InteractionType)
		io.Varint32(&interactionType)
		value.InteractionType = uint8(interactionType)
		io.Varint32(&value.InteractionEntityType)
		io.Varint32(&value.EntityVariant)
		io.Uint8(&value.EntityColour)
	case *protocol.CauldronUsedEvent:
		io.Varuint32(&value.Colour)
		potionID, fillLevel := int32(value.PotionID), int32(value.FillLevel)
		io.Varint32(&potionID)
		io.Varint32(&fillLevel)
		value.PotionID, value.FillLevel = int16(potionID), int16(fillLevel)
	case *protocol.CauldronInteractEvent:
		interactionType, itemID := int32(value.BlockInteractionType), int32(value.ItemID)
		io.Varint32(&interactionType)
		io.Varint32(&itemID)
		value.BlockInteractionType, value.ItemID = uint8(interactionType), int16(itemID)
	case *protocol.ComposterInteractEvent:
		interactionType, itemID := int32(value.BlockInteractionType), int32(value.ItemID)
		io.Varint32(&interactionType)
		io.Varint32(&itemID)
		value.BlockInteractionType, value.ItemID = uint8(interactionType), int16(itemID)
	case *protocol.BellUsedEvent:
		itemID := int32(value.ItemID)
		io.Varint32(&itemID)
		value.ItemID = int16(itemID)
	default:
		event.Marshal(io.directional())
	}
}

func marshalText(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Text)
	io.Uint8(&pk.TextType)
	io.Bool(&pk.NeedsTranslation)
	switch pk.TextType {
	case packet.TextTypeChat, packet.TextTypeWhisper, packet.TextTypeAnnouncement:
		io.String(&pk.SourceName)
		io.String(&pk.Message)
	case packet.TextTypeRaw, packet.TextTypeTip, packet.TextTypeSystem, packet.TextTypeObject, packet.TextTypeObjectWhisper, packet.TextTypeObjectAnnouncement:
		io.String(&pk.Message)
	case packet.TextTypeTranslation, packet.TextTypePopup, packet.TextTypeJukeboxPopup:
		io.String(&pk.Message)
		protocol.FuncSlice(io.directional(), &pk.Parameters, io.String)
	}
	io.String(&pk.XUID)
	io.String(&pk.PlatformChatID)
	filtered, _ := pk.FilteredMessage.Value()
	io.String(&filtered)
	if io.reading {
		if filtered == "" {
			pk.FilteredMessage = protocol.Optional[string]{}
		} else {
			pk.FilteredMessage = protocol.Option(filtered)
		}
	}
}

func marshalClientBoundDebugRenderer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundDebugRenderer)
	legacyType := pk.Type + 1
	io.Uint32(&legacyType)
	if io.reading {
		if legacyType == 0 || legacyType > 2 {
			io.UnknownEnumOption(legacyType, "client bound debug renderer type")
			return
		}
		pk.Type = legacyType - 1
	}
	if legacyType == 2 {
		io.String(&pk.Text)
		io.Vec3(&pk.Position)
		io.Float32(&pk.Red)
		io.Float32(&pk.Green)
		io.Float32(&pk.Blue)
		io.Float32(&pk.Alpha)
		io.Uint64(&pk.Duration)
	}
}

func marshalUpdateClientInputLocks(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateClientInputLocks)
	io.Varuint32(&pk.Locks)
	var position mgl32.Vec3
	io.Vec3(&position)
}

func marshalCameraInstruction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraInstruction)
	protocol.OptionalMarshaler(io.directional(), &pk.Set)
	protocol.OptionalFunc(io.directional(), &pk.Clear, io.Bool)
	protocol.OptionalMarshaler(io.directional(), &pk.Fade)
	protocol.OptionalMarshaler(io.directional(), &pk.Target)
	protocol.OptionalFunc(io.directional(), &pk.RemoveTarget, io.Bool)
	protocol.OptionalFunc(io.directional(), &pk.FieldOfView, func(fieldOfView *protocol.CameraInstructionFieldOfView) {
		io.Float32(&fieldOfView.FieldOfView)
		io.Float32(&fieldOfView.EaseTime)
		easeType := uint8(fieldOfView.EaseType)
		io.Uint8(&easeType)
		fieldOfView.EaseType = int32(easeType)
		io.Bool(&fieldOfView.Clear)
	})
	if io.reading {
		pk.Spline = protocol.Optional[protocol.CameraSplineInstruction]{}
		pk.AttachToEntity = protocol.Optional[int64]{}
		pk.DetachFromEntity = protocol.Optional[bool]{}
	}
}

func marshalCameraAimAssistPresets(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraAimAssistPresets)
	protocol.FuncIOSlice(io.directional(), &pk.Categories, func(raw protocol.IO, category *protocol.CameraAimAssistCategory) {
		legacy := asWireIO(raw)
		legacy.String(&category.Name)
		marshalCameraAimAssistPriorities827(legacy, &category.Priorities)
	})
	protocol.FuncIOSlice(io.directional(), &pk.Presets, marshalCameraAimAssistPreset827)
	io.Uint8(&pk.Operation)
}

func marshalCameraAimAssistPriorities827(io *wireIO, priorities *protocol.CameraAimAssistPriorities) {
	protocol.Slice(io.directional(), &priorities.Entities)
	protocol.Slice(io.directional(), &priorities.Blocks)
	protocol.OptionalFunc(io.directional(), &priorities.EntityDefault, io.Int32)
	protocol.OptionalFunc(io.directional(), &priorities.BlockDefault, io.Int32)
	if io.reading {
		priorities.BlockTags = nil
		priorities.EntityTypeFamilies = nil
	}
}

func marshalCameraAimAssistPreset827(raw protocol.IO, preset *protocol.CameraAimAssistPreset) {
	io := asWireIO(raw)
	io.String(&preset.Identifier)
	protocol.FuncSlice(io.directional(), &preset.BlockExclusions, io.String)
	protocol.FuncSlice(io.directional(), &preset.LiquidTargets, io.String)
	protocol.Slice(io.directional(), &preset.ItemSettings)
	protocol.OptionalFunc(io.directional(), &preset.DefaultItemSettings, io.String)
	protocol.OptionalFunc(io.directional(), &preset.HandSettings, io.String)
	if io.reading {
		preset.EntityExclusions = nil
		preset.BlockTagExclusions = nil
		preset.EntityTypeFamilyExclusions = nil
	}
}

func marshalUpdateClientOptions(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateClientOptions)
	protocol.OptionalFunc(io.directional(), &pk.GraphicsMode, io.Uint8)
	if io.reading {
		pk.FilterProfanity = protocol.Optional[bool]{}
	}
}

func marshalPlayerEnchantOptions(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerEnchantOptions)
	protocol.FuncIOSlice(io.directional(), &pk.Options, func(raw protocol.IO, option *protocol.EnchantmentOption) {
		legacy := asWireIO(raw)
		cost := uint32(option.Cost)
		legacy.Varuint32(&cost)
		option.Cost = uint8(cost)
		protocol.Single(legacy.directional(), &option.Enchantments)
		legacy.String(&option.Name)
		legacy.Varuint32(&option.RecipeNetworkID)
	})
}
