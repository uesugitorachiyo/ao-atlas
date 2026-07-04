# AO Atlas Lease Resume Wave v01

Status: completed.

This evidence root records the first post-repair recommendation wave using the
persisted lease-start and resume machinery. The wave starts at
`2026-07-04T08:34:04-07:00`, has a 120-minute minimum, and remains final-denied
until all ready nodes are complete. The final readback now records all 40 nodes
complete, no ready nodes remaining, and `final_response_allowed=true`.

## 2-3 Hour Long-Run Recommendation Wave Evidence

This wave is the durable evidence root for the operator-requested 2-3 hour
Atlas recommendation run. It uses the long-run supervisor defaults: 40 generated
nodes, 30-node minimum, 120-minute minimum lease, 180-minute maximum lease, and
checkpoint policy `after_each_node_or_timed_interval`. The lease start is
preserved in `lease-start.json`; every completed node records a Foundry import,
run-link, checkpoint readback, Command readback, Promoter no-promotion readback,
rollback record, public-safety evidence, tests, and verification output.

The final response gate is evidence-bound. A final answer is denied while ready
nodes or an exact next action remain, even when the minimum lease time is met.
When all 40 nodes complete, the final readback must show zero ready nodes,
`min_minutes_met=true`, `return_gate_status=final_response_allowed`, matching
Command/Foundry/Promoter/reconciliation artifacts, and no authority or RSI
promotion claim.

## Current Readback

- Total nodes: 40
- Completed nodes: 40
- Ready nodes: 0
- Blocked nodes: 0
- Failed nodes: 0
- Elapsed minutes: 491
- Minimum minutes met: true
- Lease health status: `all_generated_nodes_complete`
- Checkpoint freshness status: `fresh_checkpoint_required_after_each_node_or_timed_interval`
- Stale route decision status: `fresh_atlas_supervises_foundry_owns_one_active_node`
- Early-return risk status: `cleared_no_ready_nodes_remain`
- Return gate status: `final_response_allowed`
- Checkpoint count: 40
- Final response allowed: true
- Exact next action: finalize AO Atlas long-run wave with Promoter, Command,
  and public-safety readbacks.
- Exact next action readback: `finalization_ready`, no next executable node,
  `final_response_allowed`, and `final_response_allowed=true`.
- Command readback: completed; `40/40` recommendation nodes complete,
  `lease_time_status=minimum_minutes_met`, and
  `final_response_allowed=true`.
- Promoter readback: no promotion claimed; RSI remains denied.
- Final-response denial tests: ready recommendation workgraphs now reject stale
  `final_response_allowed` return gates, stale final reasons, and exact-next
  actions that do not name the first executable node.
- Final-response allowance tests: completed recommendation workgraphs now
  reject stale allowed-state status, return gate, final reason, and final exact
  next action drift.
- Long-run lease examples: `examples/valid/recommendation-wave-long-run-supervisor.json`
  and `examples/valid/recommendation-lease-start-long-run.json` cover the
  30-node, 120-180 minute, continue-if-fast lease fields.
- Blueprint skip docs: the operator runbook and README now explain that ready
  Atlas-owned tasks skip Blueprint unless requirements, authorization,
  candidate rules, or scope approval are missing.
- Foundry import ownership docs: the operator runbook and README now explain
  that Atlas imports only the current executable node, Foundry owns that
  bounded implementation evidence and run-link, and Atlas resumes from the
  checkpoint before importing the next node.
- Command readback requirements docs: the operator runbook and README now
  explain that Command stays compact and gate-bound, reporting node counts,
  lease status, checkpoint freshness, return gate, exact next action, first
  executable node, and final-response allowance without replacing Atlas
  workgraph or Foundry run-link evidence.
- Generated continuation prompt public-safety scan: production readiness now
  rejects affirmative unsafe wording in generated continuation prompts, includes
  a negative unsafe prompt fixture, and scans Mission recommendation, Blueprint
  import, and direct Foundry import continuation prompts.
- 40-node recommendation workgraph assertion: production readiness now verifies
  the generated recommendation workgraph has exactly 40 ready nodes with stable
  ids, a linear dependency chain, source digest evidence, required gates, safety
  limits, Atlas planning boundary, and no scheduling/execution/approval flags.
- Evidence README long-run section: this file now documents the 40-node,
  120-180 minute lease, checkpoint evidence requirements, final-response gate,
  cross-artifact agreement, no-promotion boundary, and RSI-denied boundary.
- Final execution readback: recommendation, execution, Command, Promoter,
  Foundry, and reconciliation artifacts now agree on 40/40 complete,
  ready_nodes=0, checkpoint_count=40, final_response_allowed=true, no authority
  promotion, and RSI remains denied.
- Foundry terminal examples: `completed`, `promoted`, `denied`, and `blocked`
  are explicit in the recommendation readback. `promoted` normalizes to
  `completed` only when Promoter and Command agree and RSI remains denied.
- Foundry denied examples: `missing_node_evidence`,
  `missing_stop_gate_evidence`, and `forbidden_surface_or_rsi_claim` require
  exact missing-evidence readbacks, keep RSI denied, and do not claim authority
  advancement.
- Blocked-node continuation: generated Atlas prompts now require the exact
  blocked node id, missing evidence or stop gate, safe repair or repack action,
  and checkpoint resume after repair.

## Root Artifacts

- `recommendation-wave.json`: 40-node Atlas-owned recommendation wave.
- `recommendation-workgraph.json`: initial generated workgraph.
- `lease-start.json`: durable lease start marker.
- `recommendation-readback.json`: authoritative resumed readback.
- `execution-readback.json`: execution ledger matched to the readback.
- `command-readback.json`: compact Command timeline with structured binding for
  elapsed lease status, return gate, first executable node, and next action.
- `promoter-readback.json`: no-promotion readback; RSI remains denied.
- `foundry-rollup.json`: Foundry node/lease summary.
- `reconciliation-packet.json`: cross-artifact return-gate agreement packet.
- `final-synthesis.json`: cross-artifact synthesis and next action.

## Fixtures

- `fixtures/all-complete-under-120-readback.json`: all nodes complete at 22
  minutes; final response remains denied.
- `fixtures/all-complete-lease-met-readback.json`: all nodes complete at 120
  minutes; final response is allowed.
- `fixtures/blocked-before-lease-readback.json`: blocked node before lease
  completion; final response remains denied with exact repair action.

## Safety

This evidence root records readback and fixture-only handoff evidence. It does
not call providers, inspect credentials, mutate `main` directly, release,
deploy, publish, upload, tag, update dependencies, widen auth/policy/config, or
claim RSI. RSI remains denied.
