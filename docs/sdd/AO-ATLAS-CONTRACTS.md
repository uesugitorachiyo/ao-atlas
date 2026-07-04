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
- `ao.atlas.foundry-continuation-handoff.v0.1`
- `ao.atlas.run-link.v0.1`
- `ao.atlas.mutation-classes.v0.1`
- `ao.atlas.ao-mission-import.v0.1`
- `ao.atlas.ao-mission-workgraph-metadata.v0.1`
- `ao.atlas.recommendation-wave.v0.1`
- `ao.atlas.recommendation-readback.v0.1`

All contracts are JSON. Validation requires explicit `contract_version` values,
stable identifiers, non-empty required fields, and public-safe paths.

`ao.atlas.instance-doctor.v0.1` records stack-instance hygiene readback:
instance validation, generated registry parity, ignored local state placement,
bounded worktree roots, shared-toolchain use, and authority boundaries. The
doctor reports `ready`, `blocked`, or `failed`, including the first failing
check and blocking next actions. It does not schedule, execute, approve,
publish, call providers, or mutate sibling repositories.

`ao.atlas.mission-status.v0.1` summarizes intake, workgraph node state, run-link
state, missing context packs, missing Foundry handoffs, completion status, and
next readback actions. It is status readback only and does not schedule or
execute work.

`ao.atlas.ao-mission-import.v0.1` binds AO Mission record, AO Command
mission-status, and AO Mission artifact-manifest readbacks before Atlas compiles
mission context into workgraphs. It carries source artifact digests and rejects
any scheduling, execution, approval, or repository-mutation authority claim. If
the artifact manifest includes artifact refs, Atlas verifies each referenced
file against its declared `sha256:` digest and blocks the import on mismatch.

`ao.atlas.ao-mission-workgraph-metadata.v0.1` binds an AO Mission import to a
validated Atlas workgraph. It records the mission id, workgraph id, target
instance, node counts, and source digests without changing the workgraph schema
or granting execution, scheduling, or approval authority.

`ao.atlas.recommendation-wave.v0.1` binds AO Mission Feature Depth
Recommendations into Atlas-owned long-run planning. The wave requires a
digest-bound source recommendation artifact, explicit minimum task count,
estimated minute budget, Atlas-owned tasks with gates, verification commands,
and safety limits, and a v0.2 long-run supervisor lease. The default lease is a
30-node minimum, 120-180 minute target, and 40-node continue-if-fast target. The
wave records `final_response_allowed=false` while ready nodes or exact next
actions remain and requires Promoter, Command, and public-safety readbacks before
closure. It is planning/readback material only: it does not schedule, execute,
approve, mutate repositories, call providers, inspect credentials, or claim broad
RSI.

`ao.atlas.recommendation-readback.v0.1` reconciles a recommendation wave with its
generated workgraph. It records node counts, executable-ready node count, lease
timing (`started_at`, `completed_at`, `elapsed_minutes`, `min_minutes_met`, and
`lease_time_status`), lease health, checkpoint freshness, stale-route decision
status, early-return risk, exact next action, per-node gate/readback evidence,
and the first 10 Feature Depth recommendations. Final response is allowed only
when no ready, blocked, or failed recommendation nodes remain and the elapsed
lease time meets `supervisor.min_minutes`. The readback is evidence only: it
does not schedule, execute, approve, mutate repositories, call providers,
inspect credentials, or claim broad RSI.

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

`ao.atlas.mutation-classes.v0.1` records the authority ladder classification
model for live mutation classes. Each class defines allowed paths, forbidden
paths, maximum file count, required gates, rollback requirements, CI
requirements, and promotion requirements. The model is classification readback
only: it carries `schedules_work=false`, `executes_work=false`, and
`approves_work=false`, and it does not grant or consume authority tickets.

`ao.atlas.foundry-import.v0.1` records fixture files emitted from
dependency-ready Atlas workgraph nodes. Each imported task carries
`mutation_class`, `write_scope`, `rollback_scope`, `required_gates`,
`required_evidence`, and `authority_boundary` so Foundry can fail closed before
requesting any downstream authority. It is a Foundry import packet only: it
does not schedule, execute, approve, mutate sibling repositories, or call
providers.

`ao.atlas.foundry-continuation-handoff.v0.1` records the operator-ready
handoff that accompanies a Foundry import. It includes the AO Foundry target
folder, `codex --yolo` command, full paste-ready prompt, source artifact paths,
first safe node, node counts, class boundary, stop conditions, and hard safety
prohibitions. It preserves the Atlas no-execution boundary and replaces
inspection-only next actions with an executable continuation instruction for
the operator.

Schemas live in `schemas/`. The CLI validators are the normative v0.1 gate.
