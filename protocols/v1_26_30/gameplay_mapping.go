package v1_26_30

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/df-multiversion/mapping"
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
	case *packet.CreativeContent:
		if items == nil {
			return nil
		}
		creative, err := p.targetCreativeContent(items)
		if err != nil {
			return nil
		}
		return []packet.Packet{creative}
	case *packet.CraftingData:
		if items == nil {
			return nil
		}
		return []packet.Packet{mapCraftingData(current, items, p.runtime.blocks)}
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

func (p Protocol) convertGameplayToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	if p.runtime == nil {
		return []packet.Packet{pk}
	}
	items := p.runtime.currentItemMapper()
	if items == nil && conn != nil {
		items, _ = p.runtime.itemMapper(conn.GameData().Items)
	}
	switch current := pk.(type) {
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
	mapped := stack
	var ok bool
	if direction == toTarget {
		mapped.NetworkID, ok = items.NativeToTarget(stack.NetworkID)
	} else {
		mapped.NetworkID, ok = items.TargetToNative(stack.NetworkID)
	}
	if !ok {
		return protocol.ItemStack{}, false
	}
	if stack.BlockRuntimeID > 0 && blocks != nil {
		if direction == toTarget {
			blockRuntimeID, _ := blocks.NativeToTarget(uint32(stack.BlockRuntimeID))
			mapped.BlockRuntimeID = int32(blockRuntimeID)
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
