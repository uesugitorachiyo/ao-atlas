package atlas

import (
	"fmt"
	"strings"
)

func BuildAOMissionFinalSynthesisReadback(synthesisPath string) (AOMissionFinalSynthesisReadback, error) {
	var synthesis AOMissionFinalSynthesis
	if err := readJSONIfPossible(synthesisPath, &synthesis); err != nil {
		return AOMissionFinalSynthesisReadback{}, err
	}
	if err := ValidateAOMissionFinalSynthesis(synthesis); err != nil {
		return AOMissionFinalSynthesisReadback{}, err
	}
	sourceDigest, err := digestFile(synthesisPath)
	if err != nil {
		return AOMissionFinalSynthesisReadback{}, err
	}
	finalAllowed, finalReason, returnGate := aoMissionFinalSynthesisReturnGate(synthesis)
	readback := AOMissionFinalSynthesisReadback{
		ContractVersion:        AOMissionFinalSynthesisReadbackContract,
		MissionID:              synthesis.Mission,
		Status:                 synthesis.Status,
		SourceDigest:           sourceDigest,
		TotalNodes:             synthesis.CompletedNodes + synthesis.ReadyNodes + synthesis.BlockedNodes,
		CompletedNodes:         synthesis.CompletedNodes,
		ReadyNodes:             synthesis.ReadyNodes,
		BlockedNodes:           synthesis.BlockedNodes,
		MinimumNodes:           synthesis.MinimumNodes,
		TargetMinutes:          synthesis.TargetMinutes,
		MaxMinutes:             synthesis.MaxMinutes,
		ReturnGateStatus:       returnGate,
		FinalResponseAllowed:   finalAllowed,
		FinalResponseReason:    finalReason,
		AtlasWorkgraphStatus:   synthesis.AtlasWorkgraphStatus,
		FoundryRollup:          synthesis.FoundryRollup,
		PromoterStatus:         synthesis.PromoterStatus,
		CommandReadback:        synthesis.CommandReadback,
		EventSearchBound:       synthesis.EventSearchBound,
		BranchCleanupBound:     synthesis.BranchCleanupBoundThroughPreviousNode,
		MergedPRsFinal:         append([]int(nil), synthesis.MergedPRsFinal...),
		ExactNextAction:        synthesis.ExactNextAction,
		FeatureDepthNextTasks:  finalSynthesisNextTasks(synthesis.FeatureDepthRecommendations),
		RSIRemainsDenied:       synthesis.RSIRemainsDenied,
		PromotionClaimed:       synthesis.PromotionClaimed,
		ClaimsAuthorityAdvance: synthesis.ClaimsAuthorityAdvance,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
	}
	if err := ValidateAOMissionFinalSynthesisReadback(readback); err != nil {
		return AOMissionFinalSynthesisReadback{}, err
	}
	return readback, nil
}

