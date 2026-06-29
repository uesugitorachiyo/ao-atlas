# AO Atlas SDD Handoff

## Objective

Build and maintain AO Atlas as the stack-instance and context-pack compiler for
the AO stack.

## v0.1 Command Surface

- `atlas instance init`
- `atlas instance validate`
- `atlas instance registry emit`
- `atlas instance inspect`
- `atlas intake validate`
- `atlas workgraph validate`
- `atlas workgraph next`
- `atlas workgraph status`
- `atlas factory-task validate`
- `atlas context-pack validate`
- `atlas foundry handoff emit`
- `atlas foundry import`
- `atlas run-link validate`

## Closure Rule

Do not close v0.1 unless `scripts/production-readiness.sh` reports
`score=100/100`.
