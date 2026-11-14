package packetio

import (
	"bytes"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestTrustedSkinBoolWritesWithoutChangingThePacket(t *testing.T) {
	for _, flag := range []protocol.Optional[bool]{{}, protocol.Option(false), protocol.Option(true)} {
		var buf bytes.Buffer
		before := flag
		TrustedSkinBool(protocol.NewWriter(&buf, 0), &flag)
		if flag != before {
			t.Fatalf("writing changed the flag from %v to %v", before, flag)
		}
		want, _ := before.Value()
		var decoded protocol.Optional[bool]
		TrustedSkinBool(protocol.NewReader(bytes.NewReader(buf.Bytes()), 0, true), &decoded)
		if value, ok := decoded.Value(); !ok || value != want {
			t.Fatalf("read %v, want set %v", decoded, want)
		}
	}
}
