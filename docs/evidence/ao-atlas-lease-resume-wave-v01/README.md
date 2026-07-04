# AO Atlas Lease Resume Wave v01

Status: in progress.

This evidence root records the first post-repair recommendation wave using the
persisted lease-start and resume machinery. The wave starts at
`2026-07-04T08:34:04-07:00`, has a 120-minute minimum, and remains final-denied
because ready nodes remain and the lease floor is not met.

## Current Readback

- Total nodes: 40
- Completed nodes: 2
- Ready nodes: 38
- Blocked nodes: 0
- Failed nodes: 0
- Elapsed minutes: 19
- Minimum minutes met: false
- Return gate status: `blocked_ready_nodes_remain`
- Checkpoint count: 2
- Final response allowed: false
- Exact next action: emit Foundry import for `mission-recommendation-next-03`
  and execute exactly one active node.

## Root Artifacts

- `recommendation-wave.json`: 40-node Atlas-owned recommendation wave.
- `recommendation-workgraph.json`: initial generated workgraph.
- `lease-start.json`: durable lease start marker.
- `recommendation-readback.json`: authoritative resumed readback.
- `execution-readback.json`: execution ledger matched to the readback.
- `command-readback.json`: compact Command timeline with elapsed lease status.
- `promoter-readback.json`: no-promotion readback; RSI remains denied.
- `foundry-rollup.json`: Foundry node/lease summary.
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
