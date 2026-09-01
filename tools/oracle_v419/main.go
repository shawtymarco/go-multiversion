package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type output struct {
	Server   map[uint32]string  `json:"server"`
	Client   map[uint32]string  `json:"client"`
	Fixtures map[string]fixture `json:"fixtures"`
	Skipped  map[string]string  `json:"skipped,omitempty"`
}

type fixture struct {
	ID  uint32 `json:"id"`
	Hex string `json:"hex"`
}

func main() {
	path := flag.String("out", "", "output JSON path")
	flag.Parse()
	if *path == "" && flag.NArg() == 3 {
		roundTrip(flag.Arg(0), flag.Arg(1), flag.Arg(2))
		return
	}
	if *path == "" {
		flag.Usage()
		os.Exit(2)
	}
	result := output{Server: map[uint32]string{}, Client: map[uint32]string{}, Fixtures: populatedFixtures(), Skipped: map[string]string{}}
	encodePool("server", packet.NewPool(), result.Server, result.Skipped)
	encodePool("client", packet.NewPool(), result.Client, result.Skipped)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(*path, data, 0o644); err != nil {
		panic(err)
	}
}

func populatedFixtures() map[string]fixture {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	item := protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 1, MetadataValue: 2}, Count: 3, NBTData: map[string]interface{}{}, CanBePlacedOn: []string{"minecraft:stone"}, CanBreak: []string{"minecraft:dirt"}}
	fixtures := map[string]packet.Packet{
		"resource_packs_info":    &packet.ResourcePacksInfo{TexturePackRequired: true, HasScripts: true, TexturePacks: []protocol.ResourcePackInfo{{UUID: id.String(), Version: "1.0.0", Size: 9, ContentKey: "k", SubPackName: "s", ContentIdentity: "c", HasScripts: true}}},
		"resource_pack_stack":    &packet.ResourcePackStack{TexturePackRequired: true, TexturePacks: []protocol.StackResourcePack{{UUID: id.String(), Version: "1.0.0", SubPackName: "s"}}, BaseGameVersion: "1.16.100", Experiments: []protocol.ExperimentData{{Name: "test", Enabled: true}}, ExperimentsPreviouslyToggled: true},
		"inventory_content":      &packet.InventoryContent{WindowID: 3, Content: []protocol.ItemInstance{{StackNetworkID: 4, Stack: item}}},
		"inventory_slot":         &packet.InventorySlot{WindowID: 3, Slot: 2, NewItem: protocol.ItemInstance{StackNetworkID: 4, Stack: item}},
		"mob_equipment":          &packet.MobEquipment{EntityRuntimeID: 9, NewItem: item, InventorySlot: 1, HotBarSlot: 1, WindowID: 0},
		"mob_armour":             &packet.MobArmourEquipment{EntityRuntimeID: 9, Helmet: item, Chestplate: item, Leggings: item, Boots: item},
		"add_item_actor":         &packet.AddItemActor{EntityUniqueID: -4, EntityRuntimeID: 9, Item: item, Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, EntityMetadata: map[uint32]interface{}{0: int64(2)}, FromFishing: true},
		"inventory_transaction":  &packet.InventoryTransaction{LegacyRequestID: 2, LegacySetItemSlots: []protocol.LegacySetItemSlot{{ContainerID: 1, Slots: []byte{2}}}, Actions: []protocol.InventoryAction{{SourceType: protocol.InventoryActionSourceContainer, WindowID: 3, InventorySlot: 2, OldItem: item, NewItem: item}}, TransactionData: &protocol.UseItemTransactionData{ActionType: protocol.UseItemActionClickBlock, BlockPosition: protocol.BlockPos{1, 64, -2}, BlockFace: 2, HotBarSlot: 1, HeldItem: item, Position: mgl32.Vec3{1, 2, 3}, ClickedPosition: mgl32.Vec3{0.25, 0.5, 0.75}, BlockRuntimeID: 9}},
		"item_stack_request":     &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: 8, Actions: []protocol.StackRequestAction{&protocol.SwapStackRequestAction{Source: protocol.StackRequestSlotInfo{ContainerID: 1, Slot: 2, StackNetworkID: 3}, Destination: protocol.StackRequestSlotInfo{ContainerID: 4, Slot: 5, StackNetworkID: 6}}, &protocol.AutoCraftRecipeStackRequestAction{CraftRecipeStackRequestAction: protocol.CraftRecipeStackRequestAction{RecipeNetworkID: 7}}, &protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: 17}}}}},
		"item_stack_response":    &packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusOK, RequestID: 8, ContainerInfo: []protocol.StackResponseContainerInfo{{ContainerID: 28, SlotInfo: []protocol.StackResponseSlotInfo{{Slot: 1, HotbarSlot: 1, Count: 2, StackNetworkID: 4}}}}}}},
		"creative_content":       &packet.CreativeContent{Items: []protocol.CreativeItem{{CreativeItemNetworkID: 17, Item: item}}},
		"crafting_data":          &packet.CraftingData{Recipes: []protocol.Recipe{&protocol.ShapelessRecipe{RecipeID: "test", Input: []protocol.RecipeIngredientItem{{NetworkID: 1, MetadataValue: 2, Count: 3}}, Output: []protocol.ItemStack{item}, UUID: id, Block: "crafting_table", Priority: 4, RecipeNetworkID: 5}}, PotionRecipes: []protocol.PotionRecipe{{InputPotionID: 1, InputPotionMetadata: 2, ReagentItemID: 3, ReagentItemMetadata: 4, OutputPotionID: 5, OutputPotionMetadata: 6}}, ClearRecipes: true},
		"player_auth_input":      &packet.PlayerAuthInput{Pitch: 1, Yaw: 2, Position: mgl32.Vec3{3, 4, 5}, MoveVector: mgl32.Vec2{0.25, -0.5}, HeadYaw: 6, InputData: packet.InputFlagJumping | packet.InputFlagSprinting, InputMode: packet.InputModeMouse, PlayMode: packet.PlayModeNormal, Tick: 99, Delta: mgl32.Vec3{0.1, 0.2, 0.3}},
		"available_commands":     &packet.AvailableCommands{Commands: []protocol.Command{{Name: "test", Description: "desc", Flags: 1, PermissionLevel: 1, Aliases: []string{"t"}, Overloads: []protocol.CommandOverload{{Parameters: []protocol.CommandParameter{{Name: "value", Type: protocol.CommandArgValid | protocol.CommandArgTypeString, Optional: true, Options: 1}}}}}}},
		"player_skin":            &packet.PlayerSkin{UUID: id, Skin: protocol.Skin{SkinID: "skin", SkinResourcePatch: []byte("{}"), SkinImageWidth: 1, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 4}, CapeImageWidth: 1, CapeImageHeight: 1, CapeData: []byte{5, 6, 7, 8}, SkinGeometry: []byte("{}"), AnimationData: []byte("{}"), PremiumSkin: true, CapeID: "cape", FullSkinID: "full", SkinColour: "#112233", ArmSize: "slim", PersonaPieces: []protocol.PersonaPiece{{PieceID: "piece", PieceType: "persona_body", PackID: id.String(), Default: true, ProductID: "product"}}, PieceTintColours: []protocol.PersonaPieceTintColour{{PieceType: "persona_body", Colours: []string{"#ff112233", "#00000000", "#00000000", "#00000000"}}}, Trusted: true}, NewSkinName: "new", OldSkinName: "old"},
		"player_list_add":        &packet.PlayerList{ActionType: packet.PlayerListActionAdd, Entries: []protocol.PlayerListEntry{{UUID: id, EntityUniqueID: -5, Username: "Steve", XUID: "42", PlatformChatID: "pc", BuildPlatform: 7, Skin: protocol.Skin{SkinID: "skin", SkinResourcePatch: []byte("{}"), SkinImageWidth: 1, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 4}, CapeImageWidth: 1, CapeImageHeight: 1, CapeData: []byte{5, 6, 7, 8}, ArmSize: "slim", SkinColour: "#112233"}, Teacher: true, Host: true}}},
		"add_player":             &packet.AddPlayer{UUID: id, Username: "Steve", EntityUniqueID: -5, EntityRuntimeID: 10, PlatformChatID: "pc", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, HeldItem: item, EntityMetadata: map[uint32]interface{}{0: int64(2)}, Flags: packet.AdventureFlagAllowFlight, CommandPermissionLevel: 1, ActionPermissions: packet.ActionPermissionBuildAndMine | packet.ActionPermissionOpenContainers, PermissionLevel: packet.PermissionLevelMember, PlayerUniqueID: -5, DeviceID: "device", BuildPlatform: 7},
		"add_actor":              &packet.AddActor{EntityUniqueID: -4, EntityRuntimeID: 9, EntityType: "minecraft:pig", Position: mgl32.Vec3{1, 2, 3}, Velocity: mgl32.Vec3{4, 5, 6}, Pitch: 7, Yaw: 8, HeadYaw: 9, Attributes: []protocol.Attribute{{Name: "minecraft:health", Min: 0, Max: 20, Value: 15}}, EntityMetadata: map[uint32]interface{}{0: int64(2), 4: "Pig"}, EntityLinks: []protocol.EntityLink{{RiddenEntityUniqueID: 1, RiderEntityUniqueID: 2, Type: 1, Immediate: true}}},
		"start_game":             &packet.StartGame{EntityUniqueID: -5, EntityRuntimeID: 10, PlayerGameMode: 1, PlayerPosition: mgl32.Vec3{1, 2, 3}, Pitch: 4, Yaw: 5, WorldSeed: 6, SpawnBiomeType: packet.SpawnBiomeTypeDefault, Generator: 1, WorldGameMode: 1, Difficulty: 2, WorldSpawn: protocol.BlockPos{1, 64, -2}, AchievementsDisabled: true, DayCycleLockTime: 7, MultiPlayerGame: true, LANBroadcastEnabled: true, CommandsEnabled: true, TexturePackRequired: true, GameRules: map[string]interface{}{"doDaylightCycle": true}, Experiments: []protocol.ExperimentData{{Name: "test", Enabled: true}}, PlayerPermissions: packet.PermissionLevelMember, ServerChunkTickRadius: 4, NewNether: true, BaseGameVersion: "1.16.100", LevelID: "level", WorldName: "world", ServerAuthoritativeMovementMode: packet.AuthoritativeMovementModeServer, Time: 8, EnchantmentSeed: 9, Blocks: []protocol.BlockEntry{{Name: "custom:block", Properties: map[string]interface{}{}}}, Items: []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}}, MultiPlayerCorrelationID: "corr", ServerAuthoritativeInventory: true},
		"level_chunk":            &packet.LevelChunk{ChunkX: -2, ChunkZ: 5, SubChunkCount: 2, CacheEnabled: true, BlobHashes: []uint64{11, 12, 13}, RawPayload: []byte{1, 2, 3}},
		"level_sound_undefined":  &packet.LevelSoundEvent{SoundType: packet.SoundEventUndefined},
		"set_title":              &packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: "hello", FadeInDuration: 1, RemainDuration: 2, FadeOutDuration: 3},
		"structure_block_update": &packet.StructureBlockUpdate{Position: protocol.BlockPos{1, 64, -2}, StructureName: "house", DataField: "data", IncludePlayers: true, ShowBoundingBox: true, StructureBlockType: packet.StructureBlockSave, Settings: protocol.StructureSettings{PaletteName: "default", Size: protocol.BlockPos{2, 3, 4}, Offset: protocol.BlockPos{1, 2, 3}, LastEditingPlayerUniqueID: -5, Rotation: 1, Mirror: 2, Integrity: 1, Seed: 9, Pivot: mgl32.Vec3{1, 2, 3}}, RedstoneSaveMode: packet.StructureRedstoneSaveModeDisk, ShouldTrigger: true},
		"animate_entity":         &packet.AnimateEntity{Animation: "animation.test", NextState: "next", StopCondition: "q.any", Controller: "controller.test", BlendOutTime: 0.5, EntityRuntimeIDs: []uint64{9, 10}},
		"camera_shake":           &packet.CameraShake{Intensity: 1.5, Duration: 2.5, Type: packet.CameraShakeTypeRotational},
		"education_settings":     &packet.EducationSettings{CodeBuilderDefaultURI: "https://example.invalid", CodeBuilderTitle: "Code", CanResizeCodeBuilder: true, OverrideURI: "https://override.invalid", HasQuiz: true},
		"event":                  &packet.Event{EntityRuntimeID: 9, EventType: packet.EventFishBucketed, UsePlayerID: 1},
		"text":                   &packet.Text{TextType: packet.TextTypeChat, NeedsTranslation: true, SourceName: "Steve", Message: "hello", Parameters: []string{"x"}, XUID: "42", PlatformChatID: "pc"},
		"client_bound_map":       &packet.ClientBoundMapItemData{MapID: 4, UpdateFlags: packet.MapUpdateFlagTexture, LockedMap: true, Scale: 1, Width: 1, Height: 1, XOffset: 2, YOffset: 3, Pixels: [][]color.RGBA{{{R: 1, G: 2, B: 3, A: 4}}}},
		"command_output":         &packet.CommandOutput{CommandOrigin: protocol.CommandOrigin{Origin: protocol.CommandOriginPlayer, UUID: id, RequestID: "req"}, OutputType: packet.CommandOutputTypeAllOutput, SuccessCount: 1, OutputMessages: []protocol.CommandOutputMessage{{Success: true, Message: "done", Parameters: []string{"x"}}}},
		"actor_pick":             &packet.ActorPickRequest{EntityUniqueID: -5, HotBarSlot: 2},
		"hurt_armour":            &packet.HurtArmour{Cause: 3, Damage: 4},
		"npc_request":            &packet.NPCRequest{EntityRuntimeID: 9, RequestType: 1, CommandString: "say hi", ActionType: 2},
		"photo_transfer":         &packet.PhotoTransfer{PhotoName: "photo.png", PhotoData: []byte{1, 2}, BookID: "book"},
	}
	result := make(map[string]fixture, len(fixtures))
	for name, pk := range fixtures {
		var buffer bytes.Buffer
		pk.Marshal(protocol.NewWriter(&buffer, -1))
		result[name] = fixture{ID: pk.ID(), Hex: hex.EncodeToString(buffer.Bytes())}
	}
	_ = math.MaxUint32
	return result
}

