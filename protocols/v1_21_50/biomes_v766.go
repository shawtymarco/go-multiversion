package v1_21_50

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	v766 "github.com/shawtymarco/go-multiversion/data/v766"
)

func marshalBiomeDefinitionList(io *wireIO, _ packet.Packet) {
	definitions := v766.BiomeDefinitions()
	io.Bytes(&definitions)
}
