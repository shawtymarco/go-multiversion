package v1_16_100

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func marshalResourcePacksInfo(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePacksInfo)
	io.Bool(&pk.TexturePackRequired)
	io.Bool(&pk.HasScripts)
	var behaviourPackCount uint16
	io.Uint16(&behaviourPackCount)
	if behaviourPackCount != 0 {
		io.InvalidValue(behaviourPackCount, "behaviour pack count", "current server has no behaviour-pack descriptors")
		return
	}
	count := uint16(len(pk.TexturePacks))
	io.Uint16(&count)
	if io.reading {
		pk.TexturePacks = make([]protocol.TexturePackInfo, count)
	}
	for index := range pk.TexturePacks {
		marshalTexturePackInfo(io, &pk.TexturePacks[index])
	}
	if io.reading {
		pk.HasAddons = false
		pk.ForceDisableVibrantVisuals = false
		pk.WorldTemplateUUID = uuid.Nil
		pk.WorldTemplateVersion = ""
	}
}

func marshalTexturePackInfo(io *wireIO, pack *protocol.TexturePackInfo) {
	identifier := pack.UUID.String()
	io.String(&identifier)
	if io.reading {
		parsed, err := uuid.Parse(identifier)
		if err != nil {
			io.InvalidValue(identifier, "texture pack UUID", err.Error())
			return
		}
		pack.UUID = parsed
	}
	io.String(&pack.Version)
	io.Uint64(&pack.Size)
	io.String(&pack.ContentKey)
	io.String(&pack.SubPackName)
	io.String(&pack.ContentIdentity)
	io.Bool(&pack.HasScripts)
	if io.reading {
		pack.AddonPack = false
		pack.RTXEnabled = false
		pack.DownloadURL = ""
	}
}

func marshalResourcePackStack(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ResourcePackStack)
	io.Bool(&pk.TexturePackRequired)
	var behaviourPacks []protocol.StackResourcePack
	protocol.Slice(io.directional(), &behaviourPacks)
	protocol.Slice(io.directional(), &pk.TexturePacks)
	io.String(&pk.BaseGameVersion)
	protocol.SliceUint32Length(io.directional(), &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	if io.reading {
		pk.IncludeEditorPacks = false
	}
}
