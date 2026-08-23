// Package multiversion provides Minecraft Bedrock protocol adapters for
// gophertunnel-based servers.
package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
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

// V1_26_44 returns the Minecraft protocol 2168 family adapter. The adapter
// supports Minecraft 1.26.40 through 1.26.44 and keeps this function name for
// API compatibility.
func V1_26_44() minecraft.Protocol {
	return v1_26_44.New()
}
