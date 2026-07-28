package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func BuildAtlasRecommendationReadback(wave AtlasRecommendationWave, workgraph Workgraph, options AtlasRecommendationReadbackOptions) (AtlasRecommendationReadback, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	if wave.TargetInstance != workgraph.TargetInstance {
		return AtlasRecommendationReadback{}, fmt.Errorf("target_instance mismatch between recommendation wave and workgraph")
	}
	if len(workgraph.Nodes) != len(wave.Tasks) {
		return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node count must match recommendation tasks")
	}
	taskByNode := map[string]AtlasRecommendationTask{}
	for _, task := range wave.Tasks {
		taskByNode[task.NodeID] = task
	}
	for _, node := range workgraph.Nodes {
		task, ok := taskByNode[node.ID]
		if !ok {
			return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node %s is not present in recommendation wave", node.ID)
		}
		if task.TaskID != node.FactoryTask.ID {
			return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node %s task_id mismatch", node.ID)
		}
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return AtlasRecommendationReadback{}, err
	}
	completed := state.NodeCounts["completed"]
	ready := state.NodeCounts["ready"]
	blocked := state.NodeCounts["blocked"]
	failed := state.NodeCounts["failed"]
	executableReady := len(state.ExecutableReadyNodeIDs)
	firstExecutable := ""
	if executableReady > 0 {
		firstExecutable = state.ExecutableReadyNodeIDs[0]
	}
	nodesComplete := completed == wave.TotalTasks && ready == 0 && blocked == 0 && failed == 0
	leaseTiming, err := buildRecommendationLeaseTiming(wave, options, nodesComplete)
	if err != nil {
		return AtlasRecommendationReadback{}, err
	}
	finalAllowed := nodesComplete && leaseTiming.MinMinutesMet
	transition := recommendationReadbackTransition(finalAllowed, nodesComplete, leaseTiming, completed, wave.MinimumTasks, ready, blocked, failed, firstExecutable)
	finalGate := recommendationFinalResponseGateEvaluation(finalAllowed, nodesComplete, leaseTiming, ready, blocked, failed, transition.ExactNextAction, firstExecutable)
	compactStatus := recommendationCompactReadbackStatus(completed, wave.TotalTasks, ready, blocked, failed, finalAllowed)
	foundryRollupStatus := "required_pending_first_node_import"
	promoterReadbackStatus := wave.PromoterReadbackStatus
	promoterNoPromotionStatus := "required_not_bound_until_promotion_evidence_exists"
	commandReadbackStatus := wave.CommandReadbackStatus
	commandTimelineStatus := "compact_timeline_required_before_closure"
	if completed > 0 {
		foundryRollupStatus = "in_progress_node_run_links_recorded"
		promoterNoPromotionStatus = "in_progress_no_promotion_recorded"
		commandTimelineStatus = "in_progress_compact_timeline_recorded"
	}
	if finalAllowed {
		foundryRollupStatus = "completed_all_node_run_links_recorded"
		promoterReadbackStatus = "no_promotion_recorded"
		promoterNoPromotionStatus = "recorded_no_promotion_for_recommendation_wave"
		commandReadbackStatus = "compact_timeline_recorded"
		commandTimelineStatus = "recorded_compact_timeline_for_completed_wave"
	}
	waveDigest := digestValue(wave)
	if strings.TrimSpace(options.WavePath) != "" {
		if digest, err := digestFile(options.WavePath); err == nil {
			waveDigest = digest
		} else {
			return AtlasRecommendationReadback{}, err
		}
	}
	workgraphDigest := digestValue(workgraph)
	if strings.TrimSpace(options.WorkgraphPath) != "" {
		if digest, err := digestFile(options.WorkgraphPath); err == nil {
			workgraphDigest = digest
		} else {
			return AtlasRecommendationReadback{}, err
		}
	}
	publicSafetyScanStatus := wave.PublicSafetyScanStatus
	if strings.TrimSpace(options.PublicSafetyScanStatus) != "" {
		publicSafetyScanStatus = strings.TrimSpace(options.PublicSafetyScanStatus)
	}
	schemaHealthStatus := "required_pending_schema_registry_health"
	if strings.TrimSpace(options.SchemaHealthStatus) != "" {
		schemaHealthStatus = strings.TrimSpace(options.SchemaHealthStatus)
	}
	readback := AtlasRecommendationReadback{
		ContractVersion:           AtlasRecommendationReadbackContract,
		MissionID:                 wave.MissionID,
		TargetInstance:            wave.TargetInstance,
		Status:                    compactStatus.ReadbackStatus,
		SourceDigest:              wave.SourceDigest,
		WaveDigest:                waveDigest,
		WorkgraphDigest:           workgraphDigest,
		EvidenceRoot:              filepath.ToSlash(strings.TrimSpace(options.EvidenceRoot)),
		Supervisor:                wave.Supervisor,
		StartedAt:                 leaseTiming.StartedAt,
		CompletedAt:               leaseTiming.CompletedAt,
		ElapsedMinutes:            leaseTiming.ElapsedMinutes,
		MinMinutesMet:             leaseTiming.MinMinutesMet,
		LeaseTimeStatus:           leaseTiming.LeaseTimeStatus,
		TotalNodes:                wave.TotalTasks,
		MinimumNodes:              wave.MinimumTasks,
		CompletedNodes:            completed,
		ReadyNodes:                ready,
		BlockedNodes:              blocked,
		FailedNodes:               failed,
		ExecutableReadyNodes:      executableReady,
		FirstExecutableNode:       firstExecutable,
		LeaseHealthStatus:         transition.LeaseHealthStatus,
		CheckpointFreshnessStatus: "fresh_checkpoint_required_after_each_node_or_timed_interval",
		StaleRouteDecisionStatus:  "fresh_atlas_supervises_foundry_owns_one_active_node",
		EarlyReturnRiskStatus:     transition.EarlyReturnRiskStatus,
		FoundryRollupStatus:       foundryRollupStatus,
		FoundryTerminalStatusReadback: map[string]string{
			"completed": "terminal_success_can_close_when_all_nodes_and_readbacks_are_complete",
			"promoted":  "terminal_success_can_close_when_promoter_and_command_agree",
			"denied":    "terminal_denial_requires_exact_missing_evidence_readback",
			"blocked":   "terminal_blocker_requires_repair_or_checkpoint_resume",
		},
		FoundryTerminalStatusExamples:   foundryTerminalStatusExamples(),
		FoundryDeniedTerminalExamples:   foundryDeniedTerminalExamples(),
		PromoterReadbackStatus:          promoterReadbackStatus,
		PromoterNoPromotionStatus:       promoterNoPromotionStatus,
		PromoterNoPromotionPlaceholders: promoterNoPromotionPlaceholders(),
		CommandReadbackStatus:           commandReadbackStatus,
		CommandTimelineStatus:           commandTimelineStatus,
		CommandTimelinePlaceholders:     commandTimelinePlaceholders(),
		PublicSafetyScanStatus:          publicSafetyScanStatus,
		SchemaHealthStatus:              schemaHealthStatus,
		ReturnGateStatus:                finalGate.ReturnGateStatus,
		CheckpointCount:                 completed,
		FinalResponseAllowed:            finalAllowed,
		FinalResponseDenialGate:         finalGate.FinalResponseDenialGate,
		FinalResponseReason:             transition.FinalResponseReason,
		ExactNextAction:                 transition.ExactNextAction,
		ContinuationContract:            finalGate.ContinuationContract,
		ExactNextActionReadback:         finalGate.ExactNextActionReadback,
		NodeEvidence:                    recommendationNodeEvidence(workgraph),
		FeatureDepthRecommendations:     featureDepthRecommendationReadback(wave.Tasks, wave.TotalTasks),
		SafetyBoundaries: map[string]bool{
			"provider_calls":                    false,
			"credential_inspection":             false,
			"direct_main_mutation":              false,
			"release_deploy_publish_upload_tag": false,
			"dependency_updates":                false,
			"auth_policy_config_widening":       false,
			"hidden_instruction_mutation":       false,
			"broad_rsi_claim":                   false,
			"rsi_remains_denied":                true,
		},
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	return readback, nil
}

func buildExactNextActionReadback(action, nextExecutableNode, returnGateStatus string, finalResponseAllowed bool) AtlasExactNextActionReadback {
	status := "continuation_required"
	if finalResponseAllowed {
		status = "ready_for_final_response"
	}
	return AtlasExactNextActionReadback{
		Status:               status,
		Action:               action,
		NextExecutableNode:   nextExecutableNode,
		ReturnGateStatus:     returnGateStatus,
		FinalResponseAllowed: finalResponseAllowed,
		Source:               "recommendation_readback",
	}
}

type atlasRecommendationReadbackTransition struct {
	FinalResponseReason   string
	ExactNextAction       string
	LeaseHealthStatus     string
	EarlyReturnRiskStatus string
}

func recommendationReadbackTransition(finalAllowed bool, nodesComplete bool, leaseTiming atlasRecommendationLeaseTiming, completed, minimum, ready, blocked, failed int, firstExecutable string) atlasRecommendationReadbackTransition {
	transition := atlasRecommendationReadbackTransition{
		FinalResponseReason:   "ready nodes or exact next actions remain",
		ExactNextAction:       "Complete dependency chain so the next Atlas recommendation node becomes executable-ready.",
		LeaseHealthStatus:     "minimum_unmet",
		EarlyReturnRiskStatus: "blocked_final_response_minimum_unmet",
	}
	if finalAllowed {
		transition.FinalResponseReason = "all generated nodes complete and no ready nodes remain"
		transition.ExactNextAction = "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks."
		transition.LeaseHealthStatus = "all_generated_nodes_complete"
		transition.EarlyReturnRiskStatus = "cleared_no_ready_nodes_remain"
		return transition
	}
	if nodesComplete && leaseTiming.LeaseTimeStatus == "lease_timing_missing" {
		transition.FinalResponseReason = "minimum lease timing evidence missing"
		transition.ExactNextAction = "Record started_at, completed_at, and elapsed_minutes before evaluating final response for the long-run lease."
		transition.LeaseHealthStatus = "minimum_minutes_timing_missing"
		transition.EarlyReturnRiskStatus = "blocked_final_response_minimum_timing_missing"
		return transition
	}
	if nodesComplete && !leaseTiming.MinMinutesMet {
		transition.FinalResponseReason = "minimum lease minutes unmet"
		transition.ExactNextAction = "Generate and execute the next useful Atlas recommendation wave until elapsed_minutes meets supervisor.min_minutes, or record a true hard blocker."
		transition.LeaseHealthStatus = "minimum_minutes_unmet_continue_next_wave"
		transition.EarlyReturnRiskStatus = "blocked_final_response_minimum_minutes_unmet"
		return transition
	}
	if blocked > 0 || failed > 0 {
		transition.FinalResponseReason = "true hard blocker requires exact repair evidence"
		transition.ExactNextAction = "Resolve blocked or failed recommendation node with exact repair evidence."
		transition.LeaseHealthStatus = "hard_blocker_requires_repair"
		transition.EarlyReturnRiskStatus = "hard_blocker_requires_exact_missing_evidence"
		return transition
	}
	if completed >= minimum {
		transition.LeaseHealthStatus = "minimum_met_continue_if_fast"
	}
	if ready > 0 {
		transition.EarlyReturnRiskStatus = "blocked_final_response_ready_nodes_remain"
	}
	if firstExecutable != "" {
		transition.ExactNextAction = fmt.Sprintf("Emit Foundry import for %s and execute exactly one active node.", firstExecutable)
	}
	return transition
}

type atlasRecommendationCompactReadbackStatus struct {
	ReadbackStatus       string
	NodeCompletionStatus string
}

func recommendationCompactReadbackStatus(completed, total, ready, blocked, failed int, finalAllowed bool) atlasRecommendationCompactReadbackStatus {
	status := atlasRecommendationCompactReadbackStatus{
		ReadbackStatus:       "ready",
		NodeCompletionStatus: "nodes_in_progress",
	}
	if finalAllowed {
		status.ReadbackStatus = "completed"
	} else if blocked > 0 || failed > 0 {
		status.ReadbackStatus = "blocked"
	} else if completed > 0 {
		status.ReadbackStatus = "in_progress"
	}
	if completed == total && ready == 0 && blocked == 0 && failed == 0 {
		status.NodeCompletionStatus = "all_nodes_complete"
	}
	if blocked > 0 || failed > 0 {
		status.NodeCompletionStatus = "blocked_or_failed_nodes_present"
	}
	return status
}

func commandTimelinePlaceholders() []AtlasCommandTimelinePlaceholder {
	return []AtlasCommandTimelinePlaceholder{
		{
			Slot:                        "checkpoint",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind completed_nodes, ready_nodes, checkpoint_count, and elapsed_minutes into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "exact_next_action",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind exact_next_action and first_executable_node into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "return_gate",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind return_gate_status and final_response_allowed into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
	}
}

func promoterNoPromotionPlaceholders() []AtlasPromoterNoPromotionPlaceholder {
	return []AtlasPromoterNoPromotionPlaceholder{
		{
			Slot:                        "promotion_claim",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind promotion_claimed=false and the no-promotion summary before closure.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "rsi_boundary",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind rsi_remains_denied=true and next_denied_class=RSI before closure.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "authority_advance",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind claims_authority_advance=false plus no scheduling, execution, or approval authority.",
			RequiredBeforeFinalResponse: true,
		},
	}
}

func foundryTerminalStatusExamples() []AtlasFoundryTerminalStatusExample {
	return []AtlasFoundryTerminalStatusExample{
		{
			SourceStatus:     "completed",
			NormalizedStatus: "completed",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Foundry rollup reports completed, all node evidence exists, and no ready nodes remain.",
		},
		{
			SourceStatus:     "promoted",
			NormalizedStatus: "completed",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Promoter and Command agree promotion is terminal, RSI remains denied, and no ready nodes remain.",
		},
		{
			SourceStatus:     "denied",
			NormalizedStatus: "denied",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Denial readback includes exact missing evidence, no ready repair node remains, and no authority advance is claimed.",
		},
		{
			SourceStatus:     "blocked",
			NormalizedStatus: "blocked",
			Terminal:         true,
			CanCloseMission:  false,
			RequiredReadback: "Blocker readback names the exact repair or resume action before final response can close.",
		},
	}
}

func foundryDeniedTerminalExamples() []AtlasFoundryDeniedTerminalExample {
	return []AtlasFoundryDeniedTerminalExample{
		{
			DenialReason:                 "missing_node_evidence",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup names the missing node id, missing evidence key, and expected evidence path.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
		{
			DenialReason:                 "missing_stop_gate_evidence",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup names the uncleared stop gate and the exact artifact needed before promotion can be reconsidered.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
		{
			DenialReason:                 "forbidden_surface_or_rsi_claim",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup records the forbidden surface or RSI claim, refuses promotion, and keeps RSI denied.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
	}
}

func recommendationReturnGateStatus(finalAllowed bool, nodesComplete bool, leaseTiming atlasRecommendationLeaseTiming, ready, blocked, failed int) string {
	if finalAllowed {
		return "final_response_allowed"
	}
	if blocked > 0 || failed > 0 {
		return "blocked_hard_blocker"
	}
	if nodesComplete && leaseTiming.LeaseTimeStatus == "lease_timing_missing" {
		return "blocked_lease_timing_missing"
	}
	if nodesComplete && !leaseTiming.MinMinutesMet {
		return "blocked_minimum_minutes_unmet"
	}
	if ready > 0 {
		return "blocked_ready_nodes_remain"
	}
	return "blocked_no_executable_ready_node"
}

type atlasRecommendationFinalResponseGateEvaluation struct {
	ReturnGateStatus        string
	FinalResponseDenialGate string
	ContinuationContract    AtlasContinuationContract
	ExactNextActionReadback AtlasExactNextActionReadback
}

func recommendationFinalResponseGateEvaluation(finalAllowed bool, nodesComplete bool, leaseTiming atlasRecommendationLeaseTiming, ready, blocked, failed int, exactNextAction, firstExecutable string) atlasRecommendationFinalResponseGateEvaluation {
	returnGateStatus := recommendationReturnGateStatus(finalAllowed, nodesComplete, leaseTiming, ready, blocked, failed)
	denialReason := recommendationReturnGateDenialReason(finalAllowed, returnGateStatus)
	return atlasRecommendationFinalResponseGateEvaluation{
		ReturnGateStatus:        returnGateStatus,
		FinalResponseDenialGate: denialReason.FinalResponseDenialGate,
		ContinuationContract:    buildAtlasContinuationContract(ready, exactNextAction, returnGateStatus, finalAllowed),
		ExactNextActionReadback: buildExactNextActionReadback(exactNextAction, firstExecutable, returnGateStatus, finalAllowed),
	}
}

func recommendationFinalResponseDenialGate(finalAllowed bool, returnGateStatus string) string {
	return recommendationReturnGateDenialReason(finalAllowed, returnGateStatus).FinalResponseDenialGate
}

type atlasRecommendationReturnGateDenialReason struct {
	Code                    string
	Summary                 string
	FinalResponseDenialGate string
	AllowsFinalResponse     bool
}

func recommendationReturnGateDenialReason(finalAllowed bool, returnGateStatus string) atlasRecommendationReturnGateDenialReason {
	if finalAllowed {
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "allow_final_response",
			Summary:                 "final response allowed by return gate",
			FinalResponseDenialGate: "allow_final_response",
			AllowsFinalResponse:     true,
		}
	}
	if returnGateStatus == "blocked_hard_blocker" {
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "hard_blocker",
			Summary:                 "hard blocker prevents final response",
			FinalResponseDenialGate: "blocked_hard_blocker",
		}
	}
	switch returnGateStatus {
	case "blocked_lease_timing_missing":
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "lease_timing_missing",
			Summary:                 "lease timing evidence missing prevents final response",
			FinalResponseDenialGate: "deny_ready_nodes_or_exact_next_action_remain",
		}
	case "blocked_minimum_minutes_unmet":
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "minimum_minutes_unmet",
			Summary:                 "minimum minutes unmet prevents final response",
			FinalResponseDenialGate: "deny_ready_nodes_or_exact_next_action_remain",
		}
	case "blocked_ready_nodes_remain":
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "ready_nodes_remain",
			Summary:                 "ready nodes remain before final response",
			FinalResponseDenialGate: "deny_ready_nodes_or_exact_next_action_remain",
		}
	case "blocked_no_executable_ready_node":
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "no_executable_ready_node",
			Summary:                 "no executable ready node is available before final response",
			FinalResponseDenialGate: "deny_ready_nodes_or_exact_next_action_remain",
		}
	default:
		return atlasRecommendationReturnGateDenialReason{
			Code:                    "return_gate_denied",
			Summary:                 "return gate denies final response",
			FinalResponseDenialGate: "deny_ready_nodes_or_exact_next_action_remain",
		}
	}
}

