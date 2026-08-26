package v1_26_20

import (
	"image/color"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/packetio"
)

const (
	mapUpdateTexture = 1 << (iota + 1)
	mapUpdateDecoration
	mapUpdateInitialisation
)

func marshalAnvilDamage(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AnvilDamage)
	var discardedDamage uint8
	io.Uint8(&discardedDamage)
	io.BlockPos(&pk.AnvilPosition)
}

func marshalClientBoundMapItemData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundMapItemData)
	io.Varint64(&pk.MapID)

	flags := uint32(0)
	if !io.reading {
		if _, ok := pk.MapsIncludedIn.Value(); ok {
			flags |= mapUpdateInitialisation
		}
		if _, ok := pk.TrackedObjects.Value(); ok {
			flags |= mapUpdateDecoration
		}
		if _, ok := pk.Pixels.Value(); ok {
			flags |= mapUpdateTexture
		}
	}
	io.Varuint32(&flags)
	io.Uint8(&pk.Dimension)
	io.Bool(&pk.LockedMap)
	io.BlockPos(&pk.Origin)

	if flags&mapUpdateInitialisation != 0 {
		values, _ := pk.MapsIncludedIn.Value()
		protocol.FuncSlice(io.directional(), &values, io.Varint64)
		pk.MapsIncludedIn = protocol.Option(values)
	}
	if flags&(mapUpdateInitialisation|mapUpdateDecoration|mapUpdateTexture) != 0 {
		value, _ := pk.Scale.Value()
		io.Uint8(&value)
		pk.Scale = protocol.Option(value)
	}
	if flags&mapUpdateDecoration != 0 {
		tracked, _ := pk.TrackedObjects.Value()
		protocol.FuncIOSlice(io.directional(), &tracked, marshalMapTrackedObject)
		pk.TrackedObjects = protocol.Option(tracked)
		decorations, _ := pk.Decorations.Value()
		protocol.FuncIOSlice(io.directional(), &decorations, marshalMapDecoration)
		pk.Decorations = protocol.Option(decorations)
	}
	if flags&mapUpdateTexture != 0 {
		width, _ := pk.Width.Value()
		height, _ := pk.Height.Value()
		xOffset, _ := pk.XOffset.Value()
		yOffset, _ := pk.YOffset.Value()
		io.Varint32(&width)
		io.Varint32(&height)
		io.Varint32(&xOffset)
		io.Varint32(&yOffset)
		pixels, _ := pk.Pixels.Value()
		protocol.FuncIOSlice(io.directional(), &pixels, marshalVarRGBA)
		pk.Width, pk.Height = protocol.Option(width), protocol.Option(height)
		pk.XOffset, pk.YOffset = protocol.Option(xOffset), protocol.Option(yOffset)
		pk.Pixels = protocol.Option(pixels)
	}
}

func marshalMapTrackedObject(raw protocol.IO, object *protocol.MapTrackedObject) {
	io := asWireIO(raw)
	io.Int32(&object.Type)
	switch object.Type {
	case protocol.MapObjectTypeEntity:
		entityUniqueID, _ := object.EntityUniqueID.Value()
		io.Varint64(&entityUniqueID)
		if io.reading {
			object.EntityUniqueID = protocol.Option(entityUniqueID)
			object.BlockPosition = protocol.Optional[protocol.BlockPos]{}
		}
	case protocol.MapObjectTypeBlock:
		blockPosition, _ := object.BlockPosition.Value()
		io.BlockPos(&blockPosition)
		if io.reading {
			object.EntityUniqueID = protocol.Optional[int64]{}
			object.BlockPosition = protocol.Option(blockPosition)
		}
	default:
		io.UnknownEnumOption(object.Type, "map tracked object type")
	}
}

func marshalMapDecoration(io protocol.IO, decoration *protocol.MapDecoration) {
	io.Uint8(&decoration.Type)
	io.Uint8(&decoration.Rotation)
	io.Uint8(&decoration.X)
	io.Uint8(&decoration.Y)
	io.String(&decoration.Label)
	marshalVarRGBA(io, &decoration.Colour)
}

