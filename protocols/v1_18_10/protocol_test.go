package v1_18_10

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"image/color"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	v486data "github.com/shawtymarco/go-multiversion/data/v486"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type oracleFixture struct {
	packet   packet.Packet
	listener bool
	hex      string
}

func TestHistoricalZeroValuePacketPools(t *testing.T) {
	data, err := os.ReadFile("testdata/zero_pool.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Packets map[uint32]string `json:"packets"`
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	p := New().(*Protocol)
	server, client := p.Packets(false), p.Packets(true)
	for id, encoded := range oracle.Packets {
		if id == packet.IDStartGame || id == packet.IDBiomeDefinitionList || id == packet.IDResourcePackClientResponse {
			continue
		}
		constructor, ok := server[id]
		if !ok {
			constructor, ok = client[id]
		}
		if !ok {
			t.Errorf("historical packet %d is missing from both target pools", id)
			continue
		}
		want, err := hex.DecodeString(encoded)
		if err != nil {
			t.Fatalf("decode packet %d fixture: %v", id, err)
		}
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("packet %d zero-value Marshal panicked: %v", id, recovered)
				}
			}()
			var buffer bytes.Buffer
			constructor().Marshal(p.NewWriter(&buffer, -1))
			if got := buffer.Bytes(); !bytes.Equal(got, want) {
				t.Errorf("packet %d zero-value bytes: got %x, want %x", id, got, want)
			}
		}()
	}
}

