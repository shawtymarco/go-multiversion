package v748

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
	want := []int{14196, 1749, 1920, 1003, 1083, 210, 215, 9, 71}
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
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "079bfb8d02826279e5768923e876464855b2f638ea2ddf5dfd494ca1d5f6ba7b"; got != want {
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
	if len(report.Blocks) != 3303 || len(report.Items) != 301 || len(report.TargetItems) != 1 {
		t.Fatalf("fallback counts: %d/%d/%d", len(report.Blocks), len(report.Items), len(report.TargetItems))
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"biome_definitions.nbt": "dc7dec09e45b333583120f826c3b4ddd986f707515673c543f442f576601954f",
		"block_states.nbt":      "4e434d9c40988697f61ed2053586330796bf5216f602cb051aa9061cd67a4ea7",
		"crafting_data.nbt":     "b6b50c57bcf3d75b1bd88c82f1c629f057995842a0f7ab6e0fc11f60991b8eac",
		"creative_items.nbt":    "9f34c8cbc40913124ca9b45e4ea953ae3daefc5cb6a918de11850a7da6b57b20",
		"furnace_data.nbt":      "926a37b7d23e6b912e2f9fd2fea7706634be6d4347c3c4edc936a440e34fecd6",
		"item_runtime_ids.nbt":  "de720c1934997838a1f21e984b37c24716c75c2f203fe7963b36df74548e4f5e",
		"potion_data.nbt":       "e34ad9d698d86e8cecf8303bb97f7d7737a593188c52a2a6925e85849cffb19d",
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
