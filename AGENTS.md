# AO Atlas Agent Instructions

## Status And Role

AO Atlas is the active compiler for authorized oversized objectives. It owns bounded workgraphs, context packs, candidate selection, single-node Foundry import material, and readback needed to resume compilation.

Atlas consumes Blueprint authorization and Mission objectives and readbacks. It does not authorize scope, schedule or execute work, approve actions, mutate sibling repositories, replace Foundry run-link evidence, or advance Mission state.

## Sources Of Truth

- [docs/sdd/AO-ATLAS-PRD.md](docs/sdd/AO-ATLAS-PRD.md), [docs/sdd/AO-ATLAS-ARCHITECTURE.md](docs/sdd/AO-ATLAS-ARCHITECTURE.md), and [docs/sdd/AO-ATLAS-CONTRACTS.md](docs/sdd/AO-ATLAS-CONTRACTS.md) define product and contract boundaries.
- [docs/sdd/AO-ATLAS-WORKGRAPH.md](docs/sdd/AO-ATLAS-WORKGRAPH.md), [docs/sdd/AO-ATLAS-CONTEXT-PACKS.md](docs/sdd/AO-ATLAS-CONTEXT-PACKS.md), and [docs/sdd/AO-ATLAS-FOUNDRY-HANDOFF.md](docs/sdd/AO-ATLAS-FOUNDRY-HANDOFF.md) own compilation and handoff semantics.
- [docs/sdd/canonical-terminal-index.md](docs/sdd/canonical-terminal-index.md) is the current terminal-artifact index. `schemas/` and `internal/atlas/` are authoritative for implemented wire behavior.
- `scripts/production-readiness.sh`, `scripts/production-readiness.ps1`, and [`.github/workflows/ci.yml`](.github/workflows/ci.yml) define the broad verification gate.

## Ownership And Boundaries

- Require exact source heads, repository-relative evidence roots, and declared SHA-256 digests where contracts bind inputs. A mismatch, absent authorization, unknown state, or authority claim must fail closed.
- Preserve historical files under `docs/evidence/` and mappings in `docs/evidence-path-map.json`; never rewrite them to support a current claim or revive and reinterpret an earlier wave.
- Treat `examples/valid/` and `examples/invalid/` as contract fixtures. Change them with their consuming tests and keep rejection cases invalid.
- Keep generated state and output under ignored `.atlas-local/`, `.atlas-state/`, `target/`, `dist/`, or `bin/`; do not hand-edit generated readbacks into source evidence.
- Do not introduce secrets, provider credentials, private paths, account identifiers, or live-provider behavior. Ordinary Atlas operation has no push, tag, release, upload, approval, or sibling-repository mutation authority.
- Release publication is limited to separately authorized exact candidates through `.github/workflows/release-finalize.yml`, the `protected-release` environment, and eligible human review. All other tag, release, and upload activity remains forbidden.
- Live finalization additionally requires matching environment-scoped `AO_ATLAS_PROTECTED_RELEASE_SENTINEL` secret and variable values. This source sentinel fails closed but never replaces external verification that `protected-release` has eligible required reviewers.
- Deployment, other publication, credentialed operation, and direct-main changes require separate explicit authority.

## Working Method

- Change the smallest owned compilation surface and preserve dependency order, single executable-node selection, context provenance, continuation gates, and non-authority flags.
- Add negative coverage for malformed, stale, digest-mismatched, traversal, or over-authority inputs. Do not weaken a rejection fixture to obtain a result.
- Preserve an explicit zero-minute recommendation lease minimum as useful-work mode. Keep target and maximum duration independent, require timing evidence to enforce the maximum, and never pad elapsed time; omitted minima retain the historical default.
- Update this file in the same pull request when durable commands, architecture, ownership, or authority boundaries change.

## Verification

- `ao-quality-gates.json` is Atlas-owned executable command truth for optional
  local commit, push, and full feedback through the AO2 quality runner. Keep
  its argument vectors, path triggers, timeouts, generated paths, and
  non-mutation policy aligned with the commands below.
- Atlas logic or contract changes: `go test ./internal/atlas -count=1`.
- Format relevant Go source with `gofmt -d cmd internal`; run `go test ./... -count=1`, `go vet ./...`, and `go build ./cmd/atlas`.
- Run `scripts/production-readiness.sh` as the full local gate. Run `scripts/atlas-foundry-roundtrip-smoke.sh` only when the cross-repository boundary changes and a synchronized sibling Foundry checkout is explicitly in scope.
- The source-owned quality manifest does not replace hosted
  `production-readiness.sh` and `production-readiness.ps1` checks or make
  optional Git hooks authoritative.
- For instruction changes run `python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-atlas`. Always run `git diff --check`.

## Evidence And Completion

- Record the Atlas source head, input source heads, command exits, and relevant input/output digests. Report skipped, unavailable, or failed checks explicitly.
- A generated workgraph, readiness score, or historical readback is evidence, not authorization to execute, release, deploy, or reinterpret an earlier campaign.
- Completion requires focused and broad gates, green pull-request CI, clean synchronized `main`, and task-branch cleanup.
