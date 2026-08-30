package v1_18_0

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAddPlayer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddPlayer)
	io.UUID(&pk.UUID)
	io.String(&pk.Username)
	io.Varint64(&pk.AbilityData.EntityUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.PlatformChatID)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.ItemInstance(&pk.HeldItem)
	io.EntityMetadata(&pk.EntityMetadata)
	flags, actions := legacyAbilityFlags(pk.AbilityData.Layers)
	io.Varuint32(&flags)
	commandPermissions := uint32(pk.AbilityData.CommandPermissions)
	io.Varuint32(&commandPermissions)
	io.Varuint32(&actions)
	playerPermissions := uint32(pk.AbilityData.PlayerPermissions)
	io.Varuint32(&playerPermissions)
	var customStoredPermissions uint32
	io.Varuint32(&customStoredPermissions)
	io.Int64(&pk.AbilityData.EntityUniqueID)
	marshalEntityLinks(io, &pk.EntityLinks)
	io.String(&pk.DeviceID)
	io.Int32(&pk.BuildPlatform)
	if io.reading {
		pk.GameType = 0
		pk.EntityProperties = protocol.EntityProperties{}
		pk.AbilityData.CommandPermissions = byte(commandPermissions)
		pk.AbilityData.PlayerPermissions = byte(playerPermissions)
		pk.AbilityData.Layers = abilityLayersFromLegacy(flags, actions)
	}
}

func marshalAddActor(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddActor)
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.EntityType)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	protocol.FuncIOSlice(io.directional(), &pk.Attributes, func(raw protocol.IO, attribute *protocol.AttributeValue) {
		legacy := asWireIO(raw)
		legacy.String(&attribute.Name)
		legacy.Float32(&attribute.Min)
		legacy.Float32(&attribute.Value)
		legacy.Float32(&attribute.Max)
	})
	io.EntityMetadata(&pk.EntityMetadata)
	marshalEntityLinks(io, &pk.EntityLinks)
	if io.reading {
		pk.BodyYaw = pk.Yaw
		pk.EntityProperties = protocol.EntityProperties{}
	}
}

func marshalSetActorData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetActorData)
	io.Varuint64(&pk.EntityRuntimeID)
	io.EntityMetadata(&pk.EntityMetadata)
	io.Varuint64(&pk.Tick)
	if io.reading {
		pk.EntityProperties = protocol.EntityProperties{}
	}
}

func marshalSetActorLink(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetActorLink)
	marshalEntityLink(io, &pk.EntityLink)
}

func marshalUpdateAttributes(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateAttributes)
	io.Varuint64(&pk.EntityRuntimeID)
	protocol.FuncIOSlice(io.directional(), &pk.Attributes, func(raw protocol.IO, attribute *protocol.Attribute) {
		legacy := asWireIO(raw)
		legacy.Float32(&attribute.Min)
		legacy.Float32(&attribute.Max)
		legacy.Float32(&attribute.Value)
		legacy.Float32(&attribute.Default)
		legacy.String(&attribute.Name)
		if legacy.reading {
			attribute.DefaultMin = attribute.Min
			attribute.DefaultMax = attribute.Max
			attribute.Modifiers = nil
		}
	})
	io.Varuint64(&pk.Tick)
}

func marshalEntityLinks(io *wireIO, links *[]protocol.EntityLink) {
	protocol.FuncIOSlice(io.directional(), links, func(raw protocol.IO, link *protocol.EntityLink) {
		marshalEntityLink(asWireIO(raw), link)
	})
}

func marshalEntityLink(io *wireIO, link *protocol.EntityLink) {
	io.Varint64(&link.RiddenEntityUniqueID)
	io.Varint64(&link.RiderEntityUniqueID)
	io.Uint8(&link.Type)
	io.Bool(&link.Immediate)
	io.Bool(&link.RiderInitiated)
	if io.reading {
		link.VehicleAngularVelocity = 0
	}
}

func marshalEntityMetadata(io *wireIO, metadata *protocol.EntityMetadata) {
	if io.reading {
		var count uint32
		io.Varuint32(&count)
		decoded := make(protocol.EntityMetadata, count)
		for range count {
			var key, kind uint32
			io.Varuint32(&key)
			io.Varuint32(&kind)
			decoded[key] = readEntityMetadataValue(io, kind)
		}
		upgradeEntityFlags(decoded)
		*metadata = decoded
		return
	}
	encoded := cloneEntityMetadata(*metadata)
	downgradeEntityFlags(encoded)
	keys := make([]int, 0, len(encoded))
	for key := range encoded {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)
	count := uint32(len(keys))
	io.Varuint32(&count)
	for _, rawKey := range keys {
		key := uint32(rawKey)
		io.Varuint32(&key)
		writeEntityMetadataValue(io, encoded[key])
	}
}

