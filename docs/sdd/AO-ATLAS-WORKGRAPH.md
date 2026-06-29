# AO Atlas Workgraph

The workgraph is factory-level. Nodes represent factory tasks, not individual
agent actions. Edges express dependencies, blocked state, and integration or
stitching work.

CLI:

```sh
atlas workgraph validate --workgraph <path>
atlas workgraph next --workgraph <path> --json
atlas workgraph materialize-next --workgraph <path> --out <dir> --dry-run
atlas workgraph complete --workgraph <path> --run-link <path> --out <path>
atlas workgraph repair-plan --workgraph <path> --run-link <path> --out <path>
atlas workgraph status --workgraph <path>
```

`workgraph next` returns the first ready node whose dependencies are completed.
Blocked nodes must explain their blockers.

`workgraph materialize-next --dry-run` selects that same next ready node and
writes a bounded factory skeleton through the factory materialization path. It
does not schedule, execute, approve, publish, upload, or call providers.

`workgraph complete` is explicit file-to-file completion. It reads an existing
workgraph and run link, marks only the matching factory-task node completed in a
new output workgraph, and refuses to overwrite the input. Completion requires a
completed run link, public-safe evidence, and completed dependencies.

`workgraph repair-plan` emits a bounded repair task when a matching run link is
blocked or failed. It writes a repair-plan artifact only; Atlas still does not
schedule, execute, approve, publish, upload, or call providers. The repair task
preserves the source task's write scope, verification commands, required
evidence, dependency refs, and context-pack refs so Foundry can schedule a
bounded follow-up without re-expanding the whole mission. The public
`examples/valid/workgraph-repair-plan-blocked-node-demo.json` fixture is the
blocked-node demo for this path.

Mission status readback:

```sh
atlas mission status --intake <path> --workgraph <path> [--run-link <path>...] [--json] [--out <path>]
```

Mission status summarizes intake, workgraph node counts including failed count,
missing context packs, missing Foundry handoffs, run-link completion state, and
the next recommended Atlas action without mutating source artifacts.
