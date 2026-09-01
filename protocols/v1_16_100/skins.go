package v1_16_100

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalPlayerList(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerList)
	action := uint8(protocol.PlayerListActionAdd)
	if !io.reading && len(pk.Entries) != 0 {
		action = pk.Entries[0].ActionType
	}
	io.Uint8(&action)
	count := uint32(len(pk.Entries))
	io.Varuint32(&count)
	if io.reading {
		pk.Entries = make([]protocol.PlayerListEntry, count)
	}
	for index := range pk.Entries {
		entry := &pk.Entries[index]
		entry.ActionType = action
		switch action {
		case protocol.PlayerListActionAdd:
			marshalPlayerListAddEntry(io, entry)
		case protocol.PlayerListActionRemove:
			io.UUID(&entry.UUID)
		default:
			io.UnknownEnumOption(action, "player list action type")
			return
		}
	}
	if action == protocol.PlayerListActionAdd {
		for index := range pk.Entries {
			io.Bool(&pk.Entries[index].Skin.Trusted)
		}
	}
}

func marshalPlayerListAddEntry(io *wireIO, entry *protocol.PlayerListEntry) {
	io.UUID(&entry.UUID)
	io.Varint64(&entry.EntityUniqueID)
	io.String(&entry.Username)
	io.String(&entry.XUID)
	io.String(&entry.PlatformChatID)
	io.Int32(&entry.BuildPlatform)
	marshalSkin(io, &entry.Skin)
	io.Bool(&entry.Teacher)
	io.Bool(&entry.Host)
	if io.reading {
		entry.SubClient = false
		entry.PlayerColour = color.RGBA{}
	}
}

func marshalPlayerSkin(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.PlayerSkin)
	io.UUID(&pk.UUID)
	marshalSkin(io, &pk.Skin)
	io.String(&pk.NewSkinName)
	io.String(&pk.OldSkinName)
	io.Bool(&pk.Skin.Trusted)
}

func marshalSkin(io *wireIO, skin *protocol.Skin) {
	io.String(&skin.SkinID)
	io.ByteSlice(&skin.SkinResourcePatch)
	io.Uint32(&skin.SkinImageWidth)
	io.Uint32(&skin.SkinImageHeight)
	io.ByteSlice(&skin.SkinData)
	funcIOSliceUint32Length(io, &skin.Animations, marshalSkinAnimation)
	io.Uint32(&skin.CapeImageWidth)
	io.Uint32(&skin.CapeImageHeight)
	io.ByteSlice(&skin.CapeData)
	io.ByteSlice(&skin.SkinGeometry)
	io.ByteSlice(&skin.AnimationData)
	io.Bool(&skin.PremiumSkin)
	io.Bool(&skin.PersonaSkin)
	io.Bool(&skin.PersonaCapeOnClassicSkin)
	io.String(&skin.CapeID)
	io.String(&skin.FullID)
	armSize := ""
	if !emptyLegacySkin(*skin) {
		armSize = legacyArmSize(skin.ArmSize)
	}
	io.String(&armSize)
	if io.reading {
		skin.ArmSize = nativeArmSize(armSize)
	}
	skinColour := ""
	if !emptyLegacySkin(*skin) {
		skinColour = legacyRGB(skin.SkinColour)
	}
	io.String(&skinColour)
	if io.reading {
		skin.SkinColour = parseLegacyColour(io, skinColour)
	}
	funcIOSliceUint32Length(io, &skin.PersonaPieces, marshalPersonaPiece)
	funcIOSliceUint32Length(io, &skin.PieceTintColours, marshalPieceTint)
	if io.reading {
		skin.PlayFabID = ""
		skin.GeometryDataEngineVersion = nil
		skin.PrimaryUser = false
		skin.OverrideAppearance = true
		skin.ProfileHash = ""
	}
}

func emptyLegacySkin(skin protocol.Skin) bool {
	return skin.SkinID == "" && skin.PlayFabID == "" && len(skin.SkinData) == 0 && len(skin.CapeData) == 0 &&
		len(skin.SkinGeometry) == 0 && len(skin.PersonaPieces) == 0 && len(skin.PieceTintColours) == 0
}

