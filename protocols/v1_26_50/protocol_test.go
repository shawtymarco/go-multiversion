package v1_26_50

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"os"
	"reflect"
	"testing"
)

func TestFrozenNativeWireOracles(t *testing.T) {
	data, err := os.ReadFile("testdata/native2193.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name     string
		Listener bool
		ID       uint32
		Hex      string
	}
	if err = json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 17 {
		t.Fatalf("fixture count %d, want 17", len(fixtures))
	}
	p := Protocol{}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			payload, err := hex.DecodeString(fixture.Hex)
			if err != nil {
				t.Fatal(err)
			}
			constructor, ok := p.Packets(fixture.Listener)[fixture.ID]
			if !ok {
				t.Fatalf("packet %d missing", fixture.ID)
			}
			decoded := constructor()
			input := bytes.NewBuffer(payload)
			decoded.Marshal(p.NewReader(input, 0, true))
			if input.Len() != 0 {
				t.Fatalf("%d unread bytes", input.Len())
			}
			var output bytes.Buffer
			decoded.Marshal(p.NewWriter(&output, 0))
			if !bytes.Equal(output.Bytes(), payload) {
				t.Fatalf("frozen bytes differ\ngot %x\nwant %x", output.Bytes(), payload)
			}
			native := UnwrapWirePacket(decoded)
			switch fixture.Name {
			case "player-list":
				entry := native.(*packet.PlayerList).Entries[0]
				if entry.PlayFabID != "" || !bytes.Equal(entry.Skin.SkinData, []byte{1, 2, 3, 255}) {
					t.Fatal("skin decode lost data or conflated cosmetic/account identity")
				}
			case "diagnostics":
				pk := native.(*packet.ServerBoundDiagnostics)
				categories, ok := pk.SystemCategories.Value()
				if !ok || len(categories) != 1 || categories[0].SystemIndex != 7 {
					t.Fatal("required categories not promoted")
				}
				position, ok := pk.EntityDiagnostics[0].Position.Value()
				if !ok || position[1] != 2 {
					t.Fatal("required diagnostic position not promoted")
				}
			case "education":
				capabilities, ok := native.(*packet.EducationSettings).AgentCapabilities.Value()
				if !ok {
					t.Fatal("agent capability missing")
				}
				value, ok := capabilities.CanModifyBlocks.Value()
				if !ok || !value {
					t.Fatal("agent flag missing")
				}
			case "attribute-transition":
				pk := native.(*packet.ClientBoundAttributeLayerSync).EnvironmentAttributes[0]
				if pk.PayloadType != protocol.EnvironmentAttributePayloadTypeTransition || pk.TransitionSettings.TotalTransitionTicks != 40 {
					t.Fatal("transition not promoted to native variant")
				}
			case "attribute-noise":
				pk := native.(*packet.ClientBoundAttributeLayerSync).Layers[0].EnvironmentAttributes[0]
				if pk.PayloadType != protocol.EnvironmentAttributePayloadTypeNoiseTransition || pk.NoiseTransitionSettings.NoiseName != "noise" {
					t.Fatal("layer noise name not promoted")
				}
			case "stack-deprecated":
				pk := native.(*packet.ItemStackRequest).Requests[0]
				if _, ok := pk.Actions[0].(*protocol.CraftNonImplementedStackRequestAction); !ok {
					t.Fatal("deprecated action index shifted into reserved native action")
				}
			}
		})
	}
}

func TestPreviewFieldsDoNotLeakIntoFrozenWire(t *testing.T) {
	p := Protocol{}
	if protocol.CurrentProtocol != 2223 || p.ID() != 2193 {
		t.Fatal("wrong native/target identity")
	}
	for _, listener := range []bool{true, false} {
		for id := range p.Packets(listener) {
			if id > packet.IDRecordStarted {
				t.Fatalf("preview-only ID %d leaked", id)
			}
		}
	}
	input := &packet.StartGame{GameVersion: "1.26.60", BaseGameVersion: "1.26.60", EditorLevelMigrationVersion: 1, PropertyData: map[string]any{}}
	before := *input
	converted := p.ConvertFromLatest(input, nil)[0]
	var payload bytes.Buffer
	converted.Marshal(p.NewWriter(&payload, 0))
	decoded := p.Packets(false)[packet.IDStartGame]()
	decoded.Marshal(p.NewReader(&payload, 0, true))
	start := UnwrapWirePacket(decoded).(*packet.StartGame)
	if start.EditorLevelMigrationVersion != 0 || start.GameVersion != Version {
		t.Fatal("preview editor field leaked")
	}
	if !reflect.DeepEqual(*input, before) {
		t.Fatal("conversion or writing mutated native input")
	}
	if len(p.ConvertFromLatest(&packet.ClientboundMatchmakingState{}, nil)) != 0 {
		t.Fatal("preview-only operation not dropped")
	}
}
