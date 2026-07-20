package atlas

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	rehearsalVersion        = "v0.2.0"
	rehearsalSourceSHA      = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	rehearsalManifestDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type rehearsalTarget struct {
	Label  string
	GOOS   string
	GOARCH string
	Binary string
}

var rehearsalTargets = []rehearsalTarget{
	{Label: "linux-x86_64", GOOS: "linux", GOARCH: "amd64", Binary: "ao-atlas"},
	{Label: "macos-x86_64", GOOS: "darwin", GOARCH: "amd64", Binary: "ao-atlas"},
	{Label: "windows-x86_64", GOOS: "windows", GOARCH: "amd64", Binary: "ao-atlas.exe"},
}

func TestAtlasVersionCommandDefaultsToDevelopmentIdentity(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("--version failed with code %d: %s", code, stderr.String())
	}
	if stdout.String() != "ao-atlas version=dev source_sha=unknown\n" {
		t.Fatalf("unexpected default version identity: %q", stdout.String())
	}
}

func TestSpecialistReleaseRehearsalWorkflowStructure(t *testing.T) {
	workflow := readReleaseRehearsalWorkflow(t)
	if err := validateReleaseRehearsalWorkflowStructure(workflow); err != nil {
		t.Fatal(err)
	}
}

func TestSpecialistReleaseRehearsalWorkflowRejectsUnsafeStructures(t *testing.T) {
	workflow := readReleaseRehearsalWorkflow(t)
	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{
			name: "non dispatch trigger",
			mutate: func(value string) string {
				return strings.Replace(value, "  workflow_dispatch:\n", "  workflow_dispatch:\n  push:\n", 1)
			},
			wantErr: "workflow_dispatch must be the only trigger",
		},
		{
			name: "write all permission",
			mutate: func(value string) string {
				return strings.Replace(value, "permissions:\n  contents: read", "permissions: write-all", 1)
			},
			wantErr: "permissions must be exactly contents read",
		},
		{
			name: "non read permission",
			mutate: func(value string) string {
				return strings.Replace(value, "  contents: read", "  contents: write", 1)
			},
			wantErr: "permissions must be exactly contents read",
		},
		{
			name: "publish job",
			mutate: func(value string) string {
				return value + "\n  publish-release:\n    runs-on: ubuntu-latest\n    steps: []\n"
			},
			wantErr: "unexpected job",
		},
		{
			name: "release action",
			mutate: func(value string) string {
				return strings.Replace(value, "uses: actions/upload-artifact@v7", "uses: softprops/action-gh-release@v2", 1)
			},
			wantErr: "action is not allowed",
		},
		{
			name: "publish command",
			mutate: func(value string) string {
				return strings.Replace(value, "set -euo pipefail", "npm publish", 1)
			},
			wantErr: "forbidden release capability",
		},
		{
			name: "source input checkout",
			mutate: func(value string) string {
				return strings.ReplaceAll(value, "ref: ${{ github.sha }}", "ref: ${{ inputs.source_commit }}")
			},
			wantErr: "checkout must bind github.sha",
		},
		{
			name: "missing smoke field",
			mutate: func(value string) string {
				return strings.ReplaceAll(value, "provider_free_functional", "provider_free")
			},
			wantErr: "missing smoke evidence field",
		},
		{
			name: "filename mode checksum",
			mutate: func(value string) string {
				return strings.Replace(value, `sha256sum < "$1" | awk '{print $1}'`, `sha256sum "$1" | awk '{print $1}'`, 1)
			},
			wantErr: "missing smoke evidence field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReleaseRehearsalWorkflowStructure(tt.mutate(workflow))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestReleaseRehearsalCandidateVerifierAcceptsExactInventory(t *testing.T) {
	candidatesDir := t.TempDir()
	binaries := buildReleaseCandidateBinaries(t)
	for _, target := range rehearsalTargets {
		writeReleaseCandidateFixture(t, candidatesDir, target.Label, target, rehearsalSourceSHA, binaries[target.Label])
	}
	plan := runReleaseCandidateVerifier(t, candidatesDir, true, "")
	candidates, ok := plan["candidates"].([]any)
	if !ok || len(candidates) != len(rehearsalTargets) {
		t.Fatalf("promotion plan candidate inventory drifted: %#v", plan)
	}
	if plan["schema_version"] != "ao.atlas.release-rehearsal-promotion-plan.v0.4" {
		t.Fatalf("promotion plan schema drifted: %#v", plan)
	}
	for _, candidateValue := range candidates {
		candidate, ok := candidateValue.(map[string]any)
		if !ok || !strings.HasPrefix(fmt.Sprint(candidate["binary_sha256"]), "sha256:") {
			t.Fatalf("promotion plan must separately bind binary digest: %#v", candidateValue)
		}
		if candidate["candidate_identity"] != "ao-atlas version="+rehearsalVersion+" source_sha="+rehearsalSourceSHA {
			t.Fatalf("promotion plan must bind executed version identity: %#v", candidateValue)
		}
		if candidate["version_command"] != fmt.Sprint(candidate["binary"])+" --version" {
			t.Fatalf("promotion plan must bind exact version command: %#v", candidateValue)
		}
		if candidate["help_command"] != fmt.Sprint(candidate["binary"])+" (no arguments)" {
			t.Fatalf("promotion plan must bind exact help command: %#v", candidateValue)
		}
		if candidate["functional_smoke_command"] != "workgraph validate --workgraph examples/valid/workgraph.json" {
			t.Fatalf("promotion plan must bind exact functional smoke command: %#v", candidateValue)
		}
		candidateDir := filepath.Join(candidatesDir, fmt.Sprint(candidate["target_label"]))
		evidenceDigests := map[string]string{
			"checksum_manifest_sha256":        "SHA256SUMS",
			"help_readback_sha256":            "help-readback.json",
			"installed_archive_smoke_sha256":  "installed-archive-smoke.json",
			"provider_free_functional_sha256": "provider-free-functional-smoke.json",
			"provenance_sha256":               "provenance.json",
			"sbom_sha256":                     "sbom.spdx.json",
			"signature_verification_sha256":   "signature-verification.json",
			"version_source_readback_sha256":  "version-source-readback.json",
		}
		for field, file := range evidenceDigests {
			expected := "sha256:" + fileSHA256(t, filepath.Join(candidateDir, file))
			if candidate[field] != expected {
				t.Fatalf("promotion plan must exactly bind %s: got %v want %s", field, candidate[field], expected)
			}
		}
	}
}

func TestReleaseRehearsalCandidateVerifierRejectsNegativeFixtures(t *testing.T) {
	binaries := buildReleaseCandidateBinaries(t)
	tests := []struct {
		name    string
		mutate  func(*testing.T, string)
		wantErr string
	}{
		{
			name: "missing candidate",
			mutate: func(t *testing.T, dir string) {
				os.RemoveAll(filepath.Join(dir, "windows-x86_64"))
			},
			wantErr: "missing candidate",
		},
		{
			name: "duplicate candidate",
			mutate: func(t *testing.T, dir string) {
				writeReleaseCandidateFixture(t, dir, "linux-duplicate", rehearsalTargets[0], rehearsalSourceSHA, binaries["linux-x86_64"])
			},
			wantErr: "duplicate candidate",
		},
		{
			name: "stale candidate",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "macos-x86_64", "candidate-summary.json")
				updateFixtureJSON(t, path, func(value map[string]any) {
					value["source_sha"] = "cccccccccccccccccccccccccccccccccccccccc"
				})
				rewriteCandidateChecksums(t, filepath.Dir(path))
			},
			wantErr: "stale candidate",
		},
		{
			name: "wrong version",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "linux-x86_64", "candidate-summary.json")
				updateFixtureJSON(t, path, func(value map[string]any) {
					value["version"] = "v9.9.9"
				})
				rewriteCandidateChecksums(t, filepath.Dir(path))
			},
			wantErr: "stale candidate version",
		},
		{
			name: "altered manifest digest",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "windows-x86_64", "candidate-summary.json")
				updateFixtureJSON(t, path, func(value map[string]any) {
					value["approved_manifest_digest"] = "sha256:" + strings.Repeat("c", 64)
				})
				rewriteCandidateChecksums(t, filepath.Dir(path))
			},
			wantErr: "stale candidate manifest digest",
		},
		{
			name: "substituted candidate",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "linux-x86_64", "ao-atlas-v0.2.0-linux-x86_64.tar.gz")
				mustWriteFile(t, path, []byte("substituted archive\n"))
			},
			wantErr: "substituted candidate",
		},
		{
			name: "altered binary digest",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "linux-x86_64", "ao-atlas")
				mustWriteFile(t, path, []byte("substituted binary\n"))
			},
			wantErr: "substituted candidate binary",
		},
		{
			name: "substituted executable payload",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				binaryPath := filepath.Join(candidateDir, "ao-atlas")
				mustWriteFile(t, binaryPath, binaries["windows-x86_64"])
				rewriteChecksumSidecar(t, candidateDir, "ao-atlas")
				updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
					value["binary_sha256"] = "sha256:" + fileSHA256(t, binaryPath)
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "executable target substitution",
		},
		{
			name: "altered binary sidecar",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				sidecar := filepath.Join(candidateDir, "ao-atlas.exe.sha256")
				mustWriteFile(t, sidecar, []byte(strings.Repeat("d", 64)+"  ao-atlas.exe\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "binary checksum sidecar mismatch",
		},
		{
			name: "target architecture substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
					value["goarch"] = "arm64"
				})
				updateFixtureJSON(t, filepath.Join(candidateDir, "go-env.json"), func(value map[string]any) {
					value["GOARCH"] = "arm64"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "target architecture substitution",
		},
		{
			name: "executable target substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				oldBinary := filepath.Join(candidateDir, "ao-atlas.exe")
				newBinary := filepath.Join(candidateDir, "ao-atlas")
				if err := os.Rename(oldBinary, newBinary); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(oldBinary+".sha256", newBinary+".sha256"); err != nil {
					t.Fatal(err)
				}
				updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
					value["binary"] = "ao-atlas"
				})
				rewriteChecksumSidecar(t, candidateDir, "ao-atlas")
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "executable target substitution",
		},
		{
			name: "archive target substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				oldArchive := "ao-atlas-v0.2.0-linux-x86_64.tar.gz"
				newArchive := "ao-atlas-v0.2.0-linux-arm64.tar.gz"
				if err := os.Rename(filepath.Join(candidateDir, oldArchive), filepath.Join(candidateDir, newArchive)); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Join(candidateDir, oldArchive+".sha256"), filepath.Join(candidateDir, newArchive+".sha256")); err != nil {
					t.Fatal(err)
				}
				updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
					value["archive"] = newArchive
				})
				rewriteChecksumSidecar(t, candidateDir, newArchive)
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "archive target substitution",
		},
		{
			name: "substituted archive payload",
			mutate: func(t *testing.T, dir string) {
				linuxDir := filepath.Join(dir, "linux-x86_64")
				linuxArchive := filepath.Join(linuxDir, "ao-atlas-v0.2.0-linux-x86_64.tar.gz")
				macosArchive := filepath.Join(dir, "macos-x86_64", "ao-atlas-v0.2.0-macos-x86_64.tar.gz")
				content, err := os.ReadFile(macosArchive)
				if err != nil {
					t.Fatal(err)
				}
				mustWriteFile(t, linuxArchive, content)
				rewriteChecksumSidecar(t, linuxDir, filepath.Base(linuxArchive))
				updateFixtureJSON(t, filepath.Join(linuxDir, "candidate-summary.json"), func(value map[string]any) {
					value["archive_sha256"] = "sha256:" + fileSHA256(t, linuxArchive)
				})
				rewriteCandidateChecksums(t, linuxDir)
			},
			wantErr: "archive target substitution",
		},
		{
			name: "archive symlink entry",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[0])
				entries[1] = rehearsalArchiveEntry{
					header: tar.Header{Name: "LICENSE", Typeflag: tar.TypeSymlink, Linkname: "ao-atlas", Mode: 0o777},
				}
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[0], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "archive hard link entry",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[1])
				entries[1] = rehearsalArchiveEntry{
					header: tar.Header{Name: "LICENSE", Typeflag: tar.TypeLink, Linkname: "ao-atlas", Mode: 0o644},
				}
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[1], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "archive device entry",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[0])
				entries[1] = rehearsalArchiveEntry{
					header: tar.Header{
						Name: "LICENSE", Typeflag: tar.TypeChar, Mode: 0o600, Devmajor: 1, Devminor: 3,
					},
				}
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[0], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "archive nested file and directory",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[2])
				entries = append(entries,
					rehearsalArchiveEntry{
						header: tar.Header{Name: "nested/", Typeflag: tar.TypeDir, Mode: 0o755},
					},
					rehearsalArchiveEntry{
						header:  tar.Header{Name: "nested/extra.txt", Typeflag: tar.TypeReg, Mode: 0o644},
						content: []byte("unexpected\n"),
					},
				)
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[2], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "archive traversal entry",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[0])
				entries = append(entries, rehearsalArchiveEntry{
					header:  tar.Header{Name: "../outside.txt", Typeflag: tar.TypeReg, Mode: 0o644},
					content: []byte("unexpected\n"),
				})
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[0], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "archive absolute entry",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				entries := canonicalReleaseArchiveEntries(t, candidateDir, rehearsalTargets[1])
				entries = append(entries, rehearsalArchiveEntry{
					header:  tar.Header{Name: filepath.Join(candidateDir, "absolute.txt"), Typeflag: tar.TypeReg, Mode: 0o644},
					content: []byte("unexpected\n"),
				})
				replaceReleaseCandidateArchive(t, candidateDir, rehearsalTargets[1], entries)
			},
			wantErr: "noncanonical archive entry",
		},
		{
			name: "unexpected candidate file",
			mutate: func(t *testing.T, dir string) {
				mustWriteFile(t, filepath.Join(dir, "windows-x86_64", "unexpected.txt"), []byte("unexpected\n"))
			},
			wantErr: "unexpected candidate inventory",
		},
		{
			name: "unexpected candidate target",
			mutate: func(t *testing.T, dir string) {
				writeReleaseCandidateFixture(t, dir, "freebsd-x86_64", rehearsalTarget{
					Label: "freebsd-x86_64", GOOS: "freebsd", GOARCH: "amd64", Binary: "ao-atlas",
				}, rehearsalSourceSHA, binaries["linux-x86_64"])
			},
			wantErr: "unexpected candidate target",
		},
		{
			name: "missing checksum sidecar",
			mutate: func(t *testing.T, dir string) {
				os.Remove(filepath.Join(dir, "linux-x86_64", "ao-atlas-v0.2.0-linux-x86_64.tar.gz.sha256"))
			},
			wantErr: "missing checksum sidecar",
		},
		{
			name: "mismatched checksum sidecar",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				sidecar := filepath.Join(candidateDir, "ao-atlas-v0.2.0-macos-x86_64.tar.gz.sha256")
				mustWriteFile(t, sidecar, []byte(strings.Repeat("d", 64)+"  ao-atlas-v0.2.0-macos-x86_64.tar.gz\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "checksum sidecar mismatch",
		},
		{
			name: "noncanonical checksum sidecar",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				sidecar := filepath.Join(candidateDir, "ao-atlas-v0.2.0-windows-x86_64.tar.gz.sha256")
				content, err := os.ReadFile(sidecar)
				if err != nil {
					t.Fatal(err)
				}
				mustWriteFile(t, sidecar, append(content, '\n'))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "checksum sidecar mismatch",
		},
		{
			name: "malformed candidate JSON",
			mutate: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "linux-x86_64", "candidate-summary.json")
				mustWriteFile(t, path, []byte("{not-json\n"))
				rewriteCandidateChecksums(t, filepath.Dir(path))
			},
			wantErr: "malformed candidate JSON",
		},
		{
			name: "altered executed version identity",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "version-source-readback.json"), func(value map[string]any) {
					value["candidate_identity"] = "ao-atlas version=v0.2.0 source_sha=" + strings.Repeat("c", 40)
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid built candidate version identity",
		},
		{
			name: "altered version command",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "version-source-readback.json"), func(value map[string]any) {
					value["candidate_command"] = "ao-atlas.exe version"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid built candidate version identity",
		},
		{
			name: "false version tag match",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "version-source-readback.json"), func(value map[string]any) {
					value["tag_matches_version"] = false
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid version source readback evidence",
		},
		{
			name: "altered functional smoke command",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "provider-free-functional-smoke.json"), func(value map[string]any) {
					value["command"] = "workgraph validate --workgraph examples/invalid/workgraph.json"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid provider-free functional smoke evidence",
		},
		{
			name: "altered functional smoke fixture digest",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "provider-free-functional-smoke.json"), func(value map[string]any) {
					value["fixture_sha256"] = "sha256:" + strings.Repeat("d", 64)
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid provider-free functional smoke evidence",
		},
		{
			name: "help schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "linux-x86_64"), "help-readback.json")
			},
			wantErr: "invalid help readback evidence",
		},
		{
			name: "altered help command",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "help-readback.json"), func(value map[string]any) {
					value["command"] = "ao-atlas.exe --help"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid help readback evidence",
		},
		{
			name: "version source schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "macos-x86_64"), "version-source-readback.json")
			},
			wantErr: "invalid version source readback evidence",
		},
		{
			name: "provider smoke schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "windows-x86_64"), "provider-free-functional-smoke.json")
			},
			wantErr: "invalid provider-free functional smoke evidence",
		},
		{
			name: "malformed installed archive smoke",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				mustWriteFile(t, filepath.Join(candidateDir, "installed-archive-smoke.json"), []byte("{not-json\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid installed archive smoke evidence",
		},
		{
			name: "installed archive smoke substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "installed-archive-smoke.json"), func(value map[string]any) {
					value["binary"] = "other"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid installed archive smoke evidence",
		},
		{
			name: "installed archive schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "windows-x86_64"), "installed-archive-smoke.json")
			},
			wantErr: "invalid installed archive smoke evidence",
		},
		{
			name: "provenance source substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "provenance.json"), func(value map[string]any) {
					value["source_sha"] = strings.Repeat("d", 40)
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid provenance evidence",
		},
		{
			name: "provenance schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "linux-x86_64"), "provenance.json")
			},
			wantErr: "invalid provenance evidence",
		},
		{
			name: "malformed provenance",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				mustWriteFile(t, filepath.Join(candidateDir, "provenance.json"), []byte("{not-json\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid provenance evidence",
		},
		{
			name: "SBOM package substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "sbom.spdx.json"), func(value map[string]any) {
					value["packages"] = []any{map[string]any{"name": "other"}}
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid SBOM evidence",
		},
		{
			name: "malformed SBOM",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "windows-x86_64")
				mustWriteFile(t, filepath.Join(candidateDir, "sbom.spdx.json"), []byte("[]\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid SBOM evidence",
		},
		{
			name: "SBOM schema version substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "sbom.spdx.json"), func(value map[string]any) {
					value["spdxVersion"] = "SPDX-1.0"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid SBOM evidence",
		},
		{
			name: "signature status substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "signature-verification.json"), func(value map[string]any) {
					value["verification_status"] = "verified"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid signature verification evidence",
		},
		{
			name: "signature schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "windows-x86_64"), "signature-verification.json")
			},
			wantErr: "invalid signature verification evidence",
		},
		{
			name: "malformed signature status",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "macos-x86_64")
				mustWriteFile(t, filepath.Join(candidateDir, "signature-verification.json"), []byte("null\n"))
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid signature verification evidence",
		},
		{
			name: "candidate summary schema substitution",
			mutate: func(t *testing.T, dir string) {
				substituteEvidenceSchema(t, filepath.Join(dir, "macos-x86_64"), "candidate-summary.json")
			},
			wantErr: "invalid candidate summary",
		},
		{
			name: "candidate summary field substitution",
			mutate: func(t *testing.T, dir string) {
				candidateDir := filepath.Join(dir, "linux-x86_64")
				updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
					value["entry_point"] = "other"
				})
				rewriteCandidateChecksums(t, candidateDir)
			},
			wantErr: "invalid candidate summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidatesDir := t.TempDir()
			for _, target := range rehearsalTargets {
				writeReleaseCandidateFixture(t, candidatesDir, target.Label, target, rehearsalSourceSHA, binaries[target.Label])
			}
			tt.mutate(t, candidatesDir)
			runReleaseCandidateVerifier(t, candidatesDir, false, tt.wantErr)
		})
	}
}

