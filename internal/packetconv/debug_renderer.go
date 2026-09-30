package packetconv

import (
	"image/color"
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LegacyDebugRenderer bridges the integer-action/float-colour layout to the
// native optional DebugMarkerData model introduced by gophertunnel 98c3d1f.
// Historical sources: c40bf828/2cb1e399 use signed durations; 268adeb5/bf05a1a
// use unsigned durations. Both use actions 1 (clear) and 2 (add), no optional
// presence byte, and four float32 colour channels. Writers never mutate pk.
func LegacyDebugRenderer(io protocol.IO, pk *packet.ClientBoundDebugRenderer, reading, signedDuration bool) {
	action := pk.Type + 1
	io.Uint32(&action)
	if action < 1 || action > 2 {
		io.UnknownEnumOption(action, "client bound debug renderer type")
		return
	}
	if action == 1 {
		if reading {
			pk.Type = packet.ClientBoundDebugRendererClear
			pk.Data = protocol.Optional[packet.DebugMarkerData]{}
		}
		return
	}
	data, present := pk.Data.Value()
	if !reading && !present {
		io.InvalidValue(present, "debug marker data", "required for legacy add-cube action")
		return
	}
	io.String(&data.Text)
	io.Vec3(&data.Position)
	channels := [4]float32{float32(data.Colour.R) / 255, float32(data.Colour.G) / 255, float32(data.Colour.B) / 255, float32(data.Colour.A) / 255}
	for i := range channels {
		io.Float32(&channels[i])
		if reading && (math.IsNaN(float64(channels[i])) || channels[i] < 0 || channels[i] > 1) {
			io.InvalidValue(channels[i], "debug marker colour", "must be finite and between zero and one")
			return
		}
	}
	if signedDuration {
		if !reading && data.Duration > math.MaxInt64 {
			io.InvalidValue(data.Duration, "debug marker duration", "exceeds legacy signed duration")
			return
		}
		duration := int64(data.Duration)
		io.Int64(&duration)
		if duration < 0 {
			io.InvalidValue(duration, "debug marker duration", "must not be negative")
			return
		}
		data.Duration = uint64(duration)
	} else {
		io.Uint64(&data.Duration)
	}
	if reading {
		// Native colours have eight bits per channel, so incoming historical
		// float colours are rounded to the nearest representable native value.
		data.Colour = color.RGBA{R: uint8(math.Round(float64(channels[0]) * 255)), G: uint8(math.Round(float64(channels[1]) * 255)),
			B: uint8(math.Round(float64(channels[2]) * 255)), A: uint8(math.Round(float64(channels[3]) * 255))}
		pk.Type = packet.ClientBoundDebugRendererAddCube
		pk.Data = protocol.Option(data)
	}
}
