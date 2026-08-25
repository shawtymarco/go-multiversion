package v1_26_45

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func marshalInventoryAction(io *wireIO, action *protocol.InventoryAction) {
	io.Varuint32(&action.SourceType)
	doubleOptionalFunc(io, &action.WindowID, io.Int8)
	doubleOptionalFunc(io, &action.SourceFlags, io.Varuint32)
	io.Varuint32(&action.InventorySlot)
	io.ItemInstance(&action.OldItem)
	io.ItemInstance(&action.NewItem)
}

func marshalInventoryTransactionData(io *wireIO, value protocol.InventoryTransactionData) {
	switch data := value.(type) {
	case *protocol.NormalTransactionData, *protocol.MismatchTransactionData:
	case *protocol.UseItemTransactionData:
		marshalUseItemTransactionData(io, data)
	case *protocol.UseItemOnEntityTransactionData, *protocol.ReleaseItemTransactionData:
		data.Marshal(io)
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value), "inventory transaction data type")
	}
}

func marshalUseItemTransactionData(io *wireIO, data *protocol.UseItemTransactionData) {
	protocol.IntegerFunc(&data.ActionType, io.Varint32)
	protocol.IntegerFunc(&data.TriggerType, io.Uint8)
	io.BlockPos(&data.BlockPosition)
	protocol.IntegerFunc(&data.BlockFace, io.Uint8)
	io.Varint32(&data.HotBarSlot)
	io.ItemInstance(&data.HeldItem)
	io.Vec3(&data.Position)
	io.Vec3(&data.ClickedPosition)
	io.Varuint32(&data.BlockRuntimeID)
	io.Uint8(&data.ClientPrediction)
	io.Uint8(&data.ClientCooldownState)
	if io.reading {
		data.Hand = protocol.HandSlotMainHand
	}
}

func marshalPlayerInventoryAction(io *wireIO, data *protocol.UseItemTransactionData) {
	io.Varint32(&data.LegacyRequestID)
	protocol.OptionalFunc(io, &data.LegacySetItemSlots, func(slots *[]protocol.LegacySetItemSlot) {
		protocol.Slice(io.directional(), slots)
	})
	actions := protocol.Option(data.Actions)
	doubleOptionalFunc(io, &actions, func(values *[]protocol.InventoryAction) {
		protocol.FuncIOSlice(io.directional(), values, func(raw protocol.IO, action *protocol.InventoryAction) {
			marshalInventoryAction(asWireIO(raw), action)
		})
	})
	if values, ok := actions.Value(); ok {
		data.Actions = values
	} else {
		data.Actions = nil
	}
	io.Varuint32(&data.ActionType)
	io.Varuint32(&data.TriggerType)
	io.BlockPos(&data.BlockPosition)
	io.Varint32(&data.BlockFace)
	io.Varint32(&data.HotBarSlot)
	io.ItemInstance(&data.HeldItem)
	io.Vec3(&data.Position)
	io.Vec3(&data.ClickedPosition)
	io.Varuint32(&data.BlockRuntimeID)
	io.Uint8(&data.ClientPrediction)
	io.Uint8(&data.ClientCooldownState)
	if io.reading {
		data.Hand = protocol.HandSlotMainHand
	}
}

func marshalStackResponse(io *wireIO, response *protocol.ItemStackResponse) {
	io.Uint8(&response.Status)
	io.Varint32(&response.RequestID)
	containers := protocol.Optional[[]protocol.StackResponseContainerInfo]{}
	if len(response.ContainerInfo) != 0 {
		containers = protocol.Option(response.ContainerInfo)
	}
	doubleOptionalFunc(io, &containers, func(values *[]protocol.StackResponseContainerInfo) {
		protocol.FuncIOSlice(io.directional(), values, func(raw protocol.IO, container *protocol.StackResponseContainerInfo) {
			legacy := asWireIO(raw)
			protocol.Single(legacy.directional(), &container.Container)
			protocol.FuncIOSlice(legacy.directional(), &container.SlotInfo, func(raw protocol.IO, slot *protocol.StackResponseSlotInfo) {
				marshalStackResponseSlot(asWireIO(raw), slot)
			})
		})
	})
	if values, ok := containers.Value(); ok {
		response.ContainerInfo = values
	} else if io.reading {
		response.ContainerInfo = nil
	}
}

