# AO Atlas Contracts

v0.1 contract names:

- `ao.atlas.stack-instance.v0.1`
- `ao.atlas.intake.v0.1`
- `ao.atlas.blueprint-request.v0.1`
- `ao.atlas.workgraph.v0.1`
- `ao.atlas.factory-task.v0.1`
- `ao.atlas.context-pack.v0.1`
- `ao.atlas.foundry-handoff.v0.1`
- `ao.atlas.run-link.v0.1`

All contracts are JSON. Validation requires explicit `contract_version` values,
stable identifiers, non-empty required fields, and public-safe paths.

`ao.atlas.blueprint-request.v0.1` is emitted when intake is not specific
enough to compile into a workgraph. It records the intake id, missing fields,
and reason for routing back to AO Blueprint. It is not build authorization and
does not allow Atlas, Foundry, or AO2 to schedule or execute work.

Schemas live in `schemas/`. The CLI validators are the normative v0.1 gate.
