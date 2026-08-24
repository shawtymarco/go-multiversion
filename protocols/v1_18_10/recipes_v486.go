package v1_18_10

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	recipeShapeless int32 = iota
	recipeShaped
	recipeFurnace
	recipeFurnaceData
	recipeMulti
	recipeUserDataShapeless
	recipeShapelessChemistry
	recipeShapedChemistry
)

func writeRecipeType(io *wireIO, value int32) { io.Varint32(&value) }

func marshalMultiRecipe(io *wireIO, recipe *protocol.MultiRecipe) {
	io.UUID(&recipe.UUID)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalCraftingDataV486(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CraftingData)
	if io.reading {
		marshalCraftingDataRead(io, pk)
	} else {
		marshalCraftingDataWrite(io, pk)
	}
	protocol.Slice(io.directional(), &pk.PotionRecipes)
	protocol.Slice(io.directional(), &pk.PotionContainerChangeRecipes)
	var materialReducers []protocol.MaterialReducer
	protocol.FuncSlice(io.directional(), &materialReducers, io.MaterialReducer)
	io.Bool(&pk.ClearRecipes)
}

func marshalCraftingDataWrite(io *wireIO, pk *packet.CraftingData) {
	count := uint32(len(pk.ShapelessRecipes) + len(pk.ShapedRecipes) + len(pk.MultiRecipes) + len(pk.UserDataShapelessRecipes))
	io.Varuint32(&count)
	for index := range pk.ShapelessRecipes {
		writeRecipeType(io, recipeShapeless)
		marshalShapelessRecipeV486(io, &pk.ShapelessRecipes[index])
	}
	for index := range pk.ShapedRecipes {
		writeRecipeType(io, recipeShaped)
		marshalShapedRecipeV486(io, &pk.ShapedRecipes[index])
	}
	for index := range pk.MultiRecipes {
		writeRecipeType(io, recipeMulti)
		marshalMultiRecipe(io, &pk.MultiRecipes[index])
	}
	for index := range pk.UserDataShapelessRecipes {
		writeRecipeType(io, recipeUserDataShapeless)
		marshalShapelessRecipeV486(io, &pk.UserDataShapelessRecipes[index].ShapelessRecipe)
	}
}

func marshalCraftingDataRead(io *wireIO, pk *packet.CraftingData) {
	var count uint32
	io.Varuint32(&count)
	for range count {
		var recipeType int32
		io.Varint32(&recipeType)
		switch recipeType {
		case recipeShapeless:
			var recipe protocol.ShapelessRecipe
			marshalShapelessRecipeV486(io, &recipe)
			pk.ShapelessRecipes = append(pk.ShapelessRecipes, recipe)
		case recipeShaped:
			var recipe protocol.ShapedRecipe
			marshalShapedRecipeV486(io, &recipe)
			pk.ShapedRecipes = append(pk.ShapedRecipes, recipe)
		case recipeMulti:
			var recipe protocol.MultiRecipe
			marshalMultiRecipe(io, &recipe)
			pk.MultiRecipes = append(pk.MultiRecipes, recipe)
		case recipeUserDataShapeless:
			var recipe protocol.UserDataShapelessRecipe
			marshalShapelessRecipeV486(io, &recipe.ShapelessRecipe)
			pk.UserDataShapelessRecipes = append(pk.UserDataShapelessRecipes, recipe)
		case recipeShapelessChemistry:
			var recipe protocol.ShapelessRecipe
			marshalShapelessRecipeV486(io, &recipe)
		case recipeShapedChemistry:
			var recipe protocol.ShapedRecipe
			marshalShapedRecipeV486(io, &recipe)
		case recipeFurnace, recipeFurnaceData:
			marshalDiscardedFurnaceRecipeV486(io, recipeType == recipeFurnaceData)
		default:
			io.UnknownEnumOption(recipeType, "crafting recipe type")
			return
		}
	}
}

