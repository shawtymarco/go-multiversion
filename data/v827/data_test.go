package v827

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
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
	furnaces, err := Furnaces()
	if err != nil {
		t.Fatal(err)
	}
	smithing, err := Smithing()
	if err != nil {
		t.Fatal(err)
	}

	if len(states) == 0 || len(items) == 0 || len(creative.Groups) == 0 || len(creative.Items) == 0 {
		t.Fatal("decoded registry contains an empty required collection")
	}
	if got, want := len(states), 15285; got != want {
		t.Fatalf("block state count: got %d, want %d", got, want)
	}
	if got, want := len(items), 1836; got != want {
		t.Fatalf("item count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Groups), 113; got != want {
		t.Fatalf("creative group count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Items), 1738; got != want {
		t.Fatalf("creative item count: got %d, want %d", got, want)
	}
	if len(crafting.Shaped)+len(crafting.Shapeless)+len(crafting.UserDataShapeless)+len(crafting.Multi) == 0 {
		t.Fatal("decoded crafting registry is empty")
	}
	if len(potions.Potions)+len(potions.ContainerChanges) == 0 {
		t.Fatal("decoded potion registry is empty")
	}
	if got, want := len(crafting.Shaped), 1056; got != want {
		t.Fatalf("shaped recipe count: got %d, want %d", got, want)
	}
	if got, want := len(crafting.Shapeless), 1097; got != want {
		t.Fatalf("shapeless recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.Potions), 210; got != want {
		t.Fatalf("potion recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.ContainerChanges), 2; got != want {
		t.Fatalf("container-change recipe count: got %d, want %d", got, want)
	}
	if got, want := len(furnaces), 231; got != want {
		t.Fatalf("furnace recipe count: got %d, want %d", got, want)
	}
	if got, want := len(smithing), 9; got != want {
		t.Fatalf("smithing recipe count: got %d, want %d", got, want)
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

func TestFallbackReport(t *testing.T) {
	data, err := os.ReadFile("fallbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "5ec7c1e05fb5d7acca08712bf8af6ffea2adcdb9c19922eb025657fc3c01862d"; got != want {
		t.Fatalf("fallback report SHA256: got %s, want %s", got, want)
	}
	var report struct {
		BlockFallbacks []json.RawMessage `json:"block_fallbacks"`
		ItemFallbacks  []json.RawMessage `json:"item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got, want := len(report.BlockFallbacks), 2214; got != want {
		t.Fatalf("block fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.ItemFallbacks), 213; got != want {
		t.Fatalf("item fallback count: got %d, want %d", got, want)
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":   "7763aaf169ff47a5b8b09733b42e669210f0eb23f1246fa31ee848b4b695c60f",
		"vanilla_items.nbt":  "10afda9a8342a4c0a3d02444335a19eb6fd4a8237f2d1c9445780009dd01b94e",
		"creative_items.nbt": "319a510d355aaa7254a3710791490d25ca260c1f7e36a23217649bf549e66011",
		"crafting_data.nbt":  "b9f29394d804faf40773fc0e000978c8ec09473d8837d0e016272efcbe29effd",
		"furnace_data.nbt":   "54947309384374360c5a1ee355c6513d0dedfc6d2b77a15bad75a2f77b13a21b",
		"potion_data.nbt":    "9ff746236b663858f7ce60ab43a31ea7f271f9aef37647dc62729327b058f8f0",
		"smithing_data.nbt":  "99c9eb3ea54fca8a43c880fa91e6f1eb0ca350a051aad7e5ba1621659e3e4f7b",
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
