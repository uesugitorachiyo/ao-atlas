package atlas

func BuildAtlasRollbackTerminalReadbackFixture() (AtlasRollbackTerminalReadbackFixture, error) {
	fixture := AtlasRollbackTerminalReadbackFixture{
		Schema:                  AtlasRollbackTerminalReadbackFixtureContract,
		Status:                  "rollback_terminal_readbacks_agree",
		TerminalState:           "rolled_back",
		Readbacks:               []string{"command", "sentinel", "promoter", "control_plane"},
		ReadbackAgreementCount:  4,
		RollbackReceiptReplayed: true,
		ReadbacksAgree:          true,
		PromotionRequested:      false,
		LiveProviderCalls:       false,
		SchedulesWork:           false,
		ExecutesWork:            false,
		ApprovesWork:            false,
		ClaimsAuthorityAdvance:  false,
		RSIRemainsDenied:        true,
	}
	if err := ValidateAtlasRollbackTerminalReadbackFixture(fixture); err != nil {
		return AtlasRollbackTerminalReadbackFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasRollbackTerminalReadbackFixture(fixture AtlasRollbackTerminalReadbackFixture) error {
	var errs []string
	requireContract(&errs, "rollback_terminal_readback_fixture", fixture.Schema, AtlasRollbackTerminalReadbackFixtureContract)
	if fixture.Status != "rollback_terminal_readbacks_agree" {
		errs = append(errs, "status must be rollback_terminal_readbacks_agree")
	}
	if fixture.TerminalState != "rolled_back" {
		errs = append(errs, "terminal_state must be rolled_back")
	}
	if fixture.ReadbackAgreementCount != len(fixture.Readbacks) || fixture.ReadbackAgreementCount != 4 {
		errs = append(errs, "readback_agreement_count must be 4")
	}
	if !fixture.RollbackReceiptReplayed {
		errs = append(errs, "rollback_receipt_replayed must be true")
	}
	if !fixture.ReadbacksAgree {
		errs = append(errs, "readbacks_agree must be true")
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
