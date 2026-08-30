package v766

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
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
	var biomes map[string]any
	if err := nbt.NewDecoder(bytes.NewReader(BiomeDefinitions())).Decode(&biomes); err != nil {
		t.Fatal(err)
	}
	counts := []int{len(states), len(items), len(creative), len(crafting.Shaped), len(crafting.Shapeless), len(potions.Potions), len(furnaces), len(smithing), len(biomes)}
	want := []int{15162, 1789, 1954, 1034, 1093, 210, 220, 9, 71}
	for index, count := range counts {
		if count != want[index] {
			t.Fatalf("registry count %d: got %d, want %d", index, count, want[index])
		}
	}
}

func TestFallbackReport(t *testing.T) {
	data, err := os.ReadFile("fallbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "e1160705b0cfe09bea22d13144087d9efa48759c072d133b2801a89d9e2bbd0f"; got != want {
		t.Fatalf("fallback SHA256: got %s, want %s", got, want)
	}
	var report struct {
		Blocks      []json.RawMessage `json:"block_fallbacks"`
		Items       []json.RawMessage `json:"item_fallbacks"`
		TargetItems []json.RawMessage `json:"target_item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Blocks) != 2337 || len(report.Items) != 261 || len(report.TargetItems) != 1 {
		t.Fatalf("fallback counts: %d/%d/%d", len(report.Blocks), len(report.Items), len(report.TargetItems))
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"biome_definitions.nbt": "dc7dec09e45b333583120f826c3b4ddd986f707515673c543f442f576601954f",
		"block_states.nbt":      "f6aeb01ca776cd142d3b457962b0b5aec951f6c5e48aca9ff0fa6445982015b3",
		"crafting_data.nbt":     "907e8af15b530cc2d50ddbed9661744d45f18ea023fbefdaa5e934b534b723ac",
		"creative_items.nbt":    "9e410ef0c46597dbbab36cfecc92b1361216665ca2e6c85d47511e79cc134cbd",
		"furnace_data.nbt":      "7fac2fec50861c6d2dc773e3ede0c994cb45f4d756afc9a7ad4c03871d0202cb",
		"item_runtime_ids.nbt":  "d7a7299446c1f79596f0aaa5bcf0d8f5d398f25a4c00154b7299140187676419",
		"potion_data.nbt":       "b6e2d5e7f45785d9ae9dfd7bea87d9c4862a1f57f0db1dc47e98a7d41fd2a7d7",
		"smithing_data.nbt":     "99c9eb3ea54fca8a43c880fa91e6f1eb0ca350a051aad7e5ba1621659e3e4f7b",
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
