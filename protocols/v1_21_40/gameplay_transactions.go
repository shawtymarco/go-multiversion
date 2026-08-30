package v1_21_40

import (
	"maps"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func mapInventoryTransaction(pk *packet.InventoryTransaction, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (*packet.InventoryTransaction, bool) {
	cloned := *pk
	cloned.Actions = make([]protocol.InventoryAction, len(pk.Actions))
	for index, action := range pk.Actions {
		cloned.Actions[index] = action
		var ok bool
		cloned.Actions[index].OldItem, ok = mapItemInstance(action.OldItem, items, blocks, direction)
		if !ok {
			return nil, false
		}
		cloned.Actions[index].NewItem, ok = mapItemInstance(action.NewItem, items, blocks, direction)
		if !ok {
			return nil, false
		}
	}
	data, ok := mapInventoryTransactionData(pk.TransactionData, items, blocks, direction)
	if !ok {
		return nil, false
	}
	cloned.TransactionData = data
	return &cloned, true
}

func mapInventoryTransactionData(value protocol.InventoryTransactionData, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (protocol.InventoryTransactionData, bool) {
	switch data := value.(type) {
	case *protocol.NormalTransactionData:
		cloned := *data
		return &cloned, true
	case *protocol.MismatchTransactionData:
		cloned := *data
		return &cloned, true
	case *protocol.UseItemTransactionData:
		cloned := *data
		var ok bool
		cloned.HeldItem, ok = mapItemInstance(data.HeldItem, items, blocks, direction)
		if !ok {
			return nil, false
		}
		if data.BlockRuntimeID != 0 && blocks != nil {
			if direction == toTarget {
				cloned.BlockRuntimeID, _ = blocks.NativeToTarget(data.BlockRuntimeID)
			} else {
				cloned.BlockRuntimeID, ok = blocks.TargetToNative(data.BlockRuntimeID)
				if !ok {
					return nil, false
				}
			}
		}
		return &cloned, true
	case *protocol.UseItemOnEntityTransactionData:
		cloned := *data
		var ok bool
		cloned.HeldItem, ok = mapItemInstance(data.HeldItem, items, blocks, direction)
		return &cloned, ok
	case *protocol.ReleaseItemTransactionData:
		cloned := *data
		var ok bool
		cloned.HeldItem, ok = mapItemInstance(data.HeldItem, items, blocks, direction)
		return &cloned, ok
	default:
		return nil, false
	}
}

func mapPlayerAuthInput(pk *packet.PlayerAuthInput, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) (*packet.PlayerAuthInput, bool) {
	cloned := *pk
	if interaction, ok := pk.ItemInteractionData.Value(); ok {
		mapped, valid := mapInventoryTransactionData(&interaction, items, blocks, direction)
		if !valid {
			return nil, false
		}
		cloned.ItemInteractionData = protocol.Option(*mapped.(*protocol.UseItemTransactionData))
	}
	return &cloned, true
}

func mapBlockChangeEntries(entries []protocol.BlockChangeEntry, blocks *mapping.BlockMapper, direction mappingDirection) []protocol.BlockChangeEntry {
	mapped := append([]protocol.BlockChangeEntry(nil), entries...)
	for index := range mapped {
		if direction == toTarget {
			mapped[index].BlockRuntimeID, _ = blocks.NativeToTarget(entries[index].BlockRuntimeID)
		} else {
			var ok bool
			mapped[index].BlockRuntimeID, ok = blocks.TargetToNative(entries[index].BlockRuntimeID)
			if !ok {
				return nil
			}
		}
	}
	return mapped
}

func mapLevelEventData(pk *packet.LevelEvent, items *mapping.ItemMapper, blocks *mapping.BlockMapper, direction mappingDirection) bool {
	switch pk.EventType {
	case packet.LevelEventParticlesDestroyBlock:
		mapped, ok := mapBlockRuntimeID(uint32(pk.EventData), blocks, direction)
		pk.EventData = int32(mapped)
		return ok
	case packet.LevelEventParticlesCrackBlock:
		face := uint32(pk.EventData) & 0xff000000
		mapped, ok := mapBlockRuntimeID(uint32(pk.EventData)&0x00ffffff, blocks, direction)
		pk.EventData = int32(face | mapped&0x00ffffff)
		return ok
	case packet.LevelEventParticleLegacyEvent | 14:
		if items == nil {
			return false
		}
		itemID, meta := int32(uint32(pk.EventData)>>16), uint32(pk.EventData)&0xffff
		var mapped int32
		var ok bool
		if direction == toTarget {
			mapped, ok = items.NativeToTarget(itemID)
		} else {
			mapped, ok = items.TargetToNative(itemID)
		}
		pk.EventData = int32(uint32(mapped)<<16 | meta)
		return ok
	default:
		return true
	}
}

func mapActorEventData(pk *packet.ActorEvent, items *mapping.ItemMapper, direction mappingDirection) bool {
	if pk.EventType != packet.ActorEventFeed {
		return true
	}
	if items == nil {
		return false
	}
	itemID, meta := int32(uint32(pk.EventData)>>16), uint32(pk.EventData)&0xffff
	var mapped int32
	var ok bool
	if direction == toTarget {
		mapped, ok = items.NativeToTarget(itemID)
	} else {
		mapped, ok = items.TargetToNative(itemID)
	}
	pk.EventData = int32(uint32(mapped)<<16 | meta)
	return ok
}

func mapFallingBlockMetadata(pk *packet.AddActor, blocks *mapping.BlockMapper, direction mappingDirection) bool {
	if pk.EntityType != "minecraft:falling_block" {
		return true
	}
	value, ok := pk.EntityMetadata[protocol.EntityDataKeyVariant].(int32)
	if !ok {
		return true
	}
	pk.EntityMetadata = maps.Clone(pk.EntityMetadata)
	mapped, found := mapBlockRuntimeID(uint32(value), blocks, direction)
	if found {
		pk.EntityMetadata[protocol.EntityDataKeyVariant] = int32(mapped)
	}
	return found
}

func mapBlockRuntimeID(runtimeID uint32, blocks *mapping.BlockMapper, direction mappingDirection) (uint32, bool) {
	if blocks == nil {
		return 0, false
	}
	if direction == toTarget {
		mapped, _ := blocks.NativeToTarget(runtimeID)
		return mapped, true
	}
	return blocks.TargetToNative(runtimeID)
}
