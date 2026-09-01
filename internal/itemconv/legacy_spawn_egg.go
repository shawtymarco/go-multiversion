// Package itemconv contains item conversions shared by historical protocol
// families whose wire registries predate the current flattened item model.
package itemconv

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
)

const (
	legacySpawnEggIdentifier    = "minecraft:spawn_egg"
	ironGolemSpawnEggIdentifier = "minecraft:iron_golem_spawn_egg"
	ironGolemSpawnEggMetadata   = uint32(20)
)

// DowngradeLegacySpawnEgg converts a current flattened spawn egg into the
// generic identifier and metadata used by Minecraft 1.18. False is returned
// when the stack is not one of the explicitly supported legacy variants.
func DowngradeLegacySpawnEgg(stack protocol.ItemStack, items *mapping.ItemMapper) (protocol.ItemStack, bool) {
	if items == nil {
		return protocol.ItemStack{}, false
	}
	name, ok := items.NativeIdentifier(stack.NetworkID)
	if !ok || name != ironGolemSpawnEggIdentifier {
		return protocol.ItemStack{}, false
	}
	targetRuntimeID, ok := items.TargetWireRuntimeID(legacySpawnEggIdentifier)
	if !ok {
		return protocol.ItemStack{}, false
	}
	mapped := stack
	mapped.NetworkID = targetRuntimeID
	mapped.MetadataValue = ironGolemSpawnEggMetadata
	return mapped, true
}

// UpgradeLegacySpawnEgg converts the Minecraft 1.18 generic spawn egg variant
// back to the current flattened identifier. False is returned for other legacy
// spawn egg metadata values so existing unsupported-item policy remains intact.
func UpgradeLegacySpawnEgg(stack protocol.ItemStack, items *mapping.ItemMapper) (protocol.ItemStack, bool) {
	if items == nil || stack.MetadataValue != ironGolemSpawnEggMetadata {
		return protocol.ItemStack{}, false
	}
	name, ok := items.TargetIdentifier(stack.NetworkID)
	if !ok || name != legacySpawnEggIdentifier {
		return protocol.ItemStack{}, false
	}
	nativeRuntimeID, ok := items.NativeRuntimeID(ironGolemSpawnEggIdentifier)
	if !ok {
		return protocol.ItemStack{}, false
	}
	mapped := stack
	mapped.NetworkID = nativeRuntimeID
	mapped.MetadataValue = 0
	return mapped, true
}
