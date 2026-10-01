package packetio

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// TransactionBody2193 omits the preview hand fields. Older clients only address
// their main hand; the native model is populated with that explicit default.
func TransactionBody2193(io protocol.IO, value protocol.InventoryTransactionData, reading bool) {
	switch x := value.(type) {
	case *protocol.UseItemOnEntityTransactionData:
		io.ActorRuntimeID(&x.TargetEntityRuntimeID)
		io.Varint32(&x.ActionType)
		io.Varint32(&x.HotBarSlot)
		io.ItemInstance(&x.HeldItem)
		io.Vec3(&x.Position)
		io.Vec3(&x.ClickedPosition)
		if reading {
			x.Hand = protocol.HandSlotMainHand
		}
	case *protocol.ReleaseItemTransactionData:
		io.Varint32(&x.ActionType)
		io.Varint32(&x.HotBarSlot)
		io.ItemInstance(&x.HeldItem)
		io.Vec3(&x.HeadPosition)
		if reading {
			x.Hand = protocol.HandSlotMainHand
		}
	default:
		value.Marshal(io)
	}
}
