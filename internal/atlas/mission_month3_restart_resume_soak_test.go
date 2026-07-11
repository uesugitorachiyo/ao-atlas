package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3RestartResumeSoakBindsCheckpointRecoveryFixtures(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-11-restart-resume-soak")
	recordedPath := filepath.Join(nodeDir, "month3-restart-resume-soak.json")
	outPath := filepath.Join(t.TempDir(), "month3-restart-resume-soak.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-restart-resume-soak",
		"--node-id", "mission-recommendation-month3-final-closure-11-restart-resume-soak",
		"--exactly-once", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-24", "exactly-once-resume-accounting-fixture.json"),
		"--kill-restart", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-38", "kill-restart-replay-fixture.json"),
		"--source-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-10-operator-dashboard-readback", "recommendation-readback-after.json"),
		"--dashboard-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-10-operator-dashboard-readback", "month3-operator-dashboard-readback.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-restart-resume-soak command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3RestartResumeSoak](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3RestartResumeSoak](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 restart resume soak fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3RestartResumeSoak(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "restart_resume_soak_ready" ||
		recorded.ScenarioCount != 4 ||
		!recorded.ExactlyOnceAccountingBound ||
		!recorded.KillRestartReplayBound ||
		!recorded.CheckpointRecoveryBound ||
		!recorded.NoLostEvidence ||
		recorded.DuplicateMutationDetected ||
		recorded.FalseCompletionDetected ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("restart resume soak lost safe recovery state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-restart-resume-soak.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-restart-resume-soak" {
		t.Fatalf("expected typed Month 3 restart resume soak validator, got %s", validator)
	}
}
