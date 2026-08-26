package packetio

import (
	"bytes"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestSubChunkOffsetsAcceptsFastTransferBurst(t *testing.T) {
	offsets := make([]protocol.SubChunkOffset, 1067)
	for index := range offsets {
		offsets[index] = protocol.SubChunkOffset{int8(index % 64), int8(index % 24), int8(index % 63)}
	}
	var buffer bytes.Buffer
	writer := protocol.NewWriter(&buffer, 0)
	SubChunkOffsets(writer, uint32(len(offsets)), &offsets)

	var decoded []protocol.SubChunkOffset
	reader := protocol.NewReader(&buffer, 0, true)
	SubChunkOffsets(reader, uint32(len(offsets)), &decoded)
	if len(decoded) != len(offsets) {
		t.Fatalf("decoded offset count: got %d, want %d", len(decoded), len(offsets))
	}
}

func TestSubChunkOffsetsRejectsOversizedBurst(t *testing.T) {
	count := MaxSubChunkRequestOffsets + 1
	buffer := bytes.NewBuffer(make([]byte, count*3))
	defer func() {
		if recover() == nil {
			t.Fatal("oversized sub-chunk request was accepted")
		}
	}()
	var decoded []protocol.SubChunkOffset
	SubChunkOffsets(protocol.NewReader(buffer, 0, true), count, &decoded)
}
