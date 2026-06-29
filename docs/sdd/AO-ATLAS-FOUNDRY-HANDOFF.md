# AO Atlas Foundry Handoff

AO Atlas converts ready workgraph nodes into Foundry-compatible handoff
material. The handoff is a local JSON artifact with task objective, target
factory repo/folder, verification commands, and required evidence.

AO Atlas does not schedule the handoff. Foundry remains the scheduler and safe
next-action selector.

CLI:

```sh
atlas foundry handoff emit --workgraph <path> --out <path>
atlas foundry import --workgraph <path> --instance <path> --out <dir> [--node <id>] [--json]
```

`foundry import` writes a manifest plus one task fixture per dependency-ready
workgraph node, or one selected ready node when `--node` is provided. The
manifest records the source workgraph and stack-instance paths with sha256
digests, preserves context-pack refs, and keeps `schedules_work=false`,
`executes_work=false`, and `approves_work=false`. The output is local fixture
material for Foundry import tests or operator review. Atlas still does not
schedule, execute, approve, publish, call providers, or mutate sibling repos.

When the source workgraph represents the first tiny docs-only live mutation
class, the import material remains non-authoritative. It may carry task
objective, docs-only write scope, acceptance criteria, context-pack refs,
dependency refs, and required evidence into Foundry, but Covenant approval,
Foundry approval gate, Forge guard, AO2 patch packet, Sentinel verdict,
Promoter boundary, rollback rehearsal, Command readback, and exact operator
approval decide whether a later docs-only PR rehearsal can proceed.

Cross-repo fixture smoke:

```sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

The smoke emits an Atlas Foundry handoff and a Foundry import packet, asks a
sibling AO Foundry checkout to validate the public `atlas-demo` registry fixture
and the `ao.atlas.foundry-import.v0.1` packet, records the returned validation
evidence as an Atlas run link, and asks Foundry to emit
`ao.foundry.atlas-readback.v0.1`. It is fixture-only readback: no scheduling,
execution, approval, provider calls, publication, or repository mutation.
