package v1_18_0

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalLevelChunk(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LevelChunk)
	io.ChunkPos(&pk.Position)
	count := pk.SubChunkCount
	io.Varuint32(&count)
	pk.SubChunkCount = count
	if io.reading {
		pk.SubChunkLimit = protocol.Optional[int32]{}
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
	io.Varint32(&pk.Dimension)
	marshalLegacySubChunkPos(io, &pk.Position)
	var entry protocol.SubChunkEntry
	if !io.reading && len(pk.SubChunkEntries) != 0 {
		entry = pk.SubChunkEntries[0]
	}
	payload, _ := entry.RawPayload.Value()
	io.ByteSlice(&payload)
	entry.RawPayload = protocol.Option(payload)
	result := int32(entry.Result)
	io.Varint32(&result)
	entry.Result = byte(result)
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
	io.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		blobHash, _ := entry.BlobHash.Value()
		io.Uint64(&blobHash)
		entry.BlobHash = protocol.Option(blobHash)
	} else if io.reading {
		entry.BlobHash = protocol.Optional[uint64]{}
	}
	if io.reading {
		entry.Offset = protocol.SubChunkOffset{}
		entry.RenderHeightMapType = protocol.HeightMapDataNone
		entry.RenderHeightMapData = protocol.Optional[[]int8]{}
		pk.SubChunkEntries = []protocol.SubChunkEntry{entry}
	}
}

func marshalSubChunkRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunkRequest)
	io.Varint32(&pk.Dimension)
	position := pk.Position
	if !io.reading && len(pk.Offsets) != 0 {
		position[0] += int32(pk.Offsets[0][0])
		position[1] += int32(pk.Offsets[0][1])
		position[2] += int32(pk.Offsets[0][2])
	}
	marshalLegacySubChunkPos(io, &position)
	if io.reading {
		pk.Position = position
		pk.Offsets = []protocol.SubChunkOffset{{0, 0, 0}}
	}
}

func marshalLegacySubChunkPos(io *wireIO, position *protocol.SubChunkPos) {
	io.Varint32(&position[0])
	io.Varint32(&position[1])
	io.Varint32(&position[2])
}
