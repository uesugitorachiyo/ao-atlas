package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosurePromoterNoActivationBoundaryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-25-promoter-no-activation")
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

	recorded := mustLoadJSON[AtlasPromoterNoActivationBoundaryFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasPromoterNoActivationBoundaryFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Promoter no-activation fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded.Decision != "no_promotion" ||
		!recorded.NoPromotionDecisionSupported ||
		recorded.ActivationExecutionOwned ||
		recorded.ReleaseExecutionOwned ||
		!containsAll(recorded.ForbiddenActions, []string{"activate", "release", "deploy", "publish", "tag"}) ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Promoter no-activation fixture lost boundary or safety state: %#v", recorded)
	}
}

func TestMonth3FinalClosurePromoterNoActivationBoundaryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-25-promoter-no-activation", "promoter-no-activation-boundary-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.promoter-no-activation-boundary-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:promoter-no-activation-boundary-fixture" {
		t.Fatalf("expected typed Promoter no-activation boundary validator, got %s", validator)
	}
}
