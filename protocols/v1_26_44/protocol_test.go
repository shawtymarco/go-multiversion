package v1_26_44

import (
	"bytes"
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

func TestServerPacketPoolUsesLegacySetScore(t *testing.T) {
	constructor, ok := Protocol{}.Packets(false)[packet.IDSetScore]
	if !ok {
		t.Fatal("SetScore missing from server packet pool")
	}
	if _, ok := constructor().(*setScore); !ok {
		t.Fatalf("unexpected SetScore packet type: %T", constructor())
	}
}
