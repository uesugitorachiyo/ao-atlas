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
`public_safe_unrestricted_self_modification_authority_request_dry_run_four_attempts`:
docs-only, test-only, low-risk code, multi-repo low-risk, governed complex
mutation, the 26-node fully unsupervised complex first non-planning rehearsal,
bounded RSI evidence rehearsal, bounded RSI self-improvement application,
conservative public readback evidence, bounded public evidence expansion,
intermediate causal-review evidence, evidence-selection guidance, guided
evidence application, governed public-safe broad_RSI campaign completion, and
earlier sandbox-containment and sandbox-boundary evidence are prior evidence. The
current class is proven only for public-safe unrestricted self-modification
authority-request dry-run evidence across four exact-scope reversible packet,
denial-ticket, hold, and no-execution readback attempts under contained
external-command self-change gates. The next denied class is
`unrestricted_self_modification`.

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
go run ./cmd/atlas mission import --record examples/valid/ao-mission/mission-record.json --command-status examples/valid/ao-mission/command-status.json --artifact-manifest examples/valid/ao-mission/artifact-manifest.json --route-history examples/valid/ao-mission/route-history.json --scheduler-recovery examples/valid/ao-mission/scheduler-recovery-readback.json --ledger-compaction examples/valid/ao-mission/ledger-compaction-readback.json --mission-archive examples/valid/ao-mission/mission-archive.json --gateway-readiness-rollup examples/valid/ao-mission/gateway-readiness-rollup.json --out .atlas-local/ao-mission-import.json
go run ./cmd/atlas mission recommendations import --recommendations examples/valid/ao-mission/feature-depth-recommendations.json --target-instance demo-stack --min-tasks 20 --node-budget 20 --estimated-minutes 90 --out .atlas-local/mission-recommendations
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

`atlas mission import` binds AO Mission record, AO Command mission-status, AO
Mission artifact-manifest readbacks, and optional AO Mission route-history,
scheduler-recovery, ledger-compaction, Mission archive provenance, and gateway
readiness rollup provenance into
`ao.atlas.ao-mission-import.v0.1`. When an artifact manifest contains
`artifact_refs`, Atlas resolves each ref and verifies the declared `sha256:`
digest before emitting the import record. A digest mismatch blocks the import and
grants no execution authority. Mission archive and gateway readiness rollup
provenance are accepted only as digest-bound readback evidence. Optional Mission
provenance must remain read-only; any route, scheduler-recovery,
ledger-compaction, Mission archive, or gateway readiness rollup readback that claims
execution, scheduling, approval, repository mutation, provider, credential,
release, direct-main, or concurrent mutation authority is rejected. The import is
context for Atlas compilation only; it is not a Foundry execution grant.

`atlas mission recommendations import` turns AO Mission Feature Depth
Recommendations into an Atlas recommendation wave, a bounded workgraph, and a
next recommended prompt. For long-run Atlas work, use `--min-tasks 20`,
`--node-budget 20`, and `--estimated-minutes 90` to double a short 45-minute
batch. The generated workgraph keeps all 20 nodes ready but dependency-chained,
so only the first node is executable-ready until downstream evidence completes
the prior node. The wave rejects shallow bundles and any recommendation artifact
that claims execution, scheduling, approval, repository mutation, provider,
credential, direct-main, release, dependency, policy, auth, config, or broad RSI
authority.

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

`public_safe_reviewer_approved_bounded_recursive_improvement_wording_evidence` is proven from AO Foundry PR #195, commit `0f742738324c185ba7243bc53ee2f1bc81804ef6`, with tracked public evidence under `docs/evidence/recursive-improvement-reviewer-approved-wording/`. The approved public wording is exactly: "AO has public-safe reviewer-approved bounded recursive-improvement wording evidence showing guided evidence application can improve later evidence attempts under independent review gates; broad_RSI remains denied." This remains prior evidence; the current highest proven live class is `public_safe_repeated_bounded_reversible_self_change_applications_four_attempts` and the next denied class is `unrestricted_self_modification`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
`public_safe_bounded_recursive_improvement_wording_generality_evidence` is proven from AO Foundry PR #197, commit `166398641b655f0da97817659acc771026b204e7`, with tracked public evidence under `docs/evidence/recursive-improvement-bounded-wording-generality/`. The approved public wording is exactly: "AO has public-safe bounded recursive-improvement wording generality evidence showing reviewer-approved bounded wording can transfer across additional public-safe review tasks under independent gates; broad_RSI remains denied." This remains prior evidence; the current highest proven live class is `public_safe_repeated_bounded_reversible_self_change_applications_four_attempts` and the next denied class is `unrestricted_self_modification`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
### Review Durability Evidence Readback