func buildAtlasContinuationContract(readyNodes int, exactNextAction, returnGateStatus string, finalResponseAllowed bool) AtlasContinuationContract {
	status := "ready_for_final_response"
	refusesFinalResponse := false
	reason := "final response allowed by recommendation readback"
	if !finalResponseAllowed {
		status = "continuation_required"
		refusesFinalResponse = true
		reason = atlasContinuationContractReason(readyNodes, exactNextAction, returnGateStatus)
		if readyNodes == 0 && strings.TrimSpace(exactNextAction) == "" {
			status = "blocked"
		}
	}
	return AtlasContinuationContract{
		ContractVersion:      "ao.atlas.continuation-contract.v0.1",
		Status:               status,
		ReadyNodes:           readyNodes,
		ExactNextAction:      exactNextAction,
		ReturnGateStatus:     returnGateStatus,
		FinalResponseAllowed: finalResponseAllowed,
		RefusesFinalResponse: refusesFinalResponse,
		Reason:               reason,
		Source:               "recommendation_readback",
	}
}

func atlasContinuationContractReason(readyNodes int, exactNextAction, returnGateStatus string) string {
	hasExactNextAction := strings.TrimSpace(exactNextAction) != ""
	switch {
	case readyNodes > 0 && hasExactNextAction:
		return "ready_nodes_or_exact_next_action_remain"
	case readyNodes > 0:
		return "ready_nodes_remain"
	case hasExactNextAction:
		return "exact_next_action_remains"
	default:
		return returnGateStatus
	}
}

