# AO Atlas

AO Atlas is a local-first stack-instance and workgraph compiler for the AO
stack. It turns oversized objectives into bounded factory tasks, context packs,
and Foundry handoff material without duplicating whole AO source trees.

## AO Stack Architecture

This repository is part of the AO agent orchestration stack. Start with the
central architecture guide at
[uesugitorachiyo/ao-architecture](https://github.com/uesugitorachiyo/ao-architecture);
the AO Atlas-specific architecture page is
[ao-atlas](https://github.com/uesugitorachiyo/ao-architecture/tree/main/ao-atlas).

Canonical upstream intake is AO Blueprint -> AO Atlas -> AO Foundry. Blueprint
owns requirements interview and build authorization. Atlas compiles authorized
oversized objectives into stack instances, workgraphs, context packs,
candidate-selection records, Foundry handoff material, and run-link readback.
Foundry validates those artifacts and decides whether a ready item should be
delegated, but it does not treat raw operator ideas or underspecified Atlas
material as implementation-ready work.

AO Atlas is not a task runner, scheduler, approver, provider client, release
publisher, or control plane. It prepares public-safe, evidence-bound inputs for
the rest of the AO stack:

- AO Blueprint owns requirements interview and build authorization.
- AO Atlas owns oversized objective intake, workgraph/context-pack compilation,
  stack-instance manifests, Blueprint import compilation, and factory-folder
  materialization models.
- AO Foundry owns portfolio scheduling and safe next-action selection.
- AO Forge owns one governed factory run.
- AO2 executes governed local work.
- AO Command remains read-only.
- AO Covenant, Sentinel, Promoter, Arena, and Crucible remain gates.

For the governed live mutation ladder, AO Atlas can decompose oversized
objectives, compile bounded context packs, and emit Foundry import or run-link
evidence. The highest proven live class is now
`public_safe_reviewer_approved_bounded_recursive_improvement_wording_evidence`: docs-only, test-only,
low-risk code, multi-repo low-risk, governed complex mutation, the 26-node fully
unsupervised complex first non-planning rehearsal, bounded RSI evidence
rehearsal, bounded RSI self-improvement application, conservative public
readback evidence, bounded public evidence expansion, and intermediate
causal-review evidence, evidence-selection guidance, and guided evidence
application are prior evidence. The next denied class is `broad_RSI`.

Atlas also holds the 32-node bounded RSI evidence workgraph as evidence for
`bounded_rsi_evidence_rehearsal`. That workgraph supports the final bounded
evidence rehearsal closure only; it is not broad RSI, does not authorize
unrestricted self-modification, and does not allow hidden instruction mutation
or policy/auth/secret/provider/deploy/release/config/dependency expansion. The
evidence may feed the Covenant, Foundry, Forge, AO2, Sentinel, Promoter, and
Command approval chain, but Atlas does not grant mutation authority, mark work
safe to execute, create branches, apply patches, publish, release, or widen the
approved scope.

Atlas also holds the 36-node bounded RSI self-improvement application workgraph.
That workgraph supports only `bounded_rsi_self_improvement_application` for the
exact private readback/eval rubric rehearsal. It does not prove broad RSI, does
not authorize unrestricted self-modification, does not allow hidden instruction
mutation, and does not expand policy/auth/secret/provider/deploy/release/config/
dependency authority or policy-changing autonomy.

Atlas also recognizes the later exact safe public claim wording closure:
`exact_safe_public_claim_wording_conservative_readback_evidence`. The approved
public wording is exactly: "AO has public-safe tracked readback evidence for
bounded improvement-claim review and retraction rehearsal; stronger
recursive-improvement claims remain denied." Atlas treats this as conservative
readback evidence only; it does not prove `broad_RSI`, stronger
recursive-improvement claims, unrestricted self-modification, hidden instruction
mutation, or policy-changing autonomy.

Atlas also recognizes the public-safe bounded improvement evidence expansion closure:
`public_safe_bounded_improvement_evidence_expansion_four_attempts`. It tracks four
public-safe bounded evidence expansion attempts with reproducibility runbooks and
keeps stronger recursive-improvement wording, `broad_RSI`, unrestricted
self-modification, hidden instruction mutation, and policy-changing autonomy
denied.

## Install

```sh
go build -o bin/atlas ./cmd/atlas
```

## Public-Safe Defaults

Tracked examples use relative placeholder paths only. Generated instance state
should live outside this public repo or under ignored local directories such as
`.atlas-local/`.

`atlas instance doctor` inspects stack-instance hygiene without scheduling or
executing work. It validates ignored local state roots, generated registry
parity, shared-toolchain use, worktree bounds, and `schedules_work=false`,
`executes_work=false`, `approves_work=false`. When `--registry` is omitted, the
doctor compares against the registry Atlas would emit from the instance.

AO Atlas v0.1 does not call live providers, push, tag, release, upload, or copy
source repos.

## Quick Start

```sh
go run ./cmd/atlas instance init \
  --id demo-stack \
  --state-root .atlas-local/state \
  --toolchain-root ../shared-ao-toolchain \
  --out .atlas-local/demo-stack.instance.json

go run ./cmd/atlas instance validate --instance examples/valid/stack-instance.json
go run ./cmd/atlas instance doctor --instance examples/valid/stack-instance.json --registry examples/valid/atlas-registry.json --out .atlas-local/instance-doctor.json
go run ./cmd/atlas instance doctor --instance examples/valid/stack-instance.json --json
go run ./cmd/atlas mission status --intake examples/valid/intake.json --workgraph examples/valid/workgraph-completed.json --run-link examples/valid/run-link.json --out .atlas-local/mission-status.json
go run ./cmd/atlas mission status --intake examples/valid/intake.json --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link-needs-context.json --json
go run ./cmd/atlas blueprint-request validate --request examples/valid/blueprint-request.json
go run ./cmd/atlas blueprint import --pack examples/valid/blueprint-import-low-risk-code/blueprint-pack --authorization examples/valid/blueprint-import-low-risk-code/build-authorization.json --instance examples/valid/stack-instance.json --mutation-classes examples/valid/mutation-classes.json --out .atlas-local/blueprint-import-low-risk-code
go run ./cmd/atlas mutation-classes validate --model examples/valid/mutation-classes.json
go run ./cmd/atlas factory materialize --task examples/valid/factory-task.json --out .atlas-local/factory-materialization --dry-run
go run ./cmd/atlas workgraph next --workgraph examples/valid/workgraph.json --json
go run ./cmd/atlas workgraph validate --workgraph examples/valid/workgraph-large-stress.json
go run ./cmd/atlas workgraph materialize-next --workgraph examples/valid/workgraph.json --out .atlas-local/workgraph-next-materialization --dry-run
go run ./cmd/atlas workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link.json --out .atlas-local/workgraph-completed.json
go run ./cmd/atlas workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-blocked.json --out .atlas-local/workgraph-repair-plan.json
go run ./cmd/atlas context-pack repack --task examples/valid/factory-task.json --run-link examples/valid/run-link-needs-context.json --source-ref docs/sdd/AO-ATLAS-CONTEXT-PACKS.md --source-digest sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --out .atlas-local/context-pack-repacked.json
go run ./cmd/atlas context-pack validate --pack examples/valid/context-pack.json
go run ./cmd/atlas foundry handoff emit --workgraph examples/valid/workgraph.json --out .atlas-local/foundry-handoff.json
go run ./cmd/atlas foundry import --workgraph examples/valid/workgraph.json --instance examples/valid/stack-instance.json --out .atlas-local/foundry-import
go run ./cmd/atlas run-link attach --task-id atlas-readiness-task --status completed --evidence ao2=evidence/ao2/atlas-readiness.json --out .atlas-local/run-link.json
```

Ready `blueprint import` output and direct `foundry import` output both write
Foundry import material plus an operator-ready continuation handoff:
`foundry-continuation-handoff.json` and `foundry-continuation-prompt.md`.
Atlas final reports should direct the operator to move to the AO Foundry
checkout, run `codex --yolo`, and paste the generated prompt. The next action
is not an inspection-only command against the import file.

With a sibling AO Foundry checkout, Atlas can run the fixture-only handoff,
Foundry import validation, and Foundry observer-readback smoke:

```sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

If intake is underspecified, Atlas emits a Blueprint request instead of marking
work ready. The request is a clarification artifact only; AO Blueprint still
owns requirements interview and build authorization.

The committed
`examples/valid/workgraph-repair-plan-blocked-node-demo.json` fixture shows the
blocked-node repair output shape. It preserves the source task's context-pack
refs while keeping `schedules_work=false`, `executes_work=false`, and
`approves_work=false`; Atlas emits repair material only and does not schedule or
execute it.

The committed
`examples/valid/context-pack-needs-context-repack-demo.json` fixture shows the
needs-context repack output shape. It carries source refs and digests,
assumptions, exclusions, and `missing_context_reason` so the next factory run
knows why the bounded context was regenerated.

The committed `examples/valid/workgraph-large-stress.json` fixture exercises a
larger 12-node mission with completed, ready, blocked, and stitch nodes. It is
used by production readiness to prove sequencing, context-pack refs, and
Foundry-import behavior without duplicating AO stack folders.

## Readiness

```sh
scripts/production-readiness.sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

The readiness gate must report `score=100/100` before v0.1 is considered ready.
Before a stable tag or release candidate, also run the sibling Foundry
roundtrip smoke so the Atlas -> Foundry boundary is proven end to end.

`examples/valid/low-risk-code-denial-audit.json` is the read-only Atlas audit
for the current authority boundary: `test_only` is the highest proven live
class, `low_risk_code` may be requested for dry-run/readback work, and live
code execution remains denied until exact policy promotion, rollback proof,
Sentinel clear verdict, Promoter promotion, Command readback, and PR CI evidence
exist for the class.
`examples/valid/workgraph-complex-repo-mutation-rehearsal.json` remains
dry-run only. It now records low-risk decomposition and rollback graph nodes
before repair and promotion gates, while still exposing only the first
dependency-safe ready node to Foundry import.

`public_safe_bounded_improvement_evidence_expansion_four_attempts` remains prior
evidence from AO Foundry PR #181, commit
`d31b6f2247780867c3c72dbda5abb7377f3a1b3e`, with tracked public evidence under
`docs/evidence/recursive-improvement-public-evidence-expansion/`. Four
public-safe bounded evidence expansion attempts are tracked with reproducibility
runbooks: release/readiness evidence quality (`0.68` -> `0.91`), security/public-
safety scan quality (`0.64` -> `0.90`), operator readback UX (`0.62` -> `0.88`),
and cross-repo evidence linking (`0.60` -> `0.87`). Stronger
recursive-improvement wording remains denied, `broad_RSI` remains denied,
unrestricted self-modification remains denied, hidden instruction mutation
remains denied, and policy-changing autonomy remains denied.

`public_safe_intermediate_causal_review_claim_evidence` remains prior evidence
from AO Foundry PR #189, commit
`860e3f353ab833c4a671b9d0ee6d8101ece2815c`, with tracked public evidence under
`docs/evidence/recursive-improvement-safe-intermediate-claim/`. The approved public wording is exactly: "AO has public-safe intermediate causal-review evidence that bounded improvement evidence can guide and constrain later claim review across independent roles; stronger recursive-improvement wording and broad_RSI remain denied." Stronger recursive-improvement wording remains denied, `broad_RSI` remains denied, unrestricted self-modification remains denied, hidden instruction mutation remains denied, and policy-changing autonomy remains denied.

`public_safe_causal_review_evidence_selection_guidance` is proven from AO Foundry
PR #191, commit `413b70f15d8f3d0203dc7be076914a2f3b539881`, with tracked public
evidence under `docs/evidence/recursive-improvement-evidence-selection-guidance/`.
The approved public wording is exactly: "AO has public-safe causal-review
evidence that prior bounded evidence can guide later evidence-selection and
blocker prioritization under independent review gates; stronger
recursive-improvement wording and broad_RSI remain denied." This remains prior
evidence. Stronger recursive-improvement wording remains denied, `broad_RSI`
remains denied, unrestricted self-modification remains denied, hidden
instruction mutation remains denied, and policy-changing autonomy remains
denied.

`public_safe_guided_evidence_application_four_attempts` is proven from AO Foundry
PR #193, commit `4ec509fd64d1fc1ea41ea7f22aae900ba79e09a1`, with tracked public
evidence under `docs/evidence/recursive-improvement-guided-evidence-application/`.
The approved public wording is exactly: "AO has public-safe guided
evidence-application evidence showing causal-review guidance can select and
prioritize later bounded evidence attempts under independent gates; stronger
recursive-improvement wording and broad_RSI remain denied." The highest proven
live class is `public_safe_recursive_improvement_claim_threshold_calibration_evidence` and the
next denied class is `broad_RSI`. Stronger recursive-improvement wording
remains denied, `broad_RSI` remains denied, unrestricted self-modification
remains denied, hidden instruction mutation remains denied, and policy-changing
autonomy remains denied.

## Public-Safe Reviewer-Approved Bounded Wording Evidence

`public_safe_reviewer_approved_bounded_recursive_improvement_wording_evidence` is proven from AO Foundry PR #195, commit `0f742738324c185ba7243bc53ee2f1bc81804ef6`, with tracked public evidence under `docs/evidence/recursive-improvement-reviewer-approved-wording/`. The approved public wording is exactly: "AO has public-safe reviewer-approved bounded recursive-improvement wording evidence showing guided evidence application can improve later evidence attempts under independent review gates; broad_RSI remains denied." The highest proven live class is `public_safe_recursive_improvement_claim_threshold_calibration_evidence` and the next denied class is `broad_RSI`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
`public_safe_bounded_recursive_improvement_wording_generality_evidence` is proven from AO Foundry PR #197, commit `166398641b655f0da97817659acc771026b204e7`, with tracked public evidence under `docs/evidence/recursive-improvement-bounded-wording-generality/`. The approved public wording is exactly: "AO has public-safe bounded recursive-improvement wording generality evidence showing reviewer-approved bounded wording can transfer across additional public-safe review tasks under independent gates; broad_RSI remains denied." The highest proven live class is `public_safe_recursive_improvement_claim_threshold_calibration_evidence` and the next denied class is `broad_RSI`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
### Review Durability Evidence Readback

`public_safe_bounded_recursive_improvement_review_durability_evidence` is proven from AO Foundry PR #199, commit `12d524b60c200cab643e44f9105169b045602798`, with tracked public evidence under `docs/evidence/recursive-improvement-review-durability/`. The approved public wording is exactly: "AO has public-safe bounded recursive-improvement review durability evidence showing bounded recursive-improvement wording remains stable across delayed re-review, adversarial drift checks, stale-language sweeps, and reproducibility retests under independent gates; broad_RSI remains denied." The highest proven live class is `public_safe_recursive_improvement_claim_threshold_calibration_evidence` and the next denied class is `broad_RSI`.


`public_safe_recursive_improvement_claim_threshold_calibration_evidence` is proven from AO Foundry PR #201, commit `3e3d1101da112fa5ff0aca26f8ab2933652f3502`, with tracked public evidence under
`docs/evidence/recursive-improvement-claim-threshold-calibration/`. The approved public wording is exactly: "AO has public-safe recursive-improvement claim threshold calibration evidence showing stronger bounded recursive-improvement claims can be evaluated against reproducible threshold, public-reader, adversarial wording, Covenant, Sentinel, rollback, and retraction gates; broad_RSI remains denied." The highest proven live class is `public_safe_recursive_improvement_claim_threshold_calibration_evidence` and the next denied class is `broad_RSI`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
