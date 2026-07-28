package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasRecommendationCheckpointReadback(readback AtlasRecommendationReadback) AtlasRecommendationCheckpointReadback {
	minMinutes := readback.ElapsedMinutes
	maxMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
		maxMinutes = readback.Supervisor.MaxMinutes
	}
	status := "fresh"
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		status = "blocked"
	}
	return AtlasRecommendationCheckpointReadback{
		Schema:                     "ao.atlas.recommendation-checkpoint-readback.v0.1",
		Status:                     status,
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		StartedAt:                  readback.StartedAt,
		CompletedAt:                readback.CompletedAt,
		ElapsedMinutes:             readback.ElapsedMinutes,
		MinMinutes:                 minMinutes,
		MaxMinutes:                 maxMinutes,
		MinMinutesMet:              readback.MinMinutesMet,
		LeaseTimeStatus:            readback.LeaseTimeStatus,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  "elapsed_minutes_recorded_after_node_checkpoint",
		ContinuationContractReason: readback.ContinuationContract.Reason,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		TotalNodes:                 readback.TotalNodes,
		FirstExecutableNode:        readback.FirstExecutableNode,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		FinalResponseReason:        readback.FinalResponseReason,
		ExactNextAction:            readback.ExactNextAction,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
	}
}

func ValidateAtlasRecommendationCheckpointReadback(checkpoint AtlasRecommendationCheckpointReadback) error {
	var errs []string
	if checkpoint.Schema != "ao.atlas.recommendation-checkpoint-readback.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-checkpoint-readback.v0.1")
	}
	if !oneOf(checkpoint.Status, "fresh", "blocked") {
		errs = append(errs, "status must be fresh or blocked")
	}
	requireField(&errs, "mission_id", checkpoint.MissionID)
	if checkpoint.CompletedNodes+checkpoint.ReadyNodes+checkpoint.BlockedNodes+checkpoint.FailedNodes != checkpoint.TotalNodes {
		errs = append(errs, "node counts must sum to total_nodes")
	}
	if checkpoint.ElapsedMinutes < 0 {
		errs = append(errs, "elapsed_minutes must be non-negative")
	}
	requireField(&errs, "lease_time_status", checkpoint.LeaseTimeStatus)
	requireField(&errs, "lease_health_status", checkpoint.LeaseHealthStatus)
	requireField(&errs, "checkpoint_freshness_status", checkpoint.CheckpointFreshnessStatus)
	requireField(&errs, "continuation_contract_reason", checkpoint.ContinuationContractReason)
	if strings.TrimSpace(checkpoint.ContinuationContractReason) != "" &&
		!oneOf(checkpoint.ContinuationContractReason,
			"ready_nodes_or_exact_next_action_remain",
			"ready_nodes_remain",
			"exact_next_action_remains",
			"final response allowed by recommendation readback",
			"blocked_hard_blocker",
			"blocked_lease_timing_missing",
			"blocked_minimum_minutes_unmet",
			"blocked_ready_nodes_remain",
			"blocked_no_executable_ready_node",
		) {
		errs = append(errs, "continuation_contract_reason has unsupported value")
	}
	requireField(&errs, "final_response_reason", checkpoint.FinalResponseReason)
	requireField(&errs, "exact_next_action", checkpoint.ExactNextAction)
	if checkpoint.FinalResponseAllowed && !checkpoint.MinMinutesMet {
		errs = append(errs, "final_response_allowed requires min_minutes_met")
	}
	if checkpoint.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if checkpoint.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if checkpoint.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if checkpoint.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationCommandReadback(readback AtlasRecommendationReadback) AtlasRecommendationCommandReadback {
	minMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
	}
	compactStatus := recommendationCompactReadbackStatus(readback.CompletedNodes, readback.TotalNodes, readback.ReadyNodes, readback.BlockedNodes, readback.FailedNodes, readback.FinalResponseAllowed)
	nodeStatus := compactStatus.NodeCompletionStatus
	compactTimeline := fmt.Sprintf("%d/%d recommendation nodes complete; ready_nodes=%d; blocked_nodes=%d; failed_nodes=%d; elapsed_minutes=%d; min_minutes=%d; min_minutes_met=%t; node_completion_status=%s; lease_time_status=%s; final_response_allowed=%t; continuation_contract_reason=%s; exact_next_action=%s", readback.CompletedNodes, readback.TotalNodes, readback.ReadyNodes, readback.BlockedNodes, readback.FailedNodes, readback.ElapsedMinutes, minMinutes, readback.MinMinutesMet, nodeStatus, readback.LeaseTimeStatus, readback.FinalResponseAllowed, readback.ContinuationContract.Reason, readback.ExactNextAction)
	return AtlasRecommendationCommandReadback{
		Schema:                     "ao.atlas.recommendation-command-readback.v0.1",
		Status:                     readback.Status,
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		TotalNodes:                 readback.TotalNodes,
		StartedAt:                  readback.StartedAt,
		CompletedAt:                readback.CompletedAt,
		ElapsedMinutes:             readback.ElapsedMinutes,
		MinMinutes:                 minMinutes,
		MinMinutesMet:              readback.MinMinutesMet,
		LeaseTimeStatus:            readback.LeaseTimeStatus,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		NodeCompletionStatus:       nodeStatus,
		ReturnGateStatus:           readback.ReturnGateStatus,
		CheckpointCount:            readback.CheckpointCount,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		FinalResponseReason:        readback.FinalResponseReason,
		ExactNextAction:            readback.ExactNextAction,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		CompactTimeline:            compactTimeline,
		CommandTimelineBinding: AtlasRecommendationCommandTimelineBinding{
			Summary:                    compactTimeline,
			FirstExecutableNode:        readback.FirstExecutableNode,
			ExactNextAction:            readback.ExactNextAction,
			ContinuationContractReason: readback.ContinuationContract.Reason,
			ReturnGateStatus:           readback.ReturnGateStatus,
			NodeCompletionStatus:       nodeStatus,
			LeaseTimeStatus:            readback.LeaseTimeStatus,
			LeaseHealthStatus:          readback.LeaseHealthStatus,
			CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
			CheckpointCount:            readback.CheckpointCount,
			CompletedNodes:             readback.CompletedNodes,
			ReadyNodes:                 readback.ReadyNodes,
			TotalNodes:                 readback.TotalNodes,
			ElapsedMinutes:             readback.ElapsedMinutes,
			MinMinutes:                 minMinutes,
			MinMinutesMet:              readback.MinMinutesMet,
			FinalResponseAllowed:       readback.FinalResponseAllowed,
		},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
	}
}

