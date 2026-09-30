package packetconv_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/protocols/v1_16_100"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_0"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_10"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_100"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_110"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_130"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_40"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_50"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_0"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_10"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_20"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_30"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_44"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_45"
)

func debugRendererProtocols() []minecraft.Protocol {
	return []minecraft.Protocol{v1_16_100.New(), v1_18_0.New(), v1_18_10.New(), v1_21_40.New(), v1_21_50.New(),
		v1_21_100.New(), v1_21_110.New(), v1_21_130.New(), v1_26_0.New(), v1_26_10.New(), v1_26_20.New(),
		v1_26_30.New(), v1_26_44.New(), v1_26_45.New(), minecraft.DefaultProtocol}
}

// The integer-action oracle follows c40bf828/2cb1e399 and
// 268adeb5/ecff04b7/49e707e/bf05a1a. The string-action oracle follows the
// corrected 2193 model (98c3d1f + 54e03d9): optional marker, packed BE ARGB.
const legacyDebugCube = "0200000004637562650000803f00000040000040400000803f00000000000000000000803f1400000000000000"

func TestDebugRendererWireOracles(t *testing.T) {
	for _, p := range debugRendererProtocols() {
		t.Run(p.Ver(), func(t *testing.T) {
			for _, cube := range []bool{false, true} {
				original := &packet.ClientBoundDebugRenderer{Type: packet.ClientBoundDebugRendererClear}
				wantHex := "01000000"
				if p.ID() > 844 {
					wantHex = "11636c65617264656275676d61726b65727300"
				}
				if cube {
					original.Type = packet.ClientBoundDebugRendererAddCube
					original.Data = protocol.Option(packet.DebugMarkerData{Text: "cube", Position: mgl32.Vec3{1, 2, 3},
						Colour: color.RGBA{R: 255, A: 255}, Duration: 20})
					wantHex = legacyDebugCube
					if p.ID() > 844 {
						wantHex = "1261646464656275676d61726b6572637562650104637562650000803f00000040000040400000ffff1400000000000000"
					}
				}
				before := *original
				converted := p.ConvertFromLatest(original, nil)
				if p.ID() == 419 {
					// Protocol 419 predates this packet. Updating its shared codec
					// must not accidentally advertise the later packet ID.
					if len(converted) != 0 || p.Packets(false)[original.ID()] != nil {
						t.Fatal("protocol 419 advertised an unsupported debug packet")
					}
					continue
				}
				if len(converted) != 1 {
					t.Fatalf("debug packet produced %d packets", len(converted))
				}
				var encoded bytes.Buffer
				converted[0].Marshal(p.NewWriter(&encoded, -1))
				if got := hex.EncodeToString(encoded.Bytes()); got != wantHex {
					t.Fatalf("cube=%v wire=%s, want %s", cube, got, wantHex)
				}
				if !reflect.DeepEqual(*original, before) {
					t.Fatal("encoding changed the native input")
				}
				decoded := p.Packets(false)[converted[0].ID()]()
				decoded.Marshal(p.NewReader(&encoded, -1, true))
				if encoded.Len() != 0 {
					t.Fatalf("decoder left %d bytes", encoded.Len())
				}
				latest := p.ConvertToLatest(decoded, nil)
				if len(latest) != 1 || !reflect.DeepEqual(latest[0], original) {
					t.Fatalf("cube=%v decoded=%#v, want %#v", cube, latest, original)
				}
			}
		})
	}
}

func TestLegacyDebugRendererFloatColoursAndInvalidInput(t *testing.T) {
	for _, p := range debugRendererProtocols()[1:7] {
		t.Run(p.Ver(), func(t *testing.T) {
			fixture, err := hex.DecodeString(legacyDebugCube)
			if err != nil {
				t.Fatal(err)
			}
			// Historical float channels map to the nearest native byte value.
			for i, f := range []float32{0.5, 0.25, 0.125, 0.75} {
				binary.LittleEndian.PutUint32(fixture[21+i*4:], math.Float32bits(f))
			}
			decoded := p.Packets(false)[packet.IDClientBoundDebugRenderer]()
			decoded.Marshal(p.NewReader(bytes.NewBuffer(fixture), -1, true))
			latest := p.ConvertToLatest(decoded, nil)[0].(*packet.ClientBoundDebugRenderer)
			data, present := latest.Data.Value()
			if !present || data.Colour != (color.RGBA{R: 128, G: 64, B: 32, A: 191}) {
				t.Fatalf("incorrect native colour: %+v", data.Colour)
			}
			// Decoding clear into a reused packet must remove the prior marker.
			decoded.Marshal(p.NewReader(bytes.NewBuffer([]byte{1, 0, 0, 0}), -1, true))
			cleared := p.ConvertToLatest(decoded, nil)[0].(*packet.ClientBoundDebugRenderer)
			if _, ok := cleared.Data.Value(); ok || cleared.Type != packet.ClientBoundDebugRendererClear {
				t.Fatal("clear retained marker data")
			}
			for _, invalid := range []struct {
				name string
				data []byte
			}{
				{"unknown_action", []byte{3, 0, 0, 0}},
				{"missing_marker", []byte{2, 0, 0, 0}},
				{"nan_colour", func() []byte {
					b := bytes.Clone(fixture)
					binary.LittleEndian.PutUint32(b[21:], math.Float32bits(float32(math.NaN())))
					return b
				}()},
			} {
				t.Run(invalid.name, func(t *testing.T) {
					defer func() {
						if recover() == nil {
							t.Fatal("invalid legacy packet was accepted")
						}
					}()
					pk := p.Packets(false)[packet.IDClientBoundDebugRenderer]()
					pk.Marshal(p.NewReader(bytes.NewBuffer(invalid.data), -1, true))
				})
			}
		})
	}
}
