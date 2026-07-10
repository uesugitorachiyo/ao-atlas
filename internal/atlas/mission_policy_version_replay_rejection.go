package atlas

import "fmt"

func BuildAtlasPolicyVersionReplayRejectionFixture(inputPath string) (AtlasPolicyVersionReplayRejectionFixture, error) {
	input, err := LoadJSON[AtlasPolicyVersionReplayRejectionInput](inputPath)
	if err != nil {
		return AtlasPolicyVersionReplayRejectionFixture{}, err
	}
	if err := ValidateAtlasPolicyVersionReplayRejectionInput(input); err != nil {
		return AtlasPolicyVersionReplayRejectionFixture{}, err
	}
	fixture := summarizePolicyVersionReplayRejection(input.Cases)
	fixture.Schema = AtlasPolicyVersionReplayRejectionFixtureContract
	fixture.Status = "stale_policy_versions_rejected"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasPolicyVersionReplayRejectionFixture(fixture); err != nil {
		return AtlasPolicyVersionReplayRejectionFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasPolicyVersionReplayRejectionInput(input AtlasPolicyVersionReplayRejectionInput) error {
	var errs []string
	requireContract(&errs, "policy_version_replay_rejection_input", input.Schema, AtlasPolicyVersionReplayRejectionInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validatePolicyVersionReplayRejectionCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasPolicyVersionReplayRejectionFixture(fixture AtlasPolicyVersionReplayRejectionFixture) error {
	var errs []string
	requireContract(&errs, "policy_version_replay_rejection_fixture", fixture.Schema, AtlasPolicyVersionReplayRejectionFixtureContract)
	if fixture.Status != "stale_policy_versions_rejected" {
		errs = append(errs, "status must be stale_policy_versions_rejected")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validatePolicyVersionReplayRejectionCases(&errs, fixture.Cases)
	expected := summarizePolicyVersionReplayRejection(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.RejectedCases != expected.RejectedCases {
		errs = append(errs, "rejected_cases must match cases")
	}
	if fixture.SafeToAccept != expected.SafeToAccept {
		errs = append(errs, "safe_to_accept must match cases")
	}
	if fixture.AllStaleVersionsRejected != expected.AllStaleVersionsRejected {
		errs = append(errs, "all_stale_versions_rejected must match cases")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validatePolicyVersionReplayRejectionCases(errs *[]string, cases []AtlasPolicyVersionReplayRejectionCase) {
	if len(cases) == 0 {
		*errs = append(*errs, "cases must not be empty")
	}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		validateRejectedTicketDigest(errs, prefix+".ticket_sha256", item.TicketSHA256)
		requireField(errs, prefix+".ticket_policy_version", item.TicketPolicyVersion)
		requireField(errs, prefix+".current_policy_version", item.CurrentPolicyVersion)
		if item.TicketPolicyVersion == item.CurrentPolicyVersion {
			*errs = append(*errs, prefix+".policy versions must differ")
		}
		if item.ExpectedDecision != "rejected" {
			*errs = append(*errs, prefix+".expected_decision must be rejected")
		}
		requireField(errs, prefix+".rejection_reason", item.RejectionReason)
	}
}

func summarizePolicyVersionReplayRejection(cases []AtlasPolicyVersionReplayRejectionCase) AtlasPolicyVersionReplayRejectionFixture {
	fixture := AtlasPolicyVersionReplayRejectionFixture{
		CaseCount:                len(cases),
		SafeToAccept:             true,
		AllStaleVersionsRejected: true,
		Cases:                    append([]AtlasPolicyVersionReplayRejectionCase(nil), cases...),
	}
	for _, item := range cases {
		stale := item.TicketPolicyVersion != item.CurrentPolicyVersion
		rejected := item.ExpectedDecision == "rejected"
		if stale {
			fixture.SafeToAccept = false
			if rejected {
				fixture.RejectedCases++
			} else {
				fixture.AllStaleVersionsRejected = false
			}
		}
	}
	return fixture
}
