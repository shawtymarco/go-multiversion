<p align="center">
  <img src="assets/go-gopher.png" alt="Go gopher" width="180">
</p>

<h1 align="center">go-multiversion</h1>

<p align="center">Minecraft Bedrock multiversion adapters for gophertunnel and Dragonfly.</p>

## ✅ Supported Versions

| Protocol ID | Minecraft version | Adapter | Support | Tested |
|------------:|-------------------|---------|:-------:|-------------|
| 2169 | 1.26.45 | Native gophertunnel | ✅ | ✅ Native advertisement |
| 2168 | 1.26.40-1.26.44 | `v1_26_44` | ✅ | ✅ |
| 1001 | 1.26.30-1.26.34, 1.26.36 | `v1_26_30` | ✅ | ✅ 1.26.33 |
| 975 | 1.26.20, 1.26.21, 1.26.23 | `v1_26_20` | ✅ | ✅ 1.26.20/1.26.21 · 🧪 1.26.23 automated |
| 944 | 1.26.10-1.26.14 | `v1_26_10` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 924 | 1.26.0-1.26.3 | `v1_26_0` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 898 | 1.21.130-1.21.132 | `v1_21_130` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 844 | 1.21.110-1.21.114 | `v1_21_110` | ✅ | ✅ 1.21.114 |
| 827 | 1.21.100-1.21.102 | `v1_21_100` | ✅ | ✅ 1.21.100 |
| 766 | 1.21.50-1.21.51 | `v1_21_50` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 748 | 1.21.40, 1.21.41, 1.21.43, 1.21.44 | `v1_21_40` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 486 | 1.18.10-1.18.12 | `v1_18_10` | ✅ | ✅ 1.18.10 · 🧪 1.18.11/1.18.12 automated |
| 475 | 1.18.0-1.18.2 | `v1_18_0` | 🧪 | 🧪 Historical wire oracle and registry integration |
| 419 | 1.16.100 | `v1_16_100` | 🧪 | 🧪 Historical wire, encryption, registry, and chunk-format tests |

> [!NOTE]
> Version coverage is explicit. Unlisted releases and previews are not implied.

## 🚀 Usage

Registry-aware adapters must be created after Dragonfly finalises its native
block registry:

```go
conf.AcceptedProtocolsProvider = func(blocks world.BlockRegistry) ([]minecraft.Protocol, error) {
	return multiversion.ProtocolsWithRegistries(blocks, dragonfly.VanillaItemEntries())
}
```

`ProtocolsWithRegistries` returns `2168`, `1001`, `975`, `944`, `924`, `898`,
`844`, `827`, `766`, `748`, `486`, `475`, and `419` in
that order. The parameterless `Protocols()` intentionally omits adapters that
need native block and item registries.

## 🔗 Dependencies

**Library**

- [Sandertv/gophertunnel](https://github.com/Sandertv/gophertunnel) provides
  the current native protocol model and generic protocol interface.
- [df-mc/worldupgrader](https://github.com/df-mc/worldupgrader) upgrades
  historical block and item identifiers before semantic mapping.
- `go-multiversion` does **not** import Dragonfly directly.

**Minecraft 1.16/1.18 transport**

Servers that want to accept protocol `486`, `475`, or `419` clients must use
[`shawtymarco/gophertunnel`](https://github.com/shawtymarco/gophertunnel) at
`1524deb2ed1a4c65f595685e30bf38534302e877` or an equivalent implementation.
The fork keeps the `github.com/sandertv/gophertunnel` module path while adding:

- RakNet v10 acceptance without changing the native v11 advertisement;
- Login-first protocol selection and legacy flate batches without an algorithm prefix;
- protocol-owned pre-spawn packets required before `PlayStatus(PlayerSpawn)`.
- protocol-selected AES-CFB8 encryption for 1.16.100 while native and 1.18 retain AES-CTR.

The standard RakNet v11 and `RequestNetworkSettings` path remains unchanged for modern clients.

**Dragonfly integration**

Dragonfly consumers require
[`shawtymarco/dragonfly`](https://github.com/shawtymarco/dragonfly) at
`cef8669ca8e873ee9622a4e5d425921fef1458d1` or an implementation with equivalent
hooks:

- `AcceptedProtocolsProvider` after block-registry finalisation;
- `VanillaItemEntries()` for native item mapping;
- protocol access on each connection;
- `BlockRuntimeIDMapper` and protocol-aware chunk encoding before cache hashing.
- protocol-neutral target range, sub-chunk-version, and 2D-biome encoding for 1.16.100.

> [!WARNING]
> Stock upstream Dragonfly does not currently provide these hooks. It cannot be
> used as a drop-in replacement for registry-aware protocols `1001`, `975`, `944`, `924`, `898`,
> `844`, `827`, `766`, `748`, `486`, `475`, or `419`. Without pre-hash palette mapping, old clients receive native block
> runtime IDs and incompatible cache blobs.

Protocols `486`, `475`, and `419` additionally require the gophertunnel fork above. The adapter
alone cannot make a stock listener accept RakNet v10 or decode Login-first
legacy batches.

## 🙏 Credits

- [Sandertv/gophertunnel](https://github.com/Sandertv/gophertunnel)
- [df-mc/dragonfly](https://github.com/df-mc/dragonfly)
- [df-mc/worldupgrader](https://github.com/df-mc/worldupgrader)
- [Mojang/bedrock-protocol-docs](https://github.com/Mojang/bedrock-protocol-docs)
- [EndstoneMC/bedrock-server-data](https://github.com/EndstoneMC/bedrock-server-data)
- [TedacMC/tedac](https://github.com/TedacMC/tedac) for the MIT-licensed protocol-419 registry cross-checks
- [Go gopher](https://go.dev/blog/gopher) by Renee French, used under
  [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