func marshalStackResponseSlot(io *wireIO, slot *protocol.StackResponseSlotInfo) {
	io.Uint8(&slot.Slot)
	io.Uint8(&slot.HotbarSlot)
	io.Uint8(&slot.Count)
	networkID := protocol.Optional[int32]{}
	if slot.StackNetworkID > 0 {
		networkID = protocol.Option(slot.StackNetworkID)
	}
	doubleOptionalFunc(io, &networkID, io.Varint32)
	if value, ok := networkID.Value(); ok {
		slot.StackNetworkID = value
	} else if io.reading {
		slot.StackNetworkID = 0
	}
	io.String(&slot.CustomName)
	io.String(&slot.FilteredCustomName)
	io.Varint32(&slot.DurabilityCorrection)
	if slot.DurabilityCorrection < -32768 || slot.DurabilityCorrection > 32767 {
		io.InvalidValue(slot.DurabilityCorrection, "durability correction", "must fit in an int16")
	}
}

func marshalCameraPreset(io *wireIO, preset *protocol.CameraPreset) {
	io.String(&preset.Name)
	io.String(&preset.Parent)
	protocol.OptionalFunc(io, &preset.PosX, io.Float32)
	protocol.OptionalFunc(io, &preset.PosY, io.Float32)
	protocol.OptionalFunc(io, &preset.PosZ, io.Float32)
	protocol.OptionalFunc(io, &preset.RotX, io.Float32)
	protocol.OptionalFunc(io, &preset.RotY, io.Float32)
	protocol.OptionalFunc(io, &preset.RotationSpeed, io.Float32)
	protocol.OptionalFunc(io, &preset.SnapToTarget, io.Bool)
	protocol.OptionalFunc(io, &preset.HorizontalRotationLimit, io.Vec2)
	protocol.OptionalFunc(io, &preset.VerticalRotationLimit, io.Vec2)
	protocol.OptionalFunc(io, &preset.ContinueTargeting, io.Bool)
	protocol.OptionalFunc(io, &preset.TrackingRadius, io.Float32)
	protocol.OptionalFunc(io, &preset.ViewOffset, io.Vec2)
	protocol.OptionalFunc(io, &preset.EntityOffset, io.Vec3)
	protocol.OptionalFunc(io, &preset.Radius, io.Float32)
	protocol.OptionalFunc(io, &preset.MinYawLimit, io.Float32)
	protocol.OptionalFunc(io, &preset.MaxYawLimit, io.Float32)
	protocol.OptionalFunc(io, &preset.AudioListener, io.Uint8)
	protocol.OptionalFunc(io, &preset.PlayerEffects, io.Bool)
	protocol.OptionalMarshaler(io.directional(), &preset.AimAssist)
	protocol.OptionalFunc(io, &preset.ControlScheme, io.Uint8)
	if io.reading {
		preset.ApplyInheritedStartingRotation = false
		preset.StartingRotation = protocol.Optional[mgl32.Vec2]{}
	}
}

func marshalDimensionDefinition(io *wireIO, definition *protocol.DimensionDefinition) {
	io.String(&definition.Name)
	maximumY, minimumY := definition.MinimumY+definition.HeightRange, definition.MinimumY
	io.Varint32(&maximumY)
	io.Varint32(&minimumY)
	io.Varint32(&definition.Generator)
	io.Varint32(&definition.DimensionType)
	io.UUID(&definition.PackID)
	if io.reading {
		definition.MinimumY = minimumY
		definition.HeightRange = maximumY - minimumY
		definition.DefaultBiome = ""
	}
}

func marshalAttributeLayer(io *wireIO, layer *protocol.AttributeLayerData) {
	io.String(&layer.Name)
	protocol.OptionalFunc(io, &layer.NoiseName, io.String)
	io.Varint32(&layer.DimensionID)
	protocol.Single(io.directional(), &layer.Settings)
	protocol.FuncIOSlice(io.directional(), &layer.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
		marshalEnvironmentAttribute(asWireIO(raw), value)
	})
}