type atlasRecommendationLeaseTiming struct {
	StartedAt       string
	CompletedAt     string
	ElapsedMinutes  int
	MinMinutesMet   bool
	LeaseTimeStatus string
}

func buildRecommendationLeaseTiming(wave AtlasRecommendationWave, options AtlasRecommendationReadbackOptions, nodesComplete bool) (atlasRecommendationLeaseTiming, error) {
	minMinutes := wave.EstimatedMinutes
	if wave.Supervisor != nil {
		minMinutes = wave.Supervisor.MinMinutes
	}
	startedAt := strings.TrimSpace(options.StartedAt)
	completedAt := strings.TrimSpace(options.CompletedAt)
	elapsedMinutes := options.ElapsedMinutes
	if elapsedMinutes < 0 {
		return atlasRecommendationLeaseTiming{}, fmt.Errorf("elapsed_minutes must be non-negative")
	}
	var started time.Time
	var completed time.Time
	var hasStarted bool
	var hasCompleted bool
	if startedAt != "" {
		parsed, err := time.Parse(time.RFC3339, startedAt)
		if err != nil {
			return atlasRecommendationLeaseTiming{}, fmt.Errorf("started_at must be RFC3339: %w", err)
		}
		started = parsed
		hasStarted = true
	}
	if completedAt != "" {
		parsed, err := time.Parse(time.RFC3339, completedAt)
		if err != nil {
			return atlasRecommendationLeaseTiming{}, fmt.Errorf("completed_at must be RFC3339: %w", err)
		}
		completed = parsed
		hasCompleted = true
	}
	if hasStarted && hasCompleted && completed.Before(started) {
		return atlasRecommendationLeaseTiming{}, fmt.Errorf("completed_at must be greater than or equal to started_at")
	}
	hasTimingEvidence := elapsedMinutes > 0 ||
		startedAt != "" ||
		completedAt != "" ||
		strings.TrimSpace(options.LeaseTimingMode) != ""
	if elapsedMinutes == 0 && hasStarted && hasCompleted {
		elapsedMinutes = ceilDurationMinutes(completed.Sub(started))
	}
	status := "in_progress_timing_pending"
	minMinutesMet := false
	if minMinutes <= 0 {
		status = "minimum_minutes_not_required"
		minMinutesMet = true
	} else if hasTimingEvidence {
		if elapsedMinutes >= minMinutes {
			status = "minimum_minutes_met"
			minMinutesMet = true
		} else {
			status = "minimum_minutes_unmet"
		}
	} else if nodesComplete {
		status = "lease_timing_missing"
	}
	return atlasRecommendationLeaseTiming{
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
		ElapsedMinutes:  elapsedMinutes,
		MinMinutesMet:   minMinutesMet,
		LeaseTimeStatus: status,
	}, nil
}