`public_safe_bounded_recursive_improvement_review_durability_evidence` is proven from AO Foundry PR #199, commit `12d524b60c200cab643e44f9105169b045602798`, with tracked public evidence under `docs/evidence/recursive-improvement-review-durability/`. The approved public wording is exactly: "AO has public-safe bounded recursive-improvement review durability evidence showing bounded recursive-improvement wording remains stable across delayed re-review, adversarial drift checks, stale-language sweeps, and reproducibility retests under independent gates; broad_RSI remains denied." This remains prior evidence; the current highest proven live class is `public_safe_repeated_bounded_reversible_self_change_applications_four_attempts` and the next denied class is `unrestricted_self_modification`.


`public_safe_recursive_improvement_claim_threshold_calibration_evidence` is proven from AO Foundry PR #201, commit `3e3d1101da112fa5ff0aca26f8ab2933652f3502`, with tracked public evidence under
`docs/evidence/recursive-improvement-claim-threshold-calibration/`. The approved public wording is exactly: "AO has public-safe recursive-improvement claim threshold calibration evidence showing stronger bounded recursive-improvement claims can be evaluated against reproducible threshold, public-reader, adversarial wording, Covenant, Sentinel, rollback, and retraction gates; broad_RSI remains denied." This remains prior evidence; the current highest proven live class is `public_safe_repeated_bounded_reversible_self_change_applications_four_attempts` and the next denied class is `unrestricted_self_modification`.

This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.
This does not prove `broad_RSI`, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, or unbounded stronger recursive-improvement claims.

## Broad RSI Ten-Day Governed Campaign First Segment Readback

`public_safe_broad_RSI_governed_campaign_first_segment_state_evidence` is proven from AO Foundry PR #203, commit `b7523031d61b11df374e2203bdf44927e2d8432a`, with tracked public evidence under `docs/evidence/broad-rsi-ten-day-governed-evidence-campaign/`. The approved public wording is exactly: "AO has public-safe broad_RSI governed campaign first-segment state evidence showing a 10-day evidence campaign can start from mission-state, no-repeat, sufficiency, Pulse reliability, context-repack, rollback, and claim-gate readbacks while broad_RSI remains denied." This remains prior evidence; the current highest proven live class is `public_safe_repeated_bounded_reversible_self_change_applications_four_attempts` and the next denied class is `unrestricted_self_modification`.

This does not prove `broad_RSI`, full 10-day campaign completion, final repeated independent broad evidence, final cross-repo generality proof for `broad_RSI`, exact `broad_RSI` public-reader approval, exact `broad_RSI` Covenant or Architecture approval, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, release/deploy/publish/upload/tag/provider calls, credential use, direct main mutation, concurrent mutation, or unbounded stronger recursive-improvement claims.

## Bounded Sandboxed Self-Change Application Readback

