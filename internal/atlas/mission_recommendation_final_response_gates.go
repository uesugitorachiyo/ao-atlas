package atlas

import "fmt"

func DefaultAtlasRecommendationFinalResponseGates() (AtlasRecommendationFinalResponseGates, error) {
	gates := AtlasRecommendationFinalResponseGates{
		Schema:                               AtlasRecommendationFinalResponseGatesContract,
		Status:                               "ready",
		FinalResponseAllowedRequiresAllGates: true,
		Gates: []AtlasRecommendationFinalResponseGateEntry{
			{Gate: "completed_nodes_equal_total", Required: true, SourceField: "completed_nodes,total_nodes", Expected: "completed_nodes == total_nodes"},
			{Gate: "ready_nodes_zero", Required: true, SourceField: "ready_nodes", Expected: "0"},
			{Gate: "blocked_nodes_zero", Required: true, SourceField: "blocked_nodes", Expected: "0"},
			{Gate: "failed_nodes_zero", Required: true, SourceField: "failed_nodes", Expected: "0"},
			{Gate: "return_gate_final_response_allowed", Required: true, SourceField: "final_response_allowed", Expected: "true"},
			{Gate: "local_verification_passed", Required: true, SourceField: "verification.status", Expected: "passed"},
			{Gate: "public_safety_scan_passed", Required: true, SourceField: "public_safety_scan_status", Expected: "passed"},
			{Gate: "promoter_no_promotion", Required: true, SourceField: "promoter_readback_status", Expected: "no_promotion_requested"},
			{Gate: "command_readback_agrees", Required: true, SourceField: "command_readback_status", Expected: "agrees"},
			{Gate: "rsi_denied", Required: true, SourceField: "rsi_remains_denied", Expected: "true"},
		},
		NoPromotionRequested:   true,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
	}
	if err := ValidateAtlasRecommendationFinalResponseGates(gates); err != nil {
		return AtlasRecommendationFinalResponseGates{}, err
	}
	return gates, nil
}

func ValidateAtlasRecommendationFinalResponseGates(gates AtlasRecommendationFinalResponseGates) error {
	var errs []string
	requireContract(&errs, "recommendation_final_response_gates", gates.Schema, AtlasRecommendationFinalResponseGatesContract)
	if gates.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if !gates.FinalResponseAllowedRequiresAllGates {
		errs = append(errs, "final_response_allowed_requires_all_gates must be true")
	}
	requiredOrder := []string{
		"completed_nodes_equal_total",
		"ready_nodes_zero",
		"blocked_nodes_zero",
		"failed_nodes_zero",
		"return_gate_final_response_allowed",
		"local_verification_passed",
		"public_safety_scan_passed",
		"promoter_no_promotion",
		"command_readback_agrees",
		"rsi_denied",
	}
	if len(gates.Gates) != len(requiredOrder) {
		errs = append(errs, fmt.Sprintf("gates must include %d entries", len(requiredOrder)))
	}
	for index, gate := range gates.Gates {
		if index < len(requiredOrder) && gate.Gate != requiredOrder[index] {
			errs = append(errs, fmt.Sprintf("gate %d must be %s", index, requiredOrder[index]))
		}
		if !gate.Required {
			errs = append(errs, fmt.Sprintf("gate %s required must be true", gate.Gate))
		}
		requireField(&errs, "gate.source_field", gate.SourceField)
		requireField(&errs, "gate.expected", gate.Expected)
	}
	if !gates.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if gates.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if gates.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !gates.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if gates.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if gates.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if gates.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if gates.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if gates.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
