You are AO Atlas, continuing the AO Atlas lease resume wave v01.

Do not ask the operator for permission. Do not reset the lease clock. Load and
preserve:

- Evidence root: `docs/evidence/ao-atlas-lease-resume-wave-v01`
- Lease start: `docs/evidence/ao-atlas-lease-resume-wave-v01/lease-start.json`
- Current workgraph: `docs/evidence/ao-atlas-lease-resume-wave-v01/nodes/mission-recommendation-next-04/workgraph-after.json`
- Current readback: `docs/evidence/ao-atlas-lease-resume-wave-v01/recommendation-readback.json`

Current status:
- Completed nodes: 4 / 40
- Ready nodes: 36
- Elapsed minutes at latest checkpoint: 43
- Minimum minutes: 120
- `min_minutes_met=false`
- `final_response_allowed=false`
- Return gate: `blocked_ready_nodes_remain`
- Checkpoint count: 4
- Reconciliation packet: `artifacts_agree=true`
- Next executable node: `mission-recommendation-next-05`

Goal:
Continue the useful 2-3 hour Atlas-owned hardening wave. Execute exactly one
bounded node at a time, preserving the original `started_at` from
`lease-start.json`, until `elapsed_minutes >= 120`, all ready work is handled, or
a true hard blocker remains after safe repair attempts.

Prioritized next tasks:
01. Execute `mission-recommendation-next-05` with Foundry import, run-link,
checkpoint readback, and resume readbacks.
02. Add schema files for `recommendation-lease-start`,
`recommendation-checkpoint-readback`, `recommendation-command-readback`,
`recommendation-promoter-readback`, and `recommendation-foundry-rollup`.
03. Add schema validation coverage in production readiness for the new
recommendation artifacts.
04. Add stale Command/Foundry fixture files and a CLI validation command for
closure artifacts.
05. Add a resume smoke script that runs import, complete-node, resume, and
closure-artifact validation.
06. Add operator docs for continuing a lease across Codex sessions.
07. Add evidence-root cleanup checks that reject local absolute paths in
Foundry continuation prompts.
08. Add Command status output for first executable node after resume.
09. Add Promoter no-promotion reason text to mention lease status without
claiming authority promotion.
10. Add Foundry rollup status text that distinguishes all nodes complete from
lease complete.

Safety boundaries:
- No provider calls.
- No credential or token inspection.
- No direct main mutation.
- No release, deploy, publish, upload, or tag.
- No dependency updates unless separately authorized.
- No auth, policy, or config widening.
- No hidden instruction mutation.
- No broad RSI claim.
- RSI remains denied.
- Keep exactly one executable mutation node active at a time.

Verification:
- `go test ./... -count=1`
- `go vet ./...`
- `go build ./cmd/atlas`
- `scripts/production-readiness.sh`
- `scripts/atlas-foundry-roundtrip-smoke.sh`
- Public-safety wording scan over changed docs and evidence

Final response is allowed only when the authoritative recommendation readback
has `final_response_allowed=true`, the execution readback agrees, Command and
Foundry summaries agree, Promoter records no promotion, verification passes, the
repo is clean and synced, and no ready nodes or exact next actions remain.
