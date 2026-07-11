package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3NonAODryRunReplayBinding(nodeID, sourceFixturePath, terminalBindingPath string) (AtlasMonth3NonAODryRunReplayBinding, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3NonAODryRunReplayBinding{}, fmt.Errorf("node id is required")
	}
	sourceFixture, err := LoadJSON[AtlasNonAOReplayBindingFixture](sourceFixturePath)
	if err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	if err := ValidateAtlasNonAOReplayBindingFixture(sourceFixture); err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	terminalBinding, err := LoadJSON[AtlasMonth3TerminalDigestBinding](terminalBindingPath)
	if err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	if err := ValidateAtlasMonth3TerminalDigestBinding(terminalBinding); err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	sourceFixtureDigest, err := digestTextFileWithNormalizedLineEndings(sourceFixturePath)
	if err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	terminalBindingDigest, err := digestTextFileWithNormalizedLineEndings(terminalBindingPath)
	if err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	binding := AtlasMonth3NonAODryRunReplayBinding{
		Schema:                       AtlasMonth3NonAODryRunReplayBindingContract,
		NodeID:                       nodeID,
		Status:                       "non_ao_dry_run_replay_bound",
		SourceFixturePath:            publicArtifactRef(sourceFixturePath),
		SourceFixtureDigest:          sourceFixtureDigest,
		TerminalBindingPath:          publicArtifactRef(terminalBindingPath),
		TerminalBindingDigest:        terminalBindingDigest,
		ReplayRepo:                   sourceFixture.ReplayRepo,
		TinyNonAORepo:                sourceFixture.TinyNonAORepo,
		ReviewedPREvidence:           sourceFixture.ReviewedPREvidence,
		ObserverReadbackBound:        sourceFixture.ObserverReadbackBound,
		NoPromotionBoundary:          sourceFixture.NoPromotionBoundary,
		FixtureOnlyExecutionEvidence: true,
		TerminalDigestBindingBound:   terminalBinding.Status == "terminal_digest_binding_ready" && terminalBinding.NodeCountsMatch,
		TerminalFinalResponseAllowed: terminalBinding.FinalResponseAllowed,
		PromotionRequested:           sourceFixture.PromotionRequested || terminalBinding.PromotionRequested,
		LiveProviderCalls:            sourceFixture.LiveProviderCalls,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       sourceFixture.ClaimsAuthorityAdvance || terminalBinding.ClaimsAuthorityAdvance,
		RSIRemainsDenied:             sourceFixture.RSIRemainsDenied && terminalBinding.RSIRemainsDenied,
	}
	if !binding.TerminalDigestBindingBound ||
		!binding.TerminalFinalResponseAllowed ||
		binding.PromotionRequested ||
		binding.LiveProviderCalls ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied {
		binding.Status = "non_ao_dry_run_replay_failed"
	}
	if err := ValidateAtlasMonth3NonAODryRunReplayBinding(binding); err != nil {
		return AtlasMonth3NonAODryRunReplayBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasMonth3NonAODryRunReplayBinding(binding AtlasMonth3NonAODryRunReplayBinding) error {
	var errs []string
	requireContract(&errs, "month3_non_ao_dry_run_replay", binding.Schema, AtlasMonth3NonAODryRunReplayBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if !oneOf(binding.Status, "non_ao_dry_run_replay_bound", "non_ao_dry_run_replay_failed") {
		errs = append(errs, "status must be non_ao_dry_run_replay_bound or non_ao_dry_run_replay_failed")
	}
	for field, value := range map[string]string{
		"source_fixture_path":   binding.SourceFixturePath,
		"terminal_binding_path": binding.TerminalBindingPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_fixture_digest":   binding.SourceFixtureDigest,
		"terminal_binding_digest": binding.TerminalBindingDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	requireField(&errs, "replay_repo", binding.ReplayRepo)
	if !binding.TinyNonAORepo {
		errs = append(errs, "tiny_non_ao_repo must be true")
	}
	if !binding.ReviewedPREvidence {
		errs = append(errs, "reviewed_pr_evidence must be true")
	}
	if !binding.ObserverReadbackBound {
		errs = append(errs, "observer_readback_bound must be true")
	}
	if !binding.NoPromotionBoundary {
		errs = append(errs, "no_promotion_boundary must be true")
	}
	if !binding.FixtureOnlyExecutionEvidence {
		errs = append(errs, "fixture_only_execution_evidence must be true")
	}
	if !binding.TerminalDigestBindingBound {
		errs = append(errs, "terminal_digest_binding_bound must be true")
	}
	if !binding.TerminalFinalResponseAllowed {
		errs = append(errs, "terminal_final_response_allowed must be true")
	}
	if binding.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if binding.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3NonAODryRunReplayBinding(path string, binding AtlasMonth3NonAODryRunReplayBinding) error {
	return WriteJSON(path, binding)
}
