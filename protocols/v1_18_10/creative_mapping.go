package v1_18_10

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

func (p Protocol) targetCreativeContent(current *packet.CreativeContent, items *mapping.ItemMapper) *packet.CreativeContent {
	creativeItems := make([]protocol.CreativeItem, 0, len(current.Items))
	for _, item := range current.Items {
		stack, ok := p.targetCreativeStack(item.Item, items)
		if !ok {
			continue
		}
		item.Item = stack
		item.GroupIndex = 0
		// The handle must stay aligned with Dragonfly's native filtered list.
		creativeItems = append(creativeItems, item)
	}
	return &packet.CreativeContent{Items: creativeItems}
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
