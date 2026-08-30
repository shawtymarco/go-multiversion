package v1_21_130

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalCameraInstruction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraInstruction)
	protocol.OptionalMarshaler(io.directional(), &pk.Set)
	protocol.OptionalFunc(io.directional(), &pk.Clear, io.Bool)
	protocol.OptionalMarshaler(io.directional(), &pk.Fade)
	protocol.OptionalMarshaler(io.directional(), &pk.Target)
	protocol.OptionalFunc(io.directional(), &pk.RemoveTarget, io.Bool)
	protocol.OptionalFunc(io.directional(), &pk.FieldOfView, func(value *protocol.CameraInstructionFieldOfView) {
		marshalCameraFieldOfView898(io, value)
	})
	protocol.OptionalFunc(io.directional(), &pk.Spline, func(value *protocol.CameraSplineInstruction) {
		marshalCameraSplineInstruction898(io, value)
	})
	protocol.OptionalFunc(io.directional(), &pk.AttachToEntity, io.Int64)
	protocol.OptionalFunc(io.directional(), &pk.DetachFromEntity, io.Bool)
}

func marshalCameraFieldOfView898(io *wireIO, value *protocol.CameraInstructionFieldOfView) {
	io.Float32(&value.FieldOfView)
	io.Float32(&value.EaseTime)
	easeType := uint8(value.EaseType)
	if !io.reading && value.EaseType > protocol.EasingTypeInOutElastic {
		easeType = uint8(protocol.EasingTypeLinear)
	}
	io.Uint8(&easeType)
	value.EaseType = int32(easeType)
	io.Bool(&value.Clear)
}

func marshalCameraSplineInstruction898(io *wireIO, value *protocol.CameraSplineInstruction) {
	io.Float32(&value.TotalTime)
	var easeType uint8
	io.Uint8(&easeType)
	protocol.FuncSlice(io.directional(), &value.Curve, io.Vec3)
	progress := make([]mgl32.Vec2, len(value.ProgressKeyFrames))
	if !io.reading {
		for index, keyframe := range value.ProgressKeyFrames {
			progress[index] = mgl32.Vec2{keyframe.Value, keyframe.Time}
		}
	}
	protocol.FuncSlice(io.directional(), &progress, io.Vec2)
	if io.reading {
		value.ProgressKeyFrames = make([]protocol.CameraProgressOption, len(progress))
		for index, keyframe := range progress {
			value.ProgressKeyFrames[index] = protocol.CameraProgressOption{Value: keyframe[0], Time: keyframe[1], EaseType: int32(easeType)}
		}
	}
	protocol.FuncIOSlice(io.directional(), &value.RotationOptions, func(raw protocol.IO, option *protocol.CameraRotationOption) {
		marshalCameraRotationOption898(asWireIO(raw), option)
	})
	if io.reading {
		value.SplineType = protocol.Optional[uint8]{}
		value.SplineIdentifier = protocol.Optional[string]{}
		value.LoadFromJson = protocol.Optional[bool]{}
	}
}

func marshalCameraRotationOption898(io *wireIO, value *protocol.CameraRotationOption) {
	io.Vec3(&value.Value)
	io.Float32(&value.Time)
	if io.reading {
		value.EaseType = protocol.EasingTypeLinear
	}
}