func TestHistoricalPacketOracles(t *testing.T) {
	inputFlags := protocol.NewInputFlags(packet.InputFlagCount)
	inputFlags.Set(0)
	identifier := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixtures := map[string]oracleFixture{
		"animate":                {packet: &packet.Animate{ActionType: 0x80, EntityRuntimeID: 19, Data: 1.25}, hex: "8002130000a03f"},
		"text":                   {packet: &packet.Text{TextType: packet.TextTypeChat, NeedsTranslation: true, SourceName: "Steve", Message: "hello", XUID: "42", PlatformChatID: "pc"}, hex: "01010553746576650568656c6c6f023432027063"},
		"container_close":        {packet: &packet.ContainerClose{WindowID: 7, ServerSide: true}, hex: "0701"},
		"player_action":          {packet: &packet.PlayerAction{EntityRuntimeID: 9, ActionType: protocol.PlayerActionStartBreak, BlockPosition: protocol.BlockPos{3, 70, -4}, ResultPosition: protocol.BlockPos{3, 70, -4}, BlockFace: 2}, listener: true, hex: "090006460704"},
		"request_chunk_radius":   {packet: &packet.RequestChunkRadius{ChunkRadius: 12, MaxChunkRadius: 12}, listener: true, hex: "18"},
		"level_chunk":            {packet: &packet.LevelChunk{Position: protocol.ChunkPos{-2, 5}, Dimension: packet.DimensionOverworld, SubChunkCount: 2, CacheEnabled: true, BlobHashes: []uint64{11, 12, 13}, RawPayload: []byte{1, 2, 3}}, hex: "030a0201030b000000000000000c000000000000000d0000000000000003010203"},
		"sub_chunk_request":      {packet: &packet.SubChunkRequest{Dimension: 0, Position: protocol.SubChunkPos{-2, 3, 4}, Offsets: []protocol.SubChunkOffset{{-1, 0, 1}, {2, -3, 4}}}, listener: true, hex: "0003060802000000ff000102fd04"},
		"sub_chunk":              {packet: &packet.SubChunk{CacheEnabled: true, Dimension: 0, Position: protocol.SubChunkPos{1, -2, 3}, SubChunkEntries: []protocol.SubChunkEntry{{Offset: protocol.SubChunkOffset{-1, 2, 3}, Result: protocol.SubChunkResultSuccess, RawPayload: protocol.Option([]byte{9, 8}), HeightMapType: protocol.HeightMapDataNone, BlobHash: protocol.Option(uint64(77))}}}, hex: "010002030601000000ff020301020908004d00000000000000"},
		"player_auth_input":      {packet: &packet.PlayerAuthInput{Pitch: 1, Yaw: 2, Position: mgl32.Vec3{3, 4, 5}, MoveVector: mgl32.Vec2{0.25, -0.5}, HeadYaw: 6, InputData: inputFlags, InputMode: packet.InputModeMouse, PlayMode: packet.PlayModeNormal, Tick: 99, Delta: mgl32.Vec3{0.1, 0.2, 0.3}}, listener: true, hex: "0000803f0000004000004040000080400000a0400000803e000000bf0000c04001010063cdcccc3dcdcc4c3e9a99993e"},
		"inventory_content":      {packet: &packet.InventoryContent{WindowID: 3, Content: []protocol.ItemInstance{{StackNetworkID: 4, Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 1, MetadataValue: 2}, BlockRuntimeID: 3, Count: 5}}}}, hex: "0301020500020108060a00000000000000000000"},
		"creative_content":       {packet: &packet.CreativeContent{Items: []protocol.CreativeItem{{CreativeItemNetworkID: 17, Item: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 1}, Count: 1}}}}, hex: "011102010000000a00000000000000000000"},
		"item_stack_request":     {packet: &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: 8, Actions: []protocol.StackRequestAction{&protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: 17, NumberOfCrafts: 1}}, FilterStrings: []string{"name"}}}}, listener: true, hex: "0110010e1101046e616d65"},
		"item_stack_response":    {packet: &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 8, ContainerInfo: []protocol.StackResponseContainerInfo{{Container: protocol.FullContainerName{ContainerID: 28}, SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 1, HotbarSlot: 1, Count: 2, StackNetworkID: 4, CustomName: "item", DurabilityCorrection: 3}}}}}}}, hex: "010010011b0101010208046974656d06"},
		"resource_packs_info":    {packet: &packet.ResourcePacksInfo{TexturePackRequired: true, TexturePacks: []protocol.TexturePackInfo{{UUID: identifier, Version: "1.0.0", Size: 9, ContentKey: "k", SubPackName: "s", ContentIdentity: "c", RTXEnabled: true}}}, hex: "010000000001002431323365343536372d653839622d313264332d613435362d34323636313431373430303005312e302e300900000000000000016b017301630001"},
		"resource_pack_stack":    {packet: &packet.ResourcePackStack{TexturePackRequired: true, TexturePacks: []protocol.StackResourcePack{{UUID: identifier.String(), Version: "1.0.0", SubPackName: "s"}}, BaseGameVersion: "1.18.12", Experiments: []protocol.ExperimentData{{Name: "test", Enabled: true}}, ExperimentsPreviouslyToggled: true}, hex: "0100012431323365343536372d653839622d313264332d613435362d34323636313431373430303005312e302e30017307312e31382e31320100000004746573740101"},
		"add_actor":              {packet: &packet.AddActor{EntityUniqueID: -4, EntityRuntimeID: 9, EntityType: "minecraft:pig", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, BodyYaw: 10, Attributes: []protocol.AttributeValue{{Name: "minecraft:health", Min: 0, Max: 20, Value: 15}}, EntityMetadata: protocol.EntityMetadata{0: int64(2), 4: "Pig"}, EntityLinks: []protocol.EntityLink{{RiddenEntityUniqueID: 1, RiderEntityUniqueID: 2, Type: 1, Immediate: true, VehicleAngularVelocity: 9}}}, hex: "07090d6d696e6563726166743a7069670000803f0000004000004040000080400000a0400000c0400000e040000000410000104101106d696e6563726166743a6865616c746800000000000070410000a04102000704040403506967010204010100"},
		"add_player":             {packet: &packet.AddPlayer{UUID: identifier, Username: "Steve", EntityRuntimeID: 10, PlatformChatID: "pc", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, EntityMetadata: protocol.EntityMetadata{0: int64(2), 4: "Steve"}, AbilityData: protocol.AbilityData{EntityUniqueID: -5, PlayerPermissions: packet.PermissionLevelMember, CommandPermissions: 1, Layers: []protocol.AbilityLayer{{Type: protocol.AbilityLayerTypeBase, Values: protocol.AbilityBuild | protocol.AbilityMine | protocol.AbilityDoorsAndSwitches | protocol.AbilityOpenContainers | protocol.AbilityAttackPlayers | protocol.AbilityAttackMobs | protocol.AbilityMayFly}}}, DeviceID: "device", BuildPlatform: 7}, hex: "d3129be867453e1200401714664256a4055374657665090a0270630000803f0000004000004040000080400000a0400000c0400000e04000000041000010410002000704040405537465766540019f010100fbffffffffffffff000664657669636507000000"},
		"update_attributes":      {packet: &packet.UpdateAttributes{EntityRuntimeID: 9, Attributes: []protocol.Attribute{{AttributeValue: protocol.AttributeValue{Name: "minecraft:health", Min: 0, Max: 20, Value: 15}, DefaultMin: -1, DefaultMax: 99, Default: 20}}, Tick: 44}, hex: "0901000000000000a041000070410000a041106d696e6563726166743a6865616c74682c"},
		"player_list_remove":     {packet: &packet.PlayerList{Entries: []protocol.PlayerListEntry{{ActionType: protocol.PlayerListActionRemove, UUID: identifier}}}, hex: "0101d3129be867453e1200401714664256a4"},
		"available_commands":     {packet: &packet.AvailableCommands{Commands: []protocol.Command{{Name: "test", Description: "desc", Flags: 1, PermissionLevel: 1, AliasesOffset: math.MaxUint32, Overloads: []protocol.CommandOverload{{Parameters: []protocol.CommandParameter{{Name: "value", Type: protocol.CommandArgValid | protocol.CommandArgTypeString, Optional: true, Options: 1}}}}}}}, hex: "0000000104746573740464657363010001ffffffff01010576616c75652000100001010000"},
		"client_bound_map":       {packet: &packet.ClientBoundMapItemData{MapID: 4, LockedMap: true, Scale: protocol.Option(byte(1)), Width: protocol.Option(int32(1)), Height: protocol.Option(int32(1)), XOffset: protocol.Option(int32(2)), YOffset: protocol.Option(int32(3)), Pixels: protocol.Option([]color.RGBA{{R: 1, G: 2, B: 3, A: 4}})}, hex: "0802000101020204060181848c20"},
		"command_block_update":   {packet: &packet.CommandBlockUpdate{Block: true, Position: protocol.BlockPos{1, 64, -2}, Mode: packet.CommandBlockRepeating, NeedsRedstone: true, Command: "say hi", LastOutput: "ok", Name: "cmd", ShouldTrackOutput: true, TickDelay: 4, ExecuteOnFirstTick: true}, listener: true, hex: "0102400301010006736179206869026f6b03636d64010400000001"},
		"command_output":         {packet: &packet.CommandOutput{CommandOrigin: protocol.CommandOrigin{Origin: protocol.CommandOriginPlayer, UUID: identifier, RequestID: "req"}, OutputType: packet.CommandOutputTypeAllOutput, SuccessCount: 1, OutputMessages: []protocol.CommandOutputMessage{{Success: true, Message: "done", Parameters: []string{"x"}}}}, hex: "00d3129be867453e1200401714664256a4037265710301010104646f6e65010178"},
		"level_sound_event":      {packet: &packet.LevelSoundEvent{SoundType: packet.SoundEventAnvilUse, Position: mgl32.Vec3{1, 2, 3}, ExtraData: 4, EntityType: "minecraft:player", BabyMob: true}, hex: "af010000803f000000400000404008106d696e6563726166743a706c617965720100"},
		"player_list_add":        {packet: &packet.PlayerList{Entries: []protocol.PlayerListEntry{{ActionType: protocol.PlayerListActionAdd, UUID: identifier, EntityUniqueID: -5, Username: "Steve", XUID: "42", PlatformChatID: "pc", BuildPlatform: 7, Skin: protocol.Skin{SkinID: "skin", PlayFabID: "pf", SkinImageWidth: 1, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 4}, CapeImageWidth: 1, CapeImageHeight: 1, CapeData: []byte{5, 6, 7, 8}, ArmSize: protocol.ArmSizeSlim, SkinColour: color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}}, Teacher: true, Host: true}}}, hex: "0001d3129be867453e1200401714664256a4090553746576650234320270630700000004736b696e02706600010000000100000004010203040000000001000000010000000405060708000000000004736c696d0723313132323333000000000000000000000000010100"},
		"structure_block_update": {packet: &packet.StructureBlockUpdate{Position: protocol.BlockPos{1, 64, -2}, StructureName: "house", DataField: "data", IncludePlayers: true, ShowBoundingBox: true, StructureBlockType: packet.StructureBlockSave, Settings: protocol.StructureSettings{PaletteName: "default", Size: protocol.BlockPos{2, 3, 4}, Offset: protocol.BlockPos{1, 2, 3}, LastEditingPlayerUniqueID: -5, Rotation: 1, Mirror: 2, AnimationMode: 1, AnimationDuration: 0.5, Integrity: 1, Seed: 9, Pivot: mgl32.Vec3{1, 2, 3}}, RedstoneSaveMode: packet.StructureRedstoneSaveModeDisk, ShouldTrigger: true}, listener: true, hex: "02400305686f75736504646174610101020764656661756c740000040308020206090102010000003f0000803f090000000000803f00000040000040400201"},
	}

	p := New().(*Protocol)
	for name, fixture := range fixtures {
		t.Run(name, func(t *testing.T) {
			want, err := hex.DecodeString(fixture.hex)
			if err != nil {
				t.Fatal(err)
			}
			got := marshalOraclePacket(t, p, fixture.packet)
			if !bytes.Equal(got, want) {
				t.Fatalf("historical bytes differ:\n got %x\nwant %x", got, want)
			}

			constructor, ok := p.Packets(fixture.listener)[fixture.packet.ID()]
			if !ok {
				t.Fatalf("packet %d missing from target pool", fixture.packet.ID())
			}
			decoded := constructor()
			decoded.Marshal(p.NewReader(bytes.NewReader(want), -1, true))
			latest := p.ConvertToLatest(decoded, nil)
			if len(latest) != 1 {
				t.Fatalf("decode conversion count: got %d, want 1", len(latest))
			}
			roundTrip := marshalOraclePacket(t, p, latest[0])
			if !bytes.Equal(roundTrip, want) {
				t.Fatalf("historical round trip differs:\n got %x\nwant %x", roundTrip, want)
			}
		})
	}
}

