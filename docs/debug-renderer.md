# Debug renderer compatibility

The native 2193 model uses optional `DebugMarkerData` and the packed-colour codec corrected by
[gophertunnel 98c3d1f](https://github.com/Sandertv/gophertunnel/commit/98c3d1f3438820b478eb5495c6ff9fb99fa758c5)
and [54e03d9](https://github.com/Sandertv/gophertunnel/commit/54e03d96befbea523f21295b9c9a6a0441906043).
[Mojang's 2193 schema](https://github.com/Mojang/bedrock-protocol-docs/blob/475bd72ed89036af4eb18426774ef3b953de7603/json/DebugRendererData.json)
confirms the string action and optional marker. The native version and registries remain unchanged.

`internal/packetconv/debug_renderer.go` owns the integer-action historical codec used by the 1.18 and
older 1.21 adapters. Exact sources are gophertunnel `c40bf828`, `2cb1e399`, `268adeb5`, `ecff04b7`,
`49e707e` and `bf05a1a`: action 1 clears, action 2 adds a cube; the cube has four float32 colour channels
and no optional-presence byte. The 1.18 duration is signed; the 1.21 duration is unsigned.
Incoming colours round to the nearest native eight-bit channel. Invalid channels and negative signed
durations are rejected. Writers preserve the original packet. Protocol 419 continues to omit packet 164.

`internal/packetconv/debug_renderer_test.go` checks fixed wire oracles across every advertised family,
clear/add semantics, source immutability, exact consumption and invalid inputs. The existing
`protocols/v1_16_100/transport_integration_test.go` gates encrypted Login-first bootstrap through spawn;
both modern and Login-first dialers must expect resource-pack info before sending Login.

The protocol-419 item snapshot is the exact LF JSON from Tedac commit
`5d16de7f9a4e0270e9c3336299564f937c263f8b`, Git blob `24b07f8893222833d34b20c2be851bdde5b02375`.
Its source SHA256 is recorded in `data/v419/manifest.yaml`; `.gitattributes` preserves those bytes on
Windows and in module downloads. The earlier checksum described a CRLF checkout, not the source blob.
