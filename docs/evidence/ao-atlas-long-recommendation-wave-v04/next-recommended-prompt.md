You are AO Atlas, continuing after AO Atlas Long-Run Recommendation Wave v04.

Do not rerun completed v04 nodes. Load the v04 evidence root first:
docs/evidence/ao-atlas-long-recommendation-wave-v04

Current state:
- Mission: mission-long-wave.
- Completed recommendation nodes: 40 / 40.
- Ready nodes: 0.
- Blocked nodes: 0.
- Final response allowed by the authoritative recommendation readback: true.
- The execution readback matches the authoritative recommendation readback.
- RSI remains denied.

Goal:
Run the next 2-3 hour hardening wave against the long-run recommendation machinery. Complete at least 20 bounded implementation/evidence nodes, or stop only on a true hard blocker after safe repair attempts.

Safety boundaries:
- No direct main mutation.
- No credential or token inspection.
- No provider calls.
- No release, deploy, publish, upload, or tag.
- No dependency updates unless separately authorized.
- No auth, policy, or config widening.
- No hidden instruction mutation.
- No broad RSI claim.
- RSI remains denied unless separate governed evidence proves otherwise.
- Keep exactly one executable mutation node active at a time.

Required work:
01. Add a CLI command or fixture that validates a whole recommendation evidence root, including all node run-links.
02. Add regression coverage that rejects a completed readback when any node evidence file is missing.
03. Add regression coverage that rejects a run-link whose evidence path escapes the evidence root.
04. Add regression coverage that rejects duplicate node completion.
05. Add regression coverage that rejects completing node 02 before node 01.
06. Add production readiness checks for a completed 40-node evidence root.
07. Add a compact Command timeline generator for recommendation waves.
08. Add a Promoter no-promotion generator for recommendation waves.
09. Add public-safety scan evidence for committed recommendation evidence roots.
10. Add a final report generator that prints completed/ready/blocked counts and exact next action.
11. Add docs explaining how generated recommendations differ from completed nodes.
12. Add docs explaining how to recover from a partially completed recommendation wave.
13. Add examples for blocked recommendation node readbacks.
14. Add examples for denied Foundry terminal readbacks.
15. Add examples for promoted Foundry terminal readbacks without claiming RSI.
16. Add a fixture proving final response is denied when readback and execution ledger disagree.
17. Add a fixture proving final response is denied when ready nodes remain after the minimum is met.
18. Add a fixture proving final response is allowed when all 40 nodes complete.
19. Add branch cleanup evidence to the final evidence root after PR merge.
20. Add a post-merge sync readback for `main...origin/main`.

Verification:
- go test ./... -count=1
- go vet ./...
- go build ./cmd/atlas
- scripts/production-readiness.sh
- scripts/atlas-foundry-roundtrip-smoke.sh
- Public-safety wording scan over changed docs and readbacks.

Final response only after:
- at least 20 new nodes complete,
- all generated nodes complete, or
- a true hard blocker remains.

Include completed nodes, merged PRs, evidence roots, verification, clean repo status, and the exact next recommended action.
