package v1_16_100

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	itemDescriptorInvalid = iota
	itemDescriptorDefault
	itemDescriptorMoLang
	itemDescriptorTag
	itemDescriptorDeferred
	itemDescriptorComplexAlias
)

func marshalItemDescriptorCount(io *wireIO, value *protocol.ItemDescriptorCount) {
	if io.reading {
		var kind uint8
		io.Uint8(&kind)
		switch kind {
		case itemDescriptorInvalid:
			value.Descriptor = &protocol.InvalidItemDescriptor{}
		case itemDescriptorDefault:
			var networkID, metadata int16
			io.Int16(&networkID)
			if networkID != 0 {
				io.Int16(&metadata)
			}
			value.Descriptor = &protocol.DefaultItemDescriptor{MetadataValue: int32(metadata)}
		case itemDescriptorMoLang:
			var expression string
			var version uint8
			io.String(&expression)
			io.Uint8(&version)
			value.Descriptor = &protocol.MoLangItemDescriptor{Expression: expression, Version: int16(version)}
		case itemDescriptorTag:
			var tag string
			io.String(&tag)
			value.Descriptor = &protocol.ItemTagItemDescriptor{Tag: tag}
		case itemDescriptorDeferred:
			var name string
			var metadata int16
			io.String(&name)
			io.Int16(&metadata)
			value.Descriptor = &protocol.DefaultItemDescriptor{Name: name, MetadataValue: int32(metadata)}
		case itemDescriptorComplexAlias:
			var name string
			io.String(&name)
			value.Descriptor = &protocol.DefaultItemDescriptor{Name: name}
		default:
			io.UnknownEnumOption(kind, "item descriptor type")
			return
		}
		io.Varint32(&value.Count)
		return
	}

	var kind uint8
	switch value.Descriptor.(type) {
	case nil, *protocol.InvalidItemDescriptor:
		kind = itemDescriptorInvalid
	case *protocol.DefaultItemDescriptor:
		kind = itemDescriptorDeferred
	case *protocol.MoLangItemDescriptor:
		kind = itemDescriptorMoLang
	case *protocol.ItemTagItemDescriptor:
		kind = itemDescriptorTag
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value.Descriptor), "item descriptor type")
		return
	}
	io.Uint8(&kind)
	switch descriptor := value.Descriptor.(type) {
	case nil, *protocol.InvalidItemDescriptor:
	case *protocol.DefaultItemDescriptor:
		io.String(&descriptor.Name)
		metadata := int16(descriptor.MetadataValue)
		io.Int16(&metadata)
	case *protocol.MoLangItemDescriptor:
		io.String(&descriptor.Expression)
		version := uint8(descriptor.Version)
		io.Uint8(&version)
	case *protocol.ItemTagItemDescriptor:
		io.String(&descriptor.Tag)
	}
	io.Varint32(&value.Count)
}

func marshalItemInstance(io *wireIO, value *protocol.ItemInstance) {
	io.Varint32(&value.StackNetworkID)
	marshalItem(io, &value.Stack)
}

func marshalItemInstanceNew(io *wireIO, value *protocol.ItemInstance) {
	marshalItemInstance(io, value)
}

func marshalItem(io *wireIO, value *protocol.ItemStack) {
	io.Varint32(&value.NetworkID)
	if value.NetworkID == 0 {
		if io.reading {
			*value = protocol.ItemStack{}
		}
		return
	}
	aux := int32(int16(value.MetadataValue))<<8 | int32(value.Count&0xff)
	io.Varint32(&aux)
	if io.reading {
		value.MetadataValue = uint32(uint16(aux >> 8))
		value.Count = uint16(aux & 0xff)
		value.BlockRuntimeID = 0
	}
	marshalItemUserData(io, value)
}

