package atlas

import (
	"fmt"
	"sort"
)

func BuildAtlasMissionReadbackDelta(sourceReadbackPath, targetReadbackPath string) (AtlasMissionReadbackDelta, error) {
	source, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMissionReadbackDelta{}, err
	}
	if err := ValidateAtlasRecommendationReadback(source); err != nil {
		return AtlasMissionReadbackDelta{}, err
	}
	target, err := LoadJSON[AtlasRecommendationReadback](targetReadbackPath)
	if err != nil {
		return AtlasMissionReadbackDelta{}, err
	}
	if err := ValidateAtlasRecommendationReadback(target); err != nil {
		return AtlasMissionReadbackDelta{}, err
	}
	sourceDigest := digestValue(source)
	targetDigest := digestValue(target)

	changed := map[string]bool{}
	numericDeltas := map[string]int{}
	booleanTransitions := map[string]AtlasMissionReadbackBooleanTransition{}
	stringTransitions := map[string]AtlasMissionReadbackStringTransition{}

	addNumeric := func(field string, before, after int) {
		if before == after {
			return
		}
		numericDeltas[field] = after - before
		changed[field] = true
	}
	addBoolean := func(field string, before, after bool, alwaysRecord bool) {
		if alwaysRecord || before != after {
			booleanTransitions[field] = AtlasMissionReadbackBooleanTransition{Before: before, After: after}
		}
		if before != after {
			changed[field] = true
		}
	}
	addString := func(field, before, after string) {
		if before == after {
			return
		}
		stringTransitions[field] = AtlasMissionReadbackStringTransition{Before: before, After: after}
		changed[field] = true
	}

	addNumeric("total_nodes", source.TotalNodes, target.TotalNodes)
	addNumeric("minimum_nodes", source.MinimumNodes, target.MinimumNodes)
	addNumeric("completed_nodes", source.CompletedNodes, target.CompletedNodes)
	addNumeric("ready_nodes", source.ReadyNodes, target.ReadyNodes)
	addNumeric("blocked_nodes", source.BlockedNodes, target.BlockedNodes)
	addNumeric("failed_nodes", source.FailedNodes, target.FailedNodes)
	addNumeric("executable_ready_nodes", source.ExecutableReadyNodes, target.ExecutableReadyNodes)
	addNumeric("checkpoint_count", source.CheckpointCount, target.CheckpointCount)
	addNumeric("elapsed_minutes", source.ElapsedMinutes, target.ElapsedMinutes)

	addBoolean("final_response_allowed", source.FinalResponseAllowed, target.FinalResponseAllowed, true)
	addBoolean("min_minutes_met", source.MinMinutesMet, target.MinMinutesMet, true)
	addBoolean("schedules_work", source.SchedulesWork, target.SchedulesWork, true)
	addBoolean("executes_work", source.ExecutesWork, target.ExecutesWork, true)
	addBoolean("approves_work", source.ApprovesWork, target.ApprovesWork, true)

	addString("status", source.Status, target.Status)
	addString("first_executable_node", source.FirstExecutableNode, target.FirstExecutableNode)
	addString("return_gate_status", source.ReturnGateStatus, target.ReturnGateStatus)
	addString("final_response_denial_gate", source.FinalResponseDenialGate, target.FinalResponseDenialGate)
	addString("exact_next_action", source.ExactNextAction, target.ExactNextAction)
	addString("lease_health_status", source.LeaseHealthStatus, target.LeaseHealthStatus)
	addString("checkpoint_freshness_status", source.CheckpointFreshnessStatus, target.CheckpointFreshnessStatus)
	addString("early_return_risk_status", source.EarlyReturnRiskStatus, target.EarlyReturnRiskStatus)
	addString("foundry_rollup_status", source.FoundryRollupStatus, target.FoundryRollupStatus)
	addString("promoter_readback_status", source.PromoterReadbackStatus, target.PromoterReadbackStatus)
	addString("promoter_no_promotion_status", source.PromoterNoPromotionStatus, target.PromoterNoPromotionStatus)
	addString("command_readback_status", source.CommandReadbackStatus, target.CommandReadbackStatus)
	addString("command_timeline_status", source.CommandTimelineStatus, target.CommandTimelineStatus)
	addString("public_safety_scan_status", source.PublicSafetyScanStatus, target.PublicSafetyScanStatus)

	changedFields := make([]string, 0, len(changed))
	for field := range changed {
		changedFields = append(changedFields, field)
	}
	sort.Strings(changedFields)
	status := "unchanged"
	if len(changedFields) > 0 {
		status = "changed"
	}

	delta := AtlasMissionReadbackDelta{
		Schema:                  AtlasMissionReadbackDeltaContract,
		Status:                  status,
		SourceReadbackPath:      publicArtifactRef(sourceReadbackPath),
		TargetReadbackPath:      publicArtifactRef(targetReadbackPath),
		SourceReadbackDigest:    sourceDigest,
		TargetReadbackDigest:    targetDigest,
		DeterministicComparison: true,
		ChangedFields:           changedFields,
		NumericDeltas:           numericDeltas,
		BooleanTransitions:      booleanTransitions,
		StringTransitions:       stringTransitions,
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
	if err := ValidateAtlasMissionReadbackDelta(delta); err != nil {
		return AtlasMissionReadbackDelta{}, err
	}
	return delta, nil
}

