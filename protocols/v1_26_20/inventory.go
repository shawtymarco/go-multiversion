package v1_26_20

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalTransactionDataType(io *wireIO, value *protocol.InventoryTransactionData) {
	var typeID uint32
	if !io.reading {
		switch (*value).(type) {
		case *protocol.NormalTransactionData:
			typeID = protocol.InventoryTransactionTypeNormal
		case *protocol.MismatchTransactionData:
			typeID = protocol.InventoryTransactionTypeMismatch
		case *protocol.UseItemTransactionData:
			typeID = protocol.InventoryTransactionTypeUseItem
		case *protocol.UseItemOnEntityTransactionData:
			typeID = protocol.InventoryTransactionTypeUseItemOnEntity
		case *protocol.ReleaseItemTransactionData:
			typeID = protocol.InventoryTransactionTypeReleaseItem
		default:
			io.UnknownEnumOption(fmt.Sprintf("%T", *value), "inventory transaction data type")
			return
		}
	}
	io.Varuint32(&typeID)
	if io.reading {
		switch typeID {
		case protocol.InventoryTransactionTypeNormal:
			*value = &protocol.NormalTransactionData{}
		case protocol.InventoryTransactionTypeMismatch:
			*value = &protocol.MismatchTransactionData{}
		case protocol.InventoryTransactionTypeUseItem:
			*value = &protocol.UseItemTransactionData{}
		case protocol.InventoryTransactionTypeUseItemOnEntity:
			*value = &protocol.UseItemOnEntityTransactionData{}
		case protocol.InventoryTransactionTypeReleaseItem:
			*value = &protocol.ReleaseItemTransactionData{}
		default:
			io.UnknownEnumOption(typeID, "inventory transaction data type")
			return
		}
	}
}

func marshalInventoryTransactionData(io *wireIO, value protocol.InventoryTransactionData) {
	switch data := value.(type) {
	case *protocol.NormalTransactionData, *protocol.MismatchTransactionData:
	case *protocol.UseItemTransactionData:
		marshalUseItemTransactionData(io, data)
	case *protocol.UseItemOnEntityTransactionData:
		io.Varuint64(&data.TargetEntityRuntimeID)
		io.Varint32(&data.ActionType)
		io.Varint32(&data.HotBarSlot)
		marshalItemInstanceNew(io, &data.HeldItem)
		io.Vec3(&data.Position)
		io.Vec3(&data.ClickedPosition)
	case *protocol.ReleaseItemTransactionData:
		io.Varint32(&data.ActionType)
		io.Varint32(&data.HotBarSlot)
		marshalItemInstanceNew(io, &data.HeldItem)
		io.Vec3(&data.HeadPosition)
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value), "inventory transaction data type")
	}
}

func marshalUseItemTransactionData(io *wireIO, data *protocol.UseItemTransactionData) {
	protocol.IntegerFunc(&data.ActionType, io.Varint32)
	protocol.IntegerFunc(&data.TriggerType, io.Uint8)
	io.BlockPos(&data.BlockPosition)
	protocol.IntegerFunc(&data.BlockFace, io.Uint8)
	io.Varint32(&data.HotBarSlot)
	marshalItemInstanceNew(io, &data.HeldItem)
	io.Vec3(&data.Position)
	io.Vec3(&data.ClickedPosition)
	io.Varuint32(&data.BlockRuntimeID)
	io.Uint8(&data.ClientPrediction)
	io.Uint8(&data.ClientCooldownState)
}

func marshalInventoryAction(io *wireIO, action *protocol.InventoryAction) {
	io.Varuint32(&action.SourceType)
	present := true
	io.Bool(&present)
	if !present {
		io.InvalidValue(present, "inventory action container ID", "outer optional must be present")
		return
	}
	hasContainerID := action.SourceType == protocol.InventoryActionSourceContainer || action.SourceType == protocol.InventoryActionSourceTODO
	io.Bool(&hasContainerID)
	if hasContainerID {
		windowID, _ := action.WindowID.Value()
		io.Int8(&windowID)
		action.WindowID = protocol.Option(windowID)
	} else if io.reading {
		action.WindowID = protocol.Optional[int8]{}
	}

	present = true
	io.Bool(&present)
	if !present {
		io.InvalidValue(present, "inventory action source flags", "outer optional must be present")
		return
	}
	hasFlags := action.SourceType == protocol.InventoryActionSourceWorld
	io.Bool(&hasFlags)
	if hasFlags {
		flags, _ := action.SourceFlags.Value()
		io.Varuint32(&flags)
		action.SourceFlags = protocol.Option(flags)
	} else if io.reading {
		action.SourceFlags = protocol.Optional[uint32]{}
	}
	io.Varuint32(&action.InventorySlot)
	marshalItemInstanceNew(io, &action.OldItem)
	marshalItemInstanceNew(io, &action.NewItem)
}

