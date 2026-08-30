package v1_21_50

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalStartGame(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StartGame)
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
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
	io.UBlockPos(&pk.WorldSpawn)
	io.Bool(&pk.AchievementsDisabled)
	io.Varint32(&pk.EditorWorldType)
	io.Bool(&pk.CreatedInEditor)
	io.Bool(&pk.ExportedFromEditor)
	io.Varint32(&pk.DayCycleLockTime)
	educationOffer := int32(pk.EducationEditionOffer)
	io.Varint32(&educationOffer)
	pk.EducationEditionOffer = uint32(educationOffer)
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
	protocol.FuncIOSlice(io.directional(), &pk.GameRules, func(io protocol.IO, rule *protocol.GameRule) {
		marshalGameRule(asWireIO(io), rule, true)
	})
	protocol.SliceUint32Length(io.directional(), &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	io.Bool(&pk.BonusChestEnabled)
	io.Bool(&pk.StartWithMapEnabled)
	permissions := int32(pk.PlayerPermissions)
	io.Varint32(&permissions)
	pk.PlayerPermissions = byte(permissions)
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
	protocol.Single(io.directional(), &pk.EducationSharedResourceURI)
	forceExperimentalGameplay, _ := pk.ForceExperimentalGameplay.Value()
	io.Bool(&forceExperimentalGameplay)
	if io.reading {
		pk.ForceExperimentalGameplay = protocol.Option(forceExperimentalGameplay)
	}
	io.Uint8(&pk.ChatRestrictionLevel)
	io.Bool(&pk.DisablePlayerInteractions)
	io.String(&pk.ServerID)
	io.String(&pk.WorldID)
	io.String(&pk.ScenarioID)
	if io.reading {
		pk.OwnerID = ""
	}
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	movementType := int32(1)
	io.Varint32(&movementType)
	io.Varint32(&pk.PlayerMovementSettings.RewindHistorySize)
	io.Bool(&pk.PlayerMovementSettings.ServerAuthoritativeBlockBreaking)
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	protocol.Slice(io.directional(), &pk.Blocks)
	var items []protocol.ItemEntry
	if !io.reading && io.runtime != nil {
		if mapper := io.runtime.currentItemMapper(); mapper != nil {
			items = mapper.TargetEntries()
		}
	}
	protocol.FuncIOSlice(io.directional(), &items, func(raw protocol.IO, item *protocol.ItemEntry) {
		legacy := asWireIO(raw)
		legacy.String(&item.Name)
		legacy.Int16(&item.RuntimeID)
		legacy.Bool(&item.ComponentBased)
	})
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	io.String(&pk.GameVersion)
	io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	io.Uint64(&pk.ServerBlockStateChecksum)
	io.UUID(&pk.WorldTemplateID)
	io.Bool(&pk.ClientSideGeneration)
	io.Bool(&pk.UseBlockNetworkIDHashes)
	io.Bool(&pk.ServerAuthoritativeSound)
	if io.reading {
		pk.ServerEditorConnectionPolicy = 0
		pk.AllowAnonymousBlockDropsInEditorWorlds = false
		pk.ServerJoinInformation = protocol.Optional[protocol.ServerJoinInformation]{}
	}
}

func marshalServerJoinInformation(io *wireIO, info *protocol.ServerJoinInformation) {
	protocol.OptionalFunc(io.directional(), &info.GatheringJoinInfo, func(join *protocol.GatheringJoinInfo) {
		io.UUID(&join.ExperienceID)
		io.String(&join.ExperienceName)
		worldID, _ := join.ExperienceWorldID.Value()
		io.UUID(&worldID)
		join.ExperienceWorldID = protocol.Option(worldID)
		worldName, _ := join.ExperienceWorldName.Value()
		io.String(&worldName)
		join.ExperienceWorldName = protocol.Option(worldName)
		io.String(&join.CreatorID)
		targetID, _ := join.TargetID.Value()
		io.UUID(&targetID)
		join.TargetID = protocol.Option(targetID)
		scenarioID, _ := join.ScenarioID.Value()
		io.String(&scenarioID)
		join.ScenarioID = protocol.Option(scenarioID)
		serverID, _ := join.ServerID.Value()
		io.String(&serverID)
		join.ServerID = protocol.Option(serverID)
	})
	protocol.OptionalFunc(io.directional(), &info.StoreEntryPointInfo, func(store *protocol.StoreEntryPointInfo) {
		io.String(&store.StoreID)
		io.String(&store.StoreName)
	})
	protocol.OptionalFunc(io.directional(), &info.PresenceInfo, func(presence *protocol.PresenceInfo) {
		var experienceName, worldName protocol.Optional[string]
		protocol.OptionalFunc(io.directional(), &experienceName, io.String)
		protocol.OptionalFunc(io.directional(), &worldName, io.String)
		richPresenceID, _ := presence.RichPresenceID.Value()
		io.String(&richPresenceID)
		presence.RichPresenceID = protocol.Option(richPresenceID)
	})
}