func BuildAtlasRecommendationPromoterReadback(readback AtlasRecommendationReadback) AtlasRecommendationPromoterReadback {
	reason := "Recommendation wave records no mutation authority promotion; RSI remains denied."
	if readback.FinalResponseAllowed {
		reason = "Recommendation wave may close its readback lease, but it does not promote mutation authority; RSI remains denied."
	}
	return AtlasRecommendationPromoterReadback{
		Schema:                     "ao.atlas.recommendation-promoter-readback.v0.1",
		Status:                     "no_promotion",
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		PromotionClaimed:           false,
		RSIRemainsDenied:           true,
		NoPromotionSummary:         "No mutation authority promotion claimed; RSI remains denied.",
		NoPromotionReasonSummary:   fmt.Sprintf("No authority promotion claimed; RSI remains denied; continuation_contract_reason=%s; final_response_allowed=%t.", readback.ContinuationContract.Reason, readback.FinalResponseAllowed),
		NextDeniedClass:            "RSI",
		Reason:                     reason,
		ElapsedMinutes:             readback.ElapsedMinutes,
		MinMinutesMet:              readback.MinMinutesMet,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
	}
}

func BuildAtlasRecommendationFoundryRollup(readback AtlasRecommendationReadback) AtlasRecommendationFoundryRollup {
	compactStatus := recommendationCompactReadbackStatus(readback.CompletedNodes, readback.TotalNodes, readback.ReadyNodes, readback.BlockedNodes, readback.FailedNodes, readback.FinalResponseAllowed)
	nodeStatus := compactStatus.NodeCompletionStatus
	status := "in_progress"
	if nodeStatus == "all_nodes_complete" {
		status = "nodes_complete_lease_pending"
	}
	if readback.FinalResponseAllowed {
		status = "completed"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		status = "blocked"
	}
	return AtlasRecommendationFoundryRollup{
		Schema:                     "ao.atlas.recommendation-foundry-rollup.v0.1",
		Status:                     status,
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		TotalNodes:                 readback.TotalNodes,
		NodeCompletionStatus:       nodeStatus,
		LeaseCompletionStatus:      readback.LeaseTimeStatus,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		ReturnGateStatus:           readback.ReturnGateStatus,
		CheckpointCount:            readback.CheckpointCount,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		ExactNextAction:            readback.ExactNextAction,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
	}
}

