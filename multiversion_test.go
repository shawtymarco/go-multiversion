package multiversion_test

import (
	"testing"

	multiversion "github.com/shawtymarco/df-multiversion"
)

func TestProtocols(t *testing.T) {
	protocols := multiversion.Protocols()
	if len(protocols) != 1 {
		t.Fatalf("expected one compatibility protocol, got %d", len(protocols))
	}
	if got, want := protocols[0].ID(), int32(2168); got != want {
		t.Fatalf("unexpected protocol ID: got %d, want %d", got, want)
	}
	if got, want := protocols[0].Ver(), "1.26.44"; got != want {
		t.Fatalf("unexpected game version: got %q, want %q", got, want)
	}
}
