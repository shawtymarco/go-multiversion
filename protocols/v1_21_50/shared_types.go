package v1_21_50

import (
	"fmt"
	"image/color"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func marshalEntityMetadata(io *wireIO, metadata *protocol.EntityMetadata) {
	if !io.reading {
		filtered := make(protocol.EntityMetadata, len(*metadata))
		for key, value := range *metadata {
			if key == protocol.EntityDataKeyInvulnerableTicks || key > protocol.EntityDataKeyVisibleMobEffects {
				continue
			}
			filtered[key] = value
		}
		metadata = &filtered
	}
	if io.reading {
		*metadata = protocol.EntityMetadata{}
	}
	count := uint32(len(*metadata))
	io.Varuint32(&count)
	if io.reading {
		for range count {
			var key, dataType uint32
			io.Varuint32(&key)
			io.Varuint32(&dataType)
			(*metadata)[key] = readEntityMetadataValue(io, dataType)
		}
		return
	}
	keys := make([]int, 0, count)
	for key := range *metadata {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)
	for _, rawKey := range keys {
		key := uint32(rawKey)
		io.Varuint32(&key)
		writeEntityMetadataValue(io, (*metadata)[key])
	}
}

func writeEntityMetadataValue(io *wireIO, value any) {
	writeType := func(dataType uint32) { io.Varuint32(&dataType) }
	switch value := value.(type) {
	case byte:
		writeType(protocol.EntityDataTypeByte)
		io.Uint8(&value)
	case int16:
		writeType(protocol.EntityDataTypeInt16)
		io.Int16(&value)
	case int32:
		writeType(protocol.EntityDataTypeInt32)
		io.Varint32(&value)
	case float32:
		writeType(protocol.EntityDataTypeFloat32)
		io.Float32(&value)
	case string:
		writeType(protocol.EntityDataTypeString)
		io.String(&value)
	case map[string]any:
		writeType(protocol.EntityDataTypeCompoundTag)
		io.NBT(&value, nbt.NetworkLittleEndian)
	case protocol.BlockPos:
		writeType(protocol.EntityDataTypeBlockPos)
		io.BlockPos(&value)
	case int64:
		writeType(protocol.EntityDataTypeInt64)
		io.Varint64(&value)
	case mgl32.Vec3:
		writeType(protocol.EntityDataTypeVec3)
		io.Vec3(&value)
	default:
		io.UnknownEnumOption(reflect.TypeOf(value), "entity metadata")
	}
}

func readEntityMetadataValue(io *wireIO, dataType uint32) any {
	switch dataType {
	case protocol.EntityDataTypeByte:
		var value byte
		io.Uint8(&value)
		return value
	case protocol.EntityDataTypeInt16:
		var value int16
		io.Int16(&value)
		return value
	case protocol.EntityDataTypeInt32:
		var value int32
		io.Varint32(&value)
		return value
	case protocol.EntityDataTypeFloat32:
		var value float32
		io.Float32(&value)
		return value
	case protocol.EntityDataTypeString:
		var value string
		io.String(&value)
		return value
	case protocol.EntityDataTypeCompoundTag:
		var value map[string]any
		io.NBT(&value, nbt.NetworkLittleEndian)
		return value
	case protocol.EntityDataTypeBlockPos:
		var value protocol.BlockPos
		io.BlockPos(&value)
		return value
	case protocol.EntityDataTypeInt64:
		var value int64
		io.Varint64(&value)
		return value
	case protocol.EntityDataTypeVec3:
		var value mgl32.Vec3
		io.Vec3(&value)
		return value
	default:
		io.UnknownEnumOption(dataType, "entity metadata")
		return nil
	}
}

func marshalGameRule(io *wireIO, rule *protocol.GameRule, legacyInteger bool) {
	io.String(&rule.Name)
	io.Bool(&rule.CanBeModifiedByPlayer)
	var kind uint32
	if !io.reading {
		switch rule.Value.(type) {
		case bool:
			kind = 1
		case uint32:
			kind = 2
		case float32:
			kind = 3
		default:
			io.UnknownEnumOption(fmt.Sprintf("%T", rule.Value), "game rule type")
			return
		}
	}
	io.Varuint32(&kind)
	switch kind {
	case 1:
		value, _ := rule.Value.(bool)
		io.Bool(&value)
		rule.Value = value
	case 2:
		value, _ := rule.Value.(uint32)
		if legacyInteger {
			io.Varuint32(&value)
		} else {
			io.Uint32(&value)
		}
		rule.Value = value
	case 3:
		value, _ := rule.Value.(float32)
		io.Float32(&value)
		rule.Value = value
	default:
		io.UnknownEnumOption(kind, "game rule type")
	}
}

func marshalAbilityData(io *wireIO, data *protocol.AbilityData) {
	io.Int64(&data.EntityUniqueID)
	io.Uint8(&data.PlayerPermissions)
	io.Uint8(&data.CommandPermissions)
	count := uint8(len(data.Layers))
	io.Uint8(&count)
	if io.reading {
		data.Layers = make([]protocol.AbilityLayer, count)
	}
	for i := range data.Layers {
		layer := &data.Layers[i]
		io.Uint16(&layer.Type)
		io.Uint32(&layer.Abilities)
		io.Uint32(&layer.Values)
		io.Float32(&layer.FlySpeed)
		io.Float32(&layer.WalkSpeed)
		if io.reading {
			layer.VerticalFlySpeed = 0
		}
	}
}

func marshalPlayerListEntry(io protocol.IO, entry *protocol.PlayerListEntry) {
	legacy := asWireIO(io)
	legacy.UUID(&entry.UUID)
	legacy.Varint64(&entry.EntityUniqueID)
	legacy.String(&entry.Username)
	legacy.String(&entry.XUID)
	legacy.String(&entry.PlatformChatID)
	legacy.Int32(&entry.BuildPlatform)
	marshalSkin(legacy, &entry.Skin)
	legacy.Bool(&entry.Teacher)
	legacy.Bool(&entry.Host)
	legacy.Bool(&entry.SubClient)
	if legacy.reading {
		entry.PlayerColour = color.RGBA{}
	}
}

func marshalLegacyARGB(io protocol.IO, value *color.RGBA) {
	packed := int32(value.A) | int32(value.R)<<8 | int32(value.G)<<16 | int32(value.B)<<24
	io.Int32(&packed)
	*value = color.RGBA{A: byte(packed), R: byte(packed >> 8), G: byte(packed >> 16), B: byte(packed >> 24)}
}

var personaPieceTypes = [...]string{
	"unknown", "persona_skeleton", "persona_body", "persona_skin", "persona_bottom", "persona_feet",
	"persona_dress", "persona_top", "persona_high_pants", "persona_hands", "persona_outerwear",
	"persona_facial_hair", "persona_mouth", "persona_eyes", "persona_hair", "persona_hood", "persona_back",
	"persona_face_accessory", "persona_head", "persona_legs", "persona_left_leg", "persona_right_leg",
	"persona_arms", "persona_left_arm", "persona_right_arm", "persona_capes", "persona_classic_skin",
	"persona_emote", "unsupported",
}

func marshalSkin(io *wireIO, skin *protocol.Skin) {
	if !io.reading {
		copySkin := *skin
		copySkin.Animations = append([]protocol.SkinAnimation(nil), skin.Animations...)
		copySkin.PersonaPieces = append([]protocol.PersonaPiece(nil), skin.PersonaPieces...)
		copySkin.PieceTintColours = append([]protocol.PersonaPieceTintColour(nil), skin.PieceTintColours...)
		skin = &copySkin
	}
	io.String(&skin.SkinID)
	io.String(&skin.PlayFabID)
	io.ByteSlice(&skin.SkinResourcePatch)
	io.Uint32(&skin.SkinImageWidth)
	io.Uint32(&skin.SkinImageHeight)
	io.ByteSlice(&skin.SkinData)
	marshalSkinAnimations(io, &skin.Animations)
	io.Uint32(&skin.CapeImageWidth)
	io.Uint32(&skin.CapeImageHeight)
	io.ByteSlice(&skin.CapeData)
	io.ByteSlice(&skin.SkinGeometry)
	io.ByteSlice(&skin.GeometryDataEngineVersion)
	io.ByteSlice(&skin.AnimationData)
	io.String(&skin.CapeID)
	io.String(&skin.FullID)
	armSize := "slim"
	if skin.ArmSize == protocol.ArmSizeWide {
		armSize = "wide"
	}
	io.String(&armSize)
	if strings.EqualFold(armSize, "wide") {
		skin.ArmSize = protocol.ArmSizeWide
	} else {
		skin.ArmSize = protocol.ArmSizeSlim
	}
	skinColour := fmt.Sprintf("#%02x%02x%02x", skin.SkinColour.R, skin.SkinColour.G, skin.SkinColour.B)
	io.String(&skinColour)
	skin.SkinColour = parseLegacyColour(skinColour, false)
	marshalPersonaPieces(io, &skin.PersonaPieces)
	marshalPersonaTints(io, &skin.PieceTintColours)
	io.Bool(&skin.PremiumSkin)
	io.Bool(&skin.PersonaSkin)
	io.Bool(&skin.PersonaCapeOnClassicSkin)
	io.Bool(&skin.PrimaryUser)
	io.Bool(&skin.OverrideAppearance)
}

func marshalSkinAnimations(io *wireIO, animations *[]protocol.SkinAnimation) {
	count := uint32(len(*animations))
	io.Uint32(&count)
	if io.reading {
		*animations = make([]protocol.SkinAnimation, count)
	}
	for i := range *animations {
		animation := &(*animations)[i]
		io.Uint32(&animation.ImageWidth)
		io.Uint32(&animation.ImageHeight)
		io.ByteSlice(&animation.ImageData)
		io.Uint32(&animation.AnimationType)
		io.Float32(&animation.FrameCount)
		io.Uint32(&animation.ExpressionType)
	}
}

func marshalPersonaPieces(io *wireIO, pieces *[]protocol.PersonaPiece) {
	count := uint32(len(*pieces))
	io.Uint32(&count)
	if io.reading {
		*pieces = make([]protocol.PersonaPiece, count)
	}
	for i := range *pieces {
		piece := &(*pieces)[i]
		io.String(&piece.PieceID)
		pieceType := "unknown"
		if int(piece.PieceType) < len(personaPieceTypes) {
			pieceType = personaPieceTypes[piece.PieceType]
		}
		io.String(&pieceType)
		for index, name := range personaPieceTypes {
			if name == pieceType {
				piece.PieceType = uint32(index)
				break
			}
		}
		packID := piece.PackID.String()
		io.String(&packID)
		if parsed, err := uuid.Parse(packID); err == nil {
			piece.PackID = parsed
		}
		io.Bool(&piece.Default)
		io.String(&piece.ProductID)
	}
}

func marshalPersonaTints(io *wireIO, tints *[]protocol.PersonaPieceTintColour) {
	count := uint32(len(*tints))
	io.Uint32(&count)
	if io.reading {
		*tints = make([]protocol.PersonaPieceTintColour, count)
	}
	for i := range *tints {
		tint := &(*tints)[i]
		io.String(&tint.PieceType)
		colourCount := uint32(len(tint.Colours))
		io.Uint32(&colourCount)
		for colourIndex := uint32(0); colourIndex < colourCount; colourIndex++ {
			colourString := "#0"
			if int(colourIndex) < len(tint.Colours) {
				value := tint.Colours[colourIndex]
				colourString = fmt.Sprintf("#%02x%02x%02x%02x", value.A, value.R, value.G, value.B)
			}
			io.String(&colourString)
			if colourIndex < uint32(len(tint.Colours)) {
				tint.Colours[colourIndex] = parseLegacyColour(colourString, true)
			}
		}
	}
}

func parseLegacyColour(value string, argb bool) color.RGBA {
	hexValue := strings.TrimPrefix(value, "#")
	if hexValue == "" || hexValue == "0" {
		return color.RGBA{}
	}
	parsed, err := strconv.ParseUint(hexValue, 16, 32)
	if err != nil {
		return color.RGBA{}
	}
	if argb && len(hexValue) > 6 {
		return color.RGBA{A: byte(parsed >> 24), R: byte(parsed >> 16), G: byte(parsed >> 8), B: byte(parsed)}
	}
	return color.RGBA{R: byte(parsed >> 16), G: byte(parsed >> 8), B: byte(parsed), A: 0xff}
}

func marshalSubChunkEntry(io *wireIO, entry *protocol.SubChunkEntry, cacheEnabled bool) {
	io.Int8(&entry.Offset[0])
	io.Int8(&entry.Offset[1])
	io.Int8(&entry.Offset[2])
	io.Uint8(&entry.Result)
	if cacheEnabled {
		if entry.Result != protocol.SubChunkResultSuccessAllAir {
			payload, _ := entry.RawPayload.Value()
			io.ByteSlice(&payload)
			entry.RawPayload = protocol.Option(payload)
		}
	} else {
		payload, _ := entry.RawPayload.Value()
		io.ByteSlice(&payload)
		entry.RawPayload = protocol.Option(payload)
	}
	io.Uint8(&entry.HeightMapType)
	if entry.HeightMapType == protocol.HeightMapDataHasData {
		data, _ := entry.HeightMapData.Value()
		protocol.FuncSliceOfLen(io.directional(), 256, &data, io.Int8)
		entry.HeightMapData = protocol.Option(data)
	}
	if io.reading {
		entry.RenderHeightMapType = protocol.HeightMapDataNone
		entry.RenderHeightMapData = protocol.Optional[[]int8]{}
	}
	if cacheEnabled {
		blobHash, _ := entry.BlobHash.Value()
		io.Uint64(&blobHash)
		entry.BlobHash = protocol.Option(blobHash)
	} else if io.reading {
		entry.BlobHash = protocol.Optional[uint64]{}
	}
}

func marshalShapeData(io *wireIO, value *protocol.ShapeData) { io.IO.ShapeData(value) }
