package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3PromoterNoActivationBoundaryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-18")
	recordedPath := filepath.Join(nodeDir, "promoter-no-activation-boundary-fixture.json")
	outPath := filepath.Join(t.TempDir(), "promoter-no-activation-boundary-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "promoter-no-activation-boundary-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("promoter-no-activation-boundary-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=no_promotion_boundary_ready",
		"no_promotion_decision_supported=true",
		"activation_execution_owned=false",
		"release_execution_owned=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("Promoter boundary output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Promoter no-activation boundary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["activation_execution_owned"] != false ||
		generated["release_execution_owned"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("Promoter boundary fixture lost activation/release or authority state: %#v", generated)
	}
}

func TestMonth3PromoterNoActivationBoundaryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-18", "promoter-no-activation-boundary-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.promoter-no-activation-boundary-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:promoter-no-activation-boundary-fixture" {
		t.Fatalf("expected typed Promoter no-activation boundary validator, got %s", validator)
	}
}

func TestMonth3PromoterNoActivationBoundaryFixtureRejectsReleaseOwnership(t *testing.T) {
	fixture, err := BuildAtlasPromoterNoActivationBoundaryFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.ReleaseExecutionOwned = true
	if err := ValidateAtlasPromoterNoActivationBoundaryFixture(fixture); err == nil || !strings.Contains(err.Error(), "release_execution_owned must be false") {
		t.Fatalf("expected release ownership rejection, got %v", err)
	}
}
