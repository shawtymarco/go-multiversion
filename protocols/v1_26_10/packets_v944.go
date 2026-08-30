package v1_26_10

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalBossEvent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BossEvent)
	io.Varint64(&pk.BossEntityUniqueID)
	eventType := uint32(pk.EventType)
	io.Varuint32(&eventType)
	pk.EventType = uint8(eventType)
	var screenDarkening uint16
	switch eventType {
	case packet.BossEventShow:
		io.String(&pk.BossBarTitle)
		io.String(&pk.FilteredBossBarTitle)
		io.Float32(&pk.HealthPercentage)
		io.Uint16(&screenDarkening)
		marshalLegacyBossAppearance(io, pk)
	case packet.BossEventRegisterPlayer, packet.BossEventUnregisterPlayer, packet.BossEventRequest:
		io.Varint64(&pk.PlayerUniqueID)
	case packet.BossEventHide:
	case packet.BossEventHealthPercentage:
		io.Float32(&pk.HealthPercentage)
	case packet.BossEventTitle:
		io.String(&pk.BossBarTitle)
		io.String(&pk.FilteredBossBarTitle)
	case packet.BossEventAppearanceProperties:
		io.Uint16(&screenDarkening)
		marshalLegacyBossAppearance(io, pk)
	case packet.BossEventTexture:
		marshalLegacyBossAppearance(io, pk)
	default:
		io.UnknownEnumOption(eventType, "boss event type")
	}
}

func marshalLegacyBossAppearance(io *wireIO, pk *packet.BossEvent) {
	colour, overlay := uint32(pk.Colour), uint32(pk.Overlay)
	if !io.reading {
		switch colour {
		case packet.BossEventColourWhite:
			colour = 6
		case packet.BossEventColourRebeccaPurple:
			colour = packet.BossEventColourPurple
		}
	}
	io.Varuint32(&colour)
	io.Varuint32(&overlay)
	if io.reading {
		if colour == 6 {
			colour = packet.BossEventColourWhite
		}
		pk.Colour, pk.Overlay = uint8(colour), uint8(overlay)
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

func marshalGraphicsOverrideParameter(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.GraphicsOverrideParameter)
	protocol.Slice(io.directional(), &pk.Values)
	protocol.OptionalFunc(io.directional(), &pk.FloatValue, io.Float32)
	protocol.OptionalFunc(io.directional(), &pk.Vec3Value, io.Vec3)
	io.String(&pk.BiomeIdentifier)
	io.Uint8(&pk.ParameterType)
	io.Bool(&pk.Reset)
	if io.reading {
		pk.PlayerIdentifier = protocol.Optional[string]{}
	}
}

var closeReasonToLegacy = map[string]uint8{
	packet.DataDrivenScreenCloseReasonProgrammaticClose:    0,
	packet.DataDrivenScreenCloseReasonProgrammaticCloseAll: 1,
	packet.DataDrivenScreenCloseReasonClientCanceled:       2,
	packet.DataDrivenScreenCloseReasonUserBusy:             3,
	packet.DataDrivenScreenCloseReasonInvalidForm:          4,
}

var closeReasonFromLegacy = map[uint8]string{
	0: packet.DataDrivenScreenCloseReasonProgrammaticClose,
	1: packet.DataDrivenScreenCloseReasonProgrammaticCloseAll,
	2: packet.DataDrivenScreenCloseReasonClientCanceled,
	3: packet.DataDrivenScreenCloseReasonUserBusy,
	4: packet.DataDrivenScreenCloseReasonInvalidForm,
}

func marshalServerBoundDataDrivenScreenClosed(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ServerBoundDataDrivenScreenClosed)
	formID := protocol.Option(pk.FormID)
	protocol.OptionalFunc(io.directional(), &formID, io.Uint32)
	if io.reading {
		pk.FormID, _ = formID.Value()
	}
	reason, ok := closeReasonToLegacy[pk.CloseReason]
	if io.reading {
		reason = 0
	}
	io.Uint8(&reason)
	if io.reading {
		pk.CloseReason, ok = closeReasonFromLegacy[reason]
	}
	if !ok {
		io.UnknownEnumOption(reason, "data-driven screen close reason")
	}
}

func marshalClientBoundDataStore(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundDataStore)
	protocol.FuncIOSlice(io.directional(), &pk.Updates, func(raw protocol.IO, entry *protocol.DataStoreChangeEntry) {
		legacy := asWireIO(raw)
		legacy.Uint32(&entry.ChangeType)
		switch entry.ChangeType {
		case protocol.DataStoreChangeTypeUpdate:
			marshalDataStoreUpdate944(legacy, &entry.Update)
		case protocol.DataStoreChangeTypeChange:
			legacy.String(&entry.Change.DataStoreName)
			legacy.String(&entry.Change.Property)
			legacy.Uint32(&entry.Change.UpdateCount)
			marshalDataStorePropertyValue944(legacy, &entry.Change.NewValue)
		case protocol.DataStoreChangeTypeRemoval:
			legacy.String(&entry.Removal.DataStoreName)
		default:
			legacy.UnknownEnumOption(entry.ChangeType, "data store change type")
		}
	})
}

