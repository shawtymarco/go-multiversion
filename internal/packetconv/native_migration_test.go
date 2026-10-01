package packetconv

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"testing"
)

func TestLegacySoundUsesEffectiveNativeOperation(t *testing.T) {
	pk := &packet.ClientboundUpdateSoundData{Update: protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateSetVolume, Volume: .5}}
	if StopOnlySoundUpdate(pk) {
		t.Fatal("native volume update was treated as Stop")
	}
	pk.Update = protocol.SoundDataUpdate{Type: protocol.SoundDataUpdateStop}
	if !StopOnlySoundUpdate(pk) {
		t.Fatal("native Stop update was not recognized")
	}
}

func TestNativeBiomeDefinitionIsNotAdvertisedAfterFallback(t *testing.T) {
	pk := &packet.BiomeDefinitionList{StringList: []string{"forest", "dappled_forest", "tag", "desert"}, BiomeDefinitions: []protocol.BiomeDefinition{{NameIndex: 0}, {NameIndex: 1}, {NameIndex: 3, Tags: protocol.Option([]uint16{2})}}}
	got := LegacyStartGame(pk).(*packet.BiomeDefinitionList)
	if len(got.BiomeDefinitions) != 2 || got.BiomeDefinitions[1].NameIndex != 3 || got.StringList[3] != "desert" {
		t.Fatal("native-only filtering changed surviving dictionary indices")
	}
	if len(pk.BiomeDefinitions) != 3 {
		t.Fatal("biome filtering mutated the source")
	}
}
