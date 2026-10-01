package v1_26_30

import (
	"bytes"
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Independent bytes from 0a2ecd5633ea1466ff97f6d4718df66ec14d054f.
// These unchanged target packets must not inherit the new native nested types.
func TestNativeUpgradeSharedWireOracles(t *testing.T) {
	fixtures := func() map[string]packet.Packet {
		return map[string]packet.Packet{
			"boss_event":       &packet.BossEvent{BossEntityUniqueID: 12, EventType: packet.BossEventShow, BossBarTitle: "boss", FilteredBossBarTitle: "clear", HealthPercentage: .5, Colour: 2},
			"camera_presets":   &packet.CameraPresets{Presets: []protocol.CameraPreset{{Name: "test", Parent: "minecraft:first_person", PosX: protocol.Option(float32(2))}}},
			"primitive_text":   &packet.PrimitiveShapes{Shapes: []protocol.PrimitiveShape{{NetworkID: 1, ExtraShapeData: &protocol.TextShape{Text: "text", DepthTest: true}}}},
			"attribute_layers": &packet.ClientBoundAttributeLayerSync{PayloadType: protocol.AttributeLayerPayloadTypeUpdateEnvironment, LayerName: "layer", EnvironmentAttributes: []protocol.EnvironmentAttributeData{{AttributeName: "fog", Attribute: protocol.AttributeData{Type: protocol.AttributeDataTypeBool, BoolValue: true}}}},
			"diagnostics":      &packet.ServerBoundDiagnostics{AverageFramesPerSecond: 60, MemoryCategoryValues: []protocol.MemoryCategoryCounter{{Category: protocol.MemoryCategoryRendering, Bytes: 99}}, EntityDiagnostics: []protocol.EntityDiagnosticTimingInfo{{DisplayName: "pig", Entity: "minecraft:pig", DurationNanos: 7, PercentOfTotal: 11}}},
		}
	}
	bodies := map[string]string{
		"attribute_layers": "AgVsYXllcgABA2ZvZwAAAQAAAAAAAAAAAAAGbGluZWFyAAAAAAA=",
		"boss_event":       "GAAABGJvc3MFY2xlYXIAAAA/AgA=",
		"camera_presets":   "AQR0ZXN0Fm1pbmVjcmFmdDpmaXJzdF9wZXJzb24BAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"diagnostics":      "AABwQgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAATtjAAAAAAAAAAEDcGlnDW1pbmVjcmFmdDpwaWcHAAAAAAAAAAsAAA==",
		"primitive_text":   "AQEAAAAAAAAAAAACBHRleHQAAAEAAA==",
	}
	for name, source := range fixtures() {
		t.Run(name, func(t *testing.T) {
			want, err := base64.StdEncoding.DecodeString(bodies[name])
			if err != nil {
				t.Fatal(err)
			}
			p := Protocol{}
			converted := p.ConvertFromLatest(source, nil)
			if len(converted) != 1 {
				t.Fatalf("conversion count = %d", len(converted))
			}
			var encoded bytes.Buffer
			converted[0].Marshal(p.NewWriter(&encoded, 0))
			if !bytes.Equal(encoded.Bytes(), want) {
				t.Fatalf("body mismatch: got %x want %x", encoded.Bytes(), want)
			}
			if !reflect.DeepEqual(source, fixtures()[name]) {
				t.Fatal("encoding mutated source")
			}
			pool := p.Packets(name == "diagnostics")
			constructor := pool[source.ID()]
			if constructor == nil {
				t.Fatal("packet missing from historical pool")
			}
			decoded := constructor()
			reader := bytes.NewBuffer(want)
			decoded.Marshal(p.NewReader(reader, 0, true))
			if reader.Len() != 0 {
				t.Fatalf("unread bytes = %d", reader.Len())
			}
			latest := p.ConvertToLatest(decoded, nil)
			if len(latest) == 1 && name == "diagnostics" {
				got := latest[0].(*packet.ServerBoundDiagnostics)
				if len(got.SystemDiagnostics) == 0 {
					got.SystemDiagnostics = nil
				}
				if len(got.WhiskerScopes) == 0 {
					got.WhiskerScopes = nil
				}
			}
			if len(latest) != 1 || !reflect.DeepEqual(latest[0], fixtures()[name]) {
				t.Fatalf("historical semantics changed: %#v", latest)
			}
		})
	}
}
