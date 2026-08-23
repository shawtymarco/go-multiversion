package v1001

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestSnapshotsDecode(t *testing.T) {
	states, err := BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	items, err := Items()
	if err != nil {
		t.Fatal(err)
	}
	creative, err := Creative()
	if err != nil {
		t.Fatal(err)
	}
	crafting, err := Crafting()
	if err != nil {
		t.Fatal(err)
	}
	potions, err := Potions()
	if err != nil {
		t.Fatal(err)
	}

	if len(states) == 0 || len(items) == 0 || len(creative.Groups) == 0 || len(creative.Items) == 0 {
		t.Fatal("decoded registry contains an empty required collection")
	}
	if got, want := len(states), 16913; got != want {
		t.Fatalf("block state count: got %d, want %d", got, want)
	}
	if got, want := len(items), 1933; got != want {
		t.Fatalf("item count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Groups), 123; got != want {
		t.Fatalf("creative group count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Items), 1875; got != want {
		t.Fatalf("creative item count: got %d, want %d", got, want)
	}
	if len(crafting.Shaped)+len(crafting.Shapeless)+len(crafting.UserDataShapeless)+len(crafting.Multi) == 0 {
		t.Fatal("decoded crafting registry is empty")
	}
	if len(potions.Potions)+len(potions.ContainerChanges) == 0 {
		t.Fatal("decoded potion registry is empty")
	}
	if got, want := len(crafting.Shaped), 1118; got != want {
		t.Fatalf("shaped recipe count: got %d, want %d", got, want)
	}
	if got, want := len(crafting.Shapeless), 1431; got != want {
		t.Fatalf("shapeless recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.Potions), 210; got != want {
		t.Fatalf("potion recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.ContainerChanges), 2; got != want {
		t.Fatalf("container-change recipe count: got %d, want %d", got, want)
	}
	if _, ok := items["minecraft:air"]; !ok {
		t.Fatal("item registry does not contain minecraft:air")
	}
	foundAir := false
	for _, state := range states {
		if state.Name == "" {
			t.Fatal("block registry contains an empty identifier")
		}
		for name, value := range state.Properties {
			switch value.(type) {
			case uint8, int32, string:
			default:
				t.Fatalf("block state %s property %s has unsupported NBT type %T", state.Name, name, value)
			}
		}
		if state.Name == "minecraft:air" {
			foundAir = true
			break
		}
	}
	if !foundAir {
		t.Fatal("block registry does not contain minecraft:air")
	}
	for index, entry := range creative.Items {
		if entry.GroupIndex < 0 || entry.GroupIndex >= int32(len(creative.Groups)) {
			t.Fatalf("creative item %d has invalid group index %d", index, entry.GroupIndex)
		}
	}

}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":   "5decc3d672adda62b872ab57a4df525d0c75f1ac53ca4e9e5f710283b1cf4de4",
		"vanilla_items.nbt":  "0ad3c87f3cc76ad97d4742a24d77ceaf1f8652ed35db2b50937c2f95ff924dec",
		"creative_items.nbt": "87ab27f1ba243ae49c649d4e7c0807035669a4ae05783fd5d5774294ef3b858d",
		"crafting_data.nbt":  "d2bb26f2fc9e6e88168862973019f01fff4255bde552ff7c74b65c1dc62fb6f8",
		"potion_data.nbt":    "54cb7f879def3f30f6c969555a7911878b10d5dcdbcb48a67e918459c395e10a",
	}
	for name, expected := range want {
		data, ok := RawSnapshot(name)
		if !ok {
			t.Fatalf("snapshot %q is not embedded", name)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != expected {
			t.Fatalf("snapshot %q SHA256: got %s, want %s", name, got, expected)
		}
	}
}
