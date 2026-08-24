package v1_21_100

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBiomeDefinitionList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BiomeDefinitionList)
	protocol.FuncIOSlice(io.directional(), &pk.BiomeDefinitions, func(raw protocol.IO, definition *protocol.BiomeDefinition) {
		marshalBiomeDefinition827(raw, definition, pk.StringList)
	})
	protocol.FuncSlice(io.directional(), &pk.StringList, io.String)
}

func marshalBiomeDefinition827(raw protocol.IO, definition *protocol.BiomeDefinition, strings []string) {
	io := asWireIO(raw)
	io.Int16(&definition.NameIndex)
	io.Int16(&definition.BiomeID)
	io.Float32(&definition.Temperature)
	io.Float32(&definition.Downfall)
	redSpores, blueSpores, ash, whiteAsh := legacyBiomeDensities(definition.NameIndex, strings)
	io.Float32(&redSpores)
	io.Float32(&blueSpores)
	io.Float32(&ash)
	io.Float32(&whiteAsh)
	if io.reading {
		definition.FoliageSnow = 0
	}
	io.Float32(&definition.Depth)
	io.Float32(&definition.Scale)
	io.Int32(&definition.MapWaterColour)
	io.Bool(&definition.Rain)
	protocol.OptionalFunc(io.directional(), &definition.Tags, func(tags *[]uint16) {
		protocol.FuncSlice(io.directional(), tags, io.Uint16)
	})
	protocol.OptionalFunc(io.directional(), &definition.ChunkGeneration, func(generation *protocol.BiomeChunkGeneration) {
		marshalBiomeChunkGeneration827(io, generation)
	})
}

func marshalBiomeChunkGeneration827(io *wireIO, generation *protocol.BiomeChunkGeneration) {
	protocol.OptionalFunc(io.directional(), &generation.Climate, func(climate *protocol.BiomeClimate) {
		io.Float32(&climate.Temperature)
		io.Float32(&climate.Downfall)
		var redSpores, blueSpores, ash, whiteAsh float32
		io.Float32(&redSpores)
		io.Float32(&blueSpores)
		io.Float32(&ash)
		io.Float32(&whiteAsh)
		io.Float32(&climate.SnowAccumulationMin)
		io.Float32(&climate.SnowAccumulationMax)
	})
	protocol.OptionalFunc(io.directional(), &generation.ConsolidatedFeatures, func(features *[]protocol.BiomeConsolidatedFeature) {
		protocol.Slice(io.directional(), features)
	})
	protocol.OptionalMarshaler(io.directional(), &generation.MountainParameters)
	protocol.OptionalFunc(io.directional(), &generation.SurfaceMaterialAdjustments, func(adjustments *[]protocol.BiomeElementData) {
		protocol.Slice(io.directional(), adjustments)
	})
	protocol.OptionalMarshaler(io.directional(), &generation.SurfaceMaterials)
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
		generation.HasDefaultOverworldSurface = false
		generation.ReplacementsData = protocol.Optional[[]protocol.BiomeReplacementData]{}
		generation.VillageType = protocol.Optional[uint8]{}
		generation.SurfaceBuilder = protocol.Optional[protocol.BiomeSurfaceBuilder]{}
		generation.SubsurfaceBuilder = protocol.Optional[protocol.BiomeSurfaceBuilder]{}
	}
}

func legacyBiomeDensities(nameIndex int16, strings []string) (redSpores, blueSpores, ash, whiteAsh float32) {
	if nameIndex < 0 || int(nameIndex) >= len(strings) {
		return
	}
	switch strings[nameIndex] {
	case "crimson_forest":
		redSpores = 0.25
	case "warped_forest":
		blueSpores = 0.25
	case "soulsand_valley":
		ash = 0.05
	case "basalt_deltas":
		whiteAsh = 2
	}
	return
}