func ValidateAtlasMissionReadbackDelta(delta AtlasMissionReadbackDelta) error {
	var errs []string
	requireContract(&errs, "mission_readback_delta", delta.Schema, AtlasMissionReadbackDeltaContract)
	if !oneOf(delta.Status, "changed", "unchanged") {
		errs = append(errs, "status must be changed or unchanged")
	}
	requireField(&errs, "source_readback_path", delta.SourceReadbackPath)
	requireField(&errs, "target_readback_path", delta.TargetReadbackPath)
	checkPublicPath(&errs, "source_readback_path", delta.SourceReadbackPath, true)
	checkPublicPath(&errs, "target_readback_path", delta.TargetReadbackPath, true)
	if !digestPattern.MatchString(delta.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(delta.TargetReadbackDigest) {
		errs = append(errs, "target_readback_digest must be sha256 digest")
	}
	if !delta.DeterministicComparison {
		errs = append(errs, "deterministic_comparison must be true")
	}
	if delta.Status == "changed" && len(delta.ChangedFields) == 0 {
		errs = append(errs, "changed status requires changed_fields")
	}
	if delta.Status == "unchanged" && len(delta.ChangedFields) != 0 {
		errs = append(errs, "unchanged status requires empty changed_fields")
	}
	if !sort.StringsAreSorted(delta.ChangedFields) {
		errs = append(errs, "changed_fields must be sorted")
	}
	seen := map[string]bool{}
	changedSet := map[string]bool{}
	for i, field := range delta.ChangedFields {
		requireField(&errs, fmt.Sprintf("changed_fields[%d]", i), field)
		checkPublicPath(&errs, fmt.Sprintf("changed_fields[%d]", i), field, true)
		if seen[field] {
			errs = append(errs, "changed_fields must not contain duplicates")
		}
		seen[field] = true
		changedSet[field] = true
	}
	for field, value := range delta.NumericDeltas {
		checkPublicPath(&errs, "numeric_deltas."+field, field, true)
		if value == 0 {
			errs = append(errs, "numeric_deltas."+field+" must not be zero")
		}
		if !changedSet[field] {
			errs = append(errs, "numeric_deltas."+field+" must be listed in changed_fields")
		}
	}
	for field, transition := range delta.BooleanTransitions {
		checkPublicPath(&errs, "boolean_transitions."+field, field, true)
		if transition.Before != transition.After && !changedSet[field] {
			errs = append(errs, "boolean_transitions."+field+" must be listed in changed_fields when values differ")
		}
	}
	for field, transition := range delta.StringTransitions {
		checkPublicPath(&errs, "string_transitions."+field, field, true)
		checkPublicPath(&errs, "string_transitions."+field+".before", transition.Before, true)
		checkPublicPath(&errs, "string_transitions."+field+".after", transition.After, true)
		if transition.Before == transition.After {
			errs = append(errs, "string_transitions."+field+" values must differ")
		}
		if !changedSet[field] {
			errs = append(errs, "string_transitions."+field+" must be listed in changed_fields")
		}
	}
	for _, field := range delta.ChangedFields {
		if _, ok := delta.NumericDeltas[field]; ok {
			continue
		}
		if transition, ok := delta.BooleanTransitions[field]; ok && transition.Before != transition.After {
			continue
		}
		if _, ok := delta.StringTransitions[field]; ok {
			continue
		}
		errs = append(errs, "changed field "+field+" must have a matching delta or transition")
	}
	if len(delta.SafetyBoundaries) == 0 {
		errs = append(errs, "safety_boundaries must not be empty")
	}
	for key, value := range delta.SafetyBoundaries {
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
	if delta.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if delta.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if delta.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if delta.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !delta.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}
