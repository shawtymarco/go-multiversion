package packetconv

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
)

// LegacyStartGame excludes native-only world definitions. Historical clients use
// their frozen palettes and derive stairs/connections locally. Biome string
// indices are retained so surviving definitions keep their original references.
func LegacyStartGame(pk packet.Packet) packet.Packet {
	if biomes, ok := pk.(*packet.BiomeDefinitionList); ok {
		cloned := *biomes
		cloned.BiomeDefinitions = make([]protocol.BiomeDefinition, 0, len(biomes.BiomeDefinitions))
		for _, entry := range biomes.BiomeDefinitions {
			if int(entry.NameIndex) < len(biomes.StringList) {
				name := strings.TrimPrefix(biomes.StringList[entry.NameIndex], "minecraft:")
				if name == "dappled_forest" {
					continue
				}
			}
			cloned.BiomeDefinitions = append(cloned.BiomeDefinitions, entry)
		}
		return &cloned
	}
	game, ok := pk.(*packet.StartGame)
	if !ok {
		return pk
	}
	cloned := *game
	cloned.Blocks = make([]protocol.BlockEntry, 0, len(game.Blocks))
	for _, entry := range game.Blocks {
		if !strings.HasPrefix(entry.Name, "minecraft:") {
			cloned.Blocks = append(cloned.Blocks, entry)
		}
	}
	return &cloned
}

// StopOnlySoundUpdate reports whether the effective native update is Stop.
// The preview native model carries one sound update discriminator.
func StopOnlySoundUpdate(pk *packet.ClientboundUpdateSoundData) bool {
	return pk.Update.Type == protocol.SoundDataUpdateStop
}

// UnsupportedNativePacket excludes additions and boss membership messages that
// cannot carry their historical player identity in the native packet model.
func UnsupportedNativePacket(pk packet.Packet) bool {
	if pk.ID() > packet.IDPartyDestinationCookieResponse {
		return true
	}
	if boss, ok := pk.(*packet.BossEvent); ok {
		switch boss.EventType {
		case packet.BossEventRegisterPlayer, packet.BossEventUnregisterPlayer, packet.BossEventRequest:
			return true
		}
	}
	return false
}
