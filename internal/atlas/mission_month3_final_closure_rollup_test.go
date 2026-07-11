package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3FinalClosureRollupBindsTerminalReadbacks(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-02-aggregate-rollup")
	recordedPath := filepath.Join(nodeDir, "month3-final-closure-rollup.json")
	outPath := filepath.Join(t.TempDir(), "month3-final-closure-rollup.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-final-closure-rollup",
		"--node-id", "mission-recommendation-month3-final-closure-02-aggregate-rollup",
		"--source-readback", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "recommendation-readback-after.json"),
		"--readiness-matrix", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "golden-path-readiness-matrix.json"),
		"--promoter", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "promoter_no_promotion.json"),
		"--command", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "command_readback.json"),
		"--public-safety", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "scoped-public-safety-scan.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-final-closure-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3FinalClosureRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3FinalClosureRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 final closure rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3FinalClosureRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "final_closure_rollup_bound" ||
		recorded.SourceCompletedNodes != 40 ||
		recorded.SourceReadyNodes != 0 ||
		recorded.SourceBlockedNodes != 0 ||
		recorded.SourceFailedNodes != 0 ||
		!recorded.SourceFinalResponseAllowed ||
		recorded.MatrixRecommendationCount != 40 ||
		recorded.PromoterStatus != "no_promotion_requested" ||
		recorded.CommandStatus != "readback_agrees_no_promotion" ||
		recorded.PublicSafetyStatus != "passed" ||
		!recorded.PublicSafetyScanPassed ||
		recorded.AggregatePromotionStatus != "no_promotion_requested" ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("Month 3 final closure rollup lost terminal no-promotion state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-final-closure-rollup.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-final-closure-rollup" {
		t.Fatalf("expected typed Month 3 final closure rollup validator, got %s", validator)
	}
}
