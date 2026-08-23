// Package multiversion provides Minecraft Bedrock protocol adapters for
// gophertunnel-based servers.
package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/shawtymarco/df-multiversion/protocols/v1_26_44"
)

// Protocols returns all non-native protocols supported by this module.
func Protocols() []minecraft.Protocol {
	return []minecraft.Protocol{v1_26_44.New()}
}

// V1_26_44 returns the Minecraft protocol 2168 family adapter. The adapter
// supports Minecraft 1.26.40 through 1.26.44 and keeps this function name for
// API compatibility.
func V1_26_44() minecraft.Protocol {
	return v1_26_44.New()
}
