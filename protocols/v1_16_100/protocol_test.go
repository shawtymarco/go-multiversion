package v1_16_100

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type populatedFixture struct {
	packet packet.Packet
	hex    string
}

func TestProtocolIdentityAndCapabilities(t *testing.T) {
	p := Protocol{}
	if p.ID() != 419 || p.Ver() != "1.16.100" {
		t.Fatalf("identity: got %d/%q", p.ID(), p.Ver())
	}
	if p.LegacyNetworkSettings() != packet.FlateCompression {
		t.Fatal("protocol 419 did not select legacy flate compression")
	}
	if minY, maxY := p.NetworkChunkRange(); minY != 0 || maxY != 255 {
		t.Fatalf("legacy chunk range: got %d..%d", minY, maxY)
	}
	if p.NetworkSubChunkVersion() != 8 || !p.NetworkBiomes2D() || p.ReuseBiomePalettes() {
		t.Fatal("protocol 419 chunk capabilities are inconsistent")
	}
	first, second := p.Encryption([32]byte{1}), p.Encryption([32]byte{1})
	if first == second {
		t.Fatal("Encryption returned shared state")
	}
}

func TestHistoricalPopulatedPacketOracles(t *testing.T) {
	data, err := os.ReadFile("testdata/zero_pool_v419.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Fixtures map[string]struct {
			ID  uint32 `json:"id"`
			Hex string `json:"hex"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	item := protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 1, MetadataValue: 2}, Count: 3, NBTData: map[string]any{}, CanBePlacedOn: []string{"minecraft:stone"}, CanBreak: []string{"minecraft:dirt"}}
	inputFlags := protocol.NewInputFlags(packet.InputFlagCount)
	inputFlags.Set(packet.InputFlagJumping)
	inputFlags.Set(packet.InputFlagSprinting)
	fixtures := map[string]packet.Packet{
		"resource_packs_info":    &packet.ResourcePacksInfo{TexturePackRequired: true, HasScripts: true, TexturePacks: []protocol.TexturePackInfo{{UUID: id, Version: "1.0.0", Size: 9, ContentKey: "k", SubPackName: "s", ContentIdentity: "c", HasScripts: true}}},
		"resource_pack_stack":    &packet.ResourcePackStack{TexturePackRequired: true, TexturePacks: []protocol.StackResourcePack{{UUID: id.String(), Version: "1.0.0", SubPackName: "s"}}, BaseGameVersion: Version, Experiments: []protocol.ExperimentData{{Name: "test", Enabled: true}}, ExperimentsPreviouslyToggled: true},
		"inventory_content":      &packet.InventoryContent{WindowID: 3, Content: []protocol.ItemInstance{{StackNetworkID: 4, Stack: item}}},
		"inventory_slot":         &packet.InventorySlot{WindowID: 3, Slot: 2, NewItem: protocol.ItemInstance{StackNetworkID: 4, Stack: item}},
		"mob_equipment":          &packet.MobEquipment{EntityRuntimeID: 9, NewItem: protocol.ItemInstance{Stack: item}, InventorySlot: 1, HotBarSlot: 1, WindowID: 0},
		"mob_armour":             &packet.MobArmourEquipment{EntityRuntimeID: 9, Helmet: protocol.ItemInstance{Stack: item}, Chestplate: protocol.ItemInstance{Stack: item}, Leggings: protocol.ItemInstance{Stack: item}, Boots: protocol.ItemInstance{Stack: item}},
		"add_item_actor":         &packet.AddItemActor{EntityUniqueID: -4, EntityRuntimeID: 9, Item: protocol.ItemInstance{Stack: item}, Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, EntityMetadata: protocol.EntityMetadata{0: int64(2)}, FromFishing: true},
		"inventory_transaction":  &packet.InventoryTransaction{LegacyRequestID: 2, LegacySetItemSlots: []protocol.LegacySetItemSlot{{ContainerID: 1, Slots: []byte{2}}}, Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.Option(int8(3)), InventorySlot: 2, OldItem: protocol.ItemInstance{Stack: item}, NewItem: protocol.ItemInstance{Stack: item}}}, TransactionData: &protocol.UseItemTransactionData{ActionType: protocol.UseItemActionClickBlock, BlockPosition: protocol.BlockPos{1, 64, -2}, BlockFace: 2, HotBarSlot: 1, HeldItem: protocol.ItemInstance{Stack: item}, Position: mgl32.Vec3{1, 2, 3}, ClickedPosition: mgl32.Vec3{0.25, 0.5, 0.75}, BlockRuntimeID: 9}},
		"item_stack_request":     &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: 8, Actions: []protocol.StackRequestAction{&protocol.SwapStackRequestAction{Source: protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: 1}, Slot: 2, StackNetworkID: 3}, Destination: protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: 4}, Slot: 5, StackNetworkID: 6}}, &protocol.AutoCraftRecipeStackRequestAction{RecipeNetworkID: 7}, &protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: 17}}}}},
		"item_stack_response":    &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 8, ContainerInfo: []protocol.StackResponseContainerInfo{{Container: protocol.FullContainerName{ContainerID: 29}, SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 1, HotbarSlot: 1, Count: 2, StackNetworkID: 4}}}}}}},
		"creative_content":       &packet.CreativeContent{Items: []protocol.CreativeItem{{CreativeItemNetworkID: 17, Item: item}}},
		"game_rules_changed":     &packet.GameRulesChanged{GameRules: []protocol.GameRule{{Name: "showcoordinates", Value: true}}},
		"player_auth_input":      &packet.PlayerAuthInput{Pitch: 1, Yaw: 2, Position: mgl32.Vec3{3, 4, 5}, MoveVector: mgl32.Vec2{0.25, -0.5}, HeadYaw: 6, InputData: inputFlags, InputMode: packet.InputModeMouse, PlayMode: packet.PlayModeNormal, Tick: 99, Delta: mgl32.Vec3{0.1, 0.2, 0.3}},
		"available_commands":     &packet.AvailableCommands{EnumValues: []string{"t"}, Enums: []protocol.CommandEnum{{Type: "testAliases", ValueIndices: []uint32{0}}}, Commands: []protocol.Command{{Name: "test", Description: "desc", Flags: 1, PermissionLevel: 1, AliasesOffset: 0, Overloads: []protocol.CommandOverload{{Parameters: []protocol.CommandParameter{{Name: "value", Type: protocol.CommandArgValid | protocol.CommandArgTypeString, Optional: true, Options: 1}}}}}}},
		"player_skin":            &packet.PlayerSkin{UUID: id, Skin: protocol.Skin{SkinID: "skin", SkinResourcePatch: []byte("{}"), SkinImageWidth: 1, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 4}, CapeImageWidth: 1, CapeImageHeight: 1, CapeData: []byte{5, 6, 7, 8}, SkinGeometry: []byte("{}"), AnimationData: []byte("{}"), PremiumSkin: true, CapeID: "cape", FullID: "full", SkinColour: color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}, ArmSize: protocol.ArmSizeSlim, PersonaPieces: []protocol.PersonaPiece{{PieceID: "piece", PieceType: protocol.PieceTypeBody, PackID: id, Default: true, ProductID: "product"}}, PieceTintColours: []protocol.PersonaPieceTintColour{{PieceType: "persona_body", Colours: [4]color.RGBA{{R: 0x11, G: 0x22, B: 0x33, A: 0xff}, {}, {}, {}}}}, Trusted: true}, NewSkinName: "new", OldSkinName: "old"},
		"add_player":             &packet.AddPlayer{UUID: id, Username: "Steve", EntityRuntimeID: 10, PlatformChatID: "pc", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, HeldItem: protocol.ItemInstance{Stack: item}, EntityMetadata: protocol.EntityMetadata{0: int64(2)}, AbilityData: protocol.AbilityData{EntityUniqueID: -5, CommandPermissions: 1, PlayerPermissions: packet.PermissionLevelMember, Layers: []protocol.AbilityLayer{{Type: protocol.AbilityLayerTypeBase, Values: protocol.AbilityMayFly | protocol.AbilityBuild | protocol.AbilityMine | protocol.AbilityOpenContainers}}}, DeviceID: "device", BuildPlatform: 7},
		"add_actor":              &packet.AddActor{EntityUniqueID: -4, EntityRuntimeID: 9, EntityType: "minecraft:pig", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, BodyYaw: 10, Attributes: []protocol.AttributeValue{{Name: "minecraft:health", Min: 0, Max: 20, Value: 15}}, EntityMetadata: protocol.EntityMetadata{0: int64(2), 4: "Pig"}, EntityLinks: []protocol.EntityLink{{RiddenEntityUniqueID: 1, RiderEntityUniqueID: 2, Type: 1, Immediate: true}}},
		"level_chunk":            &packet.LevelChunk{Position: protocol.ChunkPos{-2, 5}, SubChunkCount: 2, CacheEnabled: true, BlobHashes: []uint64{11, 12, 13}, RawPayload: []byte{1, 2, 3}},
		"level_sound_undefined":  &packet.LevelSoundEvent{SoundType: "undefined"},
		"set_title":              &packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: "hello", FadeInDuration: 1, RemainDuration: 2, FadeOutDuration: 3},
		"structure_block_update": &packet.StructureBlockUpdate{Position: protocol.BlockPos{1, 64, -2}, StructureName: "house", DataField: "data", IncludePlayers: true, ShowBoundingBox: true, StructureBlockType: packet.StructureBlockSave, Settings: protocol.StructureSettings{PaletteName: "default", Size: protocol.BlockPos{2, 3, 4}, Offset: protocol.BlockPos{1, 2, 3}, LastEditingPlayerUniqueID: -5, Rotation: 1, Mirror: 2, Integrity: 1, Seed: 9, Pivot: mgl32.Vec3{1, 2, 3}}, RedstoneSaveMode: packet.StructureRedstoneSaveModeDisk, ShouldTrigger: true},
		"animate_entity":         &packet.AnimateEntity{Animation: "animation.test", NextState: "next", StopCondition: "q.any", Controller: "controller.test", BlendOutTime: 0.5, EntityRuntimeIDs: []uint64{9, 10}},
		"camera_shake":           &packet.CameraShake{Intensity: 1.5, Duration: 2.5, Type: packet.CameraShakeTypeRotational},
		"education_settings":     &packet.EducationSettings{CodeBuilderDefaultURI: "https://example.invalid", CodeBuilderTitle: "Code", CanResizeCodeBuilder: true, OverrideURI: protocol.Option("https://override.invalid"), HasQuiz: true},
		"event":                  &packet.Event{EntityRuntimeID: 9, Event: &protocol.FishBucketedEvent{}, UsePlayerID: true},
		"actor_pick":             &packet.ActorPickRequest{EntityUniqueID: -5, HotBarSlot: 2},
		"hurt_armour":            &packet.HurtArmour{Cause: 3, Damage: 4},
		"npc_request":            &packet.NPCRequest{EntityRuntimeID: 9, RequestType: 1, CommandString: "say hi", ActionType: 2},
		"photo_transfer":         &packet.PhotoTransfer{PhotoName: "photo.png", PhotoData: []byte{1, 2}, BookID: "book"},
	}
	items, err := mapping.NewItemMapper([]protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}}, map[string]mapping.TargetItem{"minecraft:stone": {RuntimeID: 1}})
	if err != nil {
		t.Fatal(err)
	}
	p := New().(*Protocol)
	p.runtime = &runtimeData{items: items}
	startGame := &packet.StartGame{EntityUniqueID: -5, EntityRuntimeID: 10, PlayerGameMode: 1, PlayerPosition: mgl32.Vec3{1, 2, 3}, Pitch: 4, Yaw: 5, WorldSeed: 6, SpawnBiomeType: packet.SpawnBiomeTypeDefault, Generator: 1, WorldGameMode: 1, Difficulty: 2, WorldSpawn: protocol.BlockPos{1, 64, -2}, AchievementsDisabled: true, DayCycleLockTime: 7, MultiPlayerGame: true, LANBroadcastEnabled: true, CommandsEnabled: true, TexturePackRequired: true, GameRules: []protocol.GameRule{{Name: "doDaylightCycle", Value: true}}, Experiments: []protocol.ExperimentData{{Name: "test", Enabled: true}}, PlayerPermissions: packet.PermissionLevelMember, ServerChunkTickRadius: 4, NewNether: true, BaseGameVersion: Version, LevelID: "level", WorldName: "world", Time: 8, EnchantmentSeed: 9, Blocks: []protocol.BlockEntry{{Name: "custom:block", Properties: map[string]any{}}}, MultiPlayerCorrelationID: "corr", ServerAuthoritativeInventory: true}
	fixtures["start_game"] = startGame
	crafting := &packet.CraftingData{ShapelessRecipes: []protocol.ShapelessRecipe{{RecipeID: "test", Input: []protocol.ItemDescriptorCount{{Descriptor: &protocol.DefaultItemDescriptor{Name: "minecraft:stone", MetadataValue: 2}, Count: 3}}, Output: []protocol.ItemStack{item}, UUID: id, Block: "crafting_table", Priority: 4, RecipeNetworkID: 5}}, PotionRecipes: []protocol.PotionRecipe{{InputPotionID: 1, InputPotionMetadata: 2, ReagentItemID: 3, ReagentItemMetadata: 4, OutputPotionID: 5, OutputPotionMetadata: 6}}, ClearRecipes: true}
	fixtures["crafting_data"] = crafting
	for name, pk := range fixtures {
		expected, ok := oracle.Fixtures[name]
		if !ok {
			t.Fatalf("missing historical fixture %q", name)
		}
		if expected.ID != pk.ID() {
			t.Fatalf("fixture %q packet ID: got %d, want %d", name, pk.ID(), expected.ID)
		}
		want, err := hex.DecodeString(expected.Hex)
		if err != nil {
			t.Fatal(err)
		}
		got := marshalProtocol419Packet(t, p, pk)
		if !bytes.Equal(got, want) {
			t.Errorf("%s bytes differ:\n got %x\nwant %x", name, got, want)
			continue
		}
		constructor, ok := p.Packets(false)[expected.ID]
		if !ok {
			constructor, ok = p.Packets(true)[expected.ID]
		}
		if !ok {
			t.Errorf("%s packet %d is absent from both pools", name, expected.ID)
			continue
		}
		decoded := constructor()
		decoded.Marshal(p.NewReader(bytes.NewReader(want), -1, true))
		latest := p.ConvertToLatest(decoded, nil)
		if len(latest) == 0 {
			t.Errorf("%s decode conversion returned no packets", name)
			continue
		}
		roundTrip := marshalProtocol419Packet(t, p, latest[0])
		if !bytes.Equal(roundTrip, want) {
			t.Errorf("%s historical round trip differs:\n got %x\nwant %x", name, roundTrip, want)
		}
	}
}

func TestUnsupportedNativeSoundIsDropped(t *testing.T) {
	p := Protocol{runtime: &runtimeData{}}
	if got := p.convertGameplayFromLatest(&packet.LevelSoundEvent{SoundType: packet.SoundEventRecordOtherside}, nil); len(got) != 0 {
		t.Fatalf("unsupported sound converted to %#v", got)
	}
}

func TestCraftingDataUsesTargetBuiltInCatalogue(t *testing.T) {
	items, err := mapping.NewItemMapper(
		[]protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}},
		map[string]mapping.TargetItem{"minecraft:stone": {RuntimeID: 1}},
	)
	if err != nil {
		t.Fatal(err)
	}
	p := Protocol{runtime: &runtimeData{items: items}}
	input := &packet.CraftingData{
		ShapelessRecipes: []protocol.ShapelessRecipe{{RecipeID: "current", RecipeNetworkID: 1}},
		ClearRecipes:     true,
	}
	if converted := p.convertGameplayFromLatest(input, nil); len(converted) != 0 {
		t.Fatalf("CraftingData conversion count: got %d, want 0", len(converted))
	}
	if len(input.ShapelessRecipes) != 1 || input.ShapelessRecipes[0].RecipeID != "current" || !input.ClearRecipes {
		t.Fatalf("CraftingData conversion mutated input: %#v", input)
	}
}

func TestPersonaPlayerSkinOmitsIncompatibleClassicCape(t *testing.T) {
	p := Protocol{runtime: &runtimeData{}}
	playerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	input := &packet.PlayerSkin{UUID: playerID, Skin: protocol.Skin{
		SkinID: "persona", PersonaSkin: true,
		CapeImageWidth: 64, CapeImageHeight: 32,
		CapeData: []byte{1, 2, 3, 4}, CapeID: "server-cape",
	}}
	converted := p.convertGameplayFromLatest(input, nil)
	if len(converted) != 1 {
		t.Fatalf("PlayerSkin conversion count: got %d, want 1", len(converted))
	}
	target := converted[0].(*packet.PlayerSkin)
	if target == input {
		t.Fatal("PlayerSkin conversion reused mutable input")
	}
	if target.Skin.CapeImageWidth != 0 || target.Skin.CapeImageHeight != 0 || len(target.Skin.CapeData) != 0 || target.Skin.CapeID != "" {
		t.Fatalf("target persona cape was not omitted: %#v", target.Skin)
	}
	if !target.Skin.PersonaSkin || target.Skin.SkinID != "persona" {
		t.Fatalf("target persona skin was not preserved: %#v", target.Skin)
	}
	if input.Skin.CapeImageWidth != 64 || input.Skin.CapeImageHeight != 32 || len(input.Skin.CapeData) != 4 || input.Skin.CapeID != "server-cape" {
		t.Fatalf("PlayerSkin conversion mutated input: %#v", input.Skin)
	}
	if self := targetPlayerSkin(input, strings.ToUpper(playerID.String())); len(self) != 0 {
		t.Fatalf("persona self update was not omitted: %#v", self)
	}
	if input.Skin.CapeImageWidth != 64 || len(input.Skin.CapeData) != 4 {
		t.Fatalf("self-update conversion mutated input: %#v", input.Skin)
	}

	classic := &packet.PlayerSkin{Skin: protocol.Skin{SkinID: "classic", CapeImageWidth: 64, CapeImageHeight: 32, CapeData: []byte{1}, CapeID: "classic-cape"}}
	classicConverted := p.convertGameplayFromLatest(classic, nil)
	if len(classicConverted) != 1 || classicConverted[0] != classic {
		t.Fatalf("classic cape was modified: %#v", classicConverted)
	}
}

func TestPlayerListPresentsRenderedPersonaAsClassic(t *testing.T) {
	persona := &packet.PlayerList{Entries: []protocol.PlayerListEntry{{
		Username: "persona",
		Skin:     protocol.Skin{SkinID: "skin", PersonaSkin: true, PersonaCapeOnClassicSkin: true},
	}}}
	target := targetPlayerList(persona)
	if target == persona || len(target.Entries) != 1 {
		t.Fatalf("persona PlayerList was not cloned: %#v", target)
	}
	if target.Entries[0].Skin.PersonaSkin || target.Entries[0].Skin.PersonaCapeOnClassicSkin || target.Entries[0].Skin.SkinID != "skin" {
		t.Fatalf("target persona entry was not presented as classic: %#v", target.Entries[0].Skin)
	}
	if !persona.Entries[0].Skin.PersonaSkin || !persona.Entries[0].Skin.PersonaCapeOnClassicSkin {
		t.Fatalf("PlayerList conversion mutated input: %#v", persona.Entries[0].Skin)
	}

	classic := &packet.PlayerList{Entries: []protocol.PlayerListEntry{{Username: "classic", Skin: protocol.Skin{SkinID: "classic"}}}}
	if targetPlayerList(classic) != classic {
		t.Fatal("classic PlayerList was unnecessarily cloned")
	}
}

func TestGameRulesChangedUsesHistoricalRuleLayout(t *testing.T) {
	p := Protocol{runtime: &runtimeData{}}
	input := &packet.GameRulesChanged{GameRules: []protocol.GameRule{
		{Name: "showcoordinates", CanBeModifiedByPlayer: true, Value: true},
		{Name: "locatorBar", CanBeModifiedByPlayer: true, Value: true},
	}}
	converted := p.convertGameplayFromLatest(input, nil)
	if len(converted) != 1 {
		t.Fatalf("GameRulesChanged conversion count: got %d, want 1", len(converted))
	}
	target := converted[0].(*packet.GameRulesChanged)
	if len(target.GameRules) != 1 || target.GameRules[0].Name != "showcoordinates" {
		t.Fatalf("target game rules: %#v", target.GameRules)
	}
	if len(input.GameRules) != 2 || !input.GameRules[0].CanBeModifiedByPlayer || input.GameRules[1].Name != "locatorBar" {
		t.Fatalf("GameRulesChanged conversion mutated input: %#v", input)
	}
}

func TestStartGameWritesFullTargetBlockPalette(t *testing.T) {
	runtime, err := newRuntimeData(protocol419IntegrationBlockRegistry(t), []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}})
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	entries := []protocol.BlockEntry{{Name: "custom:ignored", Properties: map[string]any{}}}
	marshalStartGameBlocks(newWireIO(protocol.NewWriter(&encoded, -1), false, runtime), &entries)
	if got, want := len(entries), 6611; got != want {
		t.Fatalf("StartGame block palette count: got %d, want %d", got, want)
	}
	if entries[134].Name != "minecraft:air" {
		t.Fatalf("StartGame air entry: got %q at RID 134", entries[134].Name)
	}
	if entries[0].Name == "custom:ignored" {
		t.Fatal("native custom block leaked into protocol-419 palette")
	}
}

func marshalProtocol419Packet(t *testing.T, p *Protocol, pk packet.Packet) []byte {
	t.Helper()
	marshal, ok := packetMarshals[pk.ID()]
	if !ok {
		t.Fatalf("packet %d has no protocol-419 Marshal", pk.ID())
	}
	var buffer bytes.Buffer
	translated(pk, marshal).Marshal(p.NewWriter(&buffer, -1))
	return buffer.Bytes()
}

func TestEntityMetadataKeyMapping(t *testing.T) {
	for targetKey := uint32(0); targetKey <= 109; targetKey++ {
		target := protocol.EntityMetadata{targetKey: int32(targetKey)}
		native := upgradeEntityMetadataKeys(target)
		roundTrip := downgradeEntityMetadataKeys(native)
		if got, ok := roundTrip[targetKey]; !ok || got != int32(targetKey) {
			t.Fatalf("target metadata key %d round trip: %#v", targetKey, roundTrip)
		}
	}
	currentOnly := protocol.EntityMetadata{
		protocol.EntityDataKeySeatRotationOffsetDegrees: int32(1),
		protocol.EntityDataKeyFallDamageMultiplier:      float32(1),
	}
	if got := downgradeEntityMetadataKeys(currentOnly); len(got) != 0 {
		t.Fatalf("current-only metadata leaked into protocol 419: %#v", got)
	}
}

func TestEntityFlagMappingBoundaries(t *testing.T) {
	var first, second uint64
	for _, flag := range []int{
		protocol.EntityDataFlagPowerJump,
		protocol.EntityDataFlagDash,
		protocol.EntityDataFlagLingering,
		protocol.EntityDataFlagCelebratingSpecial,
		protocol.EntityDataFlagOutOfControl,
	} {
		if flag < 64 {
			first |= uint64(1) << uint(flag)
		} else {
			second |= uint64(1) << uint(flag-64)
		}
	}
	targetFirst, targetSecond := removeEntityFlagRange(first, second, protocol.EntityDataFlagDash, 95)
	if targetFirst&(uint64(1)<<45) == 0 || targetFirst&(uint64(1)<<46) == 0 || targetSecond&(uint64(1)<<30) == 0 {
		t.Fatalf("expected target flags missing: %x/%x", targetFirst, targetSecond)
	}
	if targetSecond&(uint64(1)<<31) != 0 {
		t.Fatalf("current-only flag leaked into protocol 419: %x", targetSecond)
	}
	nativeFirst, nativeSecond := insertEntityFlagRange(targetFirst, targetSecond, protocol.EntityDataFlagDash, 94)
	if nativeFirst&(uint64(1)<<protocol.EntityDataFlagDash) != 0 || nativeFirst&(uint64(1)<<protocol.EntityDataFlagLingering) == 0 || nativeSecond&(uint64(1)<<uint(protocol.EntityDataFlagCelebratingSpecial-64)) == 0 {
		t.Fatalf("native flag expansion differs: %x/%x", nativeFirst, nativeSecond)
	}
}

func TestHistoricalZeroValuePacketPool(t *testing.T) {
	data, err := os.ReadFile("testdata/zero_pool_v419.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Server map[uint32]string `json:"server"`
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	p := New().(*Protocol)
	server, client := p.Packets(false), p.Packets(true)
	union := make(map[uint32]struct{}, len(server)+len(client))
	for id := range server {
		union[id] = struct{}{}
	}
	for id := range client {
		union[id] = struct{}{}
	}
	for id, encoded := range oracle.Server {
		if id == packet.IDLogin || id == packet.IDStartGame || id == packet.IDBiomeDefinitionList || id == packet.IDResourcePackClientResponse {
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
				t.Errorf("packet %d (%T) zero-value bytes: got %x, want %x", id, constructor(), got, want)
			}
		}()
	}
	for id := range oracle.Server {
		if _, ok := union[id]; !ok {
			t.Errorf("historical packet %d is absent from target pool union", id)
		}
	}
	for id := range union {
		if _, ok := oracle.Server[id]; !ok {
			t.Errorf("target pool union contains non-historical packet %d", id)
		}
	}
	for id := range server {
		if id > maxLegacyPacketID {
			t.Errorf("server pool contains post-419 packet %d", id)
		}
	}
	for id := range client {
		if id > maxLegacyPacketID {
			t.Errorf("client pool contains post-419 packet %d", id)
		}
	}
}