func marshalOraclePacket(t *testing.T, p *Protocol, pk packet.Packet) []byte {
	t.Helper()
	marshal, ok := packetMarshals[pk.ID()]
	if !ok {
		t.Fatalf("packet %d has no protocol-486 Marshal", pk.ID())
	}
	var buffer bytes.Buffer
	translated(pk, marshal).Marshal(p.NewWriter(&buffer, -1))
	return buffer.Bytes()
}

func TestEntityFlagConversionRoundTrip(t *testing.T) {
	for index := 0; index < 127; index++ {
		if index == protocol.EntityDataFlagDash {
			continue
		}
		var native [2]uint64
		native[index/64] = uint64(1) << uint(index%64)
		legacyFirst, legacySecond := removeEntityFlag(native[0], native[1], protocol.EntityDataFlagDash)
		gotFirst, gotSecond := insertEntityFlag(legacyFirst, legacySecond, protocol.EntityDataFlagDash)
		if !reflect.DeepEqual([2]uint64{gotFirst, gotSecond}, native) {
			t.Fatalf("flag %d round trip: got [%x %x], want [%x %x]", index, gotFirst, gotSecond, native[0], native[1])
		}
	}
}

func TestHistoricalStartGameOracle(t *testing.T) {
	items, err := mapping.NewItemMapper(
		[]protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}},
		map[string]mapping.TargetItem{"minecraft:stone": {RuntimeID: 1}},
	)
	if err != nil {
		t.Fatal(err)
	}
	p := New().(*Protocol)
	p.runtime = &runtimeData{items: items, biomes: v486data.BiomeDefinitions()}
	pk := &packet.StartGame{
		EntityUniqueID: -5, EntityRuntimeID: 10, PlayerGameMode: 1,
		PlayerPosition: mgl32.Vec3{1, 2, 3}, Pitch: 4, Yaw: 5, WorldSeed: 6,
		SpawnBiomeType: packet.SpawnBiomeTypeDefault, Generator: 1, WorldGameMode: 1, Difficulty: 2,
		WorldSpawn: protocol.BlockPos{1, 64, -2}, AchievementsDisabled: true, DayCycleLockTime: 7,
		MultiPlayerGame: true, LANBroadcastEnabled: true, CommandsEnabled: true, TexturePackRequired: true,
		GameRules:         []protocol.GameRule{{Name: "doDaylightCycle", Value: true}},
		Experiments:       []protocol.ExperimentData{{Name: "test", Enabled: true}},
		PlayerPermissions: packet.PermissionLevelMember, ServerChunkTickRadius: 4, NewNether: true,
		BaseGameVersion: "1.18.12", LevelID: "level", WorldName: "world",
		PlayerMovementSettings: protocol.PlayerMovementSettings{RewindHistorySize: 20, ServerAuthoritativeBlockBreaking: true},
		Time:                   8, EnchantmentSeed: 9,
		Blocks:                   []protocol.BlockEntry{{Name: "custom:block", Properties: map[string]any{}}},
		MultiPlayerCorrelationID: "corr", ServerAuthoritativeInventory: true, GameVersion: "1.18.12",
		ServerBlockStateChecksum: 11,
	}
	want, err := hex.DecodeString("090a020000803f0000004000004040000080400000a0400c00000000020204024003010e000000000000000000000000010100000101010f646f4461796c696768744379636c650001010100000004746573740100000002040000000000000000000007312e31382e3132000000000000000001000000056c6576656c05776f726c640000022801080000000000000012010c637573746f6d3a626c6f636b0a0000010f6d696e6563726166743a73746f6e6501000004636f72720107312e31382e31320b00000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if got := marshalOraclePacket(t, p, pk); !bytes.Equal(got, want) {
		t.Fatalf("StartGame bytes differ:\n got %x\nwant %x", got, want)
	}
	decoded := p.Packets(false)[packet.IDStartGame]()
	decoded.Marshal(p.NewReader(bytes.NewReader(want), -1, true))
	latest := p.ConvertToLatest(decoded, nil)
	if len(latest) != 2 {
		t.Fatalf("StartGame decode conversion count: got %d, want 2", len(latest))
	}
	registry, ok := latest[1].(*packet.ItemRegistry)
	if !ok || len(registry.Items) != 1 || registry.Items[0].Name != "minecraft:stone" {
		t.Fatalf("synthetic item registry: %#v", latest[1])
	}
	if got := marshalOraclePacket(t, p, latest[0]); !bytes.Equal(got, want) {
		t.Fatalf("StartGame round trip differs:\n got %x\nwant %x", got, want)
	}
}

