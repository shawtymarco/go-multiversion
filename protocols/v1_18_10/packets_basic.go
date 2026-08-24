package v1_18_10

import (
	"bytes"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalAnimate(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Animate)
	action := int32(pk.ActionType)
	io.Varint32(&action)
	pk.ActionType = uint8(action)
	io.Varuint64(&pk.EntityRuntimeID)
	if pk.ActionType&0x80 != 0 {
		io.Float32(&pk.Data)
	} else if io.reading {
		pk.Data = 0
	}
	if io.reading {
		pk.SwingSource = 0
	}
}

func marshalChangeDimension(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ChangeDimension)
	io.Varint32(&pk.Dimension)
	io.Vec3(&pk.Position)
	io.Bool(&pk.Respawn)
	if io.reading {
		pk.LoadingScreenID = protocol.Optional[uint32]{}
	}
}

func marshalCommandRequest(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.CommandRequest)
	io.String(&pk.CommandLine)
	marshalCommandOrigin(io, &pk.CommandOrigin)
	io.Bool(&pk.Internal)
	if io.reading {
		pk.Version = ""
	}
}

func marshalCommandOrigin(io *wireIO, origin *protocol.CommandOrigin) {
	io.Varuint32(&origin.Origin)
	io.UUID(&origin.UUID)
	io.String(&origin.RequestID)
	if origin.Origin == protocol.CommandOriginDevConsole || origin.Origin == protocol.CommandOriginTest {
		io.Varint64(&origin.PlayerUniqueID)
	} else if io.reading {
		origin.PlayerUniqueID = 0
	}
}

func marshalContainerClose(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ContainerClose)
	io.Uint8(&pk.WindowID)
	io.Bool(&pk.ServerSide)
	if io.reading {
		pk.ContainerType = 0
	}
}

func marshalDisconnect(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Disconnect)
	io.Bool(&pk.HideDisconnectionScreen)
	if !pk.HideDisconnectionScreen {
		io.String(&pk.Message)
	}
	if io.reading {
		pk.Reason = packet.DisconnectReasonUnknown
		pk.FilteredMessage = ""
	}
}

func marshalEmote(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Emote)
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.EmoteID)
	io.Uint8(&pk.Flags)
	if io.reading {
		pk.EmoteLength = 0
		pk.XUID = ""
		pk.PlatformID = ""
	}
}

func marshalInventoryContent(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventoryContent)
	io.Varuint32(&pk.WindowID)
	protocol.FuncSlice(io.directional(), &pk.Content, io.ItemInstance)
	if io.reading {
		pk.Container = protocol.FullContainerName{}
		pk.StorageItem = protocol.ItemInstance{}
	}
}

func marshalInventorySlot(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.InventorySlot)
	io.Varuint32(&pk.WindowID)
	io.Varuint32(&pk.Slot)
	io.ItemInstance(&pk.NewItem)
	if io.reading {
		pk.Container = protocol.Optional[protocol.FullContainerName]{}
		pk.StorageItem = protocol.Optional[protocol.ItemInstance]{}
	}
}

func marshalModalFormResponse(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ModalFormResponse)
	io.Varuint32(&pk.FormID)
	if io.reading {
		var response []byte
		io.ByteSlice(&response)
		if bytes.Equal(bytes.TrimSpace(response), []byte("null")) {
			pk.ResponseData = protocol.Optional[[]byte]{}
			pk.CancelReason = protocol.Option(uint8(packet.ModalFormCancelReasonUserClosed))
		} else {
			pk.ResponseData = protocol.Option(response)
			pk.CancelReason = protocol.Optional[uint8]{}
		}
		return
	}
	response, ok := pk.ResponseData.Value()
	if !ok {
		if _, cancelled := pk.CancelReason.Value(); cancelled {
			response = []byte("null\n")
		}
	}
	io.ByteSlice(&response)
}

func marshalPlayerAction(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerAction)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Varint32(&pk.ActionType)
	marshalUnsignedBlockPos(io, &pk.BlockPosition)
	io.Varint32(&pk.BlockFace)
	if io.reading {
		pk.ResultPosition = pk.BlockPosition
	}
}

func marshalUnsignedBlockPos(io *wireIO, position *protocol.BlockPos) {
	io.Varint32(&position[0])
	y := uint32(position[1])
	io.Varuint32(&y)
	position[1] = int32(y)
	io.Varint32(&position[2])
}

func marshalPlayerArmourDamage(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerArmourDamage)
	var flags uint8
	var damage [4]int32
	if !io.reading {
		for _, entry := range pk.List {
			if entry.ArmourSlot < 0 || entry.ArmourSlot >= int32(len(damage)) {
				continue
			}
			flags |= 1 << uint32(entry.ArmourSlot)
			damage[entry.ArmourSlot] = int32(entry.Damage)
		}
	}
	io.Uint8(&flags)
	for index := range damage {
		if flags&(1<<uint32(index)) != 0 {
			io.Varint32(&damage[index])
		}
	}
	if io.reading {
		pk.List = make([]protocol.PlayerArmourDamageEntry, 0, 4)
		for index, value := range damage {
			if flags&(1<<uint32(index)) != 0 {
				pk.List = append(pk.List, protocol.PlayerArmourDamageEntry{ArmourSlot: int32(index), Damage: int16(value)})
			}
		}
	}
}

func marshalRequestChunkRadius(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.RequestChunkRadius)
	io.Varint32(&pk.ChunkRadius)
	if io.reading {
		pk.MaxChunkRadius = uint8(pk.ChunkRadius)
	}
}

func marshalSetTitle(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.SetTitle)
	io.Varint32(&pk.ActionType)
	io.String(&pk.Text)
	io.Varint32(&pk.FadeInDuration)
	io.Varint32(&pk.RemainDuration)
	io.Varint32(&pk.FadeOutDuration)
	io.String(&pk.XUID)
	io.String(&pk.PlatformOnlineID)
	if io.reading {
		pk.FilteredMessage = ""
	}
}

func marshalStopSound(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.StopSound)
	io.String(&pk.SoundName)
	io.Bool(&pk.StopAll)
	if io.reading {
		pk.StopMusicLegacy = false
	}
}

func marshalText(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Text)
	io.Uint8(&pk.TextType)
	io.Bool(&pk.NeedsTranslation)
	switch pk.TextType {
	case packet.TextTypeChat, packet.TextTypeWhisper, packet.TextTypeAnnouncement:
		io.String(&pk.SourceName)
		io.String(&pk.Message)
	case packet.TextTypeRaw, packet.TextTypeTip, packet.TextTypeSystem, packet.TextTypeObject, packet.TextTypeObjectWhisper:
		io.String(&pk.Message)
	case packet.TextTypeTranslation, packet.TextTypePopup, packet.TextTypeJukeboxPopup:
		io.String(&pk.Message)
		protocol.FuncSlice(io.directional(), &pk.Parameters, io.String)
	default:
		io.UnknownEnumOption(pk.TextType, "text type")
		return
	}
	io.String(&pk.XUID)
	io.String(&pk.PlatformChatID)
	if io.reading {
		pk.FilteredMessage = protocol.Optional[string]{}
	}
}
