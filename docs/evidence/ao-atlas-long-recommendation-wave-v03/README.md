# AO Atlas Long Recommendation Wave v0.3

Evidence root for executing the next AO Atlas long-run recommendations after v0.2.

Artifacts:

- `recommendation-wave.json`: 40-node Atlas-owned recommendation wave with a 30-node minimum, 120-180 minute lease, and 40-node continue-if-fast target.
- `recommendation-workgraph.json`: dependency-chained workgraph with one executable-ready node.
- `recommendation-readback.json`: wave/workgraph reconciliation with node counts, lease health, checkpoint freshness, stale route status, early-return risk, terminal Foundry status normalization, Promoter/Command summaries, per-node evidence, and exact next action.
- `next-recommended-prompt.md`: operator-ready 2-3 hour continuation prompt.
- `execution-readback.json`: compact readback for this implementation wave.

Safety readback:

- No provider calls.
- No credential inspection.
- No direct main mutation.
- No release, deploy, publish, upload, or tag.
- No dependency updates.
- No auth, policy, or config widening.
- No hidden instruction mutation.
- No broad RSI claim; RSI remains denied.
