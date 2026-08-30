package v1_21_130

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAddPlayer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddPlayer)
	io.UUID(&pk.UUID)
	io.String(&pk.Username)
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.PlatformChatID)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.ItemInstance(&pk.HeldItem)
	io.Varint32(&pk.GameType)
	io.EntityMetadata(&pk.EntityMetadata)
	protocol.Single(io.directional(), &pk.EntityProperties)
	marshalAbilityData(io, &pk.AbilityData)
	protocol.Slice(io.directional(), &pk.EntityLinks)
	io.String(&pk.DeviceID)
	io.Int32(&pk.BuildPlatform)
}

func marshalClientCheatAbility(io *wireIO, raw packet.Packet) {
	marshalAbilityData(io, &raw.(*packet.ClientCheatAbility).AbilityData)
}

func marshalUpdateAbilities(io *wireIO, raw packet.Packet) {
	marshalAbilityData(io, &raw.(*packet.UpdateAbilities).AbilityData)
}

func marshalCreativeContent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CreativeContent)
	protocol.FuncIOSlice(io.directional(), &pk.Groups, func(io protocol.IO, group *protocol.CreativeGroup) {
		legacy := asWireIO(io)
		category := int32(group.Category)
		legacy.Int32(&category)
		group.Category = byte(category)
		legacy.String(&group.Name)
		legacy.Item(&group.Icon)
	})
	protocol.Slice(io.directional(), &pk.Items)
}

func marshalAvailableCommands(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AvailableCommands)
	if io.reading {
		pk.Marshal(io.directional())
		normaliseCommandFloatTypes(pk, 2, protocol.CommandArgTypeFloat)
		return
	}
	copyPacket := *pk
	copyPacket.Commands = cloneCommands(pk.Commands)
	normaliseCommandFloatTypes(&copyPacket, protocol.CommandArgTypeFloat, 2)
	copyPacket.Marshal(io.directional())
}

func cloneCommands(commands []protocol.Command) []protocol.Command {
	cloned := make([]protocol.Command, len(commands))
	copy(cloned, commands)
	for commandIndex := range cloned {
		cloned[commandIndex].ChainedSubcommandOffsets = append([]uint32(nil), commands[commandIndex].ChainedSubcommandOffsets...)
		cloned[commandIndex].Overloads = make([]protocol.CommandOverload, len(commands[commandIndex].Overloads))
		copy(cloned[commandIndex].Overloads, commands[commandIndex].Overloads)
		for overloadIndex := range cloned[commandIndex].Overloads {
			cloned[commandIndex].Overloads[overloadIndex].Parameters = append(
				[]protocol.CommandParameter(nil), commands[commandIndex].Overloads[overloadIndex].Parameters...,
			)
		}
	}
	return cloned
}

func normaliseCommandFloatTypes(pk *packet.AvailableCommands, from, to uint32) {
	for commandIndex := range pk.Commands {
		for overloadIndex := range pk.Commands[commandIndex].Overloads {
			for parameterIndex := range pk.Commands[commandIndex].Overloads[overloadIndex].Parameters {
				parameter := &pk.Commands[commandIndex].Overloads[overloadIndex].Parameters[parameterIndex]
				if parameter.Type&0xfffff == from {
					parameter.Type = parameter.Type&^0xfffff | to
				}
			}
		}
	}
}
