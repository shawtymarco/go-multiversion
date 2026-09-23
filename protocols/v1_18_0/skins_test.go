package v1_18_0

import (
	"image/color"
	"testing"
)

func TestParseLegacyColourShorthandZero(t *testing.T) {
	if got := parseLegacyColour(nil, "#0"); got != (color.RGBA{}) {
		t.Fatalf("parseLegacyColour(#0) = %#v", got)
	}
}
