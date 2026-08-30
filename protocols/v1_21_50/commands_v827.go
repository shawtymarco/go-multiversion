package v1_21_50

import (
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAvailableCommands827(io *wireIO, pk *packet.AvailableCommands) {
	protocol.FuncSlice(io.directional(), &pk.EnumValues, io.String)
	protocol.FuncSlice(io.directional(), &pk.ChainedSubcommandValues, io.String)
	protocol.FuncSlice(io.directional(), &pk.Suffixes, io.String)
	protocol.FuncIOSlice(io.directional(), &pk.Enums, func(raw protocol.IO, value *protocol.CommandEnum) {
		marshalCommandEnum827(asWireIO(raw), value, len(pk.EnumValues))
	})
	protocol.FuncIOSlice(io.directional(), &pk.ChainedSubcommands, marshalChainedSubcommand827)
	protocol.FuncIOSlice(io.directional(), &pk.Commands, marshalCommand827)
	protocol.FuncIOSlice(io.directional(), &pk.DynamicEnums, marshalDynamicEnum827)
	protocol.FuncIOSlice(io.directional(), &pk.Constraints, marshalCommandConstraint827)
}

func marshalCommandEnum827(io *wireIO, value *protocol.CommandEnum, enumValueCount int) {
	io.String(&value.Type)
	count := uint32(len(value.ValueIndices))
	io.Varuint32(&count)
	if io.reading {
		value.ValueIndices = make([]uint32, count)
	}
	for index := range value.ValueIndices {
		switch {
		case enumValueCount <= math.MaxUint8:
			entry := uint8(value.ValueIndices[index])
			io.Uint8(&entry)
			value.ValueIndices[index] = uint32(entry)
		case enumValueCount <= math.MaxUint16:
			entry := uint16(value.ValueIndices[index])
			io.Uint16(&entry)
			value.ValueIndices[index] = uint32(entry)
		default:
			io.Uint32(&value.ValueIndices[index])
		}
	}
}

func marshalChainedSubcommand827(raw protocol.IO, value *protocol.ChainedSubcommand) {
	io := asWireIO(raw)
	io.String(&value.Name)
	protocol.FuncIOSlice(io.directional(), &value.Values, func(raw protocol.IO, entry *protocol.ChainedSubcommandValue) {
		legacy := asWireIO(raw)
		index, kind := uint16(entry.Index), uint16(entry.Value)
		legacy.Uint16(&index)
		legacy.Uint16(&kind)
		entry.Index, entry.Value = uint32(index), uint32(kind)
	})
}

func marshalCommand827(raw protocol.IO, value *protocol.Command) {
	io := asWireIO(raw)
	io.String(&value.Name)
	io.String(&value.Description)
	io.Uint16(&value.Flags)
	io.Uint8(&value.PermissionLevel)
	io.Uint32(&value.AliasesOffset)
	count := uint32(len(value.ChainedSubcommandOffsets))
	io.Varuint32(&count)
	if io.reading {
		value.ChainedSubcommandOffsets = make([]uint32, count)
	}
	for index := range value.ChainedSubcommandOffsets {
		offset := uint16(value.ChainedSubcommandOffsets[index])
		io.Uint16(&offset)
		value.ChainedSubcommandOffsets[index] = uint32(offset)
	}
	protocol.FuncIOSlice(io.directional(), &value.Overloads, marshalCommandOverload827)
}

func marshalCommandOverload827(raw protocol.IO, value *protocol.CommandOverload) {
	io := asWireIO(raw)
	io.Bool(&value.Chaining)
	protocol.FuncIOSlice(io.directional(), &value.Parameters, marshalCommandParameter827)
}

func marshalCommandParameter827(raw protocol.IO, value *protocol.CommandParameter) {
	io := asWireIO(raw)
	io.String(&value.Name)
	if io.reading {
		io.Uint32(&value.Type)
		value.Type = commandArgumentFrom827(io, value.Type)
	} else {
		legacyType := commandArgumentTo827(io, value.Type)
		io.Uint32(&legacyType)
	}
	io.Bool(&value.Optional)
	options := value.Options
	if !io.reading && options == protocol.ParamOptionAsChainedCommand {
		options = 3
	}
	io.Uint8(&options)
	if io.reading {
		if options == 3 {
			options = protocol.ParamOptionAsChainedCommand
		}
		value.Options = options
	}
}

func commandArgumentTo827(io *wireIO, value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSoftEnum|protocol.CommandArgSuffixed) != 0 {
		return value
	}
	flags, kind := value&^0xfffff, value&0xfffff
	if kind == protocol.CommandArgTypeStandaloneTarget || kind == protocol.CommandArgTypeNonIDTarget {
		return flags | 8
	}
	legacy, ok := commandArgument827ByCurrent[kind]
	if !ok {
		// Protocol 827 has no representation for newer basic argument types.
		// RValue is the historical generic parser and keeps one unsupported
		// command usage from terminating the entire server write loop.
		return flags | 4
	}
	return flags | legacy
}

func commandArgumentFrom827(io *wireIO, value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSoftEnum|protocol.CommandArgSuffixed) != 0 {
		return value
	}
	flags, kind := value&^0xfffff, value&0xfffff
	current, ok := commandArgumentCurrentBy827[kind]
	if !ok {
		io.UnknownEnumOption(kind, "protocol 766 command argument type")
		return flags
	}
	return flags | current
}

var commandArgument827ByCurrent = map[uint32]uint32{
	protocol.CommandArgTypeInt:             1,
	protocol.CommandArgTypeFloat:           3,
	protocol.CommandArgTypeRValue:          4,
	protocol.CommandArgTypeWildcardInt:     5,
	protocol.CommandArgTypeOperator:        6,
	protocol.CommandArgTypeCompareOperator: 7,
	protocol.CommandArgTypeTarget:          8,
	protocol.CommandArgTypeWildcardTarget:  10,
	protocol.CommandArgTypeFilepath:        17,
	protocol.CommandArgTypeIntegerRange:    23,
	protocol.CommandArgTypeEquipmentSlots:  47,
	protocol.CommandArgTypeString:          56,
	protocol.CommandArgTypeBlockPosition:   64,
	protocol.CommandArgTypePosition:        65,
	protocol.CommandArgTypeMessage:         67,
	protocol.CommandArgTypeRawText:         70,
	protocol.CommandArgTypeJSON:            74,
	protocol.CommandArgTypeBlockStates:     83,
	protocol.CommandArgTypeCommand:         86,
}

var commandArgumentCurrentBy827 = func() map[uint32]uint32 {
	reverse := make(map[uint32]uint32, len(commandArgument827ByCurrent))
	for current, legacy := range commandArgument827ByCurrent {
		reverse[legacy] = current
	}
	return reverse
}()

func marshalDynamicEnum827(raw protocol.IO, value *protocol.DynamicEnum) {
	io := asWireIO(raw)
	io.String(&value.Type)
	protocol.FuncSlice(io.directional(), &value.Values, io.String)
}

func marshalCommandConstraint827(raw protocol.IO, value *protocol.CommandEnumConstraint) {
	io := asWireIO(raw)
	io.Uint32(&value.EnumValueIndex)
	io.Uint32(&value.EnumIndex)
	protocol.FuncSlice(io.directional(), &value.Constraints, io.Uint8)
}
