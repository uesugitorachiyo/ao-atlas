package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3TerminalDigestBinding(nodeID, sourceReadbackPath, matrixPath string) (AtlasMonth3TerminalDigestBinding, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3TerminalDigestBinding{}, fmt.Errorf("node id is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	matrix, err := LoadJSON[AtlasGoldenPathReadinessMatrix](matrixPath)
	if err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	matrixDigest, err := digestTextFileWithNormalizedLineEndings(matrixPath)
	if err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	countsMatch := readback.CompletedNodes == matrix.CompletedNodes &&
		readback.ReadyNodes == matrix.ReadyNodes &&
		readback.BlockedNodes == matrix.BlockedNodes &&
		readback.FailedNodes == matrix.FailedNodes
	binding := AtlasMonth3TerminalDigestBinding{
		Schema:                    AtlasMonth3TerminalDigestBindingContract,
		NodeID:                    nodeID,
		Status:                    "terminal_digest_binding_ready",
		SourceReadbackPath:        publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:      readbackDigest,
		ReadinessMatrixPath:       publicArtifactRef(matrixPath),
		ReadinessMatrixDigest:     matrixDigest,
		ReadbackCompletedNodes:    readback.CompletedNodes,
		MatrixCompletedNodes:      matrix.CompletedNodes,
		ReadbackReadyNodes:        readback.ReadyNodes,
		MatrixReadyNodes:          matrix.ReadyNodes,
		ReadbackBlockedNodes:      readback.BlockedNodes,
		MatrixBlockedNodes:        matrix.BlockedNodes,
		ReadbackFailedNodes:       readback.FailedNodes,
		MatrixFailedNodes:         matrix.FailedNodes,
		NodeCountsMatch:           countsMatch,
		FinalResponseAllowed:      readback.FinalResponseAllowed,
		MatrixRecommendationCount: matrix.RankedRecommendationCount,
		NoPromotionStatus:         matrix.NoPromotionStatus,
		PromotionRequested:        matrix.PromotionRequested,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    matrix.ClaimsAuthorityAdvance,
		RSIRemainsDenied:          matrix.RSIRemainsDenied && readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !binding.NodeCountsMatch || !binding.FinalResponseAllowed || binding.PromotionRequested || binding.ClaimsAuthorityAdvance || !binding.RSIRemainsDenied {
		binding.Status = "terminal_digest_binding_failed"
	}
	if err := ValidateAtlasMonth3TerminalDigestBinding(binding); err != nil {
		return AtlasMonth3TerminalDigestBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasMonth3TerminalDigestBinding(binding AtlasMonth3TerminalDigestBinding) error {
	var errs []string
	requireContract(&errs, "month3_terminal_digest_binding", binding.Schema, AtlasMonth3TerminalDigestBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if !oneOf(binding.Status, "terminal_digest_binding_ready", "terminal_digest_binding_failed") {
		errs = append(errs, "status must be terminal_digest_binding_ready or terminal_digest_binding_failed")
	}
	for field, value := range map[string]string{
		"source_readback_path":  binding.SourceReadbackPath,
		"readiness_matrix_path": binding.ReadinessMatrixPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  binding.SourceReadbackDigest,
		"readiness_matrix_digest": binding.ReadinessMatrixDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if binding.ReadbackCompletedNodes != 40 || binding.MatrixCompletedNodes != 40 {
		errs = append(errs, "completed node counts must both equal 40")
	}
	if binding.ReadbackReadyNodes != 0 || binding.MatrixReadyNodes != 0 || binding.ReadbackBlockedNodes != 0 || binding.MatrixBlockedNodes != 0 || binding.ReadbackFailedNodes != 0 || binding.MatrixFailedNodes != 0 {
		errs = append(errs, "ready, blocked, and failed node counts must all be zero")
	}
	if !binding.NodeCountsMatch {
		errs = append(errs, "node_counts_match must be true")
	}
	if !binding.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be true for terminal digest binding")
	}
	if binding.MatrixRecommendationCount < 40 {
		errs = append(errs, "matrix_recommendation_count must be at least 40")
	}
	if binding.NoPromotionStatus != "no_promotion_requested" {
		errs = append(errs, "no_promotion_status must be no_promotion_requested")
	}
	if binding.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3TerminalDigestBinding(path string, binding AtlasMonth3TerminalDigestBinding) error {
	return WriteJSON(path, binding)
}