func ValidateAtlasRecommendationClosureArtifacts(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup) error {
	var errs []string
	if command.Schema != "ao.atlas.recommendation-command-readback.v0.1" {
		errs = append(errs, "command readback schema must be ao.atlas.recommendation-command-readback.v0.1")
	}
	if promoter.Schema != "ao.atlas.recommendation-promoter-readback.v0.1" {
		errs = append(errs, "promoter readback schema must be ao.atlas.recommendation-promoter-readback.v0.1")
	}
	if foundry.Schema != "ao.atlas.recommendation-foundry-rollup.v0.1" {
		errs = append(errs, "foundry rollup schema must be ao.atlas.recommendation-foundry-rollup.v0.1")
	}
	if command.MissionID != readback.MissionID {
		errs = append(errs, "command readback mission_id disagrees")
	}
	if command.Status != readback.Status {
		errs = append(errs, "command readback status disagrees")
	}
	if promoter.MissionID != readback.MissionID {
		errs = append(errs, "promoter readback mission_id disagrees")
	}
	if foundry.MissionID != readback.MissionID {
		errs = append(errs, "foundry rollup mission_id disagrees")
	}
	if command.CompletedNodes != readback.CompletedNodes || command.ReadyNodes != readback.ReadyNodes || command.TotalNodes != readback.TotalNodes {
		errs = append(errs, "command readback node counts disagree")
	}
	if foundry.CompletedNodes != readback.CompletedNodes || foundry.ReadyNodes != readback.ReadyNodes || foundry.TotalNodes != readback.TotalNodes {
		errs = append(errs, "foundry rollup node counts disagree")
	}
	if foundry.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "foundry rollup continuation_contract_reason disagrees")
	}
	if promoter.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "promoter readback continuation_contract_reason disagrees")
	}
	if command.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "command readback final_response_allowed disagrees")
	}
	if foundry.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "foundry rollup final_response_allowed disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && command.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "command readback return_gate_status disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && foundry.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "foundry rollup return_gate_status disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && command.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "command readback checkpoint_count disagrees")
	}
	if command.CommandTimelineBinding.Summary != command.CompactTimeline {
		errs = append(errs, "command timeline binding summary disagrees")
	}
	if !strings.Contains(command.CompactTimeline, "continuation_contract_reason="+readback.ContinuationContract.Reason) {
		errs = append(errs, "command compact timeline continuation_contract_reason missing")
	}
	if !strings.Contains(command.CompactTimeline, "exact_next_action="+readback.ExactNextAction) {
		errs = append(errs, "command compact timeline exact_next_action missing")
	}
	if command.CommandTimelineBinding.FirstExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "command timeline binding first_executable_node disagrees")
	}
	if command.CommandTimelineBinding.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "command timeline binding exact_next_action disagrees")
	}
	if command.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "command readback continuation_contract_reason disagrees")
	}
	if command.CommandTimelineBinding.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "command timeline binding continuation_contract_reason disagrees")
	}
	if command.CommandTimelineBinding.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "command timeline binding return_gate_status disagrees")
	}
	if command.CommandTimelineBinding.NodeCompletionStatus != command.NodeCompletionStatus {
		errs = append(errs, "command timeline binding node_completion_status disagrees")
	}
	if command.CommandTimelineBinding.LeaseTimeStatus != readback.LeaseTimeStatus {
		errs = append(errs, "command timeline binding lease_time_status disagrees")
	}
	if command.CommandTimelineBinding.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "command timeline binding checkpoint_count disagrees")
	}
	if command.CommandTimelineBinding.CompletedNodes != readback.CompletedNodes ||
		command.CommandTimelineBinding.ReadyNodes != readback.ReadyNodes ||
		command.CommandTimelineBinding.TotalNodes != readback.TotalNodes {
		errs = append(errs, "command timeline binding node counts disagree")
	}
	if command.CommandTimelineBinding.ElapsedMinutes != readback.ElapsedMinutes ||
		command.CommandTimelineBinding.MinMinutes != command.MinMinutes ||
		command.CommandTimelineBinding.MinMinutesMet != readback.MinMinutesMet {
		errs = append(errs, "command timeline binding lease timing disagrees")
	}
	if command.CommandTimelineBinding.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "command timeline binding final_response_allowed disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && foundry.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "foundry rollup checkpoint_count disagrees")
	}
	if command.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "command readback lease_health_status disagrees")
	}
	if command.CommandTimelineBinding.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "command timeline binding lease_health_status disagrees")
	}
	if promoter.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "promoter readback lease_health_status disagrees")
	}
	if foundry.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "foundry rollup lease_health_status disagrees")
	}
	if command.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "command readback checkpoint_freshness_status disagrees")
	}
	if command.CommandTimelineBinding.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "command timeline binding checkpoint_freshness_status disagrees")
	}
	if promoter.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "promoter readback checkpoint_freshness_status disagrees")
	}
	if foundry.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "foundry rollup checkpoint_freshness_status disagrees")
	}
	if foundry.Status == "completed" && !readback.FinalResponseAllowed {
		errs = append(errs, "foundry rollup completed while recommendation final response is denied")
	}
	if promoter.PromotionClaimed {
		errs = append(errs, "promoter readback must not claim promotion for recommendation wave")
	}
	if !promoter.RSIRemainsDenied {
		errs = append(errs, "promoter readback must keep RSI denied")
	}
	if promoter.NoPromotionSummary != "No mutation authority promotion claimed; RSI remains denied." {
		errs = append(errs, "promoter readback must include no-promotion summary")
	}
	if !strings.Contains(promoter.NoPromotionReasonSummary, "continuation_contract_reason="+readback.ContinuationContract.Reason) {
		errs = append(errs, "promoter readback no_promotion_reason_summary must include continuation_contract_reason")
	}
	if !strings.Contains(promoter.NoPromotionReasonSummary, fmt.Sprintf("final_response_allowed=%t", readback.FinalResponseAllowed)) {
		errs = append(errs, "promoter readback no_promotion_reason_summary must include final_response_allowed")
	}
	if promoter.NextDeniedClass != "RSI" {
		errs = append(errs, "promoter readback next_denied_class must be RSI")
	}
	if command.SchedulesWork || command.ExecutesWork || command.ApprovesWork || command.ClaimsAuthorityAdvance {
		errs = append(errs, "command readback must not schedule, execute, approve, or claim authority advance")
	}
	if promoter.SchedulesWork || promoter.ExecutesWork || promoter.ApprovesWork || promoter.ClaimsAuthorityAdvance {
		errs = append(errs, "promoter readback must not schedule, execute, approve, or claim authority advance")
	}
	if foundry.SchedulesWork || foundry.ExecutesWork || foundry.ApprovesWork || foundry.ClaimsAuthorityAdvance {
		errs = append(errs, "foundry rollup must not schedule, execute, approve, or claim authority advance")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationReconciliationPacket(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup) AtlasRecommendationReconciliationPacket {
	artifactsAgree := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry) == nil
	continuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	status := "continuation_required"
	if !artifactsAgree {
		status = "blocked_stale_artifact"
	} else if readback.FinalResponseAllowed {
		status = "ready"
	}
	return AtlasRecommendationReconciliationPacket{
		Schema:                       "ao.atlas.recommendation-reconciliation-packet.v0.1",
		Status:                       status,
		MissionID:                    readback.MissionID,
		EvidenceRoot:                 readback.EvidenceRoot,
		FinalStateReconciliation:     buildRecommendationFinalStateReconciliation(readback, command, promoter, foundry, status),
		CompletedNodes:               readback.CompletedNodes,
		ReadyNodes:                   readback.ReadyNodes,
		BlockedNodes:                 readback.BlockedNodes,
		FailedNodes:                  readback.FailedNodes,
		TotalNodes:                   readback.TotalNodes,
		CheckpointCount:              readback.CheckpointCount,
		ReturnGateStatus:             readback.ReturnGateStatus,
		LeaseTimeStatus:              readback.LeaseTimeStatus,
		LeaseHealthStatus:            readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:    readback.CheckpointFreshnessStatus,
		StaleRouteDecisionStatus:     readback.StaleRouteDecisionStatus,
		FinalResponseAllowed:         readback.FinalResponseAllowed,
		FinalResponseReason:          readback.FinalResponseReason,
		ExactNextAction:              readback.ExactNextAction,
		ContinuationContractReason:   readback.ContinuationContract.Reason,
		CommandReturnGateStatus:      command.ReturnGateStatus,
		CommandContinuationReason:    command.ContinuationContractReason,
		CommandFinalResponseAllowed:  command.FinalResponseAllowed,
		PromoterStatus:               promoter.Status,
		PromoterContinuationReason:   promoter.ContinuationContractReason,
		PromotionClaimed:             promoter.PromotionClaimed,
		RSIRemainsDenied:             promoter.RSIRemainsDenied,
		FoundryStatus:                foundry.Status,
		FoundryReturnGateStatus:      foundry.ReturnGateStatus,
		FoundryContinuationReason:    foundry.ContinuationContractReason,
		FoundryNodeCompletionStatus:  foundry.NodeCompletionStatus,
		FoundryLeaseCompletionStatus: foundry.LeaseCompletionStatus,
		FoundryFinalResponseAllowed:  foundry.FinalResponseAllowed,
		ContinuationReasonAgreement:  continuationReasonAgreement,
		ArtifactsAgree:               artifactsAgree,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
	}
}

func buildRecommendationFinalStateReconciliation(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, status string) AtlasFinalStateReconciliation {
	continuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	return AtlasFinalStateReconciliation{
		ContractVersion:       "ao.atlas.final-state-reconciliation.v0.1",
		Status:                status,
		WorkgraphStatus:       readback.Status,
		FoundryRollupStatus:   foundry.Status,
		PromoterVerdictStatus: promoter.Status,
		CommandReadbackStatus: command.Status,
		ExactNextAction:       readback.ExactNextAction,
		ContinuationReason:    readback.ContinuationContract.Reason,
		ContinuationAgreement: continuationReasonAgreement,
		SchedulesWork:         false,
		ExecutesWork:          false,
		ApprovesWork:          false,
	}
}

