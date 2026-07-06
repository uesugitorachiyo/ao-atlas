# AO Operator Routing Guide

Mission is the operator-facing loop. Atlas owns workgraph state and next-node continuation. Blueprint only when new authorization or governed planning is required. Foundry for ready bounded implementation.

## AO Mission

Use AO Mission as the operator-facing loop for starting, resuming, and closing supervised hardening waves. Its boundary is orchestration readback; it must not claim authority promotion from partial progress.

## AO Atlas

Use AO Atlas when workgraph state, next-node continuation, durable checkpoints, or multi-repo sequencing must be preserved. Its boundary is coordination and evidence readback for exactly one executable node at a time.

## AO Blueprint

Use AO Blueprint for new authorization, new requirements, or a new governed plan. Its boundary is planning and authorization evidence; ready bounded implementation does not need Blueprint replanning.

## AO Foundry

Use AO Foundry for ready bounded implementation after Atlas selects one executable node. Its boundary is implementation, local verification, PR, CI, merge, and cleanup evidence for that node.

## AO Promoter

Use AO Promoter when promotion or no-promotion evidence must be evaluated. Its boundary is verdict/readiness evidence; no capability class changes without explicit supporting evidence.

## AO Command

Use AO Command for readback agreement, compact timeline state, exact next action, and final-response denial or allowance. Its boundary is command-state reporting, not hidden authorization changes.

## AO Sentinel

Use AO Sentinel for public-safety wording scans, stale risk wording, and forbidden claim detection. Its boundary is risk/readback evidence; it does not execute implementation work.

## AO Architecture

Use AO Architecture for capability map, system boundary, and cross-component architecture evidence. Its boundary is architecture documentation and verification evidence, not deployment or release action.
