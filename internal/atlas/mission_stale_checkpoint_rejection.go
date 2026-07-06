package atlas

import "fmt"

func BuildAtlasMissionStaleCheckpointRejection(staleReadbackPath, latestReadbackPath, promptReadbackPath string) (AtlasMissionStaleCheckpointRejection, error) {
	stale, err := LoadJSON[AtlasRecommendationReadback](staleReadbackPath)
	if err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	if err := ValidateAtlasRecommendationReadback(stale); err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	latest, err := LoadJSON[AtlasRecommendationReadback](latestReadbackPath)
	if err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	if err := ValidateAtlasRecommendationReadback(latest); err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	prompt, err := LoadJSON[AtlasRecommendationReadback](promptReadbackPath)
	if err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	if err := ValidateAtlasRecommendationReadback(prompt); err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	if stale.MissionID != latest.MissionID || stale.TargetInstance != latest.TargetInstance {
		return AtlasMissionStaleCheckpointRejection{}, fmt.Errorf("stale and latest readbacks must belong to the same mission instance")
	}
	if digestValue(prompt) != digestValue(stale) {
		return AtlasMissionStaleCheckpointRejection{}, fmt.Errorf("prompt readback must match stale readback")
	}
	if latest.CheckpointCount <= stale.CheckpointCount {
		return AtlasMissionStaleCheckpointRejection{}, fmt.Errorf("latest checkpoint_count must be newer than stale checkpoint_count")
	}
	if latest.CompletedNodes <= stale.CompletedNodes {
		return AtlasMissionStaleCheckpointRejection{}, fmt.Errorf("latest completed_nodes must be greater than stale completed_nodes")
	}
	if latest.FirstExecutableNode == stale.FirstExecutableNode {
		return AtlasMissionStaleCheckpointRejection{}, fmt.Errorf("latest first_executable_node must differ from stale prompt node")
	}

	fixture := AtlasMissionStaleCheckpointRejection{
		Schema:                            AtlasMissionStaleCheckpointRejectionContract,
		Status:                            "rejected",
		MissionID:                         latest.MissionID,
		TargetInstance:                    latest.TargetInstance,
		StaleReadbackPath:                 publicArtifactRef(staleReadbackPath),
		LatestReadbackPath:                publicArtifactRef(latestReadbackPath),
		PromptReadbackPath:                publicArtifactRef(promptReadbackPath),
		StaleReadbackDigest:               digestValue(stale),
		LatestReadbackDigest:              digestValue(latest),
		PromptReadbackDigest:              digestValue(prompt),
		StaleCheckpoint:                   missionCheckpointSnapshot(stale),
		LatestCheckpoint:                  missionCheckpointSnapshot(latest),
		PromptNextExecutableNode:          stale.FirstExecutableNode,
		ExpectedCurrentNextExecutableNode: latest.FirstExecutableNode,
		PromptExactNextAction:             stale.ExactNextAction,
		ExpectedCurrentExactNextAction:    latest.ExactNextAction,
		RejectionReason:                   "stale_checkpoint",
		ContinuationContractReason:        latest.ContinuationContract.Reason,
		FinalResponseAllowed:              false,
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
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasMissionStaleCheckpointRejection(fixture); err != nil {
		return AtlasMissionStaleCheckpointRejection{}, err
	}
	return fixture, nil
}

func ValidateAtlasMissionStaleCheckpointRejection(fixture AtlasMissionStaleCheckpointRejection) error {
	var errs []string
	requireContract(&errs, "mission_stale_checkpoint_rejection", fixture.Schema, AtlasMissionStaleCheckpointRejectionContract)
	if fixture.Status != "rejected" {
		errs = append(errs, "status must be rejected")
	}
	requireField(&errs, "mission_id", fixture.MissionID)
	requireField(&errs, "target_instance", fixture.TargetInstance)
	for field, value := range map[string]string{
		"stale_readback_path":  fixture.StaleReadbackPath,
		"latest_readback_path": fixture.LatestReadbackPath,
		"prompt_readback_path": fixture.PromptReadbackPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"stale_readback_digest":  fixture.StaleReadbackDigest,
		"latest_readback_digest": fixture.LatestReadbackDigest,
		"prompt_readback_digest": fixture.PromptReadbackDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	validateMissionCheckpointSnapshot(&errs, "stale_checkpoint", fixture.StaleCheckpoint)
	validateMissionCheckpointSnapshot(&errs, "latest_checkpoint", fixture.LatestCheckpoint)
	if fixture.PromptReadbackDigest != fixture.StaleReadbackDigest {
		errs = append(errs, "prompt_readback_digest must match stale_readback_digest")
	}
	if fixture.LatestCheckpoint.CheckpointCount <= fixture.StaleCheckpoint.CheckpointCount {
		errs = append(errs, "latest_checkpoint checkpoint_count must be greater than stale_checkpoint checkpoint_count")
	}
	if fixture.LatestCheckpoint.CompletedNodes <= fixture.StaleCheckpoint.CompletedNodes {
		errs = append(errs, "latest_checkpoint completed_nodes must be greater than stale_checkpoint completed_nodes")
	}
	if fixture.LatestCheckpoint.ReadyNodes >= fixture.StaleCheckpoint.ReadyNodes {
		errs = append(errs, "latest_checkpoint ready_nodes must be less than stale_checkpoint ready_nodes")
	}
	requireField(&errs, "prompt_next_executable_node", fixture.PromptNextExecutableNode)
	requireField(&errs, "expected_current_next_executable_node", fixture.ExpectedCurrentNextExecutableNode)
	requireField(&errs, "prompt_exact_next_action", fixture.PromptExactNextAction)
	requireField(&errs, "expected_current_exact_next_action", fixture.ExpectedCurrentExactNextAction)
	requireField(&errs, "continuation_contract_reason", fixture.ContinuationContractReason)
	for field, value := range map[string]string{
		"prompt_next_executable_node":           fixture.PromptNextExecutableNode,
		"expected_current_next_executable_node": fixture.ExpectedCurrentNextExecutableNode,
		"prompt_exact_next_action":              fixture.PromptExactNextAction,
		"expected_current_exact_next_action":    fixture.ExpectedCurrentExactNextAction,
		"continuation_contract_reason":          fixture.ContinuationContractReason,
	} {
		checkPublicPath(&errs, field, value, true)
	}
	if fixture.PromptNextExecutableNode != fixture.StaleCheckpoint.FirstExecutableNode {
		errs = append(errs, "prompt_next_executable_node must match stale_checkpoint first_executable_node")
	}
	if fixture.ExpectedCurrentNextExecutableNode != fixture.LatestCheckpoint.FirstExecutableNode {
		errs = append(errs, "expected_current_next_executable_node must match latest_checkpoint first_executable_node")
	}
	if fixture.PromptExactNextAction != fixture.StaleCheckpoint.ExactNextAction {
		errs = append(errs, "prompt_exact_next_action must match stale_checkpoint exact_next_action")
	}
	if fixture.ExpectedCurrentExactNextAction != fixture.LatestCheckpoint.ExactNextAction {
		errs = append(errs, "expected_current_exact_next_action must match latest_checkpoint exact_next_action")
	}
	if fixture.RejectionReason != "stale_checkpoint" {
		errs = append(errs, "rejection_reason must be stale_checkpoint")
	}
	if fixture.ContinuationContractReason != fixture.LatestCheckpoint.ContinuationContractReason {
		errs = append(errs, "continuation_contract_reason must match latest_checkpoint continuation_contract_reason")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	if len(fixture.SafetyBoundaries) == 0 {
		errs = append(errs, "safety_boundaries must not be empty")
	}
	for key, value := range fixture.SafetyBoundaries {
		checkPublicPath(&errs, "safety_boundaries."+key, key, true)
		if key == "rsi_remains_denied" {
			if !value {
				errs = append(errs, "safety_boundaries.rsi_remains_denied must be true")
			}
			continue
		}
		if value {
			errs = append(errs, "safety_boundaries."+key+" must be false")
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

func missionCheckpointSnapshot(readback AtlasRecommendationReadback) AtlasMissionCheckpointSnapshot {
	return AtlasMissionCheckpointSnapshot{
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		CheckpointCount:            readback.CheckpointCount,
		FirstExecutableNode:        readback.FirstExecutableNode,
		ExactNextAction:            readback.ExactNextAction,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
	}
}

func validateMissionCheckpointSnapshot(errs *[]string, field string, snapshot AtlasMissionCheckpointSnapshot) {
	if snapshot.CompletedNodes < 0 {
		*errs = append(*errs, field+".completed_nodes must not be negative")
	}
	if snapshot.ReadyNodes < 0 {
		*errs = append(*errs, field+".ready_nodes must not be negative")
	}
	if snapshot.CheckpointCount < 0 {
		*errs = append(*errs, field+".checkpoint_count must not be negative")
	}
	requireField(errs, field+".first_executable_node", snapshot.FirstExecutableNode)
	requireField(errs, field+".exact_next_action", snapshot.ExactNextAction)
	requireField(errs, field+".continuation_contract_reason", snapshot.ContinuationContractReason)
	checkPublicPath(errs, field+".first_executable_node", snapshot.FirstExecutableNode, true)
	checkPublicPath(errs, field+".exact_next_action", snapshot.ExactNextAction, true)
	checkPublicPath(errs, field+".continuation_contract_reason", snapshot.ContinuationContractReason, true)
	if snapshot.FinalResponseAllowed {
		*errs = append(*errs, field+".final_response_allowed must be false")
	}
}
