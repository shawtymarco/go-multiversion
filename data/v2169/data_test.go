package v2169

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
	if len(states) == 0 || len(items) == 0 || len(creative.Groups) == 0 || len(creative.Items) == 0 {
		t.Fatal("decoded registry contains an empty required collection")
	}
	if _, ok := items["minecraft:air"]; !ok {
		t.Fatal("item registry does not contain minecraft:air")
	}
	counts := map[string][2]int{
		"block states":             {len(states), 17499},
		"items":                    {len(items), 1976},
		"creative groups":          {len(creative.Groups), 123},
		"creative items":           {len(creative.Items), 1875},
		"shaped recipes":           {len(crafting.Shaped), 1118},
		"shapeless recipes":        {len(crafting.Shapeless), 1431},
		"user-data recipes":        {len(crafting.UserDataShapeless), 1084},
		"multi recipes":            {len(crafting.Multi), 14},
		"potion recipes":           {len(potions.Potions), 210},
		"container-change recipes": {len(potions.ContainerChanges), 2},
	}
	for name, pair := range counts {
		if pair[0] != pair[1] {
			t.Fatalf("%s count: got %d, want %d", name, pair[0], pair[1])
		}
	}
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
	}
}

func TestFallbackReport(t *testing.T) {
	data, err := os.ReadFile("fallbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), "d8b55dc054359a5e738449645385d4d1bb811f64ca00eaec01e0545e66215aaf"; got != want {
		t.Fatalf("fallback report SHA256: got %s, want %s", got, want)
	}
	var report struct {
		BlockFallbacks []json.RawMessage `json:"block_fallbacks"`
		ItemFallbacks  []json.RawMessage `json:"item_fallbacks"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got, want := len(report.BlockFallbacks), 1417; got != want {
		t.Fatalf("block fallback count: got %d, want %d", got, want)
	}
	if got, want := len(report.ItemFallbacks), 100; got != want {
		t.Fatalf("item fallback count: got %d, want %d", got, want)
	}
}

func TestSnapshotHashes(t *testing.T) {
	want := map[string]string{
		"block_states.nbt":   "1dc6d7ea26b48b5b5e4702762e463b95e59eb109f26c0c3b74115d12cb1941a7",
		"vanilla_items.nbt":  "a06a37780c30d10f170f544517ca1bd2ad40a6bce31b22fa126ad633248f5118",
		"creative_items.nbt": "325ef5cc5992bf9b7624a762a62a436469deced5d5a5118b9c7ff959929f23e1",
		"crafting_data.nbt":  "1609f0765b76598707a118941734a01010410b3bf0cd9da2254e4f7fc18f8181",
		"potion_data.nbt":    "282b72a6fe238d7fd3cdb7a9b19eae9eec40647357c708ee0a431d65db87af5b",
		"item_tags.json":     "ac6b456e0c477acc45347d9b3ffd149081f91809518315d454219728d499e58d",
	}
	for name, expected := range want {
		data, ok := RawSnapshot(name)
		if !ok {
			t.Fatalf("snapshot %q is not embedded", name)
		}
		if name == "item_tags.json" {
			data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != expected {
			t.Fatalf("snapshot %q SHA256: got %s, want %s", name, got, expected)
		}
	}
}