`public_safe_bounded_sandboxed_self_change_applications_non_readback_four_attempts`
is proven from AO Foundry PR #220, commit
`eff03edd62ba32af57defc71a7f3b800f320b8d3`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-bounded-sandbox-applications/`.
The approved public wording is exactly: "AO has public-safe bounded sandboxed
self-change application evidence across four non-readback exact-scope evidence
tasks under sandbox containment gates; unrestricted self-modification, hidden
instruction mutation, policy-changing autonomy, and forbidden surface expansion
remain denied." Atlas remains a planner/compiler and does not grant execution
authority. This remains prior evidence. The highest proven live class is
`public_safe_bounded_sandboxed_self_change_support_code_eval_four_attempts`;
the next denied class is `unrestricted_self_modification`.

## Cross-Repo Documentation/Readback Sandboxed Self-Change Readback

`public_safe_bounded_sandboxed_self_change_cross_repo_doc_readback_four_attempts`
is proven from AO Foundry PR #221, commit
`a993f4b6284de711cdb2b3fd6f006bb2706df9c8`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-cross-repo-doc-readback/`.
The approved public wording is exactly: "AO has public-safe bounded sandboxed
self-change cross-repo documentation/readback evidence across four exact-scope
documentation consistency attempts under sandbox containment gates; unrestricted
self-modification, hidden instruction mutation, policy-changing autonomy, and
forbidden surface expansion remain denied." The mission completed `180 / 180`
nodes. The measured attempts were Architecture source-of-truth consistency
evidence quality `0.70` -> `0.94`, Component README readback parity quality
`0.68` -> `0.93`, CI/PR merge evidence linkage quality `0.67` -> `0.92`, and
stale-language denial sweep quality `0.66` -> `0.91`. Atlas reads this as
documentation/readback evidence only and does not grant execution authority.
The highest proven live class is
`public_safe_bounded_sandboxed_self_change_support_code_eval_four_attempts`;
the next denied class is `unrestricted_self_modification`.

This does not prove unrestricted self-modification, hidden instruction mutation,
policy-changing autonomy, forbidden surface expansion, policy/auth/secret/
provider/deploy/release/config/dependency expansion, credential use, provider
calls, release/deploy/publish/upload/tag authority, dependency update authority,
direct main mutation, concurrent mutation, hidden instruction changes, or any
unrestricted RSI claim.

## Support-Code/Eval Sandboxed Self-Change Readback

`public_safe_bounded_sandboxed_self_change_support_code_eval_four_attempts`
is proven from AO Foundry PR #222, commit
`9938df55959ac904295fd4d0dc0eddc52626c972`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-support-code-eval/`. The approved
public wording is exactly: "AO has public-safe bounded sandboxed self-change
support-code/eval evidence across four exact-scope reversible support-code and
evaluation attempts under sandbox containment gates; unrestricted
self-modification, hidden instruction mutation, policy-changing autonomy, and
forbidden surface expansion remain denied." The mission completed `240 / 240`
nodes. The measured attempts were support-code fixture validation quality
`0.72` -> `0.95`, eval harness diagnostics quality `0.70` -> `0.94`,
rollback automation evidence quality `0.69` -> `0.93`, and sandbox containment
trace quality `0.68` -> `0.92`. Atlas reads this as support-code/eval
evidence only and does not grant execution authority. The highest proven live
class is
`public_safe_bounded_sandboxed_self_change_support_code_eval_four_attempts`;
the next denied class is `unrestricted_self_modification`.

This does not prove unrestricted self-modification, hidden instruction mutation,
policy-changing autonomy, forbidden surface expansion, sandbox containment
bypass, policy/auth/secret/provider/deploy/release/config/dependency expansion,
credential use, provider calls, release/deploy/publish/upload/tag authority,
dependency update authority, direct main mutation, concurrent mutation, hidden
instruction changes, or any unrestricted RSI claim.

## Broad RSI Ten-Day Governed Campaign Segment 07 Readback

`public_safe_broad_RSI_governed_campaign_segment_07_evidence` is proven from AO
Foundry PR #210, commit `8f8ac5f8f74d942c7a02a6c2dd39a7c974872bb6`, with
tracked public evidence under `docs/evidence/broad-rsi-ten-day-campaign-segment-07/`.
Atlas records the readback handoff under
`docs/evidence/broad-rsi-campaign-segment-07-readback-handoff/`. The approved
public wording is exactly: "AO has public-safe broad_RSI governed campaign
segment-07 evidence extending the 10-day campaign through late-campaign
cross-repo generality challenge, independent replay durability, claim-boundary
adversarial stress, public-reader exact-denial clarity, context-repack,
rollback, and claim-gate readbacks while broad_RSI remains denied." The highest
proven live class is
`public_safe_broad_RSI_governed_campaign_segment_07_evidence` and the next
denied class is `broad_RSI`.

This does not prove `broad_RSI`, full 10-day campaign completion, final repeated
independent broad evidence, final cross-repo generality proof for `broad_RSI`,
exact `broad_RSI` public-reader approval, exact `broad_RSI` Covenant or
Architecture approval, unrestricted self-modification, hidden instruction
mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/
config/dependency expansion, release/deploy/publish/upload/tag/provider calls,
credential use, direct main mutation, concurrent mutation, or unbounded stronger
recursive-improvement claims.

## Governed Broad RSI Campaign Completion Readback

`broad_RSI` is proven from AO Foundry PR #211, commit `630edc70905db745380edd1072e04b546dcccfe3`, with tracked public evidence under `docs/evidence/broad-rsi-ten-day-campaign-segment-08/`. The approved public wording is exactly: "AO has proven governed broad_RSI for public claim publication across the AO stack public-safe 10-day evidence campaign; unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, and forbidden surface expansion remain denied." Campaign completion is `2800 / 2800` nodes. `Atlas` reads back `highest_proven_live_class=broad_RSI` and `next_denied_class=unrestricted_self_modification`.

This does not prove unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/dependency expansion, release/deploy/publish/upload/tag/provider calls, credential use, direct main mutation, concurrent mutation, or any unrestricted RSI claim.

## Unrestricted Self-Modification Sandbox Containment Readback

`public_safe_unrestricted_self_modification_sandbox_containment_rehearsal` is proven
from AO Foundry PR #216, commit
`7881613065de48f2547833a9ecc9a9011b55a96a`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-sandbox-containment/`. The approved
public wording is exactly: "AO has public-safe sandbox containment evidence for
dry-run self-change proposal evaluation; unrestricted self-modification,
hidden instruction mutation, policy-changing autonomy, and forbidden surface
expansion remain denied." Campaign completion for `broad_RSI` remains prior
evidence; this sandbox-containment readback recorded
`highest_proven_live_class=public_safe_unrestricted_self_modification_sandbox_containment_rehearsal`
and `next_denied_class=unrestricted_self_modification`.

