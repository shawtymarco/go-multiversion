package v1_26_30

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
	recipeSmithingTransform
	recipeSmithingTrim
)

func marshalCraftingData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CraftingData)
	if io.reading {
		var count uint32
		io.Varuint32(&count)
		for range count {
			var recipeType int32
			io.Varint32(&recipeType)
			switch recipeType {
			case recipeShapeless:
				var value protocol.ShapelessRecipe
				marshalShapelessRecipe(io, &value)
				pk.ShapelessRecipes = append(pk.ShapelessRecipes, value)
			case recipeShaped:
				var value protocol.ShapedRecipe
				marshalShapedRecipe(io, &value)
				pk.ShapedRecipes = append(pk.ShapedRecipes, value)
			case recipeMulti:
				var value protocol.MultiRecipe
				marshalMultiRecipe(io, &value)
				pk.MultiRecipes = append(pk.MultiRecipes, value)
			case recipeUserDataShapeless:
				var value protocol.UserDataShapelessRecipe
				marshalShapelessRecipe(io, &value.ShapelessRecipe)
				pk.UserDataShapelessRecipes = append(pk.UserDataShapelessRecipes, value)
			case recipeShapelessChemistry:
				var value protocol.ShapelessChemistryRecipe
				marshalShapelessRecipe(io, &value.ShapelessRecipe)
				pk.ShapelessChemistryRecipes = append(pk.ShapelessChemistryRecipes, value)
			case recipeShapedChemistry:
				var value protocol.ShapedChemistryRecipe
				marshalShapedRecipe(io, &value.ShapedRecipe)
				pk.ShapedChemistryRecipes = append(pk.ShapedChemistryRecipes, value)
			case recipeSmithingTransform:
				var value protocol.SmithingTransformRecipe
				marshalSmithingTransformRecipe(io, &value)
				pk.SmithingTransformRecipes = append(pk.SmithingTransformRecipes, value)
			case recipeSmithingTrim:
				var value protocol.SmithingTrimRecipe
				marshalSmithingTrimRecipe(io, &value)
				pk.SmithingTrimRecipes = append(pk.SmithingTrimRecipes, value)
			case recipeFurnace, recipeFurnaceData:
				marshalDiscardedFurnaceRecipe(io, recipeType == recipeFurnaceData)
			default:
				io.UnknownEnumOption(recipeType, "crafting data recipe type")
				return
			}
		}
	} else {
		count := len(pk.ShapelessRecipes) + len(pk.ShapedRecipes) + len(pk.MultiRecipes) + len(pk.UserDataShapelessRecipes) +
			len(pk.ShapelessChemistryRecipes) + len(pk.ShapedChemistryRecipes) + len(pk.SmithingTransformRecipes) + len(pk.SmithingTrimRecipes)
		count32 := uint32(count)
		io.Varuint32(&count32)
		for i := range pk.ShapelessRecipes {
			writeRecipeType(io, recipeShapeless)
			marshalShapelessRecipe(io, &pk.ShapelessRecipes[i])
		}
		for i := range pk.ShapedRecipes {
			writeRecipeType(io, recipeShaped)
			marshalShapedRecipe(io, &pk.ShapedRecipes[i])
		}
		for i := range pk.MultiRecipes {
			writeRecipeType(io, recipeMulti)
			marshalMultiRecipe(io, &pk.MultiRecipes[i])
		}
		for i := range pk.UserDataShapelessRecipes {
			writeRecipeType(io, recipeUserDataShapeless)
			marshalShapelessRecipe(io, &pk.UserDataShapelessRecipes[i].ShapelessRecipe)
		}
		for i := range pk.ShapelessChemistryRecipes {
			writeRecipeType(io, recipeShapelessChemistry)
			marshalShapelessRecipe(io, &pk.ShapelessChemistryRecipes[i].ShapelessRecipe)
		}
		for i := range pk.ShapedChemistryRecipes {
			writeRecipeType(io, recipeShapedChemistry)
			marshalShapedRecipe(io, &pk.ShapedChemistryRecipes[i].ShapedRecipe)
		}
		for i := range pk.SmithingTransformRecipes {
			writeRecipeType(io, recipeSmithingTransform)
			marshalSmithingTransformRecipe(io, &pk.SmithingTransformRecipes[i])
		}
		for i := range pk.SmithingTrimRecipes {
			writeRecipeType(io, recipeSmithingTrim)
			marshalSmithingTrimRecipe(io, &pk.SmithingTrimRecipes[i])
		}
	}
	protocol.Slice(io.directional(), &pk.PotionRecipes)
	protocol.Slice(io.directional(), &pk.PotionContainerChangeRecipes)
	protocol.FuncSlice(io.directional(), &pk.MaterialReducers, io.MaterialReducer)
	io.Bool(&pk.ClearRecipes)
}