func ceilDurationMinutes(duration time.Duration) int {
	if duration <= 0 {
		return 0
	}
	minutes := int(duration / time.Minute)
	if duration%time.Minute != 0 {
		minutes++
	}
	return minutes
}

func ValidateAtlasRecommendationReadback(readback AtlasRecommendationReadback) error {
	var errs []string
	requireContract(&errs, "atlas_recommendation_readback", readback.ContractVersion, AtlasRecommendationReadbackContract)
	requireField(&errs, "mission_id", readback.MissionID)
	requireField(&errs, "target_instance", readback.TargetInstance)
	if !oneOf(readback.Status, "ready", "in_progress", "blocked", "completed") {
		errs = append(errs, "status must be ready, in_progress, blocked, or completed")
	}
	if !digestPattern.MatchString(readback.SourceDigest) {
		errs = append(errs, "source_digest must be sha256 digest")
	}
	if strings.TrimSpace(readback.WaveDigest) != "" && !digestPattern.MatchString(readback.WaveDigest) {
		errs = append(errs, "wave_digest must be sha256 digest")
	}
	if strings.TrimSpace(readback.WorkgraphDigest) != "" && !digestPattern.MatchString(readback.WorkgraphDigest) {
		errs = append(errs, "workgraph_digest must be sha256 digest")
	}
	if readback.TotalNodes < 1 {
		errs = append(errs, "total_nodes must be positive")
	}
	if readback.MinimumNodes < 1 || readback.MinimumNodes > readback.TotalNodes {
		errs = append(errs, "minimum_nodes must be between 1 and total_nodes")
	}
	if readback.CompletedNodes+readback.ReadyNodes+readback.BlockedNodes+readback.FailedNodes != readback.TotalNodes {
		errs = append(errs, "node counts must sum to total_nodes")
	}
	if readback.ElapsedMinutes < 0 {
		errs = append(errs, "elapsed_minutes must be non-negative")
	}
	requireField(&errs, "lease_time_status", readback.LeaseTimeStatus)
	if readback.Supervisor != nil && readback.Supervisor.MinMinutes > 0 {
		if readback.MinMinutesMet && readback.ElapsedMinutes < readback.Supervisor.MinMinutes {
			errs = append(errs, "min_minutes_met requires elapsed_minutes to meet supervisor.min_minutes")
		}
		if readback.FinalResponseAllowed && !readback.MinMinutesMet {
			errs = append(errs, "final_response_allowed requires min_minutes_met")
		}
	}
	if readback.FinalResponseAllowed {
		requireField(&errs, "started_at", readback.StartedAt)
		requireField(&errs, "completed_at", readback.CompletedAt)
		if readback.ElapsedMinutes == 0 {
			errs = append(errs, "final_response_allowed requires elapsed_minutes")
		}
	}
	if readback.ExecutableReadyNodes > readback.ReadyNodes {
		errs = append(errs, "executable_ready_nodes cannot exceed ready_nodes")
	}
	requireField(&errs, "lease_health_status", readback.LeaseHealthStatus)
	requireField(&errs, "checkpoint_freshness_status", readback.CheckpointFreshnessStatus)
	requireField(&errs, "stale_route_decision_status", readback.StaleRouteDecisionStatus)
	requireField(&errs, "early_return_risk_status", readback.EarlyReturnRiskStatus)
	requireField(&errs, "foundry_rollup_status", readback.FoundryRollupStatus)
	for _, key := range []string{"completed", "promoted", "denied", "blocked"} {
		requireField(&errs, "foundry_terminal_status_readback."+key, readback.FoundryTerminalStatusReadback[key])
	}
	if err := validateFoundryTerminalStatusExamples(readback.FoundryTerminalStatusExamples); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateFoundryDeniedTerminalExamples(readback.FoundryDeniedTerminalExamples); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "promoter_readback_status", readback.PromoterReadbackStatus)
	requireField(&errs, "promoter_no_promotion_status", readback.PromoterNoPromotionStatus)
	if err := validatePromoterNoPromotionPlaceholders(readback.PromoterNoPromotionPlaceholders); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "command_readback_status", readback.CommandReadbackStatus)
	requireField(&errs, "command_timeline_status", readback.CommandTimelineStatus)
	if err := validateCommandTimelinePlaceholders(readback.CommandTimelinePlaceholders); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "public_safety_scan_status", readback.PublicSafetyScanStatus)
	if strings.TrimSpace(readback.ReturnGateStatus) != "" &&
		!oneOf(readback.ReturnGateStatus, "final_response_allowed", "blocked_hard_blocker", "blocked_lease_timing_missing", "blocked_minimum_minutes_unmet", "blocked_ready_nodes_remain", "blocked_no_executable_ready_node") {
		errs = append(errs, "return_gate_status has unsupported value")
	}
	if readback.CheckpointCount < 0 {
		errs = append(errs, "checkpoint_count must be non-negative")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && readback.CheckpointCount != readback.CompletedNodes {
		errs = append(errs, "checkpoint_count must match completed_nodes when return_gate_status is recorded")
	}
	requireField(&errs, "final_response_denial_gate", readback.FinalResponseDenialGate)
	requireField(&errs, "final_response_reason", readback.FinalResponseReason)
	requireField(&errs, "exact_next_action", readback.ExactNextAction)
	if err := validateAtlasContinuationContract(readback); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateExactNextActionReadback(readback); err != nil {
		errs = append(errs, err.Error())
	}
	if readback.ReadyNodes > 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		if readback.FinalResponseAllowed {
			errs = append(errs, "ready nodes require final_response_allowed=false")
		}
		if readback.ReturnGateStatus != "blocked_ready_nodes_remain" {
			errs = append(errs, "ready nodes require return_gate_status=blocked_ready_nodes_remain")
		}
		if readback.FinalResponseReason != "ready nodes or exact next actions remain" {
			errs = append(errs, "ready nodes require final_response_reason=ready nodes or exact next actions remain")
		}
		if readback.ExecutableReadyNodes > 0 && !strings.Contains(readback.ExactNextAction, readback.FirstExecutableNode) {
			errs = append(errs, "ready nodes require exact_next_action to name first_executable_node")
		}
	}
	if readback.FinalResponseAllowed {
		if readback.Status != "completed" {
			errs = append(errs, "final_response_allowed requires status=completed")
		}
		if readback.ReturnGateStatus != "final_response_allowed" {
			errs = append(errs, "final_response_allowed requires return_gate_status=final_response_allowed")
		}
		if readback.FinalResponseReason != "all generated nodes complete and no ready nodes remain" {
			errs = append(errs, "final_response_allowed requires final_response_reason=all generated nodes complete and no ready nodes remain")
		}
		if readback.ExactNextAction != "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks." {
			errs = append(errs, "final_response_allowed requires final exact_next_action")
		}
		if readback.FinalResponseDenialGate != "allow_final_response" {
			errs = append(errs, "final_response_allowed requires final_response_denial_gate=allow_final_response")
		}
	} else if readback.ReturnGateStatus == "blocked_hard_blocker" {
		if readback.FinalResponseDenialGate != "blocked_hard_blocker" {
			errs = append(errs, "hard blocker requires final_response_denial_gate=blocked_hard_blocker")
		}
	} else if readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "" {
		if readback.FinalResponseDenialGate != "deny_ready_nodes_or_exact_next_action_remain" {
			errs = append(errs, "ready nodes or exact next action require final_response_denial_gate=deny_ready_nodes_or_exact_next_action_remain")
		}
	}
	if readback.FinalResponseAllowed && (readback.ReadyNodes > 0 || readback.BlockedNodes > 0 || readback.FailedNodes > 0) {
		errs = append(errs, "final_response_allowed requires no ready, blocked, or failed nodes")
	}
	if len(readback.NodeEvidence) != readback.TotalNodes {
		errs = append(errs, "node_evidence length must match total_nodes")
	}
	if len(readback.FeatureDepthRecommendations) < 10 && readback.TotalNodes >= 10 {
		errs = append(errs, "feature_depth_recommendations must include at least 10 tasks")
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
	for i, evidence := range readback.NodeEvidence {
		prefix := fmt.Sprintf("node_evidence[%d]", i)
		requireField(&errs, prefix+".node_id", evidence.NodeID)
		requireField(&errs, prefix+".task_id", evidence.TaskID)
		requireField(&errs, prefix+".status", evidence.Status)
		requireField(&errs, prefix+".node_gate", evidence.NodeGate)
		requireField(&errs, prefix+".candidate_record", evidence.CandidateRecord)
		requireField(&errs, prefix+".rollback_record", evidence.RollbackRecord)
		requireField(&errs, prefix+".implementation_evidence", evidence.ImplementationEvidence)
		requireField(&errs, prefix+".tests", evidence.Tests)
		requireField(&errs, prefix+".verification", evidence.Verification)
		requireField(&errs, prefix+".public_safety_wording", evidence.PublicSafetyWording)
		requireField(&errs, prefix+".promoter_readback", evidence.PromoterReadback)
		requireField(&errs, prefix+".command_readback", evidence.CommandReadback)
		requireList(&errs, prefix+".required_gates", evidence.RequiredGates)
		requireList(&errs, prefix+".verification_commands", evidence.VerificationCommands)
	}
	return joinErrors(errs)
}

