package multiversion

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
)

// Config selects historical adapters without changing their protocol semantics.
// The native protocol remains accepted by gophertunnel itself.
type Config struct {
	// MinimumProtocol is an inclusive protocol ID floor. Zero enables the full
	// modern catalogue (or the legacytransport opt-in catalogue). A floor never
	// adds support for an unlisted protocol.
	MinimumProtocol int32
}

// Validate rejects a floor that cannot be enforced by the native listener.
func (c Config) Validate() error {
	if c.MinimumProtocol < 0 || c.MinimumProtocol > protocol.CurrentProtocol {
		return fmt.Errorf("minimum protocol must be between 0 and %d, got %d", protocol.CurrentProtocol, c.MinimumProtocol)
	}
	return nil
}

// ProtocolsWithRegistries builds the supported catalogue and applies the
// configured floor. Login-first clients are included only in legacytransport
// builds, which require the separately documented transport fork.
func (c Config) ProtocolsWithRegistries(native mapping.BlockRegistry, items []protocol.ItemEntry) ([]minecraft.Protocol, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	protocols, err := ProtocolsWithRegistries(native, items)
	if err != nil {
		return nil, err
	}
	return c.selectProtocols(protocols), nil
}

func (c Config) selectProtocols(protocols []minecraft.Protocol) []minecraft.Protocol {
	selected := make([]minecraft.Protocol, 0, len(protocols))
	for _, p := range protocols {
		if p.ID() >= c.MinimumProtocol {
			selected = append(selected, p)
		}
	}
	// gophertunnel appends native to the configured slice per connection.
	// Do not expose spare capacity shared by concurrent listener accepts.
	return selected[:len(selected):len(selected)]
}
