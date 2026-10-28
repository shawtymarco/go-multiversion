//go:build legacytransport

package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
	"github.com/shawtymarco/go-multiversion/protocols/v1_16_100"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_0"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_10"
)

// This opt-in build requires the documented gophertunnel legacy transport fork.
func appendLegacyProtocols(modern []minecraft.Protocol, blocks mapping.BlockRegistry, items []protocol.ItemEntry) ([]minecraft.Protocol, error) {
	for _, construct := range []func(mapping.BlockRegistry, []protocol.ItemEntry) (minecraft.Protocol, error){v1_18_10.NewWithRegistries, v1_18_0.NewWithRegistries, v1_16_100.NewWithRegistries} {
		p, err := construct(blocks, items)
		if err != nil {
			return nil, err
		}
		modern = append(modern, p)
	}
	return modern, nil
}
