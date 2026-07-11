package atlas

import "fmt"

func BuildAtlasAtomicEvidenceTransitionFixture() (AtlasAtomicEvidenceTransitionFixture, error) {
	scenarios := []AtlasAtomicEvidenceTransitionScenario{
		{Name: "crash", AtomicTransition: true, DeterministicReplay: true, DuplicateIngestIdempotent: true},
		{Name: "lease_expiry", AtomicTransition: true, DeterministicReplay: true, DuplicateIngestIdempotent: true},
		{Name: "restart", AtomicTransition: true, DeterministicReplay: true, DuplicateIngestIdempotent: true},
		{Name: "duplicate_ingest", AtomicTransition: true, DeterministicReplay: true, DuplicateIngestIdempotent: true},
	}
	fixture := AtlasAtomicEvidenceTransitionFixture{
		Schema:                      AtlasAtomicEvidenceTransitionFixtureContract,
		Status:                      "atomic_evidence_transition_ready",
		Scenarios:                   scenarios,
		ScenarioCount:               len(scenarios),
		AtomicTransitionsRequired:   true,
		DeterministicReplayRequired: true,
		DuplicateIngestIdempotent:   true,
		SchedulesWork:               false,
		ExecutesWork:                false,
		ApprovesWork:                false,
		ClaimsAuthorityAdvance:      false,
		RSIRemainsDenied:            true,
	}
	if err := ValidateAtlasAtomicEvidenceTransitionFixture(fixture); err != nil {
		return AtlasAtomicEvidenceTransitionFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasAtomicEvidenceTransitionFixture(fixture AtlasAtomicEvidenceTransitionFixture) error {
	var errs []string
	requireContract(&errs, "atomic_evidence_transition_fixture", fixture.Schema, AtlasAtomicEvidenceTransitionFixtureContract)
	if fixture.Status != "atomic_evidence_transition_ready" {
		errs = append(errs, "status must be atomic_evidence_transition_ready")
	}
	if fixture.ScenarioCount != len(fixture.Scenarios) {
		errs = append(errs, "scenario_count must match scenarios")
	}
	if fixture.ScenarioCount != 4 {
		errs = append(errs, "scenario_count must be 4")
	}
	if !fixture.AtomicTransitionsRequired {
		errs = append(errs, "atomic_transitions_required must be true")
	}
	if !fixture.DeterministicReplayRequired {
		errs = append(errs, "deterministic_replay_required must be true")
	}
	if !fixture.DuplicateIngestIdempotent {
		errs = append(errs, "duplicate_ingest_idempotent must be true")
	}
	seen := map[string]bool{}
	for i, scenario := range fixture.Scenarios {
		prefix := fmt.Sprintf("scenarios[%d]", i)
		requireField(&errs, prefix+".name", scenario.Name)
		if !scenario.AtomicTransition {
			errs = append(errs, prefix+".atomic_transition must be true")
		}
		if !scenario.DeterministicReplay {
			errs = append(errs, prefix+".deterministic_replay must be true")
		}
		if !scenario.DuplicateIngestIdempotent {
			errs = append(errs, prefix+".duplicate_ingest_idempotent must be true")
		}
		seen[scenario.Name] = true
	}
	for _, required := range []string{"crash", "lease_expiry", "restart", "duplicate_ingest"} {
		if !seen[required] {
			errs = append(errs, "scenarios must include "+required)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
