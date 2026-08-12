package atlas

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtlasReleaseFinalizeWorkflowStructure(t *testing.T) {
	path := filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read release finalizer workflow: %v", err)
	}
	workflow := string(content)
	if err := validateReleaseFinalizeWorkflowStructure(workflow); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{
			name: "missing protected environment",
			mutate: func(value string) string {
				return strings.Replace(value, "    environment: protected-release\n", "", 1)
			},
			wantErr: "publish-release must use protected-release environment",
		},
		{
			name: "top-level contents write",
			mutate: func(value string) string {
				return strings.Replace(value, "permissions:\n  actions: read\n  contents: read", "permissions:\n  actions: read\n  contents: write", 1)
			},
			wantErr: "workflow permissions must be exactly actions read and contents read",
		},
		{
			name: "validation job contents write",
			mutate: func(value string) string {
				return strings.Replace(value, "  validate-imported-release:\n    runs-on: ubuntu-latest\n    permissions:\n      actions: read\n      contents: read", "  validate-imported-release:\n    runs-on: ubuntu-latest\n    permissions:\n      actions: read\n      contents: write", 1)
			},
			wantErr: `publication capability "contents: write" must be limited to publish-release`,
		},
		{
			name: "validation job actions write",
			mutate: func(value string) string {
				return strings.Replace(value, "  validate-imported-release:\n    runs-on: ubuntu-latest\n    permissions:\n      actions: read\n      contents: read", "  validate-imported-release:\n    runs-on: ubuntu-latest\n    permissions:\n      actions: write\n      contents: read", 1)
			},
			wantErr: "write permission must be limited to publish-release",
		},
		{
			name: "missing live gate",
			mutate: func(value string) string {
				return strings.Replace(value, "inputs.dry_run == false", "inputs.dry_run", 1)
			},
			wantErr: "publish-release must require live mode",
		},
		{
			name: "missing exact live confirmation",
			mutate: func(value string) string {
				return strings.Replace(value, `test "$LIVE_CONFIRMATION" = "$expected_confirmation"`, `test -n "$LIVE_CONFIRMATION"`, 1)
			},
			wantErr: "publish-release must bind the exact live confirmation",
		},
		{
			name: "missing verified tag",
			mutate: func(value string) string {
				return strings.Replace(value, `--verify-tag`, `--target "$SOURCE_SHA"`, 1)
			},
			wantErr: "release creation must require the pre-created tag",
		},
		{
			name: "release overwrite flag",
			mutate: func(value string) string {
				return strings.Replace(value, `gh release create "$TAG"`, `gh release create "$TAG" --clobber`, 1)
			},
			wantErr: `forbidden release finalizer capability "--clobber"`,
		},
		{
			name: "draft release flag",
			mutate: func(value string) string {
				return strings.Replace(value, `gh release create "$TAG"`, `gh release create "$TAG" --draft`, 1)
			},
			wantErr: `forbidden release finalizer capability "--draft"`,
		},
		{
			name: "missing staged provenance asset",
			mutate: func(value string) string {
				return strings.Replace(value, `"linux-x86_64-provenance.json"`, `"unexpected-provenance.json"`, 1)
			},
			wantErr: "publish-release must stage the exact public inventory",
		},
		{
			name: "wrong public release target",
			mutate: func(value string) string {
				return strings.Replace(value, `.target_commitish == $source_sha`, `.target_commitish != $source_sha`, 1)
			},
			wantErr: "public release must bind the exact source target",
		},
		{
			name: "missing authorized plan anchor",
			mutate: func(value string) string {
				return strings.Replace(value, `test "$(hash_file release-assets/promotion-plan.json)" = "$PLAN_DIGEST"`, `test -s release-assets/promotion-plan.json`, 1)
			},
			wantErr: "public assets must bind to the authorized promotion plan",
		},
		{
			name: "wrong public release title",
			mutate: func(value string) string {
				return strings.Replace(value, `.name == $version`, `.name != $version`, 1)
			},
			wantErr: "public release must bind exact title and body",
		},
		{
			name: "missing public release body comparison",
			mutate: func(value string) string {
				return strings.Replace(value, `cmp public-release-body.txt "docs/release/${VERSION}.md"`, `test -s public-release-body.txt`, 1)
			},
			wantErr: "public release must bind exact title and body",
		},
		{
			name: "unsafe archive type accepted",
			mutate: func(value string) string {
				return strings.Replace(value, `if not member.isfile():`, `if member.isfile():`, 1)
			},
			wantErr: "public archive extraction must reject non-regular entries",
		},
		{
			name: "traversal guard weakened to and",
			mutate: func(value string) string {
				return strings.Replace(value, `path.is_absolute() or ".." in path.parts or "\\" in member.name`, `path.is_absolute() and ".." in path.parts and "\\" in member.name`, 1)
			},
			wantErr: "public archive extraction must reject traversal paths",
		},
		{
			name: "duplicate archive check removed",
			mutate: func(value string) string {
				return strings.Replace(value, `if member.name in seen:`, `if False:`, 1)
			},
			wantErr: "public archive extraction must reject duplicate entries",
		},
		{
			name: "exact archive inventory weakened",
			mutate: func(value string) string {
				return strings.Replace(value, `if seen != expected:`, `if not expected.issubset(seen):`, 1)
			},
			wantErr: "public archive extraction must require exact inventory",
		},
		{
			name: "bare python archive verifier",
			mutate: func(value string) string {
				return strings.Replace(value, `python3 - <<'PY'`, `python - <<'PY'`, 1)
			},
			wantErr: "public archive verifier must use python3",
		},
		{
			name: "disabled public verification",
			mutate: func(value string) string {
				return strings.Replace(value, `if: ${{ needs.publish-release.result == 'success' }}`, `if: ${{ needs.publish-release.result == 'success' && false }}`, 1)
			},
			wantErr: "public verification must run after publication",
		},
		{
			name: "stale consolidated identity",
			mutate: func(value string) string {
				return strings.Replace(value, `.version_identity == $expected_version_identity`, `.version_identity != $expected_version_identity`, 1)
			},
			wantErr: "consolidated verification must bind exact identity and results",
		},
		{
			name: "false consolidated verification accepted",
			mutate: func(value string) string {
				return strings.Replace(value, `.authorized_plan_verified == true`, `.authorized_plan_verified != true`, 1)
			},
			wantErr: "consolidated verification must bind exact identity and results",
		},
		{
			name: "wrong producer workflow",
			mutate: func(value string) string {
				return strings.ReplaceAll(value, ".github/workflows/release-rehearsal.yml", ".github/workflows/ci.yml")
			},
			wantErr: "producer workflow path must be exact",
		},
		{
			name: "missing expected plan digest",
			mutate: func(value string) string {
				return strings.Replace(value, "      expected_plan_digest:\n", "", 1)
			},
			wantErr: `missing required workflow input "expected_plan_digest:"`,
		},
		{
			name: "missing workflow source binding",
			mutate: func(value string) string {
				return strings.Replace(value, "          test \"$WORKFLOW_SHA\" = \"$SOURCE_SHA\"\n", "", 1)
			},
			wantErr: "finalizer workflow source must equal expected source",
		},
		{
			name: "suffix-only producer identity",
			mutate: func(value string) string {
				return strings.Replace(value, ".workflow_identity == $producer_identity", `.workflow_identity | endswith("/actions/runs/" + $run_id)`, 1)
			},
			wantErr: "candidate producer identity must be exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReleaseFinalizeWorkflowStructure(tt.mutate(workflow))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAtlasReleaseFinalizeWorkflowPublishesExactInventory(t *testing.T) {
	workflow := readReleaseFinalizeWorkflow(t)
	publish := yamlJobSection(yamlTopLevelSection(workflow, "jobs:"), "publish-release")
	for _, required := range []string{
		`uses: actions/download-artifact@v7`,
		`name: ao-atlas-validated-release-import-${{ inputs.producer_run_id }}`,
		`expected_confirmation="publish-imported-ao-atlas-${PRODUCER_RUN_ID}-${VERSION}-${TAG}-${SOURCE_SHA}-${MANIFEST_DIGEST}-${PLAN_DIGEST}"`,
		`test "$LIVE_CONFIRMATION" = "$expected_confirmation"`,
		`git/ref/tags/$TAG`,
		`releases/tags/$TAG`,
		`--method POST "repos/$GITHUB_REPOSITORY/git/refs"`,
		`-f ref="refs/tags/$TAG"`,
		`-f sha="$SOURCE_SHA"`,
		`gh release create "$TAG"`,
		`--verify-tag`,
		`--target "$SOURCE_SHA"`,
		`--title "$VERSION"`,
		`--notes-file "docs/release/${VERSION}.md"`,
		`"ao-atlas-${VERSION}-linux-x86_64.tar.gz"`,
		`"ao-atlas-${VERSION}-macos-aarch64.tar.gz"`,
		`"ao-atlas-${VERSION}-windows-x86_64.tar.gz"`,
		`"linux-x86_64-provenance.json"`,
		`"macos-aarch64-provenance.json"`,
		`"windows-x86_64-provenance.json"`,
		`"linux-x86_64-sbom.spdx.json"`,
		`"macos-aarch64-sbom.spdx.json"`,
		`"windows-x86_64-sbom.spdx.json"`,
		`"linux-x86_64-signature-verification.json"`,
		`"macos-aarch64-signature-verification.json"`,
		`"windows-x86_64-signature-verification.json"`,
		`"SHA256SUMS"`,
		`"promotion-plan.json"`,
		`"promotion-plan.sha256"`,
		`test "$(find validated-release -maxdepth 1 -type f | wc -l)" -eq 15`,
		`cmp actual-assets.txt expected-assets.txt`,
	} {
		if !strings.Contains(publish, required) {
			t.Errorf("publish-release missing exact inventory contract %q", required)
		}
	}
}

