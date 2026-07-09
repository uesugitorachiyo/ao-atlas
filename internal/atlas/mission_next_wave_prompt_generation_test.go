package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveNextWavePromptPreservesMinimumTwoHourBudget(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	recommendationsPath := "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-37/next-wave-feature-depth-recommendations.json"
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-38")
	recordedOutDir := filepath.Join(nodeDir, "generated-next-wave")
	recordedPromptPath := filepath.Join(recordedOutDir, "next-recommended-prompt.md")
	recordedWavePath := filepath.Join(recordedOutDir, "recommendation-wave.json")
	tmpDir := t.TempDir()
	normalizedRecommendationsPath := filepath.Join(tmpDir, "next-wave-feature-depth-recommendations.json")
	recommendationsData, err := os.ReadFile(recommendationsPath)
	if err != nil {
		t.Fatal(err)
	}
	recommendationsData = bytes.ReplaceAll(recommendationsData, []byte("\r\n"), []byte("\n"))
	recommendationsData = bytes.ReplaceAll(recommendationsData, []byte("\r"), []byte("\n"))
	if err := os.WriteFile(normalizedRecommendationsPath, recommendationsData, 0o644); err != nil {
		t.Fatal(err)
	}
	tmpOutDir := filepath.Join(tmpDir, "generated-next-wave")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", normalizedRecommendationsPath,
		"--target-instance", "ao-atlas-feature-depth-followup-wave-v01",
		"--min-tasks", "40",
		"--node-budget", "40",
		"--estimated-minutes", "150",
		"--min-minutes", "120",
		"--max-minutes", "180",
		"--continue-if-fast-target", "40",
		"--return-only-when", "all_40_feature_depth_nodes_complete_or_true_hard_blocker",
		"--checkpoint-policy", "after_each_node_with_pr_ci_merge_cleanup",
		"--evidence-policy", "evidence_after_each_node",
		"--final-report-contract", "feature_depth_recommendations_and_clean_repo_status",
		"--started-at", "2026-07-07T00:00:00Z",
		"--out", tmpOutDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("next-wave import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "min_minutes=120") ||
		!strings.Contains(out.String(), "max_minutes=180") ||
		!strings.Contains(out.String(), "continue_if_fast_target=40") {
		t.Fatalf("next-wave import output missing two-hour work budget: %s", out.String())
	}

	generatedPrompt, err := os.ReadFile(filepath.Join(tmpOutDir, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	recordedPrompt, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatal(err)
	}
	generatedPromptText := normalizeRecommendationPromptFixture(string(generatedPrompt))
	recordedPromptText := normalizeRecommendationPromptFixture(string(recordedPrompt))
	if generatedPromptText != recordedPromptText {
		t.Fatalf("next-wave prompt fixture changed\nwant digest %s\ngot  digest %s", digestValue(recordedPromptText), digestValue(generatedPromptText))
	}
	prompt := recordedPromptText
	for _, want := range []string{
		"Target 2-3 hours",
		"min_minutes: 120",
		"max_minutes: 180",
		"Lease floor stop gate: do not return before min_minutes=120",
		"Feature Depth Recommendations, at least 40 tasks",
		"If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("next-wave prompt missing required budget/stop-gate language %q:\n%s", want, prompt)
		}
	}

	recordedWave := mustLoadJSON[AtlasRecommendationWave](t, recordedWavePath)
	generatedWave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(tmpOutDir, "recommendation-wave.json"))
	if digestValue(generatedWave) != digestValue(recordedWave) {
		t.Fatalf("next-wave recommendation wave fixture changed\nwant %s\ngot  %s", digestValue(recordedWave), digestValue(generatedWave))
	}
	if err := ValidateAtlasRecommendationWave(recordedWave); err != nil {
		t.Fatal(err)
	}
	if recordedWave.TotalTasks != 40 ||
		recordedWave.MinimumTasks != 40 ||
		recordedWave.NodeBudget != 40 ||
		recordedWave.EstimatedMinutes != 150 ||
		recordedWave.Supervisor == nil ||
		recordedWave.Supervisor.MinMinutes != 120 ||
		recordedWave.Supervisor.MaxMinutes != 180 ||
		recordedWave.Supervisor.ContinueIfFastTarget != 40 ||
		recordedWave.FinalResponseAllowed ||
		recordedWave.SafeToExecute ||
		recordedWave.SchedulesWork ||
		recordedWave.ExecutesWork ||
		recordedWave.ApprovesWork {
		t.Fatalf("next-wave prompt generation lost 40-node/two-hour planning-only contract: %#v", recordedWave)
	}
}

