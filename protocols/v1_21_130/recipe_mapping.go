package v1_21_130

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func mapCraftingData(pk *packet.CraftingData, items *mapping.ItemMapper, blocks *mapping.BlockMapper) *packet.CraftingData {
	cloned := *pk
	cloned.ShapedRecipes = filterShapedRecipes(pk.ShapedRecipes, items, blocks)
	cloned.ShapelessRecipes = filterShapelessRecipes(pk.ShapelessRecipes, items, blocks)
	cloned.UserDataShapelessRecipes = filterUserDataRecipes(pk.UserDataShapelessRecipes, items, blocks)
	cloned.ShapelessChemistryRecipes = nil
	cloned.ShapedChemistryRecipes = nil
	cloned.SmithingTransformRecipes = filterSmithingTransformRecipes(pk.SmithingTransformRecipes, items, blocks)
	cloned.SmithingTrimRecipes = filterSmithingTrimRecipes(pk.SmithingTrimRecipes, items)
	cloned.PotionRecipes = filterPotionRecipes(pk.PotionRecipes, items)
	cloned.PotionContainerChangeRecipes = filterPotionContainerRecipes(pk.PotionContainerChangeRecipes, items)
	cloned.MaterialReducers = nil
	cloned.MultiRecipes = append([]protocol.MultiRecipe(nil), pk.MultiRecipes...)
	return &cloned
}

func filterShapedRecipes(recipes []protocol.ShapedRecipe, items *mapping.ItemMapper, blocks *mapping.BlockMapper) []protocol.ShapedRecipe {
	result := make([]protocol.ShapedRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		input, ok := mapDescriptorCounts(recipe.Input, items)
		output, outputOK := mapItemStacks(recipe.Output, items, blocks, toTarget)
		if !ok || !outputOK {
			continue
		}
		recipe.Input, recipe.Output = input, output
		result = append(result, recipe)
	}
	return result
}

func filterShapelessRecipes(recipes []protocol.ShapelessRecipe, items *mapping.ItemMapper, blocks *mapping.BlockMapper) []protocol.ShapelessRecipe {
	result := make([]protocol.ShapelessRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		input, ok := mapDescriptorCounts(recipe.Input, items)
		output, outputOK := mapItemStacks(recipe.Output, items, blocks, toTarget)
		if !ok || !outputOK {
			continue
		}
		recipe.Input, recipe.Output = input, output
		result = append(result, recipe)
	}
	return result
}

func filterUserDataRecipes(recipes []protocol.UserDataShapelessRecipe, items *mapping.ItemMapper, blocks *mapping.BlockMapper) []protocol.UserDataShapelessRecipe {
	result := make([]protocol.UserDataShapelessRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		mapped := filterShapelessRecipes([]protocol.ShapelessRecipe{recipe.ShapelessRecipe}, items, blocks)
		if len(mapped) == 1 {
			recipe.ShapelessRecipe = mapped[0]
			result = append(result, recipe)
		}
	}
	return result
}

func filterSmithingTransformRecipes(recipes []protocol.SmithingTransformRecipe, items *mapping.ItemMapper, blocks *mapping.BlockMapper) []protocol.SmithingTransformRecipe {
	result := make([]protocol.SmithingTransformRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		template, okOne := mapDescriptorCount(recipe.Template, items)
		base, okTwo := mapDescriptorCount(recipe.Base, items)
		addition, okThree := mapDescriptorCount(recipe.Addition, items)
		output, okFour := mapItemStack(recipe.Result, items, blocks, toTarget)
		if !okOne || !okTwo || !okThree || !okFour {
			continue
		}
		recipe.Template, recipe.Base, recipe.Addition, recipe.Result = template, base, addition, output
		result = append(result, recipe)
	}
	return result
}

func filterSmithingTrimRecipes(recipes []protocol.SmithingTrimRecipe, items *mapping.ItemMapper) []protocol.SmithingTrimRecipe {
	result := make([]protocol.SmithingTrimRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		template, okOne := mapDescriptorCount(recipe.Template, items)
		base, okTwo := mapDescriptorCount(recipe.Base, items)
		addition, okThree := mapDescriptorCount(recipe.Addition, items)
		if !okOne || !okTwo || !okThree {
			continue
		}
		recipe.Template, recipe.Base, recipe.Addition = template, base, addition
		result = append(result, recipe)
	}
	return result
}

func filterPotionRecipes(recipes []protocol.PotionRecipe, items *mapping.ItemMapper) []protocol.PotionRecipe {
	result := make([]protocol.PotionRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		input, okOne := items.NativeToTarget(recipe.InputPotionID)
		reagent, okTwo := items.NativeToTarget(recipe.ReagentItemID)
		output, okThree := items.NativeToTarget(recipe.OutputPotionID)
		if !okOne || !okTwo || !okThree {
			continue
		}
		recipe.InputPotionID, recipe.ReagentItemID, recipe.OutputPotionID = input, reagent, output
		result = append(result, recipe)
	}
	return result
}

func filterPotionContainerRecipes(recipes []protocol.PotionContainerChangeRecipe, items *mapping.ItemMapper) []protocol.PotionContainerChangeRecipe {
	result := make([]protocol.PotionContainerChangeRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		input, okOne := items.NativeToTarget(recipe.InputItemID)
		reagent, okTwo := items.NativeToTarget(recipe.ReagentItemID)
		output, okThree := items.NativeToTarget(recipe.OutputItemID)
		if !okOne || !okTwo || !okThree {
			continue
		}
		recipe.InputItemID, recipe.ReagentItemID, recipe.OutputItemID = input, reagent, output
		result = append(result, recipe)
	}
	return result
}

func mapDescriptorCounts(values []protocol.ItemDescriptorCount, items *mapping.ItemMapper) ([]protocol.ItemDescriptorCount, bool) {
	mapped := make([]protocol.ItemDescriptorCount, len(values))
	for index, value := range values {
		var ok bool
		mapped[index], ok = mapDescriptorCount(value, items)
		if !ok {
			return nil, false
		}
	}
	return mapped, true
}

func mapDescriptorCount(value protocol.ItemDescriptorCount, items *mapping.ItemMapper) (protocol.ItemDescriptorCount, bool) {
	switch descriptor := value.Descriptor.(type) {
	case nil, *protocol.InvalidItemDescriptor, *protocol.ItemTagItemDescriptor, *protocol.MoLangItemDescriptor:
		return value, true
	case *protocol.DefaultItemDescriptor:
		if descriptor.Name == "" {
			return value, true
		}
		_, ok := items.TargetRuntimeID(descriptor.Name)
		return value, ok
	default:
		return protocol.ItemDescriptorCount{}, false
	}
}

func mapItemStacks(stacks []protocol.ItemStack, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) ([]protocol.ItemStack, bool) {
	mapped := make([]protocol.ItemStack, len(stacks))
	for index, stack := range stacks {
		var ok bool
		mapped[index], ok = mapItemStack(stack, items, blocks, direction)
		if !ok {
			return nil, false
		}
	}
	return mapped, true
}
