package v844

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
	if got, want := len(states), 15845; got != want {
		t.Fatalf("block state count: got %d, want %d", got, want)
	}
	if got, want := len(items), 1888; got != want {
		t.Fatalf("item count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Groups), 120; got != want {
		t.Fatalf("creative group count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Items), 1810; got != want {
		t.Fatalf("creative item count: got %d, want %d", got, want)
	}
	if len(crafting.Shaped)+len(crafting.Shapeless)+len(crafting.UserDataShapeless)+len(crafting.Multi) == 0 {
		t.Fatal("decoded crafting registry is empty")
	}
	if len(potions.Potions)+len(potions.ContainerChanges) == 0 {
		t.Fatal("decoded potion registry is empty")
	}
	if got, want := len(crafting.Shaped), 1084; got != want {
		t.Fatalf("shaped recipe count: got %d, want %d", got, want)
	}
	if got, want := len(crafting.Shapeless), 1121; got != want {
		t.Fatalf("shapeless recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.Potions), 210; got != want {
		t.Fatalf("potion recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.ContainerChanges), 2; got != want {
		t.Fatalf("container-change recipe count: got %d, want %d", got, want)
	}
	if got, want := len(furnaces), 251; got != want {
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
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "893c63520f286239badb720ecd5450f819f9d128a2fb8230b80e13c61c800626"; got != want {
		t.Fatalf("fallback report SHA256: got %s, want %s", got, want)
	}
	var report struct {
		BlockFallbacks []json.RawMessage `json:"block_fallbacks"`
		ItemFallbacks  []json.RawMessage `json:"item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got, want := len(report.BlockFallbacks), 1654; got != want {
		t.Fatalf("block fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.ItemFallbacks), 160; got != want {
		t.Fatalf("item fallback count: got %d, want %d", got, want)
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":   "6b9ecc2b7ddd3b9cd59b5e745d2a704b35e44496c67bcde570f3ea9bcfaed068",
		"vanilla_items.nbt":  "338a5ee40257b3665355635c357151e3cc41f3032168c90bb34dd5706e322082",
		"creative_items.nbt": "0508b029fe2e0f50894a48239f4e034f5ad2c3289a8028308270f3ff6af13bcb",
		"crafting_data.nbt":  "f1f4477ee9f5e43474c092595a22f023d1a40a040474dbfb74752dc3f05423ed",
		"furnace_data.nbt":   "0b3e7c1ebb3b656e2039631a4cb3be88a5099f73ce7f74a156885ce786dfb04e",
		"potion_data.nbt":    "faaac0150e1627f39d778c99d257b6c00ad56d08f15d138c8dd56c634b815ef7",
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
