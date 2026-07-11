package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3KillRestartReplayFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-38")
	recordedPath := filepath.Join(nodeDir, "kill-restart-replay-fixture.json")
	outPath := filepath.Join(t.TempDir(), "kill-restart-replay-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "kill-restart-replay-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("kill-restart-replay-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=kill_restart_replay_ready",
		"killed_run_replayed=true",
		"no_lost_evidence=true",
		"false_completion_detected=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("kill restart replay output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("kill restart replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["killed_run_replayed"] != true ||
		generated["no_lost_evidence"] != true ||
		generated["duplicate_mutation_detected"] != false ||
		generated["false_completion_detected"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("kill restart replay fixture lost safety state: %#v", generated)
	}
}

func TestMonth3KillRestartReplayFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-38", "kill-restart-replay-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.kill-restart-replay-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:kill-restart-replay-fixture" {
		t.Fatalf("expected typed kill restart replay validator, got %s", validator)
	}
}

func TestMonth3KillRestartReplayFixtureRejectsFalseCompletion(t *testing.T) {
	fixture, err := BuildAtlasKillRestartReplayFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.FalseCompletionDetected = true
	if err := ValidateAtlasKillRestartReplayFixture(fixture); err == nil || !strings.Contains(err.Error(), "false_completion_detected must be false") {
		t.Fatalf("expected false completion rejection, got %v", err)
	}
}