func roundTrip(direction, rawID, encoded string) {
	id, err := strconv.ParseUint(rawID, 10, 32)
	if err != nil {
		panic(err)
	}
	data, err := hex.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	pool := packet.NewPool()
	if direction == "client" {
		pool = packet.NewPool()
	}
	constructor, ok := pool[uint32(id)]
	if !ok {
		panic(fmt.Sprintf("packet %d is missing", id))
	}
	pk := constructor
	pk.Unmarshal(protocol.NewReader(zeroSafeReader{Reader: bytes.NewReader(data)}, -1))
	var buffer bytes.Buffer
	pk.Marshal(protocol.NewWriter(&buffer, -1))
	_, _ = fmt.Fprint(os.Stdout, hex.EncodeToString(buffer.Bytes()))
}

type zeroSafeReader struct{ *bytes.Reader }

func (r zeroSafeReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return r.Reader.Read(p)
}

func encodePool(direction string, pool packet.Pool, encoded map[uint32]string, skipped map[string]string) {
	ids := make([]int, 0, len(pool))
	for id := range pool {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, rawID := range ids {
		id := uint32(rawID)
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					skipped[fmt.Sprintf("%s/%d", direction, id)] = fmt.Sprint(recovered)
				}
			}()
			var buffer bytes.Buffer
			pool[id].Marshal(protocol.NewWriter(&buffer, -1))
			encoded[id] = hex.EncodeToString(buffer.Bytes())
		}()
	}
}
