package v1_18_0

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type structureTemplateResponse475 struct {
	StructureName     string
	Success           bool
	ResponseType      byte
	StructureTemplate map[string]any
}

func (*structureTemplateResponse475) ID() uint32  { return packet.IDStructureTemplateDataRequest }
func (*structureTemplateResponse475) legacyOnly() {}
func (pk *structureTemplateResponse475) Marshal(io protocol.IO) {
	io.String(&pk.StructureName)
	io.Bool(&pk.Success)
	if pk.Success {
		io.NBT(&pk.StructureTemplate, nbt.NetworkLittleEndian)
	}
	io.Uint8(&pk.ResponseType)
}

const (
	idPassengerJump     uint32 = 20
	idTickSync          uint32 = 23
	idCraftingEvent     uint32 = 53
	idPlayerInput       uint32 = 57
	idItemFrameDropItem uint32 = 71
	idAddEntity         uint32 = 127
	idRemoveEntity      uint32 = 128
)

type legacyOnlyPacket interface {
	legacyOnly()
}

type passengerJump struct{ JumpStrength int32 }

func (*passengerJump) ID() uint32                { return idPassengerJump }
func (*passengerJump) legacyOnly()               {}
func (pk *passengerJump) Marshal(io protocol.IO) { io.Varint32(&pk.JumpStrength) }

type tickSync struct {
	ClientRequestTimestamp, ServerReceptionTimestamp int64
}

func (*tickSync) ID() uint32  { return idTickSync }
func (*tickSync) legacyOnly() {}
func (pk *tickSync) Marshal(io protocol.IO) {
	io.Int64(&pk.ClientRequestTimestamp)
	io.Int64(&pk.ServerReceptionTimestamp)
}

type craftingEvent struct {
	WindowID     byte
	CraftingType int32
	RecipeUUID   uuid.UUID
	Input        []protocol.ItemInstance
	Output       []protocol.ItemInstance
}

func (*craftingEvent) ID() uint32  { return idCraftingEvent }
func (*craftingEvent) legacyOnly() {}
func (pk *craftingEvent) Marshal(io protocol.IO) {
	io.Uint8(&pk.WindowID)
	io.Varint32(&pk.CraftingType)
	io.UUID(&pk.RecipeUUID)
	protocol.FuncSlice(io, &pk.Input, io.ItemInstance)
	protocol.FuncSlice(io, &pk.Output, io.ItemInstance)
}

type playerInput struct {
	Movement mgl32.Vec2
	Jumping  bool
	Sneaking bool
}

func (*playerInput) ID() uint32  { return idPlayerInput }
func (*playerInput) legacyOnly() {}
func (pk *playerInput) Marshal(io protocol.IO) {
	io.Vec2(&pk.Movement)
	io.Bool(&pk.Jumping)
	io.Bool(&pk.Sneaking)
}

type itemFrameDropItem struct{ Position protocol.BlockPos }

func (*itemFrameDropItem) ID() uint32  { return idItemFrameDropItem }
func (*itemFrameDropItem) legacyOnly() {}
func (pk *itemFrameDropItem) Marshal(raw protocol.IO) {
	marshalUnsignedBlockPos(asWireIO(raw), &pk.Position)
}

type addEntity struct{ EntityNetworkID uint64 }

func (*addEntity) ID() uint32                { return idAddEntity }
func (*addEntity) legacyOnly()               {}
func (pk *addEntity) Marshal(io protocol.IO) { io.Varuint64(&pk.EntityNetworkID) }

type removeEntity struct{ EntityNetworkID uint64 }

func (*removeEntity) ID() uint32                { return idRemoveEntity }
func (*removeEntity) legacyOnly()               {}
func (pk *removeEntity) Marshal(io protocol.IO) { io.Varuint64(&pk.EntityNetworkID) }
