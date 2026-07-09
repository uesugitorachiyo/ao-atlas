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
		!digestPattern.MatchString(recommendations.SourceReadbackDigest) ||
		recommendations.ConsumedLedgerPath != "docs/evidence/ao-atlas-refactoring-wave-v01/consumed-recommendation-ledger.json" ||
		!digestPattern.MatchString(recommendations.ConsumedLedgerDigest) ||
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
		"Feature Depth final readback digest: " + recommendations.SourceReadbackDigest,
		"Consumed recommendation ledger: " + recommendations.ConsumedLedgerPath,
		"Consumed recommendation ledger digest: " + recommendations.ConsumedLedgerDigest,
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

func TestRefactoringWaveLongRunPromptRegressionFixturePreservesTwoToThreeHourBudget(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01")
	fixture := mustLoadJSON[struct {
		Schema                 string   `json:"schema"`
		Status                 string   `json:"status"`
		PromptPath             string   `json:"prompt_path"`
		PromptDigest           string   `json:"prompt_digest"`
		RecommendationsPath    string   `json:"recommendations_path"`
		RecommendationsDigest  string   `json:"recommendations_digest"`
		MinimumNodes           int      `json:"minimum_nodes"`
		TotalNodes             int      `json:"total_nodes"`
		MinMinutes             int      `json:"min_minutes"`
		MaxMinutes             int      `json:"max_minutes"`
		FinalResponseAllowed   bool     `json:"final_response_allowed"`
		RegressionAssertions   []string `json:"regression_assertions"`
		NoPromotionRequested   bool     `json:"no_promotion_requested"`
		PromotionGranted       bool     `json:"promotion_granted"`
		ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
		RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
		SafeToExecute          bool     `json:"safe_to_execute"`
		SchedulesWork          bool     `json:"schedules_work"`
		ExecutesWork           bool     `json:"executes_work"`
		ApprovesWork           bool     `json:"approves_work"`
		MutatesRepositories    bool     `json:"mutates_repositories"`
	}](t, filepath.Join(waveRoot, "long-run-prompt-regression.json"))

	promptPath := filepath.Join(root, filepath.FromSlash(fixture.PromptPath))
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	recommendationsPath := filepath.Join(root, filepath.FromSlash(fixture.RecommendationsPath))
	recommendationsDigest, err := digestFile(recommendationsPath)
	if err != nil {
		t.Fatal(err)
	}

	if fixture.Schema != "ao.atlas.refactoring-long-run-prompt-regression.v0.1" ||
		fixture.Status != "guarded" ||
		fixture.PromptPath != "docs/evidence/ao-atlas-refactoring-wave-v01/next-recommended-prompt.md" ||
		fixture.PromptDigest != digestBytes(promptBytes) ||
		fixture.RecommendationsPath != "docs/evidence/ao-atlas-refactoring-wave-v01/refactoring-recommendations.json" ||
		fixture.RecommendationsDigest != recommendationsDigest ||
		fixture.MinimumNodes != 12 ||
		fixture.TotalNodes != 40 ||
		fixture.MinMinutes != 120 ||
		fixture.MaxMinutes != 180 ||
		fixture.FinalResponseAllowed ||
		!fixture.NoPromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories {
		t.Fatalf("long-run prompt regression fixture lost budget or safety state: %#v", fixture)
	}
	promptText := string(promptBytes)
	for _, want := range []string{
		"Target 2-3 hours",
		"Complete at least 12 bounded refactoring nodes",
		"continue toward all 40 recommendations",
		"Do not stop after one node",
		"refactoring-next-wave-40",
		"RSI remains denied",
	} {
		if !strings.Contains(promptText, want) || !containsStringValue(fixture.RegressionAssertions, want) {
			t.Fatalf("long-run prompt regression fixture missing %q", want)
		}
	}
}
