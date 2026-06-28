# AO Atlas Architecture

AO Atlas is a local-first Go CLI. It reads and writes JSON contracts, validates
public-safety boundaries, and emits Foundry handoff material for ready
factory-level workgraph nodes.

## Stack Role

```text
AO Blueprint -> AO Atlas -> AO Foundry -> AO Forge -> AO2
```

- Blueprint clarifies requirements and authorizes builds.
- Atlas compiles oversized objectives into bounded factory context.
- Foundry schedules and selects safe next actions.
- Forge owns a governed run.
- AO2 executes governed work.

## Boundaries

AO Atlas writes local JSON artifacts only. It never schedules work, approves
claims, mutates other AO repos, executes providers, or publishes releases.