func validateAtlasContinuationContract(readback AtlasRecommendationReadback) error {
	contract := readback.ContinuationContract
	if contract.ContractVersion != "ao.atlas.continuation-contract.v0.1" {
		return fmt.Errorf("continuation_contract.contract_version must be ao.atlas.continuation-contract.v0.1")
	}
	if contract.Source != "recommendation_readback" {
		return fmt.Errorf("continuation_contract.source must be recommendation_readback")
	}
	if contract.ReadyNodes != readback.ReadyNodes {
		return fmt.Errorf("continuation_contract.ready_nodes must match ready_nodes")
	}
	if contract.ExactNextAction != readback.ExactNextAction {
		return fmt.Errorf("continuation_contract.exact_next_action must match exact_next_action")
	}
	if contract.ReturnGateStatus != readback.ReturnGateStatus {
		return fmt.Errorf("continuation_contract.return_gate_status must match return_gate_status")
	}
	if contract.FinalResponseAllowed != readback.FinalResponseAllowed {
		return fmt.Errorf("continuation_contract.final_response_allowed must match final_response_allowed")
	}
	if strings.TrimSpace(contract.Reason) == "" {
		return fmt.Errorf("continuation_contract.reason is required")
	}
	if readback.FinalResponseAllowed {
		if contract.Status != "ready_for_final_response" {
			return fmt.Errorf("continuation_contract.status must be ready_for_final_response when final response is allowed")
		}
		if contract.RefusesFinalResponse {
			return fmt.Errorf("continuation_contract.refuses_final_response must be false when final response is allowed")
		}
		return nil
	}
	if readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "" {
		if contract.Status != "continuation_required" {
			return fmt.Errorf("continuation_contract.status must be continuation_required while ready nodes or exact next action remain")
		}
		if !contract.RefusesFinalResponse {
			return fmt.Errorf("continuation_contract.refuses_final_response must be true while ready nodes or exact next action remain")
		}
		expectedReason := atlasContinuationContractReason(readback.ReadyNodes, readback.ExactNextAction, readback.ReturnGateStatus)
		if contract.Reason != expectedReason {
			return fmt.Errorf("continuation_contract.reason must be %s while ready nodes or exact next action remain", expectedReason)
		}
	}
	return nil
}