func writeRecipeType(io *wireIO, value int32) { io.Varint32(&value) }

func marshalShapedRecipe(io *wireIO, recipe *protocol.ShapedRecipe) {
	io.String(&recipe.RecipeID)
	io.Varint32(&recipe.Width)
	io.Varint32(&recipe.Height)
	protocol.FuncSliceOfLen(io.directional(), uint32(recipe.Width*recipe.Height), &recipe.Input, io.ItemDescriptorCount)
	protocol.FuncSlice(io.directional(), &recipe.Output, io.Item)
	io.UUID(&recipe.UUID)
	io.String(&recipe.Block)
	io.Varint32(&recipe.Priority)
	io.Bool(&recipe.AssumeSymmetry)
	marshalLegacyRecipeUnlockRequirement(io, &recipe.UnlockRequirement)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalShapelessRecipe(io *wireIO, recipe *protocol.ShapelessRecipe) {
	io.String(&recipe.RecipeID)
	protocol.FuncSlice(io.directional(), &recipe.Input, io.ItemDescriptorCount)
	protocol.FuncSlice(io.directional(), &recipe.Output, io.Item)
	io.UUID(&recipe.UUID)
	io.String(&recipe.Block)
	io.Varint32(&recipe.Priority)
	marshalLegacyRecipeUnlockRequirement(io, &recipe.UnlockRequirement)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalLegacyRecipeUnlockRequirement(io *wireIO, optional *protocol.Optional[protocol.RecipeUnlockRequirement]) {
	requirement, _ := optional.Value()
	context := uint8(requirement.Context)
	io.Uint8(&context)
	requirement.Context = int32(context)
	if requirement.Context == protocol.RecipeUnlockContextNone {
		protocol.FuncSlice(io.directional(), &requirement.Ingredients, io.ItemDescriptorCount)
	}
	*optional = protocol.Option(requirement)
}

func marshalMultiRecipe(io *wireIO, recipe *protocol.MultiRecipe) {
	io.UUID(&recipe.UUID)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalSmithingTransformRecipe(io *wireIO, recipe *protocol.SmithingTransformRecipe) {
	io.String(&recipe.RecipeID)
	io.ItemDescriptorCount(&recipe.Template)
	io.ItemDescriptorCount(&recipe.Base)
	io.ItemDescriptorCount(&recipe.Addition)
	io.Item(&recipe.Result)
	io.String(&recipe.Block)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalSmithingTrimRecipe(io *wireIO, recipe *protocol.SmithingTrimRecipe) {
	io.String(&recipe.RecipeID)
	io.ItemDescriptorCount(&recipe.Template)
	io.ItemDescriptorCount(&recipe.Base)
	io.ItemDescriptorCount(&recipe.Addition)
	io.String(&recipe.Block)
	io.Varuint32(&recipe.RecipeNetworkID)
}

func marshalDiscardedFurnaceRecipe(io *wireIO, withMetadata bool) {
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
