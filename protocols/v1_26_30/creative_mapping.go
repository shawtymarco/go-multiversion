package v1_26_30

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/df-multiversion/mapping"
)

func (p Protocol) targetCreativeContent(current *packet.CreativeContent, items *mapping.ItemMapper) (*packet.CreativeContent, error) {
	groups := make([]protocol.CreativeGroup, len(current.Groups))
	for index, group := range current.Groups {
		groups[index] = group
		icon, ok := p.targetCreativeStack(group.Icon, items)
		if !ok {
			icon = protocol.ItemStack{}
		}
		groups[index].Icon = icon
	}
	creativeItems := make([]protocol.CreativeItem, 0, len(current.Items))
	for _, item := range current.Items {
		if item.GroupIndex >= uint32(len(groups)) {
			return nil, fmt.Errorf("creative item %d has invalid group index %d", item.CreativeItemNetworkID, item.GroupIndex)
		}
		stack, ok := p.targetCreativeStack(item.Item, items)
		if !ok {
			continue
		}
		item.Item = stack
		// CreativeItemNetworkID is deliberately preserved. Dragonfly resolves
		// client CraftCreative requests against its filtered native creative
		// slice, so assigning IDs from the raw historical blob selects unrelated
		// native items whenever either list skipped an entry.
		creativeItems = append(creativeItems, item)
	}
	return &packet.CreativeContent{Groups: groups, Items: creativeItems}, nil
}

func (p Protocol) targetCreativeStack(stack protocol.ItemStack, items *mapping.ItemMapper) (protocol.ItemStack, bool) {
	if stack.NetworkID == 0 {
		return stack, true
	}
	networkID, ok := items.NativeToTarget(stack.NetworkID)
	if !ok {
		return protocol.ItemStack{}, false
	}
	mapped := stack
	mapped.NetworkID = networkID
	if stack.BlockRuntimeID > 0 {
		blockID, valid, exact := p.runtime.blocks.MapNative(uint32(stack.BlockRuntimeID))
		if !valid || !exact {
			return protocol.ItemStack{}, false
		}
		mapped.BlockRuntimeID = int32(blockID)
	}
	return mapped, true
}