func marshalShapedRecipeV486(io *wireIO, recipe *protocol.ShapedRecipe) {
	io.String(&recipe.RecipeID)
	io.Varint32(&recipe.Width)
	io.Varint32(&recipe.Height)
	if io.reading {
		recipe.Input = make([]protocol.ItemDescriptorCount, recipe.Width*recipe.Height)
	}
	for index := range recipe.Input {
		marshalRecipeIngredientV486(io, &recipe.Input[index])
	}
	protocol.FuncSlice(io.directional(), &recipe.Output, io.Item)
	io.UUID(&recipe.UUID)
	io.String(&recipe.Block)
	io.Varint32(&recipe.Priority)
	io.Varuint32(&recipe.RecipeNetworkID)
	if io.reading {
		recipe.AssumeSymmetry = false
		recipe.UnlockRequirement = protocol.Optional[protocol.RecipeUnlockRequirement]{}
	}
}

func marshalShapelessRecipeV486(io *wireIO, recipe *protocol.ShapelessRecipe) {
	io.String(&recipe.RecipeID)
	protocol.FuncIOSlice(io.directional(), &recipe.Input, func(raw protocol.IO, ingredient *protocol.ItemDescriptorCount) {
		marshalRecipeIngredientV486(asWireIO(raw), ingredient)
	})
	protocol.FuncSlice(io.directional(), &recipe.Output, io.Item)
	io.UUID(&recipe.UUID)
	io.String(&recipe.Block)
	io.Varint32(&recipe.Priority)
	io.Varuint32(&recipe.RecipeNetworkID)
	if io.reading {
		recipe.UnlockRequirement = protocol.Optional[protocol.RecipeUnlockRequirement]{}
	}
}

func marshalRecipeIngredientV486(io *wireIO, ingredient *protocol.ItemDescriptorCount) {
	var networkID int32
	var metadata int32
	if !io.reading {
		switch descriptor := ingredient.Descriptor.(type) {
		case nil, *protocol.InvalidItemDescriptor:
		case *protocol.DefaultItemDescriptor:
			if io.runtime == nil || io.runtime.currentItemMapper() == nil {
				io.InvalidValue(descriptor.Name, "recipe ingredient", "protocol 486 item mapping is not configured")
				return
			}
			var ok bool
			networkID, ok = io.runtime.currentItemMapper().TargetRuntimeID(descriptor.Name)
			if !ok {
				io.InvalidValue(descriptor.Name, "recipe ingredient", "item is absent from protocol 486")
				return
			}
			metadata = descriptor.MetadataValue
		default:
			io.InvalidValue(descriptor, "recipe ingredient", "descriptor has no protocol-486 numeric representation")
			return
		}
	}
	io.Varint32(&networkID)
	if networkID == 0 {
		if io.reading {
			ingredient.Descriptor = &protocol.InvalidItemDescriptor{}
			ingredient.Count = 0
		}
		return
	}
	io.Varint32(&metadata)
	io.Varint32(&ingredient.Count)
	if io.reading {
		if io.runtime == nil || io.runtime.currentItemMapper() == nil {
			io.InvalidValue(networkID, "recipe ingredient", "protocol 486 item mapping is not configured")
			return
		}
		name, ok := io.runtime.currentItemMapper().TargetIdentifier(networkID)
		if !ok {
			io.InvalidValue(networkID, "recipe ingredient", "unknown protocol-486 item")
			return
		}
		ingredient.Descriptor = &protocol.DefaultItemDescriptor{Name: name, MetadataValue: metadata}
	}
}

func marshalDiscardedFurnaceRecipeV486(io *wireIO, withMetadata bool) {
	var networkID int32
	io.Varint32(&networkID)
	if withMetadata {
		var metadata int32
		io.Varint32(&metadata)
	}
	var output protocol.ItemStack
	io.Item(&output)
	var block string
	io.String(&block)
}