func readReleaseRehearsalWorkflow(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-rehearsal.yml"))
	if err != nil {
		t.Fatalf("read release rehearsal workflow: %v", err)
	}
	return string(content)
}

func validateReleaseRehearsalWorkflowStructure(workflow string) error {
	triggerKeys := yamlChildKeys(yamlTopLevelSection(workflow, "on:"), 2)
	if len(triggerKeys) != 1 || triggerKeys[0] != "workflow_dispatch" {
		return fmt.Errorf("workflow_dispatch must be the only trigger: %v", triggerKeys)
	}

	permissions := strings.TrimSpace(yamlTopLevelSection(workflow, "permissions:"))
	if permissions != "permissions:\n  contents: read" {
		return fmt.Errorf("permissions must be exactly contents read")
	}

	jobsSection := yamlTopLevelSection(workflow, "jobs:")
	wantJobs := []string{"assemble-promotion-plan", "build-native-candidate", "validate-release-inputs"}
	gotJobs := yamlChildKeys(jobsSection, 2)
	sort.Strings(gotJobs)
	if strings.Join(gotJobs, ",") != strings.Join(wantJobs, ",") {
		return fmt.Errorf("unexpected job inventory: %v", gotJobs)
	}
	if strings.Contains(jobsSection, "\n    permissions:") || strings.Contains(jobsSection, "\n    environment:") {
		return fmt.Errorf("jobs must not override permissions or declare an environment")
	}

	allowedActions := map[string]bool{
		"actions/checkout@v7":          true,
		"actions/download-artifact@v7": true,
		"actions/setup-go@v6":          true,
		"actions/upload-artifact@v7":   true,
	}
	actionPattern := regexp.MustCompile(`(?m)^\s+uses:\s+(\S+)\s*$`)
	for _, match := range actionPattern.FindAllStringSubmatch(workflow, -1) {
		if !allowedActions[match[1]] {
			return fmt.Errorf("action is not allowed: %s", match[1])
		}
	}

	checkoutCount := strings.Count(workflow, "uses: actions/checkout@v7")
	if checkoutCount < 3 || strings.Count(workflow, "ref: ${{ github.sha }}") != checkoutCount {
		return fmt.Errorf("checkout must bind github.sha exactly")
	}
	if strings.Contains(workflow, "inputs.source_commit") ||
		strings.Count(workflow, "SOURCE_SHA: ${{ github.sha }}") < 3 ||
		!strings.Contains(workflow, `test "$actual_source_sha" = "$SOURCE_SHA"`) {
		return fmt.Errorf("source binding must use github.sha dispatch head")
	}

	for _, field := range []string{
		"help_readback",
		"version_source_readback",
		"provider_free_functional",
		"installed_archive_smoke",
		"help-readback.json",
		"version-source-readback.json",
		"provider-free-functional-smoke.json",
		"macos-15-intel",
		"target_label: macos-x86_64",
		"goos: darwin",
		"goarch: amd64",
		"-X github.com/uesugitorachiyo/ao-atlas/internal/atlas.buildVersion=$VERSION",
		"-X github.com/uesugitorachiyo/ao-atlas/internal/atlas.buildSourceSHA=$SOURCE_SHA",
		`"$candidate_dir/$binary" --version`,
		"binary_sha256",
		`sha256sum < "$1" | awk '{print $1}'`,
		`shasum -a 256 < "$1" | awk '{print $1}'`,
		`printf '%s  %s\n' "$(hash_file "$checksum_file")" "$checksum_file"`,
	} {
		if !strings.Contains(workflow, field) {
			return fmt.Errorf("missing smoke evidence field %q", field)
		}
	}

	for _, forbidden := range []string{
		"contents: write",
		"write-all",
		"gh release",
		"git tag",
		"git push",
		"actions/create-release",
		"softprops/action-gh-release",
		"actions/upload-release-asset",
		"npm publish",
		"docker push",
		"secrets.",
	} {
		if strings.Contains(workflow, forbidden) {
			return fmt.Errorf("forbidden release capability %q", forbidden)
		}
	}
	return nil
}

