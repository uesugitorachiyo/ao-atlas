# AO Atlas Lease Resume Wave v01

Status: in progress.

This evidence root records the first post-repair recommendation wave using the
persisted lease-start and resume machinery. The wave starts at
`2026-07-04T08:34:04-07:00`, has a 120-minute minimum, and remains final-denied
because ready nodes remain.

## Current Readback

- Total nodes: 40
- Completed nodes: 33
- Ready nodes: 7
- Blocked nodes: 0
- Failed nodes: 0
- Elapsed minutes: 400
- Minimum minutes met: true
- Lease health status: `minimum_met_continue_if_fast`
- Checkpoint freshness status: `fresh_checkpoint_required_after_each_node_or_timed_interval`
- Stale route decision status: `fresh_atlas_supervises_foundry_owns_one_active_node`
- Early-return risk status: `blocked_final_response_ready_nodes_remain`
- Return gate status: `blocked_ready_nodes_remain`
- Checkpoint count: 33
- Final response allowed: false
- Exact next action: emit Foundry import for `mission-recommendation-next-34`
  and execute exactly one active node.
- Exact next action readback: `continuation_required`, bound to
  `mission-recommendation-next-34`, `blocked_ready_nodes_remain`, and
  `final_response_allowed=false`.
- Command timeline placeholders: `checkpoint`, `exact_next_action`, and
  `return_gate` are pending Command timeline slots required before final
  response.
- Promoter no-promotion placeholders: `promotion_claim`, `rsi_boundary`, and
  `authority_advance` are pending Promoter slots required before final
  response.
- Final-response denial tests: ready recommendation workgraphs now reject stale
  `final_response_allowed` return gates, stale final reasons, and exact-next
  actions that do not name the first executable node.
- Final-response allowance tests: completed recommendation workgraphs now
  reject stale allowed-state status, return gate, final reason, and final exact
  next action drift.
- Long-run lease examples: `examples/valid/recommendation-wave-long-run-supervisor.json`
  and `examples/valid/recommendation-lease-start-long-run.json` cover the
  30-node, 120-180 minute, continue-if-fast lease fields.
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
