package atlas

import "fmt"

func BuildAtlasFailureInjectionFuzzingFixture() (AtlasFailureInjectionFuzzingFixture, error) {
	cases := []AtlasFailureInjectionFuzzingCase{
		{Name: "malformed_gate", FailureClass: "malformed_gate", ExpectedState: "rejected_before_execution", Replayable: true},
		{Name: "lost_lease", FailureClass: "lost_lease", ExpectedState: "resume_requires_fresh_lease", Replayable: true},
		{Name: "stale_evidence", FailureClass: "stale_evidence", ExpectedState: "readback_marks_evidence_stale", Replayable: true},
		{Name: "rollback_receipt", FailureClass: "rollback_receipt", ExpectedState: "rollback_receipt_required", Replayable: true},
	}
	fixture := AtlasFailureInjectionFuzzingFixture{
		Schema:                 AtlasFailureInjectionFuzzingFixtureContract,
		Status:                 "failure_injection_fuzzing_ready",
		Cases:                  cases,
		CaseCount:              len(cases),
		DeterministicFuzzing:   true,
		ReplayableCases:        true,
		LiveProviderCalls:      false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasFailureInjectionFuzzingFixture(fixture); err != nil {
		return AtlasFailureInjectionFuzzingFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasFailureInjectionFuzzingFixture(fixture AtlasFailureInjectionFuzzingFixture) error {
	var errs []string
	requireContract(&errs, "failure_injection_fuzzing_fixture", fixture.Schema, AtlasFailureInjectionFuzzingFixtureContract)
	if fixture.Status != "failure_injection_fuzzing_ready" {
		errs = append(errs, "status must be failure_injection_fuzzing_ready")
	}
	if fixture.CaseCount != len(fixture.Cases) {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.CaseCount != 4 {
		errs = append(errs, "case_count must be 4")
	}
	if !fixture.DeterministicFuzzing {
		errs = append(errs, "deterministic_fuzzing must be true")
	}
	if !fixture.ReplayableCases {
		errs = append(errs, "replayable_cases must be true")
	}
	if fixture.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	required := map[string]bool{
		"malformed_gate":   false,
		"lost_lease":       false,
		"stale_evidence":   false,
		"rollback_receipt": false,
	}
	for i, item := range fixture.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".name", item.Name)
		requireField(&errs, prefix+".failure_class", item.FailureClass)
		requireField(&errs, prefix+".expected_state", item.ExpectedState)
		if _, ok := required[item.FailureClass]; ok {
			required[item.FailureClass] = true
		}
		if !item.Replayable {
			errs = append(errs, prefix+".replayable must be true")
		}
	}
	for name, seen := range required {
		if !seen {
			errs = append(errs, name+" case is required")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
