package atlas

import "fmt"

func BuildAtlasRepeatedTaskResultLedgerFixture() (AtlasRepeatedTaskResultLedgerFixture, error) {
	attempts := []AtlasRepeatedTaskResultLedgerAttempt{
		{Name: "task_attempt_1", ResultStatus: "passed", Replayable: true, ProviderCall: false},
		{Name: "task_attempt_2", ResultStatus: "passed", Replayable: true, ProviderCall: false},
		{Name: "task_attempt_3", ResultStatus: "passed", Replayable: true, ProviderCall: false},
	}
	fixture := AtlasRepeatedTaskResultLedgerFixture{
		Schema:                 AtlasRepeatedTaskResultLedgerFixtureContract,
		Status:                 "repeated_task_result_ledger_ready",
		Attempts:               attempts,
		AttemptCount:           len(attempts),
		DeterministicHarness:   true,
		ReplayableResultLedger: true,
		LiveProviderCalls:      false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasRepeatedTaskResultLedgerFixture(fixture); err != nil {
		return AtlasRepeatedTaskResultLedgerFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasRepeatedTaskResultLedgerFixture(fixture AtlasRepeatedTaskResultLedgerFixture) error {
	var errs []string
	requireContract(&errs, "repeated_task_result_ledger_fixture", fixture.Schema, AtlasRepeatedTaskResultLedgerFixtureContract)
	if fixture.Status != "repeated_task_result_ledger_ready" {
		errs = append(errs, "status must be repeated_task_result_ledger_ready")
	}
	if fixture.AttemptCount != len(fixture.Attempts) {
		errs = append(errs, "attempt_count must match attempts")
	}
	if fixture.AttemptCount != 3 {
		errs = append(errs, "attempt_count must be 3")
	}
	if !fixture.DeterministicHarness {
		errs = append(errs, "deterministic_harness must be true")
	}
	if !fixture.ReplayableResultLedger {
		errs = append(errs, "replayable_result_ledger must be true")
	}
	if fixture.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	for i, attempt := range fixture.Attempts {
		prefix := fmt.Sprintf("attempts[%d]", i)
		requireField(&errs, prefix+".name", attempt.Name)
		requireField(&errs, prefix+".result_status", attempt.ResultStatus)
		if !attempt.Replayable {
			errs = append(errs, prefix+".replayable must be true")
		}
		if attempt.ProviderCall {
			errs = append(errs, prefix+".provider_call must be false")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
