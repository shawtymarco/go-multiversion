# df-multiversion

Minecraft Bedrock protocol compatibility adapters for Dragonfly and other
gophertunnel-based servers.

The library keeps the server model on gophertunnel's latest native protocol and
converts older releases at the network boundary.

## Supported versions

| Minecraft version | Protocol | Role |
| --- | ---: | --- |
| 1.26.45 | 2169 | Native gophertunnel protocol |
| 1.26.44 | 2168 | `v1_26_44` compatibility adapter |

The 1.26.44 adapter is based on gophertunnel commit
`8a2b1f7939b051227fbef9d05c0f5b1d96ac2993`. Native 1.26.45 is based on
`7f058e5ddc393eaa0480dae338c5eee2feb323e6`. The adapter currently changes only
the protocol identity and the `SetScore` wire layout; it does not rewrite
resource-pack negotiation packets.

## Usage

```go
import multiversion "github.com/shawtymarco/df-multiversion"

legacyProtocols := multiversion.Protocols()
```

`Protocols` returns non-native protocols only. Pass them to the consumer's
accepted-protocol configuration alongside its native/default protocol as
required by that consumer.
