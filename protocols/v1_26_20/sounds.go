package v1_26_20

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalLevelSoundEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelSoundEvent)
	soundID := legacySoundToID[packet.SoundEventUndefined]
	if io.reading {
		io.Varuint32(&soundID)
		sound, ok := legacySoundFromID[soundID]
		if !ok {
			io.UnknownEnumOption(soundID, "level sound event")
			return
		}
		pk.SoundType = sound
	} else {
		if mapped, ok := legacySoundToID[pk.SoundType]; ok {
			soundID = mapped
		}
		io.Varuint32(&soundID)
	}
	io.Vec3(&pk.Position)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityType)
	io.Bool(&pk.BabyMob)
	io.Bool(&pk.DisableRelativeVolume)
	io.Int64(&pk.EntityUniqueID)
	protocol.OptionalFunc(io.directional(), &pk.FireAtPosition, io.Vec3)
}
