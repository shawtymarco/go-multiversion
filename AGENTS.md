# go-multiversion

Minecraft Bedrock protocol compatibility library for Dragonfly and other
gophertunnel-based servers.

## Scope

- Keep gophertunnel's current protocol as the native server model.
- Implement older client protocols under `protocols/v<version>/`.
- Own version-specific packet pools, wire layouts, and bidirectional packet
  conversion.
- Add block, item, recipe, creative inventory, or chunk data only when the
  target version actually differs.
- Do not add Dragonfly gameplay logic, server-specific policy, deployment
  configuration, or local machine paths.

## Protocol work

- Compare the exact historical gophertunnel commit directly with the current
  native commit before editing an adapter.
- Check Mojang's protocol changelog and release-specific protocol reference
  when identifying packet and supporting-type changes.
- Convert every supported version directly to and from the native model.
- Keep unchanged packets and data shared with current gophertunnel.
- Do not add speculative conversions. Document the source commits and the
  exact compatibility difference being implemented.

The current compatibility family is:

- Minecraft 1.26.45 / protocol 2169: native gophertunnel at `7f058e5`.
- Minecraft 1.26.44 / protocol 2168: double-optional `SetScore` adapter based
  on `8a2b1f7`.
- Minecraft 1.26.40-1.26.43 / protocol 2168: shared-layout adapter based on
  `0695275`, selected using the login `GameVersion`.
- Minecraft 1.26.30-1.26.34 and 1.26.36 / protocol 1001: wire adapter based on
  `0a2ecd5`, historical data from Dragonfly `705b8460`, and direct native data
  mapping against `ba86ce95`. It is enabled only through the registry-aware
  constructor; the parameterless `Protocols()` remains safe for unconfigured
  consumers.
- Minecraft 1.26.20, 1.26.21, and 1.26.23 / protocol 975: direct wire adapter
  based on `de090ae`, historical data from Dragonfly `7c304285`, semantic
  item/block mapping against the current native registries, numeric-to-string
  sound conversion, and the same pre-hash palette hook. Preview builds and
  unlisted 1.26.2x releases are not implied.
- Minecraft 1.26.10-1.26.14 / protocol 944: direct wire adapter based on
  `165bd86b`, historical data from Dragonfly `16359459`, semantic item/block
  mapping against the current native registries, the pre-hash palette hook,
  and the exact target-era furnace recipe payload. The target-only
  `editor:map_marker_spawn_egg` remains advertised but is rejected
  serverbound. Protocol 924, previews, and other 1.26.x release trains are not
  implied.
- Minecraft 1.26.0-1.26.3 / protocol 924: direct wire adapter based on
  `9c440d5f`, historical data from Dragonfly `3e4f0bbe`, unsigned-Y block
  positions, legacy camera and StartGame layouts, exact furnace data, semantic
  registry mapping, and the generic pre-hash palette hook. Preview builds and
  protocol 898 or 944 releases are not implied.
- Minecraft 1.21.130-1.21.132 / protocol 898: direct wire adapter based on
  `c839e607`, historical data from Dragonfly `8d1311b3`, legacy Text and
  BookEdit layouts, target-specific smithing audit data, semantic registry
  mapping, and the generic pre-hash palette hook. Preview builds and earlier
  1.21.x release trains are not implied.
- Minecraft 1.21.110-1.21.114 / protocol 844: direct wire adapter based on
  `bf05a1a`, historical data from Dragonfly `1872684e`, semantic item/block
  mapping against the current native registries, and version-9 palette mapping
  through Dragonfly's existing pre-hash hook. Only this 1.21.11x family is in
  scope; 1.21.120/protocol 859 and other 1.21.x families are not implied.
- Minecraft 1.21.100-1.21.102 / protocol 827: direct wire adapter based on
  `49e707e` plus the applicable later wire corrections, historical data from
  Dragonfly `93c84812`, semantic item/block mapping against the current native
  registries, and the same version-9 pre-hash palette hook. Preview builds and
  unlisted 1.21.x families are not implied.
- Minecraft 1.21.50-1.21.51 / protocol 766: direct adapter based on
  `ecff04b7`, historical data from Dragonfly `800ee178`, ungrouped creative
  content, StartGame item registry, raw biome NBT, and version-9 pre-hash
  palette mapping.
- Minecraft 1.21.40, 1.21.41, 1.21.43, and 1.21.44 / protocol 748: direct
  adapter based on `268adeb5`, historical data from Dragonfly `206bf97c`,
  uint64 input flags, legacy resource-pack UUID strings, and version-9
  pre-hash palette mapping. No stable 1.21.42 release is implied.
- Minecraft 1.18.10-1.18.12 / protocol 486: direct semantic adapter based on
  gophertunnel `2cb1e399`, historical registries from Dragonfly `677c8fa1`,
  RakNet v10/Login-first legacy flate negotiation in the gophertunnel fork,
  numeric recipe/item conversion, ungrouped creative content, and version-9
  chunk palettes through the existing pre-hash hook. Protocol 475, protocol
  503, previews, and other 1.18.x release trains are not implied.
- Minecraft 1.18.0-1.18.2 / protocol 475: direct semantic adapter based on
  gophertunnel `c40bf828`, historical registries from Dragonfly `5ac88dcd`,
  RakNet v10/Login-first legacy flate, single-response SubChunk conversion,
  and version-9 palettes through the pre-hash hook.

## Verification

After Go changes, run:

```text
gofmt on changed Go files
go test ./...
go vet ./...
git diff --check
git status --short
```

Add focused round-trip tests for every changed packet layout. Test both
conversion directions and confirm that conversions do not mutate their input
packets unexpectedly.

## Git

- Preserve unrelated user changes and stage only files changed for the task.
- Commit completed work before closing the task.
- Commit messages must always be written in English.
- Use conventional prefixes such as `feat:`, `fix:`, `test:`, `docs:`,
  `refactor:`, or `chore:`.
- Do not push unless the user explicitly requests it.
