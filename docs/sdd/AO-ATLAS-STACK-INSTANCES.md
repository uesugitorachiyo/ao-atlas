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
atlas instance doctor --instance <path> [--registry <path>] [--json] [--out <path>]
```

`instance doctor` validates stack-instance roots, generated Atlas registry
parity, ignored local state placement, bounded worktree root hygiene,
shared-toolchain use, and authority boundaries. It emits
`ao.atlas.instance-doctor.v0.1` with `status`, `checks`,
`first_failing_check`, `blocking_next_actions`, and
`maintenance_suggestions`. It is readback only and does not schedule, execute,
approve, publish, call providers, or mutate sibling repositories.

Tracked examples use relative placeholder paths. Real generated instance state
belongs outside the public repo or under ignored local directories.
