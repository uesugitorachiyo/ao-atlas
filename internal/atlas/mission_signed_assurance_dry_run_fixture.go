package atlas

import "fmt"

func BuildAtlasSignedAssuranceDryRunFixture() (AtlasSignedAssuranceDryRunFixture, error) {
	checks := []string{"signer_identity", "evidence_digest", "freshness"}
	fixture := AtlasSignedAssuranceDryRunFixture{
		Schema:                   AtlasSignedAssuranceDryRunFixtureContract,
		Status:                   "dry_run_verification_ready",
		RequiredChecks:           checks,
		RequiredCheckCount:       len(checks),
		SignerIdentityRequired:   true,
		EvidenceDigestRequired:   true,
		FreshnessRequired:        true,
		PromotionDecisionEnabled: false,
		DryRunOnly:               true,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   false,
		RSIRemainsDenied:         true,
	}
	if err := ValidateAtlasSignedAssuranceDryRunFixture(fixture); err != nil {
		return AtlasSignedAssuranceDryRunFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasSignedAssuranceDryRunFixture(fixture AtlasSignedAssuranceDryRunFixture) error {
	var errs []string
	requireContract(&errs, "signed_assurance_dry_run_fixture", fixture.Schema, AtlasSignedAssuranceDryRunFixtureContract)
	if fixture.Status != "dry_run_verification_ready" {
		errs = append(errs, "status must be dry_run_verification_ready")
	}
	if fixture.RequiredCheckCount != len(fixture.RequiredChecks) {
		errs = append(errs, "required_check_count must match required_checks")
	}
	if fixture.RequiredCheckCount != 3 || !containsAll(fixture.RequiredChecks, []string{"signer_identity", "evidence_digest", "freshness"}) {
		errs = append(errs, "required_checks must cover signer_identity, evidence_digest, and freshness")
	}
	for i, check := range fixture.RequiredChecks {
		requireField(&errs, fmt.Sprintf("required_checks[%d]", i), check)
	}
	if !fixture.SignerIdentityRequired {
		errs = append(errs, "signer_identity_required must be true")
	}
	if !fixture.EvidenceDigestRequired {
		errs = append(errs, "evidence_digest_required must be true")
	}
	if !fixture.FreshnessRequired {
		errs = append(errs, "freshness_required must be true")
	}
	if fixture.PromotionDecisionEnabled {
		errs = append(errs, "promotion_decision_enabled must be false")
	}
	if !fixture.DryRunOnly {
		errs = append(errs, "dry_run_only must be true")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
