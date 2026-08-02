# AO Atlas Acceptance Gates

Production readiness requires:

- `go test ./...`
- `go vet ./...`
- `go build ./cmd/atlas`
- JSON syntax validation for all schemas and fixtures
- CLI validation for all valid fixtures
- CLI rejection for invalid fixtures
- public-safety scan over tracked and pending files
- `git diff --check`

The readiness script prints `score=100/100` only after every gate passes.

## Developer Quality Gates

The root `ao-quality-gates.json` manifest provides deterministic local
feedback through the AO2 quality runner:

- `commit` checks the exact staged tree with `git diff --cached --check`;
- `push` runs the Atlas package tests against exact outgoing commits; and
- `full` runs all Go tests, vet, and a build against the exact source head.

Atlas owns these argument vectors and their timeouts. Fast gates are
network-disabled and all levels must remain source-non-mutating. Generated
build output is confined to `target/quality-gates`.

These local gates do not replace the platform-native production-readiness
scripts in hosted CI. In particular, the Unix and Windows scripts retain the
authoritative contract and fixture checks for their respective hosts.
Cross-repository roundtrip verification remains conditional on an explicitly
in-scope synchronized AO Foundry checkout.

## Release Candidate Checkpoint

AO Atlas v0.1 is eligible for a stable tag or release candidate only when:

- `scripts/production-readiness.sh` reports `score=100/100`;
- `scripts/atlas-foundry-roundtrip-smoke.sh` reports `status=ready`;
- the roundtrip summary records `schedules_work=false`, `executes_work=false`,
  and `approves_work=false`;
- instance doctor reports preserve `schedules_work=false`,
  `executes_work=false`, and `approves_work=false`, and fail closed on registry
  parity mismatch, non-ignored instance state, copied stack roots, or Atlas
  authority claims;
- AO Foundry validates the emitted `ao.atlas.foundry-import.v0.1` packet;
- Foundry import generation records source artifact digests, preserves
  context-pack refs, supports single-node and multi-node ready imports, and
  fails closed on blocked nodes, incomplete dependencies, missing context
  packs, unsafe paths, or authority claims;
- AO Foundry emits `ao.foundry.atlas-readback.v0.1` for the completed Atlas
  run link;
- both AO Atlas and AO Foundry are clean on synced `main`.

The checkpoint is a readiness decision, not a release action. Tags, releases,
uploads, or publication require explicit operator intent.
