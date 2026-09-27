package v1_18_0

import (
	"bytes"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProxyLevelChunkRequestMode475Oracle(t *testing.T) {
	// c40bf828 LevelChunk: X, Z, uint32 count, cache flag, payload. Mojang
	// r18 uses UINT32_MAX to request individual sub-chunks, without a limit.
	want := []byte{0, 0, 255, 255, 255, 255, 15, 0, 1, 'x'}
	p := New().(*Protocol).ProxyWireProtocol()
	input := &packet.LevelChunk{SubChunkLimit: protocol.Option(int32(-1)), RawPayload: []byte{'x'}}
	var b bytes.Buffer
	p.ConvertFromLatest(input, nil)[0].Marshal(p.NewWriter(&b, 0))
	if !bytes.Equal(b.Bytes(), want) {
		t.Fatalf("historical bytes %x != %x", b.Bytes(), want)
	}
	if input.SubChunkCount != 0 {
		t.Fatal("encoding mutated native count")
	}
	decoded := p.Packets(false)[packet.IDLevelChunk]()
	decoded.Marshal(p.NewReader(&b, 0, true))
	got := p.ConvertToLatest(decoded, nil)[0].(*packet.LevelChunk)
	limit, present := got.SubChunkLimit.Value()
	if b.Len() != 0 || !present || limit != -1 || got.SubChunkCount != 0 || !bytes.Equal(got.RawPayload, []byte{'x'}) {
		t.Fatalf("decoded %+v", got)
	}
}
