package atlas

import "fmt"

func BuildAtlasPromoterNoActivationBoundaryFixture() (AtlasPromoterNoActivationBoundaryFixture, error) {
	allowed := []string{"no_promotion", "blocked", "insufficient_evidence"}
	forbidden := []string{"activate", "release", "deploy", "publish", "tag"}
	fixture := AtlasPromoterNoActivationBoundaryFixture{
		Schema:                       AtlasPromoterNoActivationBoundaryFixtureContract,
		Status:                       "no_promotion_boundary_ready",
		Decision:                     "no_promotion",
		NoPromotionDecisionSupported: true,
		ActivationExecutionOwned:     false,
		ReleaseExecutionOwned:        false,
		AllowedOutputs:               allowed,
		AllowedOutputCount:           len(allowed),
		ForbiddenActions:             forbidden,
		ForbiddenActionCount:         len(forbidden),
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
	}
	if err := ValidateAtlasPromoterNoActivationBoundaryFixture(fixture); err != nil {
		return AtlasPromoterNoActivationBoundaryFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasPromoterNoActivationBoundaryFixture(fixture AtlasPromoterNoActivationBoundaryFixture) error {
	var errs []string
	requireContract(&errs, "promoter_no_activation_boundary_fixture", fixture.Schema, AtlasPromoterNoActivationBoundaryFixtureContract)
	if fixture.Status != "no_promotion_boundary_ready" {
		errs = append(errs, "status must be no_promotion_boundary_ready")
	}
	if fixture.Decision != "no_promotion" {
		errs = append(errs, "decision must be no_promotion")
	}
	if !fixture.NoPromotionDecisionSupported {
		errs = append(errs, "no_promotion_decision_supported must be true")
	}
	if fixture.ActivationExecutionOwned {
		errs = append(errs, "activation_execution_owned must be false")
	}
	if fixture.ReleaseExecutionOwned {
		errs = append(errs, "release_execution_owned must be false")
	}
	if fixture.AllowedOutputCount != len(fixture.AllowedOutputs) {
		errs = append(errs, "allowed_output_count must match allowed_outputs")
	}
	if fixture.AllowedOutputCount != 3 || !containsAll(fixture.AllowedOutputs, []string{"no_promotion", "blocked", "insufficient_evidence"}) {
		errs = append(errs, "allowed_outputs must cover no_promotion, blocked, and insufficient_evidence")
	}
	for i, output := range fixture.AllowedOutputs {
		requireField(&errs, fmt.Sprintf("allowed_outputs[%d]", i), output)
	}
	if fixture.ForbiddenActionCount != len(fixture.ForbiddenActions) {
		errs = append(errs, "forbidden_action_count must match forbidden_actions")
	}
	if fixture.ForbiddenActionCount != 5 || !containsAll(fixture.ForbiddenActions, []string{"activate", "release", "deploy", "publish", "tag"}) {
		errs = append(errs, "forbidden_actions must cover activate, release, deploy, publish, and tag")
	}
	for i, action := range fixture.ForbiddenActions {
		requireField(&errs, fmt.Sprintf("forbidden_actions[%d]", i), action)
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
