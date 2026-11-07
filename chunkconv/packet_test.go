package chunkconv

import (
	"bytes"
	"math"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type shiftBiomes struct{}

func (shiftBiomes) MapBlockRuntimeID(id uint32) (uint32, bool) { return id, true }
func (shiftBiomes) MapBiomeRuntimeID(id uint32) (uint32, bool) { return id + 1, true }

// biomesOnly is a LevelChunk payload without sub-chunks: one single-value biome
// section per overworld section, then no border blocks.
func biomesOnly(biome int32) []byte {
	buf := new(bytes.Buffer)
	w := protocol.NewWriter(buf, -1)
	for range 24 {
		buf.WriteByte(1)
		w.Varint32(&biome)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}

func TestPacketMapsRequestModeBiomes(t *testing.T) {
	source := &packet.LevelChunk{Dimension: packet.DimensionOverworld, SubChunkCount: math.MaxUint32 - 1, RawPayload: biomesOnly(3)}
	out, err := Packet(source, shiftBiomes{})
	if err != nil {
		t.Fatal(err)
	}
	if got := out.(*packet.LevelChunk).RawPayload; !bytes.Equal(got, biomesOnly(4)) {
		t.Fatalf("biomes were not mapped")
	}
	if !bytes.Equal(source.RawPayload, biomesOnly(3)) {
		t.Fatal("the source packet was modified")
	}
}

func TestPacketLeavesOthersAlone(t *testing.T) {
	pk := &packet.LevelChunk{Dimension: packet.DimensionOverworld, RawPayload: biomesOnly(3)}
	if out, err := Packet(pk, struct{}{}); err != nil || out != pk {
		t.Fatalf("a proto without a mapper changed the packet: %v", err)
	}
	text := &packet.Text{}
	if out, err := Packet(text, shiftBiomes{}); err != nil || out != text {
		t.Fatalf("a non-chunk packet changed: %v", err)
	}
	if _, err := Packet(&packet.LevelChunk{CacheEnabled: true}, shiftBiomes{}); err == nil {
		t.Fatal("a cached chunk was accepted")
	}
}
