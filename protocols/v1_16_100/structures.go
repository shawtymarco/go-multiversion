package v1_16_100

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalStructureBlockUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StructureBlockUpdate)
	marshalUnsignedBlockPos(io, &pk.Position)
	io.String(&pk.StructureName)
	io.String(&pk.DataField)
	io.Bool(&pk.IncludePlayers)
	io.Bool(&pk.ShowBoundingBox)
	io.Varint32(&pk.StructureBlockType)
	marshalStructureSettings(io, &pk.Settings)
	redstoneMode := int32(pk.RedstoneSaveMode)
	io.Varint32(&redstoneMode)
	pk.RedstoneSaveMode = uint8(redstoneMode)
	io.Bool(&pk.ShouldTrigger)
	if io.reading {
		pk.FilteredStructureName = ""
		pk.Waterlogged = false
	}
}

func marshalStructureTemplateDataRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StructureTemplateDataRequest)
	io.String(&pk.StructureName)
	marshalUnsignedBlockPos(io, &pk.Position)
	marshalStructureSettings(io, &pk.Settings)
	io.Uint8(&pk.RequestType)
}

func marshalStructureSettings(io *wireIO, settings *protocol.StructureSettings) {
	io.String(&settings.PaletteName)
	io.Bool(&settings.IgnoreEntities)
	io.Bool(&settings.IgnoreBlocks)
	marshalUnsignedBlockPos(io, &settings.Size)
	marshalUnsignedBlockPos(io, &settings.Offset)
	io.Varint64(&settings.LastEditingPlayerUniqueID)
	io.Uint8(&settings.Rotation)
	io.Uint8(&settings.Mirror)
	io.Float32(&settings.Integrity)
	io.Uint32(&settings.Seed)
	io.Vec3(&settings.Pivot)
	if io.reading {
		settings.AnimationMode = 0
		settings.AnimationDuration = 0
		settings.AllowNonTickingChunks = false
	}
}
