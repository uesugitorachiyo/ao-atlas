Start a Codex goal and execute AO Mission supervised P0-C Mission-to-Foundry real complete-path readiness wave.

Mission: mission-710327df54728420
Source wave: docs/evidence/ao-stack-p0b-contract-convergence-wave-v01
Source P0-C readiness criteria: docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-29/p0c-readiness-criteria.json

Goal:
Build the next bounded AO stack wave that turns the P0-B contract convergence evidence into a real Mission-to-Foundry complete-path readiness track. Use AO Mission as supervision owner and AO Atlas as workgraph owner.

Minimum work budget: generate at least 30 bounded nodes and complete at least 20 before final response unless a true hard blocker remains.

Required behavior:
1. Re-verify ao-mission, ao-atlas, ao-foundry, ao-command, ao-sentinel, and ao-promoter are clean, synced, and have no local or remote codex branches.
2. Load the P0-B terminal readback and P0-C readiness criteria.
3. Generate a P0-C Mission-to-Foundry workgraph with at least 30 bounded nodes.
4. Keep exactly one executable mutation node active at a time.
5. For each node, emit node gate, candidate record, rollback record, tests, verification, Sentinel public-safety evidence, Promoter no-promotion evidence, Command readback, run-link, checkpoint, and readback.
6. Route Atlas-owned work to Atlas and Foundry-owned compatibility work to Foundry.
7. Use PR, CI, merge, main sync, and branch cleanup evidence for every mutation branch.
8. Continue automatically while ready_nodes is greater than zero or exact_next_action is present.

P0-C required themes:
1. Mission start handoff preservation into Atlas compile input.
2. Atlas workgraph import to Foundry without rewriting governed Mission bytes.
3. Foundry import compatibility with Mission-originated workgraph context.
4. Foundry readiness packet generation for exactly one active node.
5. Mission readback after Foundry import and validation.
6. Command compact status for Mission-to-Foundry path state.
7. Sentinel public-safety scan over generated prompts, readbacks, and docs.
8. Promoter no-promotion rollup for the P0-C readiness wave.
9. Restart and resume checkpoint from the latest readback.
10. PR, CI, merge, sync, and branch cleanup ledger for the whole wave.
11. Failure replay fixtures for blocked Foundry import and stale readback.
12. Terminal P0-D recommendation exporter after P0-C closure.

Safety boundaries:
- no direct main mutation
- no credential/token inspection or secret exposure
- no provider calls
- no release, deploy, publish, upload, or tag
- no dependency updates
- no auth, policy, or config widening
- no hidden instruction mutation
- no broad RSI claim
- RSI remains denied.

Do not use provider calls or credentials.
Do not execute AO2 live mutation.
Do not claim fully_unsupervised_complex_mutation or RSI proof.

Final response is allowed only when:
- the generated P0-C node budget is satisfied
- ready_nodes=0
- blocked_nodes=0
- failed_nodes=0
- final_response_allowed=true
- local verification passes in every touched repo
- GitHub CI passes for every merged PR
- public-safety scan passes
- Promoter says no_promotion_requested or equivalent
- Command readback agrees
- RSI remains denied
- local and remote codex branches are deleted

If final response is still blocked by ready work or exact_next_action, continue automatically.
