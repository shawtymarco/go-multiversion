package packetconv

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// MapLevelSoundBlockRuntimeID maps ExtraData only for level sounds whose
// payload is a block runtime ID. Other sounds pack unrelated values such as a
// note instrument and pitch into the same field and must remain untouched.
func MapLevelSoundBlockRuntimeID(pk *packet.LevelSoundEvent, mapRuntimeID func(uint32) (uint32, bool)) bool {
	if !levelSoundUsesBlockRuntimeID(pk.SoundType) || pk.ExtraData < 0 {
		return true
	}
	if mapRuntimeID == nil {
		return false
	}
	mapped, ok := mapRuntimeID(uint32(pk.ExtraData))
	if !ok {
		return false
	}
	pk.ExtraData = int32(mapped)
	return true
}

func levelSoundUsesBlockRuntimeID(soundType string) bool {
	switch soundType {
	case packet.SoundEventDoorOpen,
		packet.SoundEventDoorClose,
		packet.SoundEventTrapdoorOpen,
		packet.SoundEventTrapdoorClose,
		packet.SoundEventFenceGateOpen,
		packet.SoundEventFenceGateClose,
		packet.SoundEventPlace,
		packet.SoundEventHit,
		packet.SoundEventItemUseOn:
		return true
	default:
		return false
	}
}