func ValidateAtlasRecommendationReconciliationPacket(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, packet AtlasRecommendationReconciliationPacket) error {
	var errs []string
	if packet.Schema != "ao.atlas.recommendation-reconciliation-packet.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-reconciliation-packet.v0.1")
	}
	if packet.MissionID != readback.MissionID {
		errs = append(errs, "reconciliation mission_id disagrees")
	}
	validateRecommendationFinalStateReconciliation(&errs, readback, command, promoter, foundry, packet)
	if packet.CompletedNodes != readback.CompletedNodes || packet.ReadyNodes != readback.ReadyNodes || packet.TotalNodes != readback.TotalNodes {
		errs = append(errs, "reconciliation node counts disagree")
	}
	if packet.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "reconciliation checkpoint_count disagrees")
	}
	if packet.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "reconciliation return_gate_status disagrees")
	}
	if packet.LeaseTimeStatus != readback.LeaseTimeStatus {
		errs = append(errs, "reconciliation lease_time_status disagrees")
	}
	if packet.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "reconciliation lease_health_status disagrees")
	}
	if packet.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "reconciliation checkpoint_freshness_status disagrees")
	}
	if packet.StaleRouteDecisionStatus != readback.StaleRouteDecisionStatus {
		errs = append(errs, "reconciliation stale_route_decision_status disagrees")
	}
	if packet.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "reconciliation final_response_allowed disagrees")
	}
	if packet.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "reconciliation exact_next_action disagrees")
	}
	if packet.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "reconciliation continuation_contract_reason disagrees")
	}
	if packet.CommandReturnGateStatus != command.ReturnGateStatus || packet.CommandFinalResponseAllowed != command.FinalResponseAllowed {
		errs = append(errs, "reconciliation command fields disagree")
	}
	if packet.CommandContinuationReason != command.ContinuationContractReason {
		errs = append(errs, "reconciliation command_continuation_contract_reason disagrees")
	}
	if packet.PromoterStatus != promoter.Status || packet.PromotionClaimed != promoter.PromotionClaimed || packet.RSIRemainsDenied != promoter.RSIRemainsDenied {
		errs = append(errs, "reconciliation promoter fields disagree")
	}
	if packet.PromoterContinuationReason != promoter.ContinuationContractReason {
		errs = append(errs, "reconciliation promoter_continuation_contract_reason disagrees")
	}
	if packet.FoundryStatus != foundry.Status ||
		packet.FoundryReturnGateStatus != foundry.ReturnGateStatus ||
		packet.FoundryNodeCompletionStatus != foundry.NodeCompletionStatus ||
		packet.FoundryLeaseCompletionStatus != foundry.LeaseCompletionStatus ||
		packet.FoundryFinalResponseAllowed != foundry.FinalResponseAllowed {
		errs = append(errs, "reconciliation foundry fields disagree")
	}
	if packet.FoundryContinuationReason != foundry.ContinuationContractReason {
		errs = append(errs, "reconciliation foundry_continuation_contract_reason disagrees")
	}
	expectedContinuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	if packet.ContinuationReasonAgreement != expectedContinuationReasonAgreement {
		errs = append(errs, "reconciliation continuation_reason_agreement disagrees")
	}
	closureErr := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry)
	if closureErr == nil && !packet.ArtifactsAgree {
		errs = append(errs, "reconciliation artifacts_agree must be true when closure artifacts agree")
	}
	if closureErr == nil && !packet.ContinuationReasonAgreement {
		errs = append(errs, "reconciliation continuation_reason_agreement must be true when closure artifacts agree")
	}
	if closureErr == nil && packet.Status == "blocked_stale_artifact" {
		errs = append(errs, "reconciliation status blocked_stale_artifact requires stale closure artifacts")
	}
	if closureErr != nil && packet.ArtifactsAgree {
		errs = append(errs, "reconciliation artifacts_agree must be false when closure artifacts disagree")
	}
	if closureErr != nil && packet.Status != "blocked_stale_artifact" {
		errs = append(errs, "reconciliation status must be blocked_stale_artifact when closure artifacts disagree")
	}
	if packet.Status == "ready" && !packet.FinalResponseAllowed {
		errs = append(errs, "reconciliation ready status requires final_response_allowed")
	}
	if packet.SchedulesWork || packet.ExecutesWork || packet.ApprovesWork || packet.ClaimsAuthorityAdvance {
		errs = append(errs, "reconciliation packet must not schedule, execute, approve, or claim authority advance")
	}
	return joinErrors(errs)
}

func validateRecommendationFinalStateReconciliation(errs *[]string, readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, packet AtlasRecommendationReconciliationPacket) {
	finalState := packet.FinalStateReconciliation
	if finalState.ContractVersion != "ao.atlas.final-state-reconciliation.v0.1" {
		*errs = append(*errs, "final_state_reconciliation.contract_version must be ao.atlas.final-state-reconciliation.v0.1")
	}
	if finalState.Status != packet.Status {
		*errs = append(*errs, "final_state_reconciliation.status must match reconciliation status")
	}
	if finalState.WorkgraphStatus != readback.Status {
		*errs = append(*errs, "final_state_reconciliation.workgraph_status disagrees")
	}
	if finalState.FoundryRollupStatus != foundry.Status {
		*errs = append(*errs, "final_state_reconciliation.foundry_rollup_status disagrees")
	}
	if finalState.PromoterVerdictStatus != promoter.Status {
		*errs = append(*errs, "final_state_reconciliation.promoter_verdict_status disagrees")
	}
	if finalState.CommandReadbackStatus != command.Status {
		*errs = append(*errs, "final_state_reconciliation.command_readback_status disagrees")
	}
	if finalState.ExactNextAction != readback.ExactNextAction {
		*errs = append(*errs, "final_state_reconciliation.exact_next_action disagrees")
	}
	if finalState.ContinuationReason != readback.ContinuationContract.Reason {
		*errs = append(*errs, "final_state_reconciliation.continuation_contract_reason disagrees")
	}
	expectedContinuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	if finalState.ContinuationAgreement != expectedContinuationReasonAgreement {
		*errs = append(*errs, "final_state_reconciliation.continuation_reason_agreement disagrees")
	}
	if finalState.SchedulesWork || finalState.ExecutesWork || finalState.ApprovesWork {
		*errs = append(*errs, "final_state_reconciliation must not schedule, execute, or approve work")
	}
	if (readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "") &&
		!readback.FinalResponseAllowed &&
		!oneOf(finalState.Status, "continuation_required", "blocked_stale_artifact") {
		*errs = append(*errs, "final_state_reconciliation must require continuation while ready nodes or exact next action remain")
	}
}

