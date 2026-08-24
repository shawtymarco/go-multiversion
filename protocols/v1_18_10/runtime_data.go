package v1_18_10

import (
	"fmt"
	"strings"
	"sync"

	"github.com/df-mc/worldupgrader/itemupgrader"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	v486 "github.com/shawtymarco/go-multiversion/data/v486"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type runtimeData struct {
	blocks        *mapping.BlockMapper
	targetItems   map[string]mapping.TargetItem
	creativeCount int
	biomes        []byte

	itemsMu sync.RWMutex
	items   *mapping.ItemMapper
}

func newRuntimeData(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (*runtimeData, error) {
	historicalStates, err := v486.BlockStates()
	if err != nil {
		return nil, err
	}
	states := make([]mapping.BlockState, len(historicalStates))
	for index, state := range historicalStates {
		states[index] = mapping.BlockState{Name: state.Name, Properties: state.Properties, Version: state.Version}
	}
	blocks, err := mapping.NewBlockMapper(native, states)
	if err != nil {
		return nil, fmt.Errorf("build block mapping: %w", err)
	}

	historicalItems, err := v486.Items()
	if err != nil {
		return nil, err
	}
	targetItems := make(map[string]mapping.TargetItem, len(historicalItems))
	for name, runtimeID := range historicalItems {
		targetItems[name] = mapping.TargetItem{RuntimeID: runtimeID}
	}
	creative, err := v486.Creative()
	if err != nil {
		return nil, err
	}
	runtime := &runtimeData{
		blocks:        blocks,
		targetItems:   targetItems,
		creativeCount: len(creative),
		biomes:        v486.BiomeDefinitions(),
	}
	if len(nativeItems) != 0 {
		if _, err := runtime.itemMapper(nativeItems); err != nil {
			return nil, fmt.Errorf("build item mapping: %w", err)
		}
	}
	return runtime, nil
}

func (data *runtimeData) itemMapper(native []protocol.ItemEntry) (*mapping.ItemMapper, error) {
	if data == nil {
		return nil, fmt.Errorf("protocol 486 runtime data is not configured")
	}
	data.itemsMu.RLock()
	items := data.items
	data.itemsMu.RUnlock()
	if items != nil {
		return items, nil
	}
	if len(native) == 0 {
		return nil, fmt.Errorf("native item registry is empty")
	}
	created, err := mapping.NewItemMapperAllowingTargetOnly(native, data.targetItems, resolveTargetItemName)
	if err != nil {
		return nil, err
	}
	data.itemsMu.Lock()
	if data.items == nil {
		data.items = created
	}
	items = data.items
	data.itemsMu.Unlock()
	return items, nil
}

func resolveTargetItemName(name string) string {
	resolved := itemupgrader.Upgrade(itemupgrader.ItemMeta{Name: name}).Name
	if strings.HasPrefix(resolved, "minecraft:item.") {
		resolved = "minecraft:" + strings.TrimPrefix(resolved, "minecraft:item.")
	}
	if resolved == "minecraft:reeds" {
		resolved = "minecraft:sugar_cane"
	}
	return resolved
}

func (data *runtimeData) currentItemMapper() *mapping.ItemMapper {
	if data == nil {
		return nil
	}
	data.itemsMu.RLock()
	defer data.itemsMu.RUnlock()
	return data.items
}
