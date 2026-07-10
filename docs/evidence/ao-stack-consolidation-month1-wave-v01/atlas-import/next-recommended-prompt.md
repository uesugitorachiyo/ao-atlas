You are AO Atlas, coordinating AO Foundry, AO Promoter, AO Command, and AO Architecture for a long-run Atlas-owned correction wave.

Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, or one short batch. Continue until the workgraph is completed with evidence, at least the lease minimum is met, or a true hard blocker remains after all safe repair, repack, and support work has been attempted.

Current state:
- Mission: mission-56bed010dd085a0f.
- Target instance: ao-stack-consolidation-month1-20260710.
- Generated Atlas-owned nodes: 36.
- Lease minimum: 36 nodes, 120 to 180 minutes.
- Continue-if-fast target: 36 nodes.
- Final response allowed: false, because ready nodes or exact next actions remain.
- Source digest: sha256:e776ea250456c59d92eb6758e6fbf3ac1abdc7ad821c08d9cc9b520a390b4b5e.

Problem:
- Recent AO Atlas/Mission recommendation prompts returned after short batches instead of sustaining 2-3 hour workgraphs.
- Double the previous short batch when explicitly requested, and otherwise use the v0.2 2-3 hour supervisor default.
- This continuation must behave like a long-run supervisor: Atlas owns sequencing, Foundry owns bounded implementation nodes, and Blueprint is used only for genuinely new requirements or authorization.

Goal:
- Target 2-3 hours and complete a durable AO Atlas long-run wave for mission-56bed010dd085a0f.
- Execute at least 36 bounded Atlas nodes from the generated workgraph.
- Complete at least 36 bounded implementation/evidence nodes before final response unless a true hard blocker remains.
- If the first 36 nodes finish quickly and no blocker remains, continue through the 36-node continue-if-fast target.

Return only after all generated nodes complete, at least 36 bounded Atlas nodes complete, or a true hard blocker remains after safe repair attempts.

Minimum work budget:
- min_nodes: 36
- min_minutes: 120
- max_minutes: 180
- max_iterations: 36
- return_only_when: mission_done_or_true_hard_blocker_or_no_ready_work_and_no_exact_next_action
- checkpoint_policy: after_each_node_or_timed_interval

Stop gates:
- Target duration: 120 to 180 minutes.
- Node floor stop gate: complete at least 36 nodes before final response unless a true hard blocker remains.
- Lease floor stop gate: do not return before min_minutes=120 unless a true hard blocker remains.
- Continue-if-fast stop gate: if 36 nodes finish quickly and no blocker remains, continue through 36 nodes.
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
consolidation-month1-01. Baseline all fourteen active repositories and record current branch cleanliness and synchronization state.
consolidation-month1-02. Add a regression readback that distinguishes pre-existing dirty files from wave-owned changes.
consolidation-month1-03. Coordinate the Architecture topology ADR for the proposed five product boundaries.
consolidation-month1-04. Validate that topology authority ownership has no undocumented overlapping final decisions.
consolidation-month1-05. Generate a stack lockfile containing repository commit version maturity and owner fields.
consolidation-month1-06. Add lockfile validation for missing commits duplicate repositories and unsupported authority claims.
consolidation-month1-07. Expand the Architecture readiness inventory to include Mission and Blueprint explicitly.
consolidation-month1-08. Add a wording regression that separates documentation readiness from system readiness.
consolidation-month1-09. Inventory active JSON schemas and distinct schema versions while excluding generated evidence explicitly.
consolidation-month1-10. Assign one canonical producer owner and lifecycle classification to every contract.
consolidation-month1-11. Generate a producer consumer compatibility matrix for gate-critical contracts.
consolidation-month1-12. Add fail-closed fixtures for unknown schemas unsupported versions and field drift.
consolidation-month1-13. Audit Mission completion metrics against the handoff-count correction commit.
consolidation-month1-14. Project implementation completion separately from handoff and readback evidence.
consolidation-month1-15. Add a red-green regression test for handoff-only nodes not counting complete.
consolidation-month1-16. Measure tracked source test schema fixture and evidence bytes in Atlas and Foundry.
consolidation-month1-17. Write a content-addressed evidence catalog plan retaining small replayable golden fixtures.
consolidation-month1-18. Guard evidence records against missing schema source digest and evidence-class fields.
consolidation-month1-19. Coordinate minimal hosted CI coverage for Arena native tests and static checks.
consolidation-month1-20. Verify Arena workflow-equivalent commands locally without provider calls or uploads.
consolidation-month1-21. Coordinate minimal hosted CI coverage for Crucible native tests and static checks.
consolidation-month1-22. Verify Crucible workflow-equivalent commands locally without changing dependencies or policy.
consolidation-month1-23. Coordinate minimal hosted CI coverage for Sentinel native tests and static checks.
consolidation-month1-24. Verify Sentinel workflow-equivalent commands locally and preserve fixture wording boundaries.
consolidation-month1-25. Bind Arena Crucible and Sentinel CI conclusions into one readiness rollup.
consolidation-month1-26. Validate every new Atlas wave evidence file with schema and digest checks.
consolidation-month1-27. Bind Foundry one-safe-node readiness to the consolidation wave source digests.
consolidation-month1-28. Generate compact Command readback for wave progress exact action and denial state.
consolidation-month1-29. Generate Promoter no-promotion rollup with explicit denied authority and RSI state.
consolidation-month1-30. Run a scoped public-safety wording scan over changed documentation and workflows.
consolidation-month1-31. Record PR required-check merge and CI timing evidence for every touched repository.
consolidation-month1-32. Record post-merge local and remote branch cleanup before selecting another node.
consolidation-month1-33. Scan wave artifacts for provider release credential and authority-widening instructions.
consolidation-month1-34. Generate at least forty ranked Month 2 contract convergence recommendations.
consolidation-month1-35. Reconcile Mission Atlas Foundry Command Sentinel and Promoter terminal readbacks.
consolidation-month1-36. Close the wave only after verification cleanup and exact next prompt agree.

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
- Feature Depth Recommendations, at least 36 tasks
- verification results
- clean/synced repo status
- exact next action