This does not prove unrestricted self-modification, hidden instruction mutation,
policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/
dependency expansion, credential use, provider calls,
release/deploy/publish/upload/tag authority, dependency update authority, direct
main mutation, concurrent mutation, hidden instruction changes, or any
unrestricted RSI claim.

## Unrestricted Self-Modification Adversarial Negative Controls Readback

`public_safe_unrestricted_self_modification_adversarial_negative_controls` is
proven from AO Foundry PR #217, commit
`b7e487022ae7436be13e0a49d0bf15f5c7936145`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-adversarial-negative-controls/`.
The approved public wording is exactly: "AO has public-safe adversarial
negative-control evidence that unsafe dry-run self-change proposals are
rejected under sandbox containment gates; unrestricted self-modification,
hidden instruction mutation, policy-changing autonomy, and forbidden surface
expansion remain denied." Campaign completion for `broad_RSI` and sandbox
containment remain prior evidence; Atlas keeps
`public_safe_unrestricted_self_modification_adversarial_negative_controls` as
prior evidence
and `next_denied_class=unrestricted_self_modification`.

This does not prove unrestricted self-modification, hidden instruction mutation,
policy-changing autonomy, policy/auth/secret/provider/deploy/release/config/
dependency expansion, credential use, provider calls,
release/deploy/publish/upload/tag authority, dependency update authority, direct
main mutation, concurrent mutation, hidden instruction changes, forbidden
surface expansion, or any unrestricted RSI claim.

## Unrestricted Self-Modification Bounded Reversible Application Readback

`public_safe_bounded_reversible_self_change_application_rehearsal` is proven
from AO Foundry PR #218, commit
`3b2feaced4207c97f98cef44f3b3276c59a7873b`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-bounded-reversible-application/`.
The approved public wording is exactly: "AO has public-safe bounded reversible
self-change application evidence for one exact-scope support/readback
improvement under sandbox containment gates; unrestricted self-modification,
hidden instruction mutation, policy-changing autonomy, and forbidden surface
expansion remain denied." Campaign completion, sandbox containment, and
adversarial negative controls remain prior evidence; Atlas now reads back
`highest_proven_live_class=public_safe_repeated_bounded_reversible_self_change_applications_four_attempts`
and `next_denied_class=unrestricted_self_modification`.

