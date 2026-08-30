package v1_21_130

import (
	"fmt"
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalDebugDrawer(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PrimitiveShapes)
	protocol.FuncIOSlice(io.directional(), &pk.Shapes, func(raw protocol.IO, shape *protocol.PrimitiveShape) {
		legacy := asWireIO(raw)
		legacy.Varuint64(&shape.NetworkID)
		protocol.OptionalFunc(raw, &shape.Type, legacy.Uint8)
		protocol.OptionalFunc(raw, &shape.Location, legacy.Vec3)
		protocol.OptionalFunc(raw, &shape.Scale, legacy.Float32)
		protocol.OptionalFunc(raw, &shape.Rotation, legacy.Vec3)
		protocol.OptionalFunc(raw, &shape.TotalTimeLeft, legacy.Float32)
		protocol.OptionalFunc(raw, &shape.Colour, legacy.BEARGB)
		dimensionID, _ := shape.DimensionID.Value()
		legacy.Varint32(&dimensionID)
		if legacy.reading {
			shape.DimensionID = protocol.Option(dimensionID)
			shape.AttachedToEntityID = protocol.Optional[int64]{}
			shape.MaxRenderDistance = protocol.Optional[float32]{}
		}
		marshalShapeData898(legacy, &shape.ExtraShapeData)
	})
}

func marshalShapeData898(io *wireIO, value *protocol.ShapeData) {
	var shapeType uint32
	if !io.reading {
		switch (*value).(type) {
		case *protocol.LastShape:
			shapeType = protocol.ShapeDataLast
		case *protocol.ArrowShape:
			shapeType = protocol.ShapeDataArrow
		case *protocol.TextShape:
			shapeType = protocol.ShapeDataText
		case *protocol.BoxShape:
			shapeType = protocol.ShapeDataBox
		case *protocol.LineShape:
			shapeType = protocol.ShapeDataLine
		case *protocol.SphereShape:
			shapeType = protocol.ShapeDataSphere
		default:
			io.UnknownEnumOption(fmt.Sprintf("%T", *value), "debug drawer shape data type")
			return
		}
	}
	io.Varuint32(&shapeType)
	if io.reading {
		switch shapeType {
		case protocol.ShapeDataLast:
			*value = &protocol.LastShape{}
		case protocol.ShapeDataArrow:
			*value = &protocol.ArrowShape{}
		case protocol.ShapeDataText:
			*value = &protocol.TextShape{}
		case protocol.ShapeDataBox:
			*value = &protocol.BoxShape{}
		case protocol.ShapeDataLine:
			*value = &protocol.LineShape{}
		case protocol.ShapeDataSphere:
			*value = &protocol.SphereShape{}
		default:
			io.UnknownEnumOption(shapeType, "debug drawer shape data type")
			return
		}
	}
	if text, ok := (*value).(*protocol.TextShape); ok {
		io.String(&text.Text)
		if io.reading {
			text.UseRotation = false
			text.BackgroundColour = protocol.Optional[color.RGBA]{}
			text.DepthTest = false
			text.ShowBackface = false
			text.ShowBackfaceText = false
		}
		return
	}
	(*value).Marshal(io.directional())
}

func marshalLocatorBar(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.LocatorBar)
	protocol.FuncIOSlice(io.directional(), &pk.Waypoints, func(raw protocol.IO, waypoint *protocol.LocatorBarWaypoint) {
		legacy := asWireIO(raw)
		legacy.UUID(&waypoint.GroupHandle)
		marshalWaypoint898(legacy, &waypoint.Waypoint)
		legacy.Uint8(&waypoint.Action)
	})
}

func marshalWaypoint898(io *wireIO, waypoint *protocol.Waypoint) {
	updateFlag := waypoint.UpdateFlag
	if !io.reading {
		updateFlag &^= protocol.WaypointUpdateFlagTextureID
	}
	io.Uint32(&updateFlag)
	protocol.OptionalFunc(io.directional(), &waypoint.Visible, io.Bool)
	protocol.OptionalMarshaler(io.directional(), &waypoint.WorldPosition)
	var textureID protocol.Optional[uint32]
	protocol.OptionalFunc(io.directional(), &textureID, io.Uint32)
	protocol.OptionalFunc(io.directional(), &waypoint.Colour, io.Int32)
	protocol.OptionalFunc(io.directional(), &waypoint.ClientPositionAuthority, io.Bool)
	protocol.OptionalFunc(io.directional(), &waypoint.ActorUniqueID, io.Varint64)
	if io.reading {
		waypoint.UpdateFlag = updateFlag &^ protocol.WaypointUpdateFlagTextureID
		waypoint.TexturePath = protocol.Optional[string]{}
		waypoint.IconSize = protocol.Optional[mgl32.Vec2]{}
	}
}
