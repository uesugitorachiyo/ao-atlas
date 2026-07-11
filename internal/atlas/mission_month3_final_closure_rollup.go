package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3FinalClosureRollup(nodeID, sourceReadbackPath, matrixPath, promoterPath, commandPath, publicSafetyPath string) (AtlasMonth3FinalClosureRollup, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3FinalClosureRollup{}, fmt.Errorf("node id is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	matrix, err := LoadJSON[AtlasGoldenPathReadinessMatrix](matrixPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	promoter, err := LoadJSON[AtlasNodePromoterNoPromotionEvidence](promoterPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	if err := ValidateAtlasNodePromoterNoPromotionEvidence(promoter); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	command, err := LoadJSON[AtlasNodeCommandReadbackEvidence](commandPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	if err := ValidateAtlasNodeCommandReadbackEvidence(command); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	publicSafety, err := LoadJSON[AtlasScopedPublicSafetyScan](publicSafetyPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	if err := ValidateAtlasScopedPublicSafetyScan(publicSafety); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	matrixDigest, err := digestTextFileWithNormalizedLineEndings(matrixPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	promoterDigest, err := digestTextFileWithNormalizedLineEndings(promoterPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	commandDigest, err := digestTextFileWithNormalizedLineEndings(commandPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	publicSafetyDigest, err := digestTextFileWithNormalizedLineEndings(publicSafetyPath)
	if err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}

	promotionRequested := matrix.PromotionRequested || promoter.PromotionClaimed
	agreement := readback.CompletedNodes == 40 &&
		readback.ReadyNodes == 0 &&
		readback.BlockedNodes == 0 &&
		readback.FailedNodes == 0 &&
		readback.FinalResponseAllowed &&
		matrix.CompletedNodes == readback.CompletedNodes &&
		matrix.ReadyNodes == readback.ReadyNodes &&
		matrix.BlockedNodes == readback.BlockedNodes &&
		matrix.FailedNodes == readback.FailedNodes &&
		matrix.NoPromotionStatus == "no_promotion_requested" &&
		promoter.Status == "no_promotion_requested" &&
		command.Status == "readback_agrees_no_promotion" &&
		publicSafety.Status == "passed" &&
		publicSafety.PublicSafetyScanPassed &&
		matrix.RSIRemainsDenied &&
		promoter.RSIRemainsDenied &&
		command.RSIRemainsDenied &&
		publicSafety.RSIRemainsDenied
	rollup := AtlasMonth3FinalClosureRollup{
		Schema:                         AtlasMonth3FinalClosureRollupContract,
		NodeID:                         nodeID,
		Status:                         "final_closure_rollup_bound",
		SourceReadbackPath:             publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:           readbackDigest,
		ReadinessMatrixPath:            publicArtifactRef(matrixPath),
		ReadinessMatrixDigest:          matrixDigest,
		PromoterPath:                   publicArtifactRef(promoterPath),
		PromoterDigest:                 promoterDigest,
		CommandPath:                    publicArtifactRef(commandPath),
		CommandDigest:                  commandDigest,
		PublicSafetyPath:               publicArtifactRef(publicSafetyPath),
		PublicSafetyDigest:             publicSafetyDigest,
		SourceCompletedNodes:           readback.CompletedNodes,
		SourceReadyNodes:               readback.ReadyNodes,
		SourceBlockedNodes:             readback.BlockedNodes,
		SourceFailedNodes:              readback.FailedNodes,
		SourceFinalResponseAllowed:     readback.FinalResponseAllowed,
		MatrixRecommendationCount:      matrix.RankedRecommendationCount,
		ProvenCapabilityCount:          len(matrix.ProvenCapabilities),
		UnresolvedBlockerCount:         len(matrix.UnresolvedBlockers),
		PromoterStatus:                 promoter.Status,
		CommandStatus:                  command.Status,
		PublicSafetyStatus:             publicSafety.Status,
		PublicSafetyScanPassed:         publicSafety.PublicSafetyScanPassed,
		AggregatePromotionStatus:       "no_promotion_requested",
		PromotionRequested:             promotionRequested,
		PromotionGranted:               false,
		TerminalReadbackBound:          readback.FinalResponseAllowed,
		MatrixBound:                    matrix.RankedRecommendationCount >= 40,
		PromoterCommandSafetyAgreement: agreement,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		ClaimsAuthorityAdvance:         matrix.ClaimsAuthorityAdvance || promoter.ClaimsAuthorityAdvance || publicSafety.ClaimsAuthorityAdvance,
		RSIRemainsDenied:               matrix.RSIRemainsDenied && promoter.RSIRemainsDenied && command.RSIRemainsDenied && publicSafety.RSIRemainsDenied,
	}
	if !rollup.PromoterCommandSafetyAgreement || rollup.PromotionRequested || rollup.ClaimsAuthorityAdvance || !rollup.RSIRemainsDenied {
		rollup.Status = "final_closure_rollup_failed"
	}
	if err := ValidateAtlasMonth3FinalClosureRollup(rollup); err != nil {
		return AtlasMonth3FinalClosureRollup{}, err
	}
	return rollup, nil
}

