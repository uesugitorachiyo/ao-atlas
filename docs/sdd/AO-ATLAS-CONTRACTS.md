# AO Atlas Contracts

v0.1 contract names:

- `ao.atlas.stack-instance.v0.1`
- `ao.atlas.foundry-registry.v0.1`
- `ao.atlas.instance-doctor.v0.1`
- `ao.atlas.intake.v0.1`
- `ao.atlas.mission-status.v0.1`
- `ao.atlas.blueprint-request.v0.1`
- `ao.atlas.workgraph.v0.1`
- `ao.atlas.workgraph-repair-plan.v0.1`
- `ao.atlas.factory-task.v0.1`
- `ao.atlas.factory-materialization.v0.1`
- `ao.atlas.context-pack.v0.1`
- `ao.atlas.foundry-handoff.v0.1`
- `ao.atlas.foundry-import.v0.1`
- `ao.atlas.run-link.v0.1`

All contracts are JSON. Validation requires explicit `contract_version` values,
stable identifiers, non-empty required fields, and public-safe paths.

`ao.atlas.instance-doctor.v0.1` records stack-instance hygiene readback:
instance validation, generated registry parity, ignored local state placement,
bounded worktree roots, shared-toolchain use, and authority boundaries. The
doctor reports `ready`, `blocked`, or `failed`, including the first failing
check and blocking next actions. It does not schedule, execute, approve,
publish, call providers, or mutate sibling repositories.

`ao.atlas.mission-status.v0.1` summarizes intake, workgraph node state, run-link
state, completion status, and next readback actions. It is status readback only
and does not schedule or execute work.

`ao.atlas.blueprint-request.v0.1` is emitted when intake is not specific
enough to compile into a workgraph. It records the intake id, missing fields,
and reason for routing back to AO Blueprint. It is not build authorization and
does not allow Atlas, Foundry, or AO2 to schedule or execute work.

`ao.atlas.factory-materialization.v0.1` records a dry-run factory skeleton
created from a validated factory task. The manifest records relative generated
files, the task digest, and explicit `executes_work=false` /
`schedules_work=false` boundaries. It must not record the local output path.

`ao.atlas.run-link.v0.1` records public-safe evidence paths for a factory task
after Foundry, Forge, or AO2 work has produced artifacts. `run-link attach`
computes a digest over the task id, status, and evidence map. A run link is
readback evidence only; it is not approval, scheduling authority, or execution.

`ao.atlas.workgraph-repair-plan.v0.1` records bounded repair tasks when a
matching run link is blocked or failed. It is advisory readback for Foundry
scheduling and carries explicit no-schedule, no-execute, and no-approval flags.

`ao.atlas.foundry-import.v0.1` records fixture files emitted from dependency-ready
Atlas workgraph nodes. It is a Foundry import packet only: it does not schedule,
execute, approve, mutate sibling repositories, or call providers.

Schemas live in `schemas/`. The CLI validators are the normative v0.1 gate.
