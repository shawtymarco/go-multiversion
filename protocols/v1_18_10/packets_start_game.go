package v1_18_10

import (
	"fmt"

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
	seed := int32(pk.WorldSeed)
	io.Varint32(&seed)
	pk.WorldSeed = int64(seed)
	io.Int16(&pk.SpawnBiomeType)
	io.String(&pk.UserDefinedBiomeName)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Generator)
	io.Varint32(&pk.WorldGameMode)
	io.Varint32(&pk.Difficulty)
	marshalUnsignedBlockPos(io, &pk.WorldSpawn)
	io.Bool(&pk.AchievementsDisabled)
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
	protocol.FuncIOSlice(io.directional(), &pk.GameRules, marshalGameRule)
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
	io.String(&pk.BaseGameVersion)
	io.Int32(&pk.LimitedWorldWidth)
	io.Int32(&pk.LimitedWorldDepth)
	io.Bool(&pk.NewNether)
	protocol.Single(io.directional(), &pk.EducationSharedResourceURI)
	forceExperimental, _ := pk.ForceExperimentalGameplay.Value()
	io.Bool(&forceExperimental)
	if forceExperimental {
		io.Bool(&forceExperimental)
	}
	if io.reading {
		pk.ForceExperimentalGameplay = protocol.Option(forceExperimental)
	}
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	marshalPlayerMovementSettings(io, &pk.PlayerMovementSettings)
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	protocol.Slice(io.directional(), &pk.Blocks)
	marshalStartGameItems(io)
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	io.String(&pk.GameVersion)
	io.Uint64(&pk.ServerBlockStateChecksum)

	if io.reading {
		pk.Hardcore = false
		pk.EditorWorldType = 0
		pk.CreatedInEditor = false
		pk.ExportedFromEditor = false
		pk.PersonaDisabled = false
		pk.CustomSkinsDisabled = false
		pk.EmoteChatMuted = false
		pk.ChatRestrictionLevel = 0
		pk.DisablePlayerInteractions = false
		pk.ClientSideGeneration = false
		pk.UseBlockNetworkIDHashes = false
		pk.ServerAuthoritativeSound = false
	}
}

func marshalGameRule(raw protocol.IO, rule *protocol.GameRule) {
	io := asWireIO(raw)
	io.String(&rule.Name)
	io.Bool(&rule.CanBeModifiedByPlayer)
	var kind uint32
	if !io.reading {
		switch rule.Value.(type) {
		case bool:
			kind = 1
		case uint32:
			kind = 2
		case float32:
			kind = 3
		default:
			io.UnknownEnumOption(fmt.Sprintf("%T", rule.Value), "game rule type")
			return
		}
	}
	io.Varuint32(&kind)
	switch kind {
	case 1:
		value, _ := rule.Value.(bool)
		io.Bool(&value)
		rule.Value = value
	case 2:
		value, _ := rule.Value.(uint32)
		io.Varuint32(&value)
		rule.Value = value
	case 3:
		value, _ := rule.Value.(float32)
		io.Float32(&value)
		rule.Value = value
	default:
		io.UnknownEnumOption(kind, "game rule type")
	}
}

func marshalPlayerMovementSettings(io *wireIO, settings *protocol.PlayerMovementSettings) {
	legacyMovementType := int32(1)
	io.Varint32(&legacyMovementType)
	io.Varint32(&settings.RewindHistorySize)
	io.Bool(&settings.ServerAuthoritativeBlockBreaking)
}

func marshalStartGameItems(io *wireIO) {
	var items []protocol.ItemEntry
	if !io.reading {
		if io.runtime == nil || io.runtime.currentItemMapper() == nil {
			io.InvalidValue(nil, "start game items", "protocol 486 item mapping is not configured")
			return
		}
		items = io.runtime.currentItemMapper().TargetEntries()
	}
	count := uint32(len(items))
	io.Varuint32(&count)
	if io.reading {
		items = make([]protocol.ItemEntry, count)
	}
	for index := range items {
		io.String(&items[index].Name)
		io.Int16(&items[index].RuntimeID)
		io.Bool(&items[index].ComponentBased)
	}
}

func marshalBiomeDefinitionList(io *wireIO, _ packet.Packet) {
	if io.reading {
		var ignored []byte
		io.Bytes(&ignored)
		return
	}
	if io.runtime == nil || len(io.runtime.biomes) == 0 {
		io.InvalidValue(nil, "biome definitions", "protocol 486 biome data is not configured")
		return
	}
	biomes := append([]byte(nil), io.runtime.biomes...)
	io.Bytes(&biomes)
}

func marshalCreativeContent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CreativeContent)
	count := uint32(len(pk.Items))
	io.Varuint32(&count)
	if io.reading {
		pk.Groups = nil
		pk.Items = make([]protocol.CreativeItem, count)
	}
	for index := range pk.Items {
		io.Varuint32(&pk.Items[index].CreativeItemNetworkID)
		io.Item(&pk.Items[index].Item)
		if io.reading {
			pk.Items[index].GroupIndex = 0
		}
	}
}
