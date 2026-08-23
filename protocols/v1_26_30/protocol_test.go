package v1_26_30

import (
	"bytes"
	"encoding/hex"
	"image/color"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetFixture struct {
	name string
	new  func() packet.Packet
}

func wireFixtures() []packetFixture {
	return []packetFixture{
		{name: "add_player", new: func() packet.Packet { return &packet.AddPlayer{} }},
		{name: "anvil_damage", new: func() packet.Packet { return &packet.AnvilDamage{} }},
		{name: "available_commands", new: func() packet.Packet { return &packet.AvailableCommands{} }},
		{name: "client_bound_map_item_data", new: func() packet.Packet { return &packet.ClientBoundMapItemData{} }},
		{name: "clientbound_update_sound_data", new: func() packet.Packet {
			return &packet.ClientboundUpdateSoundData{Stop: protocol.Option(protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateStop})}
		}},
		{name: "client_cheat_ability", new: func() packet.Packet { return &packet.ClientCheatAbility{} }},
		{name: "crafting_data", new: func() packet.Packet { return &packet.CraftingData{} }},
		{name: "creative_content", new: func() packet.Packet { return &packet.CreativeContent{} }},
		{name: "hurt_armour", new: func() packet.Packet { return &packet.HurtArmour{} }},
		{name: "inventory_content", new: func() packet.Packet { return &packet.InventoryContent{} }},
		{name: "inventory_slot", new: func() packet.Packet { return &packet.InventorySlot{} }},
		{name: "inventory_transaction", new: func() packet.Packet {
			return &packet.InventoryTransaction{TransactionData: &protocol.NormalTransactionData{}}
		}},
		{name: "item_stack_request", new: func() packet.Packet { return &packet.ItemStackRequest{} }},
		{name: "item_stack_response", new: func() packet.Packet { return &packet.ItemStackResponse{} }},
		{name: "level_chunk", new: func() packet.Packet { return &packet.LevelChunk{} }},
		{name: "mob_armour_equipment", new: func() packet.Packet { return &packet.MobArmourEquipment{} }},
		{name: "mob_equipment", new: func() packet.Packet { return &packet.MobEquipment{} }},
		{name: "move_actor_delta", new: func() packet.Packet { return &packet.MoveActorDelta{} }},
		{name: "move_player", new: func() packet.Packet { return &packet.MovePlayer{} }},
		{name: "play_sound", new: func() packet.Packet { return &packet.PlaySound{} }},
		{name: "player_auth_input", new: func() packet.Packet {
			return &packet.PlayerAuthInput{InputData: protocol.NewInputFlags(65)}
		}},
		{name: "player_list", new: func() packet.Packet { return &packet.PlayerList{} }},
		{name: "player_location", new: func() packet.Packet { return &packet.PlayerLocation{Type: packet.PlayerLocationTypeHide} }},
		{name: "player_skin", new: func() packet.Packet { return &packet.PlayerSkin{} }},
		{name: "player_update_entity_overrides", new: func() packet.Packet { return &packet.PlayerUpdateEntityOverrides{} }},
		{name: "resource_pack_client_response", new: func() packet.Packet {
			return &packet.ResourcePackClientResponse{Response: packet.PackResponseRefused}
		}},
		{name: "resource_packs_info", new: func() packet.Packet { return &packet.ResourcePacksInfo{} }},
		{name: "server_bound_diagnostics", new: func() packet.Packet { return &packet.ServerBoundDiagnostics{} }},
		{name: "set_score", new: func() packet.Packet { return &packet.SetScore{} }},
		{name: "set_scoreboard_identity", new: func() packet.Packet {
			return &packet.SetScoreboardIdentity{ActionType: packet.ScoreboardIdentityActionClear}
		}},
		{name: "start_game", new: func() packet.Packet { return &packet.StartGame{PropertyData: map[string]any{}} }},
		{name: "structure_block_update", new: func() packet.Packet { return &packet.StructureBlockUpdate{} }},
		{name: "sub_chunk", new: func() packet.Packet { return &packet.SubChunk{} }},
		{name: "sub_chunk_request", new: func() packet.Packet { return &packet.SubChunkRequest{} }},
		{name: "transfer", new: func() packet.Packet { return &packet.Transfer{} }},
		{name: "update_abilities", new: func() packet.Packet { return &packet.UpdateAbilities{} }},
		{name: "full/text", new: func() packet.Packet {
			return &packet.Text{TextType: packet.TextTypeRaw, Message: "message"}
		}},
		{name: "full/event", new: func() packet.Packet {
			return &packet.Event{EntityRuntimeID: 1, Event: &protocol.AchievementAwardedEvent{AchievementID: 2}}
		}},
		{name: "full/request_ability", new: func() packet.Packet {
			return &packet.RequestAbility{Ability: 1, Value: true}
		}},
		{name: "full/client_movement_prediction_sync", new: func() packet.Packet {
			flags := protocol.NewBitset(protocol.EntityDataFlagNotPickableFromInside)
			flags.Set(1)
			return &packet.ClientMovementPredictionSync{ActorFlags: flags, BoundingBoxScale: 1, MovementSpeed: 2, EntityUniqueID: 3, Flying: true}
		}},
		{name: "full/server_bound_pack_setting_change", new: func() packet.Packet {
			return &packet.ServerBoundPackSettingChange{PackID: uuid.MustParse("12345678-1234-1234-1234-123456789abc"), PackSetting: protocol.PackSetting{Name: "setting", Value: true}}
		}},
		{name: "full/dimension_data", new: func() packet.Packet {
			return &packet.DimensionData{Definitions: []protocol.DimensionDefinition{{Name: "dimension", Range: [2]int32{320, -64}, Generator: protocol.GeneratorOverworld, DimensionType: 1000}}}
		}},
		{name: "full/add_actor_shared_io", new: func() packet.Packet {
			return &packet.AddActor{EntityUniqueID: 1, EntityRuntimeID: 2, EntityType: "minecraft:pig", EntityMetadata: protocol.EntityMetadata{1: byte(1), 2: int32(2)}}
		}},
		{name: "full/add_item_actor_shared_io", new: func() packet.Packet {
			return &packet.AddItemActor{EntityUniqueID: 1, EntityRuntimeID: 2, Item: itemInstance(5), EntityMetadata: protocol.EntityMetadata{1: "item"}}
		}},
		{name: "full/set_actor_data_shared_io", new: func() packet.Packet {
			return &packet.SetActorData{EntityRuntimeID: 2, EntityMetadata: protocol.EntityMetadata{1: int64(3)}, Tick: 4}
		}},
		{name: "full/game_rules_changed_shared_io", new: func() packet.Packet {
			return &packet.GameRulesChanged{GameRules: []protocol.GameRule{{Name: "doDaylightCycle", Value: uint32(1)}}}
		}},
		{name: "full/add_player", new: func() packet.Packet {
			return &packet.AddPlayer{
				UUID: uuid.MustParse("11111111-2222-3333-4444-555555555555"), Username: "player", EntityRuntimeID: 99,
				Position: mgl32.Vec3{1, 64, -2}, HeldItem: itemInstance(5),
				EntityMetadata: protocol.EntityMetadata{1: byte(1), 2: int32(-3), 3: "name"},
				AbilityData:    protocol.AbilityData{EntityUniqueID: 99, Layers: []protocol.AbilityLayer{{Type: 1, FlySpeed: 0.05, WalkSpeed: 0.1}}},
			}
		}},
		{name: "full/available_commands", new: func() packet.Packet {
			return &packet.AvailableCommands{Commands: []protocol.Command{{
				Name: "speed", Description: "test", Overloads: []protocol.CommandOverload{{Parameters: []protocol.CommandParameter{{
					Name: "value", Type: protocol.CommandArgValid | protocol.CommandArgTypeFloat,
				}}}},
			}}}
		}},
		{name: "full/client_bound_map_item_data", new: func() packet.Packet {
			return &packet.ClientBoundMapItemData{
				MapID: 12, Dimension: 1, LockedMap: true, Origin: protocol.BlockPos{1, 2, 3},
				Scale: protocol.Option(byte(2)), MapsIncludedIn: protocol.Option([]int64{12, 13}),
				TrackedObjects: protocol.Option([]protocol.MapTrackedObject{{Type: protocol.MapObjectTypeEntity, EntityUniqueID: protocol.Option(int64(9))}}),
				Decorations:    protocol.Option([]protocol.MapDecoration{{Label: "x", Colour: color.RGBA{R: 1, G: 2, B: 3, A: 4}}}),
				Width:          protocol.Option(int32(1)), Height: protocol.Option(int32(1)), XOffset: protocol.Option(int32(2)),
				YOffset: protocol.Option(int32(3)), Pixels: protocol.Option([]color.RGBA{{R: 5, G: 6, B: 7, A: 8}}),
			}
		}},
		{name: "full/crafting_data", new: func() packet.Packet {
			return &packet.CraftingData{ShapelessRecipes: []protocol.ShapelessRecipe{{
				RecipeID: "recipe", Input: []protocol.ItemDescriptorCount{{Descriptor: &protocol.ItemTagItemDescriptor{Tag: "minecraft:planks"}, Count: 1}},
				Output: []protocol.ItemStack{itemStack(5)}, UUID: uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
				Block: "crafting_table", Priority: 1, UnlockRequirement: protocol.Option(protocol.RecipeUnlockRequirement{Context: protocol.RecipeUnlockContextAlwaysUnlocked}),
				RecipeNetworkID: 1,
			}}, ClearRecipes: true}
		}},
		{name: "full/creative_content", new: func() packet.Packet {
			return &packet.CreativeContent{
				Groups: []protocol.CreativeGroup{{Category: 1, Name: "group", Icon: itemStack(5)}},
				Items:  []protocol.CreativeItem{{CreativeItemNetworkID: 1, Item: itemStack(5), GroupIndex: 0}},
			}
		}},
		{name: "full/inventory_content", new: func() packet.Packet {
			return &packet.InventoryContent{WindowID: 1, Content: []protocol.ItemInstance{itemInstance(5)}, StorageItem: itemInstance(6)}
		}},
		{name: "full/inventory_transaction", new: func() packet.Packet {
			return &packet.InventoryTransaction{
				Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.Option(int8(0)), OldItem: itemInstance(5), NewItem: itemInstance(6)}},
				TransactionData: &protocol.UseItemTransactionData{ActionType: protocol.UseItemActionClickBlock, TriggerType: protocol.TriggerTypePlayerInput,
					BlockPosition: protocol.BlockPos{1, 64, 2}, BlockFace: 1, HeldItem: itemInstance(5), ClientPrediction: protocol.ClientPredictionSuccess},
			}
		}},
		{name: "full/item_stack_request", new: func() packet.Packet {
			return &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: 1, Actions: []protocol.StackRequestAction{
				&protocol.MineBlockStackRequestAction{HotbarSlot: 1, PredictedDurability: 2, StackNetworkID: 3},
			}}}}
		}},
		{name: "full/item_stack_response", new: func() packet.Packet {
			return &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 1,
				ContainerInfo: []protocol.StackResponseContainerInfo{{Container: protocol.FullContainerName{ContainerID: 1}, SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 2, HotbarSlot: 2, Count: 1, StackNetworkID: 3, CustomName: "x", FilteredCustomName: "x"}}}},
			}}}
		}},
		{name: "full/level_chunk", new: func() packet.Packet {
			return &packet.LevelChunk{Position: protocol.ChunkPos{1, 2}, Dimension: 0, SubChunkLimit: protocol.Option(int32(8)), CacheEnabled: true, BlobHashes: []uint64{9}, RawPayload: []byte{1, 2}}
		}},
		{name: "full/move_actor_delta", new: func() packet.Packet {
			return &packet.MoveActorDelta{EntityRuntimeID: 7, PositionX: protocol.Option(float32(1)), RotationY: protocol.Option(float32(90)), OnGround: true, ForceMove: true, ForceMoveLocalEntity: true}
		}},
		{name: "full/move_player", new: func() packet.Packet {
			return &packet.MovePlayer{EntityRuntimeID: 7, Position: mgl32.Vec3{1, 2, 3}, Mode: packet.MoveModeTeleport,
				TeleportData: protocol.Option(protocol.TeleportData{TeleportCause: 2, TeleportSourceEntityType: 3}), Tick: 4}
		}},
		{name: "full/player_auth_input", new: func() packet.Packet {
			flags := protocol.NewInputFlags(65)
			flags.Set(packet.InputFlagPerformBlockActions)
			flags.Set(packet.InputFlagClientPredictedVehicle)
			return &packet.PlayerAuthInput{InputData: flags, InputMode: packet.InputModeMouse, PlayMode: packet.PlayModeNormal,
				BlockActions:    protocol.Option([]protocol.PlayerBlockAction{{Action: protocol.PlayerActionStartBreak, BlockPos: protocol.BlockPos{1, 64, 2}, Face: 1}}),
				VehicleRotation: protocol.Option(mgl32.Vec2{1, 2}), ClientPredictedVehicle: protocol.Option(int64(9)),
			}
		}},
		{name: "full/player_list_add", new: func() packet.Packet {
			return &packet.PlayerList{Entries: []protocol.PlayerListEntry{{ActionType: protocol.PlayerListActionAdd,
				UUID: uuid.MustParse("99999999-8888-7777-6666-555555555555"), EntityUniqueID: 1, Username: "player", Skin: protocol.Skin{SkinID: "skin"}, PlayerColour: color.RGBA{A: 255}}}}
		}},
		{name: "full/player_skin", new: func() packet.Packet {
			return &packet.PlayerSkin{UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Skin: protocol.Skin{SkinID: "skin", PlayFabID: "pf", CapeID: "cape", FullID: "full", ArmSize: protocol.ArmSizeWide, SkinColour: color.RGBA{R: 1, G: 2, B: 3, A: 255}}, NewSkinName: "new", OldSkinName: "old"}
		}},
		{name: "full/resource_pack_client_response", new: func() packet.Packet {
			return &packet.ResourcePackClientResponse{Response: packet.PackResponseSendPacks, PacksToDownload: []string{"pack_1.0.0"}}
		}},
		{name: "full/server_bound_diagnostics", new: func() packet.Packet {
			return &packet.ServerBoundDiagnostics{AverageFramesPerSecond: 60, AverageRenderTime: 1.5}
		}},
		{name: "full/set_score_modify", new: func() packet.Packet {
			return &packet.SetScore{Entries: []protocol.ScoreboardEntry{{EntryID: 1, ObjectiveName: "obj", Score: 2, IdentityType: protocol.ScoreboardIdentityFakePlayer, DisplayName: "line"}}}
		}},
		{name: "full/set_score_remove", new: func() packet.Packet {
			return &packet.SetScore{Entries: []protocol.ScoreboardEntry{{EntryID: 1, ObjectiveName: "obj", Score: 2, IdentityType: protocol.ScoreboardIdentityRemove}}}
		}},
		{name: "full/start_game", new: func() packet.Packet {
			return &packet.StartGame{EntityUniqueID: 1, EntityRuntimeID: 2, WorldSeed: 3, PlayerPermissions: 2,
				GameRules: []protocol.GameRule{{Name: "doDaylightCycle", Value: uint32(1)}}, PropertyData: map[string]any{"x": int32(1)},
				ServerJoinInformation: protocol.Option(protocol.ServerJoinInformation{PresenceInfo: protocol.Option(protocol.PresenceInfo{RichPresenceID: protocol.Option("rich")})}),
			}
		}},
		{name: "full/sub_chunk", new: func() packet.Packet {
			height := make([]int8, 256)
			return &packet.SubChunk{CacheEnabled: false, Position: protocol.SubChunkPos{1, 2, 3}, SubChunkEntries: []protocol.SubChunkEntry{{
				Offset: protocol.SubChunkOffset{1, -1, 2}, Result: protocol.SubChunkResultSuccess, RawPayload: protocol.Option([]byte{1, 2}),
				HeightMapType: protocol.HeightMapDataHasData, HeightMapData: protocol.Option(height), RenderHeightMapType: protocol.HeightMapDataHasData, RenderHeightMapData: protocol.Option(height),
			}}}
		}},
		{name: "full/update_abilities", new: func() packet.Packet {
			return &packet.UpdateAbilities{AbilityData: protocol.AbilityData{EntityUniqueID: 1, PlayerPermissions: 2, Layers: []protocol.AbilityLayer{{Type: 1, Abilities: 2, Values: 3, FlySpeed: 0.05, WalkSpeed: 0.1}}}}
		}},
	}
}

