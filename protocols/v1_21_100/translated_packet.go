package v1_21_100

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetMarshal func(*wireIO, packet.Packet)

type translatedPacket struct {
	inner   packet.Packet
	marshal packetMarshal
}

func (pk *translatedPacket) ID() uint32 { return pk.inner.ID() }

func (pk *translatedPacket) Marshal(io protocol.IO) {
	pk.marshal(asWireIO(io), pk.inner)
}

func translatedConstructor(constructor func() packet.Packet, marshal packetMarshal) func() packet.Packet {
	return func() packet.Packet {
		return &translatedPacket{inner: constructor(), marshal: marshal}
	}
}

func translated(pk packet.Packet, marshal packetMarshal) packet.Packet {
	return &translatedPacket{inner: pk, marshal: marshal}
}

func downgradePacket(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch current := pk.(type) {
	case *packet.PlayerList:
		var additions, removals []protocol.PlayerListEntry
		for _, entry := range current.Entries {
			if entry.ActionType == protocol.PlayerListActionRemove {
				removals = append(removals, entry)
			} else {
				additions = append(additions, entry)
			}
		}
		if len(additions) != 0 && len(removals) != 0 {
			return append(
				downgradePacket(&packet.PlayerList{Entries: additions}, conn),
				downgradePacket(&packet.PlayerList{Entries: removals}, conn)...,
			)
		}
	case *packet.SetScore:
		var modifications, removals []protocol.ScoreboardEntry
		for _, entry := range current.Entries {
			if entry.IdentityType == protocol.ScoreboardIdentityRemove {
				removals = append(removals, entry)
			} else {
				modifications = append(modifications, entry)
			}
		}
		if len(modifications) != 0 && len(removals) != 0 {
			return append(
				downgradePacket(&packet.SetScore{Entries: modifications}, conn),
				downgradePacket(&packet.SetScore{Entries: removals}, conn)...,
			)
		}
	case *packet.ClientboundUpdateSoundData:
		if _, ok := current.Stop.Value(); !ok {
			return nil
		}
	case *packet.StartGame:
		cloned := *current
		version := Version
		if conn != nil && isStableGameVersion(conn.ClientData().GameVersion) {
			version = conn.ClientData().GameVersion
		}
		cloned.GameVersion, cloned.BaseGameVersion = version, version
		pk = &cloned
	}
	marshal, ok := packetMarshals[pk.ID()]
	if !ok {
		return []packet.Packet{pk}
	}
	return []packet.Packet{translated(pk, marshal)}
}

func isStableGameVersion(version string) bool {
	switch version {
	case "1.21.100", "1.21.101", "1.21.102":
		return true
	default:
		return false
	}
}
