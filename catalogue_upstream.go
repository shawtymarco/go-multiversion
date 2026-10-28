//go:build !legacytransport

package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
)

// Legacy wire codecs remain available, but upstream cannot accept Login-first sessions.
func appendLegacyProtocols(modern []minecraft.Protocol, _ mapping.BlockRegistry, _ []protocol.ItemEntry) ([]minecraft.Protocol, error) {
	return modern, nil
}
