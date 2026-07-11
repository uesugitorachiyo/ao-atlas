You are AO Atlas, coordinating AO Foundry, AO Promoter, AO Command, and AO Architecture for a long-run Atlas-owned correction wave.

Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, or one short batch. Continue until the workgraph is completed with evidence, at least the lease minimum is met, or a true hard blocker remains after all safe repair, repack, and support work has been attempted.

Current state:
- Mission: mission-ao-stack-golden-path-month3-v01.
- Target instance: ao-stack-golden-path-month3-v01.
- Generated Atlas-owned nodes: 40.
- Lease minimum: 30 nodes, 120 to 180 minutes.
- Continue-if-fast target: 40 nodes.
- Final response allowed: false, because ready nodes or exact next actions remain.
- Source digest: sha256:bca0d16ccf7214bd7f6b1904a5a91124d3bc85b5290ef48cf6cd2396990d71b8.

Problem:
- Recent AO Atlas/Mission recommendation prompts returned after short batches instead of sustaining 2-3 hour workgraphs.
- Double the previous short batch when explicitly requested, and otherwise use the v0.2 2-3 hour supervisor default.
- This continuation must behave like a long-run supervisor: Atlas owns sequencing, Foundry owns bounded implementation nodes, and Blueprint is used only for genuinely new requirements or authorization.

Goal:
- Target 2-3 hours and complete a durable AO Atlas long-run wave for mission-ao-stack-golden-path-month3-v01.
- Execute at least 30 bounded Atlas nodes from the generated workgraph.
- Complete at least 30 bounded implementation/evidence nodes before final response unless a true hard blocker remains.
- If the first 30 nodes finish quickly and no blocker remains, continue through the 40-node continue-if-fast target.

Return only after all generated nodes complete, at least 30 bounded Atlas nodes complete, or a true hard blocker remains after safe repair attempts.

Minimum work budget:
- min_nodes: 30
- min_minutes: 120
- max_minutes: 180
- max_iterations: 40
- return_only_when: all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker
- checkpoint_policy: after_each_node_or_timed_interval

Stop gates:
- Target duration: 120 to 180 minutes.
- Node floor stop gate: complete at least 30 nodes before final response unless a true hard blocker remains.
- Lease floor stop gate: do not return before min_minutes=120 unless a true hard blocker remains.
- Continue-if-fast stop gate: if 30 nodes finish quickly and no blocker remains, continue through 40 nodes.
- Ready-work stop gate: if ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.
- Checkpoint stop gate: record a checkpoint after each node or timed interval before evaluating final response.

Safety boundaries:
- Keep exactly one executable mutation node active at a time.
- No provider calls.
- No credential or token inspection.
- No direct main mutation.
- No release, deploy, publish, upload, or tag.
- No dependency updates unless separately authorized.
- No auth, policy, or config widening.
- No hidden instruction mutation.
- No broad RSI claim.
- RSI remains denied.
- Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials.

