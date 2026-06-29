# AO Atlas PRD

## Purpose

AO Atlas turns oversized AO stack objectives into bounded factory-level work.
It prevents context-window overflow by compiling stack-instance manifests,
workgraphs, factory tasks, and context packs instead of handing whole missions
or duplicated repo folders to a factory.

## Product Thesis

Multiple AO stacks should be lightweight logical instances over one shared AO
toolchain. AO Atlas models those instances and prepares evidence-bound handoff
material. It does not copy source repos and does not become a scheduler.

## Users

- Operators who need to split large objectives into governed factory runs.
- Foundry maintainers who need clean registry material and factory tasks.
- AO2 operators who need bounded context rather than entire mission history.

## Non-Goals

- No live provider calls.
- No push, tag, release, upload, or execution.
- No source-tree duplication.
- No replacement for Blueprint, Foundry, Forge, AO2, Command, or policy gates.

## v0.1 Readiness

v0.1 is ready only when the production-readiness gate reports `score=100/100`
and all tracked fixtures remain public-safe.

Stable tag or release-candidate eligibility additionally requires the
fixture-only Atlas -> Foundry roundtrip smoke to pass against a sibling AO
Foundry checkout. That smoke proves Foundry can validate the emitted
`ao.atlas.foundry-import.v0.1` packet without granting scheduling, execution,
approval, provider, publication, or sibling-repository mutation authority.