func TestHistoricalBiomeDefinitionPacket(t *testing.T) {
	p := New().(*Protocol)
	p.runtime = &runtimeData{biomes: v486data.BiomeDefinitions()}
	want := v486data.BiomeDefinitions()
	if got := marshalOraclePacket(t, p, &packet.BiomeDefinitionList{}); !bytes.Equal(got, want) {
		t.Fatalf("biome definition bytes differ: got %x, want %x", got, want)
	}
}

func TestHistoricalCraftingDataOracle(t *testing.T) {
	items, err := mapping.NewItemMapper(
		[]protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}},
		map[string]mapping.TargetItem{"minecraft:stone": {RuntimeID: 1}},
	)
	if err != nil {
		t.Fatal(err)
	}
	p := New().(*Protocol)
	p.runtime = &runtimeData{items: items}
	pk := &packet.CraftingData{
		ShapelessRecipes: []protocol.ShapelessRecipe{{
			RecipeID: "test",
			Input:    []protocol.ItemDescriptorCount{{Descriptor: &protocol.DefaultItemDescriptor{Name: "minecraft:stone", MetadataValue: 2}, Count: 3}},
			Output:   []protocol.ItemStack{{ItemType: protocol.ItemType{NetworkID: 1}, Count: 1}},
			UUID:     uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Block: "crafting_table", Priority: 4, RecipeNetworkID: 5,
		}},
		PotionRecipes: []protocol.PotionRecipe{{InputPotionID: 1, InputPotionMetadata: 2, ReagentItemID: 3, ReagentItemMetadata: 4, OutputPotionID: 5, OutputPotionMetadata: 6}},
		ClearRecipes:  true,
	}
	want, err := hex.DecodeString("01000474657374010204060102010000000a00000000000000000000d3129be867453e1200401714664256a40e6372616674696e675f7461626c65080501020406080a0c000001")
	if err != nil {
		t.Fatal(err)
	}
	if got := marshalOraclePacket(t, p, pk); !bytes.Equal(got, want) {
		t.Fatalf("CraftingData bytes differ:\n got %x\nwant %x", got, want)
	}
	decoded := p.Packets(false)[packet.IDCraftingData]()
	decoded.Marshal(p.NewReader(bytes.NewReader(want), -1, true))
	latest := p.ConvertToLatest(decoded, nil)
	if len(latest) != 1 {
		t.Fatalf("CraftingData decode conversion count: got %d, want 1", len(latest))
	}
	if got := marshalOraclePacket(t, p, latest[0]); !bytes.Equal(got, want) {
		t.Fatalf("CraftingData round trip differs:\n got %x\nwant %x", got, want)
	}
}

