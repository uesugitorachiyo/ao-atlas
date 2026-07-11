package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3NonAOReplayBindingFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-37")
	recordedPath := filepath.Join(nodeDir, "non-ao-replay-binding-fixture.json")
	outPath := filepath.Join(t.TempDir(), "non-ao-replay-binding-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "non-ao-replay-binding-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("non-ao-replay-binding-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=non_ao_replay_binding_ready",
		"reviewed_pr_evidence=true",
		"observer_readback_bound=true",
		"promotion_requested=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("non-AO replay output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("non-AO replay binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["tiny_non_ao_repo"] != true ||
		generated["reviewed_pr_evidence"] != true ||
		generated["observer_readback_bound"] != true ||
		generated["promotion_requested"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("non-AO replay binding fixture lost safety state: %#v", generated)
	}
}

func TestMonth3NonAOReplayBindingFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-37", "non-ao-replay-binding-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.non-ao-replay-binding-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:non-ao-replay-binding-fixture" {
		t.Fatalf("expected typed non-AO replay binding validator, got %s", validator)
	}
}

func TestMonth3NonAOReplayBindingFixtureRejectsPromotion(t *testing.T) {
	fixture, err := BuildAtlasNonAOReplayBindingFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.PromotionRequested = true
	if err := ValidateAtlasNonAOReplayBindingFixture(fixture); err == nil || !strings.Contains(err.Error(), "promotion_requested must be false") {
		t.Fatalf("expected promotion rejection, got %v", err)
	}
}
