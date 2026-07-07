package atlas

import (
	"fmt"
	"strings"
)

type commandPromoterDisagreementDenialCaseSpec struct {
	id       string
	mutation string
	mutate   func(*AtlasCommandPromoterAgreementRollup)
}

func BuildAtlasCommandPromoterDisagreementDenial(nodeID, sourceAgreementPath string) (AtlasCommandPromoterDisagreementDenial, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceAgreementPath = strings.TrimSpace(sourceAgreementPath)
	if nodeID == "" {
		return AtlasCommandPromoterDisagreementDenial{}, fmt.Errorf("node id is required")
	}
	if sourceAgreementPath == "" {
		return AtlasCommandPromoterDisagreementDenial{}, fmt.Errorf("source agreement path is required")
	}
	source, err := LoadJSON[AtlasCommandPromoterAgreementRollup](sourceAgreementPath)
	if err != nil {
		return AtlasCommandPromoterDisagreementDenial{}, err
	}
	if err := ValidateAtlasCommandPromoterAgreementRollup(source); err != nil {
		return AtlasCommandPromoterDisagreementDenial{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(sourceAgreementPath)
	if err != nil {
		return AtlasCommandPromoterDisagreementDenial{}, err
	}

	specs := []commandPromoterDisagreementDenialCaseSpec{
		{
			id:       "command_no_promotion_disagrees",
			mutation: "set Command no-promotion agreement false while Promoter remains no-promotion",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.CommandStatus = "readback_disagrees_with_promoter"
				rollup.CommandAgreesNoPromotion = false
			},
		},
		{
			id:       "promoter_invariant_disagrees",
			mutation: "set Promoter no-promotion invariant false while Command still reports no-promotion",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.PromoterNoPromotionInvariantHolds = false
				rollup.AggregatePromotionStatus = "promotion_blocked"
			},
		},
		{
			id:       "promoter_request_conflicts",
			mutation: "set Promoter promotion requested true while Command still reports no-promotion",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.PromotionRequested = true
				rollup.AggregatePromotionStatus = "promotion_requested"
			},
		},
		{
			id:       "readback_command_disagrees",
			mutation: "set Command expected readback agreement false while Promoter remains no-promotion",
			mutate: func(rollup *AtlasCommandPromoterAgreementRollup) {
				rollup.ReadbackAgreesWithCommand = false
			},
		},
	}

	evidence := AtlasCommandPromoterDisagreementDenial{
		Schema:                         AtlasCommandPromoterDisagreementDenialContract,
		NodeID:                         nodeID,
		Status:                         "final_response_denied_command_promoter_disagreement",
		SourceAgreementPath:            publicArtifactRef(sourceAgreementPath),
		SourceAgreementDigest:          sourceDigest,
		SourceCommandStatus:            source.CommandStatus,
		SourceAggregatePromotionStatus: source.AggregatePromotionStatus,
		SourceFinalResponseAllowed:     source.FinalResponseAllowed,
		CaseCount:                      len(specs),
		Cases:                          make([]AtlasCommandPromoterDisagreementDenialCase, 0, len(specs)),
		FinalResponseAllowed:           false,
		FinalResponseDenialGate:        "deny_command_promoter_disagreement",
		FinalResponseReason:            "command_promoter_disagreement_requires_continued_supervision",
		PromotionRequested:             false,
		PromotionGranted:               false,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		ClaimsAuthorityAdvance:         false,
		RSIRemainsDenied:               true,
	}
	for _, spec := range specs {
		mutated := source
		spec.mutate(&mutated)
		disagreement := commandPromoterDisagreementDetected(mutated)
		finalResponseAllowed := false
		finalResponseDenied := disagreement && !finalResponseAllowed
		if finalResponseDenied {
			evidence.DeniedCases++
		}
		if disagreement {
			evidence.CommandPromoterDisagreementDetected = true
		}
		evidence.Cases = append(evidence.Cases, AtlasCommandPromoterDisagreementDenialCase{
			ID:                                spec.id,
			Mutation:                          spec.mutation,
			CommandStatus:                     mutated.CommandStatus,
			AggregatePromotionStatus:          mutated.AggregatePromotionStatus,
			CommandAgreesNoPromotion:          mutated.CommandAgreesNoPromotion,
			PromoterNoPromotionInvariantHolds: mutated.PromoterNoPromotionInvariantHolds,
			ReadbackAgreesWithCommand:         mutated.ReadbackAgreesWithCommand,
			PromotionRequested:                mutated.PromotionRequested,
			PromotionGranted:                  mutated.PromotionGranted,
			ClaimsAuthorityAdvance:            mutated.ClaimsAuthorityAdvance,
			RSIRemainsDenied:                  mutated.RSIRemainsDenied,
			DisagreementDetected:              disagreement,
			FinalResponseAllowed:              finalResponseAllowed,
			FinalResponseDenied:               finalResponseDenied,
			DenialReason:                      "command_promoter_disagreement_requires_continued_supervision",
		})
	}
	if evidence.DeniedCases != evidence.CaseCount || !evidence.CommandPromoterDisagreementDetected {
		evidence.Status = "command_promoter_disagreement_denial_failed"
	}
	if err := ValidateAtlasCommandPromoterDisagreementDenial(evidence); err != nil {
		return AtlasCommandPromoterDisagreementDenial{}, err
	}
	return evidence, nil
}

