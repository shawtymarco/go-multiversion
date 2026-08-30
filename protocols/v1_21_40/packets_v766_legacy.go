package v1_21_40

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAgentAction766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.AgentAction)
	io.String(&pk.Identifier)
	io.Varint32(&pk.Action)
	io.ByteSlice(&pk.Response)
}

func marshalChangeMobProperty766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ChangeMobProperty)
	entityUniqueID := uint64(pk.EntityUniqueID)
	io.Uint64(&entityUniqueID)
	pk.EntityUniqueID = int64(entityUniqueID)
	io.String(&pk.Property)
	io.Bool(&pk.BoolValue)
	io.String(&pk.StringValue)
	io.Varint32(&pk.IntValue)
	io.Float32(&pk.FloatValue)
}

func marshalJigsawStructureData766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.JigsawStructureData)
	var data []byte
	if !io.reading && pk.StructureData != nil {
		data, _ = nbt.MarshalEncoding(pk.StructureData, nbt.NetworkLittleEndian)
	}
	io.Bytes(&data)
	if io.reading && len(data) != 0 {
		_ = nbt.UnmarshalEncoding(data, &pk.StructureData, nbt.NetworkLittleEndian)
	}
}

func marshalRemoveVolumeEntity766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.RemoveVolumeEntity)
	runtimeID := uint64(pk.EntityRuntimeID)
	io.Uint64(&runtimeID)
	pk.EntityRuntimeID = uint32(runtimeID)
	io.Varint32(&pk.Dimension)
}

func marshalCameraAimAssist766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraAimAssist)
	io.Vec2(&pk.Angle)
	io.Float32(&pk.Distance)
	io.Uint8(&pk.TargetMode)
	io.Uint8(&pk.Action)
	if io.reading {
		pk.Preset = ""
		pk.ShowDebugRender = false
	}
}

func marshalCameraPresets766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CameraPresets)
	protocol.FuncIOSlice(io.directional(), &pk.Presets, func(raw protocol.IO, preset *protocol.CameraPreset) {
		legacy := asWireIO(raw)
		legacy.String(&preset.Name)
		legacy.String(&preset.Parent)
		protocol.OptionalFunc(legacy.directional(), &preset.PosX, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.PosY, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.PosZ, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.RotX, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.RotY, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.RotationSpeed, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.SnapToTarget, legacy.Bool)
		protocol.OptionalFunc(legacy.directional(), &preset.HorizontalRotationLimit, legacy.Vec2)
		protocol.OptionalFunc(legacy.directional(), &preset.VerticalRotationLimit, legacy.Vec2)
		protocol.OptionalFunc(legacy.directional(), &preset.ContinueTargeting, legacy.Bool)
		protocol.OptionalFunc(legacy.directional(), &preset.ViewOffset, legacy.Vec2)
		protocol.OptionalFunc(legacy.directional(), &preset.EntityOffset, legacy.Vec3)
		protocol.OptionalFunc(legacy.directional(), &preset.Radius, legacy.Float32)
		protocol.OptionalFunc(legacy.directional(), &preset.AudioListener, legacy.Uint8)
		protocol.OptionalFunc(legacy.directional(), &preset.PlayerEffects, legacy.Bool)
		var alignTargetAndCameraForward protocol.Optional[bool]
		protocol.OptionalFunc(legacy.directional(), &alignTargetAndCameraForward, legacy.Bool)
		if legacy.reading {
			preset.TrackingRadius = protocol.Optional[float32]{}
			preset.AimAssist = protocol.Optional[protocol.CameraPresetAimAssist]{}
			preset.MinYawLimit = protocol.Optional[float32]{}
			preset.MaxYawLimit = protocol.Optional[float32]{}
			preset.ControlScheme = protocol.Optional[byte]{}
		}
	})
}

func marshalCorrectPlayerMovePrediction766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CorrectPlayerMovePrediction)
	io.Uint8(&pk.PredictionType)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Delta)
	if pk.PredictionType == packet.PredictionTypeVehicle {
		io.Vec2(&pk.Rotation)
		protocol.OptionalFunc(io.directional(), &pk.VehicleAngularVelocity, io.Float32)
	} else if io.reading {
		pk.Rotation = [2]float32{}
		pk.VehicleAngularVelocity = protocol.Optional[float32]{}
	}
	io.Bool(&pk.OnGround)
	io.Varuint64(&pk.Tick)
}

func marshalSetHud766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetHud)
	elements := make([]byte, len(pk.Elements))
	if !io.reading {
		for index, element := range pk.Elements {
			elements[index] = byte(element)
		}
	}
	protocol.FuncSlice(io.directional(), &elements, io.Uint8)
	visibility := byte(pk.Visibility)
	io.Uint8(&visibility)
	if io.reading {
		pk.Elements = make([]int32, len(elements))
		for index, element := range elements {
			pk.Elements[index] = int32(element)
		}
		pk.Visibility = int32(visibility)
	}
}

func marshalSetPlayerInventoryOptions766(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetPlayerInventoryOptions)
	left, right := byte(pk.LeftInventoryTab), byte(pk.RightInventoryTab)
	inventory, crafting := byte(pk.InventoryLayout), byte(pk.CraftingLayout)
	io.Uint8(&left)
	io.Uint8(&right)
	io.Bool(&pk.Filtering)
	io.Uint8(&inventory)
	io.Uint8(&crafting)
	if io.reading {
		pk.LeftInventoryTab, pk.RightInventoryTab = int32(left), int32(right)
		pk.InventoryLayout, pk.CraftingLayout = int32(inventory), int32(crafting)
	}
}
