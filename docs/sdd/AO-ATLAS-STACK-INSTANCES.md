# AO Atlas Stack Instances

A stack instance is a lightweight logical AO workspace over a shared toolchain.
It contains paths for mission, workgraph, context, evidence, and worktree state,
but it does not copy AO source repos.

CLI:

```sh
atlas instance init --id <id> --state-root <path> --toolchain-root <path> --out <path>
atlas instance validate --instance <path>
atlas instance registry emit --instance <path> --out <foundry-registry.json>
atlas instance inspect --instance <path> --json
```

Tracked examples use relative placeholder paths. Real generated instance state
belongs outside the public repo or under ignored local directories.

