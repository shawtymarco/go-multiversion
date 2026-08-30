package v924

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
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "9f83193747de73b954b75286429089db9583d9a69dc6a8ba19e52e54bbf3ee7a"; got != want {
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
		"vanilla_items.nbt":  "5e50ca77286e8e6c33d9ecd1de1e532086b36a54d0605963a6a303a642dc5bed",
		"creative_items.nbt": "39b4d67cf2836d76066722e379cc8510679adf59d296e3d37958bf6f3318090b",
		"crafting_data.nbt":  "9346c8802f043df4a279049acc72f9fb3506467fe4a0944a6372ca07500809b2",
		"furnace_data.nbt":   "4245b8c208c7f55c00ee3ce8a90a1d9c1dfa137cf85ac19248db25a962291817",
		"potion_data.nbt":    "e6d8170070b8f02d64baac0d89f02ef59dcceaa568b1d35ae4b7cdf45d52645a",
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
