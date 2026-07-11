package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3OperatorDashboardReadbackSummarizesActiveGoldenPathStatus(t *testing.T) {
	root := repoRoot(t)
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-10-operator-dashboard-readback")
	recordedPath := filepath.Join(nodeDir, "month3-operator-dashboard-readback.json")
	outPath := filepath.Join(t.TempDir(), "month3-operator-dashboard-readback.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-operator-dashboard-readback",
		"--node-id", "mission-recommendation-month3-final-closure-10-operator-dashboard-readback",
		"--source-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-09-cross-repo-ci-matrix", "recommendation-readback-after.json"),
		"--ci-matrix", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-09-cross-repo-ci-matrix", "month3-cross-repo-ci-matrix.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-operator-dashboard-readback command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3OperatorDashboardReadback](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3OperatorDashboardReadback](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 operator dashboard readback changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3OperatorDashboardReadback(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "operator_dashboard_active" ||
		recorded.CompletedNodes != 9 ||
		recorded.ReadyNodes != 21 ||
		recorded.BlockedNodes != 0 ||
		recorded.FailedNodes != 0 ||
		recorded.FirstExecutableNode != "mission-recommendation-month3-final-closure-10-operator-dashboard-readback" ||
		recorded.BlockerCount != 0 ||
		!recorded.ReadyWorkVisible ||
		!recorded.CIMatrixBound ||
		!recorded.RequiresPassBeforeMerge ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("operator dashboard readback lost active safe state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-operator-dashboard-readback.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-operator-dashboard-readback" {
		t.Fatalf("expected typed Month 3 operator dashboard readback validator, got %s", validator)
	}
}
