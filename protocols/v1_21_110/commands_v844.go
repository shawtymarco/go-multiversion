package v1_21_110

import (
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAvailableCommands844(io *wireIO, pk *packet.AvailableCommands) {
	protocol.FuncSlice(io.directional(), &pk.EnumValues, io.String)
	protocol.FuncSlice(io.directional(), &pk.ChainedSubcommandValues, io.String)
	protocol.FuncSlice(io.directional(), &pk.Suffixes, io.String)
	protocol.FuncIOSlice(io.directional(), &pk.Enums, func(raw protocol.IO, value *protocol.CommandEnum) {
		marshalCommandEnum844(asWireIO(raw), value, len(pk.EnumValues))
	})
	protocol.FuncIOSlice(io.directional(), &pk.ChainedSubcommands, marshalChainedSubcommand844)
	protocol.FuncIOSlice(io.directional(), &pk.Commands, marshalCommand844)
	protocol.FuncIOSlice(io.directional(), &pk.DynamicEnums, marshalDynamicEnum844)
	protocol.FuncIOSlice(io.directional(), &pk.Constraints, marshalCommandConstraint844)
}

func marshalCommandEnum844(io *wireIO, value *protocol.CommandEnum, enumValueCount int) {
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

func marshalChainedSubcommand844(raw protocol.IO, value *protocol.ChainedSubcommand) {
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

func marshalCommand844(raw protocol.IO, value *protocol.Command) {
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
	protocol.FuncIOSlice(io.directional(), &value.Overloads, marshalCommandOverload844)
}

func marshalCommandOverload844(raw protocol.IO, value *protocol.CommandOverload) {
	io := asWireIO(raw)
	io.Bool(&value.Chaining)
	protocol.FuncIOSlice(io.directional(), &value.Parameters, marshalCommandParameter844)
}

func marshalCommandParameter844(raw protocol.IO, value *protocol.CommandParameter) {
	io := asWireIO(raw)
	io.String(&value.Name)
	if io.reading {
		io.Uint32(&value.Type)
		value.Type = commandArgumentFrom844(io, value.Type)
	} else {
		legacyType := commandArgumentTo844(io, value.Type)
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

func commandArgumentTo844(io *wireIO, value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSoftEnum|protocol.CommandArgSuffixed) != 0 {
		return value
	}
	flags, kind := value&^0xfffff, value&0xfffff
	if kind == protocol.CommandArgTypeStandaloneTarget || kind == protocol.CommandArgTypeNonIDTarget {
		return flags | 8
	}
	legacy, ok := commandArgument844ByCurrent[kind]
	if !ok {
		// Protocol 844 has no representation for newer basic argument types.
		// RValue is the historical generic parser and keeps one unsupported
		// command usage from terminating the entire server write loop.
		return flags | 4
	}
	return flags | legacy
}

func commandArgumentFrom844(io *wireIO, value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSoftEnum|protocol.CommandArgSuffixed) != 0 {
		return value
	}
	flags, kind := value&^0xfffff, value&0xfffff
	current, ok := commandArgumentCurrentBy844[kind]
	if !ok {
		io.UnknownEnumOption(kind, "protocol 844 command argument type")
		return flags
	}
	return flags | current
}

var commandArgument844ByCurrent = map[uint32]uint32{
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

var commandArgumentCurrentBy844 = func() map[uint32]uint32 {
	reverse := make(map[uint32]uint32, len(commandArgument844ByCurrent))
	for current, legacy := range commandArgument844ByCurrent {
		reverse[legacy] = current
	}
	return reverse
}()

func marshalDynamicEnum844(raw protocol.IO, value *protocol.DynamicEnum) {
	io := asWireIO(raw)
	io.String(&value.Type)
	protocol.FuncSlice(io.directional(), &value.Values, io.String)
}

func marshalCommandConstraint844(raw protocol.IO, value *protocol.CommandEnumConstraint) {
	io := asWireIO(raw)
	io.Uint32(&value.EnumValueIndex)
	io.Uint32(&value.EnumIndex)
	protocol.FuncSlice(io.directional(), &value.Constraints, io.Uint8)
}
