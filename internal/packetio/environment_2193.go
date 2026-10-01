package packetio

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// Environment2193 is the frozen pre-2223 environment wire model. Older codecs
// marshal this view directly, without passing semantics through another version.
type Environment2193 struct {
	AttributeName                                string
	FromAttribute                                protocol.Optional[protocol.AttributeData]
	Attribute                                    protocol.AttributeData
	ToAttribute                                  protocol.Optional[protocol.AttributeData]
	CurrentTransitionTicks, TotalTransitionTicks uint32
	EaseType                                     int32
	LocalTransitionTicks                         uint32
	NoiseTransition                              bool
	NoiseAlignment                               protocol.NoiseAlignment
}

func EnvironmentFromNative(x *protocol.EnvironmentAttributeData) Environment2193 {
	value := Environment2193{AttributeName: x.AttributeName, Attribute: x.Attribute}
	switch x.PayloadType {
	case protocol.EnvironmentAttributePayloadTypeConstant:
	case protocol.EnvironmentAttributePayloadTypeTransition:
		value.FromAttribute, value.ToAttribute = protocol.Option(x.FromAttribute), protocol.Option(x.ToAttribute)
		value.Attribute = x.FromAttribute
		value.CurrentTransitionTicks = x.TransitionSettings.CurrentTransitionTicks
		value.TotalTransitionTicks = x.TransitionSettings.TotalTransitionTicks
		value.EaseType = x.TransitionSettings.EaseType
	case protocol.EnvironmentAttributePayloadTypeNoiseTransition:
		value.FromAttribute, value.ToAttribute = protocol.Option(x.FromAttribute), protocol.Option(x.ToAttribute)
		value.Attribute = x.FromAttribute
		value.CurrentTransitionTicks = x.NoiseTransitionSettings.CurrentTransitionTicks
		value.TotalTransitionTicks = x.NoiseTransitionSettings.TotalTransitionTicks
		value.EaseType = x.NoiseTransitionSettings.EaseType
		value.LocalTransitionTicks = x.NoiseTransitionSettings.LocalTransitionTicks
		value.NoiseTransition = true
		value.NoiseAlignment = x.NoiseTransitionSettings.NoiseAlignment
	}
	return value
}

func EnvironmentToNative(value Environment2193, x *protocol.EnvironmentAttributeData) {
	*x = protocol.EnvironmentAttributeData{AttributeName: value.AttributeName, Attribute: value.Attribute}
	from, hasFrom := value.FromAttribute.Value()
	to, hasTo := value.ToAttribute.Value()
	if !hasFrom {
		from = value.Attribute
	}
	if !hasTo {
		to = value.Attribute
	}
	if value.NoiseTransition {
		x.PayloadType = protocol.EnvironmentAttributePayloadTypeNoiseTransition
		x.FromAttribute, x.ToAttribute = from, to
		x.NoiseTransitionSettings = protocol.AttributeNoiseTransitionSettings{CurrentTransitionTicks: value.CurrentTransitionTicks, TotalTransitionTicks: value.TotalTransitionTicks, EaseType: value.EaseType, LocalTransitionTicks: value.LocalTransitionTicks, NoiseAlignment: value.NoiseAlignment}
	} else if hasFrom || hasTo {
		x.PayloadType = protocol.EnvironmentAttributePayloadTypeTransition
		x.FromAttribute, x.ToAttribute = from, to
		x.TransitionSettings = protocol.AttributeTransitionSettings{CurrentTransitionTicks: value.CurrentTransitionTicks, TotalTransitionTicks: value.TotalTransitionTicks, EaseType: value.EaseType}
	}
}

// LayerNoiseFromNative recovers the former layer-level noise name. Different
// per-attribute noise names cannot be represented by that older schema.
func LayerNoiseFromNative(layer *protocol.AttributeLayerData) protocol.Optional[string] {
	for _, value := range layer.EnvironmentAttributes {
		if value.PayloadType == protocol.EnvironmentAttributePayloadTypeNoiseTransition && value.NoiseTransitionSettings.NoiseName != "" {
			return protocol.Option(value.NoiseTransitionSettings.NoiseName)
		}
	}
	return protocol.Optional[string]{}
}
func LayerNoiseToNative(layer *protocol.AttributeLayerData, name protocol.Optional[string]) {
	value, _ := name.Value()
	for i := range layer.EnvironmentAttributes {
		if layer.EnvironmentAttributes[i].PayloadType == protocol.EnvironmentAttributePayloadTypeNoiseTransition {
			layer.EnvironmentAttributes[i].NoiseTransitionSettings.NoiseName = value
		}
	}
}

func RequiredSystemCategories(io protocol.IO, value *protocol.Optional[[]protocol.SystemCategory], reading bool) {
	categories, _ := value.Value()
	protocol.Slice(io, &categories)
	if reading {
		*value = protocol.Option(categories)
	}
}