func TestInitialWorldStateConversion(t *testing.T) {
	p := Protocol{runtime: &runtimeData{}}

	startGame := &packet.StartGame{
		BaseGameVersion: "1.26.45",
		GameVersion:     "1.26.45",
		GameRules: []protocol.GameRule{
			{Name: "naturalregeneration", Value: false},
			{Name: "locatorBar", Value: false},
		},
	}
	converted := p.convertGameplayFromLatest(startGame, nil)
	if len(converted) != 1 {
		t.Fatalf("StartGame conversion count: got %d, want 1", len(converted))
	}
	targetStartGame := converted[0].(*packet.StartGame)
	if targetStartGame.BaseGameVersion != Version || targetStartGame.GameVersion != Version {
		t.Fatalf("StartGame versions: got %q/%q, want %q", targetStartGame.BaseGameVersion, targetStartGame.GameVersion, Version)
	}
	if len(targetStartGame.GameRules) != 1 || targetStartGame.GameRules[0].Name != "naturalregeneration" {
		t.Fatalf("StartGame game rules: %#v", targetStartGame.GameRules)
	}
	if len(startGame.GameRules) != 2 || startGame.BaseGameVersion != "1.26.45" || startGame.GameVersion != "1.26.45" {
		t.Fatalf("StartGame conversion mutated input: %#v", startGame)
	}

	const (
		userPack            = "123e4567-e89b-12d3-a456-426614174000"
		legacyExemptedPack  = "0fba4063-dba1-4281-9b89-ff9390653530"
		currentExemptedPack = "d34cfa4b-2ad1-453d-a0db-668b429a3ea0"
	)
	resourceStack := &packet.ResourcePackStack{
		BaseGameVersion: "1.26.45",
		TexturePacks: []protocol.StackResourcePack{
			{UUID: userPack, Version: "1.0.0"},
			{UUID: legacyExemptedPack, Version: "1.0.0"},
			{UUID: currentExemptedPack, Version: "1.26.40"},
		},
		Experiments:                  []protocol.ExperimentData{{Name: "cameras", Enabled: true}},
		ExperimentsPreviouslyToggled: true,
	}
	converted = p.convertGameplayFromLatest(resourceStack, nil)
	if len(converted) != 1 {
		t.Fatalf("ResourcePackStack conversion count: got %d, want 1", len(converted))
	}
	targetStack := converted[0].(*packet.ResourcePackStack)
	if targetStack.BaseGameVersion != Version {
		t.Fatalf("ResourcePackStack base version: got %q, want %q", targetStack.BaseGameVersion, Version)
	}
	if len(targetStack.TexturePacks) != 2 || targetStack.TexturePacks[0].UUID != userPack || targetStack.TexturePacks[1].UUID != legacyExemptedPack {
		t.Fatalf("ResourcePackStack texture packs: %#v", targetStack.TexturePacks)
	}
	if len(targetStack.Experiments) != 0 || targetStack.ExperimentsPreviouslyToggled {
		t.Fatalf("ResourcePackStack experiments: %#v", targetStack.Experiments)
	}
	if len(resourceStack.TexturePacks) != 3 || len(resourceStack.Experiments) != 1 || !resourceStack.ExperimentsPreviouslyToggled {
		t.Fatalf("ResourcePackStack conversion mutated input: %#v", resourceStack)
	}
}

func TestPreSpawnPackets(t *testing.T) {
	p := Protocol{}
	packets := p.PreSpawnPackets()
	if len(packets) != 1 {
		t.Fatalf("pre-spawn packet count: got %d, want 1", len(packets))
	}
	if _, ok := packets[0].(*packet.BiomeDefinitionList); !ok {
		t.Fatalf("pre-spawn packet: got %T, want *packet.BiomeDefinitionList", packets[0])
	}
	if packets[0] == p.PreSpawnPackets()[0] {
		t.Fatal("PreSpawnPackets returned a shared packet instance")
	}
}

func TestBiomePaletteReuseDisabled(t *testing.T) {
	if (Protocol{}).ReuseBiomePalettes() {
		t.Fatal("protocol 486 enabled biome palette reuse")
	}
}
