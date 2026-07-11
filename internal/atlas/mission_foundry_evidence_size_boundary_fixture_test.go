package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FoundryEvidenceSizeBoundaryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-33")
	recordedPath := filepath.Join(nodeDir, "foundry-evidence-size-boundary-fixture.json")
	outPath := filepath.Join(t.TempDir(), "foundry-evidence-size-boundary-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-evidence-size-boundary-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-evidence-size-boundary-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=foundry_evidence_size_boundary_ready",
		"evidence_reference_count=2",
		"implementation_state_separate=true",
		"generated_campaign_bulk_separate=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("foundry evidence size boundary output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("foundry evidence size boundary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["implementation_state_separate"] != true ||
		generated["generated_campaign_bulk_separate"] != true ||
		generated["size_checks_required"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("foundry evidence size boundary fixture lost safety state: %#v", generated)
	}
}

func TestMonth3FoundryEvidenceSizeBoundaryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-33", "foundry-evidence-size-boundary-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.foundry-evidence-size-boundary-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-evidence-size-boundary-fixture" {
		t.Fatalf("expected typed foundry evidence size boundary validator, got %s", validator)
	}
}

func TestMonth3FoundryEvidenceSizeBoundaryFixtureRejectsMixedBulk(t *testing.T) {
	fixture, err := BuildAtlasFoundryEvidenceSizeBoundaryFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.GeneratedCampaignBulkSeparate = false
	if err := ValidateAtlasFoundryEvidenceSizeBoundaryFixture(fixture); err == nil || !strings.Contains(err.Error(), "generated_campaign_bulk_separate must be true") {
		t.Fatalf("expected mixed bulk rejection, got %v", err)
	}
}
