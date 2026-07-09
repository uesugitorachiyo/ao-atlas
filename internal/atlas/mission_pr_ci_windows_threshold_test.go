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

func TestWindowsCIWaitStateHelpersCoverThresholdRowsAndFinalClosureTelemetry(t *testing.T) {
	root := repoRoot(t)
	thresholdEvidence := mustLoadJSON[AtlasPRCIWindowsThresholdEvidence](t, filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-10", "pr-ci-windows-threshold-evidence.json"))
	telemetry := mustLoadJSON[struct {
		WaitThresholdSeconds        int `json:"wait_threshold_seconds"`
		WindowsCheckDurationSamples []struct {
			CheckName       string `json:"check_name"`
			FinalStatus     string `json:"final_status"`
			FinalConclusion string `json:"final_conclusion"`
			DurationSeconds int    `json:"duration_seconds"`
			WaitState       string `json:"wait_state"`
		} `json:"windows_check_duration_samples"`
	}](t, filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01", "nodes", "mission-recommendation-final-closure-consolidation-16", "windows-ci-wait-state-telemetry.json"))

	for _, row := range thresholdEvidence.Rows {
		exceeds, over, err := EvaluateAtlasWindowsCIThreshold(row.WindowsSeconds, row.ThresholdSeconds)
		if err != nil {
			t.Fatal(err)
		}
		if exceeds != row.ExceedsThreshold || over != row.OverThresholdSeconds {
			t.Fatalf("threshold helper drifted from row evidence: %#v exceeds=%v over=%d", row, exceeds, over)
		}
	}

	for _, sample := range telemetry.WindowsCheckDurationSamples {
		state, err := ClassifyAtlasWindowsCIWaitState(AtlasWindowsCIWaitStateInput{
			CheckName:        sample.CheckName,
			GitHubStatus:     sample.FinalStatus,
			GitHubConclusion: sample.FinalConclusion,
			DurationSeconds:  sample.DurationSeconds,
			ThresholdSeconds: telemetry.WaitThresholdSeconds,
		})
		if err != nil {
			t.Fatal(err)
		}
		if state.WaitState != sample.WaitState ||
			state.State != "passing" ||
			state.OperatorAction != "merge_after_all_required_checks_pass" ||
			!state.ExceedsThreshold ||
			state.OverThresholdSeconds <= 0 ||
			state.RequiresCIWait ||
			state.CIFailure ||
			state.FinalResponseAllowed {
			t.Fatalf("wait-state helper drifted from telemetry sample: sample=%#v state=%#v", sample, state)
		}
	}

	cases := []struct {
		name           string
		status         string
		conclusion     string
		duration       int
		wantState      string
		wantWaitState  string
		wantAction     string
		wantCIWait     bool
		wantCIFailure  bool
		wantOver       int
		wantFinalAllow bool
	}{
		{"pending_long_running", "IN_PROGRESS", "", 900, "pending", "long_running_pending", "wait_for_ci", true, false, 300, false},
		{"passing_under_threshold", "COMPLETED", "SUCCESS", 120, "passing", "completed_success_under_threshold", "merge_after_all_required_checks_pass", false, false, 0, false},
		{"failure_long_running", "COMPLETED", "FAILURE", 901, "failing", "long_running_before_failure", "repair_before_merge", false, true, 301, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			state, err := ClassifyAtlasWindowsCIWaitState(AtlasWindowsCIWaitStateInput{
				CheckName:        "production-readiness (windows-latest)",
				GitHubStatus:     tt.status,
				GitHubConclusion: tt.conclusion,
				DurationSeconds:  tt.duration,
				ThresholdSeconds: 600,
			})
			if err != nil {
				t.Fatal(err)
			}
			if state.State != tt.wantState ||
				state.WaitState != tt.wantWaitState ||
				state.OperatorAction != tt.wantAction ||
				state.RequiresCIWait != tt.wantCIWait ||
				state.CIFailure != tt.wantCIFailure ||
				state.OverThresholdSeconds != tt.wantOver ||
				state.FinalResponseAllowed != tt.wantFinalAllow ||
				state.ClaimsAuthorityAdvance ||
				!state.RSIRemainsDenied {
				t.Fatalf("wait-state helper mismatch: got %#v", state)
			}
		})
	}
}
