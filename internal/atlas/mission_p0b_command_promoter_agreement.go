package atlas

import "fmt"

type AtlasP0BCommandPromoterAgreement struct {
	Schema                          string                                  `json:"schema"`
	NodeID                          string                                  `json:"node_id"`
	Status                          string                                  `json:"status"`
	Source                          string                                  `json:"source"`
	CoveredNodeStart                int                                     `json:"covered_node_start"`
	CoveredNodeEnd                  int                                     `json:"covered_node_end"`
	EntryCount                      int                                     `json:"entry_count"`
	AllCommandReadbacksAgree        bool                                    `json:"all_command_readbacks_agree"`
	AllPromoterReadbacksNoPromotion bool                                    `json:"all_promoter_readbacks_no_promotion"`
	PromotionRequestedCount         int                                     `json:"promotion_requested_count"`
	PromotionGrantedCount           int                                     `json:"promotion_granted_count"`
	AuthorityAdvanceClaimCount      int                                     `json:"authority_advance_claim_count"`
	RSIDeniedCount                  int                                     `json:"rsi_denied_count"`
	CompletedNodesBefore            int                                     `json:"completed_nodes_before"`
	ReadyNodesBefore                int                                     `json:"ready_nodes_before"`
	FinalResponseAllowed            bool                                    `json:"final_response_allowed"`
	ExpectedNextNode                string                                  `json:"expected_next_node"`
	Entries                         []AtlasP0BCommandPromoterAgreementEntry `json:"entries"`
	PromotionRequested              bool                                    `json:"promotion_requested"`
	PromotionGranted                bool                                    `json:"promotion_granted"`
	ClaimsAuthorityAdvance          bool                                    `json:"claims_authority_advance"`
	RSIRemainsDenied                bool                                    `json:"rsi_remains_denied"`
	SchedulesWork                   bool                                    `json:"schedules_work"`
	ExecutesWork                    bool                                    `json:"executes_work"`
	ApprovesWork                    bool                                    `json:"approves_work"`
}

type AtlasP0BCommandPromoterAgreementEntry struct {
	NodeNumber               int    `json:"node_number"`
	NodeID                   string `json:"node_id"`
	CommandReadbackPath      string `json:"command_readback_path"`
	CommandStatus            string `json:"command_status"`
	PromoterReadbackPath     string `json:"promoter_readback_path"`
	PromoterStatus           string `json:"promoter_status"`
	CommandAgreesNoPromotion bool   `json:"command_agrees_no_promotion"`
	PromoterNoPromotion      bool   `json:"promoter_no_promotion"`
	PromotionRequested       bool   `json:"promotion_requested"`
	PromotionGranted         bool   `json:"promotion_granted"`
	ClaimsAuthorityAdvance   bool   `json:"claims_authority_advance"`
	RSIRemainsDenied         bool   `json:"rsi_remains_denied"`
}

