package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosureSentinelFreshnessSignalFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-24-sentinel-freshness-signal")
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

	recorded := mustLoadJSON[AtlasSentinelSignalStateFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasSentinelSignalStateFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Sentinel freshness signal fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded.Status != "signal_states_ready" ||
		!containsAll(recorded.Signals, []string{"ci", "runtime", "policy", "evidence_freshness"}) ||
		!containsAll(recorded.States, []string{"stale", "pending", "pass", "failure"}) ||
		recorded.MatrixCount != 16 ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Sentinel freshness signal fixture lost freshness or safety state: %#v", recorded)
	}
}

func TestMonth3FinalClosureSentinelFreshnessSignalFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-24-sentinel-freshness-signal", "sentinel-signal-state-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.sentinel-signal-state-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:sentinel-signal-state-fixture" {
		t.Fatalf("expected typed Sentinel signal state validator, got %s", validator)
	}
}
