package atlas

import "fmt"

func ValidateAtlasNodeCommandReadbackEvidence(evidence AtlasNodeCommandReadbackEvidence) error {
	var errs []string
	requireContract(&errs, "command_readback", evidence.Schema, AtlasNodeCommandReadbackContract)
	requireNodeEvidenceIdentity(&errs, "command_readback", evidence.NodeID, evidence.TaskID)
	if evidence.Status != "readback_agrees_no_promotion" {
		errs = append(errs, "status must be readback_agrees_no_promotion")
	}
	if evidence.CompletedNodesBefore < 0 {
		errs = append(errs, "completed_nodes_before must not be negative")
	}
	if evidence.ReadyNodesBefore < 0 {
		errs = append(errs, "ready_nodes_before must not be negative")
	}
	if evidence.ExpectedCompletedNodesAfter != evidence.CompletedNodesBefore+1 {
		errs = append(errs, "expected_completed_nodes_after must equal completed_nodes_before plus one")
	}
	if evidence.ExpectedReadyNodesAfter != evidence.ReadyNodesBefore-1 {
		errs = append(errs, "expected_ready_nodes_after must equal ready_nodes_before minus one")
	}
	requireField(&errs, "expected_next_executable_node_after", evidence.ExpectedNextExecutableNodeAfter)
	checkPublicPath(&errs, "expected_next_executable_node_after", evidence.ExpectedNextExecutableNodeAfter, true)
	if evidence.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	if !evidence.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func ValidateAtlasNodePromoterNoPromotionEvidence(evidence AtlasNodePromoterNoPromotionEvidence) error {
	var errs []string
	requireContract(&errs, "promoter_no_promotion", evidence.Schema, AtlasNodePromoterNoPromotionContract)
	requireNodeEvidenceIdentity(&errs, "promoter_no_promotion", evidence.NodeID, evidence.TaskID)
	if evidence.Status != "no_promotion_requested" {
		errs = append(errs, "status must be no_promotion_requested")
	}
	if evidence.PromotionClaimed {
		errs = append(errs, "promotion_claimed must be false")
	}
	if evidence.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !evidence.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func ValidateAtlasNodeSentinelPublicSafetyEvidence(evidence AtlasNodeSentinelPublicSafetyEvidence) error {
	var errs []string
	requireContract(&errs, "sentinel_public_safety", evidence.Schema, AtlasNodeSentinelPublicSafetyContract)
	requireNodeEvidenceIdentity(&errs, "sentinel_public_safety", evidence.NodeID, evidence.TaskID)
	if evidence.Status != "passed" {
		errs = append(errs, "status must be passed")
	}
	requireList(&errs, "scan_scope", evidence.ScanScope)
	for i, scope := range evidence.ScanScope {
		requireField(&errs, fmt.Sprintf("scan_scope[%d]", i), scope)
		checkPublicPath(&errs, fmt.Sprintf("scan_scope[%d]", i), scope, true)
	}
	if evidence.UnsafePublicClaimDetected {
		errs = append(errs, "unsafe_public_claim_detected must be false")
	}
	if evidence.PromotionClaimDetected {
		errs = append(errs, "promotion_claim_detected must be false")
	}
	if evidence.RSIClaimDetected {
		errs = append(errs, "rsi_claim_detected must be false")
	}
	if !evidence.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func requireNodeEvidenceIdentity(errs *[]string, prefix, nodeID, taskID string) {
	requireField(errs, prefix+".node_id", nodeID)
	requireField(errs, prefix+".task_id", taskID)
	checkPublicPath(errs, prefix+".node_id", nodeID, true)
	checkPublicPath(errs, prefix+".task_id", taskID, true)
	if nodeID != "" && taskID != "" && taskID != nodeID+"-task" {
		*errs = append(*errs, prefix+".task_id must equal node_id plus -task")
	}
}
