package v1_16_100

import (
	"testing"

	"github.com/df-mc/worldupgrader/blockupgrader"
	v419data "github.com/shawtymarco/go-multiversion/data/v419"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func protocol419IntegrationBlockRegistry(t *testing.T) blockItemTestRegistry {
	t.Helper()
	target, err := v419data.BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	states := make([]mapping.BlockState, 1, len(target))
	seen := make(map[string]struct{}, len(target))
	for _, state := range target {
		upgraded := blockupgrader.Upgrade(blockupgrader.BlockState{Name: state.Name, Properties: state.Properties, Version: state.Version})
		key, err := mapping.StateKey(upgraded.Name, upgraded.Properties)
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		native := mapping.BlockState{Name: upgraded.Name, Properties: upgraded.Properties, Version: upgraded.Version}
		if upgraded.Name == "minecraft:air" {
			states[0] = native
			continue
		}
		states = append(states, native)
	}
	if states[0].Name != "minecraft:air" {
		t.Fatal("upgraded protocol-419 registry has no air state")
	}
	return blockItemTestRegistry{states: states}
}
