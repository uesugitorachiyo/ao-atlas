# Blueprint Atlas Foundry Rewire Design

## Objective

Make AO Blueprint -> AO Atlas -> AO Foundry the enforced path for oversized,
live-mutation ladder, mutation-class, and long-running work. Blueprint remains
the requirements and build-authorization front door. Atlas becomes the mandatory
compiler that turns a ready Blueprint pack into digest-bound intake, workgraph,
context packs, candidate-selection readback, and Foundry import material.
Foundry consumes Atlas import/readback before any live-mutation or oversized
gate continues.

This slice includes the current `low_risk_code` rehearsal shape only as
public-safe dry-run/readback material. It must not execute live mutation and must
not claim `low_risk_code` is live-proven.

## Architecture

AO Atlas adds `ao.atlas.blueprint-import.v0.1` and a command:

```bash
atlas blueprint import \
  --pack <blueprint-pack-dir> \
  --authorization <build-authorization.json> \
  --instance <stack-instance.json> \
  --mutation-classes <mutation-classes.json> \
  --out <dir>
```

Ready output:

- `blueprint-import.json`
- `intake.json`
- `candidate-selection.json`
- `context-packs/*.json`
- `workgraph.json`
- `foundry-import/foundry-import.json`
- `foundry-import/tasks/*.json`

Blocked output:

- `blueprint-import.json`
- `blueprint-request.json`

Atlas emits blocked output, not a ready workgraph, when Blueprint authorization
is missing, blocked, stale, digest-mismatched, unsafe, or not scoped to the
requested work.

## Digest Binding

The Atlas import record binds all inputs and derived downstream material:

- Blueprint pack digest
- build authorization digest and declared pack digest
- implementation spec digest
- quality profile digest
- candidate rules digest
- mutation class model digest
- candidate-selection digest
- context pack digests
- workgraph digest
- downstream Foundry import digest

The record also carries `schedules_work=false`, `executes_work=false`,
`approves_work=false`, and `mutates_repositories=false`.

## Atlas Compilation Rules

The initial compiler supports the public-safe Blueprint pack shape already used
by AO Blueprint:

- `implementation-spec.md`
- `quality-profile.md`
- `sdd-plan.json`
- `ao-foundry-task.json`
- optional `candidate-rules.json`

`candidate-rules.json` supplies mutation class, target repo, write scope,
rollback scope, required gates, required evidence, and verification commands.
When it is absent, Atlas derives a conservative docs-only candidate from
`ao-foundry-task.json`; mutation-class and live ladder work must provide explicit
rules.

For `low_risk_code`, Atlas requires the candidate to remain dry-run/readback
only and to carry safety limits that deny live execution. The resulting Foundry
import is ready for Foundry validation/readback, not live execution.

## Foundry Enforcement

Foundry already validates `ao.atlas.foundry-import.v0.1` and emits
`ao.foundry.atlas-status.v0.1`. This slice hardens the gates so oversized,
live-mutation, and mutation-class paths must include Atlas import/readback before
Foundry gates advance:

- Pulse intake preflight requires an Atlas Blueprint import record when
  `--requires-atlas` is set.
- Mutation class evaluation requires a ready Atlas import/readback for
  `low_risk_code`, `multi_repo_low_risk`, and `complex_repo_mutation`.
- The low-risk live rehearsal gate blocks before downstream live checks when
  Atlas Blueprint import/readback is missing or not ready.

Foundry remains readback/gating only. It does not create branches, mutate repos,
execute live code changes, or grant mutation authority.

## Command Readback

AO Command adds an operator-facing readback that composes:

- Blueprint pack/import status
- Atlas import status
- Foundry gate status
- ready or blocked reason

The readback is read-only and preserves `operator_mode=read_only` and
`mutates_repositories=false`.

## Documentation

AO Blueprint and AO Architecture wording must say that Blueprint does not hand
directly to Foundry for oversized, mutation-class, live-mutation, or long-running
work. Atlas is the mandatory compiler between Blueprint and Foundry for those
classes. Narrow non-oversized work may still use the existing downstream
contracts where documented, but not for the governed live-mutation ladder.

## Verification

Required checks:

- `go test ./...` in `ao-atlas`
- `scripts/production-readiness.sh` in `ao-atlas`
- `go test ./...` in `ao-foundry`
- targeted Foundry scripts for Blueprint/Atlas Pulse and low-risk gate
- `go test ./...` in `ao-command`
- `go test ./...` in `ao-blueprint`
- `python3 scripts/verify_architecture.py` in `ao-architecture`

Completion must leave touched repos clean and synced to `origin/main`, with no
`codex/*` branches left from this work. PRs are merged only after CI passes.
