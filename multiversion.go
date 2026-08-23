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

// V1_26_44 returns the Minecraft 1.26.44 protocol adapter.
func V1_26_44() minecraft.Protocol {
	return v1_26_44.New()
}
