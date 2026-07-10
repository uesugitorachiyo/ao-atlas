package atlas

type AtlasP0CMissionFoundryHandoffCheck struct {
	Schema                               string   `json:"schema"`
	NodeID                               string   `json:"node_id"`
	Status                               string   `json:"status"`
	PromptPath                           string   `json:"prompt_path"`
	CriteriaPath                         string   `json:"criteria_path"`
	SourceReadback                       string   `json:"source_readback"`
	CompletedNodesBefore                 int      `json:"completed_nodes_before"`
	ReadyNodesBefore                     int      `json:"ready_nodes_before"`
	MinimumGeneratedNodes                int      `json:"minimum_generated_nodes"`
	MinimumCompletionBeforeFinalResponse int      `json:"minimum_completion_before_final_response"`
	RequiresMissionFoundryCompletePath   bool     `json:"requires_mission_foundry_complete_path"`
	RequiresSingleActiveNode             bool     `json:"requires_single_active_node"`
	ProviderCallsAllowed                 bool     `json:"provider_calls_allowed"`
	CredentialUseAllowed                 bool     `json:"credential_use_allowed"`
	AO2LiveMutationAllowed               bool     `json:"ao2_live_mutation_allowed"`
	RequiredOwners                       []string `json:"required_owners"`
	SafetyBoundaries                     []string `json:"safety_boundaries"`
	PromotionRequested                   bool     `json:"promotion_requested"`
	PromotionGranted                     bool     `json:"promotion_granted"`
	ClaimsAuthorityAdvance               bool     `json:"claims_authority_advance"`
	RSIRemainsDenied                     bool     `json:"rsi_remains_denied"`
	SchedulesWork                        bool     `json:"schedules_work"`
	ExecutesWork                         bool     `json:"executes_work"`
	ApprovesWork                         bool     `json:"approves_work"`
}

func ValidateAtlasP0CMissionFoundryHandoffCheck(check AtlasP0CMissionFoundryHandoffCheck) error {
	var errs []string
	requireContract(&errs, "p0c_mission_foundry_handoff_check", check.Schema, AtlasP0CMissionFoundryHandoffCheckContract)
	if check.Status != "handoff_ready" {
		errs = append(errs, "status must be handoff_ready")
	}
	requireField(&errs, "node_id", check.NodeID)
	checkPublicPath(&errs, "node_id", check.NodeID, true)
	requireField(&errs, "prompt_path", check.PromptPath)
	checkPublicPath(&errs, "prompt_path", check.PromptPath, true)
	requireField(&errs, "criteria_path", check.CriteriaPath)
	checkPublicPath(&errs, "criteria_path", check.CriteriaPath, true)
	requireField(&errs, "source_readback", check.SourceReadback)
	checkPublicPath(&errs, "source_readback", check.SourceReadback, true)
	if check.CompletedNodesBefore < 0 || check.ReadyNodesBefore < 0 {
		errs = append(errs, "node counts must be non-negative")
	}
	if check.MinimumGeneratedNodes < 30 {
		errs = append(errs, "minimum_generated_nodes must be at least 30")
	}
	if check.MinimumCompletionBeforeFinalResponse < 20 {
		errs = append(errs, "minimum_completion_before_final_response must be at least 20")
	}
	if !check.RequiresMissionFoundryCompletePath {
		errs = append(errs, "requires_mission_foundry_complete_path must be true")
	}
	if !check.RequiresSingleActiveNode {
		errs = append(errs, "requires_single_active_node must be true")
	}
	if check.ProviderCallsAllowed {
		errs = append(errs, "provider_calls_allowed must be false")
	}
	if check.CredentialUseAllowed {
		errs = append(errs, "credential_use_allowed must be false")
	}
	if check.AO2LiveMutationAllowed {
		errs = append(errs, "ao2_live_mutation_allowed must be false")
	}
	requireList(&errs, "required_owners", check.RequiredOwners)
	checkPublicStrings(&errs, "required_owners", check.RequiredOwners, true)
	for _, owner := range []string{"ao-mission", "ao-atlas", "ao-foundry", "ao-command", "ao-sentinel", "ao-promoter"} {
		if !containsValue(check.RequiredOwners, owner) {
			errs = append(errs, "required_owners must include "+owner)
		}
	}
	requireList(&errs, "safety_boundaries", check.SafetyBoundaries)
	checkPublicStrings(&errs, "safety_boundaries", check.SafetyBoundaries, true)
	if check.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if check.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, check.SchedulesWork, check.ExecutesWork, check.ApprovesWork, check.ClaimsAuthorityAdvance, check.RSIRemainsDenied)
	return joinErrors(errs)
}
