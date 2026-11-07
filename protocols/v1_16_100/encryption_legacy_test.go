//go:build legacytransport

package v1_16_100

import "testing"

func TestEncryptionHasIndependentDirectionState(t *testing.T) {
	p := Protocol{}
	if p.Encryption([32]byte{1}) == p.Encryption([32]byte{1}) {
		t.Fatal("Encryption returned shared state")
	}
}
