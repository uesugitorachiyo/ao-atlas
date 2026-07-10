package atlas

import "fmt"

func BuildAtlasCommandTicketBytePreservationFixture(inputPath string) (AtlasCommandTicketBytePreservationFixture, error) {
	input, err := LoadJSON[AtlasCommandTicketBytePreservationInput](inputPath)
	if err != nil {
		return AtlasCommandTicketBytePreservationFixture{}, err
	}
	if err := ValidateAtlasCommandTicketBytePreservationInput(input); err != nil {
		return AtlasCommandTicketBytePreservationFixture{}, err
	}
	fixture := summarizeCommandTicketBytePreservation(input.Cases)
	fixture.Schema = AtlasCommandTicketBytePreservationFixtureContract
	fixture.Status = "ticket_bytes_preserved"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasCommandTicketBytePreservationFixture(fixture); err != nil {
		return AtlasCommandTicketBytePreservationFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandTicketBytePreservationInput(input AtlasCommandTicketBytePreservationInput) error {
	var errs []string
	requireContract(&errs, "command_ticket_byte_preservation_input", input.Schema, AtlasCommandTicketBytePreservationInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	if len(input.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	for i, item := range input.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", item.ID)
		requireField(&errs, prefix+".command_readback_path", item.CommandReadbackPath)
		requireField(&errs, prefix+".ticket_canonical_json", item.TicketCanonicalJSON)
		requireField(&errs, prefix+".command_reported_ticket_bytes", item.CommandReportedTicketBytes)
	}
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCommandTicketBytePreservationFixture(fixture AtlasCommandTicketBytePreservationFixture) error {
	var errs []string
	requireContract(&errs, "command_ticket_byte_preservation_fixture", fixture.Schema, AtlasCommandTicketBytePreservationFixtureContract)
	if fixture.Status != "ticket_bytes_preserved" {
		errs = append(errs, "status must be ticket_bytes_preserved")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	if len(fixture.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	expected := summarizeCommandTicketBytePreservationCases(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.MismatchedCases != expected.MismatchedCases {
		errs = append(errs, "mismatched_cases must match cases")
	}
	if fixture.BytePreservationPassed != expected.BytePreservationPassed {
		errs = append(errs, "byte_preservation_passed must match cases")
	}
	for i, item := range fixture.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", item.ID)
		requireField(&errs, prefix+".command_readback_path", item.CommandReadbackPath)
		validateRejectedTicketDigest(&errs, prefix+".ticket_sha256", item.TicketSHA256)
		validateRejectedTicketDigest(&errs, prefix+".command_reported_ticket_sha256", item.CommandReportedTicketSHA256)
		if item.BytesPreserved != (item.TicketSHA256 == item.CommandReportedTicketSHA256) {
			errs = append(errs, prefix+".bytes_preserved must match digest equality")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeCommandTicketBytePreservation(cases []AtlasCommandTicketBytePreservationInCase) AtlasCommandTicketBytePreservationFixture {
	fixtureCases := make([]AtlasCommandTicketBytePreservationFixtureCase, 0, len(cases))
	for _, item := range cases {
		ticketDigest := DigestBytes([]byte(item.TicketCanonicalJSON))
		commandDigest := DigestBytes([]byte(item.CommandReportedTicketBytes))
		fixtureCases = append(fixtureCases, AtlasCommandTicketBytePreservationFixtureCase{
			ID:                          item.ID,
			CommandReadbackPath:         item.CommandReadbackPath,
			TicketSHA256:                ticketDigest,
			CommandReportedTicketSHA256: commandDigest,
			BytesPreserved:              ticketDigest == commandDigest,
		})
	}
	return summarizeCommandTicketBytePreservationCases(fixtureCases)
}

func summarizeCommandTicketBytePreservationCases(cases []AtlasCommandTicketBytePreservationFixtureCase) AtlasCommandTicketBytePreservationFixture {
	fixture := AtlasCommandTicketBytePreservationFixture{
		CaseCount:              len(cases),
		BytePreservationPassed: true,
		Cases:                  append([]AtlasCommandTicketBytePreservationFixtureCase(nil), cases...),
	}
	for _, item := range cases {
		if !item.BytesPreserved {
			fixture.MismatchedCases++
			fixture.BytePreservationPassed = false
		}
	}
	return fixture
}
