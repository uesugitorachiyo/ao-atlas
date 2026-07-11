package atlas

import "fmt"

func BuildAtlasBoundedSignerContractFixture() (AtlasBoundedSignerContractFixture, error) {
	fixture := AtlasBoundedSignerContractFixture{
		Schema:             AtlasBoundedSignerContractFixtureContract,
		Status:             "bounded_signer_contract_ready",
		ContractPurpose:    "golden_path_signed_assurance_input_verification",
		RotationBoundary:   "overlap_required",
		RevocationBoundary: "deny_on_or_after_revoked_at",
		SignerIdentities: []AtlasBoundedSignerIdentity{
			{
				ID:          "ao-assurance-primary-2026q3",
				Role:        "assurance_input_signer",
				KeyRef:      "covenant://signers/ao-assurance-primary-2026q3",
				ValidFrom:   "2026-07-10T00:00:00Z",
				ValidUntil:  "2026-10-10T00:00:00Z",
				Revoked:     false,
				Fingerprint: "sha256:1111111111111111111111111111111111111111111111111111111111111111",
			},
			{
				ID:          "ao-assurance-previous-2026q2",
				Role:        "assurance_input_signer",
				KeyRef:      "covenant://signers/ao-assurance-previous-2026q2",
				ValidFrom:   "2026-04-10T00:00:00Z",
				ValidUntil:  "2026-08-10T00:00:00Z",
				RevokedAt:   "2026-08-10T00:00:00Z",
				Revoked:     true,
				Fingerprint: "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			},
		},
		RequiredBindings: []string{
			"signer_identity",
			"signer_fingerprint",
			"key_ref",
			"evidence_digest",
			"base_commit",
			"issued_at",
			"expires_at",
			"revocation_epoch",
		},
		VerificationRules: []AtlasBoundedSignerVerification{
			{
				ID:          "validity-window",
				Description: "accept signatures only within signer and ticket validity windows",
				Required:    true,
			},
			{
				ID:          "rotation-overlap",
				Description: "require explicit overlap during rotation and deny gaps between old and new signer records",
				Required:    true,
			},
			{
				ID:          "revocation-deny",
				Description: "deny signatures issued on or after revoked_at for revoked signer records",
				Required:    true,
			},
			{
				ID:          "digest-binding",
				Description: "bind every signature to evidence_digest and base_commit before any promotion readback",
				Required:    true,
			},
		},
		LiveKeyManagement:      false,
		PolicyWidening:         false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	fixture.SignerCount = len(fixture.SignerIdentities)
	if err := ValidateAtlasBoundedSignerContractFixture(fixture); err != nil {
		return AtlasBoundedSignerContractFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasBoundedSignerContractFixture(fixture AtlasBoundedSignerContractFixture) error {
	var errs []string
	requireContract(&errs, "bounded_signer_contract_fixture", fixture.Schema, AtlasBoundedSignerContractFixtureContract)
	if fixture.Status != "bounded_signer_contract_ready" {
		errs = append(errs, "status must be bounded_signer_contract_ready")
	}
	requireField(&errs, "contract_purpose", fixture.ContractPurpose)
	if fixture.RotationBoundary != "overlap_required" {
		errs = append(errs, "rotation_boundary must be overlap_required")
	}
	if fixture.RevocationBoundary != "deny_on_or_after_revoked_at" {
		errs = append(errs, "revocation_boundary must be deny_on_or_after_revoked_at")
	}
	if fixture.SignerCount != len(fixture.SignerIdentities) {
		errs = append(errs, "signer_count must match signer_identities")
	}
	if fixture.SignerCount < 2 {
		errs = append(errs, "signer_count must be at least two to model rotation")
	}
	validateBoundedSignerIdentities(&errs, fixture.SignerIdentities)
	validateBoundedSignerRequiredBindings(&errs, fixture.RequiredBindings)
	validateBoundedSignerVerificationRules(&errs, fixture.VerificationRules)
	if fixture.LiveKeyManagement {
		errs = append(errs, "live_key_management must be false")
	}
	if fixture.PolicyWidening {
		errs = append(errs, "policy_widening must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateBoundedSignerRequiredBindings(errs *[]string, bindings []string) {
	if len(bindings) == 0 {
		*errs = append(*errs, "required_bindings must not be empty")
	}
	for i, binding := range bindings {
		requireField(errs, fmt.Sprintf("required_bindings[%d]", i), binding)
	}
}

func validateBoundedSignerIdentities(errs *[]string, identities []AtlasBoundedSignerIdentity) {
	if len(identities) == 0 {
		*errs = append(*errs, "signer_identities must not be empty")
	}
	seenActive := false
	seenRevoked := false
	for i, identity := range identities {
		prefix := fmt.Sprintf("signer_identities[%d]", i)
		requireField(errs, prefix+".id", identity.ID)
		requireField(errs, prefix+".role", identity.Role)
		requireField(errs, prefix+".key_ref", identity.KeyRef)
		checkPublicPath(errs, prefix+".key_ref", identity.KeyRef, true)
		requireField(errs, prefix+".valid_from", identity.ValidFrom)
		requireField(errs, prefix+".valid_until", identity.ValidUntil)
		validateRejectedTicketDigest(errs, prefix+".fingerprint", identity.Fingerprint)
		if identity.Revoked {
			seenRevoked = true
			requireField(errs, prefix+".revoked_at", identity.RevokedAt)
		} else {
			seenActive = true
			if identity.RevokedAt != "" {
				*errs = append(*errs, prefix+".revoked_at must be empty for active signer")
			}
		}
	}
	if !seenActive {
		*errs = append(*errs, "signer_identities must include active signer")
	}
	if !seenRevoked {
		*errs = append(*errs, "signer_identities must include revoked signer")
	}
}

func validateBoundedSignerVerificationRules(errs *[]string, rules []AtlasBoundedSignerVerification) {
	if len(rules) == 0 {
		*errs = append(*errs, "verification_rules must not be empty")
	}
	for i, rule := range rules {
		prefix := fmt.Sprintf("verification_rules[%d]", i)
		requireField(errs, prefix+".id", rule.ID)
		requireField(errs, prefix+".description", rule.Description)
		if !rule.Required {
			*errs = append(*errs, prefix+".required must be true")
		}
	}
}
