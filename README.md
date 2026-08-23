# df-multiversion

Minecraft Bedrock protocol compatibility adapters for Dragonfly and other
gophertunnel-based servers.

The library keeps the server model on gophertunnel's latest native protocol and
converts older releases at the network boundary.

## Supported versions

| Minecraft version | Protocol | Role |
| --- | ---: | --- |
| 1.26.45 | 2169 | Native gophertunnel protocol |
| 1.26.44 | 2168 | `v1_26_44` double-optional `SetScore` layout |
| 1.26.40-1.26.43 | 2168 | `v1_26_44` shared native `SetScore` layout |

The 1.26.40-1.26.43 layout is based on gophertunnel commit
`06952754f1a41e01f1d2a2f5e15414c9504f8902`, and the 1.26.44 layout is based
on `8a2b1f7939b051227fbef9d05c0f5b1d96ac2993`. Native 1.26.45 is based on
`7f058e5ddc393eaa0480dae338c5eee2feb323e6`. Releases 1.26.40 through 1.26.44
reuse protocol ID 2168, so the adapter inspects the login `GameVersion` when
choosing the outgoing `SetScore` layout. It does not rewrite resource-pack
negotiation packets or maintain separate block, item, recipe, or chunk data.

## Development adapters

| Minecraft version | Protocol | Status |
| --- | ---: | --- |
| 1.26.30-1.26.34, 1.26.36 | 1001 | Wire, registry, gameplay packet, and pre-hash chunk conversion complete; real-client release matrix pending |

`V1_26_30()` exposes the wire-only protocol-1001 adapter for focused tests.
Production consumers use `ProtocolsWithRegistries` so block and item mappings
are validated before the listener advertises protocol 1001. The adapter stays
absent from the parameterless `Protocols()` to prevent an unconfigured server
from reusing native runtime IDs. The wire
implementation is based on gophertunnel `0a2ecd5633ea1466ff97f6d4718df66ec14d054f`
and converts directly to the native model at `7f058e5ddc393eaa0480dae338c5eee2feb323e6`.
Its exact wire audit is recorded in `versions/1.26.3x-wire.md`.

## Usage

```go
import multiversion "github.com/shawtymarco/df-multiversion"

legacyProtocols := multiversion.Protocols()
```

`Protocols` returns non-native protocols only. Pass them to the consumer's
accepted-protocol configuration alongside its native/default protocol as
required by that consumer.

For protocol 1001, pass the finalised native block registry and the exact
native item entries to `ProtocolsWithRegistries`. Dragonfly consumers should
do this through its post-finalisation `AcceptedProtocolsProvider` hook.
