package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3FinalReadinessReportBindsCapabilitiesAndBlockers(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-30-final-report")
	recordedPath := filepath.Join(nodeDir, "month3-final-readiness-report.json")
	outPath := filepath.Join(t.TempDir(), "month3-final-readiness-report.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-final-readiness-report",
		"--node-id", "mission-recommendation-month3-final-closure-30-final-report",
		"--source-readback", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "recommendation-readback-after.json"),
		"--readiness-matrix", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "golden-path-readiness-matrix.json"),
		"--closure-rollup", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-02-aggregate-rollup", "month3-final-closure-rollup.json"),
		"--closure-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-29-public-safety-claims", "recommendation-readback-after.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-final-readiness-report command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3FinalReadinessReport](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3FinalReadinessReport](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 final readiness report fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3FinalReadinessReport(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "ready_for_operator_handoff" ||
		recorded.SourceCompletedNodes != 40 ||
		recorded.SourceReadyNodes != 0 ||
		recorded.SourceBlockedNodes != 0 ||
		recorded.ClosureCompletedNodesBeforeReport != 29 ||
		recorded.ClosureReadyNodesBeforeReport != 1 ||
		recorded.ClosureNextExecutableNode != "mission-recommendation-month3-final-closure-30-final-report" ||
		recorded.ProvenCapabilityCount == 0 ||
		recorded.UnresolvedBlockerCount == 0 ||
		recorded.RecommendedNextActionCount < 10 ||
		recorded.AggregatePromotionStatus != "no_promotion_requested" ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("Month 3 final readiness report lost terminal handoff boundaries: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-final-readiness-report.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-final-readiness-report" {
		t.Fatalf("expected typed Month 3 final readiness report validator, got %s", validator)
	}
}
