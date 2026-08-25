// Package v2169 exposes the outgoing Minecraft 1.26.45/protocol 2169
// Dragonfly registry snapshots using production-compatible decoders.
package v2169

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	v1001 "github.com/shawtymarco/go-multiversion/data/v1001"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
	//go:embed vanilla_items.nbt
	vanillaItemData []byte
	//go:embed creative_items.nbt
	creativeItemData []byte
	//go:embed crafting_data.nbt
	craftingData []byte
	//go:embed potion_data.nbt
	potionData []byte
	//go:embed item_tags.json
	itemTagData []byte
)

type (
	BlockState                  = v1001.BlockState
	ItemEntry                   = v1001.ItemEntry
	CreativeData                = v1001.CreativeData
	CraftingData                = v1001.CraftingData
	PotionData                  = v1001.PotionData
	CreativeGroup               = v1001.CreativeGroup
	CreativeItem                = v1001.CreativeItem
	RecipeBlockState            = v1001.RecipeBlockState
	RecipeInput                 = v1001.RecipeInput
	RecipeOutput                = v1001.RecipeOutput
	ShapedRecipe                = v1001.ShapedRecipe
	ShapelessRecipe             = v1001.ShapelessRecipe
	PotionRecipe                = v1001.PotionRecipe
	PotionContainerChangeRecipe = v1001.PotionContainerChangeRecipe
)

func BlockStates() ([]BlockState, error) {
	reader := bytes.NewReader(blockStateData)
	decoder := nbt.NewDecoder(reader)
	states := make([]BlockState, 0, 1<<16)
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
	if err := nbt.Unmarshal(vanillaItemData, &items); err != nil {
		return nil, fmt.Errorf("decode vanilla items: %w", err)
	}
	return items, nil
}

func Creative() (CreativeData, error) {
	var data CreativeData
	if err := nbt.Unmarshal(creativeItemData, &data); err != nil {
		return CreativeData{}, fmt.Errorf("decode creative data: %w", err)
	}
	return data, nil
}

func Crafting() (CraftingData, error) {
	var data CraftingData
	if err := nbt.Unmarshal(craftingData, &data); err != nil {
		return CraftingData{}, fmt.Errorf("decode crafting data: %w", err)
	}
	return data, nil
}

func Potions() (PotionData, error) {
	var data PotionData
	if err := nbt.Unmarshal(potionData, &data); err != nil {
		return PotionData{}, fmt.Errorf("decode potion data: %w", err)
	}
	return data, nil
}

func RawSnapshot(name string) ([]byte, bool) {
	var data []byte
	switch name {
	case "block_states.nbt":
		data = blockStateData
	case "vanilla_items.nbt":
		data = vanillaItemData
	case "creative_items.nbt":
		data = creativeItemData
	case "crafting_data.nbt":
		data = craftingData
	case "potion_data.nbt":
		data = potionData
	case "item_tags.json":
		data = itemTagData
	default:
		return nil, false
	}
	return bytes.Clone(data), true
}