func itemStack(networkID int32) protocol.ItemStack {
	return protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: networkID}, Count: 1}
}

func itemInstance(networkID int32) protocol.ItemInstance {
	return protocol.ItemInstance{StackNetworkID: 1, Stack: itemStack(networkID)}
}

func TestProtocolIdentityAndPools(t *testing.T) {
	var protocolVersion Protocol
	if got, want := protocolVersion.ID(), int32(1001); got != want {
		t.Fatalf("protocol ID: got %d, want %d", got, want)
	}
	if got, want := protocolVersion.Ver(), "1.26.36"; got != want {
		t.Fatalf("version: got %q, want %q", got, want)
	}
	if _, ok := protocolVersion.Packets(false)[packet.IDServerPlayerPostMovePosition]; ok {
		t.Fatal("ServerPlayerPostMovePosition is present in protocol 1001 server pool")
	}
	if _, ok := protocolVersion.Packets(true)[packet.IDUpdateBlock]; ok {
		t.Fatal("UpdateBlock is present in protocol 1001 client pool")
	}
}

func TestWireFixturesRoundTrip(t *testing.T) {
	for _, fixture := range wireFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			original := fixture.new()
			before := fixture.new()
			encoded := encodeTargetPacket(t, original)
			if !reflect.DeepEqual(original, before) {
				t.Fatalf("conversion mutated input:\ngot:  %#v\nwant: %#v", original, before)
			}

			constructor, ok := Protocol{}.Packets(false)[original.ID()]
			if !ok {
				t.Fatalf("packet ID %d missing from target server pool", original.ID())
			}
			decoded := constructor()
			buffer := bytes.NewBuffer(encoded)
			decoded.Marshal(Protocol{}.NewReader(buffer, 0, true))
			if buffer.Len() != 0 {
				t.Fatalf("%d unread bytes", buffer.Len())
			}
			latest := Protocol{}.ConvertToLatest(decoded, nil)
			if len(latest) != 1 {
				t.Fatalf("converted packet count: got %d, want 1", len(latest))
			}
			reencoded := encodeTargetPacket(t, latest[0])
			if !bytes.Equal(reencoded, encoded) {
				t.Fatalf("round-trip bytes differ:\nfirst:  %x\nsecond: %x", encoded, reencoded)
			}
		})
	}
}

