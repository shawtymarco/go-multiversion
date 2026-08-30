package v1_26_0

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func uBlockPos(io *wireIO, position *protocol.BlockPos) {
	io.Varint32(&position[0])
	y := uint32(position[1])
	io.Varuint32(&y)
	position[1] = int32(y)
	io.Varint32(&position[2])
}

func marshalBlockChangeEntry924(io *wireIO, entry *protocol.BlockChangeEntry) {
	uBlockPos(io, &entry.BlockPos)
	io.Varuint32(&entry.BlockRuntimeID)
	io.Varuint32(&entry.Flags)
	io.Varuint64(&entry.SyncedUpdateEntityUniqueID)
	io.Varuint32(&entry.SyncedUpdateType)
}

func marshalStructureSettings924(io *wireIO, settings *protocol.StructureSettings) {
	io.String(&settings.PaletteName)
	io.Bool(&settings.IgnoreEntities)
	io.Bool(&settings.IgnoreBlocks)
	io.Bool(&settings.AllowNonTickingChunks)
	uBlockPos(io, &settings.Size)
	uBlockPos(io, &settings.Offset)
	io.ActorUniqueID(&settings.LastEditingPlayerUniqueID)
	io.Uint8(&settings.Rotation)
	io.Uint8(&settings.Mirror)
	io.Uint8(&settings.AnimationMode)
	io.Float32(&settings.AnimationDuration)
	io.Float32(&settings.Integrity)
	io.Uint32(&settings.Seed)
	io.Vec3(&settings.Pivot)
}

func marshalAddVolumeEntity(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AddVolumeEntity)
	io.ActorRuntimeIDVaruint32(&pk.EntityRuntimeID)
	io.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	io.String(&pk.EncodingIdentifier)
	io.String(&pk.InstanceIdentifier)
	uBlockPos(io, &pk.Bounds[0])
	uBlockPos(io, &pk.Bounds[1])
	io.Varint32(&pk.Dimension)
	io.String(&pk.EngineVersion)
}

func marshalBlockActorData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BlockActorData)
	uBlockPos(io, &pk.Position)
	io.NBT(&pk.NBTData, nbt.NetworkLittleEndian)
}

func marshalBlockEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BlockEvent)
	uBlockPos(io, &pk.Position)
	io.Varint32(&pk.EventType)
	io.Varint32(&pk.EventData)
}

func marshalCommandBlockUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandBlockUpdate)
	io.Bool(&pk.Block)
	if pk.Block {
		uBlockPos(io, &pk.Position)
		io.Varuint32(&pk.Mode)
		io.Bool(&pk.NeedsRedstone)
		io.Bool(&pk.Conditional)
	} else {
		io.ActorRuntimeID(&pk.MinecartEntityRuntimeID)
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
	uBlockPos(io, &pk.ContainerPosition)
	io.ActorUniqueID(&pk.ContainerEntityUniqueID)
}

func marshalLecternUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LecternUpdate)
	io.Uint8(&pk.Page)
	io.Uint8(&pk.PageCount)
	uBlockPos(io, &pk.Position)
}

func marshalOpenSign(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.OpenSign)
	uBlockPos(io, &pk.Position)
	io.Bool(&pk.FrontSide)
}

func marshalPlayerAction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAction)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.Varint32(&pk.ActionType)
	uBlockPos(io, &pk.BlockPosition)
	uBlockPos(io, &pk.ResultPosition)
	io.Varint32(&pk.BlockFace)
}

func marshalSetSpawnPosition(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetSpawnPosition)
	io.Varint32(&pk.SpawnType)
	uBlockPos(io, &pk.Position)
	io.Varint32(&pk.Dimension)
	uBlockPos(io, &pk.SpawnPosition)
}

func marshalStructureTemplateDataRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StructureTemplateDataRequest)
	io.String(&pk.StructureName)
	uBlockPos(io, &pk.Position)
	marshalStructureSettings924(io, &pk.Settings)
	io.Uint8(&pk.RequestType)
}

func marshalUpdateBlock(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateBlock)
	uBlockPos(io, &pk.Position)
	io.Varuint32(&pk.NewBlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
}

func marshalUpdateBlockSynced(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateBlockSynced)
	uBlockPos(io, &pk.Position)
	io.Varuint32(&pk.NewBlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
	io.ActorUniqueIDVaruint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.TransitionType)
}

func marshalUpdateSubChunkBlocks(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateSubChunkBlocks)
	uBlockPos(io, &pk.Position)
	protocol.FuncIOSlice(io.directional(), &pk.Blocks, func(raw protocol.IO, entry *protocol.BlockChangeEntry) {
		marshalBlockChangeEntry924(asWireIO(raw), entry)
	})
	protocol.FuncIOSlice(io.directional(), &pk.Extra, func(raw protocol.IO, entry *protocol.BlockChangeEntry) {
		marshalBlockChangeEntry924(asWireIO(raw), entry)
	})
}

func marshalUpdateClientInputLocks(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.UpdateClientInputLocks)
	io.Varuint32(&pk.Locks)
	var discardedPosition mgl32.Vec3
	io.Vec3(&discardedPosition)
}

func marshalClientBoundDataDrivenUICloseScreen(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundDataDrivenUICloseScreen)
	if io.reading {
		pk.FormID = protocol.Optional[uint32]{}
	}
}

func marshalClientBoundDataDrivenUIShowScreen(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundDataDrivenUIShowScreen)
	io.String(&pk.ScreenID)
	if io.reading {
		pk.FormID = 0
		pk.DataInstanceID = protocol.Optional[uint32]{}
	}
}

func marshalVoxelShapes(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.VoxelShapes)
	protocol.Slice(io.directional(), &pk.Shapes)
	protocol.Slice(io.directional(), &pk.NameMap)
	if io.reading {
		pk.CustomShapeCount = 0
	}
}
