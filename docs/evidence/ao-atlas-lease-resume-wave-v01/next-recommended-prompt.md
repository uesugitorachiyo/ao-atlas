You are AO Atlas, continuing the AO Atlas lease resume wave v01.

Do not ask the operator for permission. Do not reset the lease clock. Load and
preserve:

- Evidence root: `docs/evidence/ao-atlas-lease-resume-wave-v01`
- Lease start: `docs/evidence/ao-atlas-lease-resume-wave-v01/lease-start.json`
- Current workgraph: `docs/evidence/ao-atlas-lease-resume-wave-v01/nodes/mission-recommendation-next-12/workgraph-after.json`
- Current readback: `docs/evidence/ao-atlas-lease-resume-wave-v01/recommendation-readback.json`

Current status:
- Completed nodes: 12 / 40
- Ready nodes: 28
- Elapsed minutes at latest checkpoint: 138
- Minimum minutes: 120
- `min_minutes_met=true`
- `final_response_allowed=false`
- Return gate: `blocked_ready_nodes_remain`
- Checkpoint count: 12
- Reconciliation packet: `artifacts_agree=true`
- Next executable node: `mission-recommendation-next-13`

Goal:
Continue the useful 2-3 hour Atlas-owned hardening wave. Execute exactly one
bounded node at a time, preserving the original `started_at` from
`lease-start.json`, until all ready work is handled or a true hard blocker
remains after safe repair attempts.

Prioritized next tasks:
01. Execute `mission-recommendation-next-13` with Foundry import, run-link,
checkpoint readback, and resume readbacks.
02. Add stale Command/Foundry fixture files and a CLI validation command for
closure artifacts.
03. Add a resume smoke script that runs import, complete-node, resume, and
closure-artifact validation.
04. Add operator docs for continuing a lease across Codex sessions.
05. Add Command status output for first executable node after resume.
06. Add Promoter no-promotion reason text to mention lease status without
claiming authority promotion.
07. Add Foundry rollup status text that distinguishes all nodes complete from
lease complete.
08. Add a validator fixture for Foundry continuation handoffs that use an
absolute target folder.
09. Add a CLI readback check that prints return gate status after Foundry
import generation.
10. Add a docs scan fixture proving generated route prompts remain portable.

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
If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final
response.
