package dragonfly

import (
	"fmt"
	"maps"
	"slices"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NativeProtocol bridges the server's local block registry to the current
// client's hashed block IDs. Local runtime IDs are not vanilla palette indexes.
// Historical protocols keep their own direct registry mappings instead.
// The registry must be finalised before constructing this immutable adapter.
func NativeProtocol(registry world.BlockRegistry) (minecraft.Protocol, error) {
	p := &nativeProtocol{Protocol: minecraft.DefaultProtocol, registry: registry, nativeIDs: make(map[uint32]uint32, registry.BlockCount())}
	for id := range registry.BlockCount() {
		hash, ok := registry.RuntimeIDToHash(uint32(id))
		if !ok {
			return nil, fmt.Errorf("native block %d has no network hash", id)
		}
		if previous, exists := p.nativeIDs[hash]; exists {
			return nil, fmt.Errorf("native block hash collision: %d and %d", previous, id)
		}
		p.nativeIDs[hash] = uint32(id)
	}
	return p, nil
}

type nativeProtocol struct {
	minecraft.Protocol
	registry  world.BlockRegistry
	nativeIDs map[uint32]uint32
}

// MapBlockRuntimeID is consumed by the connection wrapper's chunk converter.
func (p *nativeProtocol) MapBlockRuntimeID(id uint32) (uint32, bool) {
	return p.registry.RuntimeIDToHash(id)
}
func (*nativeProtocol) MapBiomeRuntimeID(id uint32) (uint32, bool) { return id, true }
func (p *nativeProtocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	return p.convert(pk, false)
}
func (p *nativeProtocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	return p.convert(pk, true)
}

func (p *nativeProtocol) convert(pk packet.Packet, incoming bool) []packet.Packet {
	convert := p.registry.RuntimeIDToHash
	if incoming {
		convert = func(hash uint32) (uint32, bool) { id, ok := p.nativeIDs[hash]; return id, ok }
	}
	m := nativeBlockConversion{convert: convert, valid: true}
	result := m.packet(pk, incoming)
	if !m.valid {
		return nil
	}
	return []packet.Packet{result}
}

type nativeBlockConversion struct {
	convert func(uint32) (uint32, bool)
	valid   bool
}

func (m *nativeBlockConversion) id(id uint32) uint32 {
	mapped, ok := m.convert(id)
	m.valid = m.valid && ok
	return mapped
}

func (m *nativeBlockConversion) stack(stack protocol.ItemStack) protocol.ItemStack {
	if stack.NetworkID != 0 && stack.BlockRuntimeID != 0 {
		stack.BlockRuntimeID = int32(m.id(uint32(stack.BlockRuntimeID)))
	}
	return stack
}
func (m *nativeBlockConversion) instance(it protocol.ItemInstance) protocol.ItemInstance {
	it.Stack = m.stack(it.Stack)
	return it
}
func (m *nativeBlockConversion) instances(items []protocol.ItemInstance) []protocol.ItemInstance {
	cloned := slices.Clone(items)
	for i := range cloned {
		cloned[i] = m.instance(cloned[i])
	}
	return cloned
}
func (m *nativeBlockConversion) stacks(items []protocol.ItemStack) []protocol.ItemStack {
	cloned := slices.Clone(items)
	for i := range cloned {
		cloned[i] = m.stack(cloned[i])
	}
	return cloned
}
func (m *nativeBlockConversion) changes(entries []protocol.BlockChangeEntry) []protocol.BlockChangeEntry {
	cloned := slices.Clone(entries)
	for i := range cloned {
		cloned[i].BlockRuntimeID = m.id(cloned[i].BlockRuntimeID)
	}
	return cloned
}
func (m *nativeBlockConversion) actions(actions []protocol.InventoryAction) []protocol.InventoryAction {
	cloned := slices.Clone(actions)
	for i := range cloned {
		cloned[i].OldItem = m.instance(cloned[i].OldItem)
		cloned[i].NewItem = m.instance(cloned[i].NewItem)
	}
	return cloned
}

func (m *nativeBlockConversion) transaction(data protocol.InventoryTransactionData) protocol.InventoryTransactionData {
	switch source := data.(type) {
	case *protocol.UseItemTransactionData:
		cloned := *source
		cloned.Actions = m.actions(source.Actions)
		cloned.HeldItem = m.instance(source.HeldItem)
		if source.BlockRuntimeID != 0 || source.ActionType != protocol.UseItemActionClickAir {
			cloned.BlockRuntimeID = m.id(source.BlockRuntimeID)
		}
		return &cloned
	case *protocol.UseItemOnEntityTransactionData:
		cloned := *source
		cloned.HeldItem = m.instance(source.HeldItem)
		return &cloned
	case *protocol.ReleaseItemTransactionData:
		cloned := *source
		cloned.HeldItem = m.instance(source.HeldItem)
		return &cloned
	default:
		return data
	}
}

func (m *nativeBlockConversion) packet(pk packet.Packet, incoming bool) packet.Packet {
	switch source := pk.(type) {
	case *packet.StartGame:
		cloned := *source
		cloned.UseBlockNetworkIDHashes = !incoming
		return &cloned
	case *packet.UpdateBlock:
		cloned := *source
		cloned.NewBlockRuntimeID = m.id(source.NewBlockRuntimeID)
		return &cloned
	case *packet.UpdateBlockSynced:
		cloned := *source
		cloned.NewBlockRuntimeID = m.id(source.NewBlockRuntimeID)
		return &cloned
	case *packet.UpdateSubChunkBlocks:
		cloned := *source
		cloned.Blocks = m.changes(source.Blocks)
		cloned.Extra = m.changes(source.Extra)
		return &cloned
	case *packet.InventoryContent:
		cloned := *source
		cloned.Content = m.instances(source.Content)
		cloned.StorageItem = m.instance(source.StorageItem)
		return &cloned
	case *packet.InventorySlot:
		cloned := *source
		cloned.NewItem = m.instance(source.NewItem)
		if storage, ok := source.StorageItem.Value(); ok {
			cloned.StorageItem = protocol.Option(m.instance(storage))
		}
		return &cloned
	case *packet.MobEquipment:
		cloned := *source
		cloned.NewItem = m.instance(source.NewItem)
		return &cloned
	case *packet.MobArmourEquipment:
		cloned := *source
		cloned.Helmet, cloned.Chestplate, cloned.Leggings, cloned.Boots, cloned.Body = m.instance(source.Helmet), m.instance(source.Chestplate), m.instance(source.Leggings), m.instance(source.Boots), m.instance(source.Body)
		return &cloned
	case *packet.AddItemActor:
		cloned := *source
		cloned.Item = m.instance(source.Item)
		return &cloned
	case *packet.AddPlayer:
		cloned := *source
		cloned.HeldItem = m.instance(source.HeldItem)
		return &cloned
	case *packet.AddActor:
		if source.EntityType != "minecraft:falling_block" {
			return pk
		}
		cloned := *source
		if id, ok := source.EntityMetadata[protocol.EntityDataKeyVariant].(int32); ok {
			cloned.EntityMetadata = maps.Clone(source.EntityMetadata)
			cloned.EntityMetadata[protocol.EntityDataKeyVariant] = int32(m.id(uint32(id)))
		}
		return &cloned
	case *packet.CreativeContent:
		cloned := *source
		cloned.Groups = slices.Clone(source.Groups)
		cloned.Items = slices.Clone(source.Items)
		for i := range cloned.Groups {
			cloned.Groups[i].Icon = m.stack(cloned.Groups[i].Icon)
		}
		for i := range cloned.Items {
			cloned.Items[i].Item = m.stack(cloned.Items[i].Item)
		}
		return &cloned
	case *packet.CraftingData:
		cloned := *source
		cloned.ShapedRecipes = slices.Clone(source.ShapedRecipes)
		for i := range cloned.ShapedRecipes {
			cloned.ShapedRecipes[i].Output = m.stacks(cloned.ShapedRecipes[i].Output)
		}
		cloned.ShapelessRecipes = slices.Clone(source.ShapelessRecipes)
		for i := range cloned.ShapelessRecipes {
			cloned.ShapelessRecipes[i].Output = m.stacks(cloned.ShapelessRecipes[i].Output)
		}
		cloned.UserDataShapelessRecipes = slices.Clone(source.UserDataShapelessRecipes)
		for i := range cloned.UserDataShapelessRecipes {
			cloned.UserDataShapelessRecipes[i].Output = m.stacks(cloned.UserDataShapelessRecipes[i].Output)
		}
		cloned.ShapedChemistryRecipes = slices.Clone(source.ShapedChemistryRecipes)
		for i := range cloned.ShapedChemistryRecipes {
			cloned.ShapedChemistryRecipes[i].Output = m.stacks(cloned.ShapedChemistryRecipes[i].Output)
		}
		cloned.ShapelessChemistryRecipes = slices.Clone(source.ShapelessChemistryRecipes)
		for i := range cloned.ShapelessChemistryRecipes {
			cloned.ShapelessChemistryRecipes[i].Output = m.stacks(cloned.ShapelessChemistryRecipes[i].Output)
		}
		cloned.SmithingTransformRecipes = slices.Clone(source.SmithingTransformRecipes)
		for i := range cloned.SmithingTransformRecipes {
			cloned.SmithingTransformRecipes[i].Result = m.stack(cloned.SmithingTransformRecipes[i].Result)
		}
		return &cloned
	case *packet.InventoryTransaction:
		cloned := *source
		cloned.Actions = m.actions(source.Actions)
		cloned.TransactionData = m.transaction(source.TransactionData)
		return &cloned
	case *packet.PlayerAuthInput:
		cloned := *source
		if data, ok := source.ItemInteractionData.Value(); ok {
			cloned.ItemInteractionData = protocol.Option(*m.transaction(&data).(*protocol.UseItemTransactionData))
		}
		return &cloned
	case *packet.LevelSoundEvent:
		switch source.SoundType {
		case packet.SoundEventPlace, packet.SoundEventHit, packet.SoundEventItemUseOn,
			packet.SoundEventDoorOpen, packet.SoundEventDoorClose, packet.SoundEventTrapdoorOpen,
			packet.SoundEventTrapdoorClose, packet.SoundEventFenceGateOpen, packet.SoundEventFenceGateClose:
			if source.ExtraData == -1 {
				return pk
			}
			cloned := *source
			cloned.ExtraData = int32(m.id(uint32(source.ExtraData)))
			return &cloned
		}
	case *packet.LevelEvent:
		cloned := *source
		switch source.EventType {
		case packet.LevelEventParticlesDestroyBlock, packet.LevelEventParticlesDestroyBlockNoSound, packet.LevelEventParticleLegacyEvent | 21:
			cloned.EventData = int32(m.id(uint32(source.EventData)))
		case packet.LevelEventParticlesCrackBlock:
			if incoming {
				return pk
			}
			face := uint32(source.EventData) >> 24
			if face > 5 {
				m.valid = false
				return pk
			}
			// The face-specific events preserve all 32 bits of the block hash.
			cloned.EventType = packet.LevelEventParticlesCrackBlockDown + int32(face)
			cloned.EventData = int32(m.id(uint32(source.EventData) & 0xffffff))
		case packet.LevelEventParticlesCrackBlockDown, packet.LevelEventParticlesCrackBlockUp, packet.LevelEventParticlesCrackBlockNorth, packet.LevelEventParticlesCrackBlockSouth, packet.LevelEventParticlesCrackBlockWest, packet.LevelEventParticlesCrackBlockEast:
			cloned.EventData = int32(m.id(uint32(source.EventData)))
			if incoming {
				cloned.EventData |= (source.EventType - packet.LevelEventParticlesCrackBlockDown) << 24
				cloned.EventType = packet.LevelEventParticlesCrackBlock
			}
		}
		return &cloned
	}
	return pk
}