func marshalVarRGBA(io protocol.IO, value *color.RGBA) {
	packed := uint32(value.R) | uint32(value.G)<<8 | uint32(value.B)<<16 | uint32(value.A)<<24
	io.Varuint32(&packed)
	*value = color.RGBA{R: byte(packed), G: byte(packed >> 8), B: byte(packed >> 16), A: byte(packed >> 24)}
}

func marshalClientboundUpdateSoundData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientboundUpdateSoundData)
	io.Uint64(&pk.ServerSoundHandle)
	event := "Stop"
	io.String(&event)
	if io.reading {
		pk.Stop = protocol.Option(protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateStop})
	}
}

func marshalClientMovementPredictionSync(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientMovementPredictionSync)
	// EntityDataFlagNotPickableFromInside was added after protocol 975. Its
	// numeric index is therefore the protocol-975 flag count.
	io.Bitset(&pk.ActorFlags, protocol.EntityDataFlagNotPickableFromInside)
	io.Float32(&pk.BoundingBoxScale)
	io.Float32(&pk.BoundingBoxWidth)
	io.Float32(&pk.BoundingBoxHeight)
	io.Float32(&pk.MovementSpeed)
	io.Float32(&pk.UnderwaterMovementSpeed)
	io.Float32(&pk.LavaMovementSpeed)
	io.Float32(&pk.JumpStrength)
	io.Float32(&pk.Health)
	io.Float32(&pk.Hunger)
	io.Float32(&pk.FrictionModifier)
	io.Float32(&pk.Bounciness)
	io.Float32(&pk.AirDragModifier)
	io.Varint64(&pk.EntityUniqueID)
	io.Bool(&pk.Flying)
}

func marshalDimensionData(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.DimensionData)
	protocol.FuncIOSlice(io.directional(), &pk.Definitions, func(raw protocol.IO, definition *protocol.DimensionDefinition) {
		legacy := asWireIO(raw)
		legacy.String(&definition.Name)
		legacy.Varint32(&definition.Range[0])
		legacy.Varint32(&definition.Range[1])
		legacy.Varint32(&definition.Generator)
		legacy.Varint32(&definition.DimensionType)
	})
}

func marshalInventoryContent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventoryContent)
	io.Varuint32(&pk.WindowID)
	protocol.FuncSlice(io.directional(), &pk.Content, func(item *protocol.ItemInstance) {
		marshalItemInstance(io, item)
	})
	protocol.Single(io.directional(), &pk.Container)
	marshalItemInstance(io, &pk.StorageItem)
}

func marshalHurtArmour(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.HurtArmour)
	io.Varint32(&pk.Cause)
	io.Varint32(&pk.Damage)
	slots := int64(pk.ArmourSlots)
	io.Varint64(&slots)
	pk.ArmourSlots = uint64(slots)
}

func marshalInventorySlot(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventorySlot)
	io.Varuint32(&pk.WindowID)
	io.Varuint32(&pk.Slot)
	protocol.OptionalMarshaler(io.directional(), &pk.Container)
	protocol.OptionalFunc(io.directional(), &pk.StorageItem, func(item *protocol.ItemInstance) {
		marshalItemInstanceNew(io, item)
	})
	marshalItemInstanceNew(io, &pk.NewItem)
}

func marshalMobArmourEquipment(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MobArmourEquipment)
	io.Varuint64(&pk.EntityRuntimeID)
	marshalItemInstance(io, &pk.Helmet)
	marshalItemInstance(io, &pk.Chestplate)
	marshalItemInstance(io, &pk.Leggings)
	marshalItemInstance(io, &pk.Boots)
	marshalItemInstance(io, &pk.Body)
}

func marshalMobEquipment(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.MobEquipment)
	io.Varuint64(&pk.EntityRuntimeID)
	marshalItemInstanceNew(io, &pk.NewItem)
	io.Uint8(&pk.InventorySlot)
	io.Uint8(&pk.HotBarSlot)
	io.Uint8(&pk.WindowID)
}