func readEntityMetadataValue(io *wireIO, kind uint32) any {
	switch kind {
	case protocol.EntityDataTypeByte:
		var value byte
		io.Uint8(&value)
		return value
	case protocol.EntityDataTypeInt16:
		var value int16
		io.Int16(&value)
		return value
	case protocol.EntityDataTypeInt32:
		var value int32
		io.Varint32(&value)
		return value
	case protocol.EntityDataTypeFloat32:
		var value float32
		io.Float32(&value)
		return value
	case protocol.EntityDataTypeString:
		var value string
		io.String(&value)
		return value
	case protocol.EntityDataTypeCompoundTag:
		var value map[string]any
		io.NBT(&value, nbt.NetworkLittleEndian)
		return value
	case protocol.EntityDataTypeBlockPos:
		var value protocol.BlockPos
		io.BlockPos(&value)
		return value
	case protocol.EntityDataTypeInt64:
		var value int64
		io.Varint64(&value)
		return value
	case protocol.EntityDataTypeVec3:
		var value mgl32.Vec3
		io.Vec3(&value)
		return value
	default:
		io.UnknownEnumOption(kind, "entity metadata type")
		return nil
	}
}

func writeEntityMetadataValue(io *wireIO, value any) {
	var kind uint32
	switch value.(type) {
	case byte:
		kind = protocol.EntityDataTypeByte
	case int16:
		kind = protocol.EntityDataTypeInt16
	case int32:
		kind = protocol.EntityDataTypeInt32
	case float32:
		kind = protocol.EntityDataTypeFloat32
	case string:
		kind = protocol.EntityDataTypeString
	case map[string]any:
		kind = protocol.EntityDataTypeCompoundTag
	case protocol.BlockPos:
		kind = protocol.EntityDataTypeBlockPos
	case int64:
		kind = protocol.EntityDataTypeInt64
	case mgl32.Vec3:
		kind = protocol.EntityDataTypeVec3
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value), "entity metadata type")
		return
	}
	io.Varuint32(&kind)
	switch value := value.(type) {
	case byte:
		io.Uint8(&value)
	case int16:
		io.Int16(&value)
	case int32:
		io.Varint32(&value)
	case float32:
		io.Float32(&value)
	case string:
		io.String(&value)
	case map[string]any:
		io.NBT(&value, nbt.NetworkLittleEndian)
	case protocol.BlockPos:
		io.BlockPos(&value)
	case int64:
		io.Varint64(&value)
	case mgl32.Vec3:
		io.Vec3(&value)
	}
}

