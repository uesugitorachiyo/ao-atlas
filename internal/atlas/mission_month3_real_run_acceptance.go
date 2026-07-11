package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3RealRunAcceptanceCriteria(nodeID, readinessMatrixPath, nonAOReplayPath string) (AtlasMonth3RealRunAcceptanceCriteria, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3RealRunAcceptanceCriteria{}, fmt.Errorf("node id is required")
	}
	matrix, err := LoadJSON[AtlasGoldenPathReadinessMatrix](readinessMatrixPath)
	if err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	replay, err := LoadJSON[AtlasMonth3NonAODryRunReplayBinding](nonAOReplayPath)
	if err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	if err := ValidateAtlasMonth3NonAODryRunReplayBinding(replay); err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	matrixDigest, err := digestTextFileWithNormalizedLineEndings(readinessMatrixPath)
	if err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	replayDigest, err := digestTextFileWithNormalizedLineEndings(nonAOReplayPath)
	if err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	repos := []AtlasMonth3ExternalRepoCriteria{
		month3ExternalRepoCriteria("external-non-ao-cli-fixture"),
		month3ExternalRepoCriteria("external-non-ao-library-fixture"),
		month3ExternalRepoCriteria("external-non-ao-docs-fixture"),
	}
	criteria := AtlasMonth3RealRunAcceptanceCriteria{
		Schema:                AtlasMonth3RealRunAcceptanceCriteriaContract,
		NodeID:                nodeID,
		Status:                "real_run_acceptance_ready",
		ReadinessMatrixPath:   publicArtifactRef(readinessMatrixPath),
		ReadinessMatrixDigest: matrixDigest,
		NonAOReplayPath:       publicArtifactRef(nonAOReplayPath),
		NonAOReplayDigest:     replayDigest,
		ExternalRepoCount:     len(repos),
		CriteriaPerRepo:       7,
		ExternalRepos:         repos,
		AcceptanceSummary: []string{
			"three external non-AO repositories must run from isolated worktrees",
			"each accepted run requires exact digest approval before mutation",
			"each accepted run requires reviewed PR evidence and rollback receipt",
			"each accepted run requires observer readback and no-promotion verdict",
		},
		NonAOReplayBound:                 replay.Status == "non_ao_dry_run_replay_bound" && replay.FixtureOnlyExecutionEvidence,
		RequiresExplicitOperatorApproval: true,
		RequiresReviewedPR:               true,
		RequiresRollbackReceipt:          true,
		RequiresObserverReadback:         true,
		PromotionRequested:               matrix.PromotionRequested || replay.PromotionRequested,
		SchedulesWork:                    false,
		ExecutesWork:                     false,
		ApprovesWork:                     false,
		ClaimsAuthorityAdvance:           matrix.ClaimsAuthorityAdvance || replay.ClaimsAuthorityAdvance,
		RSIRemainsDenied:                 matrix.RSIRemainsDenied && replay.RSIRemainsDenied,
	}
	if !criteria.NonAOReplayBound || criteria.PromotionRequested || criteria.ClaimsAuthorityAdvance || !criteria.RSIRemainsDenied {
		criteria.Status = "real_run_acceptance_failed"
	}
	if err := ValidateAtlasMonth3RealRunAcceptanceCriteria(criteria); err != nil {
		return AtlasMonth3RealRunAcceptanceCriteria{}, err
	}
	return criteria, nil
}

func month3ExternalRepoCriteria(id string) AtlasMonth3ExternalRepoCriteria {
	return AtlasMonth3ExternalRepoCriteria{
		ID:                          id,
		RepoClass:                   "external_non_ao",
		AllowedChangeClasses:        []string{"documentation", "test_fixture", "single_file_low_risk_code"},
		RequiresIsolatedWorktree:    true,
		RequiresExactDigestApproval: true,
		RequiresReviewedPR:          true,
		RequiresRollbackReceipt:     true,
		RequiresObserverReadback:    true,
		RequiresNoPromotionVerdict:  true,
		ProviderExecutionAllowed:    false,
	}
}

