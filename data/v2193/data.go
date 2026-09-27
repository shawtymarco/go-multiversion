// Package v2193 provides the native vanilla registry for standalone proxies.
// Source: df-mc/dragonfly 4c7b5074be94fa83a1cd98e9c752083ad04a6e21.
package v2193

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/shawtymarco/go-multiversion/mapping"
)

//go:embed block_states.nbt
var blocks []byte

func BlockStates() ([]mapping.BlockState, error) {
	if fmt.Sprintf("%x", sha256.Sum256(blocks)) != "f0784a6284d6ca7d98cc3472f4ce84241a11e11b18ed16f6591dfd5e6da6fbd6" {
		return nil, fmt.Errorf("native block registry checksum mismatch")
	}
	r := bytes.NewReader(blocks)
	d := nbt.NewDecoder(r)
	var result []mapping.BlockState
	for r.Len() != 0 {
		var entry struct {
			Name    string         `nbt:"name"`
			States  map[string]any `nbt:"states"`
			Version int32          `nbt:"version"`
		}
		if err := d.Decode(&entry); err != nil {
			return nil, fmt.Errorf("native block %d: %w", len(result), err)
		}
		result = append(result, mapping.BlockState{Name: entry.Name, Properties: entry.States, Version: entry.Version})
	}
	return result, nil
}
