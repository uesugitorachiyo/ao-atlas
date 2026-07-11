package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ContentAddressedEvidenceManifestFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-32")
	recordedPath := filepath.Join(nodeDir, "content-addressed-evidence-manifest-fixture.json")
	outPath := filepath.Join(t.TempDir(), "content-addressed-evidence-manifest-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "content-addressed-evidence-manifest-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("content-addressed-evidence-manifest-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=content_addressed_evidence_manifest_ready",
		"bulk_evidence_externalized=true",
		"small_replayable_fixtures_retained=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("content addressed manifest output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("content addressed manifest fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["bulk_evidence_externalized"] != true ||
		generated["small_replayable_fixtures_retained"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("content addressed manifest fixture lost safety state: %#v", generated)
	}
}

func TestMonth3ContentAddressedEvidenceManifestFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-32", "content-addressed-evidence-manifest-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.content-addressed-evidence-manifest-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:content-addressed-evidence-manifest-fixture" {
		t.Fatalf("expected typed content addressed evidence manifest validator, got %s", validator)
	}
}

func TestMonth3ContentAddressedEvidenceManifestFixtureRejectsMissingSmallFixtures(t *testing.T) {
	fixture, err := BuildAtlasContentAddressedEvidenceManifestFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.SmallReplayableFixturesRetained = false
	if err := ValidateAtlasContentAddressedEvidenceManifestFixture(fixture); err == nil || !strings.Contains(err.Error(), "small_replayable_fixtures_retained must be true") {
		t.Fatalf("expected missing small replayable fixture rejection, got %v", err)
	}
}