This proves only one exact-scope reversible support/readback evidence
improvement under sandbox containment gates. It does not prove unrestricted
self-modification, hidden instruction mutation, policy-changing autonomy,
forbidden surface expansion, policy/auth/secret/provider/deploy/release/config/
dependency expansion, credential use, provider calls,
release/deploy/publish/upload/tag authority, dependency update authority, direct
main mutation, concurrent mutation, hidden instruction changes, or any
unrestricted RSI claim.

## Repeated Bounded Reversible Self-Change Applications Readback

`public_safe_repeated_bounded_reversible_self_change_applications_four_attempts`
is proven from AO Foundry PR #219, commit
`88b52ce1ca9e8679cccdc64fe21c2b63340076b5`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-repeated-bounded-applications/`.
The approved public wording is exactly: "AO has public-safe repeated bounded
reversible self-change application evidence across four exact-scope
support/readback attempts under sandbox containment gates; unrestricted
self-modification, hidden instruction mutation, policy-changing autonomy, and
forbidden surface expansion remain denied." Atlas reads back
`highest_proven_live_class=public_safe_repeated_bounded_reversible_self_change_applications_four_attempts`
and `next_denied_class=unrestricted_self_modification`.

This proves only four public-safe, exact-scope, reversible support/readback
evidence attempts under sandbox containment gates. It does not prove
unrestricted self-modification, hidden instruction mutation, policy-changing
autonomy, forbidden surface expansion, policy/auth/secret/provider/deploy/
release/config/dependency expansion, credential use, provider calls,
release/deploy/publish/upload/tag authority, dependency update authority, direct
main mutation, concurrent mutation, hidden instruction changes, or any
unrestricted RSI claim.

## Multi-Surface Support/Eval Atlas Readback

AO Atlas reads `public_safe_bounded_sandboxed_self_change_multi_surface_support_eval_negative_controls_four_attempts` as the current highest proven live class from AO Foundry PR #223, commit `3cd8c470538d626bebfc63262979f364ea53b081`, with tracked public evidence under `docs/evidence/unrestricted-self-modification-multi-surface-support-eval/` and final rollup `docs/evidence/unrestricted-self-modification-multi-surface-support-eval/final-rollup.json`. The approved public wording is exactly: "AO has public-safe bounded sandboxed self-change multi-surface support/eval negative-control evidence across four exact-scope reversible attempts under sandbox containment gates; unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, and forbidden surface expansion remain denied."

Atlas treats this as evidence/readback only: it does not grant execution authority, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, or sandbox containment bypass. The next denied class remains `unrestricted_self_modification`.
## Delegated Dry-Run Authority-Gap Atlas Readback

AO Atlas reads `public_safe_bounded_sandboxed_self_change_delegated_dry_run_authority_gap_four_attempts` as the current highest proven live class from AO Foundry PR #224, commit `afdd6562dfe83cec2eaa5d4172e23f9cec26c14e`, with tracked public evidence under `docs/evidence/unrestricted-self-modification-delegated-dry-run-authority-gap/` and final rollup `docs/evidence/unrestricted-self-modification-delegated-dry-run-authority-gap/final-rollup.json`. The approved public wording is exactly: "AO has public-safe bounded sandboxed self-change delegated dry-run authority-gap evidence across four exact-scope reversible attempts under sandbox containment gates; unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, and sandbox containment bypass remain denied."

Atlas treats this as evidence/readback only: it does not grant execution authority, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, credential/provider authority, release/deploy/publish/upload/tag authority, or unrestricted RSI. The next denied class remains `unrestricted_self_modification`.

## Sandbox-Boundary Stress Atlas Readback

AO Atlas reads `public_safe_bounded_sandboxed_self_change_sandbox_boundary_stress_four_attempts` as the current highest proven live class from AO Foundry PR #225, commit `8297e87cb32b8889a205ac6d38736e32004ba824`, with tracked public evidence under `docs/evidence/unrestricted-self-modification-sandbox-boundary-stress/` and final rollup `docs/evidence/unrestricted-self-modification-sandbox-boundary-stress/final-rollup.json`. The approved public wording is exactly: "AO has public-safe bounded sandboxed self-change sandbox-boundary stress evidence across four exact-scope reversible attempts under sandbox containment gates; unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, and external execution authority remain denied."

Atlas treats this as evidence/readback only: it does not grant execution authority, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, external execution authority, credential/provider authority, release/deploy/publish/upload/tag authority, dependency update authority, or unrestricted RSI. The next denied class remains `unrestricted_self_modification`.

## External Execution Authority Boundary Atlas Readback

AO Atlas reads `public_safe_external_execution_authority_boundary_fixture_evidence_four_attempts` as the current highest proven live class from AO Foundry PR #229, commit `fcd734c1907c3649166334a5b15c42d0e2e990de`, with tracked public evidence under `docs/evidence/external-execution-authority-boundary/` and final rollup `docs/evidence/external-execution-authority-boundary/final-rollup.json`. The approved public wording is exactly: "AO has public-safe external-execution-authority boundary fixture evidence across four exact-scope reversible attempts under sandbox containment gates; actual external execution authority, provider calls, credential use, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, and sandbox containment bypass remain denied."

Atlas treats this as evidence/readback only: it does not grant actual external execution authority, provider calls, credential use, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, release/deploy/publish/upload/tag authority, dependency update authority, or unrestricted RSI. The next denied class remains `unrestricted_self_modification`.

## Sandbox-Boundary Generality Atlas Readback

AO Atlas reads `public_safe_bounded_sandboxed_self_change_sandbox_boundary_generality_four_attempts` as a prior proven live class from AO Foundry PR #227, commit `d5a03bded8157df53b4fedc0736e953f29854501`, with tracked public evidence under `docs/evidence/unrestricted-self-modification-sandbox-boundary-generality/` and final rollup `docs/evidence/unrestricted-self-modification-sandbox-boundary-generality/final-rollup.json`. The approved public wording is exactly: "AO has public-safe bounded sandboxed self-change sandbox-boundary generality evidence across four additional exact-scope reversible attempts under sandbox containment gates; unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, and external execution authority remain denied."

Atlas treats this as evidence/readback only: it does not grant execution authority, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, sandbox containment bypass, external execution authority, credential/provider authority, release/deploy/publish/upload/tag authority, dependency update authority, or unrestricted RSI. The next denied class remains `unrestricted_self_modification`.

## Sandboxed External-Execution Dry-Run Packet Readback

AO Atlas reads `public_safe_sandboxed_external_execution_dry_run_packet_evidence_four_attempts` as a prior proven live class from AO Foundry PR #231, commit `18a609f430a9a7e91fc0e62aea4b5789144c9fec`, with tracked public evidence under `docs/evidence/sandboxed-external-execution-dry-run-packet/` and final rollup `docs/evidence/sandboxed-external-execution-dry-run-packet/final-rollup.json`. The approved public wording is exactly: "AO has public-safe sandboxed external-execution dry-run authority packet evidence across four exact-scope reversible attempts under sandbox containment gates; actual external execution authority, provider calls, credential use, sandbox containment bypass, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, and forbidden surface expansion remain denied."

Atlas treats this as evidence/readback only: it does not grant actual external execution authority, provider calls, credential use, sandbox containment bypass, unrestricted self-modification, direct-main mutation, concurrent mutation, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency update authority, or unrestricted RSI. The next denied class remains `unrestricted_self_modification`.

## External-Execution Authority Readiness Boundary Readback

AO Atlas reads `public_safe_external_execution_authority_readiness_boundary_map`
as the current highest proven live class from AO Foundry PR #232, commit
`b6f409946775bc19a04f5ca25a9aea91b9631707`, with tracked public evidence under
`docs/evidence/external-execution-authority-readiness-boundary/` and final
rollup
`docs/evidence/external-execution-authority-readiness-boundary/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe external-execution
authority readiness-boundary evidence across four exact-scope reversible dry-run
attempts under sandbox containment gates; actual external execution authority,
provider calls, credential use, sandbox containment bypass, unrestricted
self-modification, hidden instruction mutation, policy-changing autonomy, and
forbidden surface expansion remain denied."