func marshalInventoryTransactionPacket(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventoryTransaction)
	io.Varint32(&pk.LegacyRequestID)
	if pk.LegacyRequestID != 0 {
		protocol.Slice(io.directional(), &pk.LegacySetItemSlots)
	}
	marshalTransactionDataType(io, &pk.TransactionData)
	protocol.FuncIOSlice(io.directional(), &pk.Actions, func(io protocol.IO, action *protocol.InventoryAction) {
		marshalLegacyPlayerInventoryAction(io, action)
	})
	marshalLegacyInventoryTransactionData(io, pk.TransactionData)
}

func marshalLegacyInventoryTransactionData(io *wireIO, value protocol.InventoryTransactionData) {
	switch data := value.(type) {
	case *protocol.NormalTransactionData, *protocol.MismatchTransactionData:
	case *protocol.UseItemTransactionData:
		io.Varuint32(&data.ActionType)
		io.Varuint32(&data.TriggerType)
		io.BlockPos(&data.BlockPosition)
		io.Varint32(&data.BlockFace)
		io.Varint32(&data.HotBarSlot)
		marshalItemInstance(io, &data.HeldItem)
		io.Vec3(&data.Position)
		io.Vec3(&data.ClickedPosition)
		io.Varuint32(&data.BlockRuntimeID)
		prediction := uint32(data.ClientPrediction)
		io.Varuint32(&prediction)
		data.ClientPrediction = uint8(prediction)
		io.Uint8(&data.ClientCooldownState)
	case *protocol.UseItemOnEntityTransactionData:
		io.Varuint64(&data.TargetEntityRuntimeID)
		actionType := uint32(data.ActionType)
		io.Varuint32(&actionType)
		data.ActionType = int32(actionType)
		io.Varint32(&data.HotBarSlot)
		marshalItemInstance(io, &data.HeldItem)
		io.Vec3(&data.Position)
		io.Vec3(&data.ClickedPosition)
	case *protocol.ReleaseItemTransactionData:
		actionType := uint32(data.ActionType)
		io.Varuint32(&actionType)
		data.ActionType = int32(actionType)
		io.Varint32(&data.HotBarSlot)
		marshalItemInstance(io, &data.HeldItem)
		io.Vec3(&data.HeadPosition)
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value), "inventory transaction data type")
	}
}

func marshalPlayerInventoryAction(io *wireIO, data *protocol.UseItemTransactionData) {
	io.Varint32(&data.LegacyRequestID)
	if data.LegacyRequestID < -1 && data.LegacyRequestID&1 == 0 {
		slots, _ := data.LegacySetItemSlots.Value()
		protocol.FuncIOSlice(io.directional(), &slots, func(io protocol.IO, slot *protocol.LegacySetItemSlot) {
			io.Uint8(&slot.ContainerID)
			io.ByteSlice(&slot.Slots)
		})
		data.LegacySetItemSlots = protocol.Option(slots)
	} else if io.reading {
		data.LegacySetItemSlots = protocol.Optional[[]protocol.LegacySetItemSlot]{}
	}
	actions := data.Actions
	protocol.FuncIOSlice(io.directional(), &actions, marshalLegacyPlayerInventoryAction)
	data.Actions = actions
	io.Varuint32(&data.ActionType)
	io.Varuint32(&data.TriggerType)
	io.BlockPos(&data.BlockPosition)
	io.Varint32(&data.BlockFace)
	io.Varint32(&data.HotBarSlot)
	marshalItemInstance(io, &data.HeldItem)
	io.Vec3(&data.Position)
	io.Vec3(&data.ClickedPosition)
	io.Varuint32(&data.BlockRuntimeID)
	prediction := uint32(data.ClientPrediction)
	io.Varuint32(&prediction)
	data.ClientPrediction = uint8(prediction)
	io.Uint8(&data.ClientCooldownState)
}

func marshalLegacyPlayerInventoryAction(io protocol.IO, action *protocol.InventoryAction) {
	legacy := asWireIO(io)
	legacy.Varuint32(&action.SourceType)
	switch action.SourceType {
	case protocol.InventoryActionSourceContainer, protocol.InventoryActionSourceTODO:
		windowID, _ := action.WindowID.Value()
		protocol.IntegerFunc(&windowID, legacy.Varint32)
		action.WindowID = protocol.Option(windowID)
	case protocol.InventoryActionSourceWorld:
		flags, _ := action.SourceFlags.Value()
		legacy.Varuint32(&flags)
		action.SourceFlags = protocol.Option(flags)
	}
	legacy.Varuint32(&action.InventorySlot)
	marshalItemInstance(legacy, &action.OldItem)
	marshalItemInstance(legacy, &action.NewItem)
}

func marshalAbilityValue(io *wireIO, value *any) {
	// Protocol 975 uses the pre-Cereal tagged ability value representation.
	io.IO.AbilityValue(value)
}
