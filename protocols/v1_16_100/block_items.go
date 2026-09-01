package v1_16_100

import (
	"fmt"
	"strings"

	v419data "github.com/shawtymarco/go-multiversion/data/v419"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type targetBlockItemKey struct {
	name string
	meta int16
}

var (
	targetBlockItemMetadata map[string]int16
	targetBlockItemStates   map[targetBlockItemKey]mapping.BlockState
)

func init() {
	states, err := v419data.BlockStates()
	if err != nil {
		panic(fmt.Errorf("load protocol 419 block item states: %w", err))
	}
	metadata, err := v419data.BlockItemMetadata()
	if err != nil {
		panic(fmt.Errorf("load protocol 419 block item metadata: %w", err))
	}
	if len(metadata) != len(states) {
		panic(fmt.Errorf("protocol 419 block item metadata count %d differs from state count %d", len(metadata), len(states)))
	}
	targetBlockItemMetadata = make(map[string]int16, len(states))
	targetBlockItemStates = make(map[targetBlockItemKey]mapping.BlockState, len(states))
	for runtimeID, state := range states {
		stateKey, err := mapping.StateKey(state.Name, state.Properties)
		if err != nil {
			panic(fmt.Errorf("build protocol 419 block item state key %d: %w", runtimeID, err))
		}
		targetBlockItemMetadata[stateKey] = metadata[runtimeID]
		key := targetBlockItemKey{name: normaliseTargetBlockItemName(state.Name), meta: metadata[runtimeID]}
		if _, exists := targetBlockItemStates[key]; !exists {
			targetBlockItemStates[key] = mapping.BlockState{Name: state.Name, Properties: state.Properties, Version: state.Version}
		}
	}
}

func targetBlockItemMeta(state mapping.BlockState) (int16, bool) {
	key, err := mapping.StateKey(state.Name, state.Properties)
	if err != nil {
		return 0, false
	}
	metadata, ok := targetBlockItemMetadata[key]
	return metadata, ok
}

func targetBlockItemState(name string, metadata uint32) (mapping.BlockState, bool) {
	state, ok := targetBlockItemStates[targetBlockItemKey{name: normaliseTargetBlockItemName(name), meta: int16(metadata)}]
	return state, ok
}

func normaliseTargetBlockItemName(name string) string {
	if strings.HasPrefix(name, "minecraft:item.") {
		return "minecraft:" + strings.TrimPrefix(name, "minecraft:item.")
	}
	return name
}
