package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecialistReleaseRehearsalWorkflowContract(t *testing.T) {
	root := repoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release-rehearsal.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read release rehearsal workflow: %v", err)
	}
	workflow := string(content)

	for _, want := range []string{
		"name: AO Atlas Specialist Release Rehearsal",
		"workflow_dispatch:",
		"version:",
		"tag:",
		"source_commit:",
		"approved_manifest_digest:",
		"contents: read",
		"validate-release-inputs:",
		"build-native-candidate:",
		"assemble-promotion-plan:",
		"ubuntu-latest",
		"macos-latest",
		"windows-latest",
		"linux-x86_64",
		"macos-aarch64",
		"windows-x86_64",
		"actions/upload-artifact@v7",
		"candidate-summary.json",
		"archive-identity.txt",
		"SHA256SUMS",
		"sha256sum --check --strict SHA256SUMS",
		"command -v sha256sum",
		"shasum -a 256",
		"sbom.spdx.json",
		"provenance.json",
		"signature-verification.json",
		"provider-free-smoke.txt",
		"installed-archive-smoke.txt",
		"promotion-plan.json",
		"promotion-plan.sha256",
		"dry-run-boundary.json",
		"publication_attempted\":false",
		"tag_creation_attempted\":false",
		"release_creation_attempted\":false",
		"public_upload_attempted\":false",
		"expected_inventory",
		"missing, duplicate, stale, substituted, or unexpected",
		"persist-credentials: false",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release rehearsal workflow missing %q", want)
		}
	}

	for _, forbidden := range []string{
		"contents: write",
		"gh release",
		"git tag",
		"git push",
		"actions/create-release",
		"softprops/action-gh-release",
		"actions/upload-release-asset",
		"environment:",
		"secrets.",
		"OPENAI_" + "API_KEY",
		"ANTHROPIC_" + "API_KEY",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("release rehearsal workflow contains forbidden capability %q", forbidden)
		}
	}
}
