// Frozen protocol-2193 wire from gophertunnel 9e7f28092180fd2cae6c2a0ebc32f1c8882d26f1 (MIT).
package v1_26_50

import (
	"fmt"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/internal/packetconv"
	"github.com/shawtymarco/go-multiversion/internal/packetio"
)

func marshalStartGame(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.StartGame)
	if reading {
		pk.EditorLevelMigrationVersion = 0
	}

	io.ActorUniqueID(&pk.EntityUniqueID)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.Varint32(&pk.PlayerGameMode)
	io.Vec3(&pk.PlayerPosition)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Int64(&pk.WorldSeed)
	io.Int16(&pk.SpawnBiomeType)
	io.String(&pk.UserDefinedBiomeName)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Generator)
	io.Varint32(&pk.WorldGameMode)
	io.Bool(&pk.Hardcore)
	io.Varint32(&pk.Difficulty)
	io.BlockPos(&pk.WorldSpawn)
	io.Bool(&pk.AchievementsDisabled)
	io.Varint32(&pk.EditorWorldType)
	io.Bool(&pk.CreatedInEditor)
	io.Bool(&pk.ExportedFromEditor)
	io.Varint32(&pk.DayCycleLockTime)
	io.Varuint32(&pk.EducationEditionOffer)
	io.Bool(&pk.EducationFeaturesEnabled)
	io.String(&pk.EducationProductID)
	io.Float32(&pk.RainLevel)
	io.Float32(&pk.LightningLevel)
	io.Bool(&pk.ConfirmedPlatformLockedContent)
	io.Bool(&pk.MultiPlayerGame)
	io.Bool(&pk.LANBroadcastEnabled)
	io.Varint32(&pk.XBLBroadcastMode)
	io.Varint32(&pk.PlatformBroadcastMode)
	io.Bool(&pk.CommandsEnabled)
	io.Bool(&pk.TexturePackRequired)
	protocol.FuncSlice(io, &pk.GameRules, io.GameRule)
	protocol.SliceUint32Length(io, &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	io.Bool(&pk.BonusChestEnabled)
	io.Bool(&pk.StartWithMapEnabled)
	io.Uint8(&pk.PlayerPermissions)
	io.Int32(&pk.ServerChunkTickRadius)
	io.Bool(&pk.HasLockedBehaviourPack)
	io.Bool(&pk.HasLockedTexturePack)
	io.Bool(&pk.FromLockedWorldTemplate)
	io.Bool(&pk.MSAGamerTagsOnly)
	io.Bool(&pk.FromWorldTemplate)
	io.Bool(&pk.WorldTemplateSettingsLocked)
	io.Bool(&pk.OnlySpawnV1Villagers)
	io.Bool(&pk.PersonaDisabled)
	io.Bool(&pk.CustomSkinsDisabled)
	io.Bool(&pk.EmoteChatMuted)
	io.String(&pk.BaseGameVersion)
	io.Int32(&pk.LimitedWorldWidth)
	io.Int32(&pk.LimitedWorldDepth)
	io.Bool(&pk.NewNether)
	protocol.Single(io, &pk.EducationSharedResourceURI)
	protocol.OptionalFunc(io, &pk.ForceExperimentalGameplay, io.Bool)
	io.Uint8(&pk.ChatRestrictionLevel)
	io.Bool(&pk.DisablePlayerInteractions)
	io.Varint32(&pk.ServerEditorConnectionPolicy)
	io.Bool(&pk.AllowAnonymousBlockDropsInEditorWorlds)
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	protocol.PlayerMoveSettings(io, &pk.PlayerMovementSettings)
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	protocol.Slice(io, &pk.Blocks)
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	io.String(&pk.GameVersion)
	io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	io.Uint64(&pk.ServerBlockStateChecksum)
	io.UUID(&pk.WorldTemplateID)
	io.Bool(&pk.ClientSideGeneration)
	io.Bool(&pk.UseBlockNetworkIDHashes)
	io.Bool(&pk.ServerAuthoritativeSound)
	protocol.OptionalMarshaler(io, &pk.ServerJoinInformation)
	io.String(&pk.ServerID)
	io.String(&pk.ScenarioID)
	io.String(&pk.WorldID)
	io.String(&pk.OwnerID)

}
func marshalAddActor(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.AddActor)
	if reading {
		pk.PassengerOfBlock = protocol.Optional[protocol.PassengerOfBlockData]{}
	}

	io.ActorUniqueID(&pk.EntityUniqueID)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.String(&pk.EntityType)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.Float32(&pk.BodyYaw)
	protocol.Slice(io, &pk.Attributes)
	io.EntityMetadata(&pk.EntityMetadata)
	protocol.Single(io, &pk.EntityProperties)
	protocol.Slice(io, &pk.EntityLinks)

}
func marshalAddPlayer(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.AddPlayer)
	if reading {
		pk.PassengerOfBlock = protocol.Optional[protocol.PassengerOfBlockData]{}
	}

	io.UUID(&pk.UUID)
	io.String(&pk.Username)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.String(&pk.PlatformChatID)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.ItemInstance(&pk.HeldItem)
	io.Varint32(&pk.GameType)
	io.EntityMetadata(&pk.EntityMetadata)
	protocol.Single(io, &pk.EntityProperties)
	protocol.Single(io, &pk.AbilityData)
	protocol.Slice(io, &pk.EntityLinks)
	io.String(&pk.DeviceID)
	io.Int32(&pk.BuildPlatform)

}
func marshalAnimate(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.Animate)
	if reading {
		pk.Hand = protocol.Optional[uint8]{}
	}

	var swingSource protocol.Optional[string]
	if pk.SwingSource != 0 {
		swingSource = protocol.Option(swingSourceToString(pk.SwingSource))
	}
	io.Uint8(&pk.ActionType)
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.Float32(&pk.Data)
	protocol.OptionalFunc(io, &swingSource, io.String)
	if val, ok := swingSource.Value(); ok {
		swingSourceFromString(io, &pk.SwingSource, val)
	}

}
func swingSourceFromString(io protocol.IO, x *uint8, s string) {
	switch s {
	case "none":
		*x = packet.AnimateSwingSourceNone
	case "build":
		*x = packet.AnimateSwingSourceBuild
	case "mine":
		*x = packet.AnimateSwingSourceMine
	case "interact":
		*x = packet.AnimateSwingSourceInteract
	case "attack":
		*x = packet.AnimateSwingSourceAttack
	case "useitem":
		*x = packet.AnimateSwingSourceUseItem
	case "throwitem":
		*x = packet.AnimateSwingSourceThrowItem
	case "dropitem":
		*x = packet.AnimateSwingSourceDropItem
	case "event":
		*x = packet.AnimateSwingSourceEvent
	default:
		io.InvalidValue(s, "swingSource", "unknown source")
	}
}
func swingSourceToString(x uint8) string {
	switch x {
	case packet.AnimateSwingSourceNone:
		return "none"
	case packet.AnimateSwingSourceBuild:
		return "build"
	case packet.AnimateSwingSourceMine:
		return "mine"
	case packet.AnimateSwingSourceInteract:
		return "interact"
	case packet.AnimateSwingSourceAttack:
		return "attack"
	case packet.AnimateSwingSourceUseItem:
		return "useitem"
	case packet.AnimateSwingSourceThrowItem:
		return "throwitem"
	case packet.AnimateSwingSourceDropItem:
		return "dropitem"
	case packet.AnimateSwingSourceEvent:
		return "event"
	default:
		return "unknown"
	}
}
func marshalLevelChunk(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.LevelChunk)
	if reading {
		pk.ClientBiomeUpdate = false
	}

	io.ChunkPos(&pk.Position)
	io.Varint32(&pk.Dimension)
	io.Varuint32(&pk.SubChunkCount)
	if pk.SubChunkCount > 64 {
		io.InvalidValue(pk.SubChunkCount, "level chunk sub-chunk count", "must not exceed 64")
	}
	protocol.OptionalFunc(io, &pk.SubChunkLimit, io.Varint32)
	io.Bool(&pk.CacheEnabled)
	protocol.FuncSlice(io, &pk.BlobHashes, io.Uint64)
	io.ByteSlice(&pk.RawPayload)

}
func marshalPlayerList(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.PlayerList)

	protocol.FuncIOSlice(io, &pk.Entries, func(raw protocol.IO, entry *protocol.PlayerListEntry) { marshalPlayerListEntry(raw, entry, reading) })

}
func marshalPlayerSkin(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.PlayerSkin)

	io.UUID(&pk.UUID)
	marshalSkin(io, &pk.Skin)
	io.String(&pk.NewSkinName)
	io.String(&pk.OldSkinName)

}
func marshalDimensionData(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.DimensionData)

	protocol.FuncIOSlice(io, &pk.Definitions, marshalDimensionDefinition)

}
func marshalEducationSettings(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.EducationSettings)

	io.String(&pk.CodeBuilderDefaultURI)
	io.String(&pk.CodeBuilderTitle)
	io.Bool(&pk.CanResizeCodeBuilder)
	io.Bool(&pk.DisableLegacyTitleBar)
	io.String(&pk.PostProcessFilter)
	io.String(&pk.ScreenshotBorderPath)
	capabilities, _ := pk.AgentCapabilities.Value()
	canModify := capabilities.CanModifyBlocks
	protocol.OptionalFunc(io, &canModify, io.Bool)
	if reading {
		if _, ok := canModify.Value(); ok {
			pk.AgentCapabilities = protocol.Option(protocol.EducationAgentCapabilities{CanModifyBlocks: canModify})
		} else {
			pk.AgentCapabilities = protocol.Optional[protocol.EducationAgentCapabilities]{}
		}
	}
	protocol.OptionalFunc(io, &pk.OverrideURI, io.String)
	io.Bool(&pk.HasQuiz)
	protocol.OptionalMarshaler(io, &pk.ExternalLinkSettings)

}
func marshalServerBoundDiagnostics(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.ServerBoundDiagnostics)

	io.Float32(&pk.AverageFramesPerSecond)
	io.Float32(&pk.AverageServerSimTickTime)
	io.Float32(&pk.AverageClientSimTickTime)
	io.Float32(&pk.AverageBeginFrameTime)
	io.Float32(&pk.AverageInputTime)
	io.Float32(&pk.AverageRenderTime)
	io.Float32(&pk.AverageEndFrameTime)
	io.Float32(&pk.AverageRemainderTimePercent)
	io.Float32(&pk.AverageUnaccountedTimePercent)
	protocol.FuncIOSlice(io, &pk.MemoryCategoryValues, func(raw protocol.IO, x *protocol.MemoryCategoryCounter) {
		category := x.Category
		if !reading {
			category = packetconv.MemoryCategory(ID, category, true)
		}
		raw.Uint8(&category)
		if reading {
			x.Category = packetconv.MemoryCategory(ID, category, false)
		}
		raw.Uint64(&x.Bytes)
	})
	protocol.FuncIOSlice(io, &pk.EntityDiagnostics, func(raw protocol.IO, value *protocol.EntityDiagnosticTimingInfo) {
		marshalEntityDiagnostic(raw, value, reading)
	})
	protocol.Slice(io, &pk.SystemDiagnostics)
	packetio.RequiredSystemCategories(io, &pk.SystemCategories, reading)
	protocol.Slice(io, &pk.WhiskerScopes)

}
func marshalClientboundUpdateSoundData(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.ClientboundUpdateSoundData)
	update := pk.Update

	io.Uint64(&pk.ServerSoundHandle)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)
	marshalSoundUpdate(io, &update, reading)

	if reading {
		pk.Update = update
	}

}
func marshalClientBoundAttributeLayerSync(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.ClientBoundAttributeLayerSync)

	io.Varuint32(&pk.PayloadType)
	switch pk.PayloadType {
	case protocol.AttributeLayerPayloadTypeUpdateLayers:
		protocol.FuncIOSlice(io, &pk.Layers, func(raw protocol.IO, layer *protocol.AttributeLayerData) { marshalAttributeLayer(raw, layer, reading) })
	case protocol.AttributeLayerPayloadTypeUpdateSettings:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.Single(io, &pk.Settings)
	case protocol.AttributeLayerPayloadTypeUpdateEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncIOSlice(io, &pk.EnvironmentAttributes, func(raw protocol.IO, value *protocol.EnvironmentAttributeData) {
			marshalEnvironmentAttribute(raw, value, reading)
		})
	case protocol.AttributeLayerPayloadTypeRemoveEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncSlice(io, &pk.RemoveAttributeNames, io.String)
	default:
		io.UnknownEnumOption(pk.PayloadType, "attribute layer payload type")
	}

}
func marshalSkin(r protocol.IO, x *protocol.Skin) {
	if _, reading := r.(interface{ SliceLength(uint32, uint32) }); !reading {
		copySkin := *x
		x = &copySkin
	}

	r.String(&x.SkinID)
	cosmeticPlayFabID := ""
	r.String(&cosmeticPlayFabID)
	r.ByteSlice(&x.SkinResourcePatch)
	r.Uint32(&x.SkinImageWidth)
	r.Uint32(&x.SkinImageHeight)
	r.ByteSlice(&x.SkinData)
	protocol.Slice(r, &x.Animations)
	r.Uint32(&x.CapeImageWidth)
	r.Uint32(&x.CapeImageHeight)
	r.ByteSlice(&x.CapeData)
	r.ByteSlice(&x.SkinGeometry)
	r.ByteSlice(&x.GeometryDataEngineVersion)
	r.ByteSlice(&x.AnimationData)
	r.String(&x.CapeID)
	r.String(&x.FullID)
	r.Uint8(&x.ArmSize)
	r.BEARGB(&x.SkinColour)
	protocol.Slice(r, &x.PersonaPieces)
	protocol.Slice(r, &x.PieceTintColours)
	if err := validateSkin(*x); err != nil {
		r.InvalidValue(fmt.Sprintf("Skin %v", x.SkinID), "serialised skin", err.Error())
	}
	r.Bool(&x.PremiumSkin)
	r.Bool(&x.PersonaSkin)
	r.Bool(&x.PersonaCapeOnClassicSkin)
	r.Bool(&x.PrimaryUser)
	r.Bool(&x.OverrideAppearance)
	trusted := "false"
	if x.Trusted {
		trusted = "true"
	}
	r.String(&trusted)
	x.Trusted = strings.EqualFold(trusted, "true")
	r.String(&x.ProfileHash)

}
func validateSkin(x protocol.Skin) error {

	if x.SkinImageHeight*x.SkinImageWidth*4 != uint32(len(x.SkinData)) {
		return fmt.Errorf("expected size of skin is %vx%v (%v bytes total), but got %v bytes", x.SkinImageWidth, x.SkinImageHeight, x.SkinImageHeight*x.SkinImageWidth*4, len(x.SkinData))
	}
	if x.CapeImageHeight*x.CapeImageWidth*4 != uint32(len(x.CapeData)) {
		return fmt.Errorf("expected size of cape is %vx%v (%v bytes total), but got %v bytes", x.CapeImageWidth, x.CapeImageHeight, x.CapeImageHeight*x.CapeImageWidth*4, len(x.CapeData))
	}
	for i, animation := range x.Animations {
		if animation.ImageHeight*animation.ImageWidth*4 != uint32(len(animation.ImageData)) {
			return fmt.Errorf("expected size of animation %v is %vx%v (%v bytes total), but got %v bytes", i, animation.ImageWidth, animation.ImageHeight, animation.ImageHeight*animation.ImageWidth*4, len(animation.ImageData))
		}
	}
	return nil

}
func marshalPlayerListEntry(r protocol.IO, x *protocol.PlayerListEntry, reading bool) {
	if reading {
		x.PlayFabID = ""
	}

	playerListAction(r, &x.ActionType)
	r.UUID(&x.UUID)
	if x.ActionType == protocol.PlayerListActionRemove {
		return
	}

	r.ActorUniqueID(&x.EntityUniqueID)
	r.String(&x.Username)
	r.String(&x.XUID)
	r.String(&x.PlatformChatID)
	r.Int32(&x.BuildPlatform)
	marshalSkin(r, &x.Skin)
	r.Bool(&x.Teacher)
	r.Bool(&x.Host)
	r.Bool(&x.SubClient)
	r.BEARGB(&x.PlayerColour)

}
func playerListAction(r protocol.IO, action *byte) {
	variant := uint32(0)
	if *action == protocol.PlayerListActionAdd {
		variant = 1
	}
	r.Varuint32(&variant)

	legacyAction := *action
	r.Uint8(&legacyAction)
	*action = protocol.PlayerListActionRemove
	if variant == 1 {
		*action = protocol.PlayerListActionAdd
	} else if variant != 0 {
		r.UnknownEnumOption(variant, "player list entry variant")
	}
}
func marshalDimensionDefinition(r protocol.IO, x *protocol.DimensionDefinition) {

	r.String(&x.Name)
	r.Varint32(&x.MinimumY)
	r.Varint32(&x.HeightRange)
	r.Varint32(&x.Generator)
	r.Varint32(&x.DimensionType)
	r.UUID(&x.PackID)
	r.String(&x.DefaultBiome)

}

func marshalInventoryTransaction(io protocol.IO, raw packet.Packet, reading bool) {
	pk := raw.(*packet.InventoryTransaction)

	io.Varint32(&pk.LegacyRequestID)
	hasLegacy := pk.LegacyRequestID < -1 && (pk.LegacyRequestID&1) == 0
	io.Bool(&hasLegacy)
	if hasLegacy {
		protocol.Slice(io, &pk.LegacySetItemSlots)
	}
	io.TransactionDataType(&pk.TransactionData)
	protocol.Slice(io, &pk.Actions)
	packetio.TransactionBody2193(io, pk.TransactionData, reading)
}
