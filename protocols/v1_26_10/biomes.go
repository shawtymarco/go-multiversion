package v1_26_10

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBiomeDefinitionList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BiomeDefinitionList)
	protocol.FuncIOSlice(io.directional(), &pk.BiomeDefinitions, func(raw protocol.IO, definition *protocol.BiomeDefinition) {
		marshalBiomeDefinition944(asWireIO(raw), definition)
	})
	protocol.FuncSlice(io.directional(), &pk.StringList, io.String)
}

func marshalBiomeDefinition944(io *wireIO, value *protocol.BiomeDefinition) {
	io.Int16(&value.NameIndex)
	io.Int16(&value.BiomeID)
	io.Float32(&value.Temperature)
	io.Float32(&value.Downfall)
	io.Float32(&value.FoliageSnow)
	io.Float32(&value.Depth)
	io.Float32(&value.Scale)
	io.Int32(&value.MapWaterColour)
	io.Bool(&value.Rain)
	protocol.OptionalFunc(io.directional(), &value.Tags, func(tags *[]uint16) {
		protocol.FuncSlice(io.directional(), tags, io.Uint16)
	})
	protocol.OptionalFunc(io.directional(), &value.ChunkGeneration, func(generation *protocol.BiomeChunkGeneration) {
		marshalBiomeChunkGeneration944(io, generation)
	})
}

func marshalBiomeChunkGeneration944(io *wireIO, value *protocol.BiomeChunkGeneration) {
	protocol.OptionalMarshaler(io.directional(), &value.Climate)
	protocol.OptionalFunc(io.directional(), &value.ConsolidatedFeatures, func(values *[]protocol.BiomeConsolidatedFeature) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalMarshaler(io.directional(), &value.MountainParameters)
	protocol.OptionalFunc(io.directional(), &value.SurfaceMaterialAdjustments, func(values *[]protocol.BiomeElementData) {
		protocol.Slice(io.directional(), values)
	})
	builder, _ := value.SurfaceBuilder.Value()
	protocol.OptionalMarshaler(io.directional(), &builder.SurfaceMaterials)
	io.Bool(&builder.HasDefaultOverworldSurface)
	io.Bool(&builder.HasSwampSurface)
	io.Bool(&builder.HasFrozenOceanSurface)
	io.Bool(&builder.HasEndSurface)
	protocol.OptionalMarshaler(io.directional(), &builder.MesaSurface)
	protocol.OptionalMarshaler(io.directional(), &builder.CappedSurface)
	protocol.OptionalMarshaler(io.directional(), &value.OverworldRules)
	protocol.OptionalMarshaler(io.directional(), &value.MultiNoiseRules)
	protocol.OptionalFunc(io.directional(), &value.LegacyRules, func(values *[]protocol.BiomeConditionalTransformation) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalFunc(io.directional(), &value.ReplacementsData, func(values *[]protocol.BiomeReplacementData) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalFunc(io.directional(), &value.VillageType, io.Uint8)
	if io.reading {
		builder.NoiseGradientSurface = protocol.Optional[protocol.BiomeNoiseGradientSurface]{}
		value.SurfaceBuilder = protocol.Option(builder)
		value.SubsurfaceBuilder = protocol.Optional[protocol.BiomeSurfaceBuilder]{}
	}
}
