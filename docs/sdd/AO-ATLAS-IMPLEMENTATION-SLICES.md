# AO Atlas Implementation Slices

## Slice 1

Create repo skeleton, Apache-2.0 license, public-safe docs, contract schemas,
and ignored local state paths.

## Slice 2

Implement typed Go validators and CLI commands for stack instances, intake,
workgraphs, factory tasks, context packs, Foundry handoff, and run links.

## Slice 3

Add valid and invalid fixtures plus tests that prove invalid public-safety and
contract cases are rejected.

## Slice 4

Add `scripts/production-readiness.sh` and make it report `score=100/100` only
when all v0.1 gates pass.

