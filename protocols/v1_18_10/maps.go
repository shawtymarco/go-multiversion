package v1_18_10

import (
	"image/color"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	mapUpdateTexture = 1 << (iota + 1)
	mapUpdateDecoration
	mapUpdateInitialisation
)

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
	if flags&mapUpdateInitialisation != 0 {
		values, _ := pk.MapsIncludedIn.Value()
		protocol.FuncSlice(io.directional(), &values, io.Varint64)
		pk.MapsIncludedIn = protocol.Option(values)
	} else if io.reading {
		pk.MapsIncludedIn = protocol.Optional[[]int64]{}
	}
	if flags&(mapUpdateInitialisation|mapUpdateDecoration|mapUpdateTexture) != 0 {
		value, _ := pk.Scale.Value()
		io.Uint8(&value)
		pk.Scale = protocol.Option(value)
	} else if io.reading {
		pk.Scale = protocol.Optional[byte]{}
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
	if io.reading {
		pk.Origin = protocol.BlockPos{}
	}
}

func marshalMapTrackedObject(raw protocol.IO, object *protocol.MapTrackedObject) {
	io := asWireIO(raw)
	io.Int32(&object.Type)
	switch object.Type {
	case protocol.MapObjectTypeEntity:
		entityUniqueID, _ := object.EntityUniqueID.Value()
		io.Varint64(&entityUniqueID)
		object.EntityUniqueID = protocol.Option(entityUniqueID)
	case protocol.MapObjectTypeBlock:
		blockPosition, _ := object.BlockPosition.Value()
		io.BlockPos(&blockPosition)
		object.BlockPosition = protocol.Option(blockPosition)
	default:
		io.UnknownEnumOption(object.Type, "map tracked object type")
	}
}

func marshalMapDecoration(raw protocol.IO, decoration *protocol.MapDecoration) {
	io := asWireIO(raw)
	io.Uint8(&decoration.Type)
	io.Uint8(&decoration.Rotation)
	io.Uint8(&decoration.X)
	io.Uint8(&decoration.Y)
	io.String(&decoration.Label)
	marshalVarRGBA(io, &decoration.Colour)
}

func marshalVarRGBA(raw protocol.IO, value *color.RGBA) {
	io := asWireIO(raw)
	packed := uint32(value.R) | uint32(value.G)<<8 | uint32(value.B)<<16 | uint32(value.A)<<24
	io.Varuint32(&packed)
	*value = color.RGBA{R: byte(packed), G: byte(packed >> 8), B: byte(packed >> 16), A: byte(packed >> 24)}
}