func yamlTopLevelSection(workflow, header string) string {
	lines := strings.Split(workflow, "\n")
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
		if lines[i] != "" && lines[i][0] != ' ' && lines[i][0] != '#' {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func yamlChildKeys(section string, indent int) []string {
	pattern := regexp.MustCompile(fmt.Sprintf(`^ {%d}([A-Za-z0-9_-]+):`, indent))
	var keys []string
	for _, line := range strings.Split(section, "\n") {
		if match := pattern.FindStringSubmatch(line); match != nil {
			keys = append(keys, match[1])
		}
	}
	return keys
}

func runReleaseCandidateVerifier(t *testing.T, candidatesDir string, wantSuccess bool, wantOutput string) map[string]any {
	t.Helper()
	planPath := filepath.Join(t.TempDir(), "promotion-plan.json")
	script := filepath.Join(repoRoot(t), "scripts", "verify-release-rehearsal-candidates.sh")
	cmd := exec.Command("bash", script,
		"--candidates-dir", candidatesDir,
		"--version", rehearsalVersion,
		"--tag", rehearsalVersion,
		"--source-sha", rehearsalSourceSHA,
		"--approved-manifest-digest", rehearsalManifestDigest,
		"--plan-out", planPath,
	)
	output, err := cmd.CombinedOutput()
	if wantSuccess {
		if err != nil {
			t.Fatalf("candidate verifier failed: %v\n%s", err, output)
		}
		content, readErr := os.ReadFile(planPath)
		if readErr != nil {
			t.Fatalf("read promotion plan: %v", readErr)
		}
		if !strings.Contains(string(content), `"status":"dry_run_plan_ready"`) {
			t.Fatalf("promotion plan missing ready status: %s", content)
		}
		var plan map[string]any
		if err := json.Unmarshal(content, &plan); err != nil {
			t.Fatalf("parse promotion plan: %v", err)
		}
		return plan
	}
	if err == nil {
		t.Fatalf("candidate verifier accepted negative fixture")
	}
	if !strings.Contains(string(output), wantOutput) {
		t.Fatalf("candidate verifier error missing %q: %s", wantOutput, output)
	}
	return nil
}

func buildReleaseCandidateBinaries(t *testing.T) map[string][]byte {
	t.Helper()
	root := repoRoot(t)
	outputDir := t.TempDir()
	binaries := make(map[string][]byte, len(rehearsalTargets))
	for _, target := range rehearsalTargets {
		outputPath := filepath.Join(outputDir, target.Label)
		ldflags := fmt.Sprintf(
			"-X github.com/uesugitorachiyo/ao-atlas/internal/atlas.buildVersion=%s -X github.com/uesugitorachiyo/ao-atlas/internal/atlas.buildSourceSHA=%s",
			rehearsalVersion,
			rehearsalSourceSHA,
		)
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags", ldflags, "-o", outputPath, "./cmd/atlas")
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"CGO_ENABLED=0",
			"GOOS="+target.GOOS,
			"GOARCH="+target.GOARCH,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s fixture binary: %v\n%s", target.Label, err, output)
		}
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatal(err)
		}
		binaries[target.Label] = content
	}
	return binaries
}