func validateExactNextActionReadback(readback AtlasRecommendationReadback) error {
	action := readback.ExactNextActionReadback
	if strings.TrimSpace(action.Status) == "" {
		return fmt.Errorf("exact_next_action_readback.status is required")
	}
	if action.Action != readback.ExactNextAction {
		return fmt.Errorf("exact_next_action_readback.action must match exact_next_action")
	}
	if action.NextExecutableNode != readback.FirstExecutableNode {
		return fmt.Errorf("exact_next_action_readback.next_executable_node must match first_executable_node")
	}
	if action.ReturnGateStatus != readback.ReturnGateStatus {
		return fmt.Errorf("exact_next_action_readback.return_gate_status must match return_gate_status")
	}
	if action.FinalResponseAllowed != readback.FinalResponseAllowed {
		return fmt.Errorf("exact_next_action_readback.final_response_allowed must match final_response_allowed")
	}
	if action.Source != "recommendation_readback" {
		return fmt.Errorf("exact_next_action_readback.source must be recommendation_readback")
	}
	if readback.FinalResponseAllowed {
		if action.Status != "ready_for_final_response" {
			return fmt.Errorf("exact_next_action_readback.status must be ready_for_final_response")
		}
	} else if action.Status != "continuation_required" {
		return fmt.Errorf("exact_next_action_readback.status must be continuation_required")
	}
	return nil
}

