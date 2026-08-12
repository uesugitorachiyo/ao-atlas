# Atlas Governed Release Finalize Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add and qualify a protected AO Atlas finalizer that publishes only exact artifacts imported from a successful, digest-bound release rehearsal.

**Architecture:** Keep `.github/workflows/release-rehearsal.yml` publication-disabled and bind committed release notes into its immutable plan. Add one separate `.github/workflows/release-finalize.yml` that authenticates and re-verifies the producer run, gates the sole write job behind `protected-release`, publishes the exact imported inventory, and independently verifies the public result.

**Tech Stack:** GitHub Actions YAML, Bash, `gh`, `jq`, Go fixture/contract tests, existing `scripts/verify-release-rehearsal-candidates.sh`.

## Global Constraints

- Publish only Linux x86_64, macOS aarch64, and Windows x86_64 archives; do not claim Linux aarch64.
- Keep workflow-level permissions at `actions: read` and `contents: read`; only the protected live job receives `contents: write`.
- Use `protected-release`; workflow code must not create or modify environments, reviewers, credentials, or permissions.
- Keep rehearsal dry-run-only and use its exact successful run as the sole candidate source.
- Reject stale, duplicate, symlinked, non-regular, oversized, digest-drifted, or semantically contradictory imports before mutation.
- Never delete, force-move, or overwrite a tag or release.
- Add no dependency or reusable release framework.
- Use committed `docs/release/v0.2.0.md` as the release body and bind its digest into the rehearsal plan.

---

### Task 1: Bind Committed Release Notes Into Rehearsal

**Files:**
- Create: `docs/release/v0.2.0.md`
- Modify: `.github/workflows/release-rehearsal.yml`
- Modify: `scripts/verify-release-rehearsal-candidates.sh`
- Modify: `internal/atlas/release_rehearsal_workflow_test.go`

**Interfaces:**
- Consumes: workflow input `version=v0.2.0`; verifier arguments already parsed by `scripts/verify-release-rehearsal-candidates.sh`.
- Produces: verifier argument `--release-notes-sha256 sha256:<64 lowercase hex>` and plan field `release_notes_sha256` in schema `ao.atlas.release-rehearsal-promotion-plan.v0.5`.

- [ ] **Step 1: Write failing release-note binding tests**

Add assertions to `TestSpecialistReleaseRehearsalWorkflowStructure` and `TestReleaseRehearsalCandidateVerifierAcceptsExactInventory` requiring these literal contracts:

```go
if !strings.Contains(workflow, `release_notes="docs/release/${VERSION}.md"`) ||
    !strings.Contains(workflow, `--release-notes-sha256 "sha256:$release_notes_sha256"`) {
    t.Fatal("rehearsal must bind committed release notes")
}
if plan["schema_version"] != "ao.atlas.release-rehearsal-promotion-plan.v0.5" ||
    plan["release_notes_sha256"] != "sha256:"+strings.Repeat("d", 64) {
    t.Fatalf("promotion plan release-note binding drifted: %#v", plan)
}
```

Extend `runReleaseCandidateVerifier` to pass:

```go
"--release-notes-sha256", "sha256:" + strings.Repeat("d", 64),
```

Add a negative fixture invoking the verifier with `sha256:bad` and expecting `invalid release notes SHA-256`.

- [ ] **Step 2: Run focused tests and observe RED**

Run: `go test ./internal/atlas -run 'TestSpecialistReleaseRehearsalWorkflowStructure|TestReleaseRehearsalCandidateVerifier' -count=1`

Expected: FAIL because the workflow has no release-note binding and the plan remains v0.4.

- [ ] **Step 3: Implement minimal rehearsal binding**

Create `docs/release/v0.2.0.md` with title `AO Atlas v0.2.0`, the three supported targets, provider-free installation smoke, checksum/SBOM/provenance availability, no-signing-key signature status, and explicit no-Linux-aarch64 wording.

In `validate-release-inputs`, require `VERSION == v0.2.0`, require `TAG == VERSION`, require `docs/release/$VERSION.md`, hash it, and add `release_notes_sha256` plus `release_notes_path` to the input binding.

