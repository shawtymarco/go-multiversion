package v419

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
	var biomes map[string]any
	if err := nbt.NewDecoder(bytes.NewReader(BiomeDefinitions())).Decode(&biomes); err != nil {
		t.Fatalf("decode biome definitions: %v", err)
	}
	if len(states) != 6611 || len(items) != 914 || len(biomes) != 71 {
		t.Fatalf("registry counts: blocks=%d items=%d biomes=%d", len(states), len(items), len(biomes))
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
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "9e87aa4f9279bde0ea0c77614be71ecd50294eeb49f072dee5cf1a82eae5b484"; got != want {
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
	if len(report.Blocks) != 10986 || len(report.Items) != 1097 || len(report.TargetItems) != 8 {
		t.Fatalf("fallback counts: %d/%d/%d", len(report.Blocks), len(report.Items), len(report.TargetItems))
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":      "23ceac32f48fa15b5a8125a4455c84f1687ca93fd87cdf5b7e3ed1c72cc2c224",
		"item_runtime_ids.json": "5657da3bb9118b3914242d3349537d226b119a8dd2b28ef622d6b94ef59c4d42",
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
