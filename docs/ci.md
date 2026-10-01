# Continuous integration

CI runs on pushes, pull requests and manual dispatch, on every branch. Jobs are
bounded by timeouts and stale runs on the same ref are cancelled. Actions are
pinned to immutable commits and checkout credentials are not persisted.

The primary workflow validates:

- Linux, Windows and macOS with the minimum Go 1.25 patch line and current stable Go;
- test, vet and build for every tracked Go module, including the historical oracle modules;
- Linux race tests, root coverage and pinned gophertunnel transport/session regressions;
- gofmt, module checksum integrity and a clean `go mod tidy -diff` result;
- golangci-lint govet, staticcheck, ineffassign and unused checks;
- GitHub Actions syntax through actionlint;
- independent oracle fixture regeneration/comparison, plus snapshot and semantic mapping tests;
- Linux, Windows and Darwin arm64 compilation;
- reachable vulnerability checks for the maintained root module, including tests.

The old oracle modules intentionally pin historical dependencies as wire references.
They are compiled/tested/linted, but the production vulnerability gate targets the
maintained library rather than demanding current dependencies in historical parsers.
No source vulnerability is ignored in the maintained root scan.

Lint exceptions are scoped to historical protocol paths: deprecated
ClientCheatAbility wire types, explicit base-IO calls and retained unused reference
scaffolding. Other staticcheck diagnostics remain blocking. Cleaning up that
historical scaffolding is outside this CI change.

`CI required` is one stable status name that fails if any primary job fails,
is cancelled or is skipped. Branch protection configuration is not changed by
adding the workflow. CodeQL is a separate Go analysis workflow and uploads findings
with a job-scoped security-events permission. Repository code-scanning availability
and actual GitHub execution must be checked after the workflow is published.

Coverage is reported and stored as a GitHub artifact for 14 days; no arbitrary
coverage threshold is imposed. The oracle checker generates only synthetic wire
fixtures in temporary directories. It never runs BDS, extracts registries or
overwrites checked-in fixtures.

Oracle packet bytes must match exactly, with one explicit ordering exception:
the pinned v419 AddActor writer iterates its metadata map in arbitrary order.
That populated fixture is compared as a complete decoded packet using the
independent historical reader, including full byte-consumption checks. Every
decoded value still has to match. Skipped zero-value packet keys also have to
match, while toolchain-specific panic wording may differ.

Run portable checks locally with Python 3.9+ and Go:

```text
python tools/ci/check.py format
python tools/ci/check.py modules
python tools/ci/check.py test
python tools/ci/check.py oracles
python tools/ci/check.py portability
```

`lint` also needs the pinned golangci-lint executable. `race` and `transport`
require a working race/cgo toolchain. Module discovery uses tracked go.mod files;
new oracle modules must also join `tools/ci/oracles.json`, otherwise reproducibility
validation fails.

For dual public transports, see [the proposed RakNet/NetherNet admission design](dual-transport.md).
Automated library checks do not establish actual-client or deployment support.
