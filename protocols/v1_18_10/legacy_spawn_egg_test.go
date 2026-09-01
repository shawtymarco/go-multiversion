package v1_18_10

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func TestIronGolemSpawnEggUsesLegacyVariant(t *testing.T) {
	items, err := mapping.NewItemMapperAllowingTargetOnly(
		[]protocol.ItemEntry{{Name: "minecraft:iron_golem_spawn_egg", RuntimeID: 510}},
		map[string]mapping.TargetItem{"minecraft:spawn_egg": {RuntimeID: 383}},
		func(name string) string { return name },
	)
	if err != nil {
		t.Fatal(err)
	}
	native := protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 510}, Count: 1}
	target, ok := mapItemStack(native, items, nil, toTarget)
	if !ok || target.NetworkID != 383 || target.MetadataValue != 20 {
		t.Fatalf("target stack = %+v/%v", target, ok)
	}
	roundTrip, ok := mapItemStack(target, items, nil, toNative)
	if !ok || roundTrip.NetworkID != 510 || roundTrip.MetadataValue != 0 {
		t.Fatalf("round-trip stack = %+v/%v", roundTrip, ok)
	}
}
