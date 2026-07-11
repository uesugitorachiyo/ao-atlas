package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3CrossRepoCIMatrixBindsSixReposToSentinelStates(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-09-cross-repo-ci-matrix")
	recordedPath := filepath.Join(nodeDir, "month3-cross-repo-ci-matrix.json")
	outPath := filepath.Join(t.TempDir(), "month3-cross-repo-ci-matrix.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-cross-repo-ci-matrix",
		"--node-id", "mission-recommendation-month3-final-closure-09-cross-repo-ci-matrix",
		"--sentinel-signal-state", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-16", "sentinel-signal-state-fixture.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-cross-repo-ci-matrix command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3CrossRepoCIMatrix](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3CrossRepoCIMatrix](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 cross-repo CI matrix changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3CrossRepoCIMatrix(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "cross_repo_ci_matrix_ready" ||
		recorded.RepoCount != 6 ||
		recorded.MatrixEntryCount != 24 ||
		!recorded.SentinelSignalStateBound ||
		!recorded.RequiresPassBeforeMerge ||
		!recorded.BlocksOnFailure ||
		!recorded.WaitsOnPending ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("cross-repo CI matrix lost safety state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-cross-repo-ci-matrix.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-cross-repo-ci-matrix" {
		t.Fatalf("expected typed Month 3 cross-repo CI matrix validator, got %s", validator)
	}
}
