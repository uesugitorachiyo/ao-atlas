# Blueprint Atlas Foundry Rewire Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce AO Blueprint -> AO Atlas -> AO Foundry for oversized, mutation-class, live-mutation ladder, and long-running work.

**Architecture:** AO Atlas receives the Blueprint pack and authorization, validates scope/digests, and emits ready Atlas/Foundry material or a blocked Blueprint request. AO Foundry requires the Atlas import/readback for oversized and mutation-class gates. AO Command reports Blueprint pack status, Atlas import status, Foundry gate status, and the ready/blocked reason.

**Tech Stack:** Go CLIs, JSON fixtures/contracts, shell smoke scripts, Markdown architecture docs.

## Global Constraints

- Do not execute live code mutation.
- Do not claim `low_risk_code` is live-proven.
- Keep private/local paths under `excluded/` or public-safe placeholders.
- Atlas, Foundry, and Command artifacts must remain read-only and set scheduling, execution, approval, provider-call, and mutation flags false.
- End with touched repos clean, synced to `origin/main`, no `codex/*` branches from this work, tests passing, and PRs merged only after CI.

---

### Task 1: Atlas Blueprint Import Contract

**Files:**
- Modify: `internal/atlas/models.go`
- Modify: `internal/atlas/validate.go`
- Modify: `internal/atlas/cli.go`
- Modify: `internal/atlas/atlas_test.go`
- Create: `schemas/blueprint-import.schema.json`
- Create: `schemas/blueprint-candidate-selection.schema.json`
- Create: `examples/valid/blueprint-import-low-risk-code/blueprint-pack/*`
- Create: `examples/valid/blueprint-import-low-risk-code/build-authorization.json`
- Create: `examples/invalid/blueprint-import-missing-authorization/blueprint-pack/*`
- Modify: `README.md`

**Interfaces:**
- Produces: `BlueprintImport`, `BlueprintCandidateRules`, `BlueprintCandidateSelection`, `BlueprintBuildAuthorization`, `BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error)`, `ValidateBlueprintImport(record BlueprintImport) error`.
- Consumes: existing `BuildFoundryImportForNodes`, `ValidateFoundryImport`, `ValidateContextPack`, `ValidateMutationClassModel`.

- [ ] **Step 1: Write failing Atlas ready import test**

