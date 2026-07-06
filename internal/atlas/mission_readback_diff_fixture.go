package atlas

import "fmt"

func BuildAtlasMissionReadbackDiffFixture(sourceReadbackPath, targetReadbackPath, deltaPath string) (AtlasMissionReadbackDiffFixture, error) {
	source, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	if err := ValidateAtlasRecommendationReadback(source); err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	target, err := LoadJSON[AtlasRecommendationReadback](targetReadbackPath)
	if err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	if err := ValidateAtlasRecommendationReadback(target); err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	delta, err := LoadJSON[AtlasMissionReadbackDelta](deltaPath)
	if err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	if err := ValidateAtlasMissionReadbackDelta(delta); err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	sourceRef := publicArtifactRef(sourceReadbackPath)
	targetRef := publicArtifactRef(targetReadbackPath)
	deltaRef := publicArtifactRef(deltaPath)
	if delta.SourceReadbackPath != sourceRef {
		return AtlasMissionReadbackDiffFixture{}, fmt.Errorf("delta source_readback_path must match source readback")
	}
	if delta.TargetReadbackPath != targetRef {
		return AtlasMissionReadbackDiffFixture{}, fmt.Errorf("delta target_readback_path must match target readback")
	}
	if delta.SourceReadbackDigest != digestValue(source) {
		return AtlasMissionReadbackDiffFixture{}, fmt.Errorf("delta source_readback_digest must match source readback")
	}
	if delta.TargetReadbackDigest != digestValue(target) {
		return AtlasMissionReadbackDiffFixture{}, fmt.Errorf("delta target_readback_digest must match target readback")
	}

	resumeRequired := !target.FinalResponseAllowed && target.ReadyNodes > 0
	status := "finalizable"
	if resumeRequired {
		status = "resumable"
	}
	fixture := AtlasMissionReadbackDiffFixture{
		Schema:                           AtlasMissionReadbackDiffFixtureContract,
		Status:                           status,
		MissionID:                        target.MissionID,
		TargetInstance:                   target.TargetInstance,
		SourceReadbackPath:               sourceRef,
		TargetReadbackPath:               targetRef,
		DeltaPath:                        deltaRef,
		SourceReadbackDigest:             digestValue(source),
		TargetReadbackDigest:             digestValue(target),
		DeltaDigest:                      digestValue(delta),
		CompletedNodeTransition:          numericTransition(source.CompletedNodes, target.CompletedNodes),
		ReadyNodeTransition:              numericTransition(source.ReadyNodes, target.ReadyNodes),
		CheckpointTransition:             numericTransition(source.CheckpointCount, target.CheckpointCount),
		FirstExecutableNodeBefore:        source.FirstExecutableNode,
		FirstExecutableNodeAfter:         target.FirstExecutableNode,
		ExactNextActionBefore:            source.ExactNextAction,
		ExactNextActionAfter:             target.ExactNextAction,
		ReturnGateStatusBefore:           source.ReturnGateStatus,
		ReturnGateStatusAfter:            target.ReturnGateStatus,
		ContinuationContractReasonBefore: source.ContinuationContract.Reason,
		ContinuationContractReasonAfter:  target.ContinuationContract.Reason,
		FinalResponseAllowedBefore:       source.FinalResponseAllowed,
		FinalResponseAllowedAfter:        target.FinalResponseAllowed,
		ResumeRequired:                   resumeRequired,
		ResumeReason:                     target.ContinuationContract.Reason,
		ExpectedNextAction:               target.ExactNextAction,
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
	if err := ValidateAtlasMissionReadbackDiffFixture(fixture); err != nil {
		return AtlasMissionReadbackDiffFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasMissionReadbackDiffFixture(fixture AtlasMissionReadbackDiffFixture) error {
	var errs []string
	requireContract(&errs, "mission_readback_diff_fixture", fixture.Schema, AtlasMissionReadbackDiffFixtureContract)
	if !oneOf(fixture.Status, "resumable", "finalizable") {
		errs = append(errs, "status must be resumable or finalizable")
	}
	requireField(&errs, "mission_id", fixture.MissionID)
	requireField(&errs, "target_instance", fixture.TargetInstance)
	for field, value := range map[string]string{
		"source_readback_path": fixture.SourceReadbackPath,
		"target_readback_path": fixture.TargetReadbackPath,
		"delta_path":           fixture.DeltaPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest": fixture.SourceReadbackDigest,
		"target_readback_digest": fixture.TargetReadbackDigest,
		"delta_digest":           fixture.DeltaDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	validateNumericTransition(&errs, "completed_node_transition", fixture.CompletedNodeTransition)
	validateNumericTransition(&errs, "ready_node_transition", fixture.ReadyNodeTransition)
	validateNumericTransition(&errs, "checkpoint_transition", fixture.CheckpointTransition)
	if fixture.CompletedNodeTransition.Delta != 1 {
		errs = append(errs, "completed_node_transition delta must be 1")
	}
	if fixture.ReadyNodeTransition.Delta != -1 {
		errs = append(errs, "ready_node_transition delta must be -1")
	}
	if fixture.CheckpointTransition.Delta != 1 {
		errs = append(errs, "checkpoint_transition delta must be 1")
	}
	requireField(&errs, "first_executable_node_before", fixture.FirstExecutableNodeBefore)
	requireField(&errs, "exact_next_action_before", fixture.ExactNextActionBefore)
	requireField(&errs, "exact_next_action_after", fixture.ExactNextActionAfter)
	requireField(&errs, "return_gate_status_before", fixture.ReturnGateStatusBefore)
	requireField(&errs, "return_gate_status_after", fixture.ReturnGateStatusAfter)
	requireField(&errs, "continuation_contract_reason_before", fixture.ContinuationContractReasonBefore)
	requireField(&errs, "continuation_contract_reason_after", fixture.ContinuationContractReasonAfter)
	requireField(&errs, "resume_reason", fixture.ResumeReason)
	requireField(&errs, "expected_next_action", fixture.ExpectedNextAction)
	for field, value := range map[string]string{
		"first_executable_node_before":        fixture.FirstExecutableNodeBefore,
		"first_executable_node_after":         fixture.FirstExecutableNodeAfter,
		"exact_next_action_before":            fixture.ExactNextActionBefore,
		"exact_next_action_after":             fixture.ExactNextActionAfter,
		"return_gate_status_before":           fixture.ReturnGateStatusBefore,
		"return_gate_status_after":            fixture.ReturnGateStatusAfter,
		"continuation_contract_reason_before": fixture.ContinuationContractReasonBefore,
		"continuation_contract_reason_after":  fixture.ContinuationContractReasonAfter,
		"resume_reason":                       fixture.ResumeReason,
		"expected_next_action":                fixture.ExpectedNextAction,
	} {
		checkPublicPath(&errs, field, value, true)
	}
	if fixture.Status == "resumable" {
		if !fixture.ResumeRequired {
			errs = append(errs, "resumable status requires resume_required=true")
		}
		if fixture.FinalResponseAllowedAfter {
			errs = append(errs, "resumable status requires final_response_allowed_after=false")
		}
		if fixture.FirstExecutableNodeAfter == "" {
			errs = append(errs, "resumable status requires first_executable_node_after")
		}
	} else if fixture.ResumeRequired {
		errs = append(errs, "finalizable status requires resume_required=false")
	}
	if fixture.ExpectedNextAction != fixture.ExactNextActionAfter {
		errs = append(errs, "expected_next_action must match exact_next_action_after")
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

func numericTransition(before, after int) AtlasMissionReadbackNumericTransition {
	return AtlasMissionReadbackNumericTransition{
		Before: before,
		After:  after,
		Delta:  after - before,
	}
}

func validateNumericTransition(errs *[]string, field string, transition AtlasMissionReadbackNumericTransition) {
	if transition.After-transition.Before != transition.Delta {
		*errs = append(*errs, field+" delta must equal after minus before")
	}
}