func marshalPlaySound(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlaySound)
	io.String(&pk.SoundName)
	io.SoundPos(&pk.Position)
	io.Float32(&pk.Volume)
	io.Float32(&pk.Pitch)
	protocol.OptionalFunc(io.directional(), &pk.Handle, io.Uint64)
}

func marshalPlayerSkin(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerSkin)
	io.UUID(&pk.UUID)
	marshalSkin(io, &pk.Skin)
	io.String(&pk.NewSkinName)
	io.String(&pk.OldSkinName)
	io.Bool(&pk.Skin.Trusted)
}

func marshalResourcePackClientResponse(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePackClientResponse)
	response := byte(pk.Response + 1)
	io.Uint8(&response)
	if io.reading {
		if response == 0 {
			io.InvalidValue(response, "resource pack response", "must be between 1 and 4")
			return
		}
		pk.Response = uint32(response - 1)
	}
	funcSliceUint16(io, &pk.PacksToDownload, func(value *string) { io.String(value) })
}

func marshalResourcePacksInfo(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePacksInfo)
	io.Bool(&pk.TexturePackRequired)
	io.Bool(&pk.HasAddons)
	io.Bool(&pk.HasScripts)
	io.Bool(&pk.ForceDisableVibrantVisuals)
	io.UUID(&pk.WorldTemplateUUID)
	io.String(&pk.WorldTemplateVersion)
	funcSliceUint16(io, &pk.TexturePacks, func(value *protocol.TexturePackInfo) { value.Marshal(io.directional()) })
}

func marshalServerBoundDiagnostics(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ServerBoundDiagnostics)
	io.Float32(&pk.AverageFramesPerSecond)
	io.Float32(&pk.AverageServerSimTickTime)
	io.Float32(&pk.AverageClientSimTickTime)
	io.Float32(&pk.AverageBeginFrameTime)
	io.Float32(&pk.AverageInputTime)
	io.Float32(&pk.AverageRenderTime)
	io.Float32(&pk.AverageEndFrameTime)
	io.Float32(&pk.AverageRemainderTimePercent)
	io.Float32(&pk.AverageUnaccountedTimePercent)
	protocol.Slice(io.directional(), &pk.MemoryCategoryValues)
	protocol.Slice(io.directional(), &pk.EntityDiagnostics)
	protocol.Slice(io.directional(), &pk.SystemDiagnostics)
	if io.reading {
		pk.WhiskerScopes = nil
	}
}

func marshalStructureBlockUpdate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StructureBlockUpdate)
	io.BlockPos(&pk.Position)
	io.String(&pk.StructureName)
	io.String(&pk.FilteredStructureName)
	io.String(&pk.DataField)
	io.Bool(&pk.IncludePlayers)
	io.Bool(&pk.ShowBoundingBox)
	io.Varint32(&pk.StructureBlockType)
	protocol.Single(io.directional(), &pk.Settings)
	redstoneMode := int32(pk.RedstoneSaveMode)
	io.Varint32(&redstoneMode)
	pk.RedstoneSaveMode = uint8(redstoneMode)
	io.Bool(&pk.ShouldTrigger)
	io.Bool(&pk.Waterlogged)
}

func marshalSubChunkRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SubChunkRequest)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Position[0])
	io.Varint32(&pk.Position[1])
	io.Varint32(&pk.Position[2])
	count := uint32(len(pk.Offsets))
	io.Uint32(&count)
	packetio.SubChunkOffsets(io.directional(), count, &pk.Offsets)
}

func marshalTransfer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Transfer)
	io.String(&pk.Address)
	io.Uint16(&pk.Port)
	io.Bool(&pk.ReloadWorld)
}

func funcSliceUint16[T any](io *wireIO, values *[]T, marshal func(*T)) {
	count := uint16(len(*values))
	io.Uint16(&count)
	if io.reading {
		*values = make([]T, count)
	}
	for i := range *values {
		marshal(&(*values)[i])
	}
}
