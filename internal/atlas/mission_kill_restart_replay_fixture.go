package atlas

func BuildAtlasKillRestartReplayFixture() (AtlasKillRestartReplayFixture, error) {
	fixture := AtlasKillRestartReplayFixture{
		Schema:                    AtlasKillRestartReplayFixtureContract,
		Status:                    "kill_restart_replay_ready",
		KilledRunReplayed:         true,
		RestartReadbackBound:      true,
		NoLostEvidence:            true,
		DuplicateMutationDetected: false,
		FalseCompletionDetected:   false,
		LiveProviderCalls:         false,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	if err := ValidateAtlasKillRestartReplayFixture(fixture); err != nil {
		return AtlasKillRestartReplayFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasKillRestartReplayFixture(fixture AtlasKillRestartReplayFixture) error {
	var errs []string
	requireContract(&errs, "kill_restart_replay_fixture", fixture.Schema, AtlasKillRestartReplayFixtureContract)
	if fixture.Status != "kill_restart_replay_ready" {
		errs = append(errs, "status must be kill_restart_replay_ready")
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
	if fixture.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