func ValidateAtlasMonth3RealRunAcceptanceCriteria(criteria AtlasMonth3RealRunAcceptanceCriteria) error {
	var errs []string
	requireContract(&errs, "month3_real_run_acceptance_criteria", criteria.Schema, AtlasMonth3RealRunAcceptanceCriteriaContract)
	requireField(&errs, "node_id", criteria.NodeID)
	checkPublicPath(&errs, "node_id", criteria.NodeID, true)
	if !oneOf(criteria.Status, "real_run_acceptance_ready", "real_run_acceptance_failed") {
		errs = append(errs, "status must be real_run_acceptance_ready or real_run_acceptance_failed")
	}
	for field, value := range map[string]string{
		"readiness_matrix_path": criteria.ReadinessMatrixPath,
		"non_ao_replay_path":    criteria.NonAOReplayPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"readiness_matrix_digest": criteria.ReadinessMatrixDigest,
		"non_ao_replay_digest":    criteria.NonAOReplayDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if criteria.ExternalRepoCount != 3 || len(criteria.ExternalRepos) != 3 {
		errs = append(errs, "external_repo_count must equal three external repos")
	}
	if criteria.CriteriaPerRepo < 7 {
		errs = append(errs, "criteria_per_repo must be at least 7")
	}
	requireList(&errs, "acceptance_summary", criteria.AcceptanceSummary)
	if !criteria.NonAOReplayBound {
		errs = append(errs, "non_ao_replay_bound must be true")
	}
	if !criteria.RequiresExplicitOperatorApproval {
		errs = append(errs, "requires_explicit_operator_approval must be true")
	}
	if !criteria.RequiresReviewedPR {
		errs = append(errs, "requires_reviewed_pr must be true")
	}
	if !criteria.RequiresRollbackReceipt {
		errs = append(errs, "requires_rollback_receipt must be true")
	}
	if !criteria.RequiresObserverReadback {
		errs = append(errs, "requires_observer_readback must be true")
	}
	if criteria.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	seen := map[string]bool{}
	for i, repo := range criteria.ExternalRepos {
		prefix := fmt.Sprintf("external_repos[%d]", i)
		requireField(&errs, prefix+".id", repo.ID)
		if seen[repo.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		seen[repo.ID] = true
		if repo.RepoClass != "external_non_ao" {
			errs = append(errs, prefix+".repo_class must be external_non_ao")
		}
		requireList(&errs, prefix+".allowed_change_classes", repo.AllowedChangeClasses)
		if !repo.RequiresIsolatedWorktree {
			errs = append(errs, prefix+".requires_isolated_worktree must be true")
		}
		if !repo.RequiresExactDigestApproval {
			errs = append(errs, prefix+".requires_exact_digest_approval must be true")
		}
		if !repo.RequiresReviewedPR {
			errs = append(errs, prefix+".requires_reviewed_pr must be true")
		}
		if !repo.RequiresRollbackReceipt {
			errs = append(errs, prefix+".requires_rollback_receipt must be true")
		}
		if !repo.RequiresObserverReadback {
			errs = append(errs, prefix+".requires_observer_readback must be true")
		}
		if !repo.RequiresNoPromotionVerdict {
			errs = append(errs, prefix+".requires_no_promotion_verdict must be true")
		}
		if repo.ProviderExecutionAllowed {
			errs = append(errs, prefix+".provider_execution_allowed must be false")
		}
	}
	validateNoAuthorityEffects(&errs, criteria.SchedulesWork, criteria.ExecutesWork, criteria.ApprovesWork, criteria.ClaimsAuthorityAdvance, criteria.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3RealRunAcceptanceCriteria(path string, criteria AtlasMonth3RealRunAcceptanceCriteria) error {
	return WriteJSON(path, criteria)
}
