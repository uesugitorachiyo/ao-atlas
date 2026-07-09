package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveCommandPromoterDisagreementDenialBlocksFinalResponse(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceAgreementPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-26", "command-promoter-agreement-rollup.json")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-28")
	recordedPath := filepath.Join(nodeDir, "command-promoter-disagreement-denial.json")
	outPath := filepath.Join(t.TempDir(), "command-promoter-disagreement-denial.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-promoter-disagreement-denial",
		"--node-id", "mission-recommendation-feature-depth-next-wave-28",
		"--source-agreement", sourceAgreementPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-promoter-disagreement-denial command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasCommandPromoterDisagreementDenial](t, recordedPath)
	generated := mustLoadJSON[AtlasCommandPromoterDisagreementDenial](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Command/Promoter disagreement denial fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasCommandPromoterDisagreementDenial(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "final_response_denied_command_promoter_disagreement" ||
		recorded.CaseCount != 4 ||
		recorded.DeniedCases != recorded.CaseCount ||
		!recorded.CommandPromoterDisagreementDetected ||
		recorded.FinalResponseAllowed ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Command/Promoter disagreement denial lost final-response guard state: %#v", recorded)
	}
	for _, tc := range recorded.Cases {
		if !tc.DisagreementDetected || !tc.FinalResponseDenied || tc.FinalResponseAllowed {
			t.Fatalf("disagreement case must deny final response: %#v", tc)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.command-promoter-disagreement-denial.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-promoter-disagreement-denial" {
		t.Fatalf("expected typed Command/Promoter disagreement denial validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02CommandPromoterDisagreementDenialBlocksFinalResponse(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	sourceAgreementPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-26", "command-promoter-agreement-rollup.json")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-28")
	recordedPath := filepath.Join(nodeDir, "command-promoter-disagreement-denial.json")
	outPath := filepath.Join(t.TempDir(), "command-promoter-disagreement-denial.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-promoter-disagreement-denial",
		"--node-id", "mission-recommendation-feature-depth-next-wave-28",
		"--source-agreement", sourceAgreementPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-promoter-disagreement-denial command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasCommandPromoterDisagreementDenial](t, recordedPath)
	generated := mustLoadJSON[AtlasCommandPromoterDisagreementDenial](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 Command/Promoter disagreement denial fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasCommandPromoterDisagreementDenial(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "final_response_denied_command_promoter_disagreement" ||
		recorded.CaseCount != 4 ||
		recorded.DeniedCases != recorded.CaseCount ||
		!recorded.CommandPromoterDisagreementDetected ||
		recorded.FinalResponseAllowed ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 Command/Promoter disagreement denial lost final-response guard state: %#v", recorded)
	}
	for _, tc := range recorded.Cases {
		if !tc.DisagreementDetected || !tc.FinalResponseDenied || tc.FinalResponseAllowed {
			t.Fatalf("v02 disagreement case must deny final response: %#v", tc)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.command-promoter-disagreement-denial.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-promoter-disagreement-denial" {
		t.Fatalf("expected typed Command/Promoter disagreement denial validator, got %s", validator)
	}
}