In `assemble-promotion-plan`, recompute the committed blob digest using `git cat-file blob "${SOURCE_SHA}:docs/release/${VERSION}.md" | sha256sum | awk '{print $1}'` and pass `--release-notes-sha256 "sha256:$release_notes_sha256"`.

In `scripts/verify-release-rehearsal-candidates.sh`, parse the new flag, validate it with `^sha256:[0-9a-f]{64}$`, bump the plan schema to v0.5, and emit the exact digest as `release_notes_sha256`.

- [ ] **Step 4: Run focused tests and observe GREEN**

Run: `go test ./internal/atlas -run 'TestSpecialistReleaseRehearsalWorkflowStructure|TestReleaseRehearsalCandidateVerifier' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit rehearsal binding**

```bash
git add docs/release/v0.2.0.md .github/workflows/release-rehearsal.yml scripts/verify-release-rehearsal-candidates.sh internal/atlas/release_rehearsal_workflow_test.go
git commit -m "feat: bind Atlas release notes into rehearsal"
```

### Task 2: Add Static Fail-Closed Finalizer Contract

**Files:**
- Create: `internal/atlas/release_finalize_workflow_test.go`
- Create: `.github/workflows/release-finalize.yml`

**Interfaces:**
- Consumes: `producer_run_id`, `expected_source_sha`, `expected_version`, `expected_tag`, `expected_manifest_digest`, `expected_plan_digest`, `dry_run`, `live_confirmation`.
- Produces: validated artifact `ao-atlas-validated-release-import-<producer_run_id>`, dry-run boundary artifact, and protected `publish-release` job.

- [ ] **Step 1: Write the failing workflow-structure test**

Create `TestAtlasReleaseFinalizeWorkflowStructure` that reads `.github/workflows/release-finalize.yml` and requires:

```go
required := []string{
    "workflow_dispatch:", "producer_run_id:", "expected_source_sha:",
    "expected_version:", "expected_tag:", "expected_manifest_digest:",
    "expected_plan_digest:", "dry_run:", "live_confirmation:",
    "actions: read", "contents: read", "validate-imported-release:",
    "dry-run-boundary:", "publish-release:", "verify-public-release:",
    "environment: protected-release", "contents: write",
    "inputs.dry_run == false", "actions/runs/$PRODUCER_RUN_ID",
    ".github/workflows/release-rehearsal.yml", "conclusion == \"success\"",
    "scripts/verify-release-rehearsal-candidates.sh", "gh release create",
    "gh release download", "git/ref/tags/$TAG", "provider_credentials_used:false",
}
```

Reject workflow text if global `write-all`, `pull-requests: write`, environment API calls, `gh release delete`, `git push --force`, or publication appears outside `publish-release`.

Add table mutations deleting `environment: protected-release`, changing top-level `contents: read` to write, removing `inputs.dry_run == false`, replacing the producer workflow path, and removing `expected_plan_digest`; each must return a specific error.

- [ ] **Step 2: Run the structure test and observe RED**

Run: `go test ./internal/atlas -run TestAtlasReleaseFinalizeWorkflowStructure -count=1`

Expected: FAIL because `.github/workflows/release-finalize.yml` does not exist.

- [ ] **Step 3: Implement validation and dry-run jobs**

Create `release-finalize.yml` with `workflow_dispatch` only and top-level:

```yaml
permissions:
  actions: read
  contents: read
```

`validate-imported-release` must:

1. validate all input regexes and exact `expected_tag == expected_version`;
2. fetch run and artifact metadata with `gh api` and require repository, `.github/workflows/release-rehearsal.yml`, `workflow_dispatch`, `success`, exact head SHA, age at most 14 days, unexpired artifacts, and one exact plan plus one candidate per supported target;
3. download by exact run ID using `actions/download-artifact@v7`;
4. reject symlinks/non-regular files, more than 128 files, or more than 128 MiB;
5. run every candidate `SHA256SUMS`, rerun `scripts/verify-release-rehearsal-candidates.sh`, verify imported and recomputed `promotion-plan.json` byte identity, and verify `expected_plan_digest`, manifest digest, release-note digest, source, tag, and version;
6. stage three archives, aggregate `SHA256SUMS`, promotion plan/checksum, and target-qualified `provenance`, `sbom`, and `signature-verification` assets;
7. upload exactly one validated-import artifact.

`dry-run-boundary` runs only for `inputs.dry_run == true` and emits JSON with every mutation flag false. It has read-only permissions.

- [ ] **Step 4: Run the structure test and observe GREEN**

Run: `go test ./internal/atlas -run TestAtlasReleaseFinalizeWorkflowStructure -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the imported-release boundary**