func TestAtlasReleaseFinalizeWorkflowVerifiesEveryPublicTarget(t *testing.T) {
	workflow := readReleaseFinalizeWorkflow(t)
	verify := yamlJobSection(yamlTopLevelSection(workflow, "jobs:"), "verify-public-release")
	for _, required := range []string{
		`if: ${{ needs.publish-release.result == 'success' }}`,
		`target_label: linux-x86_64`,
		`target_label: macos-aarch64`,
		`target_label: windows-x86_64`,
		`gh release download "$TAG"`,
		`.target_commitish == $source_sha`,
		`.name == $version`,
		`cmp public-release-body.txt "docs/release/${VERSION}.md"`,
		`cmp actual-assets.txt expected-assets.txt`,
		`PLAN_DIGEST: ${{ inputs.expected_plan_digest }}`,
		`test "$(hash_file release-assets/promotion-plan.json)" = "$PLAN_DIGEST"`,
		`.release_notes_sha256 == $notes_digest`,
		`test "sha256:$actual_digest" = "$expected_digest"`,
		`sha256sum --check --strict SHA256SUMS`,
		`shasum -a 256 --check SHA256SUMS`,
		`PurePosixPath(member.name)`,
		`path.is_absolute()`,
		`".." in path.parts`,
		`member.isfile()`,
		`python3 - <<'PY'`,
		`expected_version_identity="ao-atlas version=$VERSION source_sha=$SOURCE_SHA"`,
		`workgraph validate --workgraph examples/valid/workgraph.json`,
		`provider_credentials_used:false`,
		`name: ao-atlas-public-release-verification-${{ inputs.expected_tag }}-${{ matrix.target_label }}`,
	} {
		if !strings.Contains(verify, required) {
			t.Errorf("verify-public-release missing contract %q", required)
		}
	}
	for _, forbidden := range []string{"environment:", "contents: write", "gh release create", "gh release upload"} {
		if strings.Contains(verify, forbidden) {
			t.Errorf("verify-public-release has forbidden capability %q", forbidden)
		}
	}
	if !strings.Contains(workflow, `name: ao-atlas-public-release-verification-${{ inputs.expected_tag }}`) {
		t.Fatal("workflow must upload the exact consolidated public verification artifact")
	}
}

