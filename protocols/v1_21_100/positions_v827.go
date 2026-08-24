package v1_21_100

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAddVolumeEntity(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddVolumeEntity)
	io.Varuint32(&pk.EntityRuntimeID)
	io.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	io.String(&pk.EncodingIdentifier)
	io.String(&pk.InstanceIdentifier)
	io.UBlockPos(&pk.Bounds[0])
	io.UBlockPos(&pk.Bounds[1])
	io.Varint32(&pk.Dimension)
	io.String(&pk.EngineVersion)
}

func marshalBlockActorData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BlockActorData)
	io.UBlockPos(&pk.Position)
	io.NBT(&pk.NBTData, nbt.NetworkLittleEndian)
}

func marshalBlockEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BlockEvent)
	io.UBlockPos(&pk.Position)
	io.Varint32(&pk.EventType)
	io.Varint32(&pk.EventData)
}

func marshalCommandBlockUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandBlockUpdate)
	io.Bool(&pk.Block)
	if pk.Block {
		io.UBlockPos(&pk.Position)
		io.Varuint32(&pk.Mode)
		io.Bool(&pk.NeedsRedstone)
		io.Bool(&pk.Conditional)
	} else {
		io.Varuint64(&pk.MinecartEntityRuntimeID)
	}
	io.String(&pk.Command)
	io.String(&pk.LastOutput)
	io.String(&pk.Name)
	io.String(&pk.FilteredName)
	io.Bool(&pk.ShouldTrackOutput)
	io.Uint32(&pk.TickDelay)
	io.Bool(&pk.ExecuteOnFirstTick)
}

func marshalContainerOpen(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ContainerOpen)
	io.Uint8(&pk.WindowID)
	io.Uint8(&pk.ContainerType)
	io.UBlockPos(&pk.ContainerPosition)
	io.Varint64(&pk.ContainerEntityUniqueID)
}

func marshalLecternUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LecternUpdate)
	io.Uint8(&pk.Page)
	io.Uint8(&pk.PageCount)
	io.UBlockPos(&pk.Position)
}

func marshalOpenSign(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.OpenSign)
	io.UBlockPos(&pk.Position)
	io.Bool(&pk.FrontSide)
}

func marshalPlayerAction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAction)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Varint32(&pk.ActionType)
	io.UBlockPos(&pk.BlockPosition)
	io.UBlockPos(&pk.ResultPosition)
	io.Varint32(&pk.BlockFace)
}

func marshalSetSpawnPosition(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetSpawnPosition)
	io.Varint32(&pk.SpawnType)
	io.UBlockPos(&pk.Position)
	io.Varint32(&pk.Dimension)
	io.UBlockPos(&pk.SpawnPosition)
}

func marshalStructureTemplateDataRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StructureTemplateDataRequest)
	io.String(&pk.StructureName)
	io.UBlockPos(&pk.Position)
	marshalStructureSettings827(io, &pk.Settings)
	io.Uint8(&pk.RequestType)
}

func marshalUpdateBlock(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateBlock)
	io.UBlockPos(&pk.Position)
	io.Varuint32(&pk.NewBlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
}

func marshalUpdateBlockSynced(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateBlockSynced)
	io.UBlockPos(&pk.Position)
	io.Varuint32(&pk.NewBlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
	io.Varuint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.TransitionType)
}

func marshalUpdateSubChunkBlocks(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateSubChunkBlocks)
	io.UBlockPos(&pk.Position)
	protocol.FuncIOSlice(io.directional(), &pk.Blocks, marshalBlockChangeEntry827)
	protocol.FuncIOSlice(io.directional(), &pk.Extra, marshalBlockChangeEntry827)
}

func marshalBlockChangeEntry827(raw protocol.IO, entry *protocol.BlockChangeEntry) {
	io := asWireIO(raw)
	io.UBlockPos(&entry.BlockPos)
	io.Varuint32(&entry.BlockRuntimeID)
	io.Varuint32(&entry.Flags)
	io.Varuint64(&entry.SyncedUpdateEntityUniqueID)
	io.Varuint32(&entry.SyncedUpdateType)
}

func marshalStructureSettings827(io *wireIO, settings *protocol.StructureSettings) {
	io.String(&settings.PaletteName)
	io.Bool(&settings.IgnoreEntities)
	io.Bool(&settings.IgnoreBlocks)
	io.Bool(&settings.AllowNonTickingChunks)
	io.UBlockPos(&settings.Size)
	io.UBlockPos(&settings.Offset)
	io.Varint64(&settings.LastEditingPlayerUniqueID)
	io.Uint8(&settings.Rotation)
	io.Uint8(&settings.Mirror)
	io.Uint8(&settings.AnimationMode)
	io.Float32(&settings.AnimationDuration)
	io.Float32(&settings.Integrity)
	io.Uint32(&settings.Seed)
	io.Vec3(&settings.Pivot)
}

func marshalBookEdit(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BookEdit)
	actionType, inventorySlot := uint8(pk.ActionType), uint8(pk.InventorySlot)
	io.Uint8(&actionType)
	io.Uint8(&inventorySlot)
	pk.ActionType, pk.InventorySlot = uint32(actionType), int32(inventorySlot)
	switch actionType {
	case packet.BookActionReplacePage, packet.BookActionAddPage:
		page := uint8(pk.PageNumber)
		io.Uint8(&page)
		pk.PageNumber = int32(page)
		io.String(&pk.Text)
		io.String(&pk.PhotoName)
	case packet.BookActionDeletePage:
		page := uint8(pk.PageNumber)
		io.Uint8(&page)
		pk.PageNumber = int32(page)
	case packet.BookActionSwapPages:
		page, secondary := uint8(pk.PageNumber), uint8(pk.SecondaryPageNumber)
		io.Uint8(&page)
		io.Uint8(&secondary)
		pk.PageNumber, pk.SecondaryPageNumber = int32(page), int32(secondary)
	case packet.BookActionSign:
		io.String(&pk.Title)
		io.String(&pk.Author)
		io.String(&pk.XUID)
	default:
		io.UnknownEnumOption(actionType, "book edit action type")
	}
}

func marshalClientCacheBlobStatus(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientCacheBlobStatus)
	missCount, hitCount := uint32(len(pk.MissHashes)), uint32(len(pk.HitHashes))
	io.Varuint32(&missCount)
	io.Varuint32(&hitCount)
	protocol.FuncSliceOfLen(io.directional(), missCount, &pk.MissHashes, io.Uint64)
	protocol.FuncSliceOfLen(io.directional(), hitCount, &pk.HitHashes, io.Uint64)
}