func TestHistoricalGophertunnelOracle(t *testing.T) {
	oracle := os.Getenv("V1001_WIRE_ORACLE")
	if oracle == "" {
		t.Skip("V1001_WIRE_ORACLE is not set")
	}
	for _, fixture := range wireFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			pk := fixture.new()
			encoded := encodeTargetPacket(t, pk)
			command := exec.Command(oracle, "server", strconv.FormatUint(uint64(pk.ID()), 10), hex.EncodeToString(encoded))
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("historical oracle: %v\nencoded: %x\n%s", err, encoded, output)
			}
			decoded, err := hex.DecodeString(string(output))
			if err != nil {
				t.Fatalf("decode oracle output: %v", err)
			}
			if !bytes.Equal(decoded, encoded) {
				t.Fatalf("historical round-trip differs:\ninput:  %x\noutput: %x", encoded, decoded)
			}
		})
	}
}

func TestHistoricalGophertunnelZeroValuePools(t *testing.T) {
	oracle := os.Getenv("V1001_WIRE_ORACLE")
	if oracle == "" {
		t.Skip("V1001_WIRE_ORACLE is not set")
	}
	for _, direction := range []struct {
		name     string
		listener bool
		pool     packet.Pool
	}{
		{name: "server", pool: packet.NewServerPool()},
		{name: "client", listener: true, pool: packet.NewClientPool()},
	} {
		t.Run(direction.name, func(t *testing.T) {
			ids := make([]int, 0, len(direction.pool))
			for id := range direction.pool {
				ids = append(ids, int(id))
			}
			sort.Ints(ids)
			var failures, skipped []string
			for _, rawID := range ids {
				id := uint32(rawID)
				if _, exists := (Protocol{}).Packets(direction.listener)[id]; !exists {
					continue
				}
				pk := direction.pool[id]()
				encoded, ok := tryEncodeZeroValue(pk)
				if !ok {
					skipped = append(skipped, strconv.Itoa(rawID)+"="+reflect.TypeOf(pk).String())
					continue
				}
				command := exec.Command(oracle, direction.name, strconv.FormatUint(uint64(id), 10), hex.EncodeToString(encoded))
				output, err := command.CombinedOutput()
				if err != nil {
					failures = append(failures, strconv.Itoa(rawID)+": "+strings.TrimSpace(string(output)))
					continue
				}
				decoded, err := hex.DecodeString(string(output))
				if err != nil || !bytes.Equal(decoded, encoded) {
					failures = append(failures, strconv.Itoa(rawID)+": historical re-encode differs")
				}
			}
			if len(failures) != 0 {
				t.Fatalf("zero-value pool mismatches:\n%s", strings.Join(failures, "\n"))
			}
			if len(skipped) != 0 {
				t.Logf("zero-value constructors skipped after local validation panic/drop: %s", strings.Join(skipped, ", "))
			}
		})
	}
}