```bash
git add internal/atlas/release_finalize_workflow_test.go .github/workflows/release-finalize.yml
git commit -m "feat: validate imported Atlas release candidates"
```

### Task 3: Add Protected Publication And Public Verification

**Files:**
- Modify: `internal/atlas/release_finalize_workflow_test.go`
- Modify: `.github/workflows/release-finalize.yml`

**Interfaces:**
- Consumes: validated import from Task 2 and exact confirmation `publish-imported-ao-atlas-<run>-<version>-<tag>-<source>-<manifest>-<plan>`.
- Produces: exact public `v0.2.0` release inventory and `ao-atlas-public-release-verification-<tag>` artifact.

- [ ] **Step 1: Add failing mutation and inventory tests**

Require the publish script to construct and compare:

```bash
expected_confirmation="publish-imported-ao-atlas-${PRODUCER_RUN_ID}-${VERSION}-${TAG}-${SOURCE_SHA}-${MANIFEST_DIGEST}-${PLAN_DIGEST}"
test "$LIVE_CONFIRMATION" = "$expected_confirmation"
```

Require preflight API checks for absent `git/ref/tags/$TAG` and absent `releases/tags/$TAG`, `gh release create "$TAG" --verify-tag`, and exact staged inventory of 15 files: three archives, aggregate checksum, plan plus checksum, and three target-qualified files for each of provenance, SBOM, and signature verification.

Require verification matrix labels `linux-x86_64`, `macos-aarch64`, `windows-x86_64`, exact release flags, exact tag target, `gh release download`, aggregate checksum verification, archive-safe extraction, `--version` identity, and `workgraph validate --workgraph examples/valid/workgraph.json`.

- [ ] **Step 2: Run the finalizer tests and observe RED**

Run: `go test ./internal/atlas -run TestAtlasReleaseFinalizeWorkflow -count=1`

Expected: FAIL on missing exact inventory or public-verification contracts.

- [ ] **Step 3: Implement the live and verification jobs**

Add `publish-release` with:

```yaml
if: ${{ inputs.dry_run == false }}
environment: protected-release
permissions:
  actions: read
  contents: write
```

It redownloads the validated import, verifies the exact confirmation, checks remote tag/release absence, creates a lightweight tag through the GitHub Git refs API at `SOURCE_SHA`, and invokes `gh release create "$TAG" --verify-tag` with committed notes and the 15 exact assets. Do not use overwrite flags.

Add a three-runner `verify-public-release` matrix that depends on publication, uses read-only permissions, verifies API metadata and exact inventory, redownloads all assets, validates aggregate checksums, extracts only the platform archive after traversal/type checks, verifies `ao-atlas version=$VERSION source_sha=$SOURCE_SHA`, and runs the provider-free workgraph fixture. Upload a target-qualified summary with `provider_credentials_used:false`.

- [ ] **Step 4: Run focused finalizer and rehearsal tests**

