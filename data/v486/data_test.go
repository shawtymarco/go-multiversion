package v486

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
		"block states":     {len(states), 7946},
		"items":            {len(items), 1091},
		"creative entries": {len(creative), 1388},
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
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "a8f8ac92ded30dd9ef08634c3bbd8da38689cff8b10338aa02bb10c374fedadf"; got != want {
		t.Fatalf("fallback report SHA256: got %s, want %s", got, want)
	}
	var report struct {
		BlockFallbacks      []json.RawMessage `json:"block_fallbacks"`
		ItemFallbacks       []json.RawMessage `json:"item_fallbacks"`
		TargetItemFallbacks []json.RawMessage `json:"target_item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got, want := len(report.BlockFallbacks), 9617; got != want {
		t.Fatalf("block fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.ItemFallbacks), 926; got != want {
		t.Fatalf("item fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.TargetItemFallbacks), 13; got != want {
		t.Fatalf("target item fallback count: got %d, want %d", got, want)
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":      "cedd121a64ed81646012cfaeef72b7908957f38328a3d9bb4e5adfff82d1744b",
		"item_runtime_ids.nbt":  "e800bb455965cb1ab603c5d023c48e9a936371e04036a2439f4d5d330483b6bb",
		"creative_items.nbt":    "d411505e21ff37f6144ccc00b14e6b39ed1cb8cc80901076a599314f73d0e571",
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
