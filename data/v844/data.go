// Package v844 exposes the exact historical Dragonfly registry snapshots for
// Minecraft 1.21.11x/protocol 844 using typed production-compatible decoders.
package v844

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
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
	//go:embed furnace_data.nbt
	furnaceData []byte
	//go:embed smithing_data.nbt
	smithingData []byte
)

// BlockState is one exact state in the historical network block registry.
type BlockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

// ItemEntry is one identifier-keyed item registry entry.
type ItemEntry struct {
	RuntimeID      int32          `nbt:"runtime_id"`
	ComponentBased bool           `nbt:"component_based"`
	Version        int32          `nbt:"version"`
	Data           map[string]any `nbt:"data,omitempty"`
}

// CreativeData contains the historical creative groups and items.
type CreativeData struct {
	Groups []CreativeGroup `nbt:"groups"`
	Items  []CreativeItem  `nbt:"items"`
}

// CreativeGroup contains one historical creative inventory group.
type CreativeGroup struct {
	Category int32        `nbt:"category"`
	Name     string       `nbt:"name"`
	Icon     CreativeItem `nbt:"icon"`
}

// CreativeItem contains one historical creative stack description.
type CreativeItem struct {
	Name            string         `nbt:"name"`
	Meta            int16          `nbt:"meta"`
	NBT             map[string]any `nbt:"nbt,omitempty"`
	BlockProperties map[string]any `nbt:"block_properties,omitempty"`
	GroupIndex      int32          `nbt:"group_index,omitempty"`
}

// RecipeBlockState is the exact block state attached to a recipe item.
type RecipeBlockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

// RecipeInput is one item or item tag consumed by a recipe.
type RecipeInput struct {
	Name  string           `nbt:"name"`
	Meta  int32            `nbt:"meta"`
	Count int32            `nbt:"count"`
	State RecipeBlockState `nbt:"block"`
	Tag   string           `nbt:"tag"`
}

// RecipeOutput is one item stack produced by a recipe.
type RecipeOutput struct {
	Name    string           `nbt:"name"`
	Meta    int32            `nbt:"meta"`
	Count   int16            `nbt:"count"`
	State   RecipeBlockState `nbt:"block"`
	NBTData map[string]any   `nbt:"data"`
}

// ShapedRecipe is a historical shaped crafting recipe.
type ShapedRecipe struct {
	Input    []RecipeInput  `nbt:"input"`
	Output   []RecipeOutput `nbt:"output"`
	Block    string         `nbt:"block"`
	Width    int32          `nbt:"width"`
	Height   int32          `nbt:"height"`
	Priority int32          `nbt:"priority"`
}

// ShapelessRecipe is a historical shapeless crafting recipe.
type ShapelessRecipe struct {
	Input    []RecipeInput  `nbt:"input"`
	Output   []RecipeOutput `nbt:"output"`
	Block    string         `nbt:"block"`
	Priority int32          `nbt:"priority"`
}

// CraftingData contains all historical crafting recipe categories.
type CraftingData struct {
	Shaped            []ShapedRecipe    `nbt:"shaped"`
	Shapeless         []ShapelessRecipe `nbt:"shapeless"`
	UserDataShapeless []ShapelessRecipe `nbt:"shulker_box"`
	Multi             []string          `nbt:"multi"`
}

// PotionRecipe is a historical brewing recipe.
type PotionRecipe struct {
	Input   RecipeInput  `nbt:"input"`
	Reagent RecipeInput  `nbt:"reagent"`
	Output  RecipeOutput `nbt:"output"`
}

// PotionContainerChangeRecipe is a historical brewing container recipe.
type PotionContainerChangeRecipe struct {
	Input   string      `nbt:"input"`
	Reagent RecipeInput `nbt:"reagent"`
	Output  string      `nbt:"output"`
}

// PotionData contains historical brewing and container-change recipes.
type PotionData struct {
	Potions          []PotionRecipe                `nbt:"potions"`
	ContainerChanges []PotionContainerChangeRecipe `nbt:"container_changes"`
}

// FurnaceRecipe is one historical furnace-family recipe.
type FurnaceRecipe struct {
	Input  RecipeInput  `nbt:"input"`
	Output RecipeOutput `nbt:"output"`
	Block  string       `nbt:"block"`
}

// BlockStates decodes every concatenated root compound in block_states.nbt.
func BlockStates() ([]BlockState, error) {
	reader := bytes.NewReader(blockStateData)
	decoder := nbt.NewDecoder(reader)
	states := make([]BlockState, 0, 1<<16)
	for {
		var state BlockState
		if err := decoder.Decode(&state); err != nil {
			// Dragonfly stores concatenated root compounds without an outer
			// length. The decoder reports its generic buffer-end error after the
			// final complete compound, so an exhausted verified blob is EOF.
			if reader.Len() == 0 {
				return states, nil
			}
			return nil, fmt.Errorf("decode block state %d: %w", len(states), err)
		}
		states = append(states, state)
	}
}

// Items decodes the historical identifier-keyed item registry.
func Items() (map[string]ItemEntry, error) {
	items := map[string]ItemEntry{}
	if err := nbt.Unmarshal(vanillaItemData, &items); err != nil {
		return nil, fmt.Errorf("decode vanilla items: %w", err)
	}
	return items, nil
}

// Creative decodes historical creative groups and items.
func Creative() (CreativeData, error) {
	var data CreativeData
	if err := nbt.Unmarshal(creativeItemData, &data); err != nil {
		return CreativeData{}, fmt.Errorf("decode creative data: %w", err)
	}
	return data, nil
}

// Crafting decodes historical crafting recipes.
func Crafting() (CraftingData, error) {
	var data CraftingData
	if err := nbt.Unmarshal(craftingData, &data); err != nil {
		return CraftingData{}, fmt.Errorf("decode crafting data: %w", err)
	}
	return data, nil
}

// Potions decodes historical potion and container-change recipes.
func Potions() (PotionData, error) {
	var data PotionData
	if err := nbt.Unmarshal(potionData, &data); err != nil {
		return PotionData{}, fmt.Errorf("decode potion data: %w", err)
	}
	return data, nil
}

// Furnaces decodes the historical furnace-family recipes.
func Furnaces() ([]FurnaceRecipe, error) {
	var data []FurnaceRecipe
	if err := nbt.Unmarshal(furnaceData, &data); err != nil {
		return nil, fmt.Errorf("decode furnace data: %w", err)
	}
	return data, nil
}

// Smithing decodes the historical smithing transform recipes.
func Smithing() ([]ShapelessRecipe, error) {
	var data []ShapelessRecipe
	if err := nbt.Unmarshal(smithingData, &data); err != nil {
		return nil, fmt.Errorf("decode smithing data: %w", err)
	}
	return data, nil
}

// RawSnapshot returns a copy of one embedded source blob by file name.
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
	case "furnace_data.nbt":
		data = furnaceData
	case "smithing_data.nbt":
		data = smithingData
	default:
		return nil, false
	}
	return bytes.Clone(data), true
}
