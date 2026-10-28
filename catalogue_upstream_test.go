//go:build !legacytransport

package multiversion_test

import "testing"

func TestUpstreamCatalogueDoesNotAdvertiseLegacyTransport(t *testing.T) {
	all := effectsProtocols(t)
	if len(all) != 12 {
		t.Fatalf("got %d adapters, want the 12 modern families", len(all))
	}
	for _, p := range all {
		if p.ID() < 748 {
			t.Fatalf("unsupported transport protocol %d advertised", p.ID())
		}
	}
}
