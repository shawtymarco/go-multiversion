package v1_21_130

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalLevelSoundEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelSoundEvent)
	var soundID uint32
	if io.reading {
		io.Varuint32(&soundID)
		sound, ok := legacySoundFromID[soundID]
		if !ok {
			io.UnknownEnumOption(soundID, "level sound event")
			return
		}
		pk.SoundType = sound
	} else {
		mapped, ok := legacySoundToID[pk.SoundType]
		if !ok {
			io.InvalidValue(pk.SoundType, "level sound event", "sound is absent from protocol 898")
			return
		}
		soundID = mapped
		io.Varuint32(&soundID)
	}
	io.Vec3(&pk.Position)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityType)
	io.Bool(&pk.BabyMob)
	io.Bool(&pk.DisableRelativeVolume)
	io.Int64(&pk.EntityUniqueID)
	if io.reading {
		pk.FireAtPosition = protocol.Optional[mgl32.Vec3]{}
	}
}
