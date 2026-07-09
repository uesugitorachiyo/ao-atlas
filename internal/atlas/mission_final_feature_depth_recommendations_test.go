package atlas

import (
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveFinalOperatorHandoffRecommendations(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-40", "final-feature-depth-recommendations.json")
	recorded := mustLoadJSON[finalFeatureDepthRecommendationHandoff](t, fixturePath)
	generated := buildFinalFeatureDepthRecommendationHandoff(t, root, "ao-atlas-feature-depth-wave-v01")
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("final Feature Depth recommendation handoff fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	validateFinalFeatureDepthRecommendationHandoff(t, recorded)
}

func TestFeatureDepthWaveV02FinalOperatorHandoffRecommendations(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02", "nodes", "mission-recommendation-feature-depth-next-wave-40", "final-feature-depth-recommendations.json")
	recorded := mustLoadJSON[finalFeatureDepthRecommendationHandoff](t, fixturePath)
	generated := buildFinalFeatureDepthRecommendationHandoff(t, root, "ao-atlas-feature-depth-wave-v02")
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 final Feature Depth recommendation handoff fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	validateFinalFeatureDepthRecommendationHandoff(t, recorded)
}

type finalFeatureDepthRecommendationHandoff struct {
	Schema                            string                            `json:"schema"`
	NodeID                            string                            `json:"node_id"`
	Status                            string                            `json:"status"`
	SourceRecommendationsPath         string                            `json:"source_recommendations_path"`
	SourceReadbackPath                string                            `json:"source_readback_path"`
	CompletedNodesBefore              int                               `json:"completed_nodes_before"`
	ReadyNodesBefore                  int                               `json:"ready_nodes_before"`
	ExpectedCompletedNodesAfter       int                               `json:"expected_completed_nodes_after"`
	ExpectedReadyNodesAfter           int                               `json:"expected_ready_nodes_after"`
	ExpectedFinalResponseAllowedAfter bool                              `json:"expected_final_response_allowed_after"`
	TotalRecommendationCount          int                               `json:"total_recommendation_count"`
	OperatorReviewRecommendationCount int                               `json:"operator_review_recommendation_count"`
	MinimumOperatorRecommendations    int                               `json:"minimum_operator_recommendations"`
	Recommendations                   []finalFeatureDepthRecommendation `json:"recommendations"`
	NoPromotionRequested              bool                              `json:"no_promotion_requested"`
	PromotionGranted                  bool                              `json:"promotion_granted"`
	ClaimsAuthorityAdvance            bool                              `json:"claims_authority_advance"`
	RSIRemainsDenied                  bool                              `json:"rsi_remains_denied"`
	SafeToExecute                     bool                              `json:"safe_to_execute"`
	SchedulesWork                     bool                              `json:"schedules_work"`
	ExecutesWork                      bool                              `json:"executes_work"`
	ApprovesWork                      bool                              `json:"approves_work"`
	MutatesRepositories               bool                              `json:"mutates_repositories"`
}

type finalFeatureDepthRecommendation struct {
	Rank  int    `json:"rank"`
	ID    string `json:"id"`
	Theme string `json:"theme"`
	Task  string `json:"task"`
}

func buildFinalFeatureDepthRecommendationHandoff(t *testing.T, root string, wave string) finalFeatureDepthRecommendationHandoff {
	t.Helper()
	sourceRecommendationsPath := filepath.ToSlash(filepath.Join("docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-37", "next-wave-feature-depth-recommendations.json"))
	sourceReadbackPath := filepath.ToSlash(filepath.Join("docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-39", "recommendation-readback-after.json"))
	source := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, filepath.Join(root, sourceRecommendationsPath))
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(source, 40); err != nil {
		t.Fatal(err)
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, sourceReadbackPath))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	recommendations := make([]finalFeatureDepthRecommendation, 0, 10)
	for _, item := range source.Tasks {
		if len(recommendations) == 10 {
			break
		}
		recommendations = append(recommendations, finalFeatureDepthRecommendation{
			Rank:  item.Rank,
			ID:    item.ID,
			Theme: item.Theme,
			Task:  item.Task,
		})
	}
	return finalFeatureDepthRecommendationHandoff{
		Schema:                            "ao.atlas.final-feature-depth-recommendations.v0.1",
		NodeID:                            "mission-recommendation-feature-depth-next-wave-40",
		Status:                            "ready_for_operator_review",
		SourceRecommendationsPath:         filepath.ToSlash(sourceRecommendationsPath),
		SourceReadbackPath:                filepath.ToSlash(sourceReadbackPath),
		CompletedNodesBefore:              readback.CompletedNodes,
		ReadyNodesBefore:                  readback.ReadyNodes,
		ExpectedCompletedNodesAfter:       40,
		ExpectedReadyNodesAfter:           0,
		ExpectedFinalResponseAllowedAfter: true,
		TotalRecommendationCount:          source.RecommendationCount,
		OperatorReviewRecommendationCount: len(recommendations),
		MinimumOperatorRecommendations:    10,
		Recommendations:                   recommendations,
		NoPromotionRequested:              true,
		PromotionGranted:                  false,
		ClaimsAuthorityAdvance:            false,
		RSIRemainsDenied:                  true,
		SafeToExecute:                     false,
		SchedulesWork:                     false,
		ExecutesWork:                      false,
		ApprovesWork:                      false,
		MutatesRepositories:               false,
	}
}

func validateFinalFeatureDepthRecommendationHandoff(t *testing.T, fixture finalFeatureDepthRecommendationHandoff) {
	t.Helper()
	if fixture.Schema != "ao.atlas.final-feature-depth-recommendations.v0.1" ||
		fixture.NodeID != "mission-recommendation-feature-depth-next-wave-40" ||
		fixture.Status != "ready_for_operator_review" ||
		fixture.CompletedNodesBefore != 39 ||
		fixture.ReadyNodesBefore != 1 ||
		fixture.ExpectedCompletedNodesAfter != 40 ||
		fixture.ExpectedReadyNodesAfter != 0 ||
		!fixture.ExpectedFinalResponseAllowedAfter ||
		fixture.TotalRecommendationCount < 40 ||
		fixture.OperatorReviewRecommendationCount < fixture.MinimumOperatorRecommendations ||
		len(fixture.Recommendations) != fixture.OperatorReviewRecommendationCount ||
		!fixture.NoPromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories {
		t.Fatalf("final Feature Depth handoff lost recommendation or safety contract: %#v", fixture)
	}
	for i, item := range fixture.Recommendations {
		if item.Rank != i+1 || item.ID == "" || item.Theme == "" || item.Task == "" {
			t.Fatalf("recommendation %d is not ranked and review-ready: %#v", i, item)
		}
	}
}
