package v1_21_50

import (
	"bytes"
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
			if io.runtime != nil {
				if items := io.runtime.currentItemMapper(); items != nil {
					if resolved, ok := items.TargetSemanticIdentifier(name); ok {
						name = resolved
					}
				}
			}
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
		name := descriptor.Name
		if io.runtime != nil {
			if items := io.runtime.currentItemMapper(); items != nil {
				if wire, ok := items.TargetWireIdentifier(name); ok {
					name = wire
				}
			}
		}
		io.String(&name)
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
	stack := &value.Stack
	io.Varint32(&stack.NetworkID)
	if stack.NetworkID == 0 {
		if io.reading {
			*value = protocol.ItemInstance{}
		}
		return
	}
	marshalItemInstanceBody(io, value, false)
}

func marshalItemInstanceNew(io *wireIO, value *protocol.ItemInstance) {
	stack := &value.Stack
	id := int16(stack.NetworkID)
	io.Int16(&id)
	stack.NetworkID = int32(id)
	marshalItemInstanceBody(io, value, true)
}

func marshalItemInstanceBody(io *wireIO, value *protocol.ItemInstance, transitional bool) {
	stack := &value.Stack
	io.Uint16(&stack.Count)
	io.Varuint32(&stack.MetadataValue)
	hasNetworkID := value.StackNetworkID != 0
	io.Bool(&hasNetworkID)
	if hasNetworkID {
		if transitional {
			var reserved uint32
			io.Varuint32(&reserved)
		}
		io.Varint32(&value.StackNetworkID)
	} else if io.reading {
		value.StackNetworkID = 0
	}
	if transitional {
		blockRuntimeID := uint32(stack.BlockRuntimeID)
		io.Varuint32(&blockRuntimeID)
		stack.BlockRuntimeID = int32(blockRuntimeID)
	} else {
		io.Varint32(&stack.BlockRuntimeID)
	}
	marshalItemUserData(io, stack, transitional && stack.NetworkID == 0)
}

func marshalItem(io *wireIO, value *protocol.ItemStack) {
	io.Varint32(&value.NetworkID)
	if value.NetworkID == 0 {
		if io.reading {
			*value = protocol.ItemStack{}
		}
		return
	}
	io.Uint16(&value.Count)
	io.Varuint32(&value.MetadataValue)
	io.Varint32(&value.BlockRuntimeID)
	marshalItemUserData(io, value, false)
}

func marshalStackRequestItem(io *wireIO, value *protocol.StackRequestItem) {
	var networkID int32
	blockRuntimeID := value.BlockRuntimeID
	if !io.reading {
		if io.runtime == nil {
			io.InvalidValue(value.Identifier, "stack request item identifier", "protocol 766 item mapping is not configured")
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
			io.InvalidValue(value.Identifier, "stack request item identifier", "identifier is absent from protocol 766")
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
			io.InvalidValue(stack.NetworkID, "stack request item network ID", "unknown protocol 766 item")
			return
		}
		value.Identifier = identifier
		value.MetadataValue = stack.MetadataValue
		if stack.BlockRuntimeID > 0 {
			mapped, ok := io.runtime.blocks.TargetToNative(uint32(stack.BlockRuntimeID))
			if !ok {
				io.InvalidValue(stack.BlockRuntimeID, "stack request block runtime ID", "unknown protocol 766 block")
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

func marshalItemUserData(io *wireIO, stack *protocol.ItemStack, empty bool) {
	if io.reading {
		var extra []byte
		io.ByteSlice(&extra)
		if len(extra) == 0 {
			stack.NBTData, stack.CanBePlacedOn, stack.CanBreak, stack.BlockingTick = nil, nil, nil, 0
			return
		}
		reader := protocol.NewReader(bytes.NewReader(extra), io.ShieldID(), true)
		var length int16
		reader.Int16(&length)
		if length == -1 {
			var version uint8
			reader.Uint8(&version)
			if version != 1 {
				reader.UnknownEnumOption(version, "item user data version")
				return
			}
			reader.NBT(&stack.NBTData, nbt.LittleEndian)
		} else if length > 0 {
			reader.NBT(&stack.NBTData, nbt.LittleEndian)
		} else {
			stack.NBTData = nil
		}
		protocol.FuncSliceUint32Length(reader, &stack.CanBePlacedOn, reader.StringUTF)
		protocol.FuncSliceUint32Length(reader, &stack.CanBreak, reader.StringUTF)
		if stack.NetworkID == io.ShieldID() {
			reader.Int64(&stack.BlockingTick)
		} else {
			stack.BlockingTick = 0
		}
		return
	}

	if empty {
		var extra []byte
		io.ByteSlice(&extra)
		return
	}
	var buffer bytes.Buffer
	writer := protocol.NewWriter(&buffer, io.ShieldID())
	if len(stack.NBTData) != 0 {
		length, version := int16(-1), uint8(1)
		writer.Int16(&length)
		writer.Uint8(&version)
		writer.NBT(&stack.NBTData, nbt.LittleEndian)
	} else {
		var length int16
		writer.Int16(&length)
	}
	protocol.FuncSliceUint32Length(writer, &stack.CanBePlacedOn, writer.StringUTF)
	protocol.FuncSliceUint32Length(writer, &stack.CanBreak, writer.StringUTF)
	if stack.NetworkID == io.ShieldID() {
		writer.Int64(&stack.BlockingTick)
	}
	extra := buffer.Bytes()
	io.ByteSlice(&extra)
}

func marshalMaterialReducer(io *wireIO, reducer *protocol.MaterialReducer) {
	mix := (reducer.InputItem.NetworkID << 16) | int32(reducer.InputItem.MetadataValue)
	io.Varint32(&mix)
	if io.reading {
		reducer.InputItem = protocol.ItemType{NetworkID: mix << 16, MetadataValue: uint32(mix & 0x7fff)}
	}
	protocol.Slice(io.directional(), &reducer.Outputs)
}
