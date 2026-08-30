package v475

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
	legacy, err := LegacyStates()
	if err != nil {
		t.Fatal(err)
	}
	var biomes map[string]any
	if err := nbt.NewDecoder(bytes.NewReader(BiomeDefinitions())).Decode(&biomes); err != nil {
		t.Fatalf("decode biome definitions: %v", err)
	}
	counts := map[string]struct{ got, want int }{
		"block states":     {len(states), 7948},
		"items":            {len(items), 1081},
		"creative entries": {len(creative), 1386},
		"legacy states":    {len(legacy), 3307},
		"biomes":           {len(biomes), 71},
	}
	for name, count := range counts {
		if count.got != count.want {
			t.Fatalf("%s count: got %d, want %d", name, count.got, count.want)
		}
	}
	if _, ok := items["minecraft:air"]; !ok {
		t.Fatal("item registry does not contain minecraft:air")
	}
	airRuntimeID := -1
	for runtimeID, state := range states {
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
			airRuntimeID = runtimeID
		}
	}
	if airRuntimeID != 134 {
		t.Fatalf("minecraft:air runtime ID: got %d, want 134", airRuntimeID)
	}
}

func TestFallbackReport(t *testing.T) {
	data, err := os.ReadFile("fallbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "53bc2d6877dcf46099281f5840fe4e8f4456c5228094e915ef19c6c3e6b6241f"; got != want {
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
	if len(report.Blocks) != 9621 || len(report.Items) != 934 || len(report.TargetItems) != 11 {
		t.Fatalf("fallback counts: %d/%d/%d", len(report.Blocks), len(report.Items), len(report.TargetItems))
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":      "1e77f6c9e6c9eb1f57ac51066549f336883b2509cf487086f91e5989400d653b",
		"item_runtime_ids.nbt":  "76fb5abd9f18a12f1cba12898b51bdfcb3e9ab04f5c852458c4a0f6175fa3648",
		"creative_items.nbt":    "bdb590c31fc1732dd49120996d00b3421e976f3d8105b0c6e1ee68720f1111dc",
		"legacy_states.nbt":     "e892de43a47546466b94e696157056dd703d0727271f51ae82f5e4d29a5e38c1",
		"biome_definitions.nbt": "dc7dec09e45b333583120f826c3b4ddd986f707515673c543f442f576601954f",
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