func tryEncodeZeroValue(pk packet.Packet) (encoded []byte, ok bool) {
	defer func() {
		if recover() != nil {
			encoded, ok = nil, false
		}
	}()
	converted := (Protocol{}).ConvertFromLatest(pk, nil)
	if len(converted) != 1 {
		return nil, false
	}
	var buffer bytes.Buffer
	converted[0].Marshal(Protocol{}.NewWriter(&buffer, 0))
	return buffer.Bytes(), true
}

func TestMixedPacketsAreSplit(t *testing.T) {
	players := &packet.PlayerList{Entries: []protocol.PlayerListEntry{
		{ActionType: protocol.PlayerListActionAdd},
		{ActionType: protocol.PlayerListActionRemove},
	}}
	if got := (Protocol{}).ConvertFromLatest(players, nil); len(got) != 2 {
		t.Fatalf("player list split count: got %d, want 2", len(got))
	}
	scores := &packet.SetScore{Entries: []protocol.ScoreboardEntry{
		{IdentityType: protocol.ScoreboardIdentityFakePlayer},
		{IdentityType: protocol.ScoreboardIdentityRemove},
	}}
	if got := (Protocol{}).ConvertFromLatest(scores, nil); len(got) != 2 {
		t.Fatalf("set score split count: got %d, want 2", len(got))
	}
}

func TestUnsupportedPacketsAreDropped(t *testing.T) {
	if got := (Protocol{}).ConvertFromLatest(&packet.ServerPlayerPostMovePosition{}, nil); len(got) != 0 {
		t.Fatalf("ServerPlayerPostMovePosition conversion count: got %d, want 0", len(got))
	}
	update := &packet.ClientboundUpdateSoundData{
		SetVolume: protocol.Option(protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateSetVolume, Volume: 0.5}),
	}
	if got := (Protocol{}).ConvertFromLatest(update, nil); len(got) != 0 {
		t.Fatalf("non-stop sound update conversion count: got %d, want 0", len(got))
	}
}

func encodeTargetPacket(t *testing.T, pk packet.Packet) []byte {
	t.Helper()
	converted := Protocol{}.ConvertFromLatest(pk, nil)
	if len(converted) != 1 {
		t.Fatalf("converted packet count: got %d, want 1", len(converted))
	}
	var buffer bytes.Buffer
	converted[0].Marshal(Protocol{}.NewWriter(&buffer, 0))
	return buffer.Bytes()
}
