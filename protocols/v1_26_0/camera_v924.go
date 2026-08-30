package v1_26_0

import (
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
		marshalCameraFieldOfView924(io, value)
	})
	protocol.OptionalFunc(io.directional(), &pk.Spline, func(value *protocol.CameraSplineInstruction) {
		marshalCameraSplineInstruction924(io, value)
	})
	protocol.OptionalFunc(io.directional(), &pk.AttachToEntity, io.Int64)
	protocol.OptionalFunc(io.directional(), &pk.DetachFromEntity, io.Bool)
}

func marshalCameraFieldOfView924(io *wireIO, value *protocol.CameraInstructionFieldOfView) {
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

func marshalCameraSplineInstruction924(io *wireIO, value *protocol.CameraSplineInstruction) {
	io.Float32(&value.TotalTime)
	protocol.OptionalFunc(io.directional(), &value.SplineType, io.Uint8)
	protocol.FuncSlice(io.directional(), &value.Curve, io.Vec3)
	protocol.FuncIOSlice(io.directional(), &value.ProgressKeyFrames, func(raw protocol.IO, keyframe *protocol.CameraProgressOption) {
		marshalCameraProgressOption924(asWireIO(raw), keyframe)
	})
	protocol.FuncIOSlice(io.directional(), &value.RotationOptions, func(raw protocol.IO, option *protocol.CameraRotationOption) {
		marshalCameraRotationOption924(asWireIO(raw), option)
	})
	if io.reading {
		value.SplineIdentifier = protocol.Optional[string]{}
		value.LoadFromJson = protocol.Optional[bool]{}
	}
}

func marshalCameraProgressOption924(io *wireIO, value *protocol.CameraProgressOption) {
	io.Float32(&value.Value)
	io.Float32(&value.Time)
	easeType := protocol.Option(uint8(value.EaseType))
	if !io.reading && value.EaseType > protocol.EasingTypeInOutElastic {
		easeType = protocol.Option(uint8(protocol.EasingTypeLinear))
	}
	protocol.OptionalFunc(io.directional(), &easeType, io.Uint8)
	if encoded, ok := easeType.Value(); ok {
		value.EaseType = int32(encoded)
	} else if io.reading {
		value.EaseType = protocol.EasingTypeLinear
	}
}

func marshalCameraRotationOption924(io *wireIO, value *protocol.CameraRotationOption) {
	io.Vec3(&value.Value)
	io.Float32(&value.Time)
	easeType := protocol.Option(uint8(value.EaseType))
	if !io.reading && value.EaseType > protocol.EasingTypeInOutElastic {
		easeType = protocol.Option(uint8(protocol.EasingTypeLinear))
	}
	protocol.OptionalFunc(io.directional(), &easeType, io.Uint8)
	if encoded, ok := easeType.Value(); ok {
		value.EaseType = int32(encoded)
	} else if io.reading {
		value.EaseType = protocol.EasingTypeLinear
	}
}

func marshalCameraSpline(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraSpline)
	protocol.FuncIOSlice(io.directional(), &pk.Splines, func(raw protocol.IO, definition *protocol.CameraSplineDefinition) {
		legacy := asWireIO(raw)
		legacy.String(&definition.Name)
		instruction := protocol.CameraSplineInstruction{
			TotalTime:         definition.TotalTime,
			Curve:             definition.ControlPoints,
			ProgressKeyFrames: definition.ProgressKeyFrames,
			RotationOptions:   definition.RotationKeyFrames,
		}
		if splineType, ok := definition.SplineType.Value(); ok {
			switch splineType {
			case protocol.SplineTypeLinear:
				instruction.SplineType = protocol.Option(uint8(1))
			default:
				instruction.SplineType = protocol.Option(uint8(0))
			}
		}
		marshalCameraSplineInstruction924(legacy, &instruction)
		if legacy.reading {
			definition.TotalTime = instruction.TotalTime
			definition.ControlPoints = instruction.Curve
			definition.ProgressKeyFrames = instruction.ProgressKeyFrames
			definition.RotationKeyFrames = instruction.RotationOptions
			if splineType, ok := instruction.SplineType.Value(); ok {
				if splineType == 1 {
					definition.SplineType = protocol.Option(protocol.SplineTypeLinear)
				} else {
					definition.SplineType = protocol.Option(protocol.SplineTypeCatmullRom)
				}
			} else {
				definition.SplineType = protocol.Optional[string]{}
			}
		}
	})
}