var easingNames = [...]string{
	"linear", "spring", "in_quad", "out_quad", "in_out_quad", "in_cubic", "out_cubic", "in_out_cubic",
	"in_quart", "out_quart", "in_out_quart", "in_quint", "out_quint", "in_out_quint", "in_sine", "out_sine",
	"in_out_sine", "in_expo", "out_expo", "in_out_expo", "in_circ", "out_circ", "in_out_circ", "in_bounce",
	"out_bounce", "in_out_bounce", "in_back", "out_back", "in_out_back", "in_elastic", "out_elastic",
	"in_out_elastic", "inverse_lerp",
}

func marshalEnvironmentAttribute(io *wireIO, value *protocol.EnvironmentAttributeData) {
	io.String(&value.AttributeName)
	protocol.OptionalMarshaler(io.directional(), &value.FromAttribute)
	protocol.Single(io.directional(), &value.Attribute)
	protocol.OptionalMarshaler(io.directional(), &value.ToAttribute)
	io.Uint32(&value.CurrentTransitionTicks)
	io.Uint32(&value.TotalTransitionTicks)
	easing := "linear"
	if !io.reading {
		if value.EaseType < 0 || int(value.EaseType) >= len(easingNames) {
			io.InvalidValue(value.EaseType, "attribute easing type", "unknown easing type")
			return
		}
		easing = easingNames[value.EaseType]
	}
	io.String(&easing)
	if io.reading {
		value.EaseType = -1
		for index, name := range easingNames {
			if name == easing {
				value.EaseType = int32(index)
				break
			}
		}
		if value.EaseType == -1 {
			io.InvalidValue(easing, "attribute easing type", "unknown easing type")
		}
	}
	io.Uint32(&value.LocalTransitionTicks)
	io.Bool(&value.NoiseTransition)
	if io.reading {
		value.NoiseAlignment = protocol.NoiseAlignment{}
	}
}

const legacyPersonaTexturesCategory = 58

func marshalMemoryCategory(io *wireIO, value *protocol.MemoryCategoryCounter) {
	category := value.Category
	if !io.reading && category >= legacyPersonaTexturesCategory {
		category++
	}
	io.Uint8(&category)
	if io.reading {
		switch {
		case category == legacyPersonaTexturesCategory:
			value.Category = protocol.MemoryCategoryUnknown
		case category > legacyPersonaTexturesCategory:
			value.Category = category - 1
		default:
			value.Category = category
		}
	}
	io.Uint64(&value.Bytes)
}

func marshalEntityDiagnostic(io *wireIO, value *protocol.EntityDiagnosticTimingInfo) {
	io.String(&value.DisplayName)
	io.String(&value.Entity)
	io.Uint64(&value.DurationNanos)
	io.Uint8(&value.PercentOfTotal)
	if io.reading {
		value.Position = mgl32.Vec3{}
		value.Dimension = ""
	}
}

func marshalTextShape(io *wireIO, shape *protocol.TextShape) {
	io.String(&shape.Text)
	io.Bool(&shape.UseRotation)
	protocol.OptionalFunc(io, &shape.BackgroundColour, io.BEARGB)
	io.Bool(&shape.DepthTest)
	io.Bool(&shape.ShowBackface)
	io.Bool(&shape.ShowBackfaceText)
	if io.reading {
		shape.LineGapHeight = protocol.Optional[float32]{}
	}
}

func marshalSubChunkEntry(io *wireIO, entry *protocol.SubChunkEntry) {
	protocol.Single(io.directional(), &entry.Offset)
	io.Uint8(&entry.Result)
	protocol.OptionalFunc(io, &entry.RawPayload, io.ByteSlice)
	io.Uint8(&entry.HeightMapType)
	protocol.OptionalFunc(io, &entry.HeightMapData, func(data *[]int8) {
		protocol.FuncSliceOfLen(io.directional(), 256, data, io.Int8)
	})
	io.Uint8(&entry.RenderHeightMapType)
	protocol.OptionalFunc(io, &entry.RenderHeightMapData, func(data *[]int8) {
		protocol.FuncSliceOfLen(io.directional(), 256, data, io.Int8)
	})
	protocol.OptionalFunc(io, &entry.BlobHash, io.Uint64)
}