func TestFeatureDepthWaveV02NextWavePromptPreservesMinimumTwoHourBudget(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	recommendationsPath := "docs/evidence/ao-atlas-feature-depth-wave-v02/nodes/mission-recommendation-feature-depth-next-wave-37/next-wave-feature-depth-recommendations.json"
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02", "nodes", "mission-recommendation-feature-depth-next-wave-38")
	recordedOutDir := filepath.Join(nodeDir, "generated-next-wave")
	recordedPromptPath := filepath.Join(recordedOutDir, "next-recommended-prompt.md")
	recordedWavePath := filepath.Join(recordedOutDir, "recommendation-wave.json")
	tmpDir := t.TempDir()
	normalizedRecommendationsPath := filepath.Join(tmpDir, "next-wave-feature-depth-recommendations.json")
	recommendationsData, err := os.ReadFile(recommendationsPath)
	if err != nil {
		t.Fatal(err)
	}
	recommendationsData = bytes.ReplaceAll(recommendationsData, []byte("\r\n"), []byte("\n"))
	recommendationsData = bytes.ReplaceAll(recommendationsData, []byte("\r"), []byte("\n"))
	if err := os.WriteFile(normalizedRecommendationsPath, recommendationsData, 0o644); err != nil {
		t.Fatal(err)
	}
	tmpOutDir := filepath.Join(tmpDir, "generated-next-wave")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", normalizedRecommendationsPath,
		"--target-instance", "ao-atlas-feature-depth-followup-wave-v02",
		"--min-tasks", "40",
		"--node-budget", "40",
		"--estimated-minutes", "150",
		"--min-minutes", "120",
		"--max-minutes", "180",
		"--continue-if-fast-target", "40",
		"--return-only-when", "all_40_feature_depth_nodes_complete_or_true_hard_blocker",
		"--checkpoint-policy", "after_each_node_with_pr_ci_merge_cleanup",
		"--evidence-policy", "evidence_after_each_node",
		"--final-report-contract", "feature_depth_recommendations_and_clean_repo_status",
		"--started-at", "2026-07-09T00:00:00Z",
		"--out", tmpOutDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("next-wave import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "min_minutes=120") ||
		!strings.Contains(out.String(), "max_minutes=180") ||
		!strings.Contains(out.String(), "continue_if_fast_target=40") {
		t.Fatalf("next-wave import output missing two-hour work budget: %s", out.String())
	}

	generatedPrompt, err := os.ReadFile(filepath.Join(tmpOutDir, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	recordedPrompt, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatal(err)
	}
	generatedPromptText := normalizeRecommendationPromptFixture(string(generatedPrompt))
	recordedPromptText := normalizeRecommendationPromptFixture(string(recordedPrompt))
	if generatedPromptText != recordedPromptText {
		t.Fatalf("v02 next-wave prompt fixture changed\nwant digest %s\ngot  digest %s", digestValue(recordedPromptText), digestValue(generatedPromptText))
	}
	prompt := recordedPromptText
	for _, want := range []string{
		"Target 2-3 hours",
		"min_minutes: 120",
		"max_minutes: 180",
		"Lease floor stop gate: do not return before min_minutes=120",
		"Feature Depth Recommendations, at least 40 tasks",
		"If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v02 next-wave prompt missing required budget/stop-gate language %q:\n%s", want, prompt)
		}
	}

	recordedWave := mustLoadJSON[AtlasRecommendationWave](t, recordedWavePath)
	generatedWave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(tmpOutDir, "recommendation-wave.json"))
	if digestValue(generatedWave) != digestValue(recordedWave) {
		t.Fatalf("v02 next-wave recommendation wave fixture changed\nwant %s\ngot  %s", digestValue(recordedWave), digestValue(generatedWave))
	}
	if err := ValidateAtlasRecommendationWave(recordedWave); err != nil {
		t.Fatal(err)
	}
	if recordedWave.TotalTasks != 40 ||
		recordedWave.MinimumTasks != 40 ||
		recordedWave.NodeBudget != 40 ||
		recordedWave.EstimatedMinutes != 150 ||
		recordedWave.Supervisor == nil ||
		recordedWave.Supervisor.MinMinutes != 120 ||
		recordedWave.Supervisor.MaxMinutes != 180 ||
		recordedWave.Supervisor.ContinueIfFastTarget != 40 ||
		recordedWave.FinalResponseAllowed ||
		recordedWave.SafeToExecute ||
		recordedWave.SchedulesWork ||
		recordedWave.ExecutesWork ||
		recordedWave.ApprovesWork {
		t.Fatalf("v02 next-wave prompt generation lost 40-node/two-hour planning-only contract: %#v", recordedWave)
	}
}

