package v1_18_0

import (
	"testing"

	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	v475data "github.com/shawtymarco/go-multiversion/data/v475"
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
	native := protocol.ItemStack{
		ItemType:       protocol.ItemType{NetworkID: 501},
		BlockRuntimeID: 1,
		Count:          64,
	}
	target, ok := mapItemStack(native, items, blocks, toTarget)
	if !ok || target.NetworkID != 35 || target.BlockRuntimeID != int32(targetBlockRuntimeID) || target.Count != 64 {
		t.Fatalf("target wool stack = %#v, %v", target, ok)
	}
	roundTrip, ok := mapItemStack(target, items, blocks, toNative)
	if !ok || roundTrip.NetworkID != native.NetworkID || roundTrip.BlockRuntimeID != native.BlockRuntimeID || roundTrip.Count != native.Count {
		t.Fatalf("native wool round trip = %#v, %v", roundTrip, ok)
	}
}

func legacyWoolTestMappers(t *testing.T) (*mapping.BlockMapper, *mapping.ItemMapper) {
	t.Helper()
	historical, err := v475data.BlockStates()
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
		t.Fatal("protocol 475 block snapshot has no air/lime wool states")
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
