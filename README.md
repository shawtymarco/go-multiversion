<p align="center">
  <img src="assets/go-multiversion-gopher.png" alt="go-multiversion gopher" width="180">
</p>

<h1 align="center">go-multiversion</h1>

<p align="center">Minecraft Bedrock multiversion adapters for gophertunnel and Dragonfly.</p>

## Protocols

The native wire model is stable **Minecraft 1.26.60 / protocol 2225**.
Historical adapters map directly to it, including the outgoing 1.26.50 family.
[Stable source locks](versions/1.26.60-native.md) distinguish the final data from
retained Preview acquisitions. Automated checks do not establish Mojang-client
gameplay compatibility.

| Protocol ID | Minecraft version | Adapter | Snapshot coverage | Wire tests |
|------------:|-------------------|---------|:-------:|:------:|
| 2225 | 1.26.60 | Native gophertunnel + Dragonfly hash bridge | ✅ | ✅ |
| 2193 | 1.26.50–1.26.52 | `v1_26_50` | ✅ | ✅ |
| 2169 | 1.26.45 | `v1_26_45` | ✅ | ✅ |
| 2168 | 1.26.40-1.26.44 | `v1_26_44` | ✅ | ✅ |
| 1001 | 1.26.30-1.26.34, 1.26.36 | `v1_26_30` | ✅ | ✅ |
| 975 | 1.26.20, 1.26.21, 1.26.23 | `v1_26_20` | ✅ | ✅ |
| 944 | 1.26.10-1.26.14 | `v1_26_10` | ✅ | ✅ |
| 924 | 1.26.0-1.26.3 | `v1_26_0` | ✅ | ✅ |
| 898 | 1.21.130-1.21.132 | `v1_21_130` | ✅ | ✅ |
| 844 | 1.21.110-1.21.114 | `v1_21_110` | ✅ | ✅ |
| 827 | 1.21.100-1.21.102 | `v1_21_100` | ✅ | ✅ |
| 766 | 1.21.50-1.21.51 | `v1_21_50` | ✅ | ✅ |
| 748 | 1.21.40, 1.21.41, 1.21.43, 1.21.44 | `v1_21_40` | ✅ | ✅ |
| 486 (opt-in) | 1.18.10-1.18.12 | `v1_18_10` | ✅ | ✅ |
| 475 (opt-in) | 1.18.0-1.18.2 | `v1_18_0` | ✅ | ✅ |
| 419 (opt-in) | 1.16.100 | `v1_16_100` | ✅ | ✅ |

Unlisted releases and preview builds are not implied.

[CI coverage and local commands](docs/ci.md) describe the automated checks.
[Dual public transport admission](docs/dual-transport.md) records the shared
RakNet/NetherNet session owner, implementation history and integration gates.

## Usage

The core module uses **upstream gophertunnel v1.63.0**, without a fork replacement.
The default registry-aware catalogue includes the supported **1.21.40+** families
(protocol 748 and newer). Native remains stable 1.26.60 / protocol 2225.

```go
adapters, err := (multiversion.Config{MinimumProtocol: 748}).ProtocolsWithRegistries(blocks, items)
// Supply adapters to minecraft.ListenConfig.AcceptedProtocols after handling err.
```

`blocks` and `items` must describe the server's actual finalized registries.
The parameterless `Protocols()` intentionally returns no unconfigured historical
adapters. `MinimumProtocol` is an inclusive floor: zero selects the default modern
catalogue, and 2225 selects no historical adapters. Unlisted families are not added.
The core codecs do not translate opaque chunk payloads on their own: consumers
must integrate palette conversion or use the Dragonfly adapter below.

### Upstream Dragonfly integration (pending upstream release)

The optional [Dragonfly module](dragonfly/README.md) supplies a listener and
connection wrapper, actual-registry mapping, a native hash bridge, and external
chunk translation. Its intended startup API is:

```go
import mvdragonfly "github.com/shawtymarco/go-multiversion/dragonfly"

// After uc.Config(...), before conf.New():
if err := mvdragonfly.Enable(&conf, uc.Network.Address); err != nil {
    return err
}
srv := conf.New()
```

**This module is not ready for a fork-free release yet.** Dragonfly master
`1bf4e242` still refers to Skin.PlayFabID, removed in the required gophertunnel
v1.63.0. [Upstream PR 1481](https://github.com/df-mc/dragonfly/pull/1481), exact
commit `d86aefd9`, provides the compatible native update but is not merged or
resolvable as a normal df-mc Go module revision. CI validates against that PR in a
temporary modfile only. The core module has no Dragonfly dependency and is usable
independently. Existing fork integrations can continue using their generic hooks.

The new adapter uses RakNet and disables client chunk caching before Dragonfly
encodes chunks. `Enable` replaces existing listeners; explicit listener composition
is available. NetherNet integration and real-client gameplay checks remain separate
work. See [source locks and release gates](docs/upstream-integration.md).

### Optional legacy transport

Protocols 486, 475 and 419 retain their historical codecs, snapshots and wire tests,
but are **not advertised by the default catalogue**. Direct constructors are codec
APIs, not a promise that upstream gophertunnel accepts their Login-first sessions.
Existing consumers that need these families must explicitly choose both the
`legacytransport` build tag and the compatible transport fork:

```text
go mod edit -replace=github.com/sandertv/gophertunnel=github.com/shawtymarco/gophertunnel@v1.57.1-0.20261028052605-1fa0b31e78a8
go test -tags=legacytransport ./...
```

The tag restores their catalogue entries and the 1.16 CFB8 capability. It does not
add legacy chunk support to the new upstream Dragonfly module. Keep existing
legacy server integration hooks when opting into this mode.

## Credits

- [Sandertv/gophertunnel](https://github.com/Sandertv/gophertunnel)
- [df-mc/dragonfly](https://github.com/df-mc/dragonfly)
- [df-mc/worldupgrader](https://github.com/df-mc/worldupgrader)
- [Mojang/bedrock-protocol-docs](https://github.com/Mojang/bedrock-protocol-docs)
- [EndstoneMC/bedrock-server-data](https://github.com/EndstoneMC/bedrock-server-data)
- [TedacMC/tedac](https://github.com/TedacMC/tedac) for protocol-419 registry cross-checks
- Mascot adapted with OpenAI ImageGen from the [Go gopher](https://go.dev/blog/gopher) by Renee French, under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
