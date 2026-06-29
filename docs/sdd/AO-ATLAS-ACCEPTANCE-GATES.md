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

## Release Candidate Checkpoint

AO Atlas v0.1 is eligible for a stable tag or release candidate only when:

- `scripts/production-readiness.sh` reports `score=100/100`;
- `scripts/atlas-foundry-roundtrip-smoke.sh` reports `status=ready`;
- the roundtrip summary records `schedules_work=false`, `executes_work=false`,
  and `approves_work=false`;
- AO Foundry validates the emitted `ao.atlas.foundry-import.v0.1` packet;
- AO Foundry emits `ao.foundry.atlas-readback.v0.1` for the completed Atlas
  run link;
- both AO Atlas and AO Foundry are clean on synced `main`.

The checkpoint is a readiness decision, not a release action. Tags, releases,
uploads, or publication require explicit operator intent.
