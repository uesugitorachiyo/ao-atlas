package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3SentinelSignalStateFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-16")
	recordedPath := filepath.Join(nodeDir, "sentinel-signal-state-fixture.json")
	outPath := filepath.Join(t.TempDir(), "sentinel-signal-state-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "sentinel-signal-state-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("sentinel-signal-state-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=signal_states_ready",
		"signal_count=4",
		"state_count=4",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("signal state output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Sentinel signal state fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["claims_authority_advance"] != false || generated["rsi_remains_denied"] != true {
		t.Fatalf("Sentinel signal state fixture lost authority state: %#v", generated)
	}
}

func TestMonth3SentinelSignalStateFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-16", "sentinel-signal-state-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.sentinel-signal-state-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:sentinel-signal-state-fixture" {
		t.Fatalf("expected typed Sentinel signal state validator, got %s", validator)
	}
}

func TestMonth3SentinelSignalStateFixtureRejectsMissingFailure(t *testing.T) {
	fixture, err := BuildAtlasSentinelSignalStateFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.States = []string{"stale", "pending", "pass"}
	fixture.StateCount = len(fixture.States)
	if err := ValidateAtlasSentinelSignalStateFixture(fixture); err == nil || !strings.Contains(err.Error(), "states must cover stale, pending, pass, and failure") {
		t.Fatalf("expected missing failure state rejection, got %v", err)
	}
}
