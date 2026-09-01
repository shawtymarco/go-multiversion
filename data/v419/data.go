// Package v419 exposes the locked Minecraft 1.16.100/protocol-419 registry snapshots.
package v419

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
	//go:embed item_runtime_ids.json
	itemRuntimeIDData []byte
	//go:embed block_item_meta.json
	blockItemMetadataData []byte
	//go:embed biome_definitions.nbt
	biomeDefinitionData []byte
)

type BlockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

type ItemEntry struct {
	RuntimeID      int16 `json:"runtime_id"`
	ComponentBased bool  `json:"component_based"`
}

func BlockStates() ([]BlockState, error) {
	reader := bytes.NewReader(blockStateData)
	decoder := nbt.NewDecoder(reader)
	states := make([]BlockState, 0, 1<<13)
	for {
		var state BlockState
		if err := decoder.Decode(&state); err != nil {
			if reader.Len() == 0 {
				return states, nil
			}
			return nil, fmt.Errorf("decode block state %d: %w", len(states), err)
		}
		states = append(states, state)
	}
}

func Items() (map[string]ItemEntry, error) {
	items := map[string]ItemEntry{}
	if err := json.Unmarshal(itemRuntimeIDData, &items); err != nil {
		return nil, fmt.Errorf("decode item runtime IDs: %w", err)
	}
	return items, nil
}

func BlockItemMetadata() ([]int16, error) {
	var metadata []int16
	if err := json.Unmarshal(blockItemMetadataData, &metadata); err != nil {
		return nil, fmt.Errorf("decode block item metadata: %w", err)
	}
	return metadata, nil
}

func BiomeDefinitions() []byte { return bytes.Clone(biomeDefinitionData) }

func RawSnapshot(name string) ([]byte, bool) {
	var data []byte
	switch name {
	case "block_states.nbt":
		data = blockStateData
	case "item_runtime_ids.json":
		data = itemRuntimeIDData
	case "block_item_meta.json":
		data = blockItemMetadataData
	case "biome_definitions.nbt":
		data = biomeDefinitionData
	default:
		return nil, false
	}
	return bytes.Clone(data), true
}
