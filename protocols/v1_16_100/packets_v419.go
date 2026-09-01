package v1_16_100

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalActorPickRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ActorPickRequest)
	io.Int64(&pk.EntityUniqueID)
	io.Uint8(&pk.HotBarSlot)
	if io.reading {
		pk.WithData = false
	}
}

func marshalHurtArmour(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.HurtArmour)
	io.Varint32(&pk.Cause)
	io.Varint32(&pk.Damage)
	if io.reading {
		pk.ArmourSlots = 0
	}
}

func marshalEvent419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Event)
	runtimeID := uint64(pk.EntityRuntimeID)
	io.Varuint64(&runtimeID)
	pk.EntityRuntimeID = int64(runtimeID)
	if !io.reading && pk.Event == nil {
		pk.Event = &protocol.AchievementAwardedEvent{}
	}
	io.EventType(&pk.Event)
	usePlayerID := uint8(0)
	if pk.UsePlayerID {
		usePlayerID = 1
	}
	io.Uint8(&usePlayerID)
	pk.UsePlayerID = usePlayerID != 0
}

func marshalNPCRequest419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.NPCRequest)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Uint8(&pk.RequestType)
	io.String(&pk.CommandString)
	io.Uint8(&pk.ActionType)
	if io.reading {
		pk.SceneName = ""
	}
}

func marshalPhotoTransfer419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PhotoTransfer)
	io.String(&pk.PhotoName)
	io.ByteSlice(&pk.PhotoData)
	io.String(&pk.BookID)
	if io.reading {
		pk.PhotoType, pk.SourceType, pk.OwnerEntityUniqueID, pk.NewPhotoName = 0, 0, 0, ""
	}
}

func marshalEducationSettings419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.EducationSettings)
	io.String(&pk.CodeBuilderDefaultURI)
	io.String(&pk.CodeBuilderTitle)
	io.Bool(&pk.CanResizeCodeBuilder)
	override, hasOverride := pk.OverrideURI.Value()
	io.Bool(&hasOverride)
	if hasOverride {
		io.String(&override)
		pk.OverrideURI = protocol.Option(override)
	} else if io.reading {
		pk.OverrideURI = protocol.Optional[string]{}
	}
	io.Bool(&pk.HasQuiz)
	if io.reading {
		pk.DisableLegacyTitleBar = false
		pk.PostProcessFilter = ""
		pk.ScreenshotBorderPath = ""
		pk.CanModifyBlocks = protocol.Optional[bool]{}
		pk.ExternalLinkSettings = protocol.Optional[protocol.EducationExternalLinkSettings]{}
	}
}

func marshalAnimateEntity419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AnimateEntity)
	io.String(&pk.Animation)
	io.String(&pk.NextState)
	io.String(&pk.StopCondition)
	io.String(&pk.Controller)
	io.Float32(&pk.BlendOutTime)
	protocol.FuncSlice(io.directional(), &pk.EntityRuntimeIDs, io.Varuint64)
	if io.reading {
		pk.StopConditionVersion = 0
	}
}

func marshalCameraShake419(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraShake)
	io.Float32(&pk.Intensity)
	io.Float32(&pk.Duration)
	io.Uint8(&pk.Type)
	if io.reading {
		pk.Action = packet.CameraShakeActionAdd
	}
}
