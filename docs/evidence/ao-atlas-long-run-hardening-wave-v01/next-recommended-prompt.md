You are AO Atlas, coordinating AO Foundry, AO Promoter, AO Command, and AO Architecture for a long-run Atlas-owned correction wave.

Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, or one short batch. Continue until the workgraph is completed with evidence, at least the lease minimum is met, or a true hard blocker remains after all safe repair, repack, and support work has been attempted.

Current state:
- Mission: ao-atlas-long-run-hardening-wave-v01.
- Target instance: ao-atlas-long-run-hardening-wave-v01.
- Generated Atlas-owned nodes: 40.
- Lease minimum: 30 nodes, 120 to 180 minutes.
- Continue-if-fast target: 40 nodes.
- Final response allowed: false, because ready nodes or exact next actions remain.
- Source digest: sha256:c30756397352c16576216236568a34337900d81bc83b6600f43ffa710ef6cf82.

Problem:
- Recent AO Atlas/Mission recommendation prompts returned after short batches instead of sustaining 2-3 hour workgraphs.
- Double the previous short batch when explicitly requested, and otherwise use the v0.2 2-3 hour supervisor default.
- This continuation must behave like a long-run supervisor: Atlas owns sequencing, Foundry owns bounded implementation nodes, and Blueprint is used only for genuinely new requirements or authorization.

Goal:
- Target 2-3 hours and complete a durable AO Atlas long-run wave for ao-atlas-long-run-hardening-wave-v01.
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

Safety boundaries:
- Keep exactly one executable mutation node active at a time.
- Preserve no provider calls, no credential inspection, no direct main mutation, no release/deploy/publish/upload/tag, no dependency updates, no auth/policy/config widening, and no broad RSI claim.
- RSI remains denied unless separate governed evidence proves otherwise.
- Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials.

Required work:
hardening-01. Create a new Atlas workgraph lease seeded from the completed 50-node doubled wave summary.
hardening-02. Add an until-done continuation fixture proving one governed handoff cannot end a mission.
hardening-03. Add a Command readback fixture denying final response while exact next actions remain.
hardening-04. Add a resume bundle fixture requiring fresh checkpoints before any final answer.
hardening-05. Reconcile stale route decisions across Atlas Foundry Promoter and Command artifacts.
hardening-06. Extend event index bindings for route node pull request CI rollup blocker and next action evidence.
hardening-07. Generate a Foundry import that proves exactly one bounded mutation node is active.
hardening-08. Add an Atlas final-state reconciliation packet comparing workgraph Foundry Promoter and Command readbacks.
hardening-09. Add a Command compact timeline artifact summarizing the previous 50-node doubled wave.
hardening-10. Add Promoter no-promotion summary coverage for supervisor hardening without capability promotion.
hardening-11. Add Sentinel public-safety wording scan coverage over generated docs and readbacks.
hardening-12. Add an unsafe prompt fixture keeping provider token main mutation and release actions blocked.
hardening-13. Add stale Foundry rollup fixture coverage for promoted denied and blocked terminal normalization.
hardening-14. Add completed Foundry rollup fixture coverage where promoted closes only with Command agreement.
hardening-15. Add denied Foundry rollup fixture coverage that reports exact missing evidence.
hardening-16. Add blocked Foundry rollup fixture coverage preserving blocker details and safe next action.
hardening-17. Require Feature Depth Recommendations to return at least 20 actionable tasks by default.
hardening-18. Require doubled Feature Depth Recommendations to return at least 40 concrete tasks.
hardening-19. Add Atlas prompt generator coverage for target duration node floor stop gates and safety boundaries.
hardening-20. Add Command readback coverage allowing final response only after zero ready nodes and met lease minimums.
hardening-21. Add production readiness summary tying verification CI merge cleanup and evidence roots together.
hardening-22. Add evidence digest summary for route readback and prompt artifacts without absolute paths.
hardening-23. Add artifact agreement fixture tying generated prompt Command next action and workgraph status.
hardening-24. Add rollback record fixture for prompt-only nodes with no data loss boundary.
hardening-25. Add node gate fixture proving support evidence nodes cannot widen authority boundaries.
hardening-26. Add branch cleanup evidence check for local and remote codex branch removal.
hardening-27. Add GitHub pull request ledger fixture with number merge commit CI status and cleanup state.
hardening-28. Add CI readback fixture distinguishing local pass CI pending CI pass and CI failure states.
hardening-29. Add route decision readback explaining why Blueprint was not used for normal Foundry implementation.
hardening-30. Add Atlas resume prompt that continues after compaction without rerunning completed nodes.
hardening-31. Document when operators should use Mission Atlas Blueprint Foundry Promoter Command Sentinel and Architecture.
hardening-32. Document why 14 to 20 minute loops are premature returns for a 2 to 3 hour workgraph.
hardening-33. Add doctor readiness fixture for lease health checkpoint freshness stale routes shallow recommendations and early return risk.
hardening-34. Add regression coverage for early-return risk when min nodes or min minutes are unmet.
hardening-35. Add regression coverage that exact next actions carry into summaries and generated prompts.
hardening-36. Add regression coverage that public docs cannot claim RSI or unsupervised promotion without evidence.
hardening-37. Add closure regression proving Mission Atlas closes only with complete evidence CI cleanup and no forbidden surfaces.
hardening-38. Add end-to-end smoke artifact for Mission supervising Atlas through multiple Foundry imports.
hardening-39. Run full relevant verification and public-safety scans over changed docs and readbacks.
hardening-40. Build final closure artifacts with recommendations verification clean repo status and exact next action.

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
- Feature Depth Recommendations, at least 10 tasks
- verification results
- clean/synced repo status
- exact next action