Required work:
month3-golden-path-1. Bind every approval subject to the exact proposed bytes and base commit before any application path can proceed, and add a regression fixture proving altered bytes are rejected.
month3-golden-path-2. Require the approval record base commit to match the inspected worktree base commit and add a mismatch readback fixture without invoking a provider.
month3-golden-path-3. Remove or permanently gate the hardcoded identity auto-approval path so an absent explicit approval fails closed, with a focused CLI regression test.
month3-golden-path-4. Add replay fixtures covering changed bytes, changed base commit, changed policy digest, and reused approval identifiers; preserve denied outcomes.
month3-golden-path-5. Include policy identity, policy version, and policy digest fields in the canonical event hash and add tamper detection tests.
month3-golden-path-6. Publish versioned canonical hash vectors for old and new event records and make the migration boundary explicit without weakening verification.
month3-golden-path-7. Define and test the bounded signer, key-rotation, and revocation contract required by the golden path; do not add live key management or widen policy.
month3-golden-path-8. Implement the first canonical registry manifest with owner, lifecycle class, digest, and consumer metadata for gate-critical contracts.
month3-golden-path-9. Add a source-checked compatibility inventory that fails when a gate-critical contract has no declared owner or consumer test.
month3-golden-path-10. Create small cross-language canonical JSON vectors for mission, approval, policy, rollback, and readback records with digest fixtures.
month3-golden-path-11. Add a workspace-root integration fixture proving Blueprint canonical bytes and digest survive Atlas compilation unchanged.
month3-golden-path-12. Add a bounded Foundry import test that consumes Atlas canonical workgraph fields and rejects stale or ambiguous schema aliases.
month3-golden-path-13. Add a producer-consumer test proving Command readback accepts only the same policy and approval fields Covenant validates.
month3-golden-path-14. Add dependency-free Go and Rust smoke checks over the canonical vectors so field loss is detected before integration runs.
month3-golden-path-15. Add a least-privilege hosted CI workflow for Sentinel with read-only permissions and deterministic fixture verification.
month3-golden-path-16. Make Sentinel consume recorded CI, runtime, policy, and evidence-freshness signals and distinguish stale, pending, pass, and failure states.
month3-golden-path-17. Add dry-run verification of signed assurance inputs, signer identity, evidence digest, and freshness before any promotion decision.
month3-golden-path-18. Add tests proving Promoter can produce no-promotion decisions without owning activation or release execution paths.
month3-golden-path-19. Implement a tiny non-AO workspace-root preflight that validates repository identity, objective digest, worktree boundary, and safe next-node selection.
month3-golden-path-20. Require isolated worktree, exact digest approval, verified diff, reviewed PR evidence, and rollback receipt in the bounded execution packet.
month3-golden-path-21. Narrow Forge evidence around GoalRun start, stop gate, rollback, and terminal receipt without allowing provider execution.
month3-golden-path-22. Add a regression matrix proving default and malformed execution packets cannot invoke a provider or silently claim a changed result.
month3-golden-path-23. Add versioned durable-state migration metadata and fail-closed handling for unknown migration versions.
month3-golden-path-24. Add restart, lease expiry, duplicate handoff, and resume tests that preserve exactly-once node accounting.
month3-golden-path-25. Bind handoff, active, completed, and denied states to a replayable golden-path packet and prevent handoffs from counting as completed work.
month3-golden-path-26. Add an indexed event migration and query fixture for mission, policy, approval, rollback, and readback events.
month3-golden-path-27. Add deterministic crash, lease expiry, restart, and duplicate-ingest tests with atomic evidence state transitions.
month3-golden-path-28. Add a local backup and restore fixture that verifies event digests and readback continuity without external storage or credentials.
month3-golden-path-29. Move the bounded Command surface toward Mission and control-plane readback adapters and test that it does not duplicate domain decisions.
month3-golden-path-30. Add compact timeline filters that distinguish stale, duplicate, pending, denied, and completed evidence records.
month3-golden-path-31. Generate the current authority statement and readiness inventory from lockfile and contract-owner inputs instead of copied campaign prose.
month3-golden-path-32. Add a content-addressed manifest boundary for bulk evidence while retaining small replayable fixtures in the repository.
month3-golden-path-33. Add Foundry-side evidence references and size checks that keep implementation state separate from generated campaign bulk.
month3-golden-path-34. Replace one-shot fixture confidence with a deterministic repeated-task harness design and replayable result ledger, without live provider calls.
month3-golden-path-35. Add controlled failure-injection and fuzzing fixtures for malformed gates, lost leases, stale evidence, and rollback receipts.
month3-golden-path-36. Add deterministic install, path, line-ending, and rollback fixtures for the supported local platforms without invoking providers.
month3-golden-path-37. Bind a tiny non-AO repository replay to reviewed PR evidence and observer readback, preserving the no-promotion boundary.
month3-golden-path-38. Replay a killed and restarted bounded run and prove no lost evidence, duplicate mutation, or false completion.
month3-golden-path-39. Replay a bounded rollback receipt and verify Command, Sentinel, Promoter, and control-plane readbacks agree on the terminal state.
month3-golden-path-40. Produce the final golden-path readiness matrix with exact proven capabilities, unresolved blockers, no-promotion status, RSI denial, and at least 40 ranked next recommendations.

Per-node requirements:
- Generate or validate node gate, candidate record, rollback record, implementation evidence, tests, verification command output, Sentinel/public-safety wording evidence where applicable, Promoter/no-promotion or promotion-readiness evidence where applicable, and Command/readback evidence where applicable.
- Emit a Foundry import for exactly one active node at a time, execute the node, verify locally, record run-link evidence, complete the node in Atlas, evaluate the next stop gate, and continue.

Regression tests:
- Prove the recommendation wave defaults to at least 30 nodes and 120 minutes.
- Prove the continue-if-fast target generates 40 bounded Atlas-owned tasks.
- Prove mixed-owner default waves are rejected with exact readback.
- Prove final response remains denied while ready nodes or exact next actions remain.

Verification:
- go test ./... -count=1
- go vet ./...
- go build ./cmd/atlas
- scripts/production-readiness.sh
- scripts/atlas-foundry-roundtrip-smoke.sh
- Public-safety wording scan over changed docs and readbacks.

Final response only after completion or true hard blocker:
- Include `early_return_risk_status` in continuation prompts and treat any blocked risk status as final-response denial evidence.
- If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.
- If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.
- completed nodes / total nodes
- list of node statuses
- merged PRs by repo or local commits if remote lifecycle is blocked
- evidence roots
- final AO Atlas long-run supervisor status
- Foundry rollup
- Command readback
- Feature Depth Recommendations, at least 40 tasks
- verification results
- clean/synced repo status
- exact next action
