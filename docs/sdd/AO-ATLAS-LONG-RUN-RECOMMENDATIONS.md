# AO Atlas Long-Run Recommendation Waves

Use this runbook when an operator asks for a 2-3 hour Atlas-owned wave or when a Mission/Atlas prompt returns after a short batch.

## Route Choice

Use AO Mission when the request needs a lease, checkpoint policy, route history, or a final-response gate. Mission owns the supervisor contract and should keep final status denied while ready nodes or exact next actions remain.

Use AO Atlas when the work needs context-heavy sequencing, workgraph state, recommendation imports, or cross-artifact reconciliation. Atlas owns the workgraph and should expose exactly one executable node at a time.

Use AO Foundry for one bounded implementation node at a time. Atlas may emit a fixture-only Foundry import, but Foundry owns the implementation evidence for the active node.

Use AO Blueprint only when new requirements, authorization, candidate rules, or scope approval are missing. Do not route ready Atlas-owned tasks through Blueprint just to add ceremony.

Ready Atlas-owned tasks skip Blueprint because the recommendation workgraph
already contains the bounded task, owner, write scope, safety limits, required
gates, and digest-bound source evidence. Sending that node back to Blueprint
would re-ask for requirements that are already present and can stale the route
history. Atlas should emit the Foundry import for the single executable node,
Foundry should execute that bounded implementation node, and Blueprint should be
re-entered only if the node lacks requirements, authorization, candidate rules,
or scope approval.

Use AO Command for compact readback: completed nodes, ready nodes, blockers, exact next action, and whether final response is allowed.

Use AO Promoter only for a promotion or no-promotion verdict. A recommendation wave does not promote mutation authority by itself.

## Lease Defaults

For a long-run recommendation wave, use these defaults unless the operator gives stricter values:

- `min_nodes`: 30
- `min_minutes`: 120
- `max_minutes`: 180
- `continue_if_fast_target`: 40
- `return_only_when`: all generated nodes are done, the minimum is met with no ready work, or a true hard blocker remains
- `checkpoint_policy`: after each node or timed interval

Generated recommendations are not completed work. A node counts as completed only after the workgraph advances through a completed run-link with node gate, candidate, rollback, implementation, tests, verification, public-safety, Promoter, Command, Foundry import, and checkpoint evidence. Completing every generated node is still not enough to close a 2-3 hour lease: the authoritative readback must also record `started_at`, `completed_at`, `elapsed_minutes`, `min_minutes_met=true`, and `lease_time_status=minimum_minutes_met`.

`lease-start.json` is the durable start marker for the wave. Preserve it across
checkpoint and resume commands so Atlas does not reset the clock when a prompt
continues in a later session. If a wave completes all nodes before
`elapsed_minutes` reaches `supervisor.min_minutes`, keep final response denied
and generate the next useful Atlas-owned wave. Synthetic node evidence can prove
ledger mechanics, but it cannot satisfy a 2-3 hour lease by itself.

The recommendation readback exposes `return_gate_status` and `checkpoint_count`
for compact status checks. `return_gate_status` names the active final-response
gate, such as `blocked_ready_nodes_remain`,
`blocked_minimum_minutes_unmet`, or `final_response_allowed`.
`checkpoint_count` mirrors the completed-node count that has checkpoint evidence.

## Double-Size Operator Requests

When the operator says the last wave was too short, double-size means the next
Atlas recommendation wave must target the normal 2-3 hour lease, not a 90-minute
compatibility run. Use a 30-node minimum, a 40-node continue-if-fast target, and
120-180 minutes unless the operator gives stricter numbers. A 20-node or
90-minute wave is too small for this path.

Write the next prompt like a long-run correction prompt: include current state,
problem, goal, minimum work budget, safety boundaries, required routing,
per-node evidence, verification, final-report gates, and the exact next action.
The prompt should say that Atlas owns sequencing, Foundry owns one bounded
implementation node at a time, Blueprint is used only for missing requirements
or authorization, and Command/Promoter/Sentinel readbacks remain evidence gates.

Do not final after one PR, one CI pass, one Foundry import, one route decision,
one evidence artifact, or one short batch. If the first 15 nodes finish quickly
and ready nodes remain, continue into the next useful nodes instead of closing.
If all generated nodes finish before `elapsed_minutes` reaches 120, generate the
next useful Atlas-owned wave and preserve the original lease evidence. A final
answer is allowed only when the authoritative readback has
`final_response_allowed=true`, no ready nodes remain, and Command, Foundry,
Promoter, and reconciliation artifacts agree.

## Execution Pattern

