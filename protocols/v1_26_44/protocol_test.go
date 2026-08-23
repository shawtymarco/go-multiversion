package v1_26_44

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestSetScoreRoundTrip(t *testing.T) {
	want := &packet.SetScore{Entries: []protocol.ScoreboardEntry{
		{
			EntryID:       42,
			ObjectiveName: "sidebar",
			IdentityType:  protocol.ScoreboardIdentityRemove,
		},
		{
			EntryID:       43,
			ObjectiveName: "sidebar",
			Score:         7,
			IdentityType:  protocol.ScoreboardIdentityFakePlayer,
			DisplayName:   "line",
		},
	}}

	legacy := Protocol{}.ConvertFromLatest(want, nil)[0].(*setScore)
	var payload bytes.Buffer
	legacy.Marshal(Protocol{}.NewWriter(&payload, 0))

	decoded := &setScore{}
	decoded.Marshal(Protocol{}.NewReader(bytes.NewBuffer(payload.Bytes()), 0, true))
	got := Protocol{}.ConvertToLatest(decoded, nil)[0].(*packet.SetScore)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch:\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestSetScoreUsesLegacyDoubleOptional(t *testing.T) {
	latest := &packet.SetScore{Entries: []protocol.ScoreboardEntry{{
		EntryID:       1,
		ObjectiveName: "objective",
		IdentityType:  protocol.ScoreboardIdentityRemove,
	}}}

	var latestPayload bytes.Buffer
	latest.Marshal(protocol.NewWriter(&latestPayload, 0))

	legacy := Protocol{}.ConvertFromLatest(latest, nil)[0]
	var legacyPayload bytes.Buffer
	legacy.Marshal(Protocol{}.NewWriter(&legacyPayload, 0))

	if bytes.Equal(legacyPayload.Bytes(), latestPayload.Bytes()) {
		t.Fatal("1.26.44 and 1.26.45 SetScore payloads unexpectedly match")
	}
	if got, want := legacyPayload.Len(), latestPayload.Len()+1; got != want {
		t.Fatalf("legacy payload length: got %d, want %d", got, want)
	}
}

func TestSetScoreLayoutByGameVersion(t *testing.T) {
	tests := []struct {
		gameVersion string
		wantLegacy  bool
	}{
		{gameVersion: "1.26.40"},
		{gameVersion: "1.26.41"},
		{gameVersion: "1.26.42"},
		{gameVersion: "1.26.43"},
		{gameVersion: "1.26.44", wantLegacy: true},
		{gameVersion: "", wantLegacy: true},
		{gameVersion: "1.26.49", wantLegacy: true},
	}

	for _, test := range tests {
		t.Run(test.gameVersion, func(t *testing.T) {
			latest := &packet.SetScore{Entries: []protocol.ScoreboardEntry{{
				EntryID:       1,
				ObjectiveName: "objective",
				IdentityType:  protocol.ScoreboardIdentityRemove,
			}}}
			want := &packet.SetScore{Entries: append([]protocol.ScoreboardEntry(nil), latest.Entries...)}

			converted := convertFromLatestForGameVersion(latest, test.gameVersion)[0]
			_, legacy := converted.(*setScore)
			if legacy != test.wantLegacy {
				t.Fatalf("legacy SetScore layout: got %t, want %t", legacy, test.wantLegacy)
			}
			if !test.wantLegacy && converted != latest {
				t.Fatal("unchanged SetScore layout returned a new packet")
			}
			if !reflect.DeepEqual(latest, want) {
				t.Fatalf("conversion mutated input:\ngot:  %#v\nwant: %#v", latest, want)
			}
		})
	}
}

func TestSetScoreOraclePayloads(t *testing.T) {
	// The single-optional payloads match gophertunnel 0695275 (1.26.40-1.26.43).
	// The double-optional payloads match gophertunnel 8a2b1f7 (1.26.44).
	tests := []struct {
		name          string
		gameVersion   string
		objectiveName string
		wantHex       string
	}{
		{name: "1.26.40 objective", gameVersion: "1.26.40", objectiveName: "objective", wantHex: "01000672656d6f76650201096f626a656374697665"},
		{name: "1.26.41 objective", gameVersion: "1.26.41", objectiveName: "objective", wantHex: "01000672656d6f76650201096f626a656374697665"},
		{name: "1.26.42 objective", gameVersion: "1.26.42", objectiveName: "objective", wantHex: "01000672656d6f76650201096f626a656374697665"},
		{name: "1.26.43 objective", gameVersion: "1.26.43", objectiveName: "objective", wantHex: "01000672656d6f76650201096f626a656374697665"},
		{name: "1.26.44 objective", gameVersion: "1.26.44", objectiveName: "objective", wantHex: "01000672656d6f7665020101096f626a656374697665"},
		{name: "1.26.40 empty", gameVersion: "1.26.40", wantHex: "01000672656d6f76650200"},
		{name: "1.26.44 empty", gameVersion: "1.26.44", wantHex: "01000672656d6f7665020100"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			latest := &packet.SetScore{Entries: []protocol.ScoreboardEntry{{
				EntryID:       1,
				ObjectiveName: test.objectiveName,
				IdentityType:  protocol.ScoreboardIdentityRemove,
			}}}
			converted := convertFromLatestForGameVersion(latest, test.gameVersion)[0]

			var payload bytes.Buffer
			converted.Marshal(Protocol{}.NewWriter(&payload, 0))
			if got := hex.EncodeToString(payload.Bytes()); got != test.wantHex {
				t.Fatalf("payload: got %s, want %s", got, test.wantHex)
			}
		})
	}
}

func TestServerPacketPoolUsesLegacySetScore(t *testing.T) {
	constructor, ok := Protocol{}.Packets(false)[packet.IDSetScore]
	if !ok {
		t.Fatal("SetScore missing from server packet pool")
	}
	if _, ok := constructor().(*setScore); !ok {
		t.Fatalf("unexpected SetScore packet type: %T", constructor())
	}
}
