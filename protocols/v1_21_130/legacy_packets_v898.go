package v1_21_130

import (
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalText(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.Text)
	io.Bool(&pk.NeedsTranslation)
	var categoryType uint8
	if !io.reading {
		switch pk.TextType {
		case packet.TextTypeRaw, packet.TextTypeTip, packet.TextTypeSystem, packet.TextTypeObjectWhisper, packet.TextTypeObjectAnnouncement, packet.TextTypeObject:
			categoryType = packet.TextCategoryMessageOnly
		case packet.TextTypeChat, packet.TextTypeWhisper, packet.TextTypeAnnouncement:
			categoryType = packet.TextCategoryAuthoredMessage
		default:
			categoryType = packet.TextCategoryMessageWithParameters
		}
	}
	io.Uint8(&categoryType)
	switch categoryType {
	case packet.TextCategoryMessageOnly:
		for _, value := range []string{"raw", "tip", "systemMessage", "textObjectWhisper", "textObjectAnnouncement", "textObject"} {
			marshalStringConst898(io, value)
		}
	case packet.TextCategoryAuthoredMessage:
		for _, value := range []string{"chat", "whisper", "announcement"} {
			marshalStringConst898(io, value)
		}
	default:
		for _, value := range []string{"translate", "popup", "jukeboxPopup"} {
			marshalStringConst898(io, value)
		}
	}
	io.Uint8(&pk.TextType)
	switch pk.TextType {
	case packet.TextTypeChat, packet.TextTypeWhisper, packet.TextTypeAnnouncement:
		io.String(&pk.SourceName)
		io.String(&pk.Message)
	case packet.TextTypeRaw, packet.TextTypeTip, packet.TextTypeSystem, packet.TextTypeObject, packet.TextTypeObjectWhisper, packet.TextTypeObjectAnnouncement:
		io.String(&pk.Message)
	case packet.TextTypeTranslation, packet.TextTypePopup, packet.TextTypeJukeboxPopup:
		io.String(&pk.Message)
		protocol.FuncSlice(io.directional(), &pk.Parameters, io.String)
	}
	if len(pk.Message) == 0 {
		io.InvalidValue(pk.Message, "message", "string cannot be empty")
		return
	}
	io.String(&pk.XUID)
	io.String(&pk.PlatformChatID)
	protocol.OptionalFunc(io.directional(), &pk.FilteredMessage, io.String)
}

func marshalStringConst898(io *wireIO, expected string) {
	value := expected
	io.String(&value)
	if io.reading && !strings.EqualFold(value, expected) {
		io.InvalidValue(value, "text category string", "unexpected constant")
	}
}

func marshalBookEdit(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.BookEdit)
	actionType := uint8(pk.ActionType)
	if !io.reading && pk.ActionType > 255 {
		io.InvalidValue(pk.ActionType, "book edit action type", "exceeds uint8")
		return
	}
	inventorySlot := uint8(pk.InventorySlot)
	if !io.reading && (pk.InventorySlot < 0 || pk.InventorySlot > 255) {
		io.InvalidValue(pk.InventorySlot, "book inventory slot", "outside uint8")
		return
	}
	io.Uint8(&actionType)
	io.Uint8(&inventorySlot)
	pk.ActionType, pk.InventorySlot = uint32(actionType), int32(inventorySlot)
	switch pk.ActionType {
	case packet.BookActionReplacePage, packet.BookActionAddPage:
		marshalBookPageNumber898(io, &pk.PageNumber)
		io.String(&pk.Text)
		io.String(&pk.PhotoName)
	case packet.BookActionDeletePage:
		marshalBookPageNumber898(io, &pk.PageNumber)
	case packet.BookActionSwapPages:
		marshalBookPageNumber898(io, &pk.PageNumber)
		marshalBookPageNumber898(io, &pk.SecondaryPageNumber)
	case packet.BookActionSign:
		io.String(&pk.Title)
		io.String(&pk.Author)
		io.String(&pk.XUID)
	default:
		io.UnknownEnumOption(pk.ActionType, "book edit action type")
	}
}

func marshalBookPageNumber898(io *wireIO, page *int32) {
	if !io.reading && (*page < 0 || *page > 255) {
		io.InvalidValue(*page, "book page number", "outside uint8")
		return
	}
	value := uint8(*page)
	io.Uint8(&value)
	*page = int32(value)
}
