package chunkconv

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func identity(id uint32) (uint32, bool) { return id, true }

// fixturePalette independently writes the documented little-endian index words
// and zigzag palette values. Distinct cells exercise word boundaries and padding.
func fixturePalette(bits int, values []uint32) []byte {
	w := new(bytes.Buffer)
	w.WriteByte(byte(bits<<1 | 1))
	if bits != 0 {
		words := make([]uint32, (4096+32/bits-1)/(32/bits))
		for cell := range 4096 {
			words[cell/(32/bits)] |= uint32(cell%len(values)) << uint(cell%(32/bits)*bits)
		}
		_ = binary.Write(w, binary.LittleEndian, words)
		_ = protocol.WriteVarint32(w, int32(len(values)))
	}
	for _, value := range values {
		_ = protocol.WriteVarint32(w, int32(value))
	}
	return w.Bytes()
}

func TestPaletteRoundTripAndNoMutation(t *testing.T) {
	for _, bits := range []int{0, 1, 2, 3, 4, 5, 6, 8, 16} {
		count := 1
		if bits != 0 {
			count = 1 << bits
			if count > 4096 {
				count = 257
			}
		}
		values := make([]uint32, count)
		for i := range values {
			values[i] = uint32(i + 10)
		}
		source := append([]byte{9, 1, 252}, fixturePalette(bits, values)...)
		// A trailing NBT-like sentinel must not be parsed as palette data.
		source = append(source, 10, 0, 0, 8, 1, 2, 3, 0)
		before := bytes.Clone(source)
		out, err := SubChunk(source, func(id uint32) (uint32, bool) { return id ^ 0x80000000, true })
		if err != nil {
			t.Fatalf("bits %d: %v", bits, err)
		}
		back, err := SubChunk(out, func(id uint32) (uint32, bool) { return id ^ 0x80000000, true })
		if err != nil || !bytes.Equal(source, back) {
			t.Fatalf("bits %d: roundtrip differs (%v)", bits, err)
		}
		if !bytes.Equal(source, before) {
			t.Fatal("source mutated")
		}
	}
}

func TestFallbackDeduplicatesAndRepacks(t *testing.T) {
	source := append([]byte{9, 1, 0}, fixturePalette(2, []uint32{10, 20, 30})...)
	out, err := SubChunk(source, func(uint32) (uint32, bool) { return 99, true })
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte{9, 1, 0}, fixturePalette(0, []uint32{99})...)
	if !bytes.Equal(out, want) {
		t.Fatalf("dedup got %x, want %x", out, want)
	}
	out, err = SubChunk(source, func(id uint32) (uint32, bool) {
		if id == 30 {
			return 10, true
		}
		return id, true
	})
	if err != nil {
		t.Fatal(err)
	}
	if out[3] != 3 {
		t.Fatalf("palette was not reduced to 1 bit: %d", out[3])
	}
	r := bytes.NewReader(out[4:])
	words := make([]uint32, 128)
	if err := binary.Read(r, binary.LittleEndian, words); err != nil {
		t.Fatal(err)
	}
	for cell := range 4096 {
		want := uint32(0)
		if cell%3 == 1 {
			want = 1
		}
		if (words[cell/32]>>uint(cell%32))&1 != want {
			t.Fatalf("wrong cell %d", cell)
		}
	}
}

func TestLevelChunkBiomesReuseAndTail(t *testing.T) {
	block := append([]byte{9, 1, 252}, fixturePalette(0, []uint32{10})...)
	source := append(bytes.Clone(block), fixturePalette(0, []uint32{195})...)
	source = append(source, 255, 255, 0, 10, 0, 42)
	before := bytes.Clone(source)
	out, err := LevelChunk(source, 1, 3, identity, func(id uint32) (uint32, bool) {
		if id == 195 {
			return 4, true
		}
		return id, true
	})
	if err != nil {
		t.Fatal(err)
	}
	want := append(bytes.Clone(block), fixturePalette(0, []uint32{4})...)
	want = append(want, 255, 255, 0, 10, 0, 42)
	if !bytes.Equal(out, want) || !bytes.Equal(source, before) {
		t.Fatal("biome mapping, reuse or tail changed incorrectly")
	}
	if _, err := LevelChunk([]byte{255, 0}, 0, 1, identity, identity); err == nil {
		t.Fatal("first biome cannot reuse a predecessor")
	}
}

func TestMalformedPalettesFailClosed(t *testing.T) {
	good := append([]byte{9, 1, 0}, fixturePalette(1, []uint32{1, 2})...)
	for n := 0; n < len(good); n++ {
		if _, err := SubChunk(good[:n], identity); err == nil {
			t.Fatalf("accepted truncated payload length %d", n)
		}
	}
	bad := bytes.Clone(good)
	bad[3] = 15
	for _, data := range [][]byte{{8, 0, 0}, {9, 1, 0, 255}, bad} {
		if _, err := SubChunk(data, identity); err == nil {
			t.Fatalf("accepted malformed payload %x", data[:3])
		}
	}
	if _, err := SubChunk(good, func(uint32) (uint32, bool) { return 0, false }); err == nil {
		t.Fatal("unknown ID accepted")
	}
	bad = append([]byte{9, 1, 0}, fixturePalette(2, []uint32{1, 2, 3})...)
	bad[4] |= 3
	if _, err := SubChunk(bad, identity); err == nil {
		t.Fatal("out of range palette index accepted")
	}
	if _, err := LevelChunk(nil, 65, 24, identity, identity); err == nil {
		t.Fatal("oversized sub-chunk count accepted")
	}
}

func TestIdentityProducesIndependentPayload(t *testing.T) {
	source := append([]byte{9, 1, 0}, fixturePalette(0, []uint32{123})...)
	out, err := SubChunk(source, identity)
	if err != nil || !reflect.DeepEqual(source, out) {
		t.Fatal(err)
	}
	out[0] = 0
	if source[0] != 9 {
		t.Fatal("output aliases source")
	}
}
