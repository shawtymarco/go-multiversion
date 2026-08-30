package v1_26_0

import (
	"github.com/google/uuid"
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
	uBlockPos(io, &pk.WorldSpawn)
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
	forceExperimental, _ := pk.ForceExperimentalGameplay.Value()
	io.Bool(&forceExperimental)
	if io.reading {
		pk.ForceExperimentalGameplay = protocol.Option(forceExperimental)
	}
	io.Uint8(&pk.ChatRestrictionLevel)
	io.Bool(&pk.DisablePlayerInteractions)
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	protocol.PlayerMoveSettings(io.directional(), &pk.PlayerMovementSettings)
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	protocol.Slice(io.directional(), &pk.Blocks)
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	io.String(&pk.GameVersion)
	io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	io.Uint64(&pk.ServerBlockStateChecksum)
	io.UUID(&pk.WorldTemplateID)
	io.Bool(&pk.ClientSideGeneration)
	io.Bool(&pk.UseBlockNetworkIDHashes)
	io.Bool(&pk.ServerAuthoritativeSound)
	protocol.OptionalFunc(io.directional(), &pk.ServerJoinInformation, func(info *protocol.ServerJoinInformation) {
		marshalServerJoinInformation(io, info)
	})
	io.String(&pk.ServerID)
	io.String(&pk.ScenarioID)
	io.String(&pk.WorldID)
	io.String(&pk.OwnerID)
}

func marshalServerJoinInformation(io *wireIO, info *protocol.ServerJoinInformation) {
	protocol.OptionalFunc(io.directional(), &info.GatheringJoinInfo, func(join *protocol.GatheringJoinInfo) {
		experienceID := ""
		if join.ExperienceID != uuid.Nil {
			experienceID = join.ExperienceID.String()
		}
		io.String(&experienceID)
		io.String(&join.ExperienceName)
		worldID := ""
		if value, ok := join.ExperienceWorldID.Value(); ok && value != uuid.Nil {
			worldID = value.String()
		}
		io.String(&worldID)
		worldName, _ := join.ExperienceWorldName.Value()
		io.String(&worldName)
		io.String(&join.CreatorID)
		storeID := ""
		if store, ok := info.StoreEntryPointInfo.Value(); ok {
			storeID = store.StoreID
		}
		io.String(&storeID)
		if io.reading {
			join.ExperienceID, _ = uuid.Parse(experienceID)
			if parsed, err := uuid.Parse(worldID); err == nil {
				join.ExperienceWorldID = protocol.Option(parsed)
			} else {
				join.ExperienceWorldID = protocol.Optional[uuid.UUID]{}
			}
			join.ExperienceWorldName = protocol.Option(worldName)
			join.TargetID = protocol.Optional[uuid.UUID]{}
			join.ScenarioID = protocol.Optional[string]{}
			join.ServerID = protocol.Optional[string]{}
			info.StoreEntryPointInfo = protocol.Option(protocol.StoreEntryPointInfo{StoreID: storeID})
			info.PresenceInfo = protocol.Optional[protocol.PresenceInfo]{}
		}
	})
}