func marshalStackRequestItem(io *wireIO, value *protocol.StackRequestItem) {
	var networkID int32
	blockRuntimeID := value.BlockRuntimeID
	if !io.reading {
		if io.runtime == nil {
			io.InvalidValue(value.Identifier, "stack request item identifier", "protocol 419 item mapping is not configured")
			return
		}
		itemMapper := io.runtime.currentItemMapper()
		if itemMapper == nil {
			io.InvalidValue(value.Identifier, "stack request item identifier", "native item registry has not been observed")
			return
		}
		var ok bool
		networkID, ok = itemMapper.TargetRuntimeID(value.Identifier)
		if !ok {
			io.InvalidValue(value.Identifier, "stack request item identifier", "identifier is absent from protocol 419")
			return
		}
		if blockRuntimeID > 0 {
			mapped, _, _ := io.runtime.blocks.MapNative(uint32(blockRuntimeID))
			blockRuntimeID = int32(mapped)
		}
	}
	stack := protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     networkID,
			MetadataValue: value.MetadataValue,
		},
		BlockRuntimeID: blockRuntimeID,
		Count:          value.Count,
		NBTData:        value.NBTData,
		CanBePlacedOn:  value.CanBePlacedOn,
		CanBreak:       value.CanBreak,
		BlockingTick:   value.BlockingTick,
	}
	marshalItem(io, &stack)
	if io.reading {
		itemMapper := io.runtime.currentItemMapper()
		if itemMapper == nil {
			io.InvalidValue(stack.NetworkID, "stack request item network ID", "native item registry has not been observed")
			return
		}
		identifier, ok := itemMapper.TargetIdentifier(stack.NetworkID)
		if !ok {
			io.InvalidValue(stack.NetworkID, "stack request item network ID", "unknown protocol 419 item")
			return
		}
		value.Identifier = identifier
		value.MetadataValue = stack.MetadataValue
		if stack.BlockRuntimeID > 0 {
			mapped, ok := io.runtime.blocks.TargetToNative(uint32(stack.BlockRuntimeID))
			if !ok {
				io.InvalidValue(stack.BlockRuntimeID, "stack request block runtime ID", "unknown protocol 975 block")
				return
			}
			value.BlockRuntimeID = int32(mapped)
		} else {
			value.BlockRuntimeID = stack.BlockRuntimeID
		}
		value.Count = stack.Count
		value.NBTData = stack.NBTData
		value.CanBePlacedOn = stack.CanBePlacedOn
		value.CanBreak = stack.CanBreak
		value.BlockingTick = stack.BlockingTick
	}
}

func marshalItemUserData(io *wireIO, stack *protocol.ItemStack) {
	marker := int16(0)
	if !io.reading && len(stack.NBTData) != 0 {
		marker = -1
	}
	io.Int16(&marker)
	if marker == -1 {
		version := uint8(1)
		io.Uint8(&version)
		if version != 1 {
			io.UnknownEnumOption(version, "item user data version")
			return
		}
		io.NBT(&stack.NBTData, nbt.NetworkLittleEndian)
	} else if marker > 0 {
		io.NBT(&stack.NBTData, nbt.LittleEndian)
	} else if io.reading {
		stack.NBTData = nil
	}
	marshalItemStringSlice(io, &stack.CanBePlacedOn)
	marshalItemStringSlice(io, &stack.CanBreak)
	if stack.NetworkID == io.ShieldID() {
		var blockingTick int64
		io.Varint64(&blockingTick)
	}
	if io.reading {
		stack.BlockingTick = 0
	}
}

func marshalItemStringSlice(io *wireIO, values *[]string) {
	count := int32(len(*values))
	io.Varint32(&count)
	if count < 0 {
		io.InvalidValue(count, "item string count", "count must not be negative")
		return
	}
	if io.reading {
		*values = make([]string, count)
	}
	for index := range *values {
		io.String(&(*values)[index])
	}
}

func marshalMaterialReducer(io *wireIO, reducer *protocol.MaterialReducer) {
	mix := (reducer.InputItem.NetworkID << 16) | int32(reducer.InputItem.MetadataValue)
	io.Varint32(&mix)
	if io.reading {
		reducer.InputItem = protocol.ItemType{NetworkID: mix << 16, MetadataValue: uint32(mix & 0x7fff)}
	}
	protocol.Slice(io.directional(), &reducer.Outputs)
}
