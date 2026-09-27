# Standalone 1.18 proxy adapter

The `proxy` package and `multiversion.NewProxyAdapter` add opt-in, session-local
conversion for protocol 475 (1.18.0/1/2) and 486 (1.18.10/11/12) to native 2193.
Existing server-facing constructors retain their behavior.

## Locked sources

| Boundary | Exact revision |
| --- | --- |
| 475 wire | Sandertv/gophertunnel c40bf8288fb93ebd014375e46e4592424153927c |
| 486 wire | Sandertv/gophertunnel 2cb1e399e53928529916dd2bda1efd11a0ac374d |
| 2193 wire | Sandertv/gophertunnel 7a556a07335b663744b50d38062636ad8283f314 |
| 475 data | df-mc/dragonfly 5ac88dcd93d5ae79fca6142d9cd326cb23015d2f |
| 486 data | df-mc/dragonfly 677c8fa1753d662e115e9c610dfe334daed2d265 |
| Native data | df-mc/dragonfly 4c7b5074be94fa83a1cd98e9c752083ad04a6e21 |

The historical packet pools, PlayerAuthInput, StartGame, LevelChunk, SubChunk and
item registry boundaries were compared directly to the native commit. Mojang's
protocol docs at 475bd72ed89036af4eb18426774ef3b953de7603, the r18/r18_u1 changelogs,
and the SubChunk request guide corroborate the historical transition. The
protocol-475 request sentinel is `UINT32_MAX`; 2193 represents this as an optional
SubChunkLimit. A populated byte oracle tests this delta without mutating input.

Native block data are an exact Git blob, SHA256-checked by the loader. The new
snapshot is needed only because standalone consumers do not own a live Dragonfly
registry. Existing historical registries remain shared. Network hashes are
canonical sorted, typed little-endian NBT FNV-1a32; dense ordering uses stable
name-only FNV-1 uint64 ordering. Sparse network IDs are normalized independently
of direct native-to-target semantic conversion.

## Ownership

Create an Adapter per connection. `WireProtocol` handles layout conversion;
`BindUpstream` freezes the original StartGame/ItemRegistry snapshot;
`Translate` handles semantics and chunks. Drive Translate/ClientReplies from one
event loop. The connection's codec pointer changes atomically at registry bind.
The original bootstrap, metadata, input flags and nested mutable optionals are
preserved. A new connection/transfer requires a fresh Adapter.

Chunk storage is decoded, semantically mapped, deduplicated, repacked and encoded.
Biome reuse markers are expanded. Negative sub-chunk Y is preserved. Protocol
475 splits batched responses and expands SuccessAllAir into a real empty v9
payload. Blob cache must be disabled; cached payloads fail explicitly.

Live unmappable blocks/items fail explicitly instead of using implicit air or
empty-item fallbacks. Recipes/creative entries use the existing explicit filter
and preserve surviving selection IDs. Stack request IDs and net IDs are retained;
overflow/rejected requests get failure responses and inventory snapshots. New
movement edge flags are derived from observed legacy input; correction ticks are
retained. Legacy TickSync is answered using the upstream world tick. Unsupported
target-only actions fail; known native-only wire exclusions are counted by ID.

## Evidence and limits

Go tests/vet and Linux race checks cover both families, server-registry isolation,
hash/ordinal round trips, typed state hashes, palette repacking/reuse, negative Y,
clock/input behavior, and nested input immutability. Consumer integration tests
exercise encrypted RakNet v10 to native v11, resource negotiation, full bootstrap,
chunks, inventories and transfers between different item registries.

These are automated clients, not Mojang-client play validation. Real-client
validation for all six builds on Zeqa and Eliagic is pending. Do not advertise
stable support until that matrix passes. Custom dimension layouts, required pack
content unsupported by 1.18, and content without verified mappings remain explicit
compatibility limits. Native block-entity NBT is retained, not generally downgraded.
