package v1_21_50

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestFormPacketConversion(t *testing.T) {
	input := []byte(`{"type":"form","title":"games","content":"body","elements":[{"type":"header","text":"section"},{"type":"button","text":"one"},{"type":"button","text":"two"}]}`)
	tests := []struct {
		name   string
		packet packet.Packet
		data   func(packet.Packet) []byte
	}{
		{name: "modal request", packet: &packet.ModalFormRequest{FormID: 7, FormData: bytes.Clone(input)}, data: func(pk packet.Packet) []byte { return pk.(*packet.ModalFormRequest).FormData }},
		{name: "server settings", packet: &packet.ServerSettingsResponse{FormID: 8, FormData: bytes.Clone(input)}, data: func(pk packet.Packet) []byte { return pk.(*packet.ServerSettingsResponse).FormData }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converted := (Protocol{runtime: &runtimeData{}}).convertGameplayFromLatest(test.packet, nil)
			if len(converted) != 1 || converted[0] == test.packet {
				t.Fatalf("packet was not cloned: %#v", converted)
			}
			if !bytes.Equal(test.data(test.packet), input) {
				t.Fatalf("input packet was mutated: %s", test.data(test.packet))
			}
			var form map[string]any
			if err := json.Unmarshal(test.data(converted[0]), &form); err != nil {
				t.Fatal(err)
			}
			if _, exists := form["elements"]; exists {
				t.Fatalf("target form retains elements: %#v", form)
			}
			buttons, ok := form["buttons"].([]any)
			if !ok || len(buttons) != 2 {
				t.Fatalf("target buttons: %#v", form["buttons"])
			}
		})
	}
}