func ValidateAtlasMonth3FinalClosureRollup(rollup AtlasMonth3FinalClosureRollup) error {
	var errs []string
	requireContract(&errs, "month3_final_closure_rollup", rollup.Schema, AtlasMonth3FinalClosureRollupContract)
	requireField(&errs, "node_id", rollup.NodeID)
	checkPublicPath(&errs, "node_id", rollup.NodeID, true)
	if !oneOf(rollup.Status, "final_closure_rollup_bound", "final_closure_rollup_failed") {
		errs = append(errs, "status must be final_closure_rollup_bound or final_closure_rollup_failed")
	}
	for field, value := range map[string]string{
		"source_readback_path":  rollup.SourceReadbackPath,
		"readiness_matrix_path": rollup.ReadinessMatrixPath,
		"promoter_path":         rollup.PromoterPath,
		"command_path":          rollup.CommandPath,
		"public_safety_path":    rollup.PublicSafetyPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  rollup.SourceReadbackDigest,
		"readiness_matrix_digest": rollup.ReadinessMatrixDigest,
		"promoter_digest":         rollup.PromoterDigest,
		"command_digest":          rollup.CommandDigest,
		"public_safety_digest":    rollup.PublicSafetyDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if rollup.SourceCompletedNodes != 40 || rollup.SourceReadyNodes != 0 || rollup.SourceBlockedNodes != 0 || rollup.SourceFailedNodes != 0 {
		errs = append(errs, "source node counts must be 40 completed and zero ready/blocked/failed")
	}
	if !rollup.SourceFinalResponseAllowed {
		errs = append(errs, "source_final_response_allowed must be true for terminal closure")
	}
	if rollup.MatrixRecommendationCount < 40 {
		errs = append(errs, "matrix_recommendation_count must be at least 40")
	}
	if rollup.ProvenCapabilityCount <= 0 {
		errs = append(errs, "proven_capability_count must be positive")
	}
	if rollup.UnresolvedBlockerCount <= 0 {
		errs = append(errs, "unresolved_blocker_count must be positive")
	}
	if rollup.PromoterStatus != "no_promotion_requested" {
		errs = append(errs, "promoter_status must be no_promotion_requested")
	}
	if rollup.CommandStatus != "readback_agrees_no_promotion" {
		errs = append(errs, "command_status must be readback_agrees_no_promotion")
	}
	if rollup.PublicSafetyStatus != "passed" || !rollup.PublicSafetyScanPassed {
		errs = append(errs, "public safety status must be passed")
	}
	if rollup.AggregatePromotionStatus != "no_promotion_requested" {
		errs = append(errs, "aggregate_promotion_status must be no_promotion_requested")
	}
	if rollup.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if rollup.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if !rollup.TerminalReadbackBound {
		errs = append(errs, "terminal_readback_bound must be true")
	}
	if !rollup.MatrixBound {
		errs = append(errs, "matrix_bound must be true")
	}
	if !rollup.PromoterCommandSafetyAgreement {
		errs = append(errs, "promoter_command_safety_agreement must be true")
	}
	validateNoAuthorityEffects(&errs, rollup.SchedulesWork, rollup.ExecutesWork, rollup.ApprovesWork, rollup.ClaimsAuthorityAdvance, rollup.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3FinalClosureRollup(path string, rollup AtlasMonth3FinalClosureRollup) error {
	return WriteJSON(path, rollup)
}
