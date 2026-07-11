package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ReplayableStatePacketFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-25")
	recordedPath := filepath.Join(nodeDir, "replayable-state-packet-fixture.json")
	outPath := filepath.Join(t.TempDir(), "replayable-state-packet-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "replayable-state-packet-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("replayable-state-packet-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=replayable_state_packet_ready",
		"state_count=4",
		"handoff_counts_as_completed=false",
		"replayable=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("replayable state packet output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("replayable state packet fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["handoff_counts_as_completed"] != false ||
		generated["replayable"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("replayable state packet fixture lost authority state: %#v", generated)
	}
}

func TestMonth3ReplayableStatePacketFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-25", "replayable-state-packet-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.replayable-state-packet-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:replayable-state-packet-fixture" {
		t.Fatalf("expected typed replayable state packet validator, got %s", validator)
	}
}

func TestMonth3ReplayableStatePacketFixtureRejectsHandoffCompletedCount(t *testing.T) {
	fixture, err := BuildAtlasReplayableStatePacketFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.HandoffCountsAsCompleted = true
	if err := ValidateAtlasReplayableStatePacketFixture(fixture); err == nil || !strings.Contains(err.Error(), "handoff_counts_as_completed must be false") {
		t.Fatalf("expected handoff-as-completed rejection, got %v", err)
	}
}
