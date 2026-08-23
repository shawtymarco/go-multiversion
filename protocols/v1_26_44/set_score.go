package v1_26_44

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// setScore uses the 1.26.44 scoreboard entry layout. In this release, a
// removed entry's objective name is encoded as an optional nested inside an
// always-present outer optional.
type setScore struct {
	Entries []scoreboardEntry
}

func (*setScore) ID() uint32 {
	return packet.IDSetScore
}

func (pk *setScore) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Entries)
}

func (pk *setScore) latest() *packet.SetScore {
	entries := make([]protocol.ScoreboardEntry, len(pk.Entries))
	for i, entry := range pk.Entries {
		entries[i] = protocol.ScoreboardEntry(entry)
	}
	return &packet.SetScore{Entries: entries}
}

func setScoreFromLatest(pk *packet.SetScore) *setScore {
	entries := make([]scoreboardEntry, len(pk.Entries))
	for i, entry := range pk.Entries {
		entries[i] = scoreboardEntry(entry)
	}
	return &setScore{Entries: entries}
}

type scoreboardEntry protocol.ScoreboardEntry

func (entry *scoreboardEntry) Marshal(io protocol.IO) {
	variant := uint32(entry.IdentityType)
	io.Varuint32(&variant)
	entry.IdentityType = byte(variant)

	typeNames := [...]string{"remove", "changeplayer", "changeentity", "changefakeplayer"}
	if variant >= uint32(len(typeNames)) {
		io.UnknownEnumOption(variant, "scoreboard entry variant")
		return
	}
	typeName := typeNames[variant]
	io.String(&typeName)
	io.Varint64(&entry.EntryID)

	switch entry.IdentityType {
	case protocol.ScoreboardIdentityRemove:
		objective := protocol.Optional[string]{}
		if entry.ObjectiveName != "" {
			objective = protocol.Option(entry.ObjectiveName)
		}
		protocol.DoubleOptionalFunc(io, &objective, io.String)
		entry.ObjectiveName, _ = objective.Value()
	case protocol.ScoreboardIdentityEntity, protocol.ScoreboardIdentityPlayer:
		io.String(&entry.ObjectiveName)
		io.Int32(&entry.Score)
		io.ActorUniqueID(&entry.EntityUniqueID)
	case protocol.ScoreboardIdentityFakePlayer:
		io.String(&entry.ObjectiveName)
		io.Int32(&entry.Score)
		io.String(&entry.DisplayName)
	}
}
