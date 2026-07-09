package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveCommandPromoterAgreementRollupBindsNoPromotionReadbacks(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-25")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-26")
	promoterRollupPath := filepath.Join(sourceNodeDir, "promoter-no-promotion-rollup.json")
	commandReadbackPath := filepath.Join(sourceNodeDir, "command_readback.json")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "command-promoter-agreement-rollup.json")
	outPath := filepath.Join(t.TempDir(), "command-promoter-agreement-rollup.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-promoter-agreement-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-26",
		"--promoter-rollup", promoterRollupPath,
		"--command-readback", commandReadbackPath,
		"--source-readback", sourceReadbackPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-promoter-agreement-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasCommandPromoterAgreementRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasCommandPromoterAgreementRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Command/Promoter agreement rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasCommandPromoterAgreementRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "command_agrees_with_promoter_no_promotion" ||
		recorded.PromoterNoPromotionInvariantHolds != true ||
		recorded.PromoterNoPromotionFiles != 88 ||
		recorded.CommandStatus != "readback_agrees_no_promotion" ||
		!recorded.CommandAgreesNoPromotion ||
		!recorded.ReadbackAgreesWithCommand ||
		recorded.ReadbackCompletedNodes != 25 ||
		recorded.ReadbackReadyNodes != 15 ||
		recorded.ReadbackFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-26" ||
		recorded.AggregatePromotionStatus != "no_promotion_requested" ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		recorded.FinalResponseAllowed ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("Command/Promoter rollup lost no-promotion agreement: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.command-promoter-agreement-rollup.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-promoter-agreement-rollup" {
		t.Fatalf("expected typed Command/Promoter agreement rollup validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02CommandPromoterAgreementRollupBindsNoPromotionReadbacks(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	sourceNodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-25")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-26")
	promoterRollupPath := filepath.Join(sourceNodeDir, "promoter-no-promotion-rollup.json")
	commandReadbackPath := filepath.Join(sourceNodeDir, "command_readback.json")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "command-promoter-agreement-rollup.json")
	outPath := filepath.Join(t.TempDir(), "command-promoter-agreement-rollup.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-promoter-agreement-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-26",
		"--promoter-rollup", promoterRollupPath,
		"--command-readback", commandReadbackPath,
		"--source-readback", sourceReadbackPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-promoter-agreement-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasCommandPromoterAgreementRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasCommandPromoterAgreementRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 Command/Promoter agreement rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasCommandPromoterAgreementRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "command_agrees_with_promoter_no_promotion" ||
		recorded.PromoterNoPromotionInvariantHolds != true ||
		recorded.PromoterNoPromotionFiles != 128 ||
		recorded.CommandStatus != "readback_agrees_no_promotion" ||
		!recorded.CommandAgreesNoPromotion ||
		!recorded.ReadbackAgreesWithCommand ||
		recorded.ReadbackCompletedNodes != 25 ||
		recorded.ReadbackReadyNodes != 15 ||
		recorded.ReadbackFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-26" ||
		recorded.AggregatePromotionStatus != "no_promotion_requested" ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		recorded.FinalResponseAllowed ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("v02 Command/Promoter rollup lost no-promotion agreement: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.command-promoter-agreement-rollup.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-promoter-agreement-rollup" {
		t.Fatalf("expected typed Command/Promoter agreement rollup validator, got %s", validator)
	}
}

func TestCommandPromoterAgreementRollupValidatorRejectsPromotionAndRSIBoundaryDrift(t *testing.T) {
	root := repoRoot(t)
	recordedPath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-26", "command-promoter-agreement-rollup.json")
	valid := mustLoadJSON[AtlasCommandPromoterAgreementRollup](t, recordedPath)
	if err := ValidateAtlasCommandPromoterAgreementRollup(valid); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		mutate  func(*AtlasCommandPromoterAgreementRollup)
		wantErr string
	}{
		{
			name: "aggregate promotion status",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.AggregatePromotionStatus = "promotion_blocked"
			},
			wantErr: "aggregate_promotion_status must be no_promotion_requested",
		},
		{
			name: "promotion requested",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.PromotionRequested = true
			},
			wantErr: "promotion_requested must be false",
		},
		{
			name: "promotion granted",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.PromotionGranted = true
			},
			wantErr: "promotion_granted must be false",
		},
		{
			name: "authority advance claim",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.ClaimsAuthorityAdvance = true
			},
			wantErr: "claims_authority_advance must be false",
		},
		{
			name: "rsi no longer denied",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.RSIRemainsDenied = false
			},
			wantErr: "rsi_remains_denied must be true",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mutated := valid
			tt.mutate(&mutated)
			err := ValidateAtlasCommandPromoterAgreementRollup(mutated)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q validation error, got %v", tt.wantErr, err)
			}
		})
	}
}
