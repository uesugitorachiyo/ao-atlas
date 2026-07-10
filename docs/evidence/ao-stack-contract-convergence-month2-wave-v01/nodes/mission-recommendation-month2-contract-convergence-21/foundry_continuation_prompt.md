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
- Atlas import: not_provided
- Atlas workgraph: docs/evidence/ao-stack-contract-convergence-month2-wave-v01/nodes/mission-recommendation-month2-contract-convergence-20/workgraph-after.json
- Foundry import: docs/evidence/ao-stack-contract-convergence-month2-wave-v01/nodes/mission-recommendation-month2-contract-convergence-21/foundry-import/foundry-import.json

Current Atlas readback:
- first safe node: mission-recommendation-month2-contract-convergence-21
- total nodes: 40
- completed nodes: 20
- ready nodes: 20
- blocked nodes: 0
- handoff phase: pre_node_execution; the counts above are the checkpoint before the active node completes
- expected post-node readback: 21 completed, 19 ready, next node mission-recommendation-month2-contract-convergence-22, final response remains denied
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
