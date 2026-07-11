package atlas

import "reflect"

func BuildAtlasCommandCovenantFieldParityFixture() (AtlasCommandCovenantFieldParityFixture, error) {
	policyFields := []string{"policy_id", "policy_version", "policy_digest", "decision"}
	approvalFields := []string{"approval_id", "subject_digest", "base_commit", "approved_by"}
	fixture := AtlasCommandCovenantFieldParityFixture{
		Schema:                           AtlasCommandCovenantFieldParityFixtureContract,
		Status:                           "field_parity_verified",
		PolicyFields:                     policyFields,
		ApprovalFields:                   approvalFields,
		CommandReadbackPolicyFields:      append([]string(nil), policyFields...),
		CommandReadbackApprovalFields:    append([]string(nil), approvalFields...),
		CommandAcceptsOnlyCovenantFields: true,
		CovenantValidatesSameFields:      true,
		RejectedExtraFields:              []string{"policy_override", "approval_substitute", "unbound_ticket_digest"},
		SchedulesWork:                    false,
		ExecutesWork:                     false,
		ApprovesWork:                     false,
		ClaimsAuthorityAdvance:           false,
		RSIRemainsDenied:                 true,
	}
	fixture.PolicyFieldCount = len(fixture.PolicyFields)
	fixture.ApprovalFieldCount = len(fixture.ApprovalFields)
	fixture.RejectedExtraFieldCount = len(fixture.RejectedExtraFields)
	if err := ValidateAtlasCommandCovenantFieldParityFixture(fixture); err != nil {
		return AtlasCommandCovenantFieldParityFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandCovenantFieldParityFixture(fixture AtlasCommandCovenantFieldParityFixture) error {
	var errs []string
	requireContract(&errs, "command_covenant_field_parity_fixture", fixture.Schema, AtlasCommandCovenantFieldParityFixtureContract)
	if fixture.Status != "field_parity_verified" {
		errs = append(errs, "status must be field_parity_verified")
	}
	if fixture.PolicyFieldCount != len(fixture.PolicyFields) {
		errs = append(errs, "policy_field_count must match policy_fields")
	}
	if fixture.ApprovalFieldCount != len(fixture.ApprovalFields) {
		errs = append(errs, "approval_field_count must match approval_fields")
	}
	if fixture.PolicyFieldCount == 0 || fixture.ApprovalFieldCount == 0 {
		errs = append(errs, "policy and approval fields must not be empty")
	}
	if !reflect.DeepEqual(fixture.PolicyFields, fixture.CommandReadbackPolicyFields) {
		errs = append(errs, "command_readback_policy_fields must match policy_fields")
	}
	if !reflect.DeepEqual(fixture.ApprovalFields, fixture.CommandReadbackApprovalFields) {
		errs = append(errs, "command_readback_approval_fields must match approval_fields")
	}
	if !fixture.CommandAcceptsOnlyCovenantFields {
		errs = append(errs, "command_accepts_only_covenant_fields must be true")
	}
	if !fixture.CovenantValidatesSameFields {
		errs = append(errs, "covenant_validates_same_fields must be true")
	}
	if fixture.RejectedExtraFieldCount != len(fixture.RejectedExtraFields) {
		errs = append(errs, "rejected_extra_field_count must match rejected_extra_fields")
	}
	for _, field := range append(append([]string{}, fixture.PolicyFields...), fixture.ApprovalFields...) {
		requireField(&errs, "field", field)
	}
	for _, field := range fixture.RejectedExtraFields {
		requireField(&errs, "rejected_extra_field", field)
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
