package v1_21_110

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBiomeDefinitionList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BiomeDefinitionList)
	protocol.FuncIOSlice(io.directional(), &pk.BiomeDefinitions, marshalBiomeDefinition844)
	protocol.FuncSlice(io.directional(), &pk.StringList, io.String)
}

func marshalBiomeDefinition844(raw protocol.IO, definition *protocol.BiomeDefinition) {
	io := asWireIO(raw)
	io.Int16(&definition.NameIndex)
	io.Int16(&definition.BiomeID)
	io.Float32(&definition.Temperature)
	io.Float32(&definition.Downfall)
	io.Float32(&definition.FoliageSnow)
	io.Float32(&definition.Depth)
	io.Float32(&definition.Scale)
	io.Int32(&definition.MapWaterColour)
	io.Bool(&definition.Rain)
	protocol.OptionalFunc(io.directional(), &definition.Tags, func(tags *[]uint16) {
		protocol.FuncSlice(io.directional(), tags, io.Uint16)
	})
	protocol.OptionalFunc(io.directional(), &definition.ChunkGeneration, func(generation *protocol.BiomeChunkGeneration) {
		marshalBiomeChunkGeneration844(io, generation)
	})
}

func marshalBiomeChunkGeneration844(io *wireIO, generation *protocol.BiomeChunkGeneration) {
	protocol.OptionalMarshaler(io.directional(), &generation.Climate)
	protocol.OptionalFunc(io.directional(), &generation.ConsolidatedFeatures, func(features *[]protocol.BiomeConsolidatedFeature) {
		protocol.Slice(io.directional(), features)
	})
	protocol.OptionalMarshaler(io.directional(), &generation.MountainParameters)
	protocol.OptionalFunc(io.directional(), &generation.SurfaceMaterialAdjustments, func(adjustments *[]protocol.BiomeElementData) {
		protocol.Slice(io.directional(), adjustments)
	})
	protocol.OptionalMarshaler(io.directional(), &generation.SurfaceMaterials)
	io.Bool(&generation.HasDefaultOverworldSurface)
	io.Bool(&generation.HasSwampSurface)
	io.Bool(&generation.HasFrozenOceanSurface)
	io.Bool(&generation.HasEndSurface)
	protocol.OptionalMarshaler(io.directional(), &generation.MesaSurface)
	protocol.OptionalMarshaler(io.directional(), &generation.CappedSurface)
	protocol.OptionalMarshaler(io.directional(), &generation.OverworldRules)
	protocol.OptionalMarshaler(io.directional(), &generation.MultiNoiseRules)
	protocol.OptionalFunc(io.directional(), &generation.LegacyRules, func(rules *[]protocol.BiomeConditionalTransformation) {
		protocol.Slice(io.directional(), rules)
	})
	if io.reading {
		generation.ReplacementsData = protocol.Optional[[]protocol.BiomeReplacementData]{}
		generation.VillageType = protocol.Optional[uint8]{}
		generation.SurfaceBuilder = protocol.Optional[protocol.BiomeSurfaceBuilder]{}
		generation.SubsurfaceBuilder = protocol.Optional[protocol.BiomeSurfaceBuilder]{}
	}
}