func ValidateAtlasRecommendationExecutionReadback(execution AtlasRecommendationExecutionReadback, readback AtlasRecommendationReadback) error {
	var errs []string
	if execution.Schema != "ao.atlas.long-recommendation-wave-execution.v0.3" {
		errs = append(errs, "schema must be ao.atlas.long-recommendation-wave-execution.v0.3")
	}
	requireField(&errs, "status", execution.Status)
	if execution.MissionID != readback.MissionID {
		errs = append(errs, "mission_id must match recommendation readback")
	}
	if execution.TotalRecommendationNodes != readback.TotalNodes {
		errs = append(errs, "total_recommendation_nodes must match recommendation readback total_nodes")
	}
	if execution.CompletedRecommendationNodes != readback.CompletedNodes {
		errs = append(errs, "completed_recommendation_nodes must match recommendation readback completed_nodes")
	}
	if execution.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "lease_health_status must match recommendation readback")
	}
	if execution.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "checkpoint_freshness_status must match recommendation readback")
	}
	if execution.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "return_gate_status must match recommendation readback")
	}
	if execution.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "continuation_contract_reason must match recommendation readback")
	}
	if execution.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "exact_next_action must match recommendation readback")
	}
	if execution.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must match recommendation readback")
	}
	if execution.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse {
		errs = append(errs, "refuses_final_response must match recommendation readback")
	}
	if execution.GeneratedWorkgraph.TotalNodes != readback.TotalNodes {
		errs = append(errs, "generated_workgraph.total_nodes must match recommendation readback total_nodes")
	}
	if execution.GeneratedWorkgraph.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "generated_workgraph.ready_nodes must match recommendation readback ready_nodes")
	}
	if execution.GeneratedWorkgraph.ExecutableReadyNodes != readback.ExecutableReadyNodes {
		errs = append(errs, "generated_workgraph.executable_ready_nodes must match recommendation readback executable_ready_nodes")
	}
	if execution.GeneratedWorkgraph.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "generated_workgraph.final_response_allowed must match recommendation readback final_response_allowed")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && execution.GeneratedWorkgraph.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "generated_workgraph.return_gate_status must match recommendation readback return_gate_status")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && execution.GeneratedWorkgraph.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "generated_workgraph.checkpoint_count must match recommendation readback checkpoint_count")
	}
	if execution.GeneratedWorkgraph.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "generated_workgraph.lease_health_status must match recommendation readback")
	}
	if execution.GeneratedWorkgraph.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "generated_workgraph.checkpoint_freshness_status must match recommendation readback")
	}
	requireField(&errs, "foundry_run_link_readiness_summary.status", execution.FoundryRunLinkReadinessSummary.Status)
	requireField(&errs, "foundry_run_link_readiness_summary.summary", execution.FoundryRunLinkReadinessSummary.Summary)
	if execution.FoundryRunLinkReadinessSummary.CompletedRunLinks != readback.CompletedNodes {
		errs = append(errs, "foundry run-link readiness completed_run_links must match recommendation readback completed_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.RequiredRunLinks != readback.TotalNodes {
		errs = append(errs, "foundry run-link readiness required_run_links must match recommendation readback total_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.MissingRunLinks != readback.TotalNodes-readback.CompletedNodes {
		errs = append(errs, "foundry run-link readiness missing_run_links must match remaining nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "foundry run-link readiness ready_nodes must match recommendation readback ready_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.NextExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "foundry run-link readiness next_executable_node must match recommendation readback first_executable_node")
	}
	if execution.FoundryRunLinkReadinessSummary.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "foundry run-link readiness checkpoint_count must match recommendation readback checkpoint_count")
	}
	if execution.FoundryRunLinkReadinessSummary.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "foundry run-link readiness final_response_allowed must match recommendation readback final_response_allowed")
	}
	if execution.FoundryRunLinkReadinessSummary.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "foundry run-link readiness lease_health_status must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "foundry run-link readiness checkpoint_freshness_status must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "foundry run-link readiness return_gate_status must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "foundry run-link readiness continuation_contract_reason must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "foundry run-link readiness exact_next_action must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse {
		errs = append(errs, "foundry run-link readiness refuses_final_response must match recommendation readback")
	}
	if sourceDigest, ok := sourceArtifactDigest(execution.SourceArtifacts, "foundry_run_link_readiness_summary"); !ok {
		errs = append(errs, "source_artifacts must include foundry_run_link_readiness_summary")
	} else if sourceDigest != digestValue(execution.FoundryRunLinkReadinessSummary) {
		errs = append(errs, "foundry_run_link_readiness_summary source artifact digest disagrees")
	}
	validateContinuationReasonCoverageSummary(&errs, execution.ContinuationReasonCoverage, readback)
	if sourceDigest, ok := sourceArtifactDigest(execution.SourceArtifacts, "continuation_reason_coverage"); !ok {
		errs = append(errs, "source_artifacts must include continuation_reason_coverage")
	} else if sourceDigest != digestValue(execution.ContinuationReasonCoverage) {
		errs = append(errs, "continuation_reason_coverage source artifact digest disagrees")
	}
	validateReasonArtifactAgreementSummary(&errs, execution.ReasonArtifactAgreementSummary, execution, readback)
	if execution.Status == "completed" && !readback.FinalResponseAllowed {
		errs = append(errs, "status completed requires recommendation readback final_response_allowed")
	}
	return joinErrors(errs)
}

func validateContinuationReasonCoverageSummary(errs *[]string, coverage AtlasRecommendationContinuationReasonCoverage, readback AtlasRecommendationReadback) {
	requireField(errs, "continuation_reason_coverage.status", coverage.Status)
	if coverage.Status != "coverage_sources_indexed" {
		*errs = append(*errs, "continuation_reason_coverage.status must be coverage_sources_indexed")
	}
	if coverage.ExpectedReason != readback.ContinuationContract.Reason {
		*errs = append(*errs, "continuation_reason_coverage.expected_reason must match recommendation readback")
	}
	requiredSources := continuationReasonCoverageRequiredSources()
	if coverage.SourceCount != len(coverage.IndexedSources) {
		*errs = append(*errs, "continuation_reason_coverage.source_count must match indexed_sources length")
	}
	if coverage.SourceCount != len(requiredSources) {
		*errs = append(*errs, "continuation_reason_coverage.source_count must cover required sources")
	}
	for _, source := range requiredSources {
		if !containsStringValue(coverage.IndexedSources, source) {
			*errs = append(*errs, "continuation_reason_coverage.indexed_sources missing "+source)
		}
	}
	if coverage.FinalResponseAllowed != readback.FinalResponseAllowed {
		*errs = append(*errs, "continuation_reason_coverage.final_response_allowed must match recommendation readback")
	}
	if coverage.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse {
		*errs = append(*errs, "continuation_reason_coverage.refuses_final_response must match recommendation readback")
	}
	if coverage.ExactNextAction != readback.ExactNextAction {
		*errs = append(*errs, "continuation_reason_coverage.exact_next_action must match recommendation readback")
	}
	if coverage.ReturnGateStatus != readback.ReturnGateStatus {
		*errs = append(*errs, "continuation_reason_coverage.return_gate_status must match recommendation readback")
	}
	if coverage.LeaseHealthStatus != readback.LeaseHealthStatus {
		*errs = append(*errs, "continuation_reason_coverage.lease_health_status must match recommendation readback")
	}
	if coverage.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		*errs = append(*errs, "continuation_reason_coverage.checkpoint_freshness_status must match recommendation readback")
	}
	if coverage.ClaimsAuthorityAdvance {
		*errs = append(*errs, "continuation_reason_coverage.claims_authority_advance must be false")
	}
	if !coverage.RSIRemainsDenied {
		*errs = append(*errs, "continuation_reason_coverage.rsi_remains_denied must be true")
	}
}

