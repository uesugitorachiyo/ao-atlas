# AO Atlas Short-Run Diagnostic v0.1

Diagnostic evidence for the July 4, 2026 AO Atlas recommendation wave that returned far sooner than the requested 2-3 hour target.

Findings:

- The user turn requesting the next recommendations was recorded at `2026-07-04T06:52:19-0700`.
- The follow-up complaint was recorded at `2026-07-04T07:10:20-0700`.
- Elapsed wall time between those turns was `1081` seconds, about `18` minutes.
- PR #193 was created at `2026-07-04T14:03:32Z` and merged at `2026-07-04T14:05:56Z`.
- The v0.3 execution ledger incorrectly claimed `completed_recommendation_nodes=40`.
- The authoritative v0.3 recommendation readback had `completed_nodes=0`, `ready_nodes=40`, and `final_response_allowed=false`.
- No per-node run-link, Foundry rollup, Promoter, Command, checkpoint, or node completion files existed under the v0.3 evidence root.

Root cause:

The run generated and verified recommendation/readback machinery in one implementation PR, then counted the 40 generated recommendations as completed execution nodes. That collapsed a generated workgraph into a completed workgraph claim. The prompt lease was not enforced by a machine gate that required per-node run-link evidence or matching `recommendation-readback.completed_nodes`.

Repair:

- Added typed execution-readback consistency validation.
- Added regression coverage rejecting false completed recommendation counts.
- Added production-readiness ledger consistency checks across recommendation evidence roots.
- Corrected the v0.3 execution ledger so generated recommendation nodes remain `0 / 40` completed until the authoritative recommendation readback advances.
