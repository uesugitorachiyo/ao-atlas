# AO Atlas Foundry Handoff

AO Atlas converts ready workgraph nodes into Foundry-compatible handoff
material. The handoff is a local JSON artifact with task objective, target
factory repo/folder, verification commands, and required evidence.

AO Atlas does not schedule the handoff. Foundry remains the scheduler and safe
next-action selector.

CLI:

```sh
atlas foundry handoff emit --workgraph <path> --out <path>
```

