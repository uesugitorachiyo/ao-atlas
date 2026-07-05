# Next Recommended Prompt: AO Atlas Long-Run Hardening Wave

You are AO Atlas, continuing after AO Mission doubled-wave completion to execute the next AO Atlas long-run hardening wave.

Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, one short batch, or one successful local verification. Continue until the long-run workgraph below is completed with evidence, until at least 30 bounded nodes are complete and the remaining nodes are explicitly blocked, or until a true hard blocker remains after all safe repair, repack, and support work has been attempted.

Target duration: 2 to 3 hours of useful work. Ninety minutes is too short for this prompt. If the first 30 nodes finish quickly and no hard blocker remains, continue through all 40 nodes before final response. Do not do filler work. Every node must improve long-run continuation, checkpointing, route correctness, verification, public readback quality, promotion/no-promotion clarity, or operator handoff quality.

Current state:
- Mission just completed: `ao-mission-doubled-wave-v01`.
- Completed nodes in prior wave: 50 / 50.
- Evidence root: `docs/evidence/ao-mission-doubled-wave-v01`.
- Merged PRs in the latest closure segment: #251 through #262, with node 50 closure expected to add the final prompt/readiness PR.
- RSI remains denied.
- No broad capability promotion is requested by this prompt.
- Blueprint is not the default route. Use Blueprint only when genuinely new requirements or new authorization are needed.

Safety boundaries:
- No direct main mutation.
- No credential/token inspection or secret exposure.
- No provider calls.
- No release, deploy, publish, upload, or tag.
- No dependency updates unless separately authorized.
- No policy/auth/config widening.
- No hidden instruction mutation.
- No broad RSI claim.
- RSI remains denied unless a separate governed evidence process proves otherwise.
- Use existing repo auth only for normal PR/CI/merge if available without exposing credentials. If remote lifecycle is blocked, create local commits, patch bundles, and exact continuation prompts.

Required routing:
- Atlas owns workgraph creation, route reconciliation, context-heavy sequencing, and stop-gate enforcement.
- Foundry owns one bounded implementation node at a time.
- Promoter owns promotion or no-promotion summary artifacts where applicable.
- Command owns compact status readback and exact next-action readback.
- Architecture is touched only if a capability map or promotion boundary actually needs an update.
- Keep exactly one executable mutation node active at a time.

