package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RenderAtlasMissionOperatorSummary(readback AtlasRecommendationReadback) string {
	nextNode := strings.TrimSpace(readback.FirstExecutableNode)
	if nextNode == "" {
		nextNode = "none"
	}
	var b strings.Builder
	b.WriteString("# AO Atlas Feature Depth Operator Summary\n\n")
	b.WriteString(fmt.Sprintf("- Mission: `%s`\n", readback.MissionID))
	b.WriteString(fmt.Sprintf("- Target instance: `%s`\n", readback.TargetInstance))
	b.WriteString(fmt.Sprintf("- Completed nodes: %d / %d\n", readback.CompletedNodes, readback.TotalNodes))
	b.WriteString(fmt.Sprintf("- Ready nodes: %d\n", readback.ReadyNodes))
	b.WriteString(fmt.Sprintf("- Blocked nodes: %d\n", readback.BlockedNodes))
	b.WriteString(fmt.Sprintf("- Failed nodes: %d\n", readback.FailedNodes))
	b.WriteString(fmt.Sprintf("- Next executable node: `%s`\n", nextNode))
	b.WriteString(fmt.Sprintf("- Return gate: `%s`\n", readback.ReturnGateStatus))
	b.WriteString(fmt.Sprintf("- Continuation contract reason: `%s`\n", readback.ContinuationContract.Reason))
	b.WriteString(fmt.Sprintf("- Early-return risk: `%s`\n", readback.EarlyReturnRiskStatus))
	b.WriteString(fmt.Sprintf("- Final response allowed: `%t`\n\n", readback.FinalResponseAllowed))
	b.WriteString("Exact next action:\n")
	b.WriteString(fmt.Sprintf("- %s\n\n", readback.ExactNextAction))
	if !readback.FinalResponseAllowed {
		b.WriteString("Do not produce a final response while ready nodes or exact next action remain.\n\n")
	}
	b.WriteString("RSI remains denied.\n")
	return b.String()
}

func WriteAtlasMissionOperatorSummary(path string, readback AtlasRecommendationReadback) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("summary path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(RenderAtlasMissionOperatorSummary(readback)), 0o644)
}

