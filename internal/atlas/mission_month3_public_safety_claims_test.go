package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3FinalClosurePublicSafetyClaimsNegativeFixtures(t *testing.T) {
	root := repoRoot(t)
	nodeID := "mission-recommendation-month3-final-closure-29-public-safety-claims"
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", nodeID)
	recordedPath := filepath.Join(nodeDir, "authority-promotion-negative-fixtures.json")
	outPath := filepath.Join(t.TempDir(), "authority-promotion-negative-fixtures.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "authority-promotion-negative-fixtures",
		"--node-id", nodeID,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("authority-promotion-negative-fixtures command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, recordedPath)
	generated := mustLoadJSON[AtlasAuthorityPromotionNegativeFixtures](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("public safety claims fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded.NodeID != nodeID ||
		recorded.Status != "passed" ||
		recorded.FixtureEncoding != "redacted_token_sequences" ||
		recorded.CaseCount < 7 ||
		!recorded.ForbiddenPatternsRedacted ||
		recorded.UnsafeLiteralStored ||
		recorded.ExpectedScanStatus != "failed" ||
		recorded.ExpectedPublicSafetyScanPassed ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("public safety claims fixture lost negative-scan state: %#v", recorded)
	}
}

func TestMonth3FinalClosurePublicSafetyClaimsNegativeFixturesUseTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-29-public-safety-claims", "authority-promotion-negative-fixtures.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.authority-promotion-negative-fixtures.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:authority-promotion-negative-fixtures" {
		t.Fatalf("expected typed authority promotion negative fixture validator, got %s", validator)
	}
}
