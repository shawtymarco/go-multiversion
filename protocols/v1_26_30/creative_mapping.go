package v1_26_30

import (
	"fmt"
	"maps"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	v1001 "github.com/shawtymarco/df-multiversion/data/v1001"
	"github.com/shawtymarco/df-multiversion/mapping"
)

func (p Protocol) targetCreativeContent(items *mapping.ItemMapper) (*packet.CreativeContent, error) {
	groups := make([]protocol.CreativeGroup, len(p.runtime.creative.Groups))
	for index, group := range p.runtime.creative.Groups {
		name := group.Name
		if name == "" {
			name = fmt.Sprint("anon", index)
		}
		icon, _ := p.targetCreativeStack(group.Icon, items)
		groups[index] = protocol.CreativeGroup{Category: byte(group.Category), Name: name, Icon: icon}
	}
	creativeItems := make([]protocol.CreativeItem, 0, len(p.runtime.creative.Items))
	for _, item := range p.runtime.creative.Items {
		if item.GroupIndex < 0 || item.GroupIndex >= int32(len(groups)) {
			return nil, fmt.Errorf("creative item %s has invalid group index %d", item.Name, item.GroupIndex)
		}
		stack, ok := p.targetCreativeStack(item, items)
		if !ok {
			continue
		}
		creativeItems = append(creativeItems, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(len(creativeItems)) + 1,
			Item:                  stack,
			GroupIndex:            uint32(item.GroupIndex),
		})
	}
	return &packet.CreativeContent{Groups: groups, Items: creativeItems}, nil
}

func (p Protocol) targetCreativeStack(entry v1001.CreativeItem, items *mapping.ItemMapper) (protocol.ItemStack, bool) {
	if entry.Name == "" {
		return protocol.ItemStack{}, true
	}
	networkID, ok := items.TargetRuntimeID(entry.Name)
	if !ok {
		return protocol.ItemStack{}, false
	}
	var blockRuntimeID int32
	if len(entry.BlockProperties) != 0 {
		blockID, ok := p.runtime.blocks.TargetRuntimeID(entry.Name, entry.BlockProperties)
		if !ok {
			return protocol.ItemStack{}, false
		}
		blockRuntimeID = int32(blockID)
	}
	nbtData := maps.Clone(entry.NBT)
	delete(nbtData, "Damage")
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     networkID,
			MetadataValue: uint32(uint16(entry.Meta)),
		},
		Count:          1,
		BlockRuntimeID: blockRuntimeID,
		NBTData:        nbtData,
	}, true
}
