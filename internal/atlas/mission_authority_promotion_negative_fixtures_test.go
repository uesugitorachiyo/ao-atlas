package atlas

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveAuthorityPromotionNegativeFixtures(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-23")
	recordedPath := filepath.Join(nodeDir, "authority-promotion-negative-fixtures.json")
	outPath := filepath.Join(t.TempDir(), "authority-promotion-negative-fixtures.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "authority-promotion-negative-fixtures",
		"--node-id", "mission-recommendation-feature-depth-next-wave-23",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("authority-promotion-negative-fixtures command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, recordedPath)
	generated := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("authority promotion negative fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasAuthorityPromotionNegativeFixtures(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "passed" ||
		recorded.FixtureEncoding != "redacted_token_sequences" ||
		recorded.CaseCount < 7 ||
		len(recorded.Cases) != recorded.CaseCount ||
		!recorded.ForbiddenPatternsRedacted ||
		recorded.UnsafeLiteralStored ||
		recorded.ExpectedScanStatus != "failed" ||
		recorded.ExpectedPublicSafetyScanPassed ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("negative authority wording fixture lost safety state: %#v", recorded)
	}
	seen := map[string]bool{}
	for _, fixture := range recorded.Cases {
		seen[fixture.ScannerPatternID] = true
		if fixture.ExpectedUnsafeMatches <= 0 ||
			fixture.ExpectedScanStatus != "failed" ||
			len(fixture.StatementTokens) < 2 {
			t.Fatalf("negative fixture case must describe a failed scanner match without raw wording: %#v", fixture)
		}
	}
	for _, want := range []string{
		"promotion_granted_true",
		"promotion_claimed_true",
		"claims_authority_advance_true",
		"fully_unsupervised_complex_mutation_live_proven_true",
		"rsi_is_proven_phrase",
		"rsi_proof_granted_phrase",
		"fully_unsupervised_complex_mutation_is_live_proven_phrase",
	} {
		if !seen[want] {
			t.Fatalf("negative fixture missing scanner pattern %q: %#v", want, seen)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.authority-promotion-negative-fixtures.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:authority-promotion-negative-fixtures" {
		t.Fatalf("expected typed authority promotion negative fixture validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02AuthorityPromotionNegativeFixtures(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02", "nodes", "mission-recommendation-feature-depth-next-wave-23")
	recordedPath := filepath.Join(nodeDir, "authority-promotion-negative-fixtures.json")
	outPath := filepath.Join(t.TempDir(), "authority-promotion-negative-fixtures.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "authority-promotion-negative-fixtures",
		"--node-id", "mission-recommendation-feature-depth-next-wave-23",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("authority-promotion-negative-fixtures command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, recordedPath)
	generated := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 authority promotion negative fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasAuthorityPromotionNegativeFixtures(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "passed" ||
		recorded.FixtureEncoding != "redacted_token_sequences" ||
		recorded.CaseCount < 7 ||
		!recorded.ForbiddenPatternsRedacted ||
		recorded.UnsafeLiteralStored ||
		recorded.ExpectedScanStatus != "failed" ||
		recorded.ExpectedPublicSafetyScanPassed ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 negative authority wording fixture lost safety state: %#v", recorded)
	}
}

func TestFeatureDepthWaveScopedPublicSafetyScanRejectsNegativeAuthorityPromotionSamples(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "unsafe-authority.json")
	markdownPath := filepath.Join(dir, "unsafe-authority.md")

	jsonUnsafe := strings.Join([]string{
		fmt.Sprintf("%q: %s", "promotion_granted", "true"),
		fmt.Sprintf("%q: %s", "promotion_claimed", "true"),
		fmt.Sprintf("%q: %s", "claims_authority_advance", "true"),
		fmt.Sprintf("%q: %s", "fully_unsupervised_complex_mutation_live_proven", "true"),
	}, ",\n")
	if err := os.WriteFile(jsonPath, []byte("{\n"+jsonUnsafe+"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	markdownUnsafe := strings.Join([]string{
		strings.Join([]string{"RSI", "is", "proven"}, " "),
		strings.Join([]string{"RSI", "proof", "granted"}, " "),
		strings.Join([]string{"fully_unsupervised_complex_mutation", "is", "live-proven"}, " "),
	}, "\n")
	if err := os.WriteFile(markdownPath, []byte(markdownUnsafe+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	scan, err := BuildAtlasScopedPublicSafetyScan("mission-recommendation-feature-depth-next-wave-23", []string{dir})
	if err == nil {
		t.Fatalf("expected unsafe authority promotion samples to fail scoped public-safety scan: %#v", scan)
	}
	if scan.Status != "failed" ||
		scan.UnsafeMatchCount != 7 ||
		scan.PublicSafetyScanPassed ||
		scan.ChangedEvidenceFiles != 1 ||
		scan.ChangedPromptArtifacts != 1 ||
		scan.ClaimsAuthorityAdvance ||
		!scan.RSIRemainsDenied {
		t.Fatalf("negative authority samples must fail without authority effects: %#v", scan)
	}
}
