package atlas

import (
	"fmt"
	"strings"
)

func ValidateAtlasMonth6GoldenPathDryRunRehearsals(fixture AtlasMonth6GoldenPathDryRunRehearsals) error {
	var errs []string
	requireContract(&errs, "month6_golden_path_dry_run_rehearsals", fixture.Schema, AtlasMonth6GoldenPathDryRunRehearsalsContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "dry_run_rehearsals_bound" {
		errs = append(errs, "status must be dry_run_rehearsals_bound")
	}
	if fixture.SourceRecommendationRank != 3 {
		errs = append(errs, "source_recommendation_rank must be 3")
	}
	if strings.TrimSpace(fixture.SourceRecommendationTask) != "Run three dry-run golden path rehearsals on non-AO sample repos" {
		errs = append(errs, "source_recommendation_task must match Month 6 recommendation 3")
	}
	if fixture.SafetyGate != "planning_only_no_provider_no_release" {
		errs = append(errs, "safety_gate must be planning_only_no_provider_no_release")
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
	if !fixture.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if fixture.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false for fixture-only rehearsals")
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
		validateAtlasMonth6GoldenPathDryRunRehearsal(&errs, fmt.Sprintf("rehearsals[%d]", i), rehearsal, seen)
	}
	return joinErrors(errs)
}

func validateAtlasMonth6GoldenPathDryRunRehearsal(errs *[]string, prefix string, rehearsal AtlasMonth6GoldenPathDryRunRehearsal, seen map[string]bool) {
	requireField(errs, prefix+".id", rehearsal.ID)
	checkPublicPath(errs, prefix+".id", rehearsal.ID, true)
	if seen[rehearsal.ID] {
		*errs = append(*errs, prefix+".id must be unique")
	}
	seen[rehearsal.ID] = true
	requireField(errs, prefix+".repository", rehearsal.Repository)
	checkPublicPath(errs, prefix+".repository", rehearsal.Repository, true)
	if rehearsal.RepoClass != "external_non_ao" {
		*errs = append(*errs, prefix+".repo_class must be external_non_ao")
	}
	for field, value := range map[string]string{
		".objective_digest":        rehearsal.ObjectiveDigest,
		".diff_digest_placeholder": rehearsal.DiffDigestPlaceholder,
	} {
		if !digestPattern.MatchString(value) {
			*errs = append(*errs, prefix+field+" must be sha256 digest")
		}
	}
	if !lowerHex40(rehearsal.BaseCommit) {
		*errs = append(*errs, prefix+".base_commit must be 40 lowercase hex characters")
	}
	for field, value := range map[string]string{
		".rollback_receipt_ref":       rehearsal.RollbackReceiptRef,
		".command_readback_ref":       rehearsal.CommandReadbackRef,
		".sentinel_public_safety_ref": rehearsal.SentinelPublicSafetyRef,
		".promoter_no_promotion_ref":  rehearsal.PromoterNoPromotionRef,
	} {
		requireField(errs, prefix+field, value)
		checkPublicPath(errs, prefix+field, value, true)
	}
	if !rehearsal.FixtureOnly {
		*errs = append(*errs, prefix+".fixture_only must be true")
	}
	if !rehearsal.DryRunOnly {
		*errs = append(*errs, prefix+".dry_run_only must be true")
	}
	if rehearsal.ProviderCallsAllowed {
		*errs = append(*errs, prefix+".provider_calls_allowed must be false")
	}
	if rehearsal.CredentialUseAllowed {
		*errs = append(*errs, prefix+".credential_use_allowed must be false")
	}
	if rehearsal.LiveMutationAllowed {
		*errs = append(*errs, prefix+".live_mutation_allowed must be false")
	}
	if rehearsal.ReleaseOrPublishAllowed {
		*errs = append(*errs, prefix+".release_or_publish_allowed must be false")
	}
	if rehearsal.ApprovalGranted {
		*errs = append(*errs, prefix+".approval_granted must be false")
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

func lowerHex40(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
