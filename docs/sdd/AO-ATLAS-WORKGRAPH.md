# AO Atlas Workgraph

The workgraph is factory-level. Nodes represent factory tasks, not individual
agent actions. Edges express dependencies, blocked state, and integration or
stitching work.

CLI:

```sh
atlas workgraph validate --workgraph <path>
atlas workgraph binding-digest --workgraph <path> --json
atlas workgraph validate-binding --workgraph <path> --binding <path> --json
atlas workgraph next --workgraph <path> --json
atlas workgraph materialize-next --workgraph <path> --out <dir> --dry-run
atlas workgraph complete --workgraph <path> --run-link <path> --evidence-root <root> --evidence-root-id <id> --out <path>
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
completed evidence-bound run link, matching evidence-root identity, verified
evidence bytes, and completed dependencies. Path-only legacy run links remain
readable but are not eligible for state-changing completion.

## Operational binding gate

An operational Workgraph opts into the gate by carrying
`operational_binding`. Once present, the typed contract is strict: unknown
fields and missing execution, planner, retry, partition, provenance, or safety
bindings fail closed. Every node references one approved planner partition and
repeats the explicit safety boundary for that partition.

Retry arithmetic matches the Mission planner: `maximum_attempts` is between one
and four, and each partition's `retry_allowance_ms` must equal
`estimated_duration_ms * maximum_attempts` without overflow and fit both the
total-node timeout and node budget.

`workgraph validate-binding` compares a separate typed activation-binding
document to the Workgraph. Its deterministic JSON readback sets
`activation_allowed` only for an exact contract, always reports
`child_process_launches: 0`, `executes_work: false`, and
`safe_to_execute: false`, and never runs a node. A denial exits non-zero after
emitting the JSON readback. The graph-binding digest excludes mutable node
status and its own digest field, so an evidence-bound `workgraph complete`
transition preserves the approved operational contract.

Use `workgraph binding-digest` to derive that non-self-referential digest from
the actual Atlas binary before sealing the Workgraph and activation-binding
document. This readback is also inert and reports zero child-process launches.

Legacy AO Mission provenance-node augmentation is denied for an operational
Workgraph because adding readback nodes would expand the approved planner
partitions. Use the digest-bound Mission metadata readback instead.

For operational evidence gates, attach and complete with the same isolated
evidence root:

```sh
atlas run-link attach \
  --task-id <task> \
  --status completed \
  --evidence node=nodes/<node>/node-evidence.json \
  --evidence-root <campaign-root> \
  --evidence-root-id <stable-public-id> \
  --out <run-link>

atlas workgraph complete \
  --workgraph <path> \
  --run-link <run-link> \
  --evidence-root <campaign-root> \
  --evidence-root-id <stable-public-id> \
  --out <path>
```

Evidence-bound run links carry one SHA-256 per evidence entry and bind those
digests and stable evidence-root identity into the run-link digest. Completion
reopens every bounded regular non-symlink evidence file beneath the supplied
root and denies missing, changed, oversized, symlinked, escaping, or
wrong-root evidence before writing the next workgraph. An evidence-bound run
link cannot complete without matching `--evidence-root` and
`--evidence-root-id`. Legacy path-only run links remain readable for existing
evidence but cannot change workgraph state. Strict evidence attachment and
completion require Go 1.24+ for descriptor-backed root anchoring; older
runtimes fail closed.

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

For the authority-ladder workgraph, mission status also includes an
`authority_ladder` readback. It reports the current proven live mutation class,
the next unproven class, blockers, required evidence, denied higher classes,
and do-not-advance gates. This readback is descriptive only; Atlas still does
not grant tickets, schedule work, apply patches, or mark a class safe to
execute.

The public `examples/valid/workgraph-large-stress.json` fixture is the larger
sequencing demonstration. It contains 12 factory nodes across completed, ready,
blocked, and stitch states, preserves bounded context-pack refs, and imports
only dependency-ready nodes into Foundry-compatible material.

The public `examples/valid/workgraph-authority-ladder.json` fixture models the
path from the first docs-only live rehearsal toward complex repository mutation.
It includes repair nodes, context repack nodes, Sentinel/Promoter/Command
evidence nodes, blocked states, dependency gates, and explicit do-not-advance
limits so higher classes remain denied until their evidence exists.
The `workgraph-authority-ladder-multi-repo-dry-run.json` fixture extends the
ladder with a `multi_repo_low_risk` dry-run design. It records ordered PR
dependencies (`ao-atlas` first, `ao-foundry` after `ao-atlas`, `ao-command`
after `ao-foundry`), per-repo rollback scope for Atlas, Foundry, and Command,
and do-not-advance gates; it does not claim low-risk-code live evidence or
multi-repo execution authority.
The `workgraph-complex-repo-mutation-rehearsal.json` fixture models the
dry-run complex rehearsal itself with fourteen complex-class nodes, including
context repack, low-risk decomposition, dependency gates, rollback plan,
rollback graph, repair, Sentinel, Promoter, Command, CI, packet, and
live-denial gates. It remains classification-only and denies complex live
execution until every lower class has live evidence.

`docs_only_multi_file` is the next live rehearsal after the proven
`docs_only_single_file` class. Atlas represents it as classification and
workgraph evidence only: no more than two documentation files, exact write and
rollback scope, CI evidence, Sentinel no-hold readback, Promoter readiness, and
Command readback must exist before downstream tools can request execution. The
code, multi-repo, complex, and fully unsupervised classes remain denied until
their lower-class live evidence is recorded.

After `docs_only_multi_file` live evidence is recorded, Atlas can expose a
`test_only` dry-run chain. That chain is not live execution authority: it only
classifies tests-only write scope, rollback scope, required CI, Sentinel
coverage no-hold readback, Promoter readiness, and Command readback. The
test-only live rehearsal remains blocked until every gate reports ready, and
production code paths remain out of scope.
