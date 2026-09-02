package v1_16_100

import "maps"

// ConvertBlockActorNBT removes current block-state documents that protocol
// 419 cannot resolve inside otherwise valid historical block actors.
func (Protocol) ConvertBlockActorNBT(data map[string]any) map[string]any {
	if data["id"] != "FlowerPot" {
		return data
	}
	if _, ok := data["PlantBlock"]; !ok {
		return data
	}
	converted := maps.Clone(data)
	delete(converted, "PlantBlock")
	return converted
}
