package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FailureInjectionFuzzingFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-35")
	recordedPath := filepath.Join(nodeDir, "failure-injection-fuzzing-fixture.json")
	outPath := filepath.Join(t.TempDir(), "failure-injection-fuzzing-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "failure-injection-fuzzing-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("failure-injection-fuzzing-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=failure_injection_fuzzing_ready",
		"case_count=4",
		"deterministic_fuzzing=true",
		"live_provider_calls=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("failure injection fuzzing output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("failure injection fuzzing fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["deterministic_fuzzing"] != true ||
		generated["replayable_cases"] != true ||
		generated["live_provider_calls"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("failure injection fuzzing fixture lost safety state: %#v", generated)
	}
}

func TestMonth3FailureInjectionFuzzingFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-35", "failure-injection-fuzzing-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.failure-injection-fuzzing-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:failure-injection-fuzzing-fixture" {
		t.Fatalf("expected typed failure injection fuzzing validator, got %s", validator)
	}
}

func TestMonth3FailureInjectionFuzzingFixtureRequiresRollbackReceiptCase(t *testing.T) {
	fixture, err := BuildAtlasFailureInjectionFuzzingFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.Cases = fixture.Cases[:3]
	fixture.CaseCount = len(fixture.Cases)
	if err := ValidateAtlasFailureInjectionFuzzingFixture(fixture); err == nil || !strings.Contains(err.Error(), "rollback_receipt case is required") {
		t.Fatalf("expected rollback receipt case rejection, got %v", err)
	}
}
