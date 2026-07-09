package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
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

func TestFeatureDepthWaveV02PromoterNoPromotionRollupAggregatesCompletedWaves(t *testing.T) {
	root := repoRoot(t)
	hardeningRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	featureV01Root := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	featureV02Root := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(featureV02Root, "nodes", "mission-recommendation-feature-depth-next-wave-25")
	sourceReadbackPath := filepath.Join(featureV02Root, "nodes", "mission-recommendation-feature-depth-next-wave-24", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "promoter-no-promotion-rollup.json")
	outPath := filepath.Join(t.TempDir(), "promoter-no-promotion-rollup.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "promoter-no-promotion-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-25",
		"--source-readback", sourceReadbackPath,
		"--evidence-root", hardeningRoot,
		"--evidence-root", closureRoot,
		"--evidence-root", featureV01Root,
		"--evidence-root", featureV02Root,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("promoter-no-promotion-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPromoterNoPromotionRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasPromoterNoPromotionRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 promoter no-promotion rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPromoterNoPromotionRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "no_promotion_rollup_bound" ||
		recorded.SourceReadbackCompletedNodes != 24 ||
		recorded.SourceReadbackReadyNodes != 16 ||
		recorded.SourceReadbackFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-25" ||
		recorded.SourceReadbackFinalResponseAllowed ||
		recorded.CompletedNodesTotal != 128 ||
		recorded.PromoterNoPromotionFiles != 128 ||
		recorded.MissingPromoterNodesTotal != 0 ||
		recorded.NoPromotionStatusCount != 128 ||
		recorded.PromotionRequestedCount != 0 ||
		recorded.PromotionGrantedCount != 0 ||
		recorded.PromotionClaimedCount != 0 ||
		recorded.AuthorityAdvanceClaimCount != 0 ||
		recorded.RSIDeniedCount != 128 ||
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
		t.Fatalf("v02 promoter rollup lost no-promotion/no-RSI invariants: %#v", recorded)
	}
	if len(recorded.WaveSummaries) != 4 {
		t.Fatalf("expected hardening, closure, v01, and v02 wave summaries: %#v", recorded.WaveSummaries)
	}
	expectedCompleted := []int{40, 24, 40, 24}
	for i, summary := range recorded.WaveSummaries {
		if summary.CompletedNodes != expectedCompleted[i] ||
			summary.PromoterNoPromotionFiles != expectedCompleted[i] ||
			summary.MissingPromoterNodes != 0 ||
			summary.PromotionRequestedCount != 0 ||
			summary.PromotionGrantedCount != 0 ||
			summary.AuthorityAdvanceClaimCount != 0 ||
			!summary.NoPromotionInvariantHolds ||
			!summary.RSIRemainsDenied {
			t.Fatalf("v02 wave summary %d lost no-promotion coverage: %#v", i, summary)
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

func TestPromoterNoPromotionRollupValidatorRejectsPromotionAndRSIBoundaryDrift(t *testing.T) {
	root := repoRoot(t)
	recordedPath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-25", "promoter-no-promotion-rollup.json")
	valid := mustLoadJSON[AtlasPromoterNoPromotionRollup](t, recordedPath)
	if err := ValidateAtlasPromoterNoPromotionRollup(valid); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		mutate  func(*AtlasPromoterNoPromotionRollup)
		wantErr string
	}{
		{
			name: "promotion requested count",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.PromotionRequested = true
				rollup.PromotionRequestedCount = 1
				rollup.WaveSummaries[0].PromotionRequestedCount = 1
			},
			wantErr: "no_promotion_invariant_holds",
		},
		{
			name: "promotion granted count",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.PromotionGranted = true
				rollup.PromotionGrantedCount = 1
				rollup.WaveSummaries[0].PromotionGrantedCount = 1
			},
			wantErr: "no_promotion_invariant_holds",
		},
		{
			name: "authority advance claim",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.ClaimsAuthorityAdvance = true
				rollup.AuthorityAdvanceClaimCount = 1
				rollup.WaveSummaries[0].AuthorityAdvanceClaimCount = 1
			},
			wantErr: "claims_authority_advance must be false",
		},
		{
			name: "rsi no longer denied",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.RSIRemainsDenied = false
				rollup.RSIDeniedCount--
				rollup.WaveSummaries[0].RSIRemainsDenied = false
				rollup.WaveSummaries[0].RSIDeniedCount--
			},
			wantErr: "rsi_remains_denied must be true",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mutated := clonePromoterNoPromotionRollup(valid)
			tt.mutate(&mutated)
			err := ValidateAtlasPromoterNoPromotionRollup(mutated)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q validation error, got %v", tt.wantErr, err)
			}
		})
	}
}