func writeReleaseCandidateArchive(t *testing.T, path, binary string, binaryContent []byte) {
	t.Helper()
	writeReleaseCandidateArchiveEntries(t, path, []rehearsalArchiveEntry{
		{
			header:  tar.Header{Name: binary, Typeflag: tar.TypeReg, Mode: 0o755},
			content: binaryContent,
		},
		{
			header:  tar.Header{Name: "LICENSE", Typeflag: tar.TypeReg, Mode: 0o644},
			content: []byte("license\n"),
		},
	})
}

type rehearsalArchiveEntry struct {
	header  tar.Header
	content []byte
}

func writeReleaseCandidateArchiveEntries(t *testing.T, path string, entries []rehearsalArchiveEntry) {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		header := entry.header
		header.Size = int64(len(entry.content))
		if err := tarWriter.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if len(entry.content) > 0 {
			if _, err := tarWriter.Write(entry.content); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, path, compressed.Bytes())
}

func canonicalReleaseArchiveEntries(t *testing.T, candidateDir string, target rehearsalTarget) []rehearsalArchiveEntry {
	t.Helper()
	binaryContent, err := os.ReadFile(filepath.Join(candidateDir, target.Binary))
	if err != nil {
		t.Fatal(err)
	}
	return []rehearsalArchiveEntry{
		{
			header:  tar.Header{Name: target.Binary, Typeflag: tar.TypeReg, Mode: 0o755},
			content: binaryContent,
		},
		{
			header:  tar.Header{Name: "LICENSE", Typeflag: tar.TypeReg, Mode: 0o644},
			content: []byte("license\n"),
		},
	}
}