func TestRecommendationPromptGenerationUsesStructuredWaveBudgetAndStopConditions(t *testing.T) {
	wave := AtlasRecommendationWave{
		MissionID:            "ao-atlas-structured-budget-test",
		TargetInstance:       "ao-atlas-structured-budget-test-v01",
		TotalTasks:           40,
		MinimumTasks:         30,
		NodeBudget:           40,
		EstimatedMinutes:     120,
		FinalResponseAllowed: false,
		FinalResponseReason:  "ready_nodes_or_exact_next_action_remain",
		SourceDigest:         "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		Supervisor: &AtlasLongRunSupervisor{
			MinNodes:             30,
			MinMinutes:           120,
			MaxMinutes:           180,
			ContinueIfFastTarget: 40,
			ReturnOnlyWhen:       "all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker",
			CheckpointPolicy:     "after_each_node_with_pr_ci_merge_cleanup",
		},
		Tasks: []AtlasRecommendationTask{
			{ID: "next-01", Task: "Structured budget prompt test node.", Owner: "ao-atlas"},
		},
	}
	budget := BuildAtlasRecommendationPromptBudget(wave)
	if budget.MinNodes != 30 ||
		budget.MinMinutes != 120 ||
		budget.MaxMinutes != 180 ||
		budget.MaxIterations != 40 ||
		budget.ReturnOnlyWhen != "all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker" ||
		budget.CheckpointPolicy != "after_each_node_with_pr_ci_merge_cleanup" ||
		len(budget.StopConditions) != 5 {
		t.Fatalf("structured prompt budget lost supervisor fields: %#v", budget)
	}
	prompt := buildAtlasRecommendationPrompt(wave)
	for _, want := range []string{
		"min_nodes: 30",
		"min_minutes: 120",
		"max_minutes: 180",
		"max_iterations: 40",
		"return_only_when: all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker",
		"checkpoint_policy: after_each_node_with_pr_ci_merge_cleanup",
		"Lease floor stop gate: do not return before min_minutes=120 unless a true hard blocker remains.",
		"Ready-work stop gate: if ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("structured prompt budget missing %q:\n%s", want, prompt)
		}
	}
}

func TestFeatureDepthWaveImportRejectsCompletedFollowupRecommendationSource(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", "docs/evidence/ao-atlas-feature-depth-followup-durability-v04/source-feature-depth-recommendations.json",
		"--target-instance", "ao-atlas-feature-depth-followup-durability-v05",
		"--min-tasks", "40",
		"--node-budget", "40",
		"--out", filepath.Join(t.TempDir(), "generated-next-wave"),
	}, &out, &out)
	if code == 0 {
		t.Fatalf("stale completed Feature Depth follow-up recommendations were imported again: %s", out.String())
	}
	for _, want := range []string{
		"feature depth recommendations saturated",
		"completed 40/40",
		"route to AO Atlas refactoring/strategy review",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("saturated import refusal missing %q: %s", want, out.String())
		}
	}
}

func normalizeRecommendationPromptFixture(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}
