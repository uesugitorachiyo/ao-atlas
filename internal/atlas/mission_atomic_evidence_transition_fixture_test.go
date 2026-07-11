package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3AtomicEvidenceTransitionFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-27")
	recordedPath := filepath.Join(nodeDir, "atomic-evidence-transition-fixture.json")
	outPath := filepath.Join(t.TempDir(), "atomic-evidence-transition-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "atomic-evidence-transition-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("atomic-evidence-transition-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=atomic_evidence_transition_ready",
		"scenario_count=4",
		"atomic_transitions_required=true",
		"duplicate_ingest_idempotent=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("atomic evidence transition output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("atomic evidence transition fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["atomic_transitions_required"] != true ||
		generated["duplicate_ingest_idempotent"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("atomic evidence transition fixture lost authority state: %#v", generated)
	}
}

func TestMonth3AtomicEvidenceTransitionFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-27", "atomic-evidence-transition-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.atomic-evidence-transition-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:atomic-evidence-transition-fixture" {
		t.Fatalf("expected typed atomic evidence transition validator, got %s", validator)
	}
}

func TestMonth3AtomicEvidenceTransitionFixtureRejectsNonAtomicCrash(t *testing.T) {
	fixture, err := BuildAtlasAtomicEvidenceTransitionFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.Scenarios[0].AtomicTransition = false
	if err := ValidateAtlasAtomicEvidenceTransitionFixture(fixture); err == nil || !strings.Contains(err.Error(), "scenarios[0].atomic_transition must be true") {
		t.Fatalf("expected non-atomic crash transition rejection, got %v", err)
	}
}
