package v898

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
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
	furnace, err := Furnace()
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
	if got, want := len(items), 1905; got != want {
		t.Fatalf("item count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Groups), 123; got != want {
		t.Fatalf("creative group count: got %d, want %d", got, want)
	}
	if got, want := len(creative.Items), 1843; got != want {
		t.Fatalf("creative item count: got %d, want %d", got, want)
	}
	if len(crafting.Shaped)+len(crafting.Shapeless)+len(crafting.UserDataShapeless)+len(crafting.Multi) == 0 {
		t.Fatal("decoded crafting registry is empty")
	}
	if len(potions.Potions)+len(potions.ContainerChanges) == 0 {
		t.Fatal("decoded potion registry is empty")
	}
	if got, want := len(crafting.Shaped), 1090; got != want {
		t.Fatalf("shaped recipe count: got %d, want %d", got, want)
	}
	if got, want := len(crafting.Shapeless), 1121; got != want {
		t.Fatalf("shapeless recipe count: got %d, want %d", got, want)
	}
	if got, want := len(potions.Potions), 210; got != want {
		t.Fatalf("potion recipe count: got %d, want %d", got, want)
	}
	if got, want := len(furnace), 263; got != want {
		t.Fatalf("furnace recipe count: got %d, want %d", got, want)
	}
	if got, want := len(smithing), 12; got != want {
		t.Fatalf("smithing recipe count: got %d, want %d", got, want)
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

func TestTargetRuntimeIDAnchors(t *testing.T) {
	states, err := BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	sort.SliceStable(states, func(i, j int) bool {
		left, right := fnv.New64(), fnv.New64()
		_, _ = left.Write([]byte(states[i].Name))
		_, _ = right.Write([]byte(states[j].Name))
		return states[i].Name != states[j].Name && left.Sum64() < right.Sum64()
	})
	for runtimeID, name := range map[int]string{
		2532:  "minecraft:stone",
		11062: "minecraft:grass_block",
		12530: "minecraft:air",
		13079: "minecraft:bedrock",
		14388: "minecraft:oak_planks",
	} {
		if got := states[runtimeID].Name; got != name {
			t.Fatalf("target runtime ID %d: got %q, want %q", runtimeID, got, name)
		}
	}
}

func TestFallbackReport(t *testing.T) {
	data, err := os.ReadFile("fallbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "0e3c2d81880925cb543426fc8d52193c39798691267d64534dc655634b674014"; got != want {
		t.Fatalf("fallback report SHA256: got %s, want %s", got, want)
	}
	var report struct {
		BlockFallbacks []json.RawMessage `json:"block_fallbacks"`
		ItemFallbacks  []json.RawMessage `json:"item_fallbacks"`
		TargetItems    []json.RawMessage `json:"target_item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got, want := len(report.BlockFallbacks), 1654; got != want {
		t.Fatalf("block fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.ItemFallbacks), 143; got != want {
		t.Fatalf("item fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.TargetItems), 0; got != want {
		t.Fatalf("target item fallback count: got %d, want %d", got, want)
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":   "6b9ecc2b7ddd3b9cd59b5e745d2a704b35e44496c67bcde570f3ea9bcfaed068",
		"vanilla_items.nbt":  "f267b3e73603e1117af489910a02b98c629e92cc44eb7278f8e02bcf47417110",
		"creative_items.nbt": "aed147536e6509aa26587f6b453f4c1dfbdcc7c437e9ad7dd0e6dbfa6277c6ef",
		"crafting_data.nbt":  "2958de8a39b256a30910fe13fa135b2ff6f08c14640b90a5e0f2cb247f3ddb7f",
		"furnace_data.nbt":   "218db815fd0710f303724d55de95356b44f3c0a2afe09d494d973f938faff4ef",
		"smithing_data.nbt":  "e1fb5d48b576fcfc62b43c748f0fb5aaea98d3e9a359e5b33da885d79f08149f",
		"potion_data.nbt":    "4aac61af7f8059d2cdb9c22d2e1f2864d3d3915fc92ded80391f06ae0d0e5a8c",
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