func TestAtlasReleaseFinalizeWorkflowConsolidatesExactVerifiedResults(t *testing.T) {
	workflow := readReleaseFinalizeWorkflow(t)
	consolidate := yamlJobSection(yamlTopLevelSection(workflow, "jobs:"), "consolidate-public-release-verification")
	for _, required := range []string{
		`SOURCE_SHA: ${{ inputs.expected_source_sha }}`,
		`TAG: ${{ inputs.expected_tag }}`,
		`VERSION: ${{ inputs.expected_version }}`,
		`.tag == $tag`,
		`.source_sha == $source_sha`,
		`.version_identity == $expected_version_identity`,
		`.aggregate_checksums_verified == true`,
		`.archive_inventory_verified == true`,
		`.authorized_plan_verified == true`,
		`.authorized_asset_digests_verified == true`,
		`.release_metadata_verified == true`,
		`.provider_free_workgraph_validated == true`,
	} {
		if !strings.Contains(consolidate, required) {
			t.Errorf("consolidated public verification missing contract %q", required)
		}
	}
}

func readReleaseFinalizeWorkflow(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestAtlasReleaseFinalizeWorkflowAuthenticatesExactProducerArtifactInventory(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(content)
	for _, required := range []string{
		`.total_count == 5`,
		`[.artifacts[].name] | unique | length) == 5`,
		`all(.artifacts[]; .expired == false`,
		`"ao-atlas-release-input-binding-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-linux-x86_64-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-macos-aarch64-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-windows-x86_64-$SOURCE_SHA"`,
		`"ao-atlas-release-rehearsal-plan-$SOURCE_SHA"`,
		`cmp actual-artifacts.txt expected-artifacts.txt`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("producer artifact authentication missing %q", required)
		}
	}
}

