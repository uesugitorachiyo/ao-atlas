# Atlas Long-Run Lease Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make AO Atlas recommendation readbacks enforce the 120-minute long-run lease instead of treating `min_minutes` as metadata.

**Architecture:** Add timing fields to the recommendation readback contract, compute or accept elapsed minutes through builder/CLI options, and include lease timing in the final response gate. Production readiness then rejects committed evidence that claims final completion before the lease minimum is met.

**Tech Stack:** Go, shell, jq, existing AO Atlas CLI and evidence fixtures.

## Global Constraints

- No direct main mutation; work on `codex/atlas-lease-time-gate`.
- No provider calls, credential inspection, release/deploy/publish/upload/tag, dependency updates, auth/policy/config widening, hidden instruction mutation, or broad RSI claim.
- Preserve exactly one executable mutation node semantics.
- Use TDD: regression tests fail before production code changes.

---

### Task 1: Failing Lease-Gate Tests

**Files:**
- Modify: `internal/atlas/mission_recommendations_test.go`

**Interfaces:**
- Consumes: `BuildAtlasRecommendationReadback(wave, workgraph, options)`
- Produces: tests for `StartedAt`, `CompletedAt`, `ElapsedMinutes`, `MinMinutesMet`, and `LeaseTimeStatus`.

- [ ] **Step 1: Write failing tests**

Add tests that build a 40-node completed recommendation workgraph and expect:

```go
BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
    StartedAt: "2026-07-04T07:20:20-07:00",
    CompletedAt: "2026-07-04T07:42:06-07:00",
})
```

to return `FinalResponseAllowed == false`, `MinMinutesMet == false`, `LeaseTimeStatus == "minimum_minutes_unmet"`, and an exact next action to generate the next useful wave.

- [ ] **Step 2: Run the focused test**

Run: `go test ./internal/atlas -run 'TestMissionRecommendationsReadbackFinalGateTransitions|TestMissionRecommendationsDenyFinalResponseWhenLeaseMinutesUnmet|TestMissionRecommendationsDenyFinalResponseWhenLeaseTimingMissing' -count=1`

Expected: FAIL because timing fields/options do not exist yet.

### Task 2: Readback Timing Contract

**Files:**
- Modify: `internal/atlas/models.go`
- Modify: `internal/atlas/mission_recommendations.go`

**Interfaces:**
- Consumes: `AtlasRecommendationReadbackOptions`
- Produces: readback JSON fields `started_at`, `completed_at`, `elapsed_minutes`, `min_minutes_met`, and `lease_time_status`.

- [ ] **Step 1: Add timing fields**

Add string fields for `StartedAt` and `CompletedAt`, integer `ElapsedMinutes`, boolean `MinMinutesMet`, and string `LeaseTimeStatus`.

- [ ] **Step 2: Implement lease timing**

Parse RFC3339 timestamps when provided, accept explicit elapsed minutes when useful for tests/fixtures, treat missing timing as unmet for completed long-run waves, and require elapsed minutes >= supervisor min minutes before final response can close.

- [ ] **Step 3: Run the focused test**

Run: `go test ./internal/atlas -run 'TestMissionRecommendationsReadbackFinalGateTransitions|TestMissionRecommendationsDenyFinalResponseWhenLeaseMinutesUnmet|TestMissionRecommendationsDenyFinalResponseWhenLeaseTimingMissing' -count=1`

Expected: PASS.

### Task 3: CLI and Production Readiness Binding

**Files:**
- Modify: `internal/atlas/cli.go`
- Modify: `scripts/production-readiness.sh`
- Modify: `internal/atlas/mission_recommendations_test.go`

**Interfaces:**
- Consumes: `--started-at`, `--completed-at`, `--elapsed-minutes` on recommendation `readback` and `complete-node`.
- Produces: production readiness rejection for completed evidence where `elapsed_minutes < supervisor.min_minutes`.

- [ ] **Step 1: Add CLI timing flags**

Add timing flags to `mission recommendations readback` and `mission recommendations complete-node`, pass them into `AtlasRecommendationReadbackOptions`, and print lease timing in CLI output.

- [ ] **Step 2: Update readiness checks**

Make `recommendation-ledger-consistency` reject final-allowed recommendation evidence unless `min_minutes_met == true` and `elapsed_minutes >= supervisor.min_minutes`.

- [ ] **Step 3: Run tests**

Run: `go test ./... -count=1`

Expected: PASS.

### Task 4: Evidence and Docs Correction

**Files:**
- Modify: `docs/evidence/ao-atlas-long-recommendation-wave-v04/recommendation-readback.json`
- Modify: `docs/evidence/ao-atlas-long-recommendation-wave-v04/execution-readback.json`
- Modify: `docs/evidence/ao-atlas-long-recommendation-wave-v04/command-readback.json`
- Create: `docs/evidence/ao-atlas-lease-time-gate-v01/repair-readback.json`
- Modify documentation under `docs/` if an existing long-run runbook mentions node counts without timing.

**Interfaces:**
- Consumes: v04 evidence root.
- Produces: corrected evidence that denies final response for the ~22 minute run.

- [ ] **Step 1: Correct v04 final gate**

Set v04 final readback to `final_response_allowed=false`, `status="in_progress"`, `elapsed_minutes=22`, `min_minutes_met=false`, and `lease_time_status="minimum_minutes_unmet"`.

- [ ] **Step 2: Record repair evidence**

Write a compact repair readback that explains the root cause, tests, verification commands, no-promotion status, and exact next action.

### Task 5: Full Verification and PR

**Files:**
- All changed files.

**Interfaces:**
- Consumes: branch `codex/atlas-lease-time-gate`.
- Produces: merged PR if GitHub lifecycle is available.

- [ ] **Step 1: Run verification**

Run:

```bash
go test ./... -count=1
go vet ./...
go build ./cmd/atlas
scripts/production-readiness.sh
scripts/atlas-foundry-roundtrip-smoke.sh
```

Expected: all pass.

- [ ] **Step 2: Commit, PR, CI, merge, cleanup**

Use normal GitHub remote lifecycle, then sync `main` and delete local/remote `codex/*` branches.
