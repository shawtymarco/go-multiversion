// Package multiversion provides Minecraft Bedrock protocol adapters for
// gophertunnel-based servers.
package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/df-multiversion/mapping"
	"github.com/shawtymarco/df-multiversion/protocols/v1_26_30"
	"github.com/shawtymarco/df-multiversion/protocols/v1_26_44"
)

// Protocols returns all non-native protocols supported by this module.
func Protocols() []minecraft.Protocol {
	return []minecraft.Protocol{v1_26_44.New()}
}

// V1_26_30 returns the wire-only Minecraft protocol 1001 family adapter for
// Minecraft 1.26.30 through 1.26.36. The adapter is not included in Protocols
// until its registry and chunk conversions are complete.
func V1_26_30() minecraft.Protocol {
	return v1_26_30.New()
}

// V1_26_30WithBlockRegistry returns a fully configured protocol-1001 adapter
// using a current native block registry.
func V1_26_30WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_26_30.NewWithBlockRegistry(native)
}

// ProtocolsWithBlockRegistry returns all verified non-native protocols,
// including protocol 1001 configured against the current block registry.
func ProtocolsWithBlockRegistry(native mapping.BlockRegistry) ([]minecraft.Protocol, error) {
	legacy, err := V1_26_30WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	return []minecraft.Protocol{v1_26_44.New(), legacy}, nil
}

// ProtocolsWithRegistries returns all verified adapters after eagerly
// validating the current block and item registries.
func ProtocolsWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) ([]minecraft.Protocol, error) {
	legacy, err := v1_26_30.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return []minecraft.Protocol{v1_26_44.New(), legacy}, nil
}

// V1_26_44 returns the Minecraft protocol 2168 family adapter. The adapter
// supports Minecraft 1.26.40 through 1.26.44 and keeps this function name for
// API compatibility.
func V1_26_44() minecraft.Protocol {
	return v1_26_44.New()
}
