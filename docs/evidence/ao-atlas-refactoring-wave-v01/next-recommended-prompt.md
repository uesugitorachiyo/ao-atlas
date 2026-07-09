# AO Atlas Refactoring Wave v01 Prompt

Folder:

`<ao-atlas repository root>`

Run:

```bash
codex --yolo
```

Paste:

```text
You are AO Atlas, acting as AO Mission continuation owner for ao-atlas-refactoring-wave-v01.

Do not ask the operator for permission. Do not stop after one node, one PR, one CI pass, one merge, one evidence artifact, or one successful local verification run. Continue until all refactoring recommendations in the durable export are consumed with evidence, or until a true hard blocker remains after safe repair/repack/support work has been attempted.

Current completed sources:
- Feature Depth wave: docs/evidence/ao-atlas-feature-depth-wave-v01
- Feature Depth final readback: docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json
- Feature Depth final readback digest: sha256:2f8db366915afd3efdb6c7d15b6c8eafe71dadd14fb4100df2159ff319367e0e
- Final closure assertion: docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json
- Next-track decision: docs/evidence/ao-atlas-refactoring-wave-v01/next-track-decision.json
- Consumed recommendation ledger: docs/evidence/ao-atlas-refactoring-wave-v01/consumed-recommendation-ledger.json
- Consumed recommendation ledger digest: sha256:7fe7731e1a9f9b37219d4492de27ae71b1edc9ca11a093114f60fad466b690b6
- Refactoring recommendations: docs/evidence/ao-atlas-refactoring-wave-v01/refactoring-recommendations.json

Mission objective:
Execute the AO Atlas refactoring recommendations as a long-running bounded wave. Target 2-3 hours of work in a single prompt. Complete at least 12 bounded refactoring nodes before final response unless a true hard blocker remains. If 12 nodes finish quickly and ready work remains, continue toward all 40 recommendations.

Required start behavior:
1. Re-verify ao-atlas is clean, on main, synced with origin/main, with no local or remote codex/* branches.
2. Load and validate the next-track decision.
3. Load and validate refactoring-recommendations.json.
4. Confirm the recommended track is refactoring, Feature Depth is completed_saturated, no promotion is requested, and RSI remains denied.
5. Create a refactoring workgraph from the 40 ranked recommendations.
6. Select exactly one active mutation node at a time.
7. Continue automatically while ready_nodes > 0 or exact_next_action remains.

Ranked refactoring recommendations:
1. refactoring-next-wave-01: Extract recommendation routing command dispatch into a typed registry with deterministic help output.
2. refactoring-next-wave-02: Replace duplicated recommendation command lists with one shared registry-backed command catalog.
3. refactoring-next-wave-03: Bind completed Feature Depth routing decisions to refactoring wave generation without stale exports.
4. refactoring-next-wave-04: Add consumed recommendation ledger checks before any next wave exporter can run.
5. refactoring-next-wave-05: Separate planning-only recommendation export commands from mutation-capable node execution commands.
6. refactoring-next-wave-06: Refactor recommendation evidence schema registry entries into typed constructors with drift tests.
7. refactoring-next-wave-07: Move schema contract constants into grouped recommendation evidence namespaces with coverage tests.
8. refactoring-next-wave-08: Add registry-backed typed validator lookup for recommendation control-plane artifacts.
9. refactoring-next-wave-09: Collapse duplicated schema coverage failure wording into reusable validation helpers.
10. refactoring-next-wave-10: Introduce schema registry golden fixtures for command output and typed validator bindings.
11. refactoring-next-wave-11: Centralize final response gate evaluation for readback, execution readback, and closure readbacks.
12. refactoring-next-wave-12: Refactor exact next action preservation into a shared readback transition helper.
13. refactoring-next-wave-13: Add compact readback status normalization for ready, completed, blocked, and failed node states.
14. refactoring-next-wave-14: Bind return gate denial reasons to structured fields instead of repeated text fragments.
15. refactoring-next-wave-15: Create regression fixtures for stale readback rejection across all recommendation tracks.
16. refactoring-next-wave-16: Unify command run-ledger, rollup, and coverage-check builders behind a common artifact summary type.
17. refactoring-next-wave-17: Refactor run-ledger output status classification into reusable pass, ready, failed, and blocked categories.
18. refactoring-next-wave-18: Add ledger coverage checks for refactoring exporters and track routing artifacts.
19. refactoring-next-wave-19: Bind run-ledger rollups to final operator summaries without self-referential ledger requirements.
20. refactoring-next-wave-20: Create long-run ledger fixture packs for repeated command retries and resumed sessions.
21. refactoring-next-wave-21: Normalize PR and CI ledger rows across feature depth, closure, and refactoring waves.
22. refactoring-next-wave-22: Extract Windows long-running check telemetry into shared threshold and wait-state helpers.
23. refactoring-next-wave-23: Add merge readiness guard helpers that require passed checks before branch cleanup evidence.
24. refactoring-next-wave-24: Refactor post-merge branch deletion readbacks into reusable local and remote cleanup records.
25. refactoring-next-wave-25: Create PR lifecycle replay fixtures for interrupted merge, sync, and cleanup handoffs.
26. refactoring-next-wave-26: Refactor continuation prompt generation to consume structured wave budgets and stop conditions.
27. refactoring-next-wave-27: Add prompt compaction resume fixtures that preserve next node, exact action, and final gate denial.
28. refactoring-next-wave-28: Move prompt safety boundary rendering into one audited template helper.
29. refactoring-next-wave-29: Bind generated prompts to source readback digests and consumed recommendation ledgers.
30. refactoring-next-wave-30: Add long-run prompt regression fixtures for two to three hour refactoring waves.
31. refactoring-next-wave-31: Refactor mission dashboard rows to share provenance digest and freshness evaluation helpers.
32. refactoring-next-wave-32: Add dashboard filters for recommendation track, schema health, CI state, and cleanup state.
33. refactoring-next-wave-33: Bind dashboard closure rows to Promoter, Command, Sentinel, and Foundry rollup evidence.
34. refactoring-next-wave-34: Create stale dashboard evidence detection for superseded wave readbacks and old exports.
35. refactoring-next-wave-35: Add dashboard compact rendering tests for completed waves with no ready nodes.
36. refactoring-next-wave-36: Split oversized recommendation tests into focused files by routing, evidence, readback, and lifecycle domain.
37. refactoring-next-wave-37: Extract shared recommendation test fixture builders for waves, nodes, and readbacks.
38. refactoring-next-wave-38: Add table-driven validator tests for no-promotion and RSI-denied boundary fields.
39. refactoring-next-wave-39: Create targeted regression suites that avoid rerunning unrelated long-wave fixture assertions.
40. refactoring-next-wave-40: Document the refactoring wave handoff with ranked tasks and verification gates.

Per-node evidence requirements:
- node_gate
- candidate_record
- rollback_record
- tests
- verification
- sentinel_public_safety
- promoter_no_promotion or promoter_readback
- command_readback
- run-link
- checkpoint/readback
- PR/CI/merge evidence when remote lifecycle is available
- branch cleanup evidence after merge

Remote lifecycle:
- one active mutation PR at a time
- open PR when GitHub is available
- wait for all required CI checks to pass, including slow Windows checks
- merge only after CI passes
- sync main after merge
- delete local and remote codex/* branches before selecting the next node
- record PR number, merge commit, CI status, and cleanup state

Verification for AO Atlas if touched:
- targeted regression test for the node
- go test ./... -count=1
- go vet ./...
- go build ./cmd/atlas
- scripts/atlas-foundry-roundtrip-smoke.sh
- scripts/production-readiness.sh
- git diff --check
- jq validation over generated evidence
- scoped public-safety wording scan over changed docs/readbacks/tests

Safety boundaries:
- no direct main mutation
- no credential/token inspection or secret exposure
- no provider calls
- no release, deploy, publish, upload, or tag
- no dependency updates
- no auth/policy/config widening
- no hidden instruction mutation
- no broad RSI claim
- no promotion claim
- RSI remains denied

Final response is allowed only when:
- at least 12 bounded refactoring nodes are completed, or all 40 are completed
- no ready node remains within the selected completed budget, or the full 40-node wave is complete
- blocked_nodes=0 unless reporting a true hard blocker
- local verification passes
- GitHub CI passes for merged PRs
- public-safety scan passes
- Promoter says no_promotion_requested or equivalent
- Command readback agrees
- RSI remains denied
- clean/synced repo status is confirmed
- local and remote codex/* branches are deleted

Final report must include:
- completed nodes / total nodes
- node status list
- merged PRs
- final evidence roots
- final Promoter/Command rollup paths
- public-safety scan result
- verification results
- clean/synced repo status
- at least 10 next recommendations after this refactoring wave
- exact next action
```
