package v1_18_0

import (
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAvailableCommands(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AvailableCommands)
	protocol.FuncSlice(io.directional(), &pk.EnumValues, io.String)
	protocol.FuncSlice(io.directional(), &pk.Suffixes, io.String)
	protocol.FuncIOSlice(io.directional(), &pk.Enums, func(raw protocol.IO, enum *protocol.CommandEnum) {
		legacy := asWireIO(raw)
		legacy.String(&enum.Type)
		protocol.FuncSlice(legacy.directional(), &enum.ValueIndices, func(index *uint32) {
			marshalLegacyEnumIndex(legacy, index, len(pk.EnumValues))
		})
	})
	protocol.FuncIOSlice(io.directional(), &pk.Commands, marshalLegacyCommand)
	protocol.FuncIOSlice(io.directional(), &pk.DynamicEnums, func(raw protocol.IO, enum *protocol.DynamicEnum) {
		legacy := asWireIO(raw)
		legacy.String(&enum.Type)
		protocol.FuncSlice(legacy.directional(), &enum.Values, legacy.String)
	})
	protocol.Slice(io.directional(), &pk.Constraints)
	if io.reading {
		pk.ChainedSubcommandValues = nil
		pk.ChainedSubcommands = nil
	}
}

func marshalLegacyEnumIndex(io *wireIO, value *uint32, enumValueCount int) {
	switch {
	case enumValueCount <= math.MaxUint8:
		converted := uint8(*value)
		io.Uint8(&converted)
		*value = uint32(converted)
	case enumValueCount <= math.MaxUint16:
		converted := uint16(*value)
		io.Uint16(&converted)
		*value = uint32(converted)
	default:
		io.Uint32(value)
	}
}

func marshalLegacyCommand(raw protocol.IO, command *protocol.Command) {
	io := asWireIO(raw)
	io.String(&command.Name)
	io.String(&command.Description)
	io.Uint16(&command.Flags)
	io.Uint8(&command.PermissionLevel)
	aliasOffset := int32(command.AliasesOffset)
	io.Int32(&aliasOffset)
	command.AliasesOffset = uint32(aliasOffset)
	protocol.FuncIOSlice(io.directional(), &command.Overloads, func(raw protocol.IO, overload *protocol.CommandOverload) {
		legacy := asWireIO(raw)
		protocol.FuncIOSlice(legacy.directional(), &overload.Parameters, marshalLegacyCommandParameter)
		if legacy.reading {
			overload.Chaining = false
		}
	})
	if io.reading {
		command.ChainedSubcommandOffsets = nil
	}
}

func marshalLegacyCommandParameter(raw protocol.IO, parameter *protocol.CommandParameter) {
	io := asWireIO(raw)
	io.String(&parameter.Name)
	wireType := parameter.Type
	if !io.reading {
		wireType = legacyCommandParameterType(wireType)
	}
	io.Uint32(&wireType)
	if io.reading {
		parameter.Type = nativeCommandParameterType(wireType)
	}
	io.Bool(&parameter.Optional)
	io.Uint8(&parameter.Options)
}

func legacyCommandParameterType(value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSuffixed|protocol.CommandArgSoftEnum) != 0 {
		return value
	}
	base := value & 0xfffff
	mapped := map[uint32]uint32{
		protocol.CommandArgTypeInt: 1, protocol.CommandArgTypeFloat: 3,
		protocol.CommandArgTypeOperator: 6, protocol.CommandArgTypeCompareOperator: 6,
		protocol.CommandArgTypeTarget: 7, protocol.CommandArgTypeWildcardTarget: 8,
		protocol.CommandArgTypeFilepath: 16, protocol.CommandArgTypeString: 32,
		protocol.CommandArgTypeBlockPosition: 40, protocol.CommandArgTypePosition: 40,
		protocol.CommandArgTypeMessage: 44, protocol.CommandArgTypeRawText: 46,
		protocol.CommandArgTypeJSON: 50, protocol.CommandArgTypeCommand: 63,
	}
	if target, ok := mapped[base]; ok {
		return value&^0xfffff | target
	}
	return value
}

func nativeCommandParameterType(value uint32) uint32 {
	if value&(protocol.CommandArgEnum|protocol.CommandArgSuffixed|protocol.CommandArgSoftEnum) != 0 {
		return value
	}
	base := value & 0xfffff
	mapped := map[uint32]uint32{
		1: protocol.CommandArgTypeInt, 3: protocol.CommandArgTypeFloat, 6: protocol.CommandArgTypeOperator,
		7: protocol.CommandArgTypeTarget, 8: protocol.CommandArgTypeWildcardTarget,
		16: protocol.CommandArgTypeFilepath, 32: protocol.CommandArgTypeString,
		40: protocol.CommandArgTypePosition, 44: protocol.CommandArgTypeMessage,
		46: protocol.CommandArgTypeRawText, 50: protocol.CommandArgTypeJSON, 63: protocol.CommandArgTypeCommand,
	}
	if current, ok := mapped[base]; ok {
		return value&^0xfffff | current
	}
	return value
}

func marshalCommandBlockUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandBlockUpdate)
	io.Bool(&pk.Block)
	if pk.Block {
		marshalUnsignedBlockPos(io, &pk.Position)
		io.Varuint32(&pk.Mode)
		io.Bool(&pk.NeedsRedstone)
		io.Bool(&pk.Conditional)
	} else {
		io.Varuint64(&pk.MinecartEntityRuntimeID)
	}
	io.String(&pk.Command)
	io.String(&pk.LastOutput)
	io.String(&pk.Name)
	io.Bool(&pk.ShouldTrackOutput)
	tickDelay := int32(pk.TickDelay)
	io.Int32(&tickDelay)
	pk.TickDelay = uint32(tickDelay)
	io.Bool(&pk.ExecuteOnFirstTick)
	if io.reading {
		pk.FilteredName = ""
	}
}

func marshalCommandOutput(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandOutput)
	marshalCommandOrigin(io, &pk.CommandOrigin)
	io.Uint8(&pk.OutputType)
	io.Varuint32(&pk.SuccessCount)
	protocol.FuncIOSlice(io.directional(), &pk.OutputMessages, func(raw protocol.IO, message *protocol.CommandOutputMessage) {
		legacy := asWireIO(raw)
		legacy.Bool(&message.Success)
		legacy.String(&message.Message)
		protocol.FuncSlice(legacy.directional(), &message.Parameters, legacy.String)
	})
	if pk.OutputType == packet.CommandOutputTypeDataSet {
		dataSet, _ := pk.DataSet.Value()
		io.String(&dataSet)
		pk.DataSet = protocol.Option(dataSet)
	} else if io.reading {
		pk.DataSet = protocol.Optional[string]{}
	}
}
