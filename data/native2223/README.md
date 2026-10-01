# Minecraft 1.26.60.29 Preview source bundle

This directory preserves published version data for protocol 2223. It is a raw
source acquisition bundle, not an enabled Dragonfly registry or a historical
adapter. No BDS process, mod, injector or protocol dumper was run to acquire it.

The source is axolotl-pm/BedrockData commit
`5f3cd976c61e171876c93419478fbcd9e46894dc`, from `bedrock-1.26.60`.
Its protocol metadata declares 1.26.60.29 and protocol 2223. The original
`beta=false` metadata is preserved; it does not change the target's Preview
channel. The latest source commit restores education content. Exact world and
experiment settings, and the publisher's BDS binary hash, are not attested.

`raw/` contains all published JSON, NBT and binary data, plus its original README
and CC0-1.0 license. Maintenance scripts and Composer packaging are excluded.
Every source file matches its immutable upstream Git blob and size.
`manifest.json` records both the Git blob ID and SHA256 of every preserved file.
Git attributes keep their bytes unchanged on Windows and in module archives.

`bds-metadata.json` is copied from EndstoneMC/bedrock-server-data commit
`b107f380e8b1807430d09c1e82714e905469e6f3`, at
`preview/1.26.60-preview.29/metadata.json`. It records official Windows and Linux
BDS archive URLs and expected SHA256 values. These are catalog cross-references;
the archives were not downloaded, and they do not prove which binary the data
publisher used.

## Validation

`validation.json` records offline checks using the branch's pinned gophertunnel
NBT codec at `34adb6d`:

- 23,960 typed block states, with unique identifier/property combinations;
- 2,080 typed item entries, with unique signed runtime IDs;
- complete standalone NBT decoding and type/value round trips;
- 2,538 embedded LittleEndian NBT compounds from JSON, including item components
  and recipe block states, with no unread bytes and type/value round trips;
- source JSON syntax and finite-number checks.

Block properties retain their byte, int32 and string NBT types. The palette's
original order is preserved; its indexes are not asserted to be hashed native
runtime IDs. Item IDs are version-local: the source has air `-158`, stone `1`,
poplar log `-1132` and shield `358`. Empty-stack wire semantics are a separate
boundary and must not be inferred from the air registry entry.

Integrity checks establish that this bundle matches the published source.
They do not establish independent BDS provenance or actual-client correctness.

## Remaining integration

Convert source records into the exact native Dragonfly loader schemas before
replacing live assets. In particular, decode component/stack/block-state NBT using
its declared encoding, preserve typed properties and signed item IDs, assign
creative group indexes explicitly, and retain shaped/asymmetric, shapeless,
shulker-box, education, smithing and potion categories.

Native registry hash/order verification, historical-to-native mapping, consumer
pins, and real-client chunk/inventory/recipe/pack validation remain pending.
Acquiring this bundle does not enable any protocol or listener.
