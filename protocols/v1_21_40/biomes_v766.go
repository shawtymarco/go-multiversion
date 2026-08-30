package v1_21_40

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	v748 "github.com/shawtymarco/go-multiversion/data/v748"
)

func marshalBiomeDefinitionList(io *wireIO, _ packet.Packet) {
	definitions := v748.BiomeDefinitions()
	io.Bytes(&definitions)
}
