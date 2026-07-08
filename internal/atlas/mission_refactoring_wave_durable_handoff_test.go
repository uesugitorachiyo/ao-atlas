package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactoringWaveDurableHandoffFixtureRoutesPastCompletedFeatureDepth(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01")
	decision := mustLoadJSON[map[string]any](t, filepath.Join(waveRoot, "next-track-decision.json"))
	recommendations := mustLoadJSON[AOMissionRefactoringRecommendations](t, filepath.Join(waveRoot, "refactoring-recommendations.json"))

	if decision["schema"] != "ao.atlas.recommendation-next-track-decision.v0.1" ||
		decision["status"] != "routed" ||
		decision["current_track"] != "feature_depth" ||
		decision["current_track_status"] != "completed_saturated" ||
		decision["recommended_track"] != "refactoring" ||
		decision["rsi_track_status"] != "boundary_hardening_only_denied" {
		t.Fatalf("next-track decision did not route completed Feature Depth to refactoring: %#v", decision)
	}
	exactNextAction, _ := decision["exact_next_action"].(string)
	if !strings.Contains(exactNextAction, "Start AO Atlas refactoring wave") ||
		strings.Contains(strings.ToLower(exactNextAction), "feature depth") {
		t.Fatalf("next-track decision should start refactoring without looping to Feature Depth: %q", exactNextAction)
	}

	if err := ValidateAtlasNextWaveRefactoringRecommendations(recommendations, 40); err != nil {
		t.Fatal(err)
	}
	if recommendations.MissionID != "ao-atlas-refactoring-wave-v01" ||
		recommendations.Track != "refactoring" ||
		recommendations.MinimumTasks != 40 ||
		recommendations.RecommendationCount != 40 ||
		len(recommendations.Tasks) != 40 ||
		!recommendations.NoPromotionRequested ||
		recommendations.PromotionGranted ||
		recommendations.ClaimsAuthorityAdvance ||
		!recommendations.RSIRemainsDenied ||
		recommendations.SafeToExecute ||
		recommendations.SchedulesWork ||
		recommendations.ExecutesWork ||
		recommendations.ApprovesWork ||
		recommendations.MutatesRepositories {
		t.Fatalf("refactoring recommendations lost durable planning-only boundaries: %#v", recommendations)
	}

	themes := map[string]bool{}
	for _, task := range recommendations.Tasks {
		if !strings.HasPrefix(task.ID, "refactoring-next-wave-") {
			t.Fatalf("refactoring task id should be wave-scoped, got %q", task.ID)
		}
		if task.Owner != "ao-atlas" {
			t.Fatalf("refactoring task owner should stay AO Atlas, got %#v", task)
		}
		if strings.Contains(task.ID, "feature-depth") || strings.Contains(task.Task, "feature-depth-next-wave") {
			t.Fatalf("refactoring task should not schedule another Feature Depth node: %#v", task)
		}
		themes[task.Theme] = true
	}
	if len(themes) < 10 {
		t.Fatalf("expected at least 10 refactoring themes, got %d: %#v", len(themes), themes)
	}

	prompt, err := os.ReadFile(filepath.Join(waveRoot, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	promptText := string(prompt)
	for _, want := range []string{
		"ao-atlas-refactoring-wave-v01",
		"refactoring-recommendations.json",
		"Target 2-3 hours",
		"Complete at least 12 bounded refactoring nodes",
		"Do not stop after one node",
		"refactoring-next-wave-40",
		"RSI remains denied",
	} {
		if !strings.Contains(promptText, want) {
			t.Fatalf("next recommended prompt missing %q", want)
		}
	}
}