func ValidateAOMissionFinalSynthesis(synthesis AOMissionFinalSynthesis) error {
	var errs []string
	if synthesis.Schema != "ao.mission.atlas-wave-final-synthesis.v0.1" {
		errs = append(errs, "schema must be ao.mission.atlas-wave-final-synthesis.v0.1")
	}
	requireField(&errs, "mission", synthesis.Mission)
	if !oneOf(synthesis.Status, "completed", "blocked", "denied") {
		errs = append(errs, "status must be completed, blocked, or denied")
	}
	if synthesis.CompletedNodes < 0 {
		errs = append(errs, "completed_nodes must not be negative")
	}
	if synthesis.ReadyNodes < 0 {
		errs = append(errs, "ready_nodes must not be negative")
	}
	if synthesis.BlockedNodes < 0 {
		errs = append(errs, "blocked_nodes must not be negative")
	}
	if synthesis.MinimumNodes <= 0 {
		errs = append(errs, "minimum_nodes must be positive")
	}
	if synthesis.TargetMinutes <= 0 {
		errs = append(errs, "target_minutes must be positive")
	}
	if synthesis.MaxMinutes < synthesis.TargetMinutes {
		errs = append(errs, "max_minutes must be greater than or equal to target_minutes")
	}
	requireField(&errs, "atlas_workgraph_status", synthesis.AtlasWorkgraphStatus)
	requireField(&errs, "foundry_rollup", synthesis.FoundryRollup)
	requireField(&errs, "promoter_status", synthesis.PromoterStatus)
	requireField(&errs, "command_readback", synthesis.CommandReadback)
	requireField(&errs, "exact_next_action", synthesis.ExactNextAction)
	checkPublicPath(&errs, "current_node_branch", synthesis.CurrentNodeBranch, true)
	checkPublicPath(&errs, "exact_next_action", synthesis.ExactNextAction, true)
	for i, item := range synthesis.FeatureDepthRecommendations {
		checkPublicPath(&errs, fmt.Sprintf("feature_depth_recommendations[%d].id", i), item.ID, true)
		checkPublicPath(&errs, fmt.Sprintf("feature_depth_recommendations[%d].task", i), item.Task, true)
		checkPublicPath(&errs, fmt.Sprintf("feature_depth_recommendations[%d].exact_next_action", i), item.ExactNextAction, true)
	}
	if len(synthesis.FeatureDepthRecommendations) < 10 {
		errs = append(errs, "feature_depth_recommendations must contain at least 10 tasks")
	}
	if synthesis.FinalResponseAllowed {
		if synthesis.Status != "completed" {
			errs = append(errs, "final response requires completed status")
		}
		if synthesis.ReadyNodes > 0 {
			errs = append(errs, "final response cannot be allowed while ready nodes remain")
		}
		if synthesis.BlockedNodes > 0 {
			errs = append(errs, "final response cannot be allowed while blocked nodes remain")
		}
		if synthesis.CompletedNodes < synthesis.MinimumNodes {
			errs = append(errs, "final response requires completed_nodes to meet minimum_nodes")
		}
		if synthesis.AtlasWorkgraphStatus != "completed" {
			errs = append(errs, "final response requires completed Atlas workgraph status")
		}
		if synthesis.CommandReadback != "ready" {
			errs = append(errs, "final response requires ready command_readback")
		}
		if !synthesis.EventSearchBound {
			errs = append(errs, "final response requires event search evidence")
		}
		if !synthesis.BranchCleanupBoundThroughPreviousNode {
			errs = append(errs, "final response requires branch cleanup evidence")
		}
		if strings.TrimSpace(synthesis.CurrentNodeBranch) != "" && synthesis.CurrentNodeBranch != "none" {
			errs = append(errs, "final response requires current_node_branch to be none")
		}
		if synthesis.CurrentNodePRPending {
			errs = append(errs, "final response requires no pending current-node PR")
		}
	}
	if synthesis.PromotionClaimed {
		errs = append(errs, "promotion_claimed must be false")
	}
	if synthesis.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !synthesis.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if synthesis.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if synthesis.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if synthesis.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if synthesis.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func ValidateAOMissionFinalSynthesisReadback(readback AOMissionFinalSynthesisReadback) error {
	var errs []string
	requireContract(&errs, "ao_mission_final_synthesis_readback", readback.ContractVersion, AOMissionFinalSynthesisReadbackContract)
	requireField(&errs, "mission_id", readback.MissionID)
	if !oneOf(readback.Status, "completed", "blocked", "denied") {
		errs = append(errs, "status must be completed, blocked, or denied")
	}
	if !digestPattern.MatchString(readback.SourceDigest) {
		errs = append(errs, "source_digest must be sha256 digest")
	}
	if readback.TotalNodes != readback.CompletedNodes+readback.ReadyNodes+readback.BlockedNodes {
		errs = append(errs, "total_nodes must equal completed_nodes plus ready_nodes plus blocked_nodes")
	}
	if readback.FinalResponseAllowed {
		if readback.ReturnGateStatus != "final_response_allowed" {
			errs = append(errs, "return_gate_status must allow final response")
		}
		if readback.ReadyNodes != 0 || readback.BlockedNodes != 0 {
			errs = append(errs, "final response readback requires zero ready and blocked nodes")
		}
		if readback.FinalResponseReason != "completed Mission final synthesis has zero ready nodes and required closure evidence" {
			errs = append(errs, "final_response_reason does not match completed Mission final synthesis gate")
		}
	} else if readback.ReturnGateStatus == "final_response_allowed" {
		errs = append(errs, "return_gate_status cannot allow final response when final_response_allowed is false")
	}
	requireField(&errs, "final_response_reason", readback.FinalResponseReason)
	requireField(&errs, "atlas_workgraph_status", readback.AtlasWorkgraphStatus)
	requireField(&errs, "foundry_rollup", readback.FoundryRollup)
	requireField(&errs, "promoter_status", readback.PromoterStatus)
	requireField(&errs, "command_readback", readback.CommandReadback)
	requireField(&errs, "exact_next_action", readback.ExactNextAction)
	if len(readback.FeatureDepthNextTasks) < 10 {
		errs = append(errs, "feature_depth_next_tasks must contain at least 10 tasks")
	}
	checkPublicStrings(&errs, "feature_depth_next_tasks", readback.FeatureDepthNextTasks, true)
	checkPublicPath(&errs, "exact_next_action", readback.ExactNextAction, true)
	if !readback.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if readback.PromotionClaimed {
		errs = append(errs, "promotion_claimed must be false")
	}
	if readback.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if readback.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if readback.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if readback.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if readback.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if readback.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func aoMissionFinalSynthesisReturnGate(synthesis AOMissionFinalSynthesis) (bool, string, string) {
	if !synthesis.FinalResponseAllowed {
		return false, "Mission final synthesis denies final response", "blocked_source_final_response_denied"
	}
	if synthesis.ReadyNodes > 0 {
		return false, "ready nodes remain", "blocked_ready_nodes_remain"
	}
	if synthesis.BlockedNodes > 0 {
		return false, "blocked nodes remain", "blocked_nodes_remain"
	}
	if synthesis.CompletedNodes < synthesis.MinimumNodes {
		return false, "minimum nodes unmet", "blocked_minimum_nodes_unmet"
	}
	return true, "completed Mission final synthesis has zero ready nodes and required closure evidence", "final_response_allowed"
}

func finalSynthesisNextTasks(recommendations []AOMissionFinalSynthesisRecommendation) []string {
	tasks := make([]string, 0, len(recommendations))
	for _, item := range recommendations {
		id := strings.TrimSpace(item.ID)
		task := strings.TrimSpace(item.Task)
		if id == "" && task == "" {
			continue
		}
		if id == "" {
			tasks = append(tasks, task)
			continue
		}
		if task == "" {
			tasks = append(tasks, id)
			continue
		}
		tasks = append(tasks, id+": "+task)
	}
	return tasks
}
