package v1_26_30

import (
	"fmt"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	v1001 "github.com/shawtymarco/df-multiversion/data/v1001"
	"github.com/shawtymarco/df-multiversion/mapping"
)

type runtimeData struct {
	blocks      *mapping.BlockMapper
	targetItems map[string]mapping.TargetItem
	creative    v1001.CreativeData

	itemsMu sync.RWMutex
	items   *mapping.ItemMapper
}

func newRuntimeData(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) (*runtimeData, error) {
	historicalStates, err := v1001.BlockStates()
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

	historicalItems, err := v1001.Items()
	if err != nil {
		return nil, err
	}
	targetItems := make(map[string]mapping.TargetItem, len(historicalItems))
	for name, item := range historicalItems {
		targetItems[name] = mapping.TargetItem{
			RuntimeID:      item.RuntimeID,
			ComponentBased: item.ComponentBased,
			Version:        item.Version,
			Data:           item.Data,
		}
	}
	creative, err := v1001.Creative()
	if err != nil {
		return nil, err
	}
	runtime := &runtimeData{blocks: blocks, targetItems: targetItems, creative: creative}
	if len(nativeItems) != 0 {
		if _, err := runtime.itemMapper(nativeItems); err != nil {
			return nil, fmt.Errorf("build item mapping: %w", err)
		}
	}
	return runtime, nil
}

func (data *runtimeData) itemMapper(native []protocol.ItemEntry) (*mapping.ItemMapper, error) {
	if data == nil {
		return nil, fmt.Errorf("protocol 1001 runtime data is not configured")
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
	created, err := mapping.NewItemMapper(native, data.targetItems)
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

func (data *runtimeData) currentItemMapper() *mapping.ItemMapper {
	if data == nil {
		return nil
	}
	data.itemsMu.RLock()
	defer data.itemsMu.RUnlock()
	return data.items
}
