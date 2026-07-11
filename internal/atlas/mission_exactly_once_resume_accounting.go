package atlas

import "fmt"

func BuildAtlasExactlyOnceResumeAccountingFixture() (AtlasExactlyOnceResumeAccountingFixture, error) {
	scenarios := []AtlasExactlyOnceResumeAccountingScenario{
		{Name: "restart", ExpectedCompletedDelta: 0, PreservesExactlyOnce: true},
		{Name: "lease_expiry", ExpectedCompletedDelta: 0, PreservesExactlyOnce: true},
		{Name: "duplicate_handoff", ExpectedCompletedDelta: 0, PreservesExactlyOnce: true},
	}
	fixture := AtlasExactlyOnceResumeAccountingFixture{
		Schema:                             AtlasExactlyOnceResumeAccountingFixtureContract,
		Status:                             "exactly_once_resume_accounting_ready",
		Scenarios:                          scenarios,
		ScenarioCount:                      len(scenarios),
		ExactlyOnceNodeAccounting:          true,
		DuplicateHandoffDoubleCountAllowed: false,
		RestartPreservesAccounting:         true,
		LeaseExpiryPreservesAccounting:     true,
		SchedulesWork:                      false,
		ExecutesWork:                       false,
		ApprovesWork:                       false,
		ClaimsAuthorityAdvance:             false,
		RSIRemainsDenied:                   true,
	}
	if err := ValidateAtlasExactlyOnceResumeAccountingFixture(fixture); err != nil {
		return AtlasExactlyOnceResumeAccountingFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasExactlyOnceResumeAccountingFixture(fixture AtlasExactlyOnceResumeAccountingFixture) error {
	var errs []string
	requireContract(&errs, "exactly_once_resume_accounting_fixture", fixture.Schema, AtlasExactlyOnceResumeAccountingFixtureContract)
	if fixture.Status != "exactly_once_resume_accounting_ready" {
		errs = append(errs, "status must be exactly_once_resume_accounting_ready")
	}
	if fixture.ScenarioCount != len(fixture.Scenarios) {
		errs = append(errs, "scenario_count must match scenarios")
	}
	if fixture.ScenarioCount != 3 {
		errs = append(errs, "scenario_count must be 3")
	}
	if !fixture.ExactlyOnceNodeAccounting {
		errs = append(errs, "exactly_once_node_accounting must be true")
	}
	if fixture.DuplicateHandoffDoubleCountAllowed {
		errs = append(errs, "duplicate_handoff_double_count_allowed must be false")
	}
	if !fixture.RestartPreservesAccounting {
		errs = append(errs, "restart_preserves_accounting must be true")
	}
	if !fixture.LeaseExpiryPreservesAccounting {
		errs = append(errs, "lease_expiry_preserves_accounting must be true")
	}
	seen := map[string]bool{}
	for i, scenario := range fixture.Scenarios {
		prefix := fmt.Sprintf("scenarios[%d]", i)
		requireField(&errs, prefix+".name", scenario.Name)
		if scenario.ExpectedCompletedDelta != 0 {
			errs = append(errs, prefix+".expected_completed_delta must be 0")
		}
		if !scenario.PreservesExactlyOnce {
			errs = append(errs, prefix+".preserves_exactly_once must be true")
		}
		seen[scenario.Name] = true
	}
	for _, required := range []string{"restart", "lease_expiry", "duplicate_handoff"} {
		if !seen[required] {
			errs = append(errs, "scenarios must include "+required)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
