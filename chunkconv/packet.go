package chunkconv

import (
	"fmt"
	"math"
	"slices"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// BlockMapper is implemented by adapters that map native block runtime IDs to
// their own, such as protocols/v1_26_50.Protocol.
type BlockMapper interface {
	MapBlockRuntimeID(uint32) (uint32, bool)
}

// BiomeMapper is implemented by adapters whose biome runtime IDs differ.
type BiomeMapper interface {
	MapBiomeRuntimeID(uint32) (uint32, bool)
}

// sections returns the biome sections of a vanilla dimension.
func sections(dimension int32) (int, error) {
	switch dimension {
	case packet.DimensionOverworld:
		return (319 - -64 + 1) / 16, nil
	case packet.DimensionNether:
		return 128 / 16, nil
	case packet.DimensionEnd:
		return 256 / 16, nil
	}
	return 0, fmt.Errorf("unknown chunk dimension %d", dimension)
}

// Packet translates the palettes of a native LevelChunk or SubChunk for the
// adapter proto, which must implement BlockMapper; anything else, or a proto
// without a mapper, is returned unchanged. The input is never modified. Cached
// chunks carry blob hashes instead of palettes and are refused.
func Packet(pk packet.Packet, proto any) (packet.Packet, error) {
	mapper, ok := proto.(BlockMapper)
	if !ok {
		return pk, nil
	}
	blocks := Mapper(mapper.MapBlockRuntimeID)
	biomes := Mapper(func(id uint32) (uint32, bool) { return id, true })
	if b, ok := proto.(BiomeMapper); ok {
		biomes = b.MapBiomeRuntimeID
	}
	switch source := pk.(type) {
	case *packet.LevelChunk:
		if source.CacheEnabled || len(source.BlobHashes) != 0 {
			return nil, fmt.Errorf("unexpected cached LevelChunk")
		}
		n, err := sections(source.Dimension)
		if err != nil {
			return nil, err
		}
		count := source.SubChunkCount
		if count >= math.MaxUint32-1 {
			// Sub-chunk request mode: the payload holds only biomes and borders.
			count = 0
		}
		cloned := *source
		cloned.RawPayload, err = LevelChunk(source.RawPayload, count, n, blocks, biomes)
		if err != nil {
			return nil, err
		}
		return &cloned, nil
	case *packet.SubChunk:
		if source.CacheEnabled {
			return nil, fmt.Errorf("unexpected cached SubChunk")
		}
		cloned := *source
		cloned.SubChunkEntries = slices.Clone(source.SubChunkEntries)
		for i, entry := range cloned.SubChunkEntries {
			if _, exists := entry.BlobHash.Value(); exists {
				return nil, fmt.Errorf("unexpected sub-chunk blob hash")
			}
			if entry.Result != protocol.SubChunkResultSuccess {
				continue
			}
			payload, exists := entry.RawPayload.Value()
			if !exists {
				return nil, fmt.Errorf("successful sub-chunk has no payload")
			}
			mapped, err := SubChunk(payload, blocks)
			if err != nil {
				return nil, err
			}
			cloned.SubChunkEntries[i].RawPayload = protocol.Option(mapped)
		}
		return &cloned, nil
	}
	return pk, nil
}