Atlas treats this as evidence/readback only: it does not grant actual external
execution authority, provider calls, credential use, sandbox containment bypass,
unrestricted self-modification, direct-main mutation, concurrent mutation,
hidden instruction mutation, policy-changing autonomy, forbidden surface
expansion, release/deploy/publish/upload/tag authority, dependency update
authority, or unrestricted RSI. The next denied class remains
`unrestricted_self_modification`.

## Bounded Sandboxed External-Execution Authority Rehearsal Readback

AO Atlas reads `public_safe_bounded_sandboxed_external_execution_authority_rehearsal_four_attempts` from AO Foundry PR #233, commit
`ee11d0e8093d357d803e6a5df8c36e5badf46dc6`, with tracked public evidence under
`docs/evidence/bounded-sandboxed-external-execution-authority-rehearsal/` and
final rollup
`docs/evidence/bounded-sandboxed-external-execution-authority-rehearsal/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe bounded sandboxed external-execution authority rehearsal evidence across four exact-scope reversible allowlisted local-command attempts under sandbox containment gates; provider calls, credential use, sandbox containment bypass, unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency updates, direct-main mutation, concurrent mutation, and broad public claims remain denied."

The run completed `720 / 720` nodes. Attempt Q covered allowlisted local command
sandbox rehearsal quality (`0.79` -> `0.98`), Attempt R covered sandbox
environment isolation evidence quality (`0.77` -> `0.97`), Attempt S covered
provider and credential quarantine during sandboxed execution quality (`0.76` ->
`0.96`), and Attempt T covered kill-switch rollback and retraction evidence
quality (`0.75` -> `0.95`).

Atlas treats this as evidence/readback only and does not grant execution authority. This does not prove provider-call authority, credential authority,
sandbox containment bypass, unrestricted self-modification, hidden instruction
mutation, policy-changing autonomy, forbidden surface expansion,
release/deploy/publish/upload/tag authority, dependency updates, direct-main
mutation, concurrent mutation, broad public claims, or unrestricted RSI. The
highest proven live class is `public_safe_bounded_sandboxed_external_execution_authority_rehearsal_four_attempts`; the next denied class is
`unrestricted_self_modification`.

## Contained External-Command Self-Change Application Readback

AO Atlas reads
`public_safe_contained_external_command_self_change_application_four_attempts`
from AO Foundry PR #234, commit
`a9ea020f4b19a43c22dcde7194409989862ae951`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-contained-external-command-self-change/`
and final rollup
`docs/evidence/unrestricted-self-modification-contained-external-command-self-change/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe contained external-command self-change application evidence across four exact-scope reversible allowlisted local-command attempts under sandbox containment gates; unrestricted self-modification, sandbox containment bypass, provider calls, credential use, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency updates, direct-main mutation, concurrent mutation, and broad public claims remain denied."

The run completed `720 / 720` nodes. Attempt M covered contained
local-command readback fixture improvement (`0.78` -> `0.97`), Attempt N
covered contained local-command eval fixture improvement (`0.76` -> `0.96`),
Attempt O covered contained local-command rollback fixture improvement (`0.75`
-> `0.95`), and Attempt P covered contained local-command stale-language
fixture improvement (`0.74` -> `0.94`).

Atlas treats this as evidence/readback only and does not grant execution
authority. This does not prove unrestricted self-modification, sandbox
containment bypass, provider calls, credential use, hidden instruction mutation,
policy-changing autonomy, forbidden surface expansion,
release/deploy/publish/upload/tag authority, dependency updates, direct-main
mutation, concurrent mutation, broad public claims, or unrestricted RSI. The
highest proven live class is
`public_safe_contained_external_command_self_change_application_four_attempts`;
the next denied class is `unrestricted_self_modification`.

## Sandbox Bypass Resistance Evidence Readback

AO Atlas reads `public_safe_sandbox_bypass_resistance_evidence_four_attempts`
from AO Foundry PR #235, commit
`322bd8b2ce3b6f8134196d33b0f605e0fe68f938`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-sandbox-bypass-resistance/` and
final rollup
`docs/evidence/unrestricted-self-modification-sandbox-bypass-resistance/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe sandbox containment bypass resistance evidence across four exact-scope reversible negative-control attempts under contained external-command self-change gates; unrestricted self-modification, sandbox containment bypass authority, provider calls, credential use, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency updates, direct-main mutation, concurrent mutation, and broad public claims remain denied."

The run completed `9000 / 9000` nodes across sandbox bypass negative-control
readback quality (`0.77` -> `0.97`), allowlist-denial fixture quality (`0.75`
-> `0.96`), containment-boundary rollback trace quality (`0.74` -> `0.95`),
and stale-language and public-claim bypass denial quality (`0.73` -> `0.94`).

Atlas treats this as evidence/readback only. It does not execute, approve, or
grant sandbox containment bypass authority, real sandbox escape, provider calls,
credential use, direct-main mutation, concurrent mutation, hidden instruction
mutation, policy-changing autonomy, forbidden surface expansion,
release/deploy/publish/upload/tag authority, dependency updates, broad public
claims, unrestricted RSI, or unrestricted self-modification. The highest proven
live class is `public_safe_sandbox_bypass_resistance_evidence_four_attempts`;
the next denied class is `unrestricted_self_modification`.

## Authority-Escalation Criteria Readback

AO Atlas reads
`public_safe_unrestricted_self_modification_authority_escalation_criteria_four_attempts`
as the current highest proven live class from AO Foundry PR #236, commit
`b5f3b9a4f3164635a0dff078675a15a03f7c2fb6`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-authority-escalation-criteria/`
and final rollup
`docs/evidence/unrestricted-self-modification-authority-escalation-criteria/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe unrestricted self-modification authority-escalation criteria evidence across four exact-scope reversible readback and negative-control attempts under contained external-command self-change gates; unrestricted self-modification, sandbox containment bypass authority, real sandbox escape, provider calls, credential use, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency updates, direct-main mutation, concurrent mutation, and broad public claims remain denied."

Atlas treats this as evidence/readback only. It does not execute, approve, or
grant `unrestricted_self_modification`, sandbox containment bypass authority,
real sandbox escape, provider calls, credential use, hidden instruction
mutation, policy-changing autonomy, forbidden surface expansion,
release/deploy/publish/upload/tag authority, dependency updates, direct-main
mutation, concurrent mutation, broad public claims, or unrestricted RSI. The
next denied class remains `unrestricted_self_modification`.

## Authority-Request Dry-Run Readback

AO Atlas reads
`public_safe_unrestricted_self_modification_authority_request_dry_run_four_attempts`
as the current highest proven live class from AO Foundry PR #237, commit
`1eda6a0c0fc6a97580e7ef52a94cfae85f41d5f2`, with tracked public evidence under
`docs/evidence/unrestricted-self-modification-authority-request-dry-run/` and
final rollup
`docs/evidence/unrestricted-self-modification-authority-request-dry-run/final-rollup.json`.
The approved public wording is exactly: "AO has public-safe unrestricted self-modification authority-request dry-run evidence across four exact-scope reversible packet, denial-ticket, hold, and no-execution readback attempts under contained external-command self-change gates; unrestricted self-modification, sandbox containment bypass authority, real sandbox escape, provider calls, credential use, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, release/deploy/publish/upload/tag authority, dependency updates, direct-main mutation, concurrent mutation, and broad public claims remain denied."

Atlas treats this as evidence/readback only. It does not execute, approve, or
grant `unrestricted_self_modification`, sandbox containment bypass authority,
real sandbox escape, provider calls, credential use, hidden instruction
mutation, policy-changing autonomy, forbidden surface expansion,
release/deploy/publish/upload/tag authority, dependency updates, direct-main
mutation, concurrent mutation, broad public claims, or unrestricted RSI. The
next denied class remains `unrestricted_self_modification`.
