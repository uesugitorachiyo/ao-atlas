# AO Foundry Continuation Handoff

Move to AO Foundry:

```sh
cd ../ao-foundry
codex --yolo
```

Paste this prompt:

```text
You are AO Foundry. Continue from the AO Atlas first-phase handoff.

Source artifacts:
- Blueprint pack: not_provided
- Atlas import: docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/recommendation-wave.json
- Atlas workgraph: docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-23/workgraph-after.json
- Foundry import: docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-24/foundry-import.json
- Mission continuation evidence: docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-23/recommendation-readback-after.json

Current Atlas readback:
- first safe node: mission-recommendation-final-closure-consolidation-24
- total nodes: 24
- completed nodes: 23
- ready nodes: 1
- blocked nodes: 0
- class boundary: Atlas import only for low_risk_code; Foundry must preserve Atlas no-execution boundary

Required continuation behavior:
- Move to AO Foundry.
- Run codex --yolo.
- Paste this prompt.
- Import and validate the Foundry import.
- do not stop after import validation.
- do not stop after one gate artifact.
- do not stop after one node.
- Continue until all generated slices/tasks/nodes are consumed or a true hard blocker remains.
- If evidence/schema/readback support is missing and can be safely implemented, implement it with PR/CI/merge.

Hard safety prohibitions:
- Atlas must not execute live mutation
- no direct main mutation
- no release deploy publish upload tag provider call credential use dependency update auth policy widening secret env exposure or config expansion
- do not claim fully_unsupervised_complex_mutation or RSI proof
- do not claim complex_repo_mutation is live-proven unless downstream evidence proves it
- fully_unsupervised_complex_mutation remains denied.
- RSI remains denied.

Stop conditions:
- done
- final denial
- hard blocker
- CI failure
- unsafe scope drift
- kill switch

Stop only on done, final denial, hard blocker, CI failure, unsafe scope drift, or kill switch.
```