func validatePromoterNoPromotionPlaceholders(placeholders []AtlasPromoterNoPromotionPlaceholder) error {
	required := map[string]bool{
		"promotion_claim":   false,
		"rsi_boundary":      false,
		"authority_advance": false,
	}
	if len(placeholders) < len(required) {
		return fmt.Errorf("promoter_no_promotion_placeholders must include promotion_claim, rsi_boundary, and authority_advance")
	}
	for _, placeholder := range placeholders {
		slot := strings.TrimSpace(placeholder.Slot)
		if _, ok := required[slot]; !ok {
			return fmt.Errorf("promoter_no_promotion_placeholders has unsupported slot %q", placeholder.Slot)
		}
		if required[slot] {
			return fmt.Errorf("promoter_no_promotion_placeholders duplicate slot %q", slot)
		}
		required[slot] = true
		if placeholder.Source != "recommendation_readback" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s source must be recommendation_readback", slot)
		}
		if placeholder.Status != "pending_promoter_no_promotion" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s status must be pending_promoter_no_promotion", slot)
		}
		if strings.TrimSpace(placeholder.Summary) == "" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s summary is required", slot)
		}
		if !placeholder.RequiredBeforeFinalResponse {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s must be required before final response", slot)
		}
	}
	for slot, seen := range required {
		if !seen {
			return fmt.Errorf("promoter_no_promotion_placeholders missing %s", slot)
		}
	}
	return nil
}

func validateCommandTimelinePlaceholders(placeholders []AtlasCommandTimelinePlaceholder) error {
	required := map[string]bool{
		"checkpoint":        false,
		"exact_next_action": false,
		"return_gate":       false,
	}
	if len(placeholders) < len(required) {
		return fmt.Errorf("command_timeline_placeholders must include checkpoint, exact_next_action, and return_gate")
	}
	for _, placeholder := range placeholders {
		slot := strings.TrimSpace(placeholder.Slot)
		if _, ok := required[slot]; !ok {
			return fmt.Errorf("command_timeline_placeholders has unsupported slot %q", placeholder.Slot)
		}
		if required[slot] {
			return fmt.Errorf("command_timeline_placeholders duplicate slot %q", slot)
		}
		required[slot] = true
		if placeholder.Source != "recommendation_readback" {
			return fmt.Errorf("command_timeline_placeholders.%s source must be recommendation_readback", slot)
		}
		if placeholder.Status != "pending_command_timeline" {
			return fmt.Errorf("command_timeline_placeholders.%s status must be pending_command_timeline", slot)
		}
		if strings.TrimSpace(placeholder.Summary) == "" {
			return fmt.Errorf("command_timeline_placeholders.%s summary is required", slot)
		}
		if !placeholder.RequiredBeforeFinalResponse {
			return fmt.Errorf("command_timeline_placeholders.%s must be required before final response", slot)
		}
	}
	for slot, seen := range required {
		if !seen {
			return fmt.Errorf("command_timeline_placeholders missing %s", slot)
		}
	}
	return nil
}

