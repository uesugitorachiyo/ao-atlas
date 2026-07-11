package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasGoldenPathReadinessMatrix() (AtlasGoldenPathReadinessMatrix, error) {
	tasks := goldenPathReadinessRecommendations()
	recs := make([]AtlasGoldenPathRecommendation, 0, len(tasks))
	for i, task := range tasks {
		recs = append(recs, AtlasGoldenPathRecommendation{
			Rank: fmt.Sprintf("%02d", i+1),
			Task: task,
		})
	}
	matrix := AtlasGoldenPathReadinessMatrix{
		Schema:                    AtlasGoldenPathReadinessMatrixContract,
		Status:                    "golden_path_readiness_matrix_ready",
		CompletedNodes:            40,
		ReadyNodes:                0,
		BlockedNodes:              0,
		FailedNodes:               0,
		ProvenCapabilities:        []string{"bounded_fixture_replay", "readback_binding", "run_link_digest_verification", "no_promotion_rollup"},
		UnresolvedBlockers:        []string{"real provider-backed golden path remains outside this fixture-only wave"},
		NoPromotionStatus:         "no_promotion_requested",
		PromotionRequested:        false,
		RankedRecommendations:     recs,
		RankedRecommendationCount: len(recs),
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err != nil {
		return AtlasGoldenPathReadinessMatrix{}, err
	}
	return matrix, nil
}

func ValidateAtlasGoldenPathReadinessMatrix(matrix AtlasGoldenPathReadinessMatrix) error {
	var errs []string
	requireContract(&errs, "golden_path_readiness_matrix", matrix.Schema, AtlasGoldenPathReadinessMatrixContract)
	if matrix.Status != "golden_path_readiness_matrix_ready" {
		errs = append(errs, "status must be golden_path_readiness_matrix_ready")
	}
	if matrix.CompletedNodes != 40 || matrix.ReadyNodes != 0 || matrix.BlockedNodes != 0 || matrix.FailedNodes != 0 {
		errs = append(errs, "terminal node counts must be 40 completed and zero ready/blocked/failed")
	}
	if matrix.NoPromotionStatus != "no_promotion_requested" {
		errs = append(errs, "no_promotion_status must be no_promotion_requested")
	}
	if matrix.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if matrix.RankedRecommendationCount != len(matrix.RankedRecommendations) || matrix.RankedRecommendationCount < 40 {
		errs = append(errs, "ranked_recommendation_count must be at least 40")
	}
	for i, rec := range matrix.RankedRecommendations {
		prefix := fmt.Sprintf("ranked_recommendations[%d]", i)
		requireField(&errs, prefix+".task", rec.Task)
		if strings.Contains(rec.Task, "golden-path-followup-") {
			errs = append(errs, prefix+".task must be domain-specific, not a placeholder")
		}
		if len(strings.Fields(rec.Task)) < 6 {
			errs = append(errs, prefix+".task must be actionable")
		}
	}
	validateNoAuthorityEffects(&errs, false, false, false, matrix.ClaimsAuthorityAdvance, matrix.RSIRemainsDenied)
	return joinErrors(errs)
}

func goldenPathReadinessRecommendations() []string {
	return []string{
		"Add aggregate Promoter Command public-safety closure rollup over all forty Month 3 nodes.",
		"Bind terminal readback counts to the final readiness matrix with digest evidence.",
		"Add non-AO repository dry-run replay binding with fixture-only execution evidence.",
		"Define real-run acceptance criteria for three external non-AO repositories.",
		"Add control-plane observer readback adapter contract fixtures for mission timeline evidence.",
		"Draft Covenant-owned schema owner registry proposal with consumer compatibility checks.",
		"Add evidence externalization plan for large generated JSON artifacts and retained golden fixtures.",
		"Add cross-repo CI compatibility matrix for Mission Atlas Foundry Covenant Command and AO2.",
		"Add operator dashboard readback fixture for terminal golden-path status and blockers.",
		"Add restart resume soak-test fixture for long-running node waves and checkpoint recovery.",
		"Add provider and model provenance fields to every model-backed run record fixture.",
		"Add rollback receipt replay negative cases for stale base commits and digest mismatches.",
		"Add compaction-resume prompt generator from the terminal golden-path readback.",
		"Add Architecture source-of-truth correction checklist for current authority statements.",
		"Add no-promotion no-RSI assertion matrix across all terminal Month 3 artifacts.",
		"Add workspace-root preflight evidence for the real golden-path dry-run repository layout.",
		"Add Command thin-client boundary fixture that rejects duplicated domain authority.",
		"Add Foundry safe-next-work selection fixture bound to terminal readiness evidence.",
		"Add Mission recovery invariant fixture for no handoff counted as completed work.",
		"Add Blueprint authorization digest preservation fixture for downstream Atlas imports.",
		"Add AO2 approval integrity checklist binding proposed bytes and base commit.",
		"Add Covenant policy hash replay fixture for policy fields that affect acceptance.",
		"Add Sentinel freshness signal fixture using CI and evidence age instead of README wording.",
		"Add Promoter no-activation boundary fixture requiring signed assurance before promotion.",
		"Add Forge GoalRun lifecycle fixture that stays bounded without provider execution.",
		"Add control-plane durable index migration fixture for mission event search.",
		"Add rollback under failure drill plan with observer receipt and operator readback.",
		"Add public-safety wording scan for authority claims in generated summaries.",
		"Add stack lockfile draft binding repository heads without release or tag creation.",
		"Add schema compatibility downgrade fixture for deprecated contract consumers.",
		"Add golden-path PR ledger binding CI status merge head and branch cleanup.",
		"Add Windows CI wait-state telemetry fixture for long-running readiness checks.",
		"Add local platform capability fixture for sandbox and toolchain prerequisites.",
		"Add content-addressed evidence manifest fixture for externalized artifacts.",
		"Add dashboard compact filters for completed ready blocked failed and stale nodes.",
		"Add operator summary generator that reads only terminal validated evidence.",
		"Add final closure handoff prompt generator with no-promotion and RSI denial.",
		"Add replayable state packet fixture covering kill restart and resume ordering.",
		"Add command readback agreement fixture for Promoter no-promotion verdicts.",
		"Add final real-golden-path readiness report with proven capabilities and blockers.",
	}
}
