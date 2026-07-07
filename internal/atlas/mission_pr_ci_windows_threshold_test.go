package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWavePRCIWindowsThresholdEvidenceFlagsLongRunningChecks(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-10")
	summaryPath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-09", "pr-ci-timing-summary.json")
	recorded := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "pr-ci-windows-threshold-evidence.json"))
	outPath := filepath.Join(t.TempDir(), "pr-ci-windows-threshold-evidence.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "pr-ci-windows-threshold",
		"--summary", summaryPath,
		"--threshold-seconds", "720",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("pr-ci-windows-threshold command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=windows_thresholds_recorded") ||
		!strings.Contains(out.String(), "threshold_seconds=720") ||
		!strings.Contains(out.String(), "long_running_windows_checks=3") {
		t.Fatalf("pr-ci-windows-threshold output missing threshold summary: %s", out.String())
	}
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("PR/CI Windows threshold fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["status"] != "windows_thresholds_recorded" ||
		generated["threshold_seconds"] != float64(720) ||
		generated["row_count"] != float64(3) ||
		generated["long_running_windows_checks"] != float64(3) ||
		generated["max_over_threshold_seconds"] != float64(92) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("threshold evidence must bind long-running Windows rows without authority effects: %#v", generated)
	}
}

func TestFeatureDepthWavePRCIWindowsThresholdEvidenceUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-10", "pr-ci-windows-threshold-evidence.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.pr-ci-windows-threshold-evidence.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:pr-ci-windows-threshold-evidence" {
		t.Fatalf("expected typed PR/CI Windows threshold validator, got %s", validator)
	}
}