func cloneEntityMetadata(source protocol.EntityMetadata) protocol.EntityMetadata {
	cloned := make(protocol.EntityMetadata, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func downgradeEntityFlags(metadata protocol.EntityMetadata) {
	_, hadFirst := metadata[protocol.EntityDataKeyFlags]
	_, hadSecond := metadata[protocol.EntityDataKeyFlagsTwo]
	if !hadFirst && !hadSecond {
		return
	}
	first, _ := metadata[protocol.EntityDataKeyFlags].(int64)
	second, _ := metadata[protocol.EntityDataKeyFlagsTwo].(int64)
	targetFirst, targetSecond := removeEntityFlag(uint64(first), uint64(second), protocol.EntityDataFlagDash)
	metadata[protocol.EntityDataKeyFlags] = int64(targetFirst)
	if targetSecond != 0 || hadSecond {
		metadata[protocol.EntityDataKeyFlagsTwo] = int64(targetSecond)
	} else {
		delete(metadata, protocol.EntityDataKeyFlagsTwo)
	}
}

func upgradeEntityFlags(metadata protocol.EntityMetadata) {
	_, hadFirst := metadata[protocol.EntityDataKeyFlags]
	_, hadSecond := metadata[protocol.EntityDataKeyFlagsTwo]
	if !hadFirst && !hadSecond {
		return
	}
	first, _ := metadata[protocol.EntityDataKeyFlags].(int64)
	second, _ := metadata[protocol.EntityDataKeyFlagsTwo].(int64)
	nativeFirst, nativeSecond := insertEntityFlag(uint64(first), uint64(second), protocol.EntityDataFlagDash)
	metadata[protocol.EntityDataKeyFlags] = int64(nativeFirst)
	if nativeSecond != 0 || hadSecond {
		metadata[protocol.EntityDataKeyFlagsTwo] = int64(nativeSecond)
	} else {
		delete(metadata, protocol.EntityDataKeyFlagsTwo)
	}
}

func removeEntityFlag(first, second uint64, removed int) (uint64, uint64) {
	var target [2]uint64
	for current := 0; current < 128; current++ {
		word, bit := current/64, uint(current%64)
		value := [2]uint64{first, second}[word]
		if current == removed || value&(uint64(1)<<bit) == 0 {
			continue
		}
		legacy := current
		if current > removed {
			legacy--
		}
		target[legacy/64] |= uint64(1) << uint(legacy%64)
	}
	return target[0], target[1]
}

func insertEntityFlag(first, second uint64, inserted int) (uint64, uint64) {
	var target [2]uint64
	for legacy := 0; legacy < 127; legacy++ {
		word, bit := legacy/64, uint(legacy%64)
		value := [2]uint64{first, second}[word]
		if value&(uint64(1)<<bit) == 0 {
			continue
		}
		current := legacy
		if legacy >= inserted {
			current++
		}
		target[current/64] |= uint64(1) << uint(current%64)
	}
	return target[0], target[1]
}

func legacyAbilityFlags(layers []protocol.AbilityLayer) (flags, actions uint32) {
	if len(layers) == 0 {
		return 0, 0
	}
	var values uint32
	for _, layer := range layers {
		values |= layer.Values
	}
	if values&(protocol.AbilityBuild|protocol.AbilityMine) == 0 {
		flags |= packet.AdventureFlagWorldImmutable
	}
	if values&(protocol.AbilityDoorsAndSwitches|protocol.AbilityOpenContainers|protocol.AbilityAttackPlayers|protocol.AbilityAttackMobs) == 0 {
		flags |= packet.AdventureSettingsFlagsNoPvM
	}
	if values&protocol.AbilityMayFly != 0 {
		flags |= packet.AdventureFlagAllowFlight
	}
	if values&protocol.AbilityFlying != 0 {
		flags |= packet.AdventureFlagFlying
	}
	if values&protocol.AbilityNoClip != 0 {
		flags |= packet.AdventureFlagNoClip
	}
	if values&protocol.AbilityWorldBuilder != 0 {
		flags |= packet.AdventureFlagWorldBuilder
	}
	if values&protocol.AbilityMuted != 0 {
		flags |= packet.AdventureFlagMuted
	}
	permissions := []struct{ ability, action uint32 }{
		{protocol.AbilityMine, packet.ActionPermissionMine},
		{protocol.AbilityDoorsAndSwitches, packet.ActionPermissionDoorsAndSwitches},
		{protocol.AbilityOpenContainers, packet.ActionPermissionOpenContainers},
		{protocol.AbilityAttackPlayers, packet.ActionPermissionAttackPlayers},
		{protocol.AbilityAttackMobs, packet.ActionPermissionAttackMobs},
		{protocol.AbilityOperatorCommands, packet.ActionPermissionOperator},
		{protocol.AbilityTeleport, packet.ActionPermissionTeleport},
		{protocol.AbilityBuild, packet.ActionPermissionBuild},
	}
	for _, permission := range permissions {
		if values&permission.ability != 0 {
			actions |= permission.action
		}
	}
	return
}

func abilityLayersFromLegacy(flags, actions uint32) []protocol.AbilityLayer {
	var values uint32
	permissions := []struct{ action, ability uint32 }{
		{packet.ActionPermissionMine, protocol.AbilityMine},
		{packet.ActionPermissionDoorsAndSwitches, protocol.AbilityDoorsAndSwitches},
		{packet.ActionPermissionOpenContainers, protocol.AbilityOpenContainers},
		{packet.ActionPermissionAttackPlayers, protocol.AbilityAttackPlayers},
		{packet.ActionPermissionAttackMobs, protocol.AbilityAttackMobs},
		{packet.ActionPermissionOperator, protocol.AbilityOperatorCommands},
		{packet.ActionPermissionTeleport, protocol.AbilityTeleport},
		{packet.ActionPermissionBuild, protocol.AbilityBuild},
	}
	for _, permission := range permissions {
		if actions&permission.action != 0 {
			values |= permission.ability
		}
	}
	if flags&packet.AdventureFlagAllowFlight != 0 {
		values |= protocol.AbilityMayFly
	}
	if flags&packet.AdventureFlagFlying != 0 {
		values |= protocol.AbilityFlying
	}
	if flags&packet.AdventureFlagNoClip != 0 {
		values |= protocol.AbilityNoClip
	}
	if flags&packet.AdventureFlagWorldBuilder != 0 {
		values |= protocol.AbilityWorldBuilder
	}
	if flags&packet.AdventureFlagMuted != 0 {
		values |= protocol.AbilityMuted
	}
	return []protocol.AbilityLayer{{Type: protocol.AbilityLayerTypeBase, Abilities: protocol.AbilityCount - 1, Values: values, FlySpeed: protocol.AbilityBaseFlySpeed, VerticalFlySpeed: protocol.AbilityBaseVerticalFlySpeed, WalkSpeed: protocol.AbilityBaseWalkSpeed}}
}
