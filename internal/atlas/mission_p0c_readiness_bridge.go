package atlas

import "fmt"

type AtlasP0CReadinessCriteria struct {
	Schema                               string                       `json:"schema"`
	NodeID                               string                       `json:"node_id"`
	Status                               string                       `json:"status"`
	SourceReadback                       string                       `json:"source_readback"`
	SourceWorkgraph                      string                       `json:"source_workgraph"`
	CompletedNodesBefore                 int                          `json:"completed_nodes_before"`
	TotalNodes                           int                          `json:"total_nodes"`
	ReadyNodesBefore                     int                          `json:"ready_nodes_before"`
	BlockedNodesBefore                   int                          `json:"blocked_nodes_before"`
	FailedNodesBefore                    int                          `json:"failed_nodes_before"`
	FinalResponseAllowed                 bool                         `json:"final_response_allowed"`
	NextExecutableNode                   string                       `json:"next_executable_node"`
	RequiresP0BTerminalReadback          bool                         `json:"requires_p0b_terminal_readback"`
	RequiresMissionToFoundryCompletePath bool                         `json:"requires_mission_to_foundry_complete_path"`
	RequiredCriterionCount               int                          `json:"required_criterion_count"`
	RequiredCriteria                     []AtlasP0CReadinessCriterion `json:"required_criteria"`
	TerminalEvidenceRefs                 []string                     `json:"terminal_evidence_refs"`
	SafetyBoundaries                     []string                     `json:"safety_boundaries"`
	PromotionRequested                   bool                         `json:"promotion_requested"`
	PromotionGranted                     bool                         `json:"promotion_granted"`
	ClaimsAuthorityAdvance               bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied                     bool                         `json:"rsi_remains_denied"`
	SchedulesWork                        bool                         `json:"schedules_work"`
	ExecutesWork                         bool                         `json:"executes_work"`
	ApprovesWork                         bool                         `json:"approves_work"`
}

type AtlasP0CReadinessCriterion struct {
	ID                 string `json:"id"`
	Owner              string `json:"owner"`
	EvidenceKind       string `json:"evidence_kind"`
	Requirement        string `json:"requirement"`
	RequiredStatus     string `json:"required_status"`
	BlocksP0CIfMissing bool   `json:"blocks_p0c_if_missing"`
}

func ValidateAtlasP0CReadinessCriteria(criteria AtlasP0CReadinessCriteria) error {
	var errs []string
	requireContract(&errs, "p0c_readiness_criteria", criteria.Schema, AtlasP0CReadinessCriteriaContract)
	if criteria.Status != "criteria_recorded" {
		errs = append(errs, "status must be criteria_recorded")
	}
	requireField(&errs, "node_id", criteria.NodeID)
	checkPublicPath(&errs, "node_id", criteria.NodeID, true)
	requireField(&errs, "source_readback", criteria.SourceReadback)
	checkPublicPath(&errs, "source_readback", criteria.SourceReadback, true)
	requireField(&errs, "source_workgraph", criteria.SourceWorkgraph)
	checkPublicPath(&errs, "source_workgraph", criteria.SourceWorkgraph, true)
	if criteria.CompletedNodesBefore < 0 || criteria.TotalNodes <= 0 || criteria.ReadyNodesBefore < 0 || criteria.BlockedNodesBefore < 0 || criteria.FailedNodesBefore < 0 {
		errs = append(errs, "node counts must be non-negative and total_nodes must be positive")
	}
	if criteria.CompletedNodesBefore+criteria.ReadyNodesBefore+criteria.BlockedNodesBefore+criteria.FailedNodesBefore != criteria.TotalNodes {
		errs = append(errs, "node counts must add up to total_nodes")
	}
	if criteria.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while P0-B terminal handoff work remains")
	}
	requireField(&errs, "next_executable_node", criteria.NextExecutableNode)
	checkPublicPath(&errs, "next_executable_node", criteria.NextExecutableNode, true)
	if !criteria.RequiresP0BTerminalReadback {
		errs = append(errs, "requires_p0b_terminal_readback must be true")
	}
	if !criteria.RequiresMissionToFoundryCompletePath {
		errs = append(errs, "requires_mission_to_foundry_complete_path must be true")
	}
	if criteria.RequiredCriterionCount != len(criteria.RequiredCriteria) || criteria.RequiredCriterionCount < 10 {
		errs = append(errs, "required_criterion_count must match at least ten criteria")
	}
	requireList(&errs, "terminal_evidence_refs", criteria.TerminalEvidenceRefs)
	checkPublicStrings(&errs, "terminal_evidence_refs", criteria.TerminalEvidenceRefs, true)
	requireList(&errs, "safety_boundaries", criteria.SafetyBoundaries)
	checkPublicStrings(&errs, "safety_boundaries", criteria.SafetyBoundaries, true)

	seen := map[string]bool{}
	for i, item := range criteria.RequiredCriteria {
		prefix := fmt.Sprintf("required_criteria[%d]", i)
		validateP0CReadinessCriterion(&errs, prefix, item)
		if seen[item.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		seen[item.ID] = true
	}
	if criteria.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if criteria.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, criteria.SchedulesWork, criteria.ExecutesWork, criteria.ApprovesWork, criteria.ClaimsAuthorityAdvance, criteria.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateP0CReadinessCriterion(errs *[]string, prefix string, item AtlasP0CReadinessCriterion) {
	requireField(errs, prefix+".id", item.ID)
	checkPublicPath(errs, prefix+".id", item.ID, true)
	requireField(errs, prefix+".owner", item.Owner)
	checkPublicPath(errs, prefix+".owner", item.Owner, true)
	requireField(errs, prefix+".evidence_kind", item.EvidenceKind)
	checkPublicPath(errs, prefix+".evidence_kind", item.EvidenceKind, true)
	requireField(errs, prefix+".requirement", item.Requirement)
	checkPublicPath(errs, prefix+".requirement", item.Requirement, true)
	requireField(errs, prefix+".required_status", item.RequiredStatus)
	checkPublicPath(errs, prefix+".required_status", item.RequiredStatus, true)
	if !item.BlocksP0CIfMissing {
		*errs = append(*errs, prefix+".blocks_p0c_if_missing must be true")
	}
}
