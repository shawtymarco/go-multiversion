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
	if got := multiversion.V1_26_20(); got.ID() != 975 || got.Ver() != "1.26.23" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 975/%q", got.ID(), got.Ver(), "1.26.23")
	}
	if got := multiversion.V1_26_10(); got.ID() != 944 || got.Ver() != "1.26.14" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 944/%q", got.ID(), got.Ver(), "1.26.14")
	}
	if got := multiversion.V1_26_0(); got.ID() != 924 || got.Ver() != "1.26.3" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 924/%q", got.ID(), got.Ver(), "1.26.3")
	}
	if got := multiversion.V1_21_130(); got.ID() != 898 || got.Ver() != "1.21.132" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 898/%q", got.ID(), got.Ver(), "1.21.132")
	}
	if got := multiversion.V1_21_110(); got.ID() != 844 || got.Ver() != "1.21.114" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 844/%q", got.ID(), got.Ver(), "1.21.114")
	}
	if got := multiversion.V1_21_100(); got.ID() != 827 || got.Ver() != "1.21.102" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 827/%q", got.ID(), got.Ver(), "1.21.102")
	}
	if got := multiversion.V1_21_50(); got.ID() != 766 || got.Ver() != "1.21.51" {
		t.Fatalf("unexpected protocol: got %d/%q, want 766/%q", got.ID(), got.Ver(), "1.21.51")
	}
	if got := multiversion.V1_21_40(); got.ID() != 748 || got.Ver() != "1.21.44" {
		t.Fatalf("unexpected protocol: got %d/%q, want 748/%q", got.ID(), got.Ver(), "1.21.44")
	}
	if got := multiversion.V1_18_10(); got.ID() != 486 || got.Ver() != "1.18.12" {
		t.Fatalf("unexpected development protocol: got %d/%q, want 486/%q", got.ID(), got.Ver(), "1.18.12")
	}
	if got := multiversion.V1_18_0(); got.ID() != 475 || got.Ver() != "1.18.2" {
		t.Fatalf("unexpected protocol: got %d/%q, want 475/%q", got.ID(), got.Ver(), "1.18.2")
	}
	if got := multiversion.V1_16_100(); got.ID() != 419 || got.Ver() != "1.16.100" {
		t.Fatalf("unexpected protocol: got %d/%q, want 419/%q", got.ID(), got.Ver(), "1.16.100")
	}
	for _, supported := range protocols {
		if supported.ID() == 1001 || supported.ID() == 975 || supported.ID() == 944 || supported.ID() == 924 || supported.ID() == 898 || supported.ID() == 844 || supported.ID() == 827 || supported.ID() == 766 || supported.ID() == 748 || supported.ID() == 486 || supported.ID() == 475 {
			t.Fatalf("registry-aware protocol %d must not be enabled by Protocols", supported.ID())
		}
	}
}
