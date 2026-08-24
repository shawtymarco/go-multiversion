package v1_26_20

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBiomeDefinitionList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BiomeDefinitionList)
	protocol.FuncIOSlice(io.directional(), &pk.BiomeDefinitions, func(raw protocol.IO, definition *protocol.BiomeDefinition) {
		marshalBiomeDefinition975(asWireIO(raw), definition)
	})
	protocol.FuncSlice(io.directional(), &pk.StringList, io.String)
}

func marshalBiomeDefinition975(io *wireIO, value *protocol.BiomeDefinition) {
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
		marshalBiomeChunkGeneration975(io, generation)
	})
}

func marshalBiomeChunkGeneration975(io *wireIO, value *protocol.BiomeChunkGeneration) {
	protocol.OptionalMarshaler(io.directional(), &value.Climate)
	protocol.OptionalFunc(io.directional(), &value.ConsolidatedFeatures, func(values *[]protocol.BiomeConsolidatedFeature) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalMarshaler(io.directional(), &value.MountainParameters)
	protocol.OptionalFunc(io.directional(), &value.SurfaceMaterialAdjustments, func(values *[]protocol.BiomeElementData) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalMarshaler(io.directional(), &value.OverworldRules)
	protocol.OptionalMarshaler(io.directional(), &value.MultiNoiseRules)
	protocol.OptionalFunc(io.directional(), &value.LegacyRules, func(values *[]protocol.BiomeConditionalTransformation) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalFunc(io.directional(), &value.ReplacementsData, func(values *[]protocol.BiomeReplacementData) {
		protocol.Slice(io.directional(), values)
	})
	protocol.OptionalFunc(io.directional(), &value.VillageType, io.Uint8)
	protocol.OptionalFunc(io.directional(), &value.SurfaceBuilder, func(builder *protocol.BiomeSurfaceBuilder) {
		marshalBiomeSurfaceBuilder975(io, builder)
	})
	protocol.OptionalFunc(io.directional(), &value.SubsurfaceBuilder, func(builder *protocol.BiomeSurfaceBuilder) {
		marshalBiomeSurfaceBuilder975(io, builder)
	})
}

func marshalBiomeSurfaceBuilder975(io *wireIO, value *protocol.BiomeSurfaceBuilder) {
	protocol.OptionalMarshaler(io.directional(), &value.SurfaceMaterials)
	io.Bool(&value.HasDefaultOverworldSurface)
	io.Bool(&value.HasSwampSurface)
	io.Bool(&value.HasFrozenOceanSurface)
	io.Bool(&value.HasEndSurface)
	protocol.OptionalMarshaler(io.directional(), &value.MesaSurface)
	protocol.OptionalMarshaler(io.directional(), &value.CappedSurface)
	protocol.OptionalFunc(io.directional(), &value.NoiseGradientSurface, func(surface *protocol.BiomeNoiseGradientSurface) {
		marshalBiomeNoiseGradientSurface975(io, surface)
	})
}

func marshalBiomeNoiseGradientSurface975(io *wireIO, value *protocol.BiomeNoiseGradientSurface) {
	protocol.FuncSlice(io.directional(), &value.NonReplaceableBlocks, io.Uint32)
	blocks := make([]uint32, len(value.GradientBlocks))
	if !io.reading {
		for index, specifier := range value.GradientBlocks {
			blocks[index] = specifier.Block
		}
	}
	protocol.FuncSlice(io.directional(), &blocks, io.Uint32)
	if io.reading {
		value.GradientBlocks = make([]protocol.NoiseBlockSpecifier, len(blocks))
		for index, block := range blocks {
			value.GradientBlocks[index].Block = block
		}
	}
	io.String(&value.Noise.Name)
	io.Int32(&value.Noise.FirstOctave)
	protocol.FuncSlice(io.directional(), &value.Noise.Amplitudes, io.Float32)
}
