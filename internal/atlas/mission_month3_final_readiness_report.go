package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3FinalReadinessReport(nodeID, sourceReadbackPath, matrixPath, closureRollupPath, closureReadbackPath string) (AtlasMonth3FinalReadinessReport, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3FinalReadinessReport{}, fmt.Errorf("node id is required")
	}
	sourceReadback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	if err := ValidateAtlasRecommendationReadback(sourceReadback); err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	matrix, err := LoadJSON[AtlasGoldenPathReadinessMatrix](matrixPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	rollup, err := LoadJSON[AtlasMonth3FinalClosureRollup](closureRollupPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	if err := ValidateAtlasMonth3FinalClosureRollup(rollup); err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	closureReadback, err := LoadJSON[AtlasRecommendationReadback](closureReadbackPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	if err := ValidateAtlasRecommendationReadback(closureReadback); err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	matrixDigest, err := digestTextFileWithNormalizedLineEndings(matrixPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	rollupDigest, err := digestTextFileWithNormalizedLineEndings(closureRollupPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	closureReadbackDigest, err := digestTextFileWithNormalizedLineEndings(closureReadbackPath)
	if err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}

	nextActions := make([]string, 0, len(matrix.RankedRecommendations))
	for _, recommendation := range matrix.RankedRecommendations {
		nextActions = append(nextActions, recommendation.Rank+": "+recommendation.Task)
	}
	report := AtlasMonth3FinalReadinessReport{
		Schema:                             AtlasMonth3FinalReadinessReportContract,
		NodeID:                             nodeID,
		Status:                             "ready_for_operator_handoff",
		SourceReadbackPath:                 publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:               sourceDigest,
		ReadinessMatrixPath:                publicArtifactRef(matrixPath),
		ReadinessMatrixDigest:              matrixDigest,
		ClosureRollupPath:                  publicArtifactRef(closureRollupPath),
		ClosureRollupDigest:                rollupDigest,
		ClosureReadbackPath:                publicArtifactRef(closureReadbackPath),
		ClosureReadbackDigest:              closureReadbackDigest,
		SourceCompletedNodes:               sourceReadback.CompletedNodes,
		SourceReadyNodes:                   sourceReadback.ReadyNodes,
		SourceBlockedNodes:                 sourceReadback.BlockedNodes,
		SourceFailedNodes:                  sourceReadback.FailedNodes,
		SourceFinalResponseAllowed:         sourceReadback.FinalResponseAllowed,
		ClosureCompletedNodesBeforeReport:  closureReadback.CompletedNodes,
		ClosureReadyNodesBeforeReport:      closureReadback.ReadyNodes,
		ClosureBlockedNodesBeforeReport:    closureReadback.BlockedNodes,
		ClosureFailedNodesBeforeReport:     closureReadback.FailedNodes,
		ClosureFinalResponseAllowed:        closureReadback.FinalResponseAllowed,
		ClosureNextExecutableNode:          closureReadback.FirstExecutableNode,
		ProvenCapabilities:                 append([]string(nil), matrix.ProvenCapabilities...),
		ProvenCapabilityCount:              len(matrix.ProvenCapabilities),
		UnresolvedBlockers:                 append([]string(nil), matrix.UnresolvedBlockers...),
		UnresolvedBlockerCount:             len(matrix.UnresolvedBlockers),
		RecommendedNextActions:             nextActions,
		RecommendedNextActionCount:         len(nextActions),
		AggregatePromotionStatus:           rollup.AggregatePromotionStatus,
		PromotionRequested:                 matrix.PromotionRequested || rollup.PromotionRequested,
		PromotionGranted:                   rollup.PromotionGranted,
		PromoterCommandSafetyAgreement:     rollup.PromoterCommandSafetyAgreement,
		TerminalReadbackBound:              sourceReadback.FinalResponseAllowed && rollup.TerminalReadbackBound,
		FinalClosureReadyForCompletionNode: closureReadback.CompletedNodes == 29 && closureReadback.ReadyNodes == 1 && closureReadback.FirstExecutableNode == nodeID,
		SchedulesWork:                      false,
		ExecutesWork:                       false,
		ApprovesWork:                       false,
		ClaimsAuthorityAdvance:             matrix.ClaimsAuthorityAdvance || rollup.ClaimsAuthorityAdvance || closureReadback.SchedulesWork || closureReadback.ExecutesWork || closureReadback.ApprovesWork,
		RSIRemainsDenied:                   matrix.RSIRemainsDenied && rollup.RSIRemainsDenied && closureReadback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !report.PromoterCommandSafetyAgreement || !report.TerminalReadbackBound || !report.FinalClosureReadyForCompletionNode || report.PromotionRequested || report.PromotionGranted || report.ClaimsAuthorityAdvance || !report.RSIRemainsDenied {
		report.Status = "readiness_report_failed"
	}
	if err := ValidateAtlasMonth3FinalReadinessReport(report); err != nil {
		return AtlasMonth3FinalReadinessReport{}, err
	}
	return report, nil
}

func ValidateAtlasMonth3FinalReadinessReport(report AtlasMonth3FinalReadinessReport) error {
	var errs []string
	requireContract(&errs, "month3_final_readiness_report", report.Schema, AtlasMonth3FinalReadinessReportContract)
	requireField(&errs, "node_id", report.NodeID)
	checkPublicPath(&errs, "node_id", report.NodeID, true)
	if !oneOf(report.Status, "ready_for_operator_handoff", "readiness_report_failed") {
		errs = append(errs, "status must be ready_for_operator_handoff or readiness_report_failed")
	}
	for field, value := range map[string]string{
		"source_readback_path":  report.SourceReadbackPath,
		"readiness_matrix_path": report.ReadinessMatrixPath,
		"closure_rollup_path":   report.ClosureRollupPath,
		"closure_readback_path": report.ClosureReadbackPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  report.SourceReadbackDigest,
		"readiness_matrix_digest": report.ReadinessMatrixDigest,
		"closure_rollup_digest":   report.ClosureRollupDigest,
		"closure_readback_digest": report.ClosureReadbackDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if report.SourceCompletedNodes != 40 || report.SourceReadyNodes != 0 || report.SourceBlockedNodes != 0 || report.SourceFailedNodes != 0 {
		errs = append(errs, "source node counts must be 40 completed and zero ready/blocked/failed")
	}
	if !report.SourceFinalResponseAllowed {
		errs = append(errs, "source_final_response_allowed must be true")
	}
	if report.ClosureCompletedNodesBeforeReport != 29 || report.ClosureReadyNodesBeforeReport != 1 || report.ClosureBlockedNodesBeforeReport != 0 || report.ClosureFailedNodesBeforeReport != 0 {
		errs = append(errs, "closure readback must be positioned at the final report node")
	}
	if report.ClosureFinalResponseAllowed {
		errs = append(errs, "closure_final_response_allowed must be false before final report completion")
	}
	if report.ClosureNextExecutableNode != report.NodeID {
		errs = append(errs, "closure_next_executable_node must match node_id")
	}
	if report.ProvenCapabilityCount != len(report.ProvenCapabilities) || report.ProvenCapabilityCount <= 0 {
		errs = append(errs, "proven capabilities must be non-empty and count-bound")
	}
	if report.UnresolvedBlockerCount != len(report.UnresolvedBlockers) || report.UnresolvedBlockerCount <= 0 {
		errs = append(errs, "unresolved blockers must be non-empty and count-bound")
	}
	if report.RecommendedNextActionCount != len(report.RecommendedNextActions) || report.RecommendedNextActionCount < 10 {
		errs = append(errs, "recommended next actions must include at least ten count-bound actions")
	}
	if report.AggregatePromotionStatus != "no_promotion_requested" {
		errs = append(errs, "aggregate_promotion_status must be no_promotion_requested")
	}
	if report.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if report.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if !report.PromoterCommandSafetyAgreement {
		errs = append(errs, "promoter_command_safety_agreement must be true")
	}
	if !report.TerminalReadbackBound {
		errs = append(errs, "terminal_readback_bound must be true")
	}
	if !report.FinalClosureReadyForCompletionNode {
		errs = append(errs, "final_closure_ready_for_completion_node must be true")
	}
	validateNoAuthorityEffects(&errs, report.SchedulesWork, report.ExecutesWork, report.ApprovesWork, report.ClaimsAuthorityAdvance, report.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3FinalReadinessReport(path string, report AtlasMonth3FinalReadinessReport) error {
	return WriteJSON(path, report)
}
