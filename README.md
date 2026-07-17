# AO Atlas

AO Atlas compiles authorized, oversized objectives into bounded workgraphs and context packs. It maps dependencies, selects executable nodes, packages only the context each node needs, and emits handoff material for portfolio coordination. Use Atlas when a task needs decomposition, sequencing, or context management beyond a single bounded implementation item.

## How it fits in AO

- **Primary responsibility:** Workgraph and context-pack compilation.
- **Inputs:** Blueprint packs and authorization, Mission objectives, repository context, and completed-node readbacks.
- **Outputs:** Stack instances, workgraphs, context packs, candidate selections, Foundry imports, and run-link readbacks.
- **Upstream:** AO Blueprint and AO Mission.
- **Downstream:** AO Foundry, with progress and closure read back to AO Mission.

See the
[AO Architecture guide](https://github.com/uesugitorachiyo/ao-architecture)
and the
[AO Atlas component page](https://github.com/uesugitorachiyo/ao-architecture/blob/main/components/ao-atlas.md)
for the cross-repository flow.

## Install

```sh
go build -o bin/atlas ./cmd/atlas
```

## Public-Safe Defaults

Tracked examples use relative placeholder paths only. Generated instance state
should live outside this public repo or under ignored local directories such as
`.atlas-local/`.

Tracked repo-relative paths must stay at or below 180 characters. This leaves
space for a native Windows checkout prefix such as `C:\ao\factory\ao-atlas\`
plus Git lock files, temporary suffixes, and generated evidence suffixes without
depending on `core.longpaths`, substituted drives, or unusually short clone
roots. Historical evidence paths compacted for this budget are recorded in
`docs/evidence-path-map.json`; compact identifiers such as `fi`, `bi`, `cp`, and
`t` preserve audit meaning while keeping paths portable.

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
go run ./cmd/atlas mission recommendations import --recommendations examples/valid/ao-mission/feature-depth-recommendations.json --target-instance demo-stack --min-tasks 30 --node-budget 40 --min-minutes 120 --max-minutes 180 --continue-if-fast-target 40 --out .atlas-local/mission-recommendations
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
next recommended prompt plus `recommendation-readback.json`. For 2-3 hour Atlas
work, use `--min-tasks 30`, `--node-budget 40`, `--min-minutes 120`,
`--max-minutes 180`, and `--continue-if-fast-target 40`. With no budget flags,
the importer uses that same v0.2 long-run supervisor default. The generated
workgraph keeps all 40 nodes ready but dependency-chained, so only the first node
is executable-ready until downstream evidence completes the prior node. The wave
records the supervisor lease, checkpoint policy, Promoter/Command/public-safety
readback requirements, and `final_response_allowed=false` while ready nodes or
exact next actions remain. The readback reconciles the wave and workgraph into
node counts, lease health, checkpoint freshness, stale-route decision status,
early-return risk, per-node evidence, the first 10 Feature Depth
recommendations, and the exact next action. `atlas mission recommendations
readback --wave ... --workgraph ... --out ...` regenerates that readback after a
workgraph changes. The wave rejects shallow bundles and any recommendation
artifact that claims execution, scheduling, approval, repository mutation,
provider, credential, direct-main, release, dependency, policy, auth, config, or
authority outside the requested scope. Explicit 20-node/90-minute imports
remain available only as compatibility coverage for older double-size waves.

When an operator says the prior wave was too short, use the 2-3 hour supervisor
defaults instead of a 20-node or 90-minute compatibility wave. Preserve the
lease-start artifact across resumes, execute one Atlas-owned node at a time, and
do not produce a final response while ready nodes or an exact next action remain.
Ready Atlas-owned recommendation tasks skip Blueprint: Atlas already has the
bounded task, safety limits, required gates, and digest-bound source evidence,
so Blueprint is only re-entered when requirements, authorization, candidate
rules, or scope approval are missing.
Foundry import ownership stays single-node: Atlas emits import material for the
current executable node only, Foundry owns that bounded implementation evidence
and run-link, and Atlas resumes from the checkpoint before importing the next
node.
Command readback requirements stay compact and gate-bound: Command reports node
counts, lease status, checkpoint freshness, return gate, exact next action,
first executable node, and final-response allowance without replacing Atlas
workgraph or Foundry run-link evidence.
The detailed operator runbook is
`docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md`.

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