func validateReasonArtifactAgreementSummary(errs *[]string, summary AtlasRecommendationReasonArtifactAgreementSummary, execution AtlasRecommendationExecutionReadback, readback AtlasRecommendationReadback) {
	requireField(errs, "reason_artifact_agreement_summary.status", summary.Status)
	if summary.Status != "agreement" {
		*errs = append(*errs, "reason_artifact_agreement_summary.status must be agreement")
	}
	if summary.ExpectedReason != readback.ContinuationContract.Reason {
		*errs = append(*errs, "reason_artifact_agreement_summary.expected_reason must match recommendation readback")
	}
	if summary.SourceCount != len(summary.IndexedSources) {
		*errs = append(*errs, "reason_artifact_agreement_summary.source_count must match indexed_sources length")
	}
	if summary.SourceCount != execution.ContinuationReasonCoverage.SourceCount {
		*errs = append(*errs, "reason_artifact_agreement_summary.source_count must match continuation_reason_coverage source_count")
	}
	requiredSources := continuationReasonCoverageRequiredSources()
	for _, source := range requiredSources {
		if !containsStringValue(summary.IndexedSources, source) {
			*errs = append(*errs, "reason_artifact_agreement_summary.indexed_sources missing "+source)
		}
	}
	if !summary.AllRequiredSourcesIndexed {
		*errs = append(*errs, "reason_artifact_agreement_summary.all_required_sources_indexed must be true")
	}
	if summary.SourceArtifactCount != len(execution.SourceArtifacts) {
		*errs = append(*errs, "reason_artifact_agreement_summary.source_artifact_count must match source_artifacts length")
	}
	for _, ref := range []string{"foundry_run_link_readiness_summary", "continuation_reason_coverage"} {
		if !containsStringValue(summary.SourceArtifactRefs, ref) {
			*errs = append(*errs, "reason_artifact_agreement_summary.source_artifact_refs missing "+ref)
		}
	}
	if !summary.SourceArtifactsAgree {
		*errs = append(*errs, "reason_artifact_agreement_summary.source_artifacts_agree must be true")
	}
	if digest, ok := sourceArtifactDigest(execution.SourceArtifacts, "foundry_run_link_readiness_summary"); !ok {
		*errs = append(*errs, "reason_artifact_agreement_summary requires foundry_run_link_readiness_summary digest")
	} else if summary.FoundryRunLinkReadinessDigest != digest {
		*errs = append(*errs, "reason_artifact_agreement_summary.foundry_run_link_readiness_digest disagrees")
	}
	if digest, ok := sourceArtifactDigest(execution.SourceArtifacts, "continuation_reason_coverage"); !ok {
		*errs = append(*errs, "reason_artifact_agreement_summary requires continuation_reason_coverage digest")
	} else if summary.ContinuationReasonCoverageDigest != digest {
		*errs = append(*errs, "reason_artifact_agreement_summary.continuation_reason_coverage_digest disagrees")
	}
	if summary.FinalResponseAllowed != readback.FinalResponseAllowed {
		*errs = append(*errs, "reason_artifact_agreement_summary.final_response_allowed must match recommendation readback")
	}
	if summary.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse {
		*errs = append(*errs, "reason_artifact_agreement_summary.refuses_final_response must match recommendation readback")
	}
	if summary.ExactNextAction != readback.ExactNextAction {
		*errs = append(*errs, "reason_artifact_agreement_summary.exact_next_action must match recommendation readback")
	}
	if summary.ReturnGateStatus != readback.ReturnGateStatus {
		*errs = append(*errs, "reason_artifact_agreement_summary.return_gate_status must match recommendation readback")
	}
	if summary.ClaimsAuthorityAdvance {
		*errs = append(*errs, "reason_artifact_agreement_summary.claims_authority_advance must be false")
	}
	if !summary.RSIRemainsDenied {
		*errs = append(*errs, "reason_artifact_agreement_summary.rsi_remains_denied must be true")
	}
}

func continuationReasonCoverageRequiredSources() []string {
	return []string{
		"recommendation_readback",
		"checkpoint_readback",
		"workgraph_readiness_packet",
		"command_readback",
		"command_timeline_binding",
		"promoter_readback",
		"foundry_rollup",
		"reconciliation_packet",
		"reconciliation_command",
		"reconciliation_promoter",
		"reconciliation_foundry",
		"final_state_reconciliation",
		"resume_prompt",
	}
}

func containsStringValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func BuildAtlasRecommendationWorkgraphReadinessPacket(readback AtlasRecommendationReadback, options AtlasRecommendationWorkgraphReadinessPacketOptions) (AtlasRecommendationWorkgraphReadinessPacket, error) {
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasRecommendationWorkgraphReadinessPacket{}, err
	}
	waveDigest := readback.WaveDigest
	if strings.TrimSpace(options.WavePath) != "" {
		digest, err := digestFile(options.WavePath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		waveDigest = digest
	}
	if strings.TrimSpace(waveDigest) == "" {
		waveDigest = digestValue(readback.MissionID + readback.SourceDigest)
	}
	workgraphDigest := readback.WorkgraphDigest
	if strings.TrimSpace(options.WorkgraphPath) != "" {
		digest, err := digestFile(options.WorkgraphPath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		workgraphDigest = digest
	}
	if strings.TrimSpace(workgraphDigest) == "" {
		workgraphDigest = digestValue(readback.TotalNodes)
	}
	readbackDigest := digestValue(readback)
	if strings.TrimSpace(options.ReadbackPath) != "" {
		digest, err := digestFile(options.ReadbackPath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		readbackDigest = digest
	}
	continueIfFastTarget := readback.TotalNodes
	if readback.Supervisor != nil && readback.Supervisor.ContinueIfFastTarget > 0 {
		continueIfFastTarget = readback.Supervisor.ContinueIfFastTarget
	}
	status := "continuation_required"
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		status = "blocked"
	}
	if readback.FinalResponseAllowed {
		status = "ready_for_final_response"
	}
	packet := AtlasRecommendationWorkgraphReadinessPacket{
		Schema:                          "ao.atlas.recommendation-workgraph-readiness-packet.v0.1",
		Status:                          status,
		MissionID:                       readback.MissionID,
		TargetInstance:                  readback.TargetInstance,
		EvidenceRoot:                    readback.EvidenceRoot,
		WaveDigest:                      waveDigest,
		WorkgraphDigest:                 workgraphDigest,
		ReadbackDigest:                  readbackDigest,
		TotalNodes:                      readback.TotalNodes,
		MinimumNodes:                    readback.MinimumNodes,
		NodeBudget:                      readback.TotalNodes,
		ContinueIfFastTarget:            continueIfFastTarget,
		CompletedNodes:                  readback.CompletedNodes,
		ReadyNodes:                      readback.ReadyNodes,
		BlockedNodes:                    readback.BlockedNodes,
		FailedNodes:                     readback.FailedNodes,
		ExecutableReadyNodes:            readback.ExecutableReadyNodes,
		FirstExecutableNode:             readback.FirstExecutableNode,
		LeaseHealthStatus:               readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:       readback.CheckpointFreshnessStatus,
		ReturnGateStatus:                readback.ReturnGateStatus,
		CheckpointCount:                 readback.CheckpointCount,
		EarlyReturnRiskStatus:           readback.EarlyReturnRiskStatus,
		ContinuationBudgetStatus:        recommendationContinuationBudgetStatus(readback, continueIfFastTarget),
		FinalResponseAllowed:            readback.FinalResponseAllowed,
		FinalResponseReason:             readback.FinalResponseReason,
		ExactNextAction:                 readback.ExactNextAction,
		ContinuationContractReason:      readback.ContinuationContract.Reason,
		OneExecutableMutationNodeActive: readback.ExecutableReadyNodes == 1,
		RefusesFinalResponse:            !readback.FinalResponseAllowed,
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(packet, readback); err != nil {
		return AtlasRecommendationWorkgraphReadinessPacket{}, err
	}
	return packet, nil
}

func recommendationContinuationBudgetStatus(readback AtlasRecommendationReadback, continueIfFastTarget int) string {
	if readback.FinalResponseAllowed {
		return "all_generated_nodes_complete"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		return "hard_blocker_requires_repair"
	}
	if readback.ReadyNodes > 0 && readback.CompletedNodes < readback.MinimumNodes {
		return "minimum_nodes_unmet_continue_to_40_node_budget"
	}
	if readback.ReadyNodes > 0 && readback.CompletedNodes >= readback.MinimumNodes && readback.CompletedNodes < continueIfFastTarget {
		return "minimum_met_continue_if_fast_budget_open"
	}
	if readback.ReadyNodes == 0 && !readback.MinMinutesMet {
		return "node_budget_complete_waiting_for_lease_evidence"
	}
	return "continuation_required"
}

func ValidateAtlasRecommendationWorkgraphReadinessPacket(packet AtlasRecommendationWorkgraphReadinessPacket, readback AtlasRecommendationReadback) error {
	var errs []string
	if packet.Schema != "ao.atlas.recommendation-workgraph-readiness-packet.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-workgraph-readiness-packet.v0.1")
	}
	if !oneOf(packet.Status, "continuation_required", "ready_for_final_response", "blocked") {
		errs = append(errs, "status must be continuation_required, ready_for_final_response, or blocked")
	}
	if packet.MissionID != readback.MissionID {
		errs = append(errs, "mission_id must match recommendation readback")
	}
	if packet.TargetInstance != readback.TargetInstance {
		errs = append(errs, "target_instance must match recommendation readback")
	}
	for field, digest := range map[string]string{
		"wave_digest":      packet.WaveDigest,
		"workgraph_digest": packet.WorkgraphDigest,
		"readback_digest":  packet.ReadbackDigest,
	} {
		if !digestPattern.MatchString(digest) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if packet.TotalNodes != readback.TotalNodes {
		errs = append(errs, "total_nodes must match recommendation readback")
	}
	if packet.MinimumNodes != readback.MinimumNodes {
		errs = append(errs, "minimum_nodes must match recommendation readback")
	}
	if packet.NodeBudget != readback.TotalNodes {
		errs = append(errs, "node_budget must match recommendation readback total_nodes")
	}
	expectedContinueTarget := readback.TotalNodes
	if readback.Supervisor != nil && readback.Supervisor.ContinueIfFastTarget > 0 {
		expectedContinueTarget = readback.Supervisor.ContinueIfFastTarget
	}
	if packet.ContinueIfFastTarget != expectedContinueTarget {
		errs = append(errs, "continue_if_fast_target must match supervisor continue_if_fast_target")
	}
	if packet.CompletedNodes != readback.CompletedNodes {
		errs = append(errs, "completed_nodes must match recommendation readback")
	}
	if packet.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "ready_nodes must match recommendation readback")
	}
	if packet.BlockedNodes != readback.BlockedNodes {
		errs = append(errs, "blocked_nodes must match recommendation readback")
	}
	if packet.FailedNodes != readback.FailedNodes {
		errs = append(errs, "failed_nodes must match recommendation readback")
	}
	if packet.ExecutableReadyNodes != readback.ExecutableReadyNodes {
		errs = append(errs, "executable_ready_nodes must match recommendation readback")
	}
	if packet.FirstExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "first_executable_node must match recommendation readback")
	}
	if packet.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "lease_health_status must match recommendation readback")
	}
	if packet.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "checkpoint_freshness_status must match recommendation readback")
	}
	if packet.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "return_gate_status must match recommendation readback")
	}
	if packet.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "checkpoint_count must match recommendation readback")
	}
	if packet.EarlyReturnRiskStatus != readback.EarlyReturnRiskStatus {
		errs = append(errs, "early_return_risk_status must match recommendation readback")
	}
	expectedBudgetStatus := recommendationContinuationBudgetStatus(readback, expectedContinueTarget)
	if packet.ContinuationBudgetStatus != expectedBudgetStatus {
		errs = append(errs, "continuation_budget_status must match recommendation readback")
	}
	if packet.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must match recommendation readback")
	}
	if packet.FinalResponseReason != readback.FinalResponseReason {
		errs = append(errs, "final_response_reason must match recommendation readback")
	}
	if packet.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "exact_next_action must match recommendation readback")
	}
	if packet.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "continuation_contract_reason must match recommendation readback")
	}
	if readback.ReadyNodes > 0 {
		if packet.ReturnGateStatus != "blocked_ready_nodes_remain" {
			errs = append(errs, "ready nodes require return_gate_status=blocked_ready_nodes_remain")
		}
		if !packet.OneExecutableMutationNodeActive {
			errs = append(errs, "ready nodes require one_executable_mutation_node_active=true")
		}
		if readback.FirstExecutableNode != "" && !strings.Contains(packet.ExactNextAction, readback.FirstExecutableNode) {
			errs = append(errs, "ready nodes require exact_next_action to name first_executable_node")
		}
		if packet.FinalResponseAllowed {
			errs = append(errs, "ready nodes require final_response_allowed=false")
		}
	}
	if readback.FinalResponseAllowed {
		if packet.Status != "ready_for_final_response" {
			errs = append(errs, "final_response_allowed requires status=ready_for_final_response")
		}
		if packet.RefusesFinalResponse {
			errs = append(errs, "final_response_allowed requires refuses_final_response=false")
		}
	} else {
		if packet.Status == "ready_for_final_response" {
			errs = append(errs, "status ready_for_final_response requires final_response_allowed=true")
		}
		if !packet.RefusesFinalResponse {
			errs = append(errs, "final_response_allowed=false requires refuses_final_response=true")
		}
	}
	if packet.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if packet.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if packet.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if packet.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !packet.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationExecutionReadback(readback AtlasRecommendationReadback) AtlasRecommendationExecutionReadback {
	status := "implementation_wave_completed_generated_workgraph_ready"
	if readback.Status == "in_progress" || readback.CompletedNodes > 0 {
		status = "in_progress"
	}
	if readback.Status == "blocked" {
		status = "blocked"
	}
	if readback.FinalResponseAllowed {
		status = "completed"
	}
	readinessStatus := "pending_first_run_link"
	if readback.CompletedNodes > 0 {
		readinessStatus = "partial_run_links_recorded"
	}
	if readback.CompletedNodes == readback.TotalNodes && readback.ReadyNodes == 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		readinessStatus = "all_required_run_links_recorded"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		readinessStatus = "blocked_or_failed_run_links_need_repair"
	}
	runLinkSummary := AtlasRecommendationFoundryRunLinkReadinessSummary{
		Status:                     readinessStatus,
		Summary:                    fmt.Sprintf("%d/%d Foundry run-links recorded; ready_nodes=%d; next_executable_node=%s", readback.CompletedNodes, readback.TotalNodes, readback.ReadyNodes, readback.FirstExecutableNode),
		CompletedRunLinks:          readback.CompletedNodes,
		RequiredRunLinks:           readback.TotalNodes,
		MissingRunLinks:            readback.TotalNodes - readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		NextExecutableNode:         readback.FirstExecutableNode,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		ReturnGateStatus:           readback.ReturnGateStatus,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		ExactNextAction:            readback.ExactNextAction,
		RefusesFinalResponse:       readback.ContinuationContract.RefusesFinalResponse,
		CheckpointCount:            readback.CheckpointCount,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
	}
	continuationReasonCoverage := AtlasRecommendationContinuationReasonCoverage{
		Status:                    "coverage_sources_indexed",
		ExpectedReason:            readback.ContinuationContract.Reason,
		IndexedSources:            continuationReasonCoverageRequiredSources(),
		SourceCount:               len(continuationReasonCoverageRequiredSources()),
		FinalResponseAllowed:      readback.FinalResponseAllowed,
		RefusesFinalResponse:      readback.ContinuationContract.RefusesFinalResponse,
		ExactNextAction:           readback.ExactNextAction,
		ReturnGateStatus:          readback.ReturnGateStatus,
		LeaseHealthStatus:         readback.LeaseHealthStatus,
		CheckpointFreshnessStatus: readback.CheckpointFreshnessStatus,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	foundryDigest := digestValue(runLinkSummary)
	coverageDigest := digestValue(continuationReasonCoverage)
	sourceArtifacts := []SourceRef{
		{Ref: "foundry_run_link_readiness_summary", Digest: foundryDigest},
		{Ref: "continuation_reason_coverage", Digest: coverageDigest},
	}
	sourceArtifactRefs := make([]string, 0, len(sourceArtifacts))
	for _, source := range sourceArtifacts {
		sourceArtifactRefs = append(sourceArtifactRefs, source.Ref)
	}
	reasonArtifactAgreementSummary := AtlasRecommendationReasonArtifactAgreementSummary{
		Status:                           "agreement",
		ExpectedReason:                   continuationReasonCoverage.ExpectedReason,
		IndexedSources:                   append([]string(nil), continuationReasonCoverage.IndexedSources...),
		SourceCount:                      continuationReasonCoverage.SourceCount,
		AllRequiredSourcesIndexed:        true,
		SourceArtifactRefs:               sourceArtifactRefs,
		SourceArtifactCount:              len(sourceArtifacts),
		SourceArtifactsAgree:             true,
		FoundryRunLinkReadinessDigest:    foundryDigest,
		ContinuationReasonCoverageDigest: coverageDigest,
		FinalResponseAllowed:             readback.FinalResponseAllowed,
		RefusesFinalResponse:             readback.ContinuationContract.RefusesFinalResponse,
		ExactNextAction:                  readback.ExactNextAction,
		ReturnGateStatus:                 readback.ReturnGateStatus,
		ClaimsAuthorityAdvance:           false,
		RSIRemainsDenied:                 true,
	}
	return AtlasRecommendationExecutionReadback{
		Schema:                       "ao.atlas.long-recommendation-wave-execution.v0.3",
		Status:                       status,
		MissionID:                    readback.MissionID,
		EvidenceRoot:                 readback.EvidenceRoot,
		LeaseHealthStatus:            readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:    readback.CheckpointFreshnessStatus,
		ReturnGateStatus:             readback.ReturnGateStatus,
		ContinuationContractReason:   readback.ContinuationContract.Reason,
		ExactNextAction:              readback.ExactNextAction,
		FinalResponseAllowed:         readback.FinalResponseAllowed,
		RefusesFinalResponse:         readback.ContinuationContract.RefusesFinalResponse,
		CompletedRecommendationNodes: readback.CompletedNodes,
		TotalRecommendationNodes:     readback.TotalNodes,
		GeneratedWorkgraph: AtlasRecommendationGeneratedWorkgraphReadback{
			TotalNodes:                readback.TotalNodes,
			ReadyNodes:                readback.ReadyNodes,
			ExecutableReadyNodes:      readback.ExecutableReadyNodes,
			FirstExecutableNode:       readback.FirstExecutableNode,
			LeaseHealthStatus:         readback.LeaseHealthStatus,
			CheckpointFreshnessStatus: readback.CheckpointFreshnessStatus,
			ReturnGateStatus:          readback.ReturnGateStatus,
			CheckpointCount:           readback.CheckpointCount,
			FinalResponseAllowed:      readback.FinalResponseAllowed,
			FinalResponseReason:       readback.FinalResponseReason,
		},
		FoundryRunLinkReadinessSummary: runLinkSummary,
		ContinuationReasonCoverage:     continuationReasonCoverage,
		ReasonArtifactAgreementSummary: reasonArtifactAgreementSummary,
		SourceArtifacts:                sourceArtifacts,
	}
}

func sourceArtifactDigest(sources []SourceRef, ref string) (string, bool) {
	for _, source := range sources {
		if source.Ref == ref {
			return source.Digest, true
		}
	}
	return "", false
}