func replaceReleaseCandidateArchive(t *testing.T, candidateDir string, target rehearsalTarget, entries []rehearsalArchiveEntry) {
	t.Helper()
	archive := fmt.Sprintf("ao-atlas-%s-%s.tar.gz", rehearsalVersion, target.Label)
	archivePath := filepath.Join(candidateDir, archive)
	writeReleaseCandidateArchiveEntries(t, archivePath, entries)
	rewriteChecksumSidecar(t, candidateDir, archive)
	updateFixtureJSON(t, filepath.Join(candidateDir, "candidate-summary.json"), func(value map[string]any) {
		value["archive_sha256"] = "sha256:" + fileSHA256(t, archivePath)
	})
	rewriteCandidateChecksums(t, candidateDir)
}

func writeReleaseCandidateFixture(t *testing.T, root, directoryName string, target rehearsalTarget, sourceSHA string, binaryContent []byte) {
	t.Helper()
	candidateDir := filepath.Join(root, directoryName)
	if err := os.MkdirAll(candidateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binary := target.Binary
	archive := fmt.Sprintf("ao-atlas-%s-%s.tar.gz", rehearsalVersion, target.Label)
	mustWriteExecutable(t, filepath.Join(candidateDir, binary), binaryContent)
	binaryDigest := fileSHA256(t, filepath.Join(candidateDir, binary))
	mustWriteFile(t, filepath.Join(candidateDir, binary+".sha256"), []byte(binaryDigest+"  "+binary+"\n"))
	writeReleaseCandidateArchive(t, filepath.Join(candidateDir, archive), binary, binaryContent)
	archiveDigest := fileSHA256(t, filepath.Join(candidateDir, archive))
	mustWriteFile(t, filepath.Join(candidateDir, archive+".sha256"), []byte(archiveDigest+"  "+archive+"\n"))
	workflowIdentity := "https://github.com/uesugitorachiyo/ao-atlas/actions/runs/12345"
	functionalFixture := "examples/valid/workgraph.json"
	functionalFixtureDigest := fileSHA256(t, filepath.Join(repoRoot(t), functionalFixture))
	helpOutput := "atlas <instance|intake|blueprint|mission|blueprint-request|workgraph|mutation-classes|factory-task|factory|context-pack|foundry|run-link> ..."
	files := map[string][]byte{
		"LICENSE":              []byte("license\n"),
		"archive-identity.txt": []byte(binary + "\nLICENSE\n"),
		"go-env.json":          []byte(fmt.Sprintf("{\"GOOS\":%q,\"GOARCH\":%q}\n", target.GOOS, target.GOARCH)),
		"help-readback.json": []byte(fmt.Sprintf(
			"{\"schema_version\":\"ao.atlas.release-rehearsal-help-readback.v0.1\",\"status\":\"passed\",\"classification\":\"help_readback_not_functional_smoke\",\"documented_reference\":\"internal/atlas/cli.go:usage\",\"command\":%q,\"expected_exit_code\":2,\"actual_exit_code\":2,\"output\":%q}\n",
			binary+" (no arguments)", helpOutput,
		)),
		"installed-archive-smoke.json": []byte(fmt.Sprintf(
			"{\"schema_version\":\"ao.atlas.release-rehearsal-installed-archive-smoke.v0.1\",\"status\":\"passed\",\"archive\":%q,\"binary\":%q,\"license_present\":true}\n",
			archive, binary,
		)),
		"provider-free-functional-smoke.json": []byte(fmt.Sprintf(
			"{\"schema_version\":\"ao.atlas.release-rehearsal-provider-free-functional-smoke.v0.1\",\"status\":\"passed\",\"command\":\"workgraph validate --workgraph examples/valid/workgraph.json\",\"fixture\":%q,\"fixture_sha256\":\"sha256:%s\",\"installed_archive\":true,\"provider_credentials_used\":false}\n",
			functionalFixture, functionalFixtureDigest,
		)),
		"provenance.json": []byte(fmt.Sprintf(
			"{\"schema_version\":\"ao.atlas.release-rehearsal-provenance.v0.2\",\"builder\":\"github-actions\",\"source_sha\":%q,\"target_label\":%q,\"workflow_identity\":%q}\n",
			sourceSHA, target.Label, workflowIdentity,
		)),
		"sbom.spdx.json": []byte(fmt.Sprintf(
			"{\"SPDXID\":\"SPDXRef-DOCUMENT\",\"spdxVersion\":\"SPDX-2.3\",\"name\":\"ao-atlas release rehearsal candidate\",\"documentNamespace\":\"https://github.com/uesugitorachiyo/ao-atlas/rehearsal/%s/%s\",\"creationInfo\":{\"creators\":[\"Tool: AO Atlas release rehearsal\"],\"created\":\"1970-01-01T00:00:00Z\"},\"packages\":[{\"SPDXID\":\"SPDXRef-Package\",\"name\":\"ao-atlas\",\"versionInfo\":\"rehearsal\",\"downloadLocation\":\"NOASSERTION\",\"filesAnalyzed\":false}]}\n",
			sourceSHA, target.Label,
		)),
		"signature-verification.json": []byte(`{"schema_version":"ao.atlas.release-rehearsal-signature-verification.v0.1","signature_present":false,"verification_status":"not_applicable_rehearsal_no_signing_key"}` + "\n"),
		"version-source-readback.json": []byte(fmt.Sprintf(
			"{\"schema_version\":\"ao.atlas.release-rehearsal-version-source-readback.v0.2\",\"status\":\"passed\",\"source\":\"workflow_dispatch.inputs.version\",\"repository_version_source\":\"go.mod\",\"module_path\":\"github.com/uesugitorachiyo/ao-atlas\",\"go_directive\":\"1.22\",\"version\":%q,\"source_sha\":%q,\"tag\":%q,\"tag_matches_version\":true,\"candidate_command\":%q,\"candidate_identity\":%q}\n",
			rehearsalVersion, sourceSHA, rehearsalVersion, binary+" --version", "ao-atlas version="+rehearsalVersion+" source_sha="+sourceSHA,
		)),
	}
	for name, content := range files {
		mustWriteFile(t, filepath.Join(candidateDir, name), content)
	}
	summary := map[string]any{
		"approved_manifest_digest":  rehearsalManifestDigest,
		"archive":                   archive,
		"archive_sha256":            "sha256:" + archiveDigest,
		"binary":                    binary,
		"binary_sha256":             "sha256:" + binaryDigest,
		"goarch":                    target.GOARCH,
		"goos":                      target.GOOS,
		"go_version":                "go version fixture",
		"help_readback":             "help-readback.json",
		"immutable":                 true,
		"installed_archive_smoke":   "installed-archive-smoke.json",
		"license_files":             []string{"LICENSE"},
		"provider_free_functional":  "provider-free-functional-smoke.json",
		"provider_credentials_used": false,
		"provenance":                "provenance.json",
		"repository":                "ao-atlas",
		"runner_os":                 target.GOOS,
		"sbom":                      "sbom.spdx.json",
		"schema_version":            "ao.atlas.release-rehearsal-candidate.v0.3",
		"signature_verification":    "signature-verification.json",
		"source_sha":                sourceSHA,
		"target_label":              target.Label,
		"tag":                       rehearsalVersion,
		"version":                   rehearsalVersion,
		"version_source_readback":   "version-source-readback.json",
		"workflow_identity":         workflowIdentity,
		"entry_point":               "ao-atlas",
	}
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(candidateDir, "candidate-summary.json"), append(summaryBytes, '\n'))
	rewriteCandidateChecksums(t, candidateDir)
}

