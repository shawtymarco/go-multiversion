package v1_16_100

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/itemconv"
	"github.com/shawtymarco/go-multiversion/internal/packetconv"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type mappingDirection uint8

const (
	toTarget mappingDirection = iota
	toNative
)

func (p Protocol) convertGameplayFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if p.runtime == nil {
		return []packet.Packet{pk}
	}
	if registry, ok := pk.(*packet.ItemRegistry); ok {
		items, err := p.runtime.itemMapper(registry.Items)
		if err != nil {
			return nil
		}
		return []packet.Packet{&packet.ItemRegistry{Items: items.TargetEntries()}}
	}
	items := p.runtime.currentItemMapper()
	if items == nil && conn != nil {
		items, _ = p.runtime.itemMapper(conn.GameData().Items)
	}

	switch current := pk.(type) {
	case *packet.StartGame:
		cloned := *current
		version := Version
		if conn != nil && isStableGameVersion(conn.ClientData().GameVersion) {
			version = conn.ClientData().GameVersion
		}
		cloned.GameVersion, cloned.BaseGameVersion = version, version
		cloned.GameRules = targetGameRules(current.GameRules)
		return []packet.Packet{&cloned}
	case *packet.ResourcePackStack:
		cloned := *current
		version := Version
		if conn != nil && isStableGameVersion(conn.ClientData().GameVersion) {
			version = conn.ClientData().GameVersion
		}
		cloned.BaseGameVersion = version
		cloned.TexturePacks = targetTexturePacks(current.TexturePacks)
		// The native stack enables the current-only cameras experiment. The
		// protocol 419 server stack predates it and sent no experiments here.
		cloned.Experiments = nil
		cloned.ExperimentsPreviouslyToggled = false
		return []packet.Packet{&cloned}
	case *packet.ModalFormRequest:
		cloned := *current
		// Forms carry a versioned JSON schema inside an otherwise unchanged
		// packet, so downgrade the document at the protocol boundary.
		cloned.FormData = targetFormData(current.FormData)
		return []packet.Packet{&cloned}
	case *packet.GameRulesChanged:
		cloned := *current
		cloned.GameRules = targetGameRules(current.GameRules)
		return []packet.Packet{&cloned}
	case *packet.UpdateAbilities:
		flags, actions := legacyAbilityFlags(current.AbilityData.Layers)
		return []packet.Packet{&packet.AdventureSettings{
			Flags:                  flags,
			CommandPermissionLevel: uint32(current.AbilityData.CommandPermissions),
			ActionPermissions:      actions,
			PermissionLevel:        uint32(current.AbilityData.PlayerPermissions),
			PlayerUniqueID:         current.AbilityData.EntityUniqueID,
		}}
	case *packet.CreativeContent:
		if items == nil {
			return nil
		}
		return []packet.Packet{p.targetCreativeContent(current, items)}
	case *packet.CraftingData:
		// Minecraft 1.16.100 keeps its built-in recipe catalogue. Feeding the
		// current Dragonfly catalogue into the retail client crashes it while
		// initialising the recipe book, before it can process world chunks.
		return nil
	case *packet.AddPlayer:
		cloned := *current
		if mapped, ok := mapItemInstance(current.HeldItem, items, p.runtime.blocks, toTarget); ok {
			cloned.HeldItem = mapped
		} else {
			cloned.HeldItem = protocol.ItemInstance{}
		}
		return []packet.Packet{&cloned}
	case *packet.AddItemActor:
		cloned := *current
		mapped, ok := mapItemInstance(current.Item, items, p.runtime.blocks, toTarget)
		if !ok {
			return nil
		}
		cloned.Item = mapped
		return []packet.Packet{&cloned}
	case *packet.InventoryContent:
		cloned := *current
		cloned.Content = mapItemInstances(current.Content, items, p.runtime.blocks, toTarget, true)
		cloned.StorageItem, _ = mapItemInstance(current.StorageItem, items, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	case *packet.InventorySlot:
		cloned := *current
		cloned.NewItem, _ = mapItemInstance(current.NewItem, items, p.runtime.blocks, toTarget)
		if storage, ok := current.StorageItem.Value(); ok {
			mapped, _ := mapItemInstance(storage, items, p.runtime.blocks, toTarget)
			cloned.StorageItem = protocol.Option(mapped)
		}
		return []packet.Packet{&cloned}
	case *packet.MobEquipment:
		cloned := *current
		cloned.NewItem, _ = mapItemInstance(current.NewItem, items, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	case *packet.MobArmourEquipment:
		cloned := *current
		cloned.Helmet, _ = mapItemInstance(current.Helmet, items, p.runtime.blocks, toTarget)
		cloned.Chestplate, _ = mapItemInstance(current.Chestplate, items, p.runtime.blocks, toTarget)
		cloned.Leggings, _ = mapItemInstance(current.Leggings, items, p.runtime.blocks, toTarget)
		cloned.Boots, _ = mapItemInstance(current.Boots, items, p.runtime.blocks, toTarget)
		cloned.Body, _ = mapItemInstance(current.Body, items, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	case *packet.InventoryTransaction:
		mapped, ok := mapInventoryTransaction(current, items, p.runtime.blocks, toTarget)
		if !ok {
			return nil
		}
		return []packet.Packet{mapped}
	case *packet.PlayerAuthInput:
		mapped, ok := mapPlayerAuthInput(current, items, p.runtime.blocks, toTarget)
		if !ok {
			return nil
		}
		return []packet.Packet{mapped}
	case *packet.UpdateBlock:
		cloned := *current
		cloned.NewBlockRuntimeID, _ = p.runtime.blocks.NativeToTarget(current.NewBlockRuntimeID)
		return []packet.Packet{&cloned}
	case *packet.UpdateBlockSynced:
		cloned := *current
		cloned.NewBlockRuntimeID, _ = p.runtime.blocks.NativeToTarget(current.NewBlockRuntimeID)
		return []packet.Packet{&cloned}
	case *packet.UpdateSubChunkBlocks:
		cloned := *current
		cloned.Blocks = mapBlockChangeEntries(current.Blocks, p.runtime.blocks, toTarget)
		cloned.Extra = mapBlockChangeEntries(current.Extra, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	case *packet.LevelSoundEvent:
		cloned := *current
		if _, ok := legacySoundEvents[cloned.SoundType]; !ok {
			return nil
		}
		if !packetconv.MapLevelSoundBlockRuntimeID(&cloned, func(runtimeID uint32) (uint32, bool) {
			return mapBlockRuntimeID(runtimeID, p.runtime.blocks, toTarget)
		}) {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.LevelEvent:
		cloned := *current
		mapLevelEventData(&cloned, items, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	case *packet.ActorEvent:
		cloned := *current
		mapActorEventData(&cloned, items, toTarget)
		return []packet.Packet{&cloned}
	case *packet.AddActor:
		cloned := *current
		mapFallingBlockMetadata(&cloned, p.runtime.blocks, toTarget)
		return []packet.Packet{&cloned}
	default:
		return []packet.Packet{pk}
	}
}

func targetGameRules(rules []protocol.GameRule) []protocol.GameRule {
	target := make([]protocol.GameRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Name == "locatorBar" {
			continue
		}
		target = append(target, rule)
	}
	return target
}

func targetTexturePacks(packs []protocol.StackResourcePack) []protocol.StackResourcePack {
	target := make([]protocol.StackResourcePack, 0, len(packs))
	for _, pack := range packs {
		switch pack.UUID {
		case "d34cfa4b-2ad1-453d-a0db-668b429a3ea0", // 1.26.40
			"b41c2785-c512-4a49-af56-3a87afd47c57", // 1.21.30
			"a4df0cb3-17be-4163-88d7-fcf7002b935d", // 1.21.20
			"d19adffe-a2e1-4b02-8436-ca4583368c89", // 1.21.10
			"85d5603d-2824-4b21-8044-34f441f4fce1", // 1.21.0
			"e977cd13-0a11-4618-96fb-03dfe9c43608", // 1.20.60
			"0674721c-a0aa-41a1-9ba8-1ed33ea3e7ed": // 1.20.50
			continue
		}
		target = append(target, pack)
	}
	return target
}

func (p Protocol) convertGameplayToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if p.runtime == nil {
		return []packet.Packet{pk}
	}
	items := p.runtime.currentItemMapper()
	if items == nil && conn != nil {
		items, _ = p.runtime.itemMapper(conn.GameData().Items)
	}
	switch current := pk.(type) {
	case *packet.Emote:
		cloned := *current
		if conn != nil {
			cloned.XUID = conn.IdentityData().XUID
			cloned.PlatformID = conn.ClientData().PlatformOnlineID
		}
		return []packet.Packet{&cloned}
	case *packet.InventoryContent:
		cloned := *current
		mapped := mapItemInstances(current.Content, items, p.runtime.blocks, toNative, false)
		if mapped == nil && len(current.Content) != 0 {
			return nil
		}
		cloned.Content = mapped
		var ok bool
		cloned.StorageItem, ok = mapItemInstance(current.StorageItem, items, p.runtime.blocks, toNative)
		if !ok {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.InventorySlot:
		cloned := *current
		var ok bool
		cloned.NewItem, ok = mapItemInstance(current.NewItem, items, p.runtime.blocks, toNative)
		if !ok {
			return nil
		}
		if storage, present := current.StorageItem.Value(); present {
			mapped, ok := mapItemInstance(storage, items, p.runtime.blocks, toNative)
			if !ok {
				return nil
			}
			cloned.StorageItem = protocol.Option(mapped)
		}
		return []packet.Packet{&cloned}
	case *packet.MobEquipment:
		cloned := *current
		var ok bool
		cloned.NewItem, ok = mapItemInstance(current.NewItem, items, p.runtime.blocks, toNative)
		if !ok {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.InventoryTransaction:
		mapped, ok := mapInventoryTransaction(current, items, p.runtime.blocks, toNative)
		if !ok {
			return nil
		}
		return []packet.Packet{mapped}
	case *packet.PlayerAuthInput:
		mapped, ok := mapPlayerAuthInput(current, items, p.runtime.blocks, toNative)
		if !ok {
			return nil
		}
		return []packet.Packet{mapped}
	case *packet.UpdateBlock:
		cloned := *current
		var ok bool
		cloned.NewBlockRuntimeID, ok = p.runtime.blocks.TargetToNative(current.NewBlockRuntimeID)
		if !ok {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.UpdateBlockSynced:
		cloned := *current
		var ok bool
		cloned.NewBlockRuntimeID, ok = p.runtime.blocks.TargetToNative(current.NewBlockRuntimeID)
		if !ok {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.UpdateSubChunkBlocks:
		cloned := *current
		cloned.Blocks = mapBlockChangeEntries(current.Blocks, p.runtime.blocks, toNative)
		cloned.Extra = mapBlockChangeEntries(current.Extra, p.runtime.blocks, toNative)
		if (cloned.Blocks == nil && len(current.Blocks) != 0) || (cloned.Extra == nil && len(current.Extra) != 0) {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.LevelSoundEvent:
		cloned := *current
		if !packetconv.MapLevelSoundBlockRuntimeID(&cloned, func(runtimeID uint32) (uint32, bool) {
			return mapBlockRuntimeID(runtimeID, p.runtime.blocks, toNative)
		}) {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.LevelEvent:
		cloned := *current
		if !mapLevelEventData(&cloned, items, p.runtime.blocks, toNative) {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.ActorEvent:
		cloned := *current
		if !mapActorEventData(&cloned, items, toNative) {
			return nil
		}
		return []packet.Packet{&cloned}
	case *packet.AddActor:
		cloned := *current
		if !mapFallingBlockMetadata(&cloned, p.runtime.blocks, toNative) {
			return nil
		}
		return []packet.Packet{&cloned}
	default:
		return []packet.Packet{pk}
	}
}

func mapItemStack(stack protocol.ItemStack, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (protocol.ItemStack, bool) {
	if stack.NetworkID == 0 {
		return stack, true
	}
	if items == nil {
		return protocol.ItemStack{}, false
	}
	if direction == toNative && stack.BlockRuntimeID == 0 && blocks != nil {
		if targetSemanticName, found := items.TargetIdentifier(stack.NetworkID); found {
			targetName, _ := items.TargetWireIdentifier(targetSemanticName)
			if targetState, found := targetBlockItemState(targetName, stack.MetadataValue); found {
				if targetBlockRuntimeID, found := blocks.TargetRuntimeID(targetState.Name, targetState.Properties); found {
					if nativeBlockRuntimeID, found := blocks.TargetToNative(targetBlockRuntimeID); found {
						if nativeState, found := blocks.NativeState(nativeBlockRuntimeID); found {
							if nativeItemRuntimeID, found := items.NativeRuntimeID(nativeState.Name); found {
								mapped := stack
								mapped.NetworkID = nativeItemRuntimeID
								mapped.MetadataValue = 0
								mapped.BlockRuntimeID = int32(nativeBlockRuntimeID)
								return mapped, true
							}
						}
					}
				}
			}
		}
	}
	mapped := stack
	var ok bool
	if direction == toTarget {
		if mapped, ok = itemconv.DowngradeLegacySpawnEgg(stack, items); !ok {
			mapped = stack
			mapped.NetworkID, ok = items.NativeToTarget(stack.NetworkID)
		}
	} else {
		if mapped, ok = itemconv.UpgradeLegacySpawnEgg(stack, items); !ok {
			mapped = stack
			mapped.NetworkID, ok = items.TargetToNative(stack.NetworkID)
		}
	}
	if stack.BlockRuntimeID > 0 && blocks != nil {
		if blockNetworkID, found := blockItemNetworkID(stack.BlockRuntimeID, items, blocks, direction); found {
			mapped.NetworkID, ok = blockNetworkID, true
		}
	}
	if !ok {
		return protocol.ItemStack{}, false
	}
	if stack.BlockRuntimeID > 0 && blocks != nil {
		if direction == toTarget {
			blockRuntimeID, valid, exact := blocks.MapNative(uint32(stack.BlockRuntimeID))
			if !valid || !exact {
				return protocol.ItemStack{}, false
			}
			targetState, found := blocks.TargetState(blockRuntimeID)
			if !found {
				return protocol.ItemStack{}, false
			}
			metadata, found := targetBlockItemMeta(targetState)
			if !found {
				return protocol.ItemStack{}, false
			}
			mapped.MetadataValue = uint32(uint16(metadata))
			mapped.BlockRuntimeID = 0
		} else {
			blockRuntimeID, found := blocks.TargetToNative(uint32(stack.BlockRuntimeID))
			if !found {
				return protocol.ItemStack{}, false
			}
			mapped.BlockRuntimeID = int32(blockRuntimeID)
		}
	}
	return mapped, true
}

func blockItemNetworkID(blockRuntimeID int32, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (int32, bool) {
	if direction == toTarget {
		targetRuntimeID, valid, exact := blocks.MapNative(uint32(blockRuntimeID))
		if !valid || !exact {
			return 0, false
		}
		state, ok := blocks.TargetState(targetRuntimeID)
		if !ok {
			return 0, false
		}
		return items.TargetWireRuntimeID(state.Name)
	}
	nativeRuntimeID, ok := blocks.TargetToNative(uint32(blockRuntimeID))
	if !ok {
		return 0, false
	}
	state, ok := blocks.NativeState(nativeRuntimeID)
	if !ok {
		return 0, false
	}
	return items.NativeRuntimeID(state.Name)
}

func mapItemInstance(instance protocol.ItemInstance, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (protocol.ItemInstance, bool) {
	mapped := instance
	stack, ok := mapItemStack(instance.Stack, items, blocks, direction)
	if !ok {
		return protocol.ItemInstance{}, false
	}
	mapped.Stack = stack
	return mapped, true
}

func mapItemInstances(instances []protocol.ItemInstance, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection, hideUnknown bool) []protocol.ItemInstance {
	mapped := make([]protocol.ItemInstance, len(instances))
	for index, instance := range instances {
		converted, ok := mapItemInstance(instance, items, blocks, direction)
		if !ok {
			if !hideUnknown {
				return nil
			}
			converted = protocol.ItemInstance{}
		}
		mapped[index] = converted
	}
	return mapped
}
