package v1_21_130

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
	case *packet.ServerStoreInfo, *packet.ServerPresenceInfo,
		*packet.ClientBoundDataStore, *packet.ServerBoundDataStore,
		*packet.ResourcePacksReadyForValidation, *packet.LocatorBar,
		*packet.PartyChanged, *packet.ServerBoundDataDrivenScreenClosed,
		*packet.SyncWorldClocks, *packet.ClientBoundAttributeLayerSync:
		return nil
	case *packet.ClientBoundDataDrivenUIShowScreen, *packet.ClientBoundDataDrivenUICloseScreen,
		*packet.ClientBoundDataDrivenUIReload, *packet.ClientBoundTextureShift,
		*packet.VoxelShapes, *packet.CameraSpline, *packet.CameraAimAssistActorPriority:
		return nil
	case *packet.ActorEvent:
		if current.EventType >= packet.ActorEventHurtWithoutReceivingDamage {
			return nil
		}
	case *packet.GraphicsOverrideParameter:
		if current.ParameterType >= protocol.GraphicsOverrideParameterTypeChlorophyll {
			return nil
		}
	case *packet.UpdateClientInputLocks:
		cloned := *current
		cloned.Locks &= packet.ClientInputLockCamera | packet.ClientInputLockMovement
		pk = &cloned
	case *packet.LevelSoundEvent:
		if _, ok := legacySoundToID[current.SoundType]; !ok {
			return nil
		}
	case *packet.MoveActorAbsolute:
		cloned := *current
		cloned.Flags &= packet.MoveFlagOnGround | packet.MoveFlagTeleport
		pk = &cloned
	case *packet.Disconnect:
		if current.Reason >= packet.DisconnectReasonHostSignedOut {
			cloned := *current
			cloned.Reason = packet.DisconnectReasonNoReason
			pk = &cloned
		}
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
	case "1.21.130", "1.21.131", "1.21.132":
		return true
	default:
		return false
	}
}
