package atlas

import "fmt"

func BuildAtlasCommandCompactRejectionReasonFixture(inputPath string) (AtlasCommandCompactRejectionReasonFixture, error) {
	input, err := LoadJSON[AtlasCommandCompactRejectionReasonInput](inputPath)
	if err != nil {
		return AtlasCommandCompactRejectionReasonFixture{}, err
	}
	if err := ValidateAtlasCommandCompactRejectionReasonInput(input); err != nil {
		return AtlasCommandCompactRejectionReasonFixture{}, err
	}
	fixture := summarizeCommandCompactRejectionReason(input.Cases)
	fixture.Schema = AtlasCommandCompactRejectionReasonFixtureContract
	fixture.Status = "command_compact_rejection_reasons_rendered"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasCommandCompactRejectionReasonFixture(fixture); err != nil {
		return AtlasCommandCompactRejectionReasonFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandCompactRejectionReasonInput(input AtlasCommandCompactRejectionReasonInput) error {
	var errs []string
	requireContract(&errs, "command_compact_rejection_reason_input", input.Schema, AtlasCommandCompactRejectionReasonInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateCommandCompactRejectionReasonCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCommandCompactRejectionReasonFixture(fixture AtlasCommandCompactRejectionReasonFixture) error {
	var errs []string
	requireContract(&errs, "command_compact_rejection_reason_fixture", fixture.Schema, AtlasCommandCompactRejectionReasonFixtureContract)
	if fixture.Status != "command_compact_rejection_reasons_rendered" {
		errs = append(errs, "status must be command_compact_rejection_reasons_rendered")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validateCommandCompactRejectionReasonCases(&errs, fixture.Cases)
	expected := summarizeCommandCompactRejectionReason(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.ReasonsRendered != expected.ReasonsRendered {
		errs = append(errs, "reasons_rendered must match cases")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateCommandCompactRejectionReasonCases(errs *[]string, cases []AtlasCommandCompactRejectionReasonCase) {
	if len(cases) == 0 {
		*errs = append(*errs, "cases must not be empty")
	}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		requireField(errs, prefix+".command_compact_status_path", item.CommandCompactStatusPath)
		checkPublicPath(errs, prefix+".command_compact_status_path", item.CommandCompactStatusPath, true)
		requireField(errs, prefix+".covenant_native_reason", item.CovenantNativeReason)
		requireField(errs, prefix+".command_compact_reason", item.CommandCompactReason)
		if item.CovenantNativeReason != item.CommandCompactReason {
			*errs = append(*errs, prefix+".command_compact_reason must match covenant_native_reason")
		}
	}
}

func summarizeCommandCompactRejectionReason(cases []AtlasCommandCompactRejectionReasonCase) AtlasCommandCompactRejectionReasonFixture {
	fixture := AtlasCommandCompactRejectionReasonFixture{
		CaseCount:       len(cases),
		ReasonsRendered: true,
		Cases:           append([]AtlasCommandCompactRejectionReasonCase(nil), cases...),
	}
	for _, item := range cases {
		if item.CovenantNativeReason == "" || item.CommandCompactReason != item.CovenantNativeReason {
			fixture.ReasonsRendered = false
			break
		}
	}
	return fixture
}