func BuildAtlasMissionOperatorSummaryCheck(readbackPath, summaryPath string) (AtlasMissionOperatorSummaryCheck, error) {
	readback, err := LoadJSON[AtlasRecommendationReadback](readbackPath)
	if err != nil {
		return AtlasMissionOperatorSummaryCheck{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMissionOperatorSummaryCheck{}, err
	}
	summaryBytes, err := os.ReadFile(summaryPath)
	if err != nil {
		return AtlasMissionOperatorSummaryCheck{}, err
	}
	summary := string(summaryBytes)
	exactNextActionOccurrences := strings.Count(summary, readback.ExactNextAction)
	fixture := AtlasMissionOperatorSummaryCheck{
		Schema:                            AtlasMissionOperatorSummaryCheckContract,
		Status:                            "passed",
		MissionID:                         readback.MissionID,
		TargetInstance:                    readback.TargetInstance,
		SourceReadbackPath:                publicArtifactRef(readbackPath),
		SummaryMarkdownPath:               publicArtifactRef(summaryPath),
		SourceReadbackDigest:              digestValue(readback),
		CompletedNodes:                    readback.CompletedNodes,
		TotalNodes:                        readback.TotalNodes,
		ReadyNodes:                        readback.ReadyNodes,
		BlockedNodes:                      readback.BlockedNodes,
		FailedNodes:                       readback.FailedNodes,
		FirstExecutableNode:               readback.FirstExecutableNode,
		ExactNextAction:                   readback.ExactNextAction,
		ExactNextActionOccurrences:        exactNextActionOccurrences,
		ExactNextActionWordingPresent:     exactNextActionOccurrences == 1,
		NextExecutableNodeWordingPresent:  strings.Contains(summary, fmt.Sprintf("Next executable node: `%s`", readback.FirstExecutableNode)),
		FinalResponseDeniedWordingPresent: !readback.FinalResponseAllowed && strings.Contains(summary, "Do not produce a final response while ready nodes or exact next action remain."),
		ReturnGateStatus:                  readback.ReturnGateStatus,
		ContinuationContractReason:        readback.ContinuationContract.Reason,
		EarlyReturnRiskStatus:             readback.EarlyReturnRiskStatus,
		FinalResponseAllowed:              readback.FinalResponseAllowed,
		RefusesFinalResponse:              readback.ContinuationContract.RefusesFinalResponse,
		SummaryAssertions: []string{
			"exact_next_action_wording_preserved",
			"next_executable_node_wording_preserved",
			"final_response_denial_wording_preserved",
			"rsi_denial_preserved",
		},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasMissionOperatorSummaryCheck(fixture); err != nil {
		return AtlasMissionOperatorSummaryCheck{}, err
	}
	return fixture, nil
}

func ValidateAtlasMissionOperatorSummaryCheck(fixture AtlasMissionOperatorSummaryCheck) error {
	var errs []string
	requireContract(&errs, "mission_operator_summary_check", fixture.Schema, AtlasMissionOperatorSummaryCheckContract)
	if fixture.Status != "passed" {
		errs = append(errs, "status must be passed")
	}
	requireField(&errs, "mission_id", fixture.MissionID)
	requireField(&errs, "target_instance", fixture.TargetInstance)
	requireField(&errs, "source_readback_path", fixture.SourceReadbackPath)
	requireField(&errs, "summary_markdown_path", fixture.SummaryMarkdownPath)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	checkPublicPath(&errs, "summary_markdown_path", fixture.SummaryMarkdownPath, true)
	if !digestPattern.MatchString(fixture.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if fixture.CompletedNodes < 0 {
		errs = append(errs, "completed_nodes must not be negative")
	}
	if fixture.TotalNodes <= 0 {
		errs = append(errs, "total_nodes must be greater than zero")
	}
	if fixture.ReadyNodes < 0 {
		errs = append(errs, "ready_nodes must not be negative")
	}
	if fixture.BlockedNodes < 0 {
		errs = append(errs, "blocked_nodes must not be negative")
	}
	if fixture.FailedNodes < 0 {
		errs = append(errs, "failed_nodes must not be negative")
	}
	requireField(&errs, "first_executable_node", fixture.FirstExecutableNode)
	requireField(&errs, "exact_next_action", fixture.ExactNextAction)
	requireField(&errs, "return_gate_status", fixture.ReturnGateStatus)
	requireField(&errs, "continuation_contract_reason", fixture.ContinuationContractReason)
	requireField(&errs, "early_return_risk_status", fixture.EarlyReturnRiskStatus)
	for field, value := range map[string]string{
		"first_executable_node":        fixture.FirstExecutableNode,
		"exact_next_action":            fixture.ExactNextAction,
		"return_gate_status":           fixture.ReturnGateStatus,
		"continuation_contract_reason": fixture.ContinuationContractReason,
		"early_return_risk_status":     fixture.EarlyReturnRiskStatus,
	} {
		checkPublicPath(&errs, field, value, true)
	}
	if fixture.ExactNextActionOccurrences != 1 {
		errs = append(errs, "exact_next_action_occurrences must be 1")
	}
	if !fixture.ExactNextActionWordingPresent {
		errs = append(errs, "exact_next_action_wording_present must be true")
	}
	if !fixture.NextExecutableNodeWordingPresent {
		errs = append(errs, "next_executable_node_wording_present must be true")
	}
	if !fixture.FinalResponseDeniedWordingPresent {
		errs = append(errs, "final_response_denied_wording_present must be true")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	if !fixture.RefusesFinalResponse {
		errs = append(errs, "refuses_final_response must be true")
	}
	for _, want := range []string{
		"exact_next_action_wording_preserved",
		"next_executable_node_wording_preserved",
		"final_response_denial_wording_preserved",
		"rsi_denial_preserved",
	} {
		if !containsStringValue(fixture.SummaryAssertions, want) {
			errs = append(errs, "summary_assertions missing "+want)
		}
	}
	if fixture.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if fixture.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if fixture.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if fixture.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !fixture.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}
