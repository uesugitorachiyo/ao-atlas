package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWavePromoterNoPromotionRollupAggregatesCompletedWaves(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	hardeningRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-25")
	sourceReadbackPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-24", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "promoter-no-promotion-rollup.json")
	outPath := filepath.Join(t.TempDir(), "promoter-no-promotion-rollup.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "promoter-no-promotion-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-25",
		"--source-readback", sourceReadbackPath,
		"--evidence-root", hardeningRoot,
		"--evidence-root", closureRoot,
		"--evidence-root", featureRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("promoter-no-promotion-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPromoterNoPromotionRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasPromoterNoPromotionRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("promoter no-promotion rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPromoterNoPromotionRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "no_promotion_rollup_bound" ||
		recorded.CompletedNodesTotal != 88 ||
		recorded.PromoterNoPromotionFiles != 88 ||
		recorded.MissingPromoterNodesTotal != 0 ||
		recorded.NoPromotionStatusCount != 88 ||
		recorded.PromotionRequestedCount != 0 ||
		recorded.PromotionGrantedCount != 0 ||
		recorded.PromotionClaimedCount != 0 ||
		recorded.AuthorityAdvanceClaimCount != 0 ||
		recorded.RSIDeniedCount != 88 ||
		recorded.AggregatePromotionStatus != "no_promotion_requested" ||
		!recorded.AllCompletedNodesCovered ||
		!recorded.AllPromoterStatusesNoPromotion ||
		!recorded.NoPromotionInvariantHolds ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("promoter rollup lost no-promotion/no-RSI invariants: %#v", recorded)
	}
	if len(recorded.WaveSummaries) != 3 {
		t.Fatalf("expected hardening, closure, and feature-depth wave summaries: %#v", recorded.WaveSummaries)
	}
	expectedCompleted := []int{40, 24, 24}
	for i, summary := range recorded.WaveSummaries {
		if summary.CompletedNodes != expectedCompleted[i] ||
			summary.PromoterNoPromotionFiles != expectedCompleted[i] ||
			summary.MissingPromoterNodes != 0 ||
			summary.PromotionRequestedCount != 0 ||
			summary.PromotionGrantedCount != 0 ||
			summary.AuthorityAdvanceClaimCount != 0 ||
			!summary.NoPromotionInvariantHolds ||
			!summary.RSIRemainsDenied {
			t.Fatalf("wave summary %d lost no-promotion coverage: %#v", i, summary)
		}
	}
	for _, path := range recorded.PromoterEvidenceFiles {
		if path == "" || filepath.IsAbs(path) {
			t.Fatalf("promoter evidence paths must be portable relative paths: %q", path)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.promoter-no-promotion-rollup.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:promoter-no-promotion-rollup" {
		t.Fatalf("expected typed promoter no-promotion rollup validator, got %s", validator)
	}
}