func TestAtlasReleaseFinalizeProducerIdentityRejectsDifferentRepository(t *testing.T) {
	predicate := `all(.candidates[]; .workflow_identity == $producer_identity)`
	expected := "https://github.com/uesugitorachiyo/ao-atlas/actions/runs/123"
	for _, tt := range []struct {
		name     string
		identity string
		wantOK   bool
	}{
		{name: "exact", identity: expected, wantOK: true},
		{name: "different repository", identity: "https://github.com/other/ao-atlas/actions/runs/123"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixture := fmt.Sprintf(`{"candidates":[{"workflow_identity":%q}]}`, tt.identity)
			cmd := exec.Command("jq", "-e", "--arg", "producer_identity", expected, predicate)
			cmd.Stdin = strings.NewReader(fixture)
			err := cmd.Run()
			if (err == nil) != tt.wantOK {
				t.Fatalf("identity %q acceptance = %v, want %v", tt.identity, err == nil, tt.wantOK)
			}
		})
	}
}

func TestAtlasReleaseFinalizeWorkflowRejectsUntrustedArtifactsBeforeExtraction(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(content)
	required := []string{
		`WORKFLOW_SHA: ${{ github.workflow_sha }}`,
		`test "$WORKFLOW_SHA" = "$SOURCE_SHA"`,
		`map(.size_in_bytes) | add <= 128 * 1024 * 1024`,
		`repos/$GITHUB_REPOSITORY/actions/artifacts/$artifact_id/zip`,
		`scripts/inspect-release-artifact-zips.py`,
		`expected_identity="$GITHUB_SERVER_URL/$GITHUB_REPOSITORY/actions/runs/$PRODUCER_RUN_ID"`,
		`all(.candidates[]; .workflow_identity == $producer_identity)`,
		`release-input-binding.sha256`,
		`test -z "$(find "$plan_dir" -mindepth 2 -print -quit)"`,
	}
	for _, value := range required {
		if !strings.Contains(workflow, value) {
			t.Errorf("release import boundary missing %q", value)
		}
	}
	checkout := strings.Index(workflow, "uses: actions/checkout@v7")
	workflowSHA := strings.Index(workflow, `test "$WORKFLOW_SHA" = "$SOURCE_SHA"`)
	rawDownload := strings.Index(workflow, `actions/artifacts/$artifact_id/zip`)
	inspection := strings.Index(workflow, `scripts/inspect-release-artifact-zips.py`)
	extraction := strings.Index(workflow, "uses: actions/download-artifact@v7")
	if workflowSHA < 0 || checkout < 0 || workflowSHA >= checkout {
		t.Error("workflow SHA must be authenticated before checkout")
	}
	if rawDownload < 0 || inspection < 0 || extraction < 0 || rawDownload >= inspection || inspection >= extraction {
		t.Error("raw artifacts must be inspected before actions/download-artifact extracts them")
	}
}

