package v1_26_45

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestHistoricalOraclePayloads(t *testing.T) {
	flags := protocol.NewInputFlags(packet.InputFlagCount)
	flags.Set(packet.InputFlagJumping)
	heightMap := make([]int8, 256)
	heightMap[0], heightMap[255] = -1, 1

	tests := []struct {
		name string
		pk   packet.Packet
		len  int
		hash string
	}{
		{name: "boss_event", pk: &packet.BossEvent{BossEntityUniqueID: 1, EventType: packet.BossEventShow, BossBarTitle: "boss", FilteredBossBarTitle: "filtered", HealthPercentage: 0.5, Colour: 5, Overlay: 2}, len: 23, hash: "a057b5302e81c9d0cae2863947bbd738f4e1ac176137ae2eaa477c3eccc923cc"},
		{name: "move_actor_delta", pk: &packet.MoveActorDelta{EntityRuntimeID: 7, PositionX: protocol.Option(float32(1)), RotationY: protocol.Option(float32(90)), OnGround: true, ForceMove: true}, len: 16, hash: "4aa3cc0b6a868b3055f0ee73369c99791cdb8daf9155b92f6bc55f6ace75f75e"},
		{name: "play_sound", pk: &packet.PlaySound{SoundName: "note.harp", Position: mgl32.Vec3{1, 2, 3}, Volume: 1, Pitch: 0.5, LoopCount: 2, Handle: protocol.Option(uint64(9))}, len: 31, hash: "33bd444692c0845b683397adb46015384dd4449c92e73f3307bb5be8fac44d41"},
		{name: "player_auth_input", pk: &packet.PlayerAuthInput{Pitch: 1, Yaw: 2, Position: mgl32.Vec3{3, 4, 5}, MoveVector: mgl32.Vec2{0.25, -0.5}, HeadYaw: 6, InputData: flags, InputMode: packet.InputModeMouse, PlayMode: packet.PlayModeNormal, InteractionModel: packet.InteractionModelCrosshair, Tick: 8, Delta: mgl32.Vec3{0.1, 0.2, 0.3}}, len: 97, hash: "23bf8bbc027ea084c030fb1e0e3fd844d3b7c1316c97b28a923fda821993efcb"},
		{name: "inventory_transaction", pk: &packet.InventoryTransaction{Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.Option(int8(1)), InventorySlot: 2}}, TransactionData: &protocol.NormalTransactionData{}}, len: 29, hash: "3ec4e8040f61f56213bc4c2cc421aaa4e2d40f3a8578852bd772dd362f78dadc"},
		{name: "item_stack_response", pk: &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 3, ContainerInfo: []protocol.StackResponseContainerInfo{{SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 1, HotbarSlot: 2, Count: 3, StackNetworkID: 7, CustomName: "item", DurabilityCorrection: 4}}}}}}}, len: 22, hash: "c78a157da499b33fa74ac0bae0f6e84eab4ac1a7d57adad9835c872233d18ae7"},
		{name: "dimension_data", pk: &packet.DimensionData{Definitions: []protocol.DimensionDefinition{{Name: "test", MinimumY: -64, HeightRange: 384, Generator: protocol.GeneratorOverworld, DimensionType: 1000}}}, len: 28, hash: "752e49256f086b198beb2c6eaf28ee776ee781b57d65b5a1d7f238ef6066eafc"},
		{name: "camera_presets", pk: &packet.CameraPresets{Presets: []protocol.CameraPreset{{Name: "test", Parent: "minecraft:first_person", PosX: protocol.Option(float32(2))}}}, len: 53, hash: "3eaa41998f37685859cf32b938ccfe0d3555520992c0b84ac61d74c9a5cd821c"},
		{name: "sub_chunk", pk: &packet.SubChunk{Position: protocol.SubChunkPos{1, 2, 3}, SubChunkEntries: []protocol.SubChunkEntry{{Result: protocol.SubChunkResultSuccess, HeightMapType: protocol.HeightMapDataHasData, HeightMapData: protocol.Option(heightMap)}}}, len: 281, hash: "08517333c1fe4a0b6a0902270acf746b2a1eb63255002e351119ab71116a6799"},
		{name: "primitive_shapes", pk: &packet.PrimitiveShapes{Shapes: []protocol.PrimitiveShape{{NetworkID: 1, ExtraShapeData: &protocol.TextShape{Text: "text", DepthTest: true}}}}, len: 22, hash: "dd2c2200f742c450b0c94a67c4a7ccf553fa14354e4f22ebd4392994ad2499c6"},
		{name: "attribute_layers", pk: &packet.ClientBoundAttributeLayerSync{PayloadType: protocol.AttributeLayerPayloadTypeUpdateEnvironment, LayerName: "layer", EnvironmentAttributes: []protocol.EnvironmentAttributeData{{AttributeName: "fog", Attribute: protocol.AttributeData{Type: protocol.AttributeDataTypeBool, BoolValue: true}, EaseType: 0}}}, len: 38, hash: "7f4d377df1109dc1a5bcf5e79baa570994a3d4041f899e1b41bcb6fd2dbdf437"},
		{name: "diagnostics", pk: &packet.ServerBoundDiagnostics{AverageFramesPerSecond: 60, MemoryCategoryValues: []protocol.MemoryCategoryCounter{{Category: protocol.MemoryCategoryPersonaCharacters, Bytes: 9}}, EntityDiagnostics: []protocol.EntityDiagnosticTimingInfo{{DisplayName: "zombie", Entity: "minecraft:zombie", DurationNanos: 10, PercentOfTotal: 11}}}, len: 83, hash: "91fc5118da955ec34cca7f60f247468ba64bc67a28fe880e2be9cd2dd94ca60b"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converted := Protocol{}.ConvertFromLatest(test.pk, nil)
			if len(converted) != 1 {
				t.Fatalf("converted packet count: got %d, want 1", len(converted))
			}
			var payload bytes.Buffer
			converted[0].Marshal(Protocol{}.NewWriter(&payload, 0))
			if payload.Len() != test.len {
				t.Fatalf("payload length: got %d, want %d", payload.Len(), test.len)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(payload.Bytes())); got != test.hash {
				t.Fatalf("payload SHA256: got %s, want %s", got, test.hash)
			}
		})
	}
}

