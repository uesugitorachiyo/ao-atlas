# AO Atlas Workgraph

The workgraph is factory-level. Nodes represent factory tasks, not individual
agent actions. Edges express dependencies, blocked state, and integration or
stitching work.

CLI:

```sh
atlas workgraph validate --workgraph <path>
atlas workgraph next --workgraph <path> --json
atlas workgraph status --workgraph <path>
```

`workgraph next` returns the first ready node whose dependencies are completed.
Blocked nodes must explain their blockers.

