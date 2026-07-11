package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3RollbackReplayNegativeRejectsStaleBaseAndDigestMismatch(t *testing.T) {
	root := repoRoot(t)
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-13-rollback-replay-negative")
	recordedPath := filepath.Join(nodeDir, "month3-rollback-replay-negative.json")
	outPath := filepath.Join(t.TempDir(), "month3-rollback-replay-negative.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-rollback-replay-negative",
		"--node-id", "mission-recommendation-month3-final-closure-13-rollback-replay-negative",
		"--source-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-12-provider-model-provenance", "recommendation-readback-after.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-rollback-replay-negative command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3RollbackReplayNegative](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3RollbackReplayNegative](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 rollback replay negative fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3RollbackReplayNegative(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "rollback_replay_negative_ready" ||
		recorded.CaseCount != 2 ||
		!recorded.StaleBaseCommitRejected ||
		!recorded.ReceiptDigestMismatchRejected ||
		recorded.AcceptedCaseCount != 0 ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("rollback replay negative fixture lost denial state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-rollback-replay-negative.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-rollback-replay-negative" {
		t.Fatalf("expected typed Month 3 rollback replay negative validator, got %s", validator)
	}
}