func marshalDataStoreUpdate944(io *wireIO, value *protocol.DataStoreUpdate) {
	io.String(&value.DataStoreName)
	io.String(&value.Property)
	io.String(&value.Path)
	io.Uint32(&value.ControlType)
	switch value.ControlType {
	case protocol.DataStoreControlDouble:
		io.Float64(&value.DoubleValue)
	case protocol.DataStoreControlBoolean:
		io.Bool(&value.BoolValue)
	case protocol.DataStoreControlString:
		io.String(&value.StringValue)
	default:
		io.UnknownEnumOption(value.ControlType, "data store control type")
	}
	io.Uint32(&value.PropertyUpdateCount)
	io.Uint32(&value.PathUpdateCount)
}

func marshalServerBoundDataStore(io *wireIO, raw packet.Packet) {
	marshalDataStoreUpdate944(io, &raw.(*packet.ServerBoundDataStore).Update)
}

func marshalDataStorePropertyValue944(io *wireIO, value *protocol.DataStorePropertyValue) {
	io.Int32(&value.Type)
	switch value.Type {
	case protocol.DataStorePropertyTypeNone:
	case protocol.DataStorePropertyTypeBool:
		io.Bool(&value.BoolValue)
	case protocol.DataStorePropertyTypeInt64:
		io.Int64(&value.Int64Value)
	case protocol.DataStorePropertyTypeString:
		io.String(&value.StringValue)
	case protocol.DataStorePropertyTypeMap:
		protocol.FuncIOSlice(io.directional(), &value.MapValue, func(raw protocol.IO, entry *protocol.DataStoreMapEntry) {
			legacy := asWireIO(raw)
			legacy.String(&entry.Key)
			marshalDataStorePropertyValue944(legacy, &entry.Value)
		})
	default:
		io.UnknownEnumOption(value.Type, "data store property type")
	}
}

var legacyEasingNames = [...]string{
	"linear", "spring", "in_quad", "out_quad", "in_out_quad", "in_cubic", "out_cubic", "in_out_cubic",
	"in_quart", "out_quart", "in_out_quart", "in_quint", "out_quint", "in_out_quint", "in_sine", "out_sine",
	"in_out_sine", "in_expo", "out_expo", "in_out_expo", "in_circ", "out_circ", "in_out_circ", "in_bounce",
	"out_bounce", "in_out_bounce", "in_back", "out_back", "in_out_back", "in_elastic", "out_elastic",
	"in_out_elastic", "inverse_lerp",
}

func marshalClientBoundAttributeLayerSync(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ClientBoundAttributeLayerSync)
	io.Varuint32(&pk.PayloadType)
	switch pk.PayloadType {
	case protocol.AttributeLayerPayloadTypeUpdateLayers:
		protocol.FuncIOSlice(io.directional(), &pk.Layers, func(raw protocol.IO, layer *protocol.AttributeLayerData) {
			marshalAttributeLayer944(asWireIO(raw), layer)
		})
	case protocol.AttributeLayerPayloadTypeUpdateSettings:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		marshalAttributeLayerSettings944(io, &pk.Settings)
	case protocol.AttributeLayerPayloadTypeUpdateEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncIOSlice(io.directional(), &pk.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
			marshalEnvironmentAttribute944(asWireIO(raw), value)
		})
	case protocol.AttributeLayerPayloadTypeRemoveEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncSlice(io.directional(), &pk.RemoveAttributeNames, io.String)
	default:
		io.UnknownEnumOption(pk.PayloadType, "attribute layer payload type")
	}
}

func marshalAttributeLayer944(io *wireIO, layer *protocol.AttributeLayerData) {
	io.String(&layer.Name)
	io.Varint32(&layer.DimensionID)
	marshalAttributeLayerSettings944(io, &layer.Settings)
	protocol.FuncIOSlice(io.directional(), &layer.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
		marshalEnvironmentAttribute944(asWireIO(raw), value)
	})
	if io.reading {
		layer.NoiseName = protocol.Optional[string]{}
	}
}

func marshalAttributeLayerSettings944(io *wireIO, settings *protocol.AttributeLayerSettings) {
	io.Int32(&settings.Priority)
	weightType := uint32(0)
	io.Varuint32(&weightType)
	switch weightType {
	case 0:
		io.Float32(&settings.FloatWeight)
	case 1:
		var discardedStringWeight string
		io.String(&discardedStringWeight)
		if io.reading {
			settings.FloatWeight = 0
		}
	default:
		io.UnknownEnumOption(weightType, "attribute layer weight type")
	}
	io.Bool(&settings.Enabled)
	io.Bool(&settings.TransitionsPaused)
}

func marshalEnvironmentAttribute944(io *wireIO, value *protocol.EnvironmentAttributeData) {
	io.String(&value.AttributeName)
	protocol.OptionalMarshaler(io.directional(), &value.FromAttribute)
	protocol.Single(io.directional(), &value.Attribute)
	protocol.OptionalMarshaler(io.directional(), &value.ToAttribute)
	io.Uint32(&value.CurrentTransitionTicks)
	io.Uint32(&value.TotalTransitionTicks)
	easing := "linear"
	if !io.reading {
		if value.EaseType < 0 || int(value.EaseType) >= len(legacyEasingNames) {
			io.InvalidValue(value.EaseType, "attribute easing type", "unknown easing type")
			return
		}
		easing = legacyEasingNames[value.EaseType]
	}
	io.String(&easing)
	if io.reading {
		value.EaseType = -1
		for index, name := range legacyEasingNames {
			if name == easing {
				value.EaseType = int32(index)
				break
			}
		}
		if value.EaseType == -1 {
			io.InvalidValue(easing, "attribute easing type", "unknown easing type")
		}
		value.LocalTransitionTicks = 0
		value.NoiseTransition = false
	}
}
