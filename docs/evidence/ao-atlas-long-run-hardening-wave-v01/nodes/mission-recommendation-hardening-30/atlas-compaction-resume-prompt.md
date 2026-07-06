# AO Atlas Compaction Resume Prompt

Resume from latest durable checkpoint:
`docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-29/recommendation-readback-after.json`

If context was compacted, trust the checkpoint readback and continue from the exact next action recorded there.

State readback:
- completed_nodes=29
- ready_nodes=11
- blocked_nodes=0
- final_response_allowed=false
- Next executable node: mission-recommendation-hardening-30

Execution rule:
- Do not execute completed nodes 1 through 29.
- Keep exactly one executable mutation node active.
- Execute only mission-recommendation-hardening-30 until its run-link, verification, PR, CI, merge, sync, and branch cleanup evidence are recorded.
- Continue to mission-recommendation-hardening-31 after node 30 closes.

Safety:
- No provider calls, credential inspection, direct main mutation, release, deploy, publish, upload, tag, dependency update, auth or policy widening, hidden instruction mutation, or broad RSI claim.
- RSI remains denied.
