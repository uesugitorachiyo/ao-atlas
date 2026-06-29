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

## Install

```sh
go build -o bin/atlas ./cmd/atlas
```

## Public-Safe Defaults

Tracked examples use relative placeholder paths only. Generated instance state
should live outside this public repo or under ignored local directories such as
`.atlas-local/`.

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
go run ./cmd/atlas mission status --intake examples/valid/intake.json --workgraph examples/valid/workgraph-completed.json --run-link examples/valid/run-link.json --out .atlas-local/mission-status.json
go run ./cmd/atlas blueprint-request validate --request examples/valid/blueprint-request.json
go run ./cmd/atlas factory materialize --task examples/valid/factory-task.json --out .atlas-local/factory-materialization --dry-run
go run ./cmd/atlas workgraph next --workgraph examples/valid/workgraph.json --json
go run ./cmd/atlas workgraph materialize-next --workgraph examples/valid/workgraph.json --out .atlas-local/workgraph-next-materialization --dry-run
go run ./cmd/atlas workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link.json --out .atlas-local/workgraph-completed.json
go run ./cmd/atlas workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-blocked.json --out .atlas-local/workgraph-repair-plan.json
go run ./cmd/atlas context-pack repack --task examples/valid/factory-task.json --run-link examples/valid/run-link-needs-context.json --source-ref docs/sdd/AO-ATLAS-CONTEXT-PACKS.md --source-digest sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --out .atlas-local/context-pack-repacked.json
go run ./cmd/atlas context-pack validate --pack examples/valid/context-pack.json
go run ./cmd/atlas foundry handoff emit --workgraph examples/valid/workgraph.json --out .atlas-local/foundry-handoff.json
go run ./cmd/atlas foundry import --workgraph examples/valid/workgraph.json --out .atlas-local/foundry-import
go run ./cmd/atlas run-link attach --task-id atlas-readiness-task --status completed --evidence ao2=evidence/ao2/atlas-readiness.json --out .atlas-local/run-link.json
```

With a sibling AO Foundry checkout, Atlas can run the fixture-only handoff
readback smoke:

```sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

If intake is underspecified, Atlas emits a Blueprint request instead of marking
work ready. The request is a clarification artifact only; AO Blueprint still
owns requirements interview and build authorization.

## Readiness

```sh
scripts/production-readiness.sh
```

The readiness gate must report `score=100/100` before v0.1 is considered ready.
