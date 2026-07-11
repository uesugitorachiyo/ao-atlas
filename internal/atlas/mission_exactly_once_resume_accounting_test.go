package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ExactlyOnceResumeAccountingFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-24")
	recordedPath := filepath.Join(nodeDir, "exactly-once-resume-accounting-fixture.json")
	outPath := filepath.Join(t.TempDir(), "exactly-once-resume-accounting-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "exactly-once-resume-accounting-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("exactly-once-resume-accounting-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=exactly_once_resume_accounting_ready",
		"scenario_count=3",
		"exactly_once_node_accounting=true",
		"duplicate_handoff_double_count_allowed=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("exactly-once resume accounting output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("exactly-once resume accounting fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["exactly_once_node_accounting"] != true ||
		generated["duplicate_handoff_double_count_allowed"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("exactly-once resume accounting fixture lost authority state: %#v", generated)
	}
}

func TestMonth3ExactlyOnceResumeAccountingFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-24", "exactly-once-resume-accounting-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.exactly-once-resume-accounting-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:exactly-once-resume-accounting-fixture" {
		t.Fatalf("expected typed exactly-once resume accounting validator, got %s", validator)
	}
}

func TestMonth3ExactlyOnceResumeAccountingFixtureRejectsDuplicateDoubleCount(t *testing.T) {
	fixture, err := BuildAtlasExactlyOnceResumeAccountingFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.DuplicateHandoffDoubleCountAllowed = true
	if err := ValidateAtlasExactlyOnceResumeAccountingFixture(fixture); err == nil || !strings.Contains(err.Error(), "duplicate_handoff_double_count_allowed must be false") {
		t.Fatalf("expected duplicate handoff double-count rejection, got %v", err)
	}
}