Run: `go test ./internal/atlas -run 'TestAtlasReleaseFinalizeWorkflow|TestSpecialistReleaseRehearsalWorkflow|TestReleaseRehearsalCandidateVerifier' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit protected publication**

```bash
git add internal/atlas/release_finalize_workflow_test.go .github/workflows/release-finalize.yml
git commit -m "feat: publish Atlas releases through protected review"
```

### Task 4: Update Operator Truth And Run Full Qualification

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `scripts/production-readiness.sh`

**Interfaces:**
- Consumes: finalized workflow contract from Tasks 1–3.
- Produces: durable operator commands/boundaries and a full local gate that covers both release workflows.

- [ ] **Step 1: Write failing documentation/readiness assertions**

Extend the release workflow test to require README literals `release-finalize.yml`, `protected-release`, and `v0.2.0`. Extend `scripts/production-readiness.sh` required files with `.github/workflows/release-finalize.yml` and `docs/release/v0.2.0.md`.

- [ ] **Step 2: Run focused tests and observe RED**

Run: `go test ./internal/atlas -run 'TestAtlasReleaseFinalizeWorkflow|TestSpecialistReleaseRehearsalWorkflow' -count=1`

Expected: FAIL because README still says Atlas cannot release.

- [ ] **Step 3: Update durable operator truth**

Change README installation status to say `v0.2.0` is published only after the governed finalizer succeeds; document the rehearsal/finalizer dispatch sequence and three supported assets. Preserve that ordinary Atlas operation has no publication authority.

Update AGENTS.md to state that release publication is limited to separately authorized exact candidates through `release-finalize.yml`, `protected-release`, and eligible human review; all other tag/release/upload activity remains forbidden.

Update `scripts/production-readiness.sh` to require both workflow files and release notes and run the focused workflow contract tests as part of its existing Go gate, without adding another test framework.

- [ ] **Step 4: Run focused and full local qualification**

Run these commands independently and retain each exit code:

```bash
go test ./internal/atlas -count=1
gofmt -d cmd internal
go test ./... -count=1
go vet ./...
go build ./cmd/atlas
scripts/production-readiness.sh
git diff --check
```

Expected: every command exits 0 and `gofmt -d cmd internal` prints nothing. The cross-repository instruction-layout verifier runs after merge from the canonical synchronized checkout in Task 5.

- [ ] **Step 5: Commit operator truth**

```bash
git add README.md AGENTS.md scripts/production-readiness.sh
git commit -m "docs: document governed Atlas publication"
```

### Task 5: Review, Merge, And Hosted Qualification

**Files:**
- Review: all branch changes from `77ec1577^..HEAD`
- Evidence only: external campaign root outside the public checkouts

**Interfaces:**
- Consumes: clean reviewed branch with all local gates green.
- Produces: merged synchronized `main`, fresh rehearsal run ID and plan digest, successful finalizer dry run, and exact inputs for human-approved live publication.

- [ ] **Step 1: Run independent code and spec review**

Review the diff against `docs/superpowers/specs/2026-08-11-atlas-governed-release-finalize-design.md`. Reject any unbound publication path, broader permission, missing negative validation, asset collision, unsupported platform claim, or historical-evidence edit.

- [ ] **Step 2: Push and open a ready pull request**

Push `codex/atlas-governed-release-finalize`, open a non-draft PR against `main`, and include the exact local gate results. Wait for all hosted checks; fix failures through the same TDD loop.

- [ ] **Step 3: Merge and synchronize**

Merge only after green PR CI and resolved review. Fetch `main`, verify local `main == origin/main`, and remove the task branch/worktree only after retained evidence identifies the merge head.

From the canonical `ao-mission` checkout, run:

```bash
python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-atlas
```

Require exit 0 against the merged canonical Atlas checkout.

- [ ] **Step 4: Dispatch fresh rehearsal**

Dispatch `.github/workflows/release-rehearsal.yml` from the merged exact `main` head with:

```text
version=v0.2.0
tag=v0.2.0
approved_manifest_digest=sha256:6dc6436d59337e829191c48805f65ebbd42998f34066ef50e82ae90236469eb2
dry_run=true
```

Wait for success, download every artifact, verify all checksums and exact source bindings, and record the raw SHA-256 of `promotion-plan.json`.

- [ ] **Step 5: Dispatch finalizer dry run**

Dispatch `.github/workflows/release-finalize.yml` with the successful rehearsal run ID, merged source SHA, `v0.2.0`, approved manifest digest, exact plan digest, `dry_run=true`, and empty live confirmation. Require a successful dry-run boundary and zero tag/release/public-asset mutations.

- [ ] **Step 6: Configure and approve the human gate**

Pause only if `protected-release` does not exist or lacks an eligible required reviewer. The repository administrator creates/configures it; an eligible human approves the exact live deployment. Codex does not change environment policy or self-approve.

- [ ] **Step 7: Dispatch live finalizer and independently verify**

Use the exact confirmation string defined above with `dry_run=false`. After human approval and workflow success, independently resolve tag `v0.2.0`, inspect release flags and exact 15-asset inventory, redownload all assets, recompute SHA-256, run native installed-artifact smokes, and preserve evidence before resuming AO2/Mission/Atlas campaign Gates 6–7.
