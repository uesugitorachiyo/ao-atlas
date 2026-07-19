package atlas

import "fmt"

func BuildAtlasForgeGoalRunEvidenceFixture() (AtlasForgeGoalRunEvidenceFixture, error) {
	required := []string{"goalrun_start", "stop_gate", "rollback_record", "terminal_receipt"}
	fixture := AtlasForgeGoalRunEvidenceFixture{
		Schema:                   AtlasForgeGoalRunEvidenceFixtureContract,
		Status:                   "forge_goalrun_evidence_ready",
		RequiredEvidence:         required,
		RequiredEvidenceCount:    len(required),
		GoalRunStartRequired:     true,
		StopGateRequired:         true,
		RollbackRecordRequired:   true,
		TerminalReceiptRequired:  true,
		ProviderExecutionAllowed: false,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   false,
		RSIRemainsDenied:         true,
	}
	if err := ValidateAtlasForgeGoalRunEvidenceFixture(fixture); err != nil {
		return AtlasForgeGoalRunEvidenceFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasForgeGoalRunEvidenceFixture(fixture AtlasForgeGoalRunEvidenceFixture) error {
	var errs []string
	requireContract(&errs, "forge_goalrun_evidence_fixture", fixture.Schema, AtlasForgeGoalRunEvidenceFixtureContract)
	if fixture.Status != "forge_goalrun_evidence_ready" {
		errs = append(errs, "status must be forge_goalrun_evidence_ready")
	}
	if fixture.RequiredEvidenceCount != len(fixture.RequiredEvidence) {
		errs = append(errs, "required_evidence_count must match required_evidence")
	}
	if fixture.RequiredEvidenceCount != 4 || !containsAll(fixture.RequiredEvidence, []string{"goalrun_start", "stop_gate", "rollback_record", "terminal_receipt"}) {
		errs = append(errs, "required_evidence must cover goalrun_start, stop_gate, rollback_record, and terminal_receipt")
	}
	for i, evidence := range fixture.RequiredEvidence {
		requireField(&errs, fmt.Sprintf("required_evidence[%d]", i), evidence)
	}
	if !fixture.GoalRunStartRequired {
		errs = append(errs, "goalrun_start_required must be true")
	}
	if !fixture.StopGateRequired {
		errs = append(errs, "stop_gate_required must be true")
	}
	if !fixture.RollbackRecordRequired {
		errs = append(errs, "rollback_record_required must be true")
	}
	if !fixture.TerminalReceiptRequired {
		errs = append(errs, "terminal_receipt_required must be true")
	}
	if fixture.ProviderExecutionAllowed {
		errs = append(errs, "provider_execution_allowed must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
