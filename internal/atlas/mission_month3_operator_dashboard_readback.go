package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3OperatorDashboardReadback(nodeID, sourceReadbackPath, ciMatrixPath string) (AtlasMonth3OperatorDashboardReadback, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3OperatorDashboardReadback{}, fmt.Errorf("node id is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	matrix, err := LoadJSON[AtlasMonth3CrossRepoCIMatrix](ciMatrixPath)
	if err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	if err := ValidateAtlasMonth3CrossRepoCIMatrix(matrix); err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	matrixDigest, err := digestTextFileWithNormalizedLineEndings(ciMatrixPath)
	if err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	blockers := month3OperatorDashboardBlockers(readback)
	fixture := AtlasMonth3OperatorDashboardReadback{
		Schema:                   AtlasMonth3OperatorDashboardReadbackContract,
		NodeID:                   nodeID,
		Status:                   "operator_dashboard_active",
		SourceReadbackPath:       publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:     readbackDigest,
		CIMatrixPath:             publicArtifactRef(ciMatrixPath),
		CIMatrixDigest:           matrixDigest,
		TerminalGoldenPathStatus: "active_ready_work",
		CompletedNodes:           readback.CompletedNodes,
		ReadyNodes:               readback.ReadyNodes,
		BlockedNodes:             readback.BlockedNodes,
		FailedNodes:              readback.FailedNodes,
		FirstExecutableNode:      readback.FirstExecutableNode,
		ExactNextAction:          readback.ExactNextAction,
		ReturnGateStatus:         readback.ReturnGateStatus,
		Blockers:                 blockers,
		BlockerCount:             len(blockers),
		ReadyWorkVisible:         readback.ReadyNodes > 0 && readback.ExactNextAction != "",
		CIMatrixBound:            matrix.Status == "cross_repo_ci_matrix_ready",
		RequiresPassBeforeMerge:  matrix.RequiresPassBeforeMerge,
		WaitsOnPending:           matrix.WaitsOnPending,
		BlocksOnFailure:          matrix.BlocksOnFailure,
		FinalResponseAllowed:     readback.FinalResponseAllowed,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   matrix.ClaimsAuthorityAdvance,
		RSIRemainsDenied:         matrix.RSIRemainsDenied && readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if fixture.BlockerCount > 0 || fixture.BlockedNodes > 0 || fixture.FailedNodes > 0 {
		fixture.Status = "operator_dashboard_blocked"
		fixture.TerminalGoldenPathStatus = "blocked"
	}
	if err := ValidateAtlasMonth3OperatorDashboardReadback(fixture); err != nil {
		return AtlasMonth3OperatorDashboardReadback{}, err
	}
	return fixture, nil
}

func month3OperatorDashboardBlockers(readback AtlasRecommendationReadback) []string {
	var blockers []string
	if readback.BlockedNodes > 0 {
		blockers = append(blockers, "blocked_nodes_present")
	}
	if readback.FailedNodes > 0 {
		blockers = append(blockers, "failed_nodes_present")
	}
	if strings.TrimSpace(readback.FirstExecutableNode) == "" && readback.ReadyNodes > 0 {
		blockers = append(blockers, "missing_first_executable_node")
	}
	return blockers
}

func ValidateAtlasMonth3OperatorDashboardReadback(fixture AtlasMonth3OperatorDashboardReadback) error {
	var errs []string
	requireContract(&errs, "month3_operator_dashboard_readback", fixture.Schema, AtlasMonth3OperatorDashboardReadbackContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if !oneOf(fixture.Status, "operator_dashboard_active", "operator_dashboard_blocked") {
		errs = append(errs, "status must be operator_dashboard_active or operator_dashboard_blocked")
	}
	for field, value := range map[string]string{
		"source_readback_path":        fixture.SourceReadbackPath,
		"ci_matrix_path":              fixture.CIMatrixPath,
		"terminal_golden_path_status": fixture.TerminalGoldenPathStatus,
		"first_executable_node":       fixture.FirstExecutableNode,
		"exact_next_action":           fixture.ExactNextAction,
		"return_gate_status":          fixture.ReturnGateStatus,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest": fixture.SourceReadbackDigest,
		"ci_matrix_digest":       fixture.CIMatrixDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if fixture.CompletedNodes <= 0 || fixture.ReadyNodes <= 0 || fixture.BlockedNodes != 0 || fixture.FailedNodes != 0 {
		errs = append(errs, "dashboard readback must show active ready work with zero blocked or failed nodes")
	}
	if fixture.BlockerCount != len(fixture.Blockers) {
		errs = append(errs, "blocker_count must match blockers")
	}
	if fixture.BlockerCount != 0 {
		errs = append(errs, "blocker_count must be zero for active dashboard readback")
	}
	if !fixture.ReadyWorkVisible {
		errs = append(errs, "ready_work_visible must be true")
	}
	if !fixture.CIMatrixBound {
		errs = append(errs, "ci_matrix_bound must be true")
	}
	if !fixture.RequiresPassBeforeMerge {
		errs = append(errs, "requires_pass_before_merge must be true")
	}
	if !fixture.WaitsOnPending {
		errs = append(errs, "waits_on_pending must be true")
	}
	if !fixture.BlocksOnFailure {
		errs = append(errs, "blocks_on_failure must be true")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3OperatorDashboardReadback(path string, fixture AtlasMonth3OperatorDashboardReadback) error {
	return WriteJSON(path, fixture)
}
