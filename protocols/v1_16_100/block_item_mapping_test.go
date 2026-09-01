package v1_16_100

import (
	"testing"

	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	v419data "github.com/shawtymarco/go-multiversion/data/v419"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type blockItemTestRegistry struct {
	states []mapping.BlockState
}

func (r blockItemTestRegistry) BlockCount() int    { return len(r.states) }
func (blockItemTestRegistry) AirRuntimeID() uint32 { return 0 }
func (r blockItemTestRegistry) RuntimeIDToState(runtimeID uint32) (string, map[string]any, bool) {
	if runtimeID >= uint32(len(r.states)) {
		return "", nil, false
	}
	state := r.states[runtimeID]
	return state.Name, state.Properties, true
}

func TestLegacyBlockItemUsesBlockVariantForNetworkID(t *testing.T) {
	blocks, items := legacyWoolTestMappers(t)
	targetBlockRuntimeID, valid, exact := blocks.MapNative(1)
	if !valid || !exact {
		t.Fatal("lime wool block did not map exactly")
	}
	targetState, ok := blocks.TargetState(targetBlockRuntimeID)
	if !ok {
		t.Fatal("lime wool target state is absent")
	}
	if metadata, ok := targetBlockItemMeta(targetState); !ok || metadata != 5 {
		t.Fatalf("lime wool target metadata = %d/%v, want 5/true", metadata, ok)
	}
	native := protocol.ItemStack{
		ItemType:       protocol.ItemType{NetworkID: 501},
		BlockRuntimeID: 1,
		Count:          64,
	}
	target, ok := mapItemStack(native, items, blocks, toTarget)
	if !ok || target.NetworkID != 35 || target.MetadataValue != 5 || target.BlockRuntimeID != 0 || target.Count != 64 {
		t.Fatalf("target wool stack = %#v, %v", target, ok)
	}
	legacyState, ok := targetBlockItemState("minecraft:wool", target.MetadataValue)
	if !ok {
		t.Fatal("lime wool metadata did not resolve a target state")
	}
	targetSemanticName, ok := items.TargetIdentifier(target.NetworkID)
	if !ok || targetSemanticName != "minecraft:white_wool" {
		t.Fatalf("lime wool target semantic identifier = %q/%v", targetSemanticName, ok)
	}
	if targetName, ok := items.TargetWireIdentifier(targetSemanticName); !ok || targetName != "minecraft:wool" {
		t.Fatalf("lime wool target wire identifier = %q/%v", targetName, ok)
	}
	legacyRID, ok := blocks.TargetRuntimeID(legacyState.Name, legacyState.Properties)
	if !ok {
		t.Fatalf("lime wool target state is absent from mapper: %#v", legacyState)
	}
	nativeRID, ok := blocks.TargetToNative(legacyRID)
	if !ok || nativeRID != 1 {
		t.Fatalf("lime wool target state maps to native RID %d/%v, want 1/true", nativeRID, ok)
	}
	nativeState, ok := blocks.NativeState(nativeRID)
	if !ok {
		t.Fatal("lime wool native state is absent")
	}
	if itemRID, ok := items.NativeRuntimeID(nativeState.Name); !ok || itemRID != 501 {
		t.Fatalf("lime wool native item = %d/%v, want 501/true", itemRID, ok)
	}
	roundTrip, ok := mapItemStack(target, items, blocks, toNative)
	if !ok || roundTrip.NetworkID != native.NetworkID || roundTrip.BlockRuntimeID != native.BlockRuntimeID || roundTrip.Count != native.Count {
		t.Fatalf("native wool round trip = %#v, %v", roundTrip, ok)
	}
}

func TestLegacyBlockSoundUsesTargetRuntimeID(t *testing.T) {
	blocks, items := legacyWoolTestMappers(t)
	p := Protocol{runtime: &runtimeData{blocks: blocks, items: items}}
	targetRuntimeID, valid, exact := blocks.MapNative(1)
	if !valid || !exact {
		t.Fatal("lime wool block did not map exactly")
	}
	for _, soundType := range []string{
		packet.SoundEventPlace,
		packet.SoundEventHit,
		packet.SoundEventItemUseOn,
	} {
		t.Run(soundType, func(t *testing.T) {
			original := &packet.LevelSoundEvent{SoundType: soundType, ExtraData: 1}
			converted := p.convertGameplayFromLatest(original, nil)
			if original.ExtraData != 1 {
				t.Fatalf("mapping mutated input extra data to %d", original.ExtraData)
			}
			target := converted[0].(*packet.LevelSoundEvent)
			if got, want := target.ExtraData, int32(targetRuntimeID); got != want {
				t.Fatalf("mapped sound block runtime ID: got %d, want %d", got, want)
			}
			native := p.convertGameplayToLatest(target, nil)
			if got := native[0].(*packet.LevelSoundEvent).ExtraData; got != 1 {
				t.Fatalf("round-trip sound block runtime ID: got %d, want 1", got)
			}
		})
	}
	if got := p.convertGameplayFromLatest(&packet.LevelSoundEvent{SoundType: packet.SoundEventDoorOpen}, nil); len(got) != 0 {
		t.Fatalf("unsupported door sound converted to %#v", got)
	}

	note := &packet.LevelSoundEvent{SoundType: packet.SoundEventNote, ExtraData: 0x1234}
	if got := p.convertGameplayFromLatest(note, nil)[0].(*packet.LevelSoundEvent).ExtraData; got != note.ExtraData {
		t.Fatalf("note payload changed to %d, want %d", got, note.ExtraData)
	}
}

func legacyWoolTestMappers(t *testing.T) (*mapping.BlockMapper, *mapping.ItemMapper) {
	t.Helper()
	historical, err := v419data.BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	var targetAir, targetWool mapping.BlockState
	var nativeAir, nativeWool mapping.BlockState
	for _, state := range historical {
		properties := cloneBlockItemTestProperties(state.Properties)
		upgraded := blockupgrader.Upgrade(blockupgrader.BlockState{Name: state.Name, Properties: cloneBlockItemTestProperties(properties), Version: state.Version})
		switch upgraded.Name {
		case "minecraft:air":
			if targetAir.Name == "" {
				targetAir = mapping.BlockState{Name: state.Name, Properties: properties, Version: state.Version}
				nativeAir = mapping.BlockState{Name: upgraded.Name, Properties: upgraded.Properties, Version: upgraded.Version}
			}
		case "minecraft:lime_wool":
			if state.Name == "minecraft:wool" {
				targetWool = mapping.BlockState{Name: state.Name, Properties: properties, Version: state.Version}
				nativeWool = mapping.BlockState{Name: upgraded.Name, Properties: upgraded.Properties, Version: upgraded.Version}
			}
		}
	}
	if targetAir.Name == "" || targetWool.Name == "" {
		t.Fatal("protocol 419 block snapshot has no air/lime wool states")
	}
	blocks, err := mapping.NewBlockMapperWithTargetOrder(blockItemTestRegistry{states: []mapping.BlockState{nativeAir, nativeWool}}, []mapping.BlockState{targetAir, targetWool})
	if err != nil {
		t.Fatal(err)
	}
	items, err := mapping.NewItemMapperAllowingTargetOnly(
		[]protocol.ItemEntry{{Name: "minecraft:white_wool", RuntimeID: 500}, {Name: "minecraft:lime_wool", RuntimeID: 501}},
		map[string]mapping.TargetItem{"minecraft:wool": {RuntimeID: 35}},
		func(string) string { return "minecraft:white_wool" },
	)
	if err != nil {
		t.Fatal(err)
	}
	return blocks, items
}

func cloneBlockItemTestProperties(properties map[string]any) map[string]any {
	cloned := make(map[string]any, len(properties))
	for key, value := range properties {
		cloned[key] = value
	}
	return cloned
}
