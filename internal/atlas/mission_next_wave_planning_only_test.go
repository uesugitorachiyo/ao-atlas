package atlas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveNextWaveRecommendationsRemainPlanningOnlyUntilImported(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-39", "next-wave-planning-only-validation.json")
	recorded := mustLoadJSON[nextWavePlanningOnlyValidationFixture](t, fixturePath)
	generated := buildNextWavePlanningOnlyValidationFixture(t, root, "ao-atlas-feature-depth-wave-v01")
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("next-wave planning-only validation fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	validateNextWavePlanningOnlyValidationFixture(t, recorded)
}

func TestFeatureDepthWaveV02NextWaveRecommendationsRemainPlanningOnlyUntilImported(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02", "nodes", "mission-recommendation-feature-depth-next-wave-39", "next-wave-planning-only-validation.json")
	recorded := mustLoadJSON[nextWavePlanningOnlyValidationFixture](t, fixturePath)
	generated := buildNextWavePlanningOnlyValidationFixture(t, root, "ao-atlas-feature-depth-wave-v02")
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 next-wave planning-only validation fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	validateNextWavePlanningOnlyValidationFixture(t, recorded)
}

type nextWavePlanningOnlyValidationFixture struct {
	Schema                               string          `json:"schema"`
	NodeID                               string          `json:"node_id"`
	Status                               string          `json:"status"`
	SourceRecommendationsPath            string          `json:"source_recommendations_path"`
	GeneratedRecommendationWavePath      string          `json:"generated_recommendation_wave_path"`
	GeneratedRecommendationReadbackPath  string          `json:"generated_recommendation_readback_path"`
	GeneratedRecommendationWorkgraphPath string          `json:"generated_recommendation_workgraph_path"`
	SourceRecommendationCount            int             `json:"source_recommendation_count"`
	GeneratedNodeCount                   int             `json:"generated_node_count"`
	GeneratedReadyNodes                  int             `json:"generated_ready_nodes"`
	GeneratedCompletedNodes              int             `json:"generated_completed_nodes"`
	FirstExecutableNode                  string          `json:"first_executable_node"`
	ExactNextAction                      string          `json:"exact_next_action"`
	SourceRecommendationsPlanningOnly    bool            `json:"source_recommendations_planning_only"`
	GeneratedWavePlanningOnly            bool            `json:"generated_wave_planning_only"`
	GeneratedReadbackPlanningOnly        bool            `json:"generated_readback_planning_only"`
	GeneratedNextWaveFoundryImportAbsent bool            `json:"generated_next_wave_foundry_import_absent"`
	GeneratedNextWaveRunLinkAbsent       bool            `json:"generated_next_wave_run_link_absent"`
	ContinuationRefusesFinalResponse     bool            `json:"continuation_refuses_final_response"`
	ContinuationReason                   string          `json:"continuation_reason"`
	FinalResponseAllowed                 bool            `json:"final_response_allowed"`
	ClaimsAuthorityAdvance               bool            `json:"claims_authority_advance"`
	PromotionGranted                     bool            `json:"promotion_granted"`
	RSIRemainsDenied                     bool            `json:"rsi_remains_denied"`
	SafeToExecute                        bool            `json:"safe_to_execute"`
	SchedulesWork                        bool            `json:"schedules_work"`
	ExecutesWork                         bool            `json:"executes_work"`
	ApprovesWork                         bool            `json:"approves_work"`
	MutatesRepositories                  bool            `json:"mutates_repositories"`
	SafetyBoundaries                     map[string]bool `json:"safety_boundaries"`
}

func buildNextWavePlanningOnlyValidationFixture(t *testing.T, root string, waveName string) nextWavePlanningOnlyValidationFixture {
	t.Helper()
	sourcePath := filepath.ToSlash(filepath.Join("docs", "evidence", waveName, "nodes", "mission-recommendation-feature-depth-next-wave-37", "next-wave-feature-depth-recommendations.json"))
	generatedWavePath := filepath.ToSlash(filepath.Join("docs", "evidence", waveName, "nodes", "mission-recommendation-feature-depth-next-wave-38", "generated-next-wave", "recommendation-wave.json"))
	generatedReadbackPath := filepath.ToSlash(filepath.Join("docs", "evidence", waveName, "nodes", "mission-recommendation-feature-depth-next-wave-38", "generated-next-wave", "recommendation-readback.json"))
	generatedWorkgraphPath := filepath.ToSlash(filepath.Join("docs", "evidence", waveName, "nodes", "mission-recommendation-feature-depth-next-wave-38", "generated-next-wave", "recommendation-workgraph.json"))

	source := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, filepath.Join(root, sourcePath))
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(source, 40); err != nil {
		t.Fatal(err)
	}
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, generatedWavePath))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, generatedReadbackPath))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}

	sourcePlanningOnly := source.Status == "ready" &&
		source.RecommendationCount >= 40 &&
		!source.SafeToExecute &&
		!source.SchedulesWork &&
		!source.ExecutesWork &&
		!source.ApprovesWork &&
		!source.MutatesRepositories
	wavePlanningOnly := wave.Status == "ready" &&
		wave.TotalTasks == 40 &&
		wave.MinimumTasks == 40 &&
		!wave.SafeToExecute &&
		!wave.SchedulesWork &&
		!wave.ExecutesWork &&
		!wave.ApprovesWork
	readbackPlanningOnly := readback.CompletedNodes == 0 &&
		readback.ReadyNodes == 40 &&
		!readback.FinalResponseAllowed &&
		readback.ContinuationContract.RefusesFinalResponse &&
		readback.ContinuationContract.Reason == "ready_nodes_or_exact_next_action_remain" &&
		!readback.SchedulesWork &&
		!readback.ExecutesWork &&
		!readback.ApprovesWork
	generatedRoot := filepath.Dir(generatedWavePath)
	foundryImportAbsent := fileAbsent(t, filepath.Join(root, generatedRoot, "foundry-import.json"))
	runLinkAbsent := fileAbsent(t, filepath.Join(root, generatedRoot, "run-link.json"))
	status := "passed"
	if !sourcePlanningOnly || !wavePlanningOnly || !readbackPlanningOnly || !foundryImportAbsent || !runLinkAbsent {
		status = "failed"
	}

	return nextWavePlanningOnlyValidationFixture{
		Schema:                               "ao.atlas.next-wave-planning-only-validation.v0.1",
		NodeID:                               "mission-recommendation-feature-depth-next-wave-39",
		Status:                               status,
		SourceRecommendationsPath:            filepath.ToSlash(sourcePath),
		GeneratedRecommendationWavePath:      filepath.ToSlash(generatedWavePath),
		GeneratedRecommendationReadbackPath:  filepath.ToSlash(generatedReadbackPath),
		GeneratedRecommendationWorkgraphPath: filepath.ToSlash(generatedWorkgraphPath),
		SourceRecommendationCount:            source.RecommendationCount,
		GeneratedNodeCount:                   wave.TotalTasks,
		GeneratedReadyNodes:                  readback.ReadyNodes,
		GeneratedCompletedNodes:              readback.CompletedNodes,
		FirstExecutableNode:                  readback.FirstExecutableNode,
		ExactNextAction:                      readback.ExactNextAction,
		SourceRecommendationsPlanningOnly:    sourcePlanningOnly,
		GeneratedWavePlanningOnly:            wavePlanningOnly,
		GeneratedReadbackPlanningOnly:        readbackPlanningOnly,
		GeneratedNextWaveFoundryImportAbsent: foundryImportAbsent,
		GeneratedNextWaveRunLinkAbsent:       runLinkAbsent,
		ContinuationRefusesFinalResponse:     readback.ContinuationContract.RefusesFinalResponse,
		ContinuationReason:                   readback.ContinuationContract.Reason,
		FinalResponseAllowed:                 readback.FinalResponseAllowed,
		ClaimsAuthorityAdvance:               false,
		PromotionGranted:                     false,
		RSIRemainsDenied:                     true,
		SafeToExecute:                        false,
		SchedulesWork:                        false,
		ExecutesWork:                         false,
		ApprovesWork:                         false,
		MutatesRepositories:                  false,
		SafetyBoundaries: map[string]bool{
			"no_provider_calls":              true,
			"no_credential_inspection":       true,
			"no_direct_main_mutation":        true,
			"no_release_deploy_publish_tag":  true,
			"no_dependency_updates":          true,
			"no_auth_policy_config_widening": true,
			"rsi_remains_denied":             true,
		},
	}
}

