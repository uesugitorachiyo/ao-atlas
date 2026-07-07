package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasResumeDenialEvidence(readbackPath string) (AtlasResumeDenialEvidence, error) {
	readback, err := LoadJSON[AtlasRecommendationReadback](readbackPath)
	if err != nil {
		return AtlasResumeDenialEvidence{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasResumeDenialEvidence{}, err
	}
	if readback.FinalResponseAllowed {
		return AtlasResumeDenialEvidence{}, fmt.Errorf("readback already allows final response")
	}
	if readback.ReadyNodes <= 0 {
		return AtlasResumeDenialEvidence{}, fmt.Errorf("readback has no ready nodes to deny final response")
	}
	if !readback.ContinuationContract.RefusesFinalResponse {
		return AtlasResumeDenialEvidence{}, fmt.Errorf("readback continuation contract does not refuse final response")
	}
	if strings.TrimSpace(readback.FirstExecutableNode) == "" {
		return AtlasResumeDenialEvidence{}, fmt.Errorf("readback missing first executable node")
	}
	evidence := AtlasResumeDenialEvidence{
		Schema:                     AtlasResumeDenialEvidenceContract,
		Status:                     "denied_ready_work_remains",
		SourceReadbackPath:         publicArtifactRef(readbackPath),
		SourceReadbackDigest:       digestValue(readback),
		MissionID:                  readback.MissionID,
		TargetInstance:             readback.TargetInstance,
		CompletedNodes:             readback.CompletedNodes,
		TotalNodes:                 readback.TotalNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		CheckpointCount:            readback.CheckpointCount,
		CurrentNextExecutableNode:  readback.FirstExecutableNode,
		ExactNextAction:            readback.ExactNextAction,
		ReturnGateStatus:           readback.ReturnGateStatus,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		FinalResponseDenialGate:    readback.FinalResponseDenialGate,
		FinalResponseReason:        readback.FinalResponseReason,
		RefusesFinalResponse:       readback.ContinuationContract.RefusesFinalResponse,
		DenialAssertions: []string{
			"ready_work_remains",
			"exact_next_action_remains",
			"continuation_required",
			"final_response_denied_until_ready_work_consumed",
			"rsi_denial_preserved",
		},
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
	if err := ValidateAtlasResumeDenialEvidence(evidence); err != nil {
		return AtlasResumeDenialEvidence{}, err
	}
	return evidence, nil
}

func ValidateAtlasResumeDenialEvidence(evidence AtlasResumeDenialEvidence) error {
	var errs []string
	requireContract(&errs, "resume_denial_evidence", evidence.Schema, AtlasResumeDenialEvidenceContract)
	if evidence.Status != "denied_ready_work_remains" {
		errs = append(errs, "status must be denied_ready_work_remains")
	}
	requireField(&errs, "source_readback_path", evidence.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", evidence.SourceReadbackPath, true)
	if !digestPattern.MatchString(evidence.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	requireField(&errs, "mission_id", evidence.MissionID)
	requireField(&errs, "target_instance", evidence.TargetInstance)
	checkPublicPath(&errs, "mission_id", evidence.MissionID, true)
	checkPublicPath(&errs, "target_instance", evidence.TargetInstance, true)
	if evidence.CompletedNodes <= 0 || evidence.TotalNodes < evidence.CompletedNodes || evidence.ReadyNodes <= 0 || evidence.BlockedNodes < 0 || evidence.FailedNodes < 0 {
		errs = append(errs, "node counts must be positive and internally consistent with ready work remaining")
	}
	if evidence.CheckpointCount < evidence.CompletedNodes {
		errs = append(errs, "checkpoint_count must cover completed nodes")
	}
	requireField(&errs, "current_next_executable_node", evidence.CurrentNextExecutableNode)
	checkPublicPath(&errs, "current_next_executable_node", evidence.CurrentNextExecutableNode, true)
	requireField(&errs, "exact_next_action", evidence.ExactNextAction)
	checkPublicStrings(&errs, "exact_next_action", []string{evidence.ExactNextAction}, true)
	requireField(&errs, "return_gate_status", evidence.ReturnGateStatus)
	requireField(&errs, "continuation_contract_reason", evidence.ContinuationContractReason)
	requireField(&errs, "final_response_denial_gate", evidence.FinalResponseDenialGate)
	requireField(&errs, "final_response_reason", evidence.FinalResponseReason)
	checkPublicStrings(&errs, "final_response_reason", []string{evidence.FinalResponseReason}, true)
	if evidence.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	if !evidence.RefusesFinalResponse {
		errs = append(errs, "refuses_final_response must be true while ready work remains")
	}
	if evidence.ReturnGateStatus != "blocked_ready_nodes_remain" {
		errs = append(errs, "return_gate_status must be blocked_ready_nodes_remain")
	}
	if evidence.ContinuationContractReason != "ready_nodes_or_exact_next_action_remain" {
		errs = append(errs, "continuation_contract_reason must be ready_nodes_or_exact_next_action_remain")
	}
	if evidence.FinalResponseDenialGate != "deny_ready_nodes_or_exact_next_action_remain" {
		errs = append(errs, "final_response_denial_gate must be deny_ready_nodes_or_exact_next_action_remain")
	}
	for _, assertion := range []string{
		"ready_work_remains",
		"exact_next_action_remains",
		"continuation_required",
		"final_response_denied_until_ready_work_consumed",
		"rsi_denial_preserved",
	} {
		if !containsStringValue(evidence.DenialAssertions, assertion) {
			errs = append(errs, "denial_assertions missing "+assertion)
		}
	}
	for key, value := range evidence.SafetyBoundaries {
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
	if len(evidence.SafetyBoundaries) == 0 {
		errs = append(errs, "safety_boundaries must not be empty")
	}
	validateNoAuthorityEffects(&errs, evidence.SchedulesWork, evidence.ExecutesWork, evidence.ApprovesWork, evidence.ClaimsAuthorityAdvance, evidence.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasResumeDenialEvidence(path string, evidence AtlasResumeDenialEvidence) error {
	return WriteJSON(path, evidence)
}
