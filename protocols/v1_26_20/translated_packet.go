package v1_26_20

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
	case *packet.ClientboundUpdateSoundData, *packet.SendPartyDestinationCookie:
		return nil
	case *packet.MovementEffect:
		if current.Type != packet.MovementEffectTypeGlideBoost {
			return nil
		}
	case *packet.PrimitiveShapes:
		cloned := *current
		cloned.Shapes = make([]protocol.PrimitiveShape, 0, len(current.Shapes))
		for _, shape := range current.Shapes {
			shapeType, hasType := shape.Type.Value()
			if hasType && shapeType > protocol.PrimitiveShapeArrow {
				continue
			}
			if !isLegacyShapeData(shape.ExtraShapeData) {
				continue
			}
			cloned.Shapes = append(cloned.Shapes, shape)
		}
		pk = &cloned
	case *packet.ClientBoundDataStore:
		cloned := *current
		cloned.Updates = make([]protocol.DataStoreChangeEntry, 0, len(current.Updates))
		for _, entry := range current.Updates {
			if entry.ChangeType == protocol.DataStoreChangeTypeChange && !isLegacyDataStoreValue(entry.Change.NewValue) {
				continue
			}
			cloned.Updates = append(cloned.Updates, entry)
		}
		pk = &cloned
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

func isLegacyShapeData(value protocol.ShapeData) bool {
	switch value.(type) {
	case *protocol.LastShape, *protocol.ArrowShape, *protocol.TextShape, *protocol.BoxShape, *protocol.LineShape, *protocol.SphereShape:
		return true
	default:
		return false
	}
}

func isLegacyDataStoreValue(value protocol.DataStorePropertyValue) bool {
	switch value.Type {
	case protocol.DataStorePropertyTypeNone, protocol.DataStorePropertyTypeBool, protocol.DataStorePropertyTypeInt64, protocol.DataStorePropertyTypeString:
		return true
	case protocol.DataStorePropertyTypeMap:
		for _, entry := range value.MapValue {
			if !isLegacyDataStoreValue(entry.Value) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isStableGameVersion(version string) bool {
	switch version {
	case "1.26.20", "1.26.21", "1.26.23":
		return true
	default:
		return false
	}
}