func funcIOSliceUint32Length[T any](io *wireIO, values *[]T, marshal func(protocol.IO, *T)) {
	count := uint32(len(*values))
	io.Uint32(&count)
	if io.reading {
		*values = make([]T, count)
	}
	for index := range *values {
		marshal(io.directional(), &(*values)[index])
	}
}

func marshalSkinAnimation(raw protocol.IO, animation *protocol.SkinAnimation) {
	io := asWireIO(raw)
	io.Uint32(&animation.ImageWidth)
	io.Uint32(&animation.ImageHeight)
	io.ByteSlice(&animation.ImageData)
	io.Uint32(&animation.AnimationType)
	io.Float32(&animation.FrameCount)
	io.Uint32(&animation.ExpressionType)
}

func marshalPersonaPiece(raw protocol.IO, piece *protocol.PersonaPiece) {
	io := asWireIO(raw)
	io.String(&piece.PieceID)
	pieceType := legacyPieceType(piece.PieceType)
	io.String(&pieceType)
	if io.reading {
		piece.PieceType = nativePieceType(pieceType)
	}
	packID := piece.PackID.String()
	io.String(&packID)
	if io.reading {
		parsed, err := uuid.Parse(packID)
		if err != nil {
			io.InvalidValue(packID, "persona piece pack ID", err.Error())
			return
		}
		piece.PackID = parsed
	}
	io.Bool(&piece.Default)
	io.String(&piece.ProductID)
}

func marshalPieceTint(raw protocol.IO, tint *protocol.PersonaPieceTintColour) {
	io := asWireIO(raw)
	io.String(&tint.PieceType)
	count := uint32(len(tint.Colours))
	io.Uint32(&count)
	if count != uint32(len(tint.Colours)) {
		io.InvalidValue(count, "persona tint colour count", "must be four")
		return
	}
	for index := uint32(0); index < count; index++ {
		value := legacyARGB(tint.Colours[index])
		io.String(&value)
		if io.reading {
			tint.Colours[index] = parseLegacyColour(io, value)
		}
	}
}

func legacyArmSize(value uint8) string {
	if value == protocol.ArmSizeSlim {
		return "slim"
	}
	return "wide"
}

func nativeArmSize(value string) uint8 {
	if strings.EqualFold(value, "slim") {
		return protocol.ArmSizeSlim
	}
	return protocol.ArmSizeWide
}

var pieceTypes = []string{
	"unsupported", "persona_skeleton", "persona_body", "persona_skin", "persona_bottom", "persona_feet",
	"persona_dress", "persona_top", "persona_high_pants", "persona_hands", "persona_outerwear", "persona_facial_hair",
	"persona_mouth", "persona_eyes", "persona_hair", "persona_hood", "persona_back", "persona_face_accessory",
	"persona_head", "persona_legs", "persona_left_leg", "persona_right_leg", "persona_arms", "persona_left_arm",
	"persona_right_arm", "persona_capes", "persona_classic_skin", "persona_emote", "unsupported",
}

func legacyPieceType(value uint32) string {
	if int(value) >= len(pieceTypes) {
		return "unsupported"
	}
	return pieceTypes[value]
}

func nativePieceType(value string) uint32 {
	for index, candidate := range pieceTypes {
		if value == candidate {
			return uint32(index)
		}
	}
	return protocol.PieceTypeUnsupported
}

func legacyRGB(value color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", value.R, value.G, value.B)
}

func legacyARGB(value color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x%02x", value.A, value.R, value.G, value.B)
}

func parseLegacyColour(io *wireIO, value string) color.RGBA {
	hex := strings.TrimPrefix(value, "#")
	parsed, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		io.InvalidValue(value, "skin colour", err.Error())
		return color.RGBA{}
	}
	switch len(hex) {
	case 6:
		return color.RGBA{R: byte(parsed >> 16), G: byte(parsed >> 8), B: byte(parsed), A: 0xff}
	case 8:
		return color.RGBA{A: byte(parsed >> 24), R: byte(parsed >> 16), G: byte(parsed >> 8), B: byte(parsed)}
	default:
		io.InvalidValue(value, "skin colour", "must use RGB or ARGB hex")
		return color.RGBA{}
	}
}
