package multiversion_test

import (
	"testing"

	multiversion "github.com/shawtymarco/go-multiversion"
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
	if got := multiversion.V1_26_30(); got.ID() != 1001 || got.Ver() != "1.26.36" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 1001/%q", got.ID(), got.Ver(), "1.26.36")
	}
	if got := multiversion.V1_21_110(); got.ID() != 844 || got.Ver() != "1.21.114" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 844/%q", got.ID(), got.Ver(), "1.21.114")
	}
	if got := multiversion.V1_21_100(); got.ID() != 827 || got.Ver() != "1.21.102" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 827/%q", got.ID(), got.Ver(), "1.21.102")
	}
	for _, supported := range protocols {
		if supported.ID() == 1001 || supported.ID() == 844 || supported.ID() == 827 {
			t.Fatalf("registry-aware protocol %d must not be enabled by Protocols", supported.ID())
		}
	}
}
