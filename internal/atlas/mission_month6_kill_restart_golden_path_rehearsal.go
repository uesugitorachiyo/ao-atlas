package atlas

import (
	"fmt"
	"strings"
)

func ValidateAtlasMonth6KillRestartGoldenPathRehearsal(fixture AtlasMonth6KillRestartGoldenPathRehearsal) error {
	var errs []string
	requireContract(&errs, "month6_kill_restart_golden_path_rehearsal", fixture.Schema, AtlasMonth6KillRestartGoldenPathRehearsalContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "kill_restart_rehearsal_bound" {
		errs = append(errs, "status must be kill_restart_rehearsal_bound")
	}
	if fixture.SourceRecommendationRank != 15 {
		errs = append(errs, "source_recommendation_rank must be 15")
	}
	if strings.TrimSpace(fixture.SourceRecommendationTask) != "Run kill-restart golden path rehearsal without provider execution" {
		errs = append(errs, "source_recommendation_task must match Month 6 recommendation 15")
	}
	if fixture.SafetyGate != "planning_only_no_provider_no_release" {
		errs = append(errs, "safety_gate must be planning_only_no_provider_no_release")
	}
	requireField(&errs, "source_rehearsal_ref", fixture.SourceRehearsalRef)
	checkPublicPath(&errs, "source_rehearsal_ref", fixture.SourceRehearsalRef, true)
	if !strings.Contains(fixture.SourceRehearsalRef, "month6-recommendation-03-golden-path-dry-run-rehearsals") {
		errs = append(errs, "source_rehearsal_ref must bind Month 6 dry-run rehearsal source")
	}
	if fixture.InterruptionMarker != "after_foundry_import_before_ao2_execution" {
		errs = append(errs, "interruption_marker must stop before AO2 execution")
	}
	if fixture.RestartResumeMarker != "mission_event_index_rebuilt_before_next_node_selection" {
		errs = append(errs, "restart_resume_marker must rebuild Mission event index before next node selection")
	}
	if fixture.RehearsalCount != 3 || len(fixture.Rehearsals) != 3 {
		errs = append(errs, "rehearsal_count must equal three rehearsals")
	}
	if !fixture.FixtureOnly {
		errs = append(errs, "fixture_only must be true")
	}
	if !fixture.DryRunOnly {
		errs = append(errs, "dry_run_only must be true")
	}
	if !fixture.KilledRunReplayed {
		errs = append(errs, "killed_run_replayed must be true")
	}
	if !fixture.RestartReadbackBound {
		errs = append(errs, "restart_readback_bound must be true")
	}
	if !fixture.NoLostEvidence {
		errs = append(errs, "no_lost_evidence must be true")
	}
	if fixture.DuplicateMutationDetected {
		errs = append(errs, "duplicate_mutation_detected must be false")
	}
	if fixture.FalseCompletionDetected {
		errs = append(errs, "false_completion_detected must be false")
	}
	if !fixture.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if fixture.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if fixture.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if fixture.ProviderCallsAllowed {
		errs = append(errs, "provider_calls_allowed must be false")
	}
	if fixture.CredentialUseAllowed {
		errs = append(errs, "credential_use_allowed must be false")
	}
	if fixture.LiveMutationAllowed {
		errs = append(errs, "live_mutation_allowed must be false")
	}
	if fixture.ReleaseOrPublishAllowed {
		errs = append(errs, "release_or_publish_allowed must be false")
	}
	if fixture.ApprovalGranted {
		errs = append(errs, "approval_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)

	seen := map[string]bool{}
	for i, rehearsal := range fixture.Rehearsals {
		validateAtlasMonth6KillRestartGoldenPathRehearsalRow(&errs, fmt.Sprintf("rehearsals[%d]", i), rehearsal, seen)
	}
	return joinErrors(errs)
}

func validateAtlasMonth6KillRestartGoldenPathRehearsalRow(errs *[]string, prefix string, rehearsal AtlasMonth6KillRestartGoldenPathRehearsalRow, seen map[string]bool) {
	requireField(errs, prefix+".id", rehearsal.ID)
	checkPublicPath(errs, prefix+".id", rehearsal.ID, true)
	if seen[rehearsal.ID] {
		*errs = append(*errs, prefix+".id must be unique")
	}
	seen[rehearsal.ID] = true
	if !rehearsal.KilledAfterCheckpoint {
		*errs = append(*errs, prefix+".killed_after_checkpoint must be true")
	}
	for field, value := range map[string]string{
		".restart_readback_ref": rehearsal.RestartReadbackRef,
		".rollback_receipt_ref": rehearsal.RollbackReceiptRef,
		".command_readback_ref": rehearsal.CommandReadbackRef,
	} {
		requireField(errs, prefix+field, value)
		checkPublicPath(errs, prefix+field, value, true)
	}
	if !strings.Contains(rehearsal.RestartReadbackRef, "month6-recommendation-15-kill-restart-golden-path-rehearsal") {
		*errs = append(*errs, prefix+".restart_readback_ref must bind recommendation 15 restart readback")
	}
	if !strings.Contains(rehearsal.RollbackReceiptRef, "month6-recommendation-03-golden-path-dry-run-rehearsals") ||
		!strings.Contains(rehearsal.CommandReadbackRef, "month6-recommendation-03-golden-path-dry-run-rehearsals") {
		*errs = append(*errs, prefix+" rollback and command refs must bind recommendation 3 dry-run rehearsal evidence")
	}
	if !rehearsal.EventIndexRebuilt {
		*errs = append(*errs, prefix+".event_index_rebuilt must be true")
	}
	if !rehearsal.ResumeSelectedSameNextNode {
		*errs = append(*errs, prefix+".resume_selected_same_next_node must be true")
	}
	if !rehearsal.NoLostEvidence {
		*errs = append(*errs, prefix+".no_lost_evidence must be true")
	}
	if rehearsal.DuplicateMutationDetected {
		*errs = append(*errs, prefix+".duplicate_mutation_detected must be false")
	}
	if rehearsal.FalseCompletionDetected {
		*errs = append(*errs, prefix+".false_completion_detected must be false")
	}
	if rehearsal.ProviderCallsAllowed {
		*errs = append(*errs, prefix+".provider_calls_allowed must be false")
	}
	if rehearsal.SafeToExecute {
		*errs = append(*errs, prefix+".safe_to_execute must be false")
	}
	if rehearsal.ExecutesWork {
		*errs = append(*errs, prefix+".executes_work must be false")
	}
	if rehearsal.MutatesRepository {
		*errs = append(*errs, prefix+".mutates_repository must be false")
	}
}