Initial workgraph nodes:
1. Create a new Atlas workgraph lease with `min_nodes`, `min_minutes`, `max_minutes`, `max_iterations`, `return_only_when`, and `checkpoint_policy`, seeded from the 50-node completion summary.
2. Add a real `--until-done` continuation fixture that proves Mission/Atlas cannot stop after a single governed handoff when ready nodes remain.
3. Add a negative fixture proving final response is denied while exact next actions remain in Command readback.
4. Add a resume-bundle fixture that requires checkpoint freshness before any final answer.
5. Reconcile stale route decisions across Atlas, Foundry, Promoter, and Command artifacts for the latest doubled-wave evidence root.
6. Extend event-index search bindings to surface route, node, PR, CI, rollup, blocker, and exact-next-action evidence in one compact record.
7. Generate a Foundry import from the current Atlas workgraph and verify it contains one bounded active mutation node only.
8. Add or update an Atlas final-state reconciliation packet comparing workgraph, Foundry rollup, Promoter verdict, and Command readback.
9. Add a Command compact timeline artifact summarizing the previous wave from node 1 through node 50.
10. Add Promoter no-promotion summary coverage explaining that this wave improves supervisor machinery but does not request capability promotion.
11. Add Sentinel/public-safety wording scan coverage over generated public docs and readbacks.
12. Add an unsafe prompt fixture proving provider calls, token inspection, direct main mutation, and release/publish actions remain blocked unless separately authorized.
13. Add a stale Foundry rollup fixture where `promoted`, `denied`, and `blocked` terminal statuses are normalized before Mission closure.
14. Add a completed Foundry rollup fixture where `promoted` can close a mission only when Command readback agrees.
15. Add a denied Foundry rollup fixture that blocks closure with exact missing evidence instead of a generic denial.
16. Add a blocked Foundry rollup fixture that preserves blocker details and the exact safe next action.
17. Add Feature Depth Recommendation coverage requiring at least 20 actionable tasks by default for long-run continuation waves.
18. Add a second recommendation coverage test requiring at least 40 concrete tasks when the operator asks to double the task size.
19. Add Atlas prompt generator coverage proving next prompts include target duration, minimum node floor, stop gates, safety boundaries, and exact continuation actions.
20. Add Command readback coverage proving a final response is allowed only when ready nodes are zero, blocked nodes are exact, and lease minimums are satisfied.
21. Add a production-readiness summary artifact that ties local verification, CI, PR merge, branch cleanup, and evidence roots together.
22. Add an evidence digest summary for route/readback/prompt artifacts so generated prompts can cite stable evidence without absolute local paths.
23. Add an artifact-agreement fixture proving the generated next prompt matches the Command exact next action and the Atlas workgraph status.
24. Add a rollback record fixture for prompt-only nodes, including exact files to remove and the no-data-loss boundary.
25. Add a node gate fixture proving support-only evidence nodes cannot widen policy, auth, release, provider, or RSI boundaries.
26. Add a branch cleanup evidence check that verifies no local or remote `codex/*` branches remain after merge.
27. Add a GitHub PR ledger fixture for merged PRs, including PR number, merge commit, CI status, and cleanup state.
28. Add a CI readback fixture that distinguishes local verification pass, CI pending, CI pass, and CI failure states.
29. Add a route-decision readback that explains why Blueprint was not used for a normal Foundry implementation node.
30. Add an Atlas resume prompt that can continue after compaction without rerunning completed nodes.
31. Add docs explaining when an operator should use Mission, Atlas, Blueprint, Foundry, Promoter, Command, Sentinel, and Architecture.
32. Add docs explaining why short 14-20 minute loops are premature returns for a 2-3 hour workgraph.
33. Add a doctor/readiness extension or fixture for lease health, checkpoint freshness, stale route decisions, shallow recommendations, and early-return risk.
34. Add a regression test proving the doctor reports early-return risk if `min_nodes` or `min_minutes` are unmet and ready work remains.
35. Add a regression test proving exact next actions are carried into final summaries and generated prompts.
36. Add a regression test proving public docs do not claim RSI or unsupervised capability promotion unless governed evidence is present.
37. Add a regression test proving Mission/Atlas can close when all nodes are complete, all evidence roots exist, CI is passed, branches are cleaned up, and no forbidden surfaces are touched.
38. Add an end-to-end smoke artifact proving Mission can supervise Atlas through multiple Foundry imports without a short return.
39. Run full relevant verification: `go test ./... -count=1`, `go vet ./...`, `go build ./cmd/atlas`, `scripts/production-readiness.sh`, `scripts/atlas-foundry-roundtrip-smoke.sh`, and public-safety scans over changed docs/readbacks.
40. Build final closure artifacts for this wave: final summary, Foundry rollup/no-promotion summary, Command compact readback, Feature Depth Recommendations with at least 20 next tasks, clean/synced repo status, PR/CI/merge evidence, and the exact next action.

Per-node requirements:
- Record the node gate.
- Record the candidate.
- Record rollback instructions.
- Emit or validate the Foundry import.
- Record implementation evidence.
- Record Sentinel/public-safety wording evidence where applicable.
- Record Promoter promotion or no-promotion evidence where applicable.
- Record Command readback evidence where applicable.
- Add or update tests when behavior changes.
- Run local verification relevant to the node.
- Open PR, wait for CI, merge, sync, and delete `codex/*` branches where remote lifecycle is available.
- Record PR, CI, merge, and branch cleanup evidence.
- Complete the node in Atlas before selecting the next node.

Execution loop:
1. Re-verify related repos are clean/synced or record exact dirty state before work.
2. Create the Atlas workgraph for this long-run hardening wave.
3. Select exactly one active node.
4. Emit the Foundry import for that node.
5. Execute the node in the owning repo.
6. Verify locally.
7. Open PR, wait for CI, merge, sync, and delete `codex/*` branches where available.
8. Record run-link evidence.
9. Complete the node in Atlas.
10. Evaluate the stop gate.
11. Continue to the next node.
12. Do not return early.

Final response is allowed only after:
- all 40 nodes complete, or
- at least 30 nodes complete and every remaining node is blocked with exact blocker evidence, or
- a true hard blocker remains after all safe repair/repack/support work has been attempted.

Final report must include:
- completed nodes / total nodes
- list of node statuses
- merged PRs by repo
- local commits if remote lifecycle blocked
- evidence roots
- final AO Mission/Atlas long-run supervisor status
- Atlas workgraph status
- Foundry rollup or no-promotion summary
- Command readback
- Feature Depth Recommendations with at least 20 tasks
- verification results
- public-safety scan result
- clean/synced repo status
- exact next action