1. Confirm the repo is on a branch, clean enough for the wave, and not mutating `main` directly.
2. Import the Mission Feature Depth Recommendations into an Atlas recommendation wave.
3. Preserve `lease-start.json`; it is the authoritative `started_at` source for resume.
4. Inspect the readback. If `final_response_allowed` is false, do not return final status.
5. Emit a Foundry import for the first executable node.
6. Record the required evidence bundle for that node.
7. Attach a completed run-link to the node.
8. Run `atlas mission recommendations complete-node` with `--lease-start` and `--out-checkpoint-readback`.
9. Use `atlas mission recommendations resume` when continuing from a checkpoint; it carries the original `started_at` into recommendation, execution, Command, Promoter, and Foundry readbacks.
10. Repeat from the next executable node until the wave is complete and the lease is met, or a hard blocker remains.
11. Run local verification, open a PR, wait for CI, merge, sync, and remove `codex/*` branches when remote lifecycle is available.

## Commands

```sh
atlas mission recommendations import \
  --recommendations examples/valid/ao-mission/feature-depth-recommendations.json \
  --target-instance demo-stack \
  --min-tasks 30 \
  --node-budget 40 \
  --min-minutes 120 \
  --max-minutes 180 \
  --continue-if-fast-target 40 \
  --started-at 2026-07-04T08:00:00-07:00 \
  --out docs/evidence/<wave>

atlas foundry import \
  --workgraph docs/evidence/<wave>/recommendation-workgraph.json \
  --instance examples/valid/stack-instance.json \
  --node mission-recommendation-next-01 \
  --json

atlas mission recommendations complete-node \
  --wave docs/evidence/<wave>/recommendation-wave.json \
  --workgraph docs/evidence/<wave>/recommendation-workgraph.json \
  --run-link docs/evidence/<wave>/nodes/mission-recommendation-next-01/run-link.json \
  --expected-node mission-recommendation-next-01 \
  --evidence-root . \
  --readback-evidence-root docs/evidence/<wave> \
  --lease-start docs/evidence/<wave>/lease-start.json \
  --completed-at 2026-07-04T08:17:00-07:00 \
  --out-workgraph docs/evidence/<wave>/nodes/mission-recommendation-next-01/workgraph-after.json \
  --out-readback docs/evidence/<wave>/nodes/mission-recommendation-next-01/recommendation-readback-after.json \
  --out-execution-readback docs/evidence/<wave>/nodes/mission-recommendation-next-01/execution-readback-after.json \
  --out-checkpoint-readback docs/evidence/<wave>/nodes/mission-recommendation-next-01/checkpoint-readback-after.json

atlas mission recommendations resume \
  --wave docs/evidence/<wave>/recommendation-wave.json \
  --workgraph docs/evidence/<wave>/nodes/mission-recommendation-next-01/workgraph-after.json \
  --lease-start docs/evidence/<wave>/lease-start.json \
  --completed-at 2026-07-04T08:25:00-07:00 \
  --evidence-root docs/evidence/<wave> \
  --out-readback docs/evidence/<wave>/recommendation-readback.json \
  --out-execution-readback docs/evidence/<wave>/execution-readback.json \
  --out-command-readback docs/evidence/<wave>/command-readback.json \
  --out-promoter-readback docs/evidence/<wave>/promoter-readback.json \
  --out-foundry-rollup docs/evidence/<wave>/foundry-rollup.json \
  --out-reconciliation-packet docs/evidence/<wave>/reconciliation-packet.json

atlas mission recommendations readback \
  --wave docs/evidence/<wave>/recommendation-wave.json \
  --workgraph docs/evidence/<wave>/recommendation-workgraph.json \
  --evidence-root docs/evidence/<wave> \
  --started-at 2026-07-04T07:20:00-07:00 \
  --completed-at 2026-07-04T09:20:00-07:00 \
  --elapsed-minutes 120 \
  --lease-timing-mode actual \
  --out docs/evidence/<wave>/recommendation-readback.json
```

## Final Report Gate

A final response is allowed only when the authoritative recommendation readback agrees with the execution readback. The counts must match, `ready_nodes` must be zero for a completed wave, `elapsed_minutes` must meet or exceed `supervisor.min_minutes`, `min_minutes_met` must be true, and `final_response_allowed` must be true. The reconciliation packet must also report `artifacts_agree=true` across the recommendation readback, Command readback, Promoter readback, and Foundry rollup.

If ready nodes remain, report the exact next executable node instead of a final answer. If all nodes are complete but the time lease is missing or short, report the exact timing evidence to record or generate the next useful Atlas recommendation wave instead of a final answer.
