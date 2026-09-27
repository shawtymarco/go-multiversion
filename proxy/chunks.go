package proxy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type paletteMapper func(uint32) (uint32, error)

// rewriteStorage decodes a bounded network palette, remaps semantic values and
// repacks every index. Reuse markers are expanded for both 1.18 client families.
func rewriteStorage(r *bytes.Reader, mapValue paletteMapper, previous []byte) ([]byte, error) {
	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if header == 0xff {
		if previous == nil {
			return nil, fmt.Errorf("palette reuse without preceding biome")
		}
		return append([]byte(nil), previous...), nil
	}
	size := int(header >> 1)
	if header&1 == 0 {
		return nil, fmt.Errorf("disk palette in network payload")
	}
	if size != 0 && size != 1 && size != 2 && size != 3 && size != 4 && size != 5 && size != 6 && size != 8 && size != 16 {
		return nil, fmt.Errorf("invalid palette bits %d", size)
	}
	var words []uint32
	if size != 0 {
		per := 32 / size
		count := (4096 + per - 1) / per
		words = make([]uint32, count)
		if err := binary.Read(r, binary.LittleEndian, words); err != nil {
			return nil, err
		}
	}
	count := int32(1)
	if size != 0 {
		if err := protocol.Varint32(r, &count); err != nil {
			return nil, err
		}
	}
	if count <= 0 || count > 4096 {
		return nil, fmt.Errorf("invalid palette count %d", count)
	}
	values := make([]uint32, count)
	for i := range values {
		var n int32
		if err := protocol.Varint32(r, &n); err != nil {
			return nil, err
		}
		values[i], err = mapValue(uint32(n))
		if err != nil {
			return nil, err
		}
	}
	indices := make([]uint16, 4096)
	unique := make([]uint32, 0, len(values))
	lookup := make(map[uint32]uint16, len(values))
	remap := make([]uint16, len(values))
	for i, value := range values {
		index, ok := lookup[value]
		if !ok {
			index = uint16(len(unique))
			lookup[value] = index
			unique = append(unique, value)
		}
		remap[i] = index
	}
	if size != 0 {
		per := 32 / size
		mask := uint32(1<<size) - 1
		for i := range indices {
			index := (words[i/per] >> ((i % per) * size)) & mask
			if int(index) >= len(remap) {
				return nil, fmt.Errorf("palette index %d exceeds %d", index, len(remap))
			}
			indices[i] = remap[index]
		}
	}
	newSize := bits.Len(uint(len(unique) - 1))
	if newSize == 7 {
		newSize = 8
	}
	if newSize > 8 {
		newSize = 16
	}
	var out bytes.Buffer
	out.WriteByte(byte(newSize<<1 | 1))
	if newSize != 0 {
		per := 32 / newSize
		packed := make([]uint32, (4096+per-1)/per)
		for i, index := range indices {
			packed[i/per] |= uint32(index) << ((i % per) * newSize)
		}
		_ = binary.Write(&out, binary.LittleEndian, packed)
		_ = protocol.WriteVarint32(&out, int32(len(unique)))
	}
	for _, value := range unique {
		_ = protocol.WriteVarint32(&out, int32(value))
	}
	return out.Bytes(), nil
}

func rewriteSubChunk(r *bytes.Reader, w *bytes.Buffer, mapValue paletteMapper) error {
	version, err := r.ReadByte()
	if err != nil {
		return err
	}
	if version != 9 && version != 8 {
		return fmt.Errorf("unsupported sub-chunk version %d", version)
	}
	layers, err := r.ReadByte()
	if err != nil {
		return err
	}
	if layers > 16 {
		return fmt.Errorf("too many sub-chunk layers %d", layers)
	}
	w.WriteByte(version)
	w.WriteByte(layers)
	if version == 9 {
		y, err := r.ReadByte()
		if err != nil {
			return err
		}
		w.WriteByte(y)
	}
	for range layers {
		storage, err := rewriteStorage(r, mapValue, nil)
		if err != nil {
			return err
		}
		w.Write(storage)
	}
	return nil
}

func rewriteLevelPayload(payload []byte, subChunks, biomeChunks int, blocks, biomes paletteMapper) ([]byte, error) {
	if subChunks < 0 || subChunks > 64 || biomeChunks < 1 || biomeChunks > 64 {
		return nil, fmt.Errorf("invalid chunk dimensions")
	}
	r := bytes.NewReader(payload)
	var out bytes.Buffer
	for range subChunks {
		if err := rewriteSubChunk(r, &out, blocks); err != nil {
			return nil, err
		}
	}
	var last []byte
	for range biomeChunks {
		storage, err := rewriteStorage(r, biomes, last)
		if err != nil {
			return nil, err
		}
		last = storage
		out.Write(storage)
	}
	// Border-block bytes and block-entity NBT retain their exact framing.
	_, _ = io.Copy(&out, r)
	return out.Bytes(), nil
}

func rewriteSubPayload(payload []byte, blocks paletteMapper) ([]byte, error) {
	r := bytes.NewReader(payload)
	var out bytes.Buffer
	if err := rewriteSubChunk(r, &out, blocks); err != nil {
		return nil, err
	}
	_, _ = io.Copy(&out, r)
	return out.Bytes(), nil
}
