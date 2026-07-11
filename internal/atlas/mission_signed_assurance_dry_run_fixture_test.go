package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3SignedAssuranceDryRunFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-17")
	recordedPath := filepath.Join(nodeDir, "signed-assurance-dry-run-fixture.json")
	outPath := filepath.Join(t.TempDir(), "signed-assurance-dry-run-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "signed-assurance-dry-run-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("signed-assurance-dry-run-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=dry_run_verification_ready",
		"required_check_count=3",
		"promotion_decision_enabled=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("dry-run output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("signed assurance dry-run fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["promotion_decision_enabled"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("dry-run fixture lost no-promotion or authority state: %#v", generated)
	}
}

func TestMonth3SignedAssuranceDryRunFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-17", "signed-assurance-dry-run-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.signed-assurance-dry-run-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:signed-assurance-dry-run-fixture" {
		t.Fatalf("expected typed signed assurance dry-run validator, got %s", validator)
	}
}

func TestMonth3SignedAssuranceDryRunFixtureRejectsPromotionEnabled(t *testing.T) {
	fixture, err := BuildAtlasSignedAssuranceDryRunFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.PromotionDecisionEnabled = true
	if err := ValidateAtlasSignedAssuranceDryRunFixture(fixture); err == nil || !strings.Contains(err.Error(), "promotion_decision_enabled must be false") {
		t.Fatalf("expected enabled promotion rejection, got %v", err)
	}
}
