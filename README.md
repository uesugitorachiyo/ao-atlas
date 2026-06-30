# AO Atlas

AO Atlas is a local-first stack-instance and workgraph compiler for the AO
stack. It turns oversized objectives into bounded factory tasks, context packs,
and Foundry handoff material without duplicating whole AO source trees.

AO Atlas is not a task runner, scheduler, approver, provider client, release
publisher, or control plane. It prepares public-safe, evidence-bound inputs for
the rest of the AO stack:

- AO Blueprint owns requirements interview and build authorization.
- AO Atlas owns oversized objective intake, workgraph/context-pack compilation,
  stack-instance manifests, and factory-folder materialization models.
- AO Foundry owns portfolio scheduling and safe next-action selection.
- AO Forge owns one governed factory run.
- AO2 executes governed local work.
- AO Command remains read-only.
- AO Covenant, Sentinel, Promoter, Arena, and Crucible remain gates.

For the docs-only live mutation classes, AO Atlas can decompose the oversized
objective, compile bounded context packs, and emit Foundry import or run-link
evidence. The governed live ladder has now proven `docs_only_multi_file` through
a docs-only PR rehearsal. The next designed rung is the `test_only` dry-run
chain: at most one test file in later live rehearsal, with tests-only write
scope, rollback plan, Sentinel coverage no-hold evidence, Promoter readiness,
Command readback, and clean CI. That evidence may feed the later Covenant,
Foundry, Forge, AO2, Sentinel, Promoter, and Command approval chain, but Atlas
does not grant mutation authority, mark work safe to execute, create branches,
apply patches, publish, release, or widen the approved scope.

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