```go
func TestBlueprintImportCompilesLowRiskCodePackIntoAtlasAndFoundryMaterial(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "import")
	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack"),
		"--authorization", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("blueprint import failed: %s", out.String())
	}
	record := mustLoadJSON[BlueprintImport](t, filepath.Join(outDir, "blueprint-import.json"))
	if record.Status != "ready" || record.MutationClass != "low_risk_code" || record.LiveExecutionProven {
		t.Fatalf("unexpected import record: %#v", record)
	}
	if record.Digests["downstream_foundry_import"] == "" || record.Digests["candidate_rules"] == "" {
		t.Fatalf("record missing digest bindings: %#v", record.Digests)
	}
	if _, err := os.Stat(filepath.Join(outDir, "workgraph.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "foundry-import", "foundry-import.json")); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run red test**

Run: `go test ./internal/atlas -run TestBlueprintImportCompilesLowRiskCodePackIntoAtlasAndFoundryMaterial -count=1`
Expected: FAIL because `blueprint import` is unknown.

- [ ] **Step 3: Implement minimal Atlas import**

Add models, validation, digest helpers, fixture-safe pack parsing, candidate selection, context pack/workgraph generation, and CLI wiring. Ready low-risk output must include `do_not_advance:low_risk_code_live_execution_denied` and `safe_to_execute=false` evidence language.

- [ ] **Step 4: Verify ready path**

Run: `go test ./internal/atlas -run TestBlueprintImportCompilesLowRiskCodePackIntoAtlasAndFoundryMaterial -count=1`
Expected: PASS.

- [ ] **Step 5: Write failing blocked authorization tests**

```go
func TestBlueprintImportBlocksWithoutAuthorization(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "blocked")
	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected missing authorization to fail closed")
	}
	request := mustLoadJSON[BlueprintRequest](t, filepath.Join(outDir, "blueprint-request.json"))
	if request.Status != "blueprint_required" || !containsValue(request.Missing, "build_authorization") {
		t.Fatalf("unexpected request: %#v", request)
	}
	if _, err := os.Stat(filepath.Join(outDir, "workgraph.json")); !os.IsNotExist(err) {
		t.Fatalf("blocked import must not write ready workgraph: %v", err)
	}
}
```

- [ ] **Step 6: Run red blocked test**

Run: `go test ./internal/atlas -run TestBlueprintImportBlocksWithoutAuthorization -count=1`
Expected: FAIL until blocked output is implemented.

- [ ] **Step 7: Implement fail-closed blocked path**

Emit `blueprint-import.json` with `status=blocked`, `ready_for_foundry=false`,
and a `blueprint-request.json`; return non-zero from CLI for blocked imports.

- [ ] **Step 8: Verify Atlas full gate**

Run: `go test ./...`
Run: `scripts/production-readiness.sh`
Expected: PASS.

### Task 2: Foundry Requires Atlas Import/Readback

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `scripts/low-risk-code-live-rehearsal-gate.sh`
- Modify: `README.md`

**Interfaces:**
- Consumes: `ao.atlas.blueprint-import.v0.1`, `ao.atlas.foundry-import.v0.1`, `ao.foundry.atlas-status.v0.1`.
- Produces: hardened `PulseIntakePreflight`, mutation-class gate, and low-risk live rehearsal gate decisions.

- [ ] **Step 1: Write failing Foundry preflight test**

```go
func TestPulseIntakePreflightRequiresAtlasBlueprintImportWhenAtlasRequired(t *testing.T) {
	out := filepath.Join(t.TempDir(), "preflight.json")
	code, _, stderr := runCLI("pulse", "intake-preflight",
		"--blueprint-authorization", "examples/pulse-intake/blueprint-authorization.ready.json",
		"--requires-atlas",
		"--atlas-import", "examples/atlas/foundry-import.json",
		"--atlas-status", "examples/contract-fixtures/valid/foundry-atlas-status-v0.1.json",
		"--out", out)
	if code == 0 {
		t.Fatalf("expected missing Atlas Blueprint import to block, stderr=%s", stderr)
	}
}
```

- [ ] **Step 2: Implement `--atlas-blueprint-import`**

Add an optional flag that becomes required when `--requires-atlas` is set. Validate status ready, `ready_for_foundry=true`, `downstream_foundry_import` digest present, no authority claims, and matching import id/workgraph id when available.

- [ ] **Step 3: Harden low-risk live rehearsal gate**

Add `--atlas-blueprint-import` and `--atlas-status` inputs to the shell gate. Missing or non-ready Atlas evidence must set `first_failing_check=atlas_blueprint_import` before live-policy checks.

- [ ] **Step 4: Verify Foundry**

Run: `go test ./...`
Run: `scripts/blueprint-atlas-pulse-e2e-dry-run.sh --out tmp/blueprint-atlas-pulse-e2e-rewire`
Run: `scripts/governed-live-mutation-dry-run-chain.sh --mutation-class low_risk_code --out tmp/low-risk-chain-rewire`
Run: `scripts/low-risk-code-live-rehearsal-gate.sh --chain tmp/low-risk-chain-rewire/summary.json --out tmp/low-risk-chain-rewire/gate.json`
Expected: PASS commands; low-risk gate output remains blocked for live execution.

### Task 3: Command Operator Readback

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `examples/pulse-gate/ready.preflight.json`
- Modify: `README.md`

**Interfaces:**
- Produces: `ao.command.blueprint-atlas-foundry-status.v0.1` through `ao-command blueprint-atlas-foundry status`.
- Consumes: Blueprint authorization/request, Atlas Blueprint import, Foundry Atlas status, and Foundry gate/preflight.

- [ ] **Step 1: Write failing Command readback test**

```go
func TestBlueprintAtlasFoundryStatusReportsReadyAndBlockedReason(t *testing.T) {
	summary := writeBlueprintAtlasFoundryFixtures(t, "ready")
	code, stdout, stderr := runWithFake([]string{"blueprint-atlas-foundry", "status", "--summary", summary}, &fakeRunner{})
	if code != 0 {
		t.Fatalf("status exit=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"blueprint_pack_status=ready",
		"atlas_import_status=ready",
		"foundry_gate_status=ready",
		"ready_reason=Blueprint authorization, Atlas import, and Foundry gate are ready.",
		"mutates_repositories=false",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}
```

- [ ] **Step 2: Implement readback command**

Add `blueprint-atlas-foundry status --summary <json> [--json]` with validation and read-only summary fields.

- [ ] **Step 3: Verify Command**

Run: `go test ./...`
Expected: PASS.

### Task 4: Blueprint and Architecture Wording

**Files:**
- Modify in `ao-blueprint`: `README.md`, `docs/sdd/AO-BLUEPRINT-ARCHITECTURE.md`, `docs/sdd/AO-BLUEPRINT-CONTRACTS.md`
- Modify in `ao-architecture`: `README.md`, `scripts/verify_architecture.py`, relevant mirrored docs if present

**Interfaces:**
- Produces documentation stating Atlas is mandatory between Blueprint and Foundry for oversized, mutation-class, live-mutation, and long-running work.

- [ ] **Step 1: Update docs with exact boundary wording**

Use: "Blueprint does not hand directly to Foundry for oversized, mutation-class, live-mutation, or long-running work. Atlas is the mandatory compiler between Blueprint and Foundry for those classes."

- [ ] **Step 2: Add Architecture verifier phrase checks**

Require the mandatory compiler wording and `Blueprint -> Atlas -> Foundry`.

- [ ] **Step 3: Verify docs repos**

Run in `ao-blueprint`: `go test ./...`
Run in `ao-architecture`: `python3 scripts/verify_architecture.py`
Expected: PASS.

### Task 5: PRs, CI, Merge, and Cleanup

**Files:**
- All touched repos.

**Interfaces:**
- Produces merged PRs or explicit blockers if CI/merge cannot complete.

- [ ] **Step 1: Run final local verification**

Run the commands from Tasks 1-4 plus `git diff --check` in each touched repo.

- [ ] **Step 2: Commit per repo**

Use concise messages:

```bash
git commit -m "feat: add Atlas Blueprint import path"
git commit -m "feat: require Atlas import before live gates"
git commit -m "feat: add Blueprint Atlas Foundry readback"
git commit -m "docs: route oversized Blueprint work through Atlas"
git commit -m "docs: clarify Atlas compiler boundary"
```

- [ ] **Step 3: Push and open PRs**

Push `blueprint-atlas-foundry-rewire` in each touched repo. Open PRs against `main`.

- [ ] **Step 4: Wait for CI and merge only green PRs**

Use GitHub checks. Merge after required CI passes.

- [ ] **Step 5: Sync and cleanup**

Switch each repo to `main`, pull `origin/main`, delete local and remote feature branches. Remove the related low-risk stashes only if the new merged work supersedes them; otherwise leave them and report that they remain.
