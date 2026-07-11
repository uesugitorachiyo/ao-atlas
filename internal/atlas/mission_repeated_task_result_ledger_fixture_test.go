package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3RepeatedTaskResultLedgerFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-34")
	recordedPath := filepath.Join(nodeDir, "repeated-task-result-ledger-fixture.json")
	outPath := filepath.Join(t.TempDir(), "repeated-task-result-ledger-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "repeated-task-result-ledger-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("repeated-task-result-ledger-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=repeated_task_result_ledger_ready",
		"attempt_count=3",
		"replayable_result_ledger=true",
		"live_provider_calls=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("repeated task ledger output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("repeated task ledger fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["deterministic_harness"] != true ||
		generated["replayable_result_ledger"] != true ||
		generated["live_provider_calls"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("repeated task ledger fixture lost safety state: %#v", generated)
	}
}

func TestMonth3RepeatedTaskResultLedgerFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-34", "repeated-task-result-ledger-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.repeated-task-result-ledger-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:repeated-task-result-ledger-fixture" {
		t.Fatalf("expected typed repeated task result ledger validator, got %s", validator)
	}
}

func TestMonth3RepeatedTaskResultLedgerFixtureRejectsProviderCalls(t *testing.T) {
	fixture, err := BuildAtlasRepeatedTaskResultLedgerFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.LiveProviderCalls = true
	if err := ValidateAtlasRepeatedTaskResultLedgerFixture(fixture); err == nil || !strings.Contains(err.Error(), "live_provider_calls must be false") {
		t.Fatalf("expected provider call rejection, got %v", err)
	}
}