func substituteEvidenceSchema(t *testing.T, candidateDir, name string) {
	t.Helper()
	updateFixtureJSON(t, filepath.Join(candidateDir, name), func(value map[string]any) {
		value["schema_version"] = "substituted.v9"
	})
	rewriteCandidateChecksums(t, candidateDir)
}

func rewriteCandidateChecksums(t *testing.T, candidateDir string) {
	t.Helper()
	entries, err := os.ReadDir(candidateDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "SHA256SUMS" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	var manifest strings.Builder
	for _, name := range names {
		fmt.Fprintf(&manifest, "%s  %s\n", fileSHA256(t, filepath.Join(candidateDir, name)), name)
	}
	mustWriteFile(t, filepath.Join(candidateDir, "SHA256SUMS"), []byte(manifest.String()))
}

func rewriteChecksumSidecar(t *testing.T, candidateDir, name string) {
	t.Helper()
	digest := fileSHA256(t, filepath.Join(candidateDir, name))
	mustWriteFile(t, filepath.Join(candidateDir, name+".sha256"), []byte(digest+"  "+name+"\n"))
}

func updateFixtureJSON(t *testing.T, path string, update func(map[string]any)) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(content, &value); err != nil {
		t.Fatal(err)
	}
	update(value)
	updated, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, path, append(updated, '\n'))
}

func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustWriteExecutable(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatal(err)
	}
}