func validateFoundryTerminalStatusExamples(examples []AtlasFoundryTerminalStatusExample) error {
	required := map[string]bool{
		"completed": false,
		"promoted":  false,
		"denied":    false,
		"blocked":   false,
	}
	if len(examples) != len(required) {
		return fmt.Errorf("foundry_terminal_status_examples must include completed, promoted, denied, and blocked examples")
	}
	for _, example := range examples {
		source := strings.TrimSpace(example.SourceStatus)
		if _, ok := required[source]; !ok {
			return fmt.Errorf("foundry_terminal_status_examples has unsupported source_status %q", example.SourceStatus)
		}
		if required[source] {
			return fmt.Errorf("foundry_terminal_status_examples duplicate source_status %q", source)
		}
		required[source] = true
		if strings.TrimSpace(example.NormalizedStatus) == "" {
			return fmt.Errorf("foundry_terminal_status_examples.%s normalized_status is required", source)
		}
		if strings.TrimSpace(example.RequiredReadback) == "" {
			return fmt.Errorf("foundry_terminal_status_examples.%s required_readback is required", source)
		}
		if !example.Terminal {
			return fmt.Errorf("foundry_terminal_status_examples.%s must be terminal", source)
		}
		switch source {
		case "completed":
			if example.NormalizedStatus != "completed" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.completed must close as completed")
			}
		case "promoted":
			if example.NormalizedStatus != "completed" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.promoted must close as completed")
			}
		case "denied":
			if example.NormalizedStatus != "denied" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.denied must close with exact denial evidence")
			}
		case "blocked":
			if example.NormalizedStatus != "blocked" || example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.blocked must remain open for repair or resume")
			}
		}
	}
	for source, seen := range required {
		if !seen {
			return fmt.Errorf("foundry_terminal_status_examples missing %s", source)
		}
	}
	return nil
}

func validateFoundryDeniedTerminalExamples(examples []AtlasFoundryDeniedTerminalExample) error {
	required := map[string]bool{
		"missing_node_evidence":          false,
		"missing_stop_gate_evidence":     false,
		"forbidden_surface_or_rsi_claim": false,
	}
	if len(examples) < len(required) {
		return fmt.Errorf("foundry_denied_terminal_examples must include missing node, stop gate, and forbidden surface examples")
	}
	for _, example := range examples {
		reason := strings.TrimSpace(example.DenialReason)
		if _, ok := required[reason]; !ok {
			return fmt.Errorf("foundry_denied_terminal_examples has unsupported denial_reason %q", example.DenialReason)
		}
		if required[reason] {
			return fmt.Errorf("foundry_denied_terminal_examples duplicate denial_reason %q", reason)
		}
		required[reason] = true
		if example.NormalizedStatus != "denied" {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must normalize to denied", reason)
		}
		if !example.Terminal {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must be terminal", reason)
		}
		if !example.CanCloseMission {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must be closable as final denial", reason)
		}
		if !example.RequiresExactMissingEvidence {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must require exact missing evidence", reason)
		}
		if strings.TrimSpace(example.RequiredReadback) == "" {
			return fmt.Errorf("foundry_denied_terminal_examples.%s required_readback is required", reason)
		}
		if !example.RSIRemainsDenied {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must keep RSI denied", reason)
		}
		if example.AuthorityAdvanceClaimed {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must not claim authority advance", reason)
		}
	}
	for reason, seen := range required {
		if !seen {
			return fmt.Errorf("foundry_denied_terminal_examples missing %s", reason)
		}
	}
	return nil
}

func ValidateAtlasRecommendationLeaseStart(leaseStart AtlasRecommendationLeaseStart) error {
	var errs []string
	if leaseStart.Schema != "ao.atlas.recommendation-lease-start.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-lease-start.v0.1")
	}
	if leaseStart.Status != "started" {
		errs = append(errs, "status must be started")
	}
	requireField(&errs, "mission_id", leaseStart.MissionID)
	requireField(&errs, "target_instance", leaseStart.TargetInstance)
	requireField(&errs, "started_at", leaseStart.StartedAt)
	if strings.TrimSpace(leaseStart.StartedAt) != "" {
		if _, err := time.Parse(time.RFC3339, leaseStart.StartedAt); err != nil {
			errs = append(errs, "started_at must be RFC3339")
		}
	}
	if leaseStart.MinMinutes < 1 {
		errs = append(errs, "min_minutes must be positive")
	}
	if leaseStart.MaxMinutes < leaseStart.MinMinutes {
		errs = append(errs, "max_minutes must be greater than or equal to min_minutes")
	}
	if leaseStart.ContinueIfFastTarget < 1 {
		errs = append(errs, "continue_if_fast_target must be positive")
	}
	if !digestPattern.MatchString(leaseStart.WaveDigest) {
		errs = append(errs, "wave_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(leaseStart.WorkgraphDigest) {
		errs = append(errs, "workgraph_digest must be sha256 digest")
	}
	if leaseStart.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false for lease start marker")
	}
	requireField(&errs, "final_response_reason", leaseStart.FinalResponseReason)
	if leaseStart.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if leaseStart.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if leaseStart.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if leaseStart.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if leaseStart.CallsProviders {
		errs = append(errs, "calls_providers must be false")
	}
	if leaseStart.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	return joinErrors(errs)
}
