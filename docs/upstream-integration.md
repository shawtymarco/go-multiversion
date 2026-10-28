# Upstream integration boundary

## Exact sources

- Core native wire: upstream gophertunnel v1.63.0,
  `8c30047d86617a56e1f1871e3a1d47ec4f670d99`.
- Previous transport fork: `1fa0b31e78a8e76dc261e86792aacd3c2da8ff25`.
  Its diff against the native upstream commit contains legacy negotiation and
  encryption, receive-order/spawn fixes, GameData synchronization, configurable
  MTU, skin normalization and dual-transport lifecycle changes. Switching to
  upstream does not imply these unrelated fixes were merged or are unnecessary.
- Official Dragonfly master inspected: `1bf4e242c81f1f66ba6bdf8fcdcc80c70125a499`.
- Compatible native Dragonfly PR source: df-mc/dragonfly PR 1481,
  `d86aefd9ff9b6a357d6dfadab2093af811545256`.
  This is the same stable registry source already locked in `data/v2225`.
- Native hash bridge ported from this project's existing Dragonfly integration
  (shawtymarco/dragonfly `0f954a8c0a128e2cc63f27cf048b93255d6061c6`,
  `server/native_protocol.go`), preserving its packet
  semantics while moving ownership into the optional module.

No historical packet wire layout or snapshot is changed. The existing per-version
source locks and independent byte oracles remain applicable.

## Implemented ownership change

The root module no longer replaces gophertunnel. Its default registry-aware
catalogue includes only protocol 748 and newer. Historical 419/475/486 codecs
remain buildable and tested, but registering them requires an explicit transport
decision. `legacytransport` restores their catalogue entries and the 419 CFB8
method; that opt-in still requires the old transport fork.

The optional `dragonfly` module uses Config.Listeners and session.Conn rather than
fork-specific AcceptedProtocolsProvider or pre-hash encoder hooks. The actual
finalized BlockRegistry already satisfies mapping.BlockRegistry. Initial item
mapping comes from upstream GameData/ItemRegistry, avoiding a separate item-table
export and preserving custom-entry updates.

Chunk conversion is based on the exact PR's `server/world/chunk/encode.go`,
`encoding.go`, `decode.go` and `server/session/chunk.go`: version-9 sub-chunks,
numeric zigzag palette entries, little-endian packed indices, and one 3D biome
storage per dimension section. Each palette is remapped and deduplicated, indices
are repacked at the smallest valid width, and biome reuse markers retain their
previous-section meaning. The converter handles both full LevelChunk payloads
and request-mode biome packets followed by SubChunk responses. It preserves
trailing border/block-entity bytes and rejects malformed input.

The connection reports ClientCacheEnabled=false to Dragonfly before any encoding
or cache hashing. Cached LevelChunk/SubChunk or cache-miss output is rejected.
This intentionally trades cache efficiency for a self-contained integration.

Native block/item/event conversion remains direct between local runtime IDs and
full state hashes. Historical adapters still map directly to/from the local native
registry; no version chaining is introduced.

## Validation and remaining gates

The root tests and vet run against upstream gophertunnel. The optional module's
dedicated CI runner substitutes the exact PR source in a temporary modfile only.
It checks:

- palette round trips, high-bit hashes, deduplication/width changes, invalid
  indices, truncation, biome reuse and preserved NBT tails;
- the real upstream chunk encoder through every modern adapter in all three
  vanilla dimensions, with matching LevelChunk/SubChunk semantics;
- native hash packet conversions in both directions and non-mutation;
- automated Dragonfly server joins through the native and all modern wire codecs;
- explicit cache rejection and configuration validation.

Dragonfly PR merge/publication is a release prerequisite. Before publishing this
module, remove the workspace core replacement, pin a published core revision and
a compatible official Dragonfly commit, and rerun ordinary builds with no
replacement. The root module can be released independently.

Real Mojang clients, custom resource-pack content, sustained chunk traffic,
inventory/crafting interactions, transfers, and upstream transport race/load
behaviour remain validation gates. This change does not deploy a server or claim
that automated joins establish gameplay support. NetherNet and cache-enabled
translation are outside this first adapter.
