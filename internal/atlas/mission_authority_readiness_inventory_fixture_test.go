package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3AuthorityReadinessInventoryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-31")
	recordedPath := filepath.Join(nodeDir, "authority-readiness-inventory-fixture.json")
	outPath := filepath.Join(t.TempDir(), "authority-readiness-inventory-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "authority-readiness-inventory-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("authority-readiness-inventory-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=authority_readiness_inventory_ready",
		"input_count=2",
		"generated_from_inputs=true",
		"copied_campaign_prose_allowed=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("authority readiness inventory output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("authority readiness inventory fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["generated_from_inputs"] != true ||
		generated["copied_campaign_prose_allowed"] != false ||
		generated["executes_work"] != false ||
		generated["approves_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("authority readiness inventory fixture lost safety state: %#v", generated)
	}
}

func TestMonth3AuthorityReadinessInventoryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-31", "authority-readiness-inventory-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.authority-readiness-inventory-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:authority-readiness-inventory-fixture" {
		t.Fatalf("expected typed authority readiness inventory validator, got %s", validator)
	}
}

func TestMonth3AuthorityReadinessInventoryFixtureRejectsCopiedCampaignProse(t *testing.T) {
	fixture, err := BuildAtlasAuthorityReadinessInventoryFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.CopiedCampaignProseAllowed = true
	if err := ValidateAtlasAuthorityReadinessInventoryFixture(fixture); err == nil || !strings.Contains(err.Error(), "copied_campaign_prose_allowed must be false") {
		t.Fatalf("expected copied campaign prose rejection, got %v", err)
	}
}
