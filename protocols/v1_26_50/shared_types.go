package v1_26_50

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/internal/packetio"
)

var easingNames = [...]string{"linear", "spring", "in_quad", "out_quad", "in_out_quad", "in_cubic", "out_cubic", "in_out_cubic", "in_quart", "out_quart", "in_out_quart", "in_quint", "out_quint", "in_out_quint", "in_sine", "out_sine", "in_out_sine", "in_expo", "out_expo", "in_out_expo", "in_circ", "out_circ", "in_out_circ", "in_bounce", "out_bounce", "in_out_bounce", "in_back", "out_back", "in_out_back", "in_elastic", "out_elastic", "in_out_elastic", "inverse_lerp"}

func marshalAttributeLayer(io protocol.IO, x *protocol.AttributeLayerData, reading bool) {
	io.String(&x.Name)
	noise := packetio.LayerNoiseFromNative(x)
	protocol.OptionalFunc(io, &noise, io.String)
	io.Varint32(&x.DimensionID)
	protocol.Single(io, &x.Settings)
	protocol.FuncIOSlice(io, &x.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
		marshalEnvironmentAttribute(raw, value, reading)
	})
	if reading {
		packetio.LayerNoiseToNative(x, noise)
	}
}

func marshalEnvironmentAttribute(io protocol.IO, x *protocol.EnvironmentAttributeData, reading bool) {
	value := packetio.EnvironmentFromNative(x)
	io.String(&value.AttributeName)
	protocol.OptionalMarshaler(io, &value.FromAttribute)
	protocol.Single(io, &value.Attribute)
	protocol.OptionalMarshaler(io, &value.ToAttribute)
	io.Uint32(&value.CurrentTransitionTicks)
	io.Uint32(&value.TotalTransitionTicks)
	easing := "linear"
	if !reading {
		if value.EaseType < 0 || int(value.EaseType) >= len(easingNames) {
			io.InvalidValue(value.EaseType, "environment easing", "unknown type")
			return
		}
		easing = easingNames[value.EaseType]
	}
	io.String(&easing)
	if reading {
		value.EaseType = -1
		for i, name := range easingNames {
			if easing == name {
				value.EaseType = int32(i)
				break
			}
		}
		if value.EaseType < 0 {
			io.InvalidValue(easing, "environment easing", "unknown type")
			return
		}
	}
	io.Uint32(&value.LocalTransitionTicks)
	io.Bool(&value.NoiseTransition)
	protocol.Single(io, &value.NoiseAlignment)
	if reading {
		packetio.EnvironmentToNative(value, x)
	}
}

func marshalSoundUpdate(io protocol.IO, x *protocol.SoundDataUpdate, reading bool) {
	kind := uint32(x.Type)
	io.Varuint32(&kind)
	if kind > protocol.SoundDataUpdateResume {
		io.UnknownEnumOption(kind, "legacy sound update")
		return
	}
	if reading {
		x.Type = uint8(kind)
	}
	switch kind {
	case protocol.SoundDataUpdateSetVolume:
		io.Float32(&x.Volume)
	case protocol.SoundDataUpdateSetPitch:
		io.Float32(&x.Pitch)
	case protocol.SoundDataUpdateFade:
		io.Float32(&x.Duration)
		io.Float32(&x.TargetVolume)
	case protocol.SoundDataUpdateSeekTo:
		io.Float32(&x.Seconds)
	}
}

func marshalEntityDiagnostic(io protocol.IO, x *protocol.EntityDiagnosticTimingInfo, reading bool) {
	io.String(&x.DisplayName)
	io.String(&x.Entity)
	io.Uint64(&x.DurationNanos)
	io.Uint8(&x.PercentOfTotal)
	position, _ := x.Position.Value()
	dimension, _ := x.Dimension.Value()
	io.Vec3(&position)
	io.String(&dimension)
	if reading {
		x.Position = protocol.Option(position)
		x.Dimension = protocol.Option(dimension)
	}
}