func TestAtlasReleaseFinalizeWorkflowRejectsUnexpectedEmptyArtifactDirectories(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(content)
	for _, tt := range []struct {
		name     string
		variable string
		files    []string
	}{
		{name: "input binding", variable: "binding_dir", files: []string{"release-input-binding.json", "release-input-binding.sha256"}},
		{name: "plan", variable: "plan_dir", files: []string{"promotion-plan.json", "promotion-plan.sha256", "dry-run-boundary.json"}},
		{name: "candidate", variable: "candidate_dir", files: []string{"candidate-summary.json"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			immediateGuard := fmt.Sprintf(`test -z "$(find "$%s" -mindepth 1 -maxdepth 1 ! -type f -print -quit)"`, tt.variable)
			nestedGuard := fmt.Sprintf(`test -z "$(find "$%s" -mindepth 2 -print -quit)"`, tt.variable)
			for _, guard := range []string{immediateGuard, nestedGuard} {
				if !strings.Contains(workflow, guard) {
					t.Fatalf("workflow missing exact %s inventory guard %q", tt.name, guard)
				}
			}
			dir := t.TempDir()
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := runArtifactInventoryGuard(immediateGuard+"\n"+nestedGuard, tt.variable, dir); err != nil {
				t.Fatalf("exact %s inventory rejected: %v", tt.name, err)
			}
			if err := os.Mkdir(filepath.Join(dir, "unexpected"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := runArtifactInventoryGuard(immediateGuard+"\n"+nestedGuard, tt.variable, dir); err == nil {
				t.Fatalf("%s inventory accepted unexpected empty directory", tt.name)
			}
			if err := os.WriteFile(filepath.Join(dir, "unexpected", "nested"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := runArtifactInventoryGuard(immediateGuard+"\n"+nestedGuard, tt.variable, dir); err == nil {
				t.Fatalf("%s inventory accepted unexpected nested file", tt.name)
			}
		})
	}
}

func runArtifactInventoryGuard(guard, variable, dir string) error {
	cmd := exec.Command("bash", "-c", "set -euo pipefail\n"+guard)
	cmd.Env = append(os.Environ(), variable+"="+dir)
	return cmd.Run()
}

func TestReleaseArtifactZipInspectorRejectsUnsafeEntriesAndBounds(t *testing.T) {
	script := filepath.Join(repoRoot(t), "scripts", "inspect-release-artifact-zips.py")
	tests := []struct {
		name       string
		entries    []releaseArtifactZipEntry
		maxEntries string
		maxBytes   string
		wantErr    string
	}{
		{name: "safe", entries: []releaseArtifactZipEntry{{name: "candidate.json", content: "{}\n"}}, maxEntries: "1", maxBytes: "3"},
		{name: "safe stored", entries: []releaseArtifactZipEntry{{name: "candidate.json", content: "{}\n", stored: true}}, maxEntries: "1", maxBytes: "3"},
		{name: "traversal", entries: []releaseArtifactZipEntry{{name: "../candidate.json", content: "{}\n"}}, maxEntries: "1", maxBytes: "3", wantErr: "unsafe path"},
		{name: "symlink", entries: []releaseArtifactZipEntry{{name: "candidate.json", mode: os.ModeSymlink | 0o777}}, maxEntries: "1", maxBytes: "3", wantErr: "non-regular entry"},
		{name: "entry bound", entries: []releaseArtifactZipEntry{{name: "one", content: "1"}, {name: "two", content: "2"}}, maxEntries: "1", maxBytes: "2", wantErr: "entry count exceeds"},
		{name: "expanded bound", entries: []releaseArtifactZipEntry{{name: "candidate.json", content: "1234"}}, maxEntries: "1", maxBytes: "3", wantErr: "expanded size exceeds"},
		{name: "canonical file directory collision", entries: []releaseArtifactZipEntry{{name: "entry", content: "1"}, {name: "entry/", mode: os.ModeDir | 0o755}}, maxEntries: "2", maxBytes: "1", wantErr: "duplicate path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := writeReleaseArtifactZip(t, tt.entries)
			cmd := exec.Command("python3", script, "--max-entries", tt.maxEntries, "--max-expanded-bytes", tt.maxBytes, archive)
			output, err := cmd.CombinedOutput()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("safe artifact rejected: %v\n%s", err, output)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(string(output), tt.wantErr)) {
				t.Fatalf("expected %q, got %v\n%s", tt.wantErr, err, output)
			}
		})
	}
}

func TestReleaseArtifactZipInspectorRejectsForgedShortPrefixMetadata(t *testing.T) {
	archive := writeReleaseArtifactZip(t, []releaseArtifactZipEntry{{name: "candidate.json", content: strings.Repeat("x", 1024)}})
	content, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	centralDirectory := bytes.LastIndex(content, []byte{'P', 'K', 1, 2})
	if centralDirectory < 0 {
		t.Fatal("ZIP central directory not found")
	}
	prefixCRC := crc32.ChecksumIEEE([]byte("x"))
	binary.LittleEndian.PutUint32(content[centralDirectory+16:centralDirectory+20], prefixCRC)
	binary.LittleEndian.PutUint32(content[centralDirectory+24:centralDirectory+28], 1)
	dataDescriptor := bytes.LastIndex(content[:centralDirectory], []byte{'P', 'K', 7, 8})
	if dataDescriptor < 0 {
		t.Fatal("ZIP data descriptor not found")
	}
	binary.LittleEndian.PutUint32(content[dataDescriptor+4:dataDescriptor+8], prefixCRC)
	binary.LittleEndian.PutUint32(content[dataDescriptor+12:dataDescriptor+16], 1)
	if err := os.WriteFile(archive, content, 0o600); err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(repoRoot(t), "scripts", "inspect-release-artifact-zips.py")
	cmd := exec.Command("python3", script, "--max-entries", "1", "--max-expanded-bytes", "2048", archive)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("inspector accepted forged short-prefix size and CRC\n%s", output)
	}
}

func TestReleaseArtifactZipInspectorRejectsUnsupportedCompressionMethod(t *testing.T) {
	archive := writeReleaseArtifactZip(t, []releaseArtifactZipEntry{{name: "candidate.json", content: "{}\n"}})
	content, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	centralDirectory := bytes.LastIndex(content, []byte{'P', 'K', 1, 2})
	localHeader := bytes.Index(content, []byte{'P', 'K', 3, 4})
	if centralDirectory < 0 || localHeader < 0 {
		t.Fatal("ZIP headers not found")
	}
	binary.LittleEndian.PutUint16(content[centralDirectory+10:centralDirectory+12], 99)
	binary.LittleEndian.PutUint16(content[localHeader+8:localHeader+10], 99)
	if err := os.WriteFile(archive, content, 0o600); err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(repoRoot(t), "scripts", "inspect-release-artifact-zips.py")
	cmd := exec.Command("python3", script, "--max-entries", "1", "--max-expanded-bytes", "3", archive)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "unsupported compression method") {
		t.Fatalf("unsupported compression method accepted: %v\n%s", err, output)
	}
}

