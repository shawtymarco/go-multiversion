package v1_18_10

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalLevelSoundEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelSoundEvent)
	soundID := uint32(0)
	if pk.SoundType != "" {
		var ok bool
		soundID, ok = legacySoundEvents[pk.SoundType]
		if !ok {
			io.InvalidValue(pk.SoundType, "level sound event", "sound is absent from protocol 486")
			return
		}
	}
	io.Varuint32(&soundID)
	if io.reading {
		name, ok := legacySoundNames[soundID]
		if !ok {
			io.UnknownEnumOption(soundID, "level sound event")
			return
		}
		pk.SoundType = name
	}
	io.Vec3(&pk.Position)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityType)
	io.Bool(&pk.BabyMob)
	io.Bool(&pk.DisableRelativeVolume)
	if io.reading {
		pk.EntityUniqueID = 0
		pk.FireAtPosition = protocol.Optional[mgl32.Vec3]{}
	}
}
