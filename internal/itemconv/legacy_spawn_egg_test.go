package itemconv

import (
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func TestIronGolemSpawnEggRoundTrip(t *testing.T) {
	items, err := mapping.NewItemMapperAllowingTargetOnly(
		[]protocol.ItemEntry{{Name: ironGolemSpawnEggIdentifier, RuntimeID: 510}},
		map[string]mapping.TargetItem{legacySpawnEggIdentifier: {RuntimeID: 383}},
		func(name string) string { return name },
	)
	if err != nil {
		t.Fatal(err)
	}
	native := protocol.ItemStack{
		ItemType: protocol.ItemType{NetworkID: 510}, Count: 1,
		NBTData: map[string]any{"display": map[string]any{"Name": "Dream Defender"}},
	}
	target, ok := DowngradeLegacySpawnEgg(native, items)
	if !ok || target.NetworkID != 383 || target.MetadataValue != 20 || target.Count != native.Count || !reflect.DeepEqual(target.NBTData, native.NBTData) {
		t.Fatalf("downgrade = %+v/%v", target, ok)
	}
	if native.NetworkID != 510 || native.MetadataValue != 0 {
		t.Fatalf("downgrade mutated input: %+v", native)
	}
	roundTrip, ok := UpgradeLegacySpawnEgg(target, items)
	if !ok || roundTrip.NetworkID != 510 || roundTrip.MetadataValue != 0 || !reflect.DeepEqual(roundTrip.NBTData, native.NBTData) {
		t.Fatalf("upgrade = %+v/%v", roundTrip, ok)
	}
}

func TestUnsupportedLegacySpawnEggMetadataIsRejected(t *testing.T) {
	items, err := mapping.NewItemMapperAllowingTargetOnly(
		[]protocol.ItemEntry{{Name: ironGolemSpawnEggIdentifier, RuntimeID: 510}},
		map[string]mapping.TargetItem{legacySpawnEggIdentifier: {RuntimeID: 383}},
		func(name string) string { return name },
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := UpgradeLegacySpawnEgg(protocol.ItemStack{
		ItemType: protocol.ItemType{NetworkID: 383, MetadataValue: 19}, Count: 1,
	}, items); ok {
		t.Fatal("unsupported spawn egg metadata was accepted")
	}
}
