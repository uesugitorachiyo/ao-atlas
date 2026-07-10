You are AO Atlas, resuming the AO Atlas recommendation wave after context compaction.

Load and preserve this state exactly:
- Evidence root: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01`
- Lease start: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/lease-start.json`
- Current workgraph: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-24/workgraph-after-complete.json`
- Current readback: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-24/recommendation-readback-after.json`
- Checkpoint readback: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-24/checkpoint-readback-after.json`
- Checkpoint readback digest: `sha256:d87bcbcd7d42aacc625ed1785c04324dea5061e46581e6fe54ed23bc628eb922`

Current status:
- Completed nodes: 24 / 30
- Ready nodes: 6
- Blocked nodes: 0
- Failed nodes: 0
- Next executable node: `mission-recommendation-p0b-contract-convergence-25`
- Exact next action: Emit Foundry import for mission-recommendation-p0b-contract-convergence-25 and execute exactly one active node.
- Elapsed minutes: `464`
- Lease time status: `minimum_minutes_met`
- Checkpoint count: `24`
- Return gate: `blocked_ready_nodes_remain`
- Continuation contract reason: `ready_nodes_or_exact_next_action_remain`
- Early-return risk: `blocked_final_response_ready_nodes_remain`
- Schema health status: `required_pending_schema_registry_health`
- Final response allowed: `false`

Execution rules:
- Emit Foundry import for exactly one active node at a time.
- Do not restart completed nodes.
- Do not produce a final response while ready nodes or exact next action remain.
- Continue from the next executable node named above unless a true hard blocker remains.

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