type releaseArtifactZipEntry struct {
	name    string
	content string
	mode    os.FileMode
	stored  bool
}

func writeReleaseArtifactZip(t *testing.T, entries []releaseArtifactZipEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifact.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		method := uint16(zip.Deflate)
		if entry.stored {
			method = zip.Store
		}
		header := &zip.FileHeader{Name: entry.name, Method: method}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		part, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func validateReleaseFinalizeWorkflowStructure(workflow string) error {
	required := []string{
		"workflow_dispatch:", "producer_run_id:", "expected_source_sha:",
		"expected_version:", "expected_tag:", "expected_manifest_digest:",
		"expected_plan_digest:", "dry_run:", "live_confirmation:",
		"actions: read", "contents: read", "validate-imported-release:",
		"dry-run-boundary:", "publish-release:", "verify-public-release:",
		"environment: protected-release", "contents: write",
		"inputs.dry_run == false", "actions/runs/$PRODUCER_RUN_ID",
		".github/workflows/release-rehearsal.yml", `conclusion == "success"`,
		"scripts/verify-release-rehearsal-candidates.sh", "gh release create",
		"gh release download", "git/ref/tags/$TAG", "provider_credentials_used:false",
		"consolidate-public-release-verification:",
	}
	for _, value := range required {
		if !strings.Contains(workflow, value) {
			switch value {
			case "actions: read", "contents: read":
				return fmt.Errorf("workflow permissions must be exactly actions read and contents read")
			case "expected_plan_digest:":
				return fmt.Errorf("missing required workflow input %q", value)
			case "environment: protected-release":
				return fmt.Errorf("publish-release must use protected-release environment")
			case "inputs.dry_run == false":
				return fmt.Errorf("publish-release must require live mode")
			case ".github/workflows/release-rehearsal.yml":
				return fmt.Errorf("producer workflow path must be exact")
			default:
				return fmt.Errorf("release finalizer workflow missing %q", value)
			}
		}
	}
	if !strings.Contains(workflow, `test "$WORKFLOW_SHA" = "$SOURCE_SHA"`) {
		return fmt.Errorf("finalizer workflow source must equal expected source")
	}
	if !strings.Contains(workflow, ".workflow_identity == $producer_identity") {
		return fmt.Errorf("candidate producer identity must be exact")
	}

	triggers := yamlChildKeys(yamlTopLevelSection(workflow, "on:"), 2)
	if len(triggers) != 1 || triggers[0] != "workflow_dispatch" {
		return fmt.Errorf("workflow_dispatch must be the only trigger: %v", triggers)
	}
	permissions := strings.TrimSpace(yamlTopLevelSection(workflow, "permissions:"))
	if permissions != "permissions:\n  actions: read\n  contents: read" {
		return fmt.Errorf("workflow permissions must be exactly actions read and contents read")
	}

	jobs := yamlTopLevelSection(workflow, "jobs:")
	publish := yamlJobSection(jobs, "publish-release")
	if publish == "" || !strings.Contains(publish, "environment: protected-release") {
		return fmt.Errorf("publish-release must use protected-release environment")
	}
	if !strings.Contains(publish, "inputs.dry_run == false") {
		return fmt.Errorf("publish-release must require live mode")
	}
	if !strings.Contains(publish, "permissions:\n      actions: read\n      contents: write") {
		return fmt.Errorf("publish-release permissions must be exactly actions read and contents write")
	}
	if !strings.Contains(publish, `test "$LIVE_CONFIRMATION" = "$expected_confirmation"`) {
		return fmt.Errorf("publish-release must bind the exact live confirmation")
	}
	if !strings.Contains(publish, "--verify-tag") {
		return fmt.Errorf("release creation must require the pre-created tag")
	}
	for _, asset := range []string{
		`"ao-atlas-${VERSION}-linux-x86_64.tar.gz"`,
		`"ao-atlas-${VERSION}-macos-aarch64.tar.gz"`,
		`"ao-atlas-${VERSION}-windows-x86_64.tar.gz"`,
		`"linux-x86_64-provenance.json"`, `"linux-x86_64-sbom.spdx.json"`, `"linux-x86_64-signature-verification.json"`,
		`"macos-aarch64-provenance.json"`, `"macos-aarch64-sbom.spdx.json"`, `"macos-aarch64-signature-verification.json"`,
		`"windows-x86_64-provenance.json"`, `"windows-x86_64-sbom.spdx.json"`, `"windows-x86_64-signature-verification.json"`,
		`"SHA256SUMS"`, `"promotion-plan.json"`, `"promotion-plan.sha256"`,
	} {
		if !strings.Contains(publish, asset) {
			return fmt.Errorf("publish-release must stage the exact public inventory")
		}
	}
	verify := yamlJobSection(jobs, "verify-public-release")
	if verify == "" || !strings.Contains(verify, `if: ${{ needs.publish-release.result == 'success' }}`) || strings.Contains(verify, "&& false") {
		return fmt.Errorf("public verification must run after publication")
	}
	if !strings.Contains(verify, `.target_commitish == $source_sha`) {
		return fmt.Errorf("public release must bind the exact source target")
	}
	if !strings.Contains(verify, `.name == $version`) || !strings.Contains(verify, `cmp public-release-body.txt "docs/release/${VERSION}.md"`) {
		return fmt.Errorf("public release must bind exact title and body")
	}
	for _, binding := range []string{
		`PLAN_DIGEST: ${{ inputs.expected_plan_digest }}`,
		`test "$(hash_file release-assets/promotion-plan.json)" = "$PLAN_DIGEST"`,
		`.release_notes_sha256 == $notes_digest`,
		`test "sha256:$actual_digest" = "$expected_digest"`,
	} {
		if !strings.Contains(verify, binding) {
			return fmt.Errorf("public assets must bind to the authorized promotion plan")
		}
	}
	if !strings.Contains(verify, `if not member.isfile():`) {
		return fmt.Errorf("public archive extraction must reject non-regular entries")
	}
	if !strings.Contains(verify, `path.is_absolute() or ".." in path.parts or "\\" in member.name`) {
		return fmt.Errorf("public archive extraction must reject traversal paths")
	}
	if !strings.Contains(verify, `if member.name in seen:`) {
		return fmt.Errorf("public archive extraction must reject duplicate entries")
	}
	if !strings.Contains(verify, `if seen != expected:`) {
		return fmt.Errorf("public archive extraction must require exact inventory")
	}
	if !strings.Contains(verify, `python3 - <<'PY'`) || strings.Contains(verify, `python - <<'PY'`) {
		return fmt.Errorf("public archive verifier must use python3")
	}
	consolidate := yamlJobSection(jobs, "consolidate-public-release-verification")
	for _, binding := range []string{
		`SOURCE_SHA: ${{ inputs.expected_source_sha }}`,
		`TAG: ${{ inputs.expected_tag }}`,
		`VERSION: ${{ inputs.expected_version }}`,
		`.tag == $tag`, `.source_sha == $source_sha`, `.version_identity == $expected_version_identity`,
		`.aggregate_checksums_verified == true`, `.archive_inventory_verified == true`,
		`.authorized_plan_verified == true`, `.authorized_asset_digests_verified == true`,
		`.release_metadata_verified == true`, `.provider_free_workgraph_validated == true`,
	} {
		if !strings.Contains(consolidate, binding) {
			return fmt.Errorf("consolidated verification must bind exact identity and results")
		}
	}
	nonPublish := strings.Replace(jobs, publish, "", 1)
	for _, capability := range []string{"contents: write", "gh release create", "gh release upload", "git tag", "git push"} {
		if strings.Contains(nonPublish, capability) {
			return fmt.Errorf("publication capability %q must be limited to publish-release", capability)
		}
	}
	if strings.Contains(nonPublish, ": write") {
		return fmt.Errorf("write permission must be limited to publish-release")
	}
	for _, forbidden := range []string{"write-all", "pull-requests: write", "/environments", "gh release delete", "gh release upload", "git push --force", "--clobber", "--draft", "--prerelease", "--force"} {
		if strings.Contains(workflow, forbidden) {
			return fmt.Errorf("forbidden release finalizer capability %q", forbidden)
		}
	}
	return nil
}

func yamlJobSection(jobs, name string) string {
	header := "  " + name + ":"
	lines := strings.Split(jobs, "\n")
	start := -1
	for i, line := range lines {
		if line == header {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "  ") && !strings.HasPrefix(lines[i], "    ") && strings.HasSuffix(lines[i], ":") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}
