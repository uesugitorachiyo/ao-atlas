# Atlas Governed Release Finalize Design

## Purpose

AO Atlas has a qualified three-platform `v0.2.0` rehearsal, but its frozen source exposes no governed live-promotion workflow. This change adds the missing source-owned finalizer and qualifies a new exact source head. It does not publish by itself, configure repository permissions, create a GitHub environment, or approve an environment review.

## Chosen Approach

Add a separate `.github/workflows/release-finalize.yml` that consumes one successful `release-rehearsal.yml` run. Keep rehearsal publication-disabled. This preserves the authority split: rehearsal builds and seals candidates; finalization authenticates and imports those exact artifacts, waits for protected-environment review, publishes them, and verifies the resulting public release.

Extending rehearsal with a live mode would mix qualification and publication authority. Reusing another repository's finalizer would make Atlas release availability depend on a sibling repository. Neither tradeoff is needed.

## Release Contract

The finalizer accepts these dispatch inputs:

- `producer_run_id`: successful Atlas release-rehearsal workflow run;
- `expected_source_sha`: exact 40-character source commit;
- `expected_version`: exact semantic version, including the existing Atlas `v` prefix;
- `expected_tag`: exact tag, equal to `expected_version`;
- `expected_manifest_digest`: `sha256:<64 lowercase hex>` approved manifest digest;
- `expected_plan_digest`: raw 64-character SHA-256 of `promotion-plan.json`;
- `dry_run`: defaults to `true` and performs no public mutation;
- `live_confirmation`: exact string derived from all frozen inputs.

The exact live confirmation format is:

```text
publish-imported-ao-atlas-<producer_run_id>-<expected_version>-<expected_tag>-<expected_source_sha>-<expected_manifest_digest>-<expected_plan_digest>
```

The workflow validates the producer through the GitHub Actions API. Repository, workflow path, event, conclusion, source SHA, age, and artifact names must match. It downloads artifacts only from that run, rejects expired or duplicate artifacts, bounds file count and total size, rejects symlinks and non-regular files, and verifies every candidate `SHA256SUMS` before evaluating JSON.

The existing `scripts/verify-release-rehearsal-candidates.sh` remains the candidate authority. The finalizer reruns it against imported candidates, rebuilds a comparison promotion plan, and requires byte identity with the imported plan and the supplied plan digest.

## Published Assets

The live job creates an exact lightweight tag at `expected_source_sha`, then a public, non-draft, non-prerelease release. It uploads:

- the Linux x86_64, macOS aarch64, and Windows x86_64 archives exactly once;
- one aggregate `SHA256SUMS` for those archives;
- `promotion-plan.json` and `promotion-plan.sha256`;
- target-qualified provenance, SBOM, and signature-verification evidence copied from each sealed candidate;
- committed `docs/release/v0.2.0.md` as the release body.

Evidence filenames include the target label so files from different candidates cannot collide. The workflow does not claim or upload Linux aarch64 support.

The rehearsal plan binds the committed release-notes digest. Rehearsal validation fails if the requested version is not `v0.2.0`, the committed notes are absent, or their digest is not carried into the promotion plan. This fresh qualification replaces the old frozen candidate; it does not reinterpret run `31551988501`.

## Authority And Environment

Workflow-level permissions remain read-only. Only the live publish job receives `contents: write`, and only when `dry_run == false`, every validation dependency succeeds, and `live_confirmation` matches exactly.

The publish job references `protected-release`. A repository administrator must create that environment and configure eligible required reviewers. The workflow must not create or modify the environment. The operator's instruction to fix, qualify, validate, and continue releasing authorizes publication of the resulting new head after every source-owned gate passes. Codex may dispatch and monitor that exact qualified head, but an eligible human performs the protected-environment approval.

The administrator must also configure both the environment-scoped secret and variable `AO_ATLAS_PROTECTED_RELEASE_SENTINEL` to `protected-release-required-reviewers-configured`. Missing or mismatched sentinels fail before mutation if GitHub auto-creates or misconfigures the environment. The sentinels do not prove reviewer policy; external verification of eligible required reviewers remains mandatory before merge and live dispatch.

Dry-run finalization emits a retained boundary artifact proving zero tag, release, public upload, deployment, provider, credential, and permission mutations.

## Failure And Recovery

The finalizer fails before mutation on source, tag, version, manifest, plan, workflow-run, artifact, release-note, inventory, checksum, or confirmation drift. It also requires remote tag and release absence immediately before publication.

Once the tag is created, a release-creation failure is reported as a partial publication requiring operator reconciliation. The workflow never deletes, rewrites, or force-moves a tag. Reruns find the existing tag and stop. Recovery must independently inspect the exact tag target and either complete the release under separate authority or preserve the partial state for repair.

Post-publication verification redownloads the public assets into a clean directory, requires exact inventory and SHA-256 agreement with the imported candidates, verifies tag target and release flags, runs the installed provider-free smoke on the native runner matrix, and uploads a verification summary.

## Source Changes

- Add `.github/workflows/release-finalize.yml`.
- Add `docs/release/v0.2.0.md`.
- Extend `.github/workflows/release-rehearsal.yml` only to bind the committed release-note digest into the immutable plan.
- Extend `scripts/verify-release-rehearsal-candidates.sh` with the release-note digest input and promotion-plan field.
- Add focused workflow-contract and verifier tests under `internal/atlas/`.
- Update `README.md` and `AGENTS.md` because the durable release command and Atlas publication boundary change.

No new dependency or reusable release framework is introduced.

## Verification

Tests are written first and observed failing before workflow or verifier changes. Focused tests prove:

- the finalizer has the exact inputs, minimal permissions, protected environment, live gate, producer authentication, immutable artifact import, pre-publication absence check, exact asset inventory, and post-publication verification;
- dry runs cannot reach publication;
- malformed IDs, stale or wrong producer runs, digest drift, duplicate candidates, unsafe files, wrong confirmation, and existing tags/releases fail closed;
- rehearsal binds committed release notes and still cannot publish;
- the public inventory contains only the three supported archives and target-qualified evidence.

Full qualification runs:

```text
go test ./internal/atlas -count=1
gofmt -d cmd internal
go test ./... -count=1
go vet ./...
go build ./cmd/atlas
scripts/production-readiness.sh
python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-atlas
git diff --check
```

After merge and synchronized `main`, dispatch a fresh three-platform release rehearsal for `v0.2.0`, retain its exact artifacts, independently verify its promotion-plan digest, then run the finalizer in dry-run mode. Live publication proceeds only after `protected-release` has eligible reviewers and an eligible human approves the exact qualified inputs.
