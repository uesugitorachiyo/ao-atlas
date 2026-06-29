# AO Atlas Foundry Handoff

AO Atlas converts ready workgraph nodes into Foundry-compatible handoff
material. The handoff is a local JSON artifact with task objective, target
factory repo/folder, verification commands, and required evidence.

AO Atlas does not schedule the handoff. Foundry remains the scheduler and safe
next-action selector.

CLI:

```sh
atlas foundry handoff emit --workgraph <path> --out <path>
atlas foundry import --workgraph <path> --out <dir>
```

`foundry import` writes a manifest plus one task fixture per dependency-ready
workgraph node. The output is local fixture material for Foundry import tests or
operator review. Atlas still does not schedule or execute the tasks.

Cross-repo fixture smoke:

```sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

The smoke emits an Atlas Foundry handoff, asks a sibling AO Foundry checkout to
validate the public `atlas-demo` registry fixture, and records the returned
validation evidence as an Atlas run link. It is fixture-only readback: no
scheduling, execution, approval, provider calls, publication, or repository
mutation.