func TestProtocolIdentityAndPools(t *testing.T) {
	p := Protocol{}
	if got, want := p.ID(), int32(2169); got != want {
		t.Fatalf("protocol ID: got %d, want %d", got, want)
	}
	if got, want := p.Ver(), "1.26.45"; got != want {
		t.Fatalf("version: got %q, want %q", got, want)
	}
	for _, listener := range []bool{false, true} {
		pool := p.Packets(listener)
		if _, ok := pool[packet.IDSetPlayerFurnaceOptions]; ok {
			t.Fatalf("listener=%t contains native-only SetPlayerFurnaceOptions", listener)
		}
		if _, ok := pool[packet.IDRecordStarted]; ok {
			t.Fatalf("listener=%t contains native-only RecordStarted", listener)
		}
	}
}

func TestChangedPacketRoundTrips(t *testing.T) {
	flags := protocol.NewInputFlags(packet.InputFlagCount)
	flags.Set(packet.InputFlagPerformItemInteraction)

	tests := []struct {
		name     string
		listener bool
		packet   packet.Packet
		check    func(*testing.T, packet.Packet)
	}{
		{
			name: "play sound",
			packet: &packet.PlaySound{SoundName: "note.harp", Position: mgl32.Vec3{1, 2, 3}, Volume: 1, Pitch: 0.5, LoopCount: 2,
				Handle: protocol.Option(uint64(9)), BypassListenerRangeCheck: true, PlaybackPositionSeconds: protocol.Option(float32(4))},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.PlaySound)
				if got.SoundName != "note.harp" || got.BypassListenerRangeCheck {
					t.Fatalf("unexpected decoded PlaySound: %#v", got)
				}
				if _, ok := got.PlaybackPositionSeconds.Value(); ok {
					t.Fatal("protocol-2192 playback position leaked into protocol 2169")
				}
			},
		},
		{
			name:   "move actor delta",
			packet: &packet.MoveActorDelta{EntityRuntimeID: 7, PositionX: protocol.Option(float32(1)), OnGround: true, Ticks: 20},
			check: func(t *testing.T, raw packet.Packet) {
				if got := raw.(*packet.MoveActorDelta); got.EntityRuntimeID != 7 || got.Ticks != 0 {
					t.Fatalf("unexpected decoded MoveActorDelta: %#v", got)
				}
			},
		},
		{
			name: "dimension data",
			packet: &packet.DimensionData{Definitions: []protocol.DimensionDefinition{{
				Name: "test", MinimumY: -64, HeightRange: 384, Generator: protocol.GeneratorOverworld, DimensionType: 1000, DefaultBiome: "minecraft:plains",
			}}},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.DimensionData).Definitions[0]
				if got.MinimumY != -64 || got.HeightRange != 384 || got.DefaultBiome != "" {
					t.Fatalf("unexpected decoded dimension: %#v", got)
				}
			},
		},
		{
			name: "camera presets",
			packet: &packet.CameraPresets{Presets: []protocol.CameraPreset{{
				Name: "test", Parent: "minecraft:first_person", PosX: protocol.Option(float32(2)),
				ApplyInheritedStartingRotation: true, StartingRotation: protocol.Option(mgl32.Vec2{4, 5}),
			}}},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.CameraPresets).Presets[0]
				if got.Name != "test" || got.ApplyInheritedStartingRotation {
					t.Fatalf("unexpected decoded camera preset: %#v", got)
				}
				if _, ok := got.StartingRotation.Value(); ok {
					t.Fatal("protocol-2192 starting rotation leaked into protocol 2169")
				}
			},
		},
		{
			name: "sub chunk",
			packet: &packet.SubChunk{Position: protocol.SubChunkPos{1, 2, 3}, SubChunkEntries: []protocol.SubChunkEntry{{
				Result: protocol.SubChunkResultSuccess, HeightMapType: protocol.HeightMapDataHasData,
				HeightMapData: protocol.Option(make([]int8, 256)),
			}}},
			check: func(t *testing.T, raw packet.Packet) {
				values, ok := raw.(*packet.SubChunk).SubChunkEntries[0].HeightMapData.Value()
				if !ok || len(values) != 256 {
					t.Fatalf("height map: present=%t len=%d", ok, len(values))
				}
			},
		},
		{
			name: "primitive text shape",
			packet: &packet.PrimitiveShapes{Shapes: []protocol.PrimitiveShape{{NetworkID: 1, ExtraShapeData: &protocol.TextShape{
				Text: "text", LineGapHeight: protocol.Option(float32(2)), DepthTest: true,
			}}}},
			check: func(t *testing.T, raw packet.Packet) {
				shape := raw.(*packet.PrimitiveShapes).Shapes[0].ExtraShapeData.(*protocol.TextShape)
				if shape.Text != "text" {
					t.Fatalf("unexpected text shape: %#v", shape)
				}
				if _, ok := shape.LineGapHeight.Value(); ok {
					t.Fatal("protocol-2192 line gap leaked into protocol 2169")
				}
			},
		},
		{
			name:     "inventory transaction",
			listener: true,
			packet: &packet.InventoryTransaction{Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.Option(int8(1))}},
				TransactionData: &protocol.UseItemTransactionData{ActionType: protocol.UseItemActionClickBlock, Hand: protocol.HandSlotOffHand}},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.InventoryTransaction)
				if len(got.Actions) != 1 || got.TransactionData.(*protocol.UseItemTransactionData).Hand != protocol.HandSlotMainHand {
					t.Fatalf("unexpected decoded inventory transaction: %#v", got)
				}
			},
		},
		{
			name:     "player auth input",
			listener: true,
			packet: &packet.PlayerAuthInput{InputData: flags, ItemInteractionData: protocol.Option(protocol.UseItemTransactionData{
				Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceWorld, SourceFlags: protocol.Option(uint32(1))}}, Hand: protocol.HandSlotOffHand,
			})},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.PlayerAuthInput)
				value, ok := got.ItemInteractionData.Value()
				if !ok || len(value.Actions) != 1 || value.Hand != protocol.HandSlotMainHand {
					t.Fatalf("unexpected decoded PlayerAuthInput item data: %#v", value)
				}
			},
		},
		{
			name: "item stack response",
			packet: &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 3,
				ContainerInfo: []protocol.StackResponseContainerInfo{{SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 1, Count: 2, StackNetworkID: 7}}}},
			}}},
			check: func(t *testing.T, raw packet.Packet) {
				got := raw.(*packet.ItemStackResponse).Responses[0]
				if got.RequestID != 3 || got.ContainerInfo[0].SlotInfo[0].StackNetworkID != 7 {
					t.Fatalf("unexpected decoded item stack response: %#v", got)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoded := roundTrip(t, test.packet, test.listener)
			test.check(t, decoded)
		})
	}
}