func fileAbsent(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	t.Fatal(err)
	return false
}

func validateNextWavePlanningOnlyValidationFixture(t *testing.T, fixture nextWavePlanningOnlyValidationFixture) {
	t.Helper()
	if fixture.Schema != "ao.atlas.next-wave-planning-only-validation.v0.1" ||
		fixture.NodeID != "mission-recommendation-feature-depth-next-wave-39" ||
		fixture.Status != "passed" ||
		fixture.SourceRecommendationCount < 40 ||
		fixture.GeneratedNodeCount != 40 ||
		fixture.GeneratedReadyNodes != 40 ||
		fixture.GeneratedCompletedNodes != 0 ||
		fixture.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-01" ||
		fixture.ExactNextAction == "" ||
		!fixture.SourceRecommendationsPlanningOnly ||
		!fixture.GeneratedWavePlanningOnly ||
		!fixture.GeneratedReadbackPlanningOnly ||
		!fixture.GeneratedNextWaveFoundryImportAbsent ||
		!fixture.GeneratedNextWaveRunLinkAbsent ||
		!fixture.ContinuationRefusesFinalResponse ||
		fixture.ContinuationReason != "ready_nodes_or_exact_next_action_remain" ||
		fixture.FinalResponseAllowed ||
		fixture.ClaimsAuthorityAdvance ||
		fixture.PromotionGranted ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories {
		t.Fatalf("next-wave planning-only fixture lost safety/readback binding: %#v", fixture)
	}
	for key, value := range fixture.SafetyBoundaries {
		if !value {
			t.Fatalf("safety boundary %s is not asserted", key)
		}
	}
}
