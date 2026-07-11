package atlas

func BuildAtlasNonAOReplayBindingFixture() (AtlasNonAOReplayBindingFixture, error) {
	fixture := AtlasNonAOReplayBindingFixture{
		Schema:                 AtlasNonAOReplayBindingFixtureContract,
		Status:                 "non_ao_replay_binding_ready",
		ReplayRepo:             "tiny-non-ao-fixture-repo",
		TinyNonAORepo:          true,
		ReviewedPREvidence:    true,
		ObserverReadbackBound: true,
		NoPromotionBoundary:   true,
		PromotionRequested:    false,
		LiveProviderCalls:      false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasNonAOReplayBindingFixture(fixture); err != nil {
		return AtlasNonAOReplayBindingFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasNonAOReplayBindingFixture(fixture AtlasNonAOReplayBindingFixture) error {
	var errs []string
	requireContract(&errs, "non_ao_replay_binding_fixture", fixture.Schema, AtlasNonAOReplayBindingFixtureContract)
	if fixture.Status != "non_ao_replay_binding_ready" {
		errs = append(errs, "status must be non_ao_replay_binding_ready")
	}
	requireField(&errs, "replay_repo", fixture.ReplayRepo)
	if !fixture.TinyNonAORepo {
		errs = append(errs, "tiny_non_ao_repo must be true")
	}
	if !fixture.ReviewedPREvidence {
		errs = append(errs, "reviewed_pr_evidence must be true")
	}
	if !fixture.ObserverReadbackBound {
		errs = append(errs, "observer_readback_bound must be true")
	}
	if !fixture.NoPromotionBoundary {
		errs = append(errs, "no_promotion_boundary must be true")
	}
	if fixture.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if fixture.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
