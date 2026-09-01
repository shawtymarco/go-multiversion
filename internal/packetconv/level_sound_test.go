package packetconv

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestMapLevelSoundBlockRuntimeID(t *testing.T) {
	blockSounds := []string{
		packet.SoundEventDoorOpen,
		packet.SoundEventDoorClose,
		packet.SoundEventTrapdoorOpen,
		packet.SoundEventTrapdoorClose,
		packet.SoundEventFenceGateOpen,
		packet.SoundEventFenceGateClose,
		packet.SoundEventPlace,
		packet.SoundEventHit,
		packet.SoundEventItemUseOn,
	}
	for _, soundType := range blockSounds {
		t.Run(soundType, func(t *testing.T) {
			pk := &packet.LevelSoundEvent{SoundType: soundType, ExtraData: 41}
			if !MapLevelSoundBlockRuntimeID(pk, func(runtimeID uint32) (uint32, bool) {
				if runtimeID != 41 {
					t.Fatalf("mapper received runtime ID %d, want 41", runtimeID)
				}
				return 9, true
			}) {
				t.Fatal("mapping unexpectedly failed")
			}
			if pk.ExtraData != 9 {
				t.Fatalf("mapped extra data = %d, want 9", pk.ExtraData)
			}
		})
	}
}

func TestMapLevelSoundBlockRuntimeIDPreservesOtherPayloads(t *testing.T) {
	for _, test := range []struct {
		name string
		pk   packet.LevelSoundEvent
	}{
		{name: "note payload", pk: packet.LevelSoundEvent{SoundType: packet.SoundEventNote, ExtraData: 0x1234}},
		{name: "negative sentinel", pk: packet.LevelSoundEvent{SoundType: packet.SoundEventPlace, ExtraData: -1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			called := false
			if !MapLevelSoundBlockRuntimeID(&test.pk, func(uint32) (uint32, bool) {
				called = true
				return 0, true
			}) {
				t.Fatal("mapping unexpectedly failed")
			}
			if called {
				t.Fatal("runtime ID mapper was called for a non-block payload")
			}
		})
	}
}

func TestMapLevelSoundBlockRuntimeIDRejectsUnknownTarget(t *testing.T) {
	pk := &packet.LevelSoundEvent{SoundType: packet.SoundEventPlace, ExtraData: 41}
	if MapLevelSoundBlockRuntimeID(pk, func(uint32) (uint32, bool) { return 0, false }) {
		t.Fatal("mapping unexpectedly accepted an unknown runtime ID")
	}
	if pk.ExtraData != 41 {
		t.Fatalf("failed mapping changed extra data to %d", pk.ExtraData)
	}
}
