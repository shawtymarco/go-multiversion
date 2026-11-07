// Package chunkconv translates network chunk palettes without accessing a server's
// live world. Its input is Dragonfly's version-9 numeric network encoding.
package chunkconv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Mapper func(uint32) (uint32, bool)

// SubChunk maps one version-9 sub-chunk and preserves the trailing block NBT.
func SubChunk(data []byte, blocks Mapper) ([]byte, error) {
	r, w := bytes.NewReader(data), new(bytes.Buffer)
	if err := subChunk(r, w, blocks); err != nil {
		return nil, err
	}
	_, _ = io.Copy(w, r)
	return w.Bytes(), nil
}

// LevelChunk maps the declared sub-chunks and all biome sections, preserving
// border blocks and block NBT verbatim. Cache blobs are deliberately not accepted.
func LevelChunk(data []byte, count uint32, sections int, blocks, biomes Mapper) ([]byte, error) {
	if sections < 1 || sections > 64 || count > uint32(sections) {
		return nil, fmt.Errorf("invalid chunk dimensions: %d sub-chunks, %d biome sections", count, sections)
	}
	r, w := bytes.NewReader(data), new(bytes.Buffer)
	for range count {
		if err := subChunk(r, w, blocks); err != nil {
			return nil, err
		}
	}
	for i := range sections {
		if err := palette(r, w, biomes, i > 0); err != nil {
			return nil, fmt.Errorf("biome section %d: %w", i, err)
		}
	}
	if r.Len() == 0 {
		return nil, fmt.Errorf("missing border-block count")
	}
	_, _ = io.Copy(w, r)
	return w.Bytes(), nil
}

func subChunk(r *bytes.Reader, w *bytes.Buffer, blocks Mapper) error {
	var header [3]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	if header[0] != 9 {
		return fmt.Errorf("unsupported sub-chunk version %d", header[0])
	}
	w.Write(header[:])
	for range header[1] {
		if err := palette(r, w, blocks, false); err != nil {
			return err
		}
	}
	return nil
}

func palette(r *bytes.Reader, w *bytes.Buffer, mapID Mapper, reuse bool) error {
	header, err := r.ReadByte()
	if err != nil {
		return err
	}
	if header == 255 && reuse {
		w.WriteByte(header)
		return nil
	}
	bits := int(header >> 1)
	if header&1 == 0 || !validBits(bits) {
		return fmt.Errorf("invalid network palette header %d", header)
	}
	var words []uint32
	if bits != 0 {
		words = make([]uint32, (4096+32/bits-1)/(32/bits))
		if err := binary.Read(r, binary.LittleEndian, words); err != nil {
			return err
		}
	}
	count := int32(1)
	if bits != 0 {
		if err := protocol.Varint32(r, &count); err != nil {
			return err
		}
	}
	if count < 1 || count > 4096 || (bits != 0 && count > 1<<bits) {
		return fmt.Errorf("invalid palette length %d", count)
	}
	remap, values := make([]uint32, count), make([]uint32, 0, count)
	indices := make(map[uint32]uint32, count)
	for i := range remap {
		var value int32
		if err := protocol.Varint32(r, &value); err != nil {
			return err
		}
		mapped, ok := mapID(uint32(value))
		if !ok {
			return fmt.Errorf("unknown palette ID %d", uint32(value))
		}
		index, exists := indices[mapped]
		if !exists {
			index = uint32(len(values))
			indices[mapped] = index
			values = append(values, mapped)
		}
		remap[i] = index
	}
	newBits := 0
	for _, candidate := range []int{0, 1, 2, 3, 4, 5, 6, 8, 16} {
		if len(values) <= 1<<candidate {
			newBits = candidate
			break
		}
	}
	var packed []uint32
	if newBits != 0 {
		packed = make([]uint32, (4096+32/newBits-1)/(32/newBits))
	}
	for cell := range 4096 {
		old := uint32(0)
		if bits != 0 {
			old = (words[cell/(32/bits)] >> (uint(cell%(32/bits)) * uint(bits))) & ((1 << bits) - 1)
		}
		if old >= uint32(len(remap)) {
			return fmt.Errorf("palette index %d outside length %d", old, len(remap))
		}
		if newBits != 0 {
			packed[cell/(32/newBits)] |= remap[old] << (uint(cell%(32/newBits)) * uint(newBits))
		}
	}
	w.WriteByte(byte(newBits<<1 | 1))
	_ = binary.Write(w, binary.LittleEndian, packed)
	if newBits != 0 {
		_ = protocol.WriteVarint32(w, int32(len(values)))
	}
	for _, value := range values {
		_ = protocol.WriteVarint32(w, int32(value))
	}
	return nil
}

func validBits(bits int) bool {
	switch bits {
	case 0, 1, 2, 3, 4, 5, 6, 8, 16:
		return true
	}
	return false
}