func commandPromoterDisagreementDetected(rollup AtlasCommandPromoterAgreementRollup) bool {
	return rollup.CommandStatus != "readback_agrees_no_promotion" ||
		!rollup.CommandAgreesNoPromotion ||
		!rollup.PromoterNoPromotionInvariantHolds ||
		!rollup.ReadbackAgreesWithCommand ||
		rollup.PromotionRequested ||
		rollup.PromotionGranted ||
		rollup.ClaimsAuthorityAdvance ||
		!rollup.RSIRemainsDenied
}

func ValidateAtlasCommandPromoterDisagreementDenial(evidence AtlasCommandPromoterDisagreementDenial) error {
	var errs []string
	requireContract(&errs, "command_promoter_disagreement_denial", evidence.Schema, AtlasCommandPromoterDisagreementDenialContract)
	requireField(&errs, "node_id", evidence.NodeID)
	checkPublicPath(&errs, "node_id", evidence.NodeID, true)
	if !oneOf(evidence.Status, "final_response_denied_command_promoter_disagreement", "command_promoter_disagreement_denial_failed") {
		errs = append(errs, "status must be final_response_denied_command_promoter_disagreement or command_promoter_disagreement_denial_failed")
	}
	requireField(&errs, "source_agreement_path", evidence.SourceAgreementPath)
	checkPublicPath(&errs, "source_agreement_path", evidence.SourceAgreementPath, true)
	if !digestPattern.MatchString(evidence.SourceAgreementDigest) {
		errs = append(errs, "source_agreement_digest must be sha256 digest")
	}
	if evidence.SourceCommandStatus != "readback_agrees_no_promotion" {
		errs = append(errs, "source_command_status must be readback_agrees_no_promotion")
	}
	if evidence.SourceAggregatePromotionStatus != "no_promotion_requested" {
		errs = append(errs, "source_aggregate_promotion_status must be no_promotion_requested")
	}
	if evidence.SourceFinalResponseAllowed {
		errs = append(errs, "source_final_response_allowed must be false")
	}
	if evidence.CaseCount != 4 || len(evidence.Cases) != evidence.CaseCount {
		errs = append(errs, "case_count must be 4 and match cases length")
	}
	if evidence.DeniedCases != evidence.CaseCount {
		errs = append(errs, "denied_cases must match case_count")
	}
	expectedCaseIDs := map[string]bool{
		"command_no_promotion_disagrees": false,
		"promoter_invariant_disagrees":   false,
		"promoter_request_conflicts":     false,
		"readback_command_disagrees":     false,
	}
	for i, c := range evidence.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", c.ID)
		requireField(&errs, prefix+".mutation", c.Mutation)
		requireField(&errs, prefix+".command_status", c.CommandStatus)
		requireField(&errs, prefix+".aggregate_promotion_status", c.AggregatePromotionStatus)
		requireField(&errs, prefix+".denial_reason", c.DenialReason)
		if _, ok := expectedCaseIDs[c.ID]; !ok {
			errs = append(errs, prefix+".id is not an expected disagreement case")
		} else if expectedCaseIDs[c.ID] {
			errs = append(errs, prefix+".id is duplicated")
		} else {
			expectedCaseIDs[c.ID] = true
		}
		if !c.DisagreementDetected {
			errs = append(errs, prefix+".disagreement_detected must be true")
		}
		if c.FinalResponseAllowed {
			errs = append(errs, prefix+".final_response_allowed must be false")
		}
		if !c.FinalResponseDenied {
			errs = append(errs, prefix+".final_response_denied must be true")
		}
		if c.DenialReason != "command_promoter_disagreement_requires_continued_supervision" {
			errs = append(errs, prefix+".denial_reason must be command_promoter_disagreement_requires_continued_supervision")
		}
		if c.PromotionGranted {
			errs = append(errs, prefix+".promotion_granted must be false")
		}
		if c.ClaimsAuthorityAdvance {
			errs = append(errs, prefix+".claims_authority_advance must be false")
		}
		if !c.RSIRemainsDenied {
			errs = append(errs, prefix+".rsi_remains_denied must be true")
		}
	}
	for id, seen := range expectedCaseIDs {
		if !seen {
			errs = append(errs, "missing disagreement denial case "+id)
		}
	}
	if !evidence.CommandPromoterDisagreementDetected {
		errs = append(errs, "command_promoter_disagreement_detected must be true")
	}
	if evidence.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	if evidence.FinalResponseDenialGate != "deny_command_promoter_disagreement" {
		errs = append(errs, "final_response_denial_gate must be deny_command_promoter_disagreement")
	}
	if evidence.FinalResponseReason != "command_promoter_disagreement_requires_continued_supervision" {
		errs = append(errs, "final_response_reason must be command_promoter_disagreement_requires_continued_supervision")
	}
	if evidence.PromotionRequested {
		errs = append(errs, "promotion_requested must be false for the evidence artifact")
	}
	if evidence.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if evidence.Status == "final_response_denied_command_promoter_disagreement" && evidence.DeniedCases != evidence.CaseCount {
		errs = append(errs, "denial status requires all disagreement cases denied")
	}
	validateNoAuthorityEffects(&errs, evidence.SchedulesWork, evidence.ExecutesWork, evidence.ApprovesWork, evidence.ClaimsAuthorityAdvance, evidence.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasCommandPromoterDisagreementDenial(path string, evidence AtlasCommandPromoterDisagreementDenial) error {
	return WriteJSON(path, evidence)
}
