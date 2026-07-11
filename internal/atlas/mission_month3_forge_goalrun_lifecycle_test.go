package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosureForgeGoalRunLifecycleFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-26-forge-goalrun-lifecycle")
	recordedPath := filepath.Join(nodeDir, "forge-goalrun-evidence-fixture.json")
	outPath := filepath.Join(t.TempDir(), "forge-goalrun-evidence-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "forge-goalrun-evidence-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("forge-goalrun-evidence-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=forge_goalrun_evidence_ready",
		"goalrun_start_required=true",
		"provider_execution_allowed=false",
		"terminal_receipt_required=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("Forge GoalRun evidence output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[AtlasForgeGoalRunEvidenceFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasForgeGoalRunEvidenceFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Forge GoalRun lifecycle fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if !recorded.GoalRunStartRequired ||
		!recorded.StopGateRequired ||
		!recorded.RollbackRecordRequired ||
		!recorded.TerminalReceiptRequired ||
		recorded.ProviderExecutionAllowed ||
		recorded.ExecutesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Forge GoalRun lifecycle fixture lost bounded lifecycle state: %#v", recorded)
	}
}

func TestMonth3FinalClosureForgeGoalRunLifecycleFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-26-forge-goalrun-lifecycle", "forge-goalrun-evidence-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.forge-goalrun-evidence-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:forge-goalrun-evidence-fixture" {
		t.Fatalf("expected typed Forge GoalRun evidence validator, got %s", validator)
	}
}
