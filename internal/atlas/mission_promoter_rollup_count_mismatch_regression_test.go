package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWavePromoterRollupCountMismatchRegressionRejectsDrift(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceRollupPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-25", "promoter-no-promotion-rollup.json")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-27")
	recordedPath := filepath.Join(nodeDir, "promoter-rollup-count-mismatch-regression.json")
	outPath := filepath.Join(t.TempDir(), "promoter-rollup-count-mismatch-regression.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "promoter-rollup-count-mismatch-regression",
		"--node-id", "mission-recommendation-feature-depth-next-wave-27",
		"--source-rollup", sourceRollupPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("promoter-rollup-count-mismatch-regression command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPromoterRollupCountMismatchRegression](t, recordedPath)
	generated := mustLoadJSON[AtlasPromoterRollupCountMismatchRegression](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Promoter count mismatch regression fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPromoterRollupCountMismatchRegression(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "mismatch_regression_recorded" ||
		recorded.CaseCount != 5 ||
		recorded.RejectedCases != recorded.CaseCount ||
		!recorded.CompletedNodesMismatchRejected ||
		!recorded.PromoterFilesMismatchRejected ||
		!recorded.MissingNodesMismatchRejected ||
		!recorded.NoPromotionStatusMismatchRejected ||
		!recorded.RSIDeniedMismatchRejected ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Promoter count mismatch regression lost rejection coverage: %#v", recorded)
	}
	for _, tc := range recorded.Cases {
		if !tc.Rejected || tc.ExpectedErrorContains == "" {
			t.Fatalf("regression case must record rejection and expected error text: %#v", tc)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.promoter-rollup-count-mismatch-regression.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:promoter-rollup-count-mismatch-regression" {
		t.Fatalf("expected typed Promoter count mismatch regression validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02PromoterRollupCountMismatchRegressionRejectsDrift(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	sourceRollupPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-25", "promoter-no-promotion-rollup.json")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-27")
	recordedPath := filepath.Join(nodeDir, "promoter-rollup-count-mismatch-regression.json")
	outPath := filepath.Join(t.TempDir(), "promoter-rollup-count-mismatch-regression.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "promoter-rollup-count-mismatch-regression",
		"--node-id", "mission-recommendation-feature-depth-next-wave-27",
		"--source-rollup", sourceRollupPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("promoter-rollup-count-mismatch-regression command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPromoterRollupCountMismatchRegression](t, recordedPath)
	generated := mustLoadJSON[AtlasPromoterRollupCountMismatchRegression](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 Promoter count mismatch regression fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPromoterRollupCountMismatchRegression(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "mismatch_regression_recorded" ||
		recorded.CaseCount != 5 ||
		recorded.RejectedCases != recorded.CaseCount ||
		!recorded.CompletedNodesMismatchRejected ||
		!recorded.PromoterFilesMismatchRejected ||
		!recorded.MissingNodesMismatchRejected ||
		!recorded.NoPromotionStatusMismatchRejected ||
		!recorded.RSIDeniedMismatchRejected ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 Promoter count mismatch regression lost rejection coverage: %#v", recorded)
	}
	for _, tc := range recorded.Cases {
		if !tc.Rejected || tc.ExpectedErrorContains == "" {
			t.Fatalf("v02 regression case must record rejection and expected error text: %#v", tc)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.promoter-rollup-count-mismatch-regression.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:promoter-rollup-count-mismatch-regression" {
		t.Fatalf("expected typed Promoter count mismatch regression validator, got %s", validator)
	}
}
