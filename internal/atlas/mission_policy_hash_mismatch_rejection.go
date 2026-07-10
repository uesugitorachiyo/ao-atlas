package atlas

import "fmt"

func BuildAtlasPolicyHashMismatchRejectionFixture(inputPath string) (AtlasPolicyHashMismatchRejectionFixture, error) {
	input, err := LoadJSON[AtlasPolicyHashMismatchRejectionInput](inputPath)
	if err != nil {
		return AtlasPolicyHashMismatchRejectionFixture{}, err
	}
	if err := ValidateAtlasPolicyHashMismatchRejectionInput(input); err != nil {
		return AtlasPolicyHashMismatchRejectionFixture{}, err
	}
	fixture := summarizePolicyHashMismatchRejection(input.Cases)
	fixture.Schema = AtlasPolicyHashMismatchRejectionFixtureContract
	fixture.Status = "policy_hash_mismatches_rejected"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasPolicyHashMismatchRejectionFixture(fixture); err != nil {
		return AtlasPolicyHashMismatchRejectionFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasPolicyHashMismatchRejectionInput(input AtlasPolicyHashMismatchRejectionInput) error {
	var errs []string
	requireContract(&errs, "policy_hash_mismatch_rejection_input", input.Schema, AtlasPolicyHashMismatchRejectionInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validatePolicyHashMismatchRejectionCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasPolicyHashMismatchRejectionFixture(fixture AtlasPolicyHashMismatchRejectionFixture) error {
	var errs []string
	requireContract(&errs, "policy_hash_mismatch_rejection_fixture", fixture.Schema, AtlasPolicyHashMismatchRejectionFixtureContract)
	if fixture.Status != "policy_hash_mismatches_rejected" {
		errs = append(errs, "status must be policy_hash_mismatches_rejected")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validatePolicyHashMismatchRejectionCases(&errs, fixture.Cases)
	expected := summarizePolicyHashMismatchRejection(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.RejectedCases != expected.RejectedCases {
		errs = append(errs, "rejected_cases must match cases")
	}
	if fixture.SafeToAccept != expected.SafeToAccept {
		errs = append(errs, "safe_to_accept must match cases")
	}
	if fixture.AllMismatchesRejected != expected.AllMismatchesRejected {
		errs = append(errs, "all_mismatches_rejected must match cases")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validatePolicyHashMismatchRejectionCases(errs *[]string, cases []AtlasPolicyHashMismatchRejectionCase) {
	if len(cases) == 0 {
		*errs = append(*errs, "cases must not be empty")
	}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		validateRejectedTicketDigest(errs, prefix+".ticket_policy_sha256", item.TicketPolicySHA256)
		validateRejectedTicketDigest(errs, prefix+".current_policy_sha256", item.CurrentPolicySHA256)
		if item.TicketPolicySHA256 == item.CurrentPolicySHA256 {
			*errs = append(*errs, prefix+".policy hashes must differ")
		}
		if item.ExpectedDecision != "rejected" {
			*errs = append(*errs, prefix+".expected_decision must be rejected")
		}
		requireField(errs, prefix+".rejection_reason", item.RejectionReason)
	}
}

func summarizePolicyHashMismatchRejection(cases []AtlasPolicyHashMismatchRejectionCase) AtlasPolicyHashMismatchRejectionFixture {
	fixture := AtlasPolicyHashMismatchRejectionFixture{
		CaseCount:             len(cases),
		SafeToAccept:          true,
		AllMismatchesRejected: true,
		Cases:                 append([]AtlasPolicyHashMismatchRejectionCase(nil), cases...),
	}
	for _, item := range cases {
		mismatch := item.TicketPolicySHA256 != item.CurrentPolicySHA256
		rejected := item.ExpectedDecision == "rejected"
		if mismatch {
			fixture.SafeToAccept = false
			if rejected {
				fixture.RejectedCases++
			} else {
				fixture.AllMismatchesRejected = false
			}
		}
	}
	return fixture
}