func TestDropsUnrepresentableNativePacketsWithoutMutation(t *testing.T) {
	p := Protocol{}
	if got := p.ConvertFromLatest(&packet.SetPlayerFurnaceOptions{}, nil); len(got) != 0 {
		t.Fatalf("SetPlayerFurnaceOptions conversion: got %#v, want no packets", got)
	}
	if got := p.ConvertFromLatest(&packet.RecordStarted{}, nil); len(got) != 0 {
		t.Fatalf("RecordStarted conversion: got %#v, want no packets", got)
	}
	if got := p.ConvertFromLatest(&packet.BossEvent{EventType: packet.BossEventRegisterPlayer}, nil); len(got) != 0 {
		t.Fatalf("BossEvent registration conversion: got %#v, want no packets", got)
	}

	start := &packet.StartGame{GameVersion: "1.26.50", BaseGameVersion: "1.26.50"}
	converted := p.ConvertFromLatest(start, nil)[0].(*packet.StartGame)
	if converted.GameVersion != Version || converted.BaseGameVersion != Version {
		t.Fatalf("downgraded versions: got %q/%q", converted.GameVersion, converted.BaseGameVersion)
	}
	if start.GameVersion != "1.26.50" || start.BaseGameVersion != "1.26.50" {
		t.Fatal("StartGame conversion mutated its input")
	}
}

func roundTrip(t *testing.T, current packet.Packet, listener bool) packet.Packet {
	t.Helper()
	p := Protocol{}
	converted := p.ConvertFromLatest(current, nil)
	if len(converted) != 1 {
		t.Fatalf("converted packet count: got %d, want 1", len(converted))
	}
	var payload bytes.Buffer
	converted[0].Marshal(p.NewWriter(&payload, 0))

	constructor, ok := p.Packets(listener)[current.ID()]
	if !ok {
		t.Fatalf("packet %d missing from listener=%t pool", current.ID(), listener)
	}
	decoded := constructor()
	reader := bytes.NewBuffer(bytes.Clone(payload.Bytes()))
	decoded.Marshal(p.NewReader(reader, 0, true))
	if reader.Len() != 0 {
		t.Fatalf("packet %d has %d unread bytes", current.ID(), reader.Len())
	}
	latest := p.ConvertToLatest(decoded, nil)
	if len(latest) != 1 {
		t.Fatalf("latest packet count: got %d, want 1", len(latest))
	}
	return latest[0]
}
