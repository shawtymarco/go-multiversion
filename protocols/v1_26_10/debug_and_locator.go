package v1_26_10

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
		protocol.OptionalFunc(raw, &shape.DimensionID, legacy.Varint32)

		var attached protocol.Optional[uint64]
		if value, ok := shape.AttachedToEntityID.Value(); ok {
			if value < 0 {
				legacy.InvalidValue(value, "debug drawer attached entity", "must not be negative")
				return
			}
			attached = protocol.Option(uint64(value))
		}
		protocol.OptionalFunc(raw, &attached, legacy.Varuint64)
		if legacy.reading {
			if value, ok := attached.Value(); ok {
				if value > uint64(^uint64(0)>>1) {
					legacy.InvalidValue(value, "debug drawer attached entity", "exceeds int64")
					return
				}
				shape.AttachedToEntityID = protocol.Option(int64(value))
			} else {
				shape.AttachedToEntityID = protocol.Optional[int64]{}
			}
			shape.MaxRenderDistance = protocol.Optional[float32]{}
		}
		marshalShapeData944(legacy, &shape.ExtraShapeData)
	})
}

func marshalShapeData944(io *wireIO, value *protocol.ShapeData) {
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
		marshalWaypoint944(legacy, &waypoint.Waypoint)
		legacy.Uint8(&waypoint.Action)
	})
}

func marshalWaypoint944(io *wireIO, waypoint *protocol.Waypoint) {
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
