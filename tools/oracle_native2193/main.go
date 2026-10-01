// This independent oracle imports only the frozen outgoing native dependency.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"os"
)

type fixture struct {
	Name     string
	Listener bool
	ID       uint32
	Hex      string
}

func main() {
	uid := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	skin := protocol.Skin{SkinID: "test", SkinImageWidth: 1, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 255}, Trusted: true, ProfileHash: "profile"}
	sound := protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateFade, Duration: 2, TargetVolume: 0.5}
	tests := []struct {
		name     string
		listener bool
		pk       packet.Packet
	}{
		{"start-game", false, &packet.StartGame{EntityUniqueID: 7, EntityRuntimeID: 8, WorldName: "frozen", GameVersion: "1.26.50", BaseGameVersion: "1.26.50", PropertyData: map[string]any{"test": int32(1)}}},
		{"add-actor", false, &packet.AddActor{EntityUniqueID: 7, EntityRuntimeID: 8, EntityType: "minecraft:pig"}},
		{"add-player", false, &packet.AddPlayer{UUID: uid, Username: "Frozen", EntityRuntimeID: 8, DeviceID: "device"}},
		{"animate", true, &packet.Animate{ActionType: packet.AnimateActionSwingArm, EntityRuntimeID: 8, SwingSource: packet.AnimateSwingSourceAttack}},
		{"level-chunk", false, &packet.LevelChunk{Position: protocol.ChunkPos{-2, 3}, Dimension: -1, SubChunkCount: 1, RawPayload: []byte{1, 2, 3}}},
		{"player-list", false, &packet.PlayerList{Entries: []protocol.PlayerListEntry{{UUID: uid, Username: "Frozen", XUID: "1234", Skin: skin}}}},
		{"player-skin", true, &packet.PlayerSkin{UUID: uid, Skin: skin, NewSkinName: "new"}},
		{"dimension", false, &packet.DimensionData{Definitions: []protocol.DimensionDefinition{{Name: "frozen", MinimumY: -64, HeightRange: 383, DefaultBiome: "minecraft:plains"}}}},
		{"education", false, &packet.EducationSettings{CodeBuilderTitle: "agent", CanModifyBlocks: protocol.Option(true), OverrideURI: protocol.Option("test"), HasQuiz: true}},
		{"sound", false, &packet.ClientboundUpdateSoundData{ServerSoundHandle: 7, Stop: sound, SetVolume: sound, SetPitch: sound, Fade: sound, SeekTo: sound, Pause: sound, Resume: sound}},
		{"diagnostics", true, &packet.ServerBoundDiagnostics{AverageFramesPerSecond: 60, SystemCategories: []protocol.SystemCategory{{CategoryName: "test", SystemIndex: 7}}, EntityDiagnostics: []protocol.EntityDiagnosticTimingInfo{{DisplayName: "pig", Entity: "minecraft:pig", DurationNanos: 3, PercentOfTotal: 4, Position: mgl32.Vec3{1, 2, 3}, Dimension: "overworld"}}}},
		{"attribute-constant", false, &packet.ClientBoundAttributeLayerSync{PayloadType: protocol.AttributeLayerPayloadTypeUpdateEnvironment, LayerName: "constant", EnvironmentAttributes: []protocol.EnvironmentAttributeData{{AttributeName: "fog", Attribute: protocol.AttributeData{Type: protocol.AttributeDataTypeBool, BoolValue: true}}}}},
		{"attribute-transition", false, &packet.ClientBoundAttributeLayerSync{PayloadType: protocol.AttributeLayerPayloadTypeUpdateEnvironment, LayerName: "transition", EnvironmentAttributes: []protocol.EnvironmentAttributeData{{AttributeName: "fog", Attribute: protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 1}, FromAttribute: protocol.Option(protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 1}), ToAttribute: protocol.Option(protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 2}), CurrentTransitionTicks: 3, TotalTransitionTicks: 40, EaseType: 1}}}},
		{"attribute-noise", false, &packet.ClientBoundAttributeLayerSync{PayloadType: protocol.AttributeLayerPayloadTypeUpdateLayers, Layers: []protocol.AttributeLayerData{{Name: "noise", NoiseName: protocol.Option("noise"), EnvironmentAttributes: []protocol.EnvironmentAttributeData{{AttributeName: "fog", Attribute: protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 1}, FromAttribute: protocol.Option(protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 1}), ToAttribute: protocol.Option(protocol.AttributeData{Type: protocol.AttributeDataTypeFloat, FloatValue: 2}), NoiseTransition: true, CurrentTransitionTicks: 3, TotalTransitionTicks: 40, LocalTransitionTicks: 5, EaseType: 1}}}}}},
		{"inventory-entity", true, &packet.InventoryTransaction{TransactionData: &protocol.UseItemOnEntityTransactionData{TargetEntityRuntimeID: 7, ActionType: 1, HotBarSlot: 2, Position: mgl32.Vec3{1, 2, 3}}}},
		{"inventory-release", true, &packet.InventoryTransaction{TransactionData: &protocol.ReleaseItemTransactionData{ActionType: 1, HotBarSlot: 2, HeadPosition: mgl32.Vec3{1, 2, 3}}}},
		{"stack-deprecated", true, &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: -1, Actions: []protocol.StackRequestAction{&protocol.CraftNonImplementedStackRequestAction{}, &protocol.CraftResultsDeprecatedStackRequestAction{TimesCrafted: 2}}}}}},
	}
	var result []fixture
	for _, test := range tests {
		var buf bytes.Buffer
		test.pk.Marshal(protocol.NewWriter(&buf, 0))
		payload := bytes.Clone(buf.Bytes())
		pool := packet.NewServerPool()
		if test.listener {
			pool = packet.NewClientPool()
		}
		decoded := pool[test.pk.ID()]()
		decoded.Marshal(protocol.NewReader(&buf, 0, true))
		if buf.Len() != 0 {
			panic("unread oracle bytes: " + test.name)
		}
		result = append(result, fixture{test.name, test.listener, test.pk.ID(), hex.EncodeToString(payload)})
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		panic(err)
	}
}
