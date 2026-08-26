package v1_18_10

import (
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/packetio"
)

func marshalLevelChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelChunk)
	io.ChunkPos(&pk.Position)
	count := pk.SubChunkCount
	if !io.reading {
		if limit, ok := pk.SubChunkLimit.Value(); ok {
			if limit < 0 {
				count = math.MaxUint32
			} else {
				count = math.MaxUint32 - 1
			}
		}
	}
	io.Varuint32(&count)
	switch count {
	case math.MaxUint32:
		pk.SubChunkLimit = protocol.Option(int32(-1))
	case math.MaxUint32 - 1:
		limit, _ := pk.SubChunkLimit.Value()
		highest := uint16(limit)
		io.Uint16(&highest)
		pk.SubChunkLimit = protocol.Option(int32(highest))
	default:
		pk.SubChunkCount = count
		if io.reading {
			pk.SubChunkLimit = protocol.Optional[int32]{}
		}
	}
	io.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		protocol.FuncSlice(io.directional(), &pk.BlobHashes, io.Uint64)
	} else if io.reading {
		pk.BlobHashes = nil
	}
	io.ByteSlice(&pk.RawPayload)
	if io.reading {
		pk.Dimension = packet.DimensionOverworld
	}
}

func marshalSubChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunk)
	io.Bool(&pk.CacheEnabled)
	io.Varint32(&pk.Dimension)
	marshalLegacySubChunkPos(io, &pk.Position)
	count := uint32(len(pk.SubChunkEntries))
	io.Uint32(&count)
	if io.reading {
		pk.SubChunkEntries = make([]protocol.SubChunkEntry, count)
	}
	for index := range pk.SubChunkEntries {
		marshalSubChunkEntry(io, &pk.SubChunkEntries[index], pk.CacheEnabled)
	}
}

func marshalSubChunkEntry(io *wireIO, entry *protocol.SubChunkEntry, cacheEnabled bool) {
	io.Int8(&entry.Offset[0])
	io.Int8(&entry.Offset[1])
	io.Int8(&entry.Offset[2])
	io.Uint8(&entry.Result)
	if !cacheEnabled || entry.Result != protocol.SubChunkResultSuccessAllAir {
		payload, _ := entry.RawPayload.Value()
		io.ByteSlice(&payload)
		entry.RawPayload = protocol.Option(payload)
	} else if io.reading {
		entry.RawPayload = protocol.Optional[[]byte]{}
	}
	io.Uint8(&entry.HeightMapType)
	if entry.HeightMapType == protocol.HeightMapDataHasData {
		heightMap, _ := entry.HeightMapData.Value()
		if io.reading {
			heightMap = make([]int8, 256)
		}
		for index := 0; index < 256; index++ {
			io.Int8(&heightMap[index])
		}
		entry.HeightMapData = protocol.Option(heightMap)
	} else if io.reading {
		entry.HeightMapData = protocol.Optional[[]int8]{}
	}
	if cacheEnabled {
		blobHash, _ := entry.BlobHash.Value()
		io.Uint64(&blobHash)
		entry.BlobHash = protocol.Option(blobHash)
	} else if io.reading {
		entry.BlobHash = protocol.Optional[uint64]{}
	}
	if io.reading {
		entry.RenderHeightMapType = protocol.HeightMapDataNone
		entry.RenderHeightMapData = protocol.Optional[[]int8]{}
	}
}

func marshalSubChunkRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunkRequest)
	io.Varint32(&pk.Dimension)
	marshalLegacySubChunkPos(io, &pk.Position)
	count := uint32(len(pk.Offsets))
	io.Uint32(&count)
	packetio.SubChunkOffsets(io.directional(), count, &pk.Offsets)
}

func marshalLegacySubChunkPos(io *wireIO, position *protocol.SubChunkPos) {
	io.Varint32(&position[0])
	io.Varint32(&position[1])
	io.Varint32(&position[2])
}
