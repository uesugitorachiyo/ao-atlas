package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveV02SeedImportsFortyRecommendationNodes(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")

	source := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, filepath.Join(waveRoot, "source-feature-depth-recommendations.json"))
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(source, 40); err != nil {
		t.Fatal(err)
	}
	if source.MissionID != "ao-atlas-next-feature-depth-wave-v02" ||
		source.RecommendationCount != 40 ||
		source.MinimumTasks != 40 ||
		source.Status != "ready" ||
		source.SafeToExecute ||
		source.SchedulesWork ||
		source.ExecutesWork ||
		source.ApprovesWork ||
		source.MutatesRepositories {
		t.Fatalf("v02 source recommendations must be a 40-task planning-only bundle: %#v", source)
	}

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(waveRoot, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "ao-atlas-next-feature-depth-wave-v02" ||
		wave.TargetInstance != "ao-atlas-feature-depth-wave-v02" ||
		wave.TotalTasks != 40 ||
		wave.MinimumTasks != 40 ||
		wave.NodeBudget != 40 ||
		wave.EstimatedMinutes != 150 ||
		wave.FinalResponseAllowed ||
		wave.SafeToExecute ||
		wave.SchedulesWork ||
		wave.ExecutesWork ||
		wave.ApprovesWork {
		t.Fatalf("v02 wave must preserve long-run planning-only import state: %#v", wave)
	}
	if wave.Supervisor == nil ||
		wave.Supervisor.MinNodes != 40 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 ||
		wave.Supervisor.ReturnOnlyWhen != "all_40_feature_depth_nodes_complete_or_true_hard_blocker" {
		t.Fatalf("v02 wave must preserve 2-3 hour supervisor contract: %#v", wave.Supervisor)
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(waveRoot, "recommendation-readback.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 0 ||
		readback.ReadyNodes != 40 ||
		readback.BlockedNodes != 0 ||
		readback.FailedNodes != 0 ||
		readback.FinalResponseAllowed ||
		readback.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-01" ||
		readback.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		!readback.ContinuationContract.RefusesFinalResponse ||
		readback.ContinuationContract.Reason != "ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("v02 readback must require continuation across all ready nodes: %#v", readback)
	}

	summary := mustLoadJSON[featureDepthWaveV02SeedSummary](t, filepath.Join(waveRoot, "feature-depth-wave-v02-seed-summary.json"))
	validateFeatureDepthWaveV02SeedSummary(t, summary)

	promptBytes, err := os.ReadFile(filepath.Join(waveRoot, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Mission: ao-atlas-next-feature-depth-wave-v02",
		"Target instance: ao-atlas-feature-depth-wave-v02",
		"Lease minimum: 40 nodes, 120 to 180 minutes.",
		"Do not ask the operator for permission.",
		"feature-depth-next-wave-40. Generate final Feature Depth recommendations for operator handoff review.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v02 prompt missing %q", want)
		}
	}
}

type featureDepthWaveV02SeedSummary struct {
	Schema                    string `json:"schema"`
	Status                    string `json:"status"`
	SourceWaveRoot            string `json:"source_wave_root"`
	SourceRecommendationsPath string `json:"source_recommendations_path"`
	SourceReadbackPath        string `json:"source_readback_path"`
	TargetWaveRoot            string `json:"target_wave_root"`
	MissionID                 string `json:"mission_id"`
	TargetInstance            string `json:"target_instance"`
	RecommendationCount       int    `json:"recommendation_count"`
	ReadyNodes                int    `json:"ready_nodes"`
	CompletedNodes            int    `json:"completed_nodes"`
	FinalResponseAllowed      bool   `json:"final_response_allowed"`
	FirstExecutableNode       string `json:"first_executable_node"`
	ExactNextAction           string `json:"exact_next_action"`
	NoPromotionRequested      bool   `json:"no_promotion_requested"`
	PromotionGranted          bool   `json:"promotion_granted"`
	ClaimsAuthorityAdvance    bool   `json:"claims_authority_advance"`
	RSIRemainsDenied          bool   `json:"rsi_remains_denied"`
	SafeToExecute             bool   `json:"safe_to_execute"`
	SchedulesWork             bool   `json:"schedules_work"`
	ExecutesWork              bool   `json:"executes_work"`
	ApprovesWork              bool   `json:"approves_work"`
	MutatesRepositories       bool   `json:"mutates_repositories"`
}

func validateFeatureDepthWaveV02SeedSummary(t *testing.T, summary featureDepthWaveV02SeedSummary) {
	t.Helper()
	if summary.Schema != "ao.atlas.feature-depth-wave-v02-seed-summary.v0.1" ||
		summary.Status != "ready_for_long_run_execution" ||
		summary.SourceWaveRoot != "docs/evidence/ao-atlas-feature-depth-wave-v01" ||
		summary.SourceRecommendationsPath != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-37/next-wave-feature-depth-recommendations.json" ||
		summary.SourceReadbackPath != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json" ||
		summary.TargetWaveRoot != "docs/evidence/ao-atlas-feature-depth-wave-v02" ||
		summary.MissionID != "ao-atlas-next-feature-depth-wave-v02" ||
		summary.TargetInstance != "ao-atlas-feature-depth-wave-v02" ||
		summary.RecommendationCount != 40 ||
		summary.ReadyNodes != 40 ||
		summary.CompletedNodes != 0 ||
		summary.FinalResponseAllowed ||
		summary.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-01" ||
		!strings.Contains(summary.ExactNextAction, "mission-recommendation-feature-depth-next-wave-01") ||
		!summary.NoPromotionRequested ||
		summary.PromotionGranted ||
		summary.ClaimsAuthorityAdvance ||
		!summary.RSIRemainsDenied ||
		summary.SafeToExecute ||
		summary.SchedulesWork ||
		summary.ExecutesWork ||
		summary.ApprovesWork ||
		summary.MutatesRepositories {
		t.Fatalf("v02 seed summary lost long-run or safety contract: %#v", summary)
	}
}
