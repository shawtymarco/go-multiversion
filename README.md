# 🐉 df-multiversion

A gophertunnel protocol interface implementation for running selected older
Minecraft Bedrock clients against a current Dragonfly server.

> [!IMPORTANT]
> The server always keeps the latest gophertunnel protocol as its native model.
> Every older release converts directly to and from native at the network boundary.

## ✅ Supported Versions

| Protocol ID | Minecraft version | Adapter | Support | Real client |
|------------:|-------------------|---------|:-------:|-------------|
| 2169 | 1.26.45 | Native gophertunnel | ✅ | ✅ Native advertisement |
| 2168 | 1.26.40-1.26.44 | `v1_26_44` | ✅ | ✅ |
| 1001 | 1.26.30-1.26.34, 1.26.36 | `v1_26_30` | ✅ | ✅ 1.26.33 |
| 844 | 1.21.110-1.21.114 | `v1_21_110` | ✅ | ✅ 1.21.114 |
| 827 | 1.21.100-1.21.102 | `v1_21_100` | ✅ | ✅ 1.21.100 |

> [!NOTE]
> ✅ marks adapters enabled by the registry-aware protocol catalogue. The real-client
> column records representative builds that have connected successfully to CastleOnline.

> [!WARNING]
> Version coverage is explicit. A supported protocol family does not imply support for
> unlisted previews, releases, or another release train.

## 🔌 How It Works

- **2168** selects the correct 1.26.40-1.26.44 `SetScore` layout from the login
  `GameVersion` because those releases reuse one protocol ID.
- **1001, 844, and 827** use exact historical block/item data, semantic registry
  mappings, creative-selection preservation, recipe filtering, and protocol-aware
  chunk palette encoding before cache hashing.
- Unknown blocks use recorded fallbacks; unknown items are hidden clientbound and
  rejected serverbound instead of reusing numeric runtime IDs.
- Native gameplay state is never downgraded in memory. Conversion happens per
  connection at the protocol boundary.

## 🚀 Usage

Registry-aware consumers should initialise every mapped adapter only after their
native block and item registries are finalised:

```go
protocols, err := multiversion.ProtocolsWithRegistries(nativeBlocks, nativeItems)
if err != nil {
	return err
}
```

`ProtocolsWithRegistries` returns the verified non-native protocols in this order:

```text
2168, 1001, 844, 827
```

The parameterless `Protocols()` intentionally returns only adapters that do not
need runtime registries. This prevents an unconfigured consumer from reusing native
runtime IDs for an older client.

## 📚 Version Evidence

| Family | Source lock | Wire audit | Mapping | Chunks |
|--------|-------------|------------|---------|--------|
| 1.26.3x | [`1.26.3x.yaml`](versions/1.26.3x.yaml) | [`wire`](versions/1.26.3x-wire.md) | [`mapping`](versions/1.26.3x-mapping.md) | [`chunks`](versions/1.26.3x-chunks.md) |
| 1.21.11x | [`1.21.11x.yaml`](versions/1.21.11x.yaml) | [`wire`](versions/1.21.11x-wire.md) | [`mapping`](versions/1.21.11x-mapping.md) | [`chunks`](versions/1.21.11x-chunks.md) |
| 1.21.10x | [`1.21.10x.yaml`](versions/1.21.10x.yaml) | [`wire`](versions/1.21.10x-wire.md) | [`mapping`](versions/1.21.10x-mapping.md) | [`chunks`](versions/1.21.10x-chunks.md) |

## 🙏 Credits

- [Sandertv/gophertunnel](https://github.com/Sandertv/gophertunnel)
- [df-mc/dragonfly](https://github.com/df-mc/dragonfly)
- [Mojang/bedrock-protocol-docs](https://github.com/Mojang/bedrock-protocol-docs)
- [iAmFrogger/legacy-version](https://github.com/iAmFrogger/legacy-version) — README table inspiration

See [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md) for bundled historical-source notices.
