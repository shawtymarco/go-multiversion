# df-multiversion

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
- Minecraft 1.21.110-1.21.114 / protocol 844: direct wire adapter based on
  `bf05a1a`, historical data from Dragonfly `1872684e`, semantic item/block
  mapping against the current native registries, and version-9 palette mapping
  through Dragonfly's existing pre-hash hook. Only this 1.21.11x family is in
  scope; 1.21.120/protocol 859 and other 1.21.x families are not implied.

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
