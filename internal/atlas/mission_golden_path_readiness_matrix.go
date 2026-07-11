package atlas

import "fmt"

func BuildAtlasGoldenPathReadinessMatrix() (AtlasGoldenPathReadinessMatrix, error) {
	recs := make([]AtlasGoldenPathRecommendation, 40)
	for i := range recs {
		recs[i] = AtlasGoldenPathRecommendation{
			Rank: fmt.Sprintf("%02d", i+1),
			Task: fmt.Sprintf("golden-path-followup-%02d", i+1),
		}
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
	validateNoAuthorityEffects(&errs, false, false, false, matrix.ClaimsAuthorityAdvance, matrix.RSIRemainsDenied)
	return joinErrors(errs)
}
