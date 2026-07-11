package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3WorkspaceRootPreflightFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-19")
	recordedPath := filepath.Join(nodeDir, "workspace-root-preflight-fixture.json")
	outPath := filepath.Join(t.TempDir(), "workspace-root-preflight-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "workspace-root-preflight-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("workspace-root-preflight-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=preflight_ready",
		"repository_identity_validated=true",
		"objective_digest_validated=true",
		"worktree_boundary_validated=true",
		"safe_next_node_selected=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("workspace preflight output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("workspace-root preflight fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["non_ao_repository"] != true ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("workspace-root preflight fixture lost non-AO or authority state: %#v", generated)
	}
}

func TestMonth3WorkspaceRootPreflightFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-19", "workspace-root-preflight-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.workspace-root-preflight-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:workspace-root-preflight-fixture" {
		t.Fatalf("expected typed workspace-root preflight validator, got %s", validator)
	}
}

func TestMonth3FinalClosureWorkspaceRootPreflightFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-17-workspace-root-preflight")
	recordedPath := filepath.Join(nodeDir, "workspace-root-preflight-fixture.json")
	outPath := filepath.Join(t.TempDir(), "workspace-root-preflight-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "workspace-root-preflight-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("workspace-root-preflight-fixture command failed: %s", out.String())
	}

	recorded := mustLoadJSON[AtlasWorkspaceRootPreflightFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasWorkspaceRootPreflightFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("final-closure workspace-root preflight fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasWorkspaceRootPreflightFixture(recorded); err != nil {
		t.Fatalf("recorded final-closure workspace-root preflight fixture invalid: %v", err)
	}
	if recorded.RepositoryIdentity != "tiny-non-ao-repository" ||
		!recorded.NonAORepository ||
		!recorded.WorktreeBoundaryValidated ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("final-closure workspace preflight fixture lost dry-run boundary: %#v", recorded)
	}
}

func TestMonth3WorkspaceRootPreflightFixtureRejectsMissingWorktreeBoundary(t *testing.T) {
	fixture, err := BuildAtlasWorkspaceRootPreflightFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.WorktreeBoundaryValidated = false
	if err := ValidateAtlasWorkspaceRootPreflightFixture(fixture); err == nil || !strings.Contains(err.Error(), "worktree_boundary_validated must be true") {
		t.Fatalf("expected missing worktree boundary rejection, got %v", err)
	}
}
