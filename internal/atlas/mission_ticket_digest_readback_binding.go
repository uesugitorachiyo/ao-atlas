package atlas

import "fmt"

func BuildAtlasTicketDigestReadbackBindingFixture(inputPath string) (AtlasTicketDigestReadbackBindingFixture, error) {
	input, err := LoadJSON[AtlasTicketDigestReadbackBindingInput](inputPath)
	if err != nil {
		return AtlasTicketDigestReadbackBindingFixture{}, err
	}
	if err := ValidateAtlasTicketDigestReadbackBindingInput(input); err != nil {
		return AtlasTicketDigestReadbackBindingFixture{}, err
	}
	fixture := summarizeTicketDigestReadbackBinding(input.Cases)
	fixture.Schema = AtlasTicketDigestReadbackBindingFixtureContract
	fixture.Status = "ticket_digest_readbacks_bound"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasTicketDigestReadbackBindingFixture(fixture); err != nil {
		return AtlasTicketDigestReadbackBindingFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasTicketDigestReadbackBindingInput(input AtlasTicketDigestReadbackBindingInput) error {
	var errs []string
	requireContract(&errs, "ticket_digest_readback_binding_input", input.Schema, AtlasTicketDigestReadbackBindingInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateTicketDigestReadbackBindingCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasTicketDigestReadbackBindingFixture(fixture AtlasTicketDigestReadbackBindingFixture) error {
	var errs []string
	requireContract(&errs, "ticket_digest_readback_binding_fixture", fixture.Schema, AtlasTicketDigestReadbackBindingFixtureContract)
	if fixture.Status != "ticket_digest_readbacks_bound" {
		errs = append(errs, "status must be ticket_digest_readbacks_bound")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validateTicketDigestReadbackBindingCases(&errs, fixture.Cases)
	expected := summarizeTicketDigestReadbackBinding(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.MismatchedCases != expected.MismatchedCases {
		errs = append(errs, "mismatched_cases must match cases")
	}
	if fixture.DigestBindingPassed != expected.DigestBindingPassed {
		errs = append(errs, "digest_binding_passed must match cases")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateTicketDigestReadbackBindingCases(errs *[]string, cases []AtlasTicketDigestReadbackBindingCase) {
	if len(cases) == 0 {
		*errs = append(*errs, "cases must not be empty")
	}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		requireField(errs, prefix+".command_readback_path", item.CommandReadbackPath)
		requireField(errs, prefix+".covenant_readback_path", item.CovenantReadbackPath)
		validateRejectedTicketDigest(errs, prefix+".command_ticket_sha256", item.CommandTicketSHA256)
		validateRejectedTicketDigest(errs, prefix+".covenant_ticket_sha256", item.CovenantTicketSHA256)
	}
}

func summarizeTicketDigestReadbackBinding(cases []AtlasTicketDigestReadbackBindingCase) AtlasTicketDigestReadbackBindingFixture {
	fixture := AtlasTicketDigestReadbackBindingFixture{
		CaseCount:           len(cases),
		DigestBindingPassed: true,
		Cases:               append([]AtlasTicketDigestReadbackBindingCase(nil), cases...),
	}
	for _, item := range cases {
		if item.CommandTicketSHA256 != item.CovenantTicketSHA256 {
			fixture.MismatchedCases++
			fixture.DigestBindingPassed = false
		}
	}
	return fixture
}
