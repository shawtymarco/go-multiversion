# Upstream Dragonfly adapter — pending PR 1481

This is an optional integration module. The core go-multiversion module does not
import Dragonfly and uses unmodified upstream gophertunnel v1.63.0.

**Release blocker:** Dragonfly master `1bf4e242c81f` requires the old Skin.PlayFabID
field and does not compile with gophertunnel v1.63.0. The compatible native update
is [upstream PR 1481](https://github.com/df-mc/dragonfly/pull/1481), locked to
`d86aefd9ff9b6a357d6dfadab2093af811545256`. It is currently available only through
the PR author's repository, not as a resolvable official module revision.
Do not publish or describe this module as ready for installation yet.

## Intended integration

After the compatible Dragonfly update is merged and this module is released:

```go
import mvdragonfly "github.com/shawtymarco/go-multiversion/dragonfly"

// uc and conf are the application's existing Dragonfly configurations.
if err := mvdragonfly.Enable(&conf, uc.Network.Address); err != nil {
    return err
}
srv := conf.New()
```

`Enable` replaces all configured listeners with one RakNet listener. It preserves
authentication, admission policy, resource packs, compression, status provider
and maximum player count. For explicit composition or a protocol floor:

```go
conf.Listeners = []func(server.Config) (server.Listener, error){
    (mvdragonfly.Config{MinimumProtocol: 827}).Listener(uc.Network.Address),
}
```

The factory runs after Dragonfly finalises the actual block registry. Item
mapping uses the real GameData/ItemRegistry, including server custom items.
Historical packet codecs run in gophertunnel; chunk payloads are translated by
the session connection wrapper before those codecs. Native connections use
full block-state hashes; historical connections use their own numeric registries.

Chunk caching is disabled at the session connection boundary for all clients.
LevelChunk, SubChunk and biome palettes are decoded, mapped, deduplicated and
repacked without changing world chunks. Border/NBT bytes and sub-chunk indices
are preserved. Unexpected cached payloads fail closed. Existing native cache
hashes are never reused after translation.

Only the catalogue's 1.21.40+ families and native 1.26.60 are enabled. The
`legacytransport` tag does not extend this module. NetherNet listener setup is
not implemented here. A later upstream encoder/hash change must be audited
before advancing the Dragonfly pin; applying these conversions twice is invalid.

## Reproducible development check

From the repository root:

```text
python tools/ci/upstream_dragonfly.py
```

This copies the module files into a temporary directory, substitutes only the
exact PR source for verification, and runs tests, vet and build. The checked-in
module files retain the official Dragonfly dependency and therefore **plain
`go test ./...` inside this directory remains blocked by upstream master**.
The root suite runs independently without this pending module.

The `go-multiversion => ..` replacement is repository-relative development
wiring, not a published dependency pin. Before release, pin the published core
commit, remove that replacement, pin a compatible official Dragonfly revision,
and pass ordinary module tests without the PR override. Automated joins do not
replace Mojang-client checks for inventory, crafting, block rendering and gameplay.