func ValidateAtlasP0BCommandPromoterAgreement(agreement AtlasP0BCommandPromoterAgreement) error {
	var errs []string
	requireContract(&errs, "p0b_command_promoter_agreement", agreement.Schema, AtlasP0BCommandPromoterAgreementContract)
	if agreement.Status != "command_agrees_with_promoter_no_promotion" {
		errs = append(errs, "status must be command_agrees_with_promoter_no_promotion")
	}
	requireField(&errs, "node_id", agreement.NodeID)
	checkPublicPath(&errs, "node_id", agreement.NodeID, true)
	requireField(&errs, "source", agreement.Source)
	checkPublicPath(&errs, "source", agreement.Source, true)
	if agreement.CoveredNodeStart <= 0 || agreement.CoveredNodeEnd < agreement.CoveredNodeStart {
		errs = append(errs, "covered node range must be positive and ordered")
	}
	expectedCount := agreement.CoveredNodeEnd - agreement.CoveredNodeStart + 1
	if agreement.EntryCount != len(agreement.Entries) || agreement.EntryCount != expectedCount {
		errs = append(errs, "entry_count must match covered node range and entries length")
	}
	if agreement.PromotionRequestedCount < 0 || agreement.PromotionGrantedCount < 0 || agreement.AuthorityAdvanceClaimCount < 0 || agreement.RSIDeniedCount < 0 || agreement.CompletedNodesBefore < 0 || agreement.ReadyNodesBefore < 0 {
		errs = append(errs, "summary counts must be non-negative")
	}
	if agreement.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while P0-B work remains")
	}
	requireField(&errs, "expected_next_node", agreement.ExpectedNextNode)
	checkPublicPath(&errs, "expected_next_node", agreement.ExpectedNextNode, true)

	seenNodes := map[int]bool{}
	commandAgreementCount := 0
	promoterNoPromotionCount := 0
	promotionRequestedCount := 0
	promotionGrantedCount := 0
	authorityClaimCount := 0
	rsiDeniedCount := 0
	for i, entry := range agreement.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		validateP0BCommandPromoterAgreementEntry(&errs, prefix, entry)
		if entry.NodeNumber < agreement.CoveredNodeStart || entry.NodeNumber > agreement.CoveredNodeEnd {
			errs = append(errs, prefix+".node_number must be inside covered range")
		}
		if seenNodes[entry.NodeNumber] {
			errs = append(errs, prefix+".node_number must be unique")
		}
		seenNodes[entry.NodeNumber] = true
		if entry.CommandAgreesNoPromotion {
			commandAgreementCount++
		}
		if entry.PromoterNoPromotion {
			promoterNoPromotionCount++
		}
		if entry.PromotionRequested {
			promotionRequestedCount++
		}
		if entry.PromotionGranted {
			promotionGrantedCount++
		}
		if entry.ClaimsAuthorityAdvance {
			authorityClaimCount++
		}
		if entry.RSIRemainsDenied {
			rsiDeniedCount++
		}
	}
	if agreement.AllCommandReadbacksAgree != (commandAgreementCount == len(agreement.Entries) && len(agreement.Entries) > 0) {
		errs = append(errs, "all_command_readbacks_agree must match entries")
	}
	if agreement.AllPromoterReadbacksNoPromotion != (promoterNoPromotionCount == len(agreement.Entries) && len(agreement.Entries) > 0) {
		errs = append(errs, "all_promoter_readbacks_no_promotion must match entries")
	}
	if agreement.PromotionRequestedCount != promotionRequestedCount {
		errs = append(errs, "promotion_requested_count must match entries")
	}
	if agreement.PromotionGrantedCount != promotionGrantedCount {
		errs = append(errs, "promotion_granted_count must match entries")
	}
	if agreement.AuthorityAdvanceClaimCount != authorityClaimCount {
		errs = append(errs, "authority_advance_claim_count must match entries")
	}
	if agreement.RSIDeniedCount != rsiDeniedCount {
		errs = append(errs, "rsi_denied_count must match entries")
	}
	if agreement.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if agreement.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, agreement.SchedulesWork, agreement.ExecutesWork, agreement.ApprovesWork, agreement.ClaimsAuthorityAdvance, agreement.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateP0BCommandPromoterAgreementEntry(errs *[]string, prefix string, entry AtlasP0BCommandPromoterAgreementEntry) {
	requireField(errs, prefix+".node_id", entry.NodeID)
	checkPublicPath(errs, prefix+".node_id", entry.NodeID, true)
	requireField(errs, prefix+".command_readback_path", entry.CommandReadbackPath)
	checkPublicPath(errs, prefix+".command_readback_path", entry.CommandReadbackPath, true)
	requireField(errs, prefix+".promoter_readback_path", entry.PromoterReadbackPath)
	checkPublicPath(errs, prefix+".promoter_readback_path", entry.PromoterReadbackPath, true)
	if entry.CommandStatus != "readback_agrees_no_promotion" {
		*errs = append(*errs, prefix+".command_status must be readback_agrees_no_promotion")
	}
	if entry.PromoterStatus != "no_promotion_requested" {
		*errs = append(*errs, prefix+".promoter_status must be no_promotion_requested")
	}
	if !entry.CommandAgreesNoPromotion {
		*errs = append(*errs, prefix+".command_agrees_no_promotion must be true")
	}
	if !entry.PromoterNoPromotion {
		*errs = append(*errs, prefix+".promoter_no_promotion must be true")
	}
	if entry.PromotionRequested {
		*errs = append(*errs, prefix+".promotion_requested must be false")
	}
	if entry.PromotionGranted {
		*errs = append(*errs, prefix+".promotion_granted must be false")
	}
	if entry.ClaimsAuthorityAdvance {
		*errs = append(*errs, prefix+".claims_authority_advance must be false")
	}
	if !entry.RSIRemainsDenied {
		*errs = append(*errs, prefix+".rsi_remains_denied must be true")
	}
}
