package atlas

import "fmt"

func BuildAtlasCommandCovenantQuarantineFixture(inputPath string) (AtlasCommandCovenantQuarantineFixture, error) {
	input, err := LoadJSON[AtlasCommandCovenantQuarantineInput](inputPath)
	if err != nil {
		return AtlasCommandCovenantQuarantineFixture{}, err
	}
	if err := ValidateAtlasCommandCovenantQuarantineInput(input); err != nil {
		return AtlasCommandCovenantQuarantineFixture{}, err
	}
	fixture := summarizeCommandCovenantQuarantine(input.Cases)
	fixture.Schema = AtlasCommandCovenantQuarantineFixtureContract
	fixture.Status = "rejected_paths_quarantined"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasCommandCovenantQuarantineFixture(fixture); err != nil {
		return AtlasCommandCovenantQuarantineFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandCovenantQuarantineInput(input AtlasCommandCovenantQuarantineInput) error {
	var errs []string
	requireContract(&errs, "command_covenant_quarantine_input", input.Schema, AtlasCommandCovenantQuarantineInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	if len(input.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	for i, item := range input.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", item.ID)
		requireField(&errs, prefix+".command_validation_path", item.CommandValidationPath)
		requireField(&errs, prefix+".ticket_canonical_json", item.TicketCanonicalJSON)
		if item.CovenantDecision != "rejected" && item.CovenantDecision != "accepted" {
			errs = append(errs, prefix+".covenant_decision must be accepted or rejected")
		}
		if item.CovenantDecision == "rejected" {
			requireField(&errs, prefix+".covenant_native_reason", item.CovenantNativeReason)
		}
	}
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCommandCovenantQuarantineFixture(fixture AtlasCommandCovenantQuarantineFixture) error {
	var errs []string
	requireContract(&errs, "command_covenant_quarantine_fixture", fixture.Schema, AtlasCommandCovenantQuarantineFixtureContract)
	if fixture.Status != "rejected_paths_quarantined" {
		errs = append(errs, "status must be rejected_paths_quarantined")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	if len(fixture.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	expected := summarizeCommandCovenantQuarantineCases(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.RejectedCommandAcceptancePaths != expected.RejectedCommandAcceptancePaths {
		errs = append(errs, "rejected_command_acceptance_paths must match cases")
	}
	if fixture.QuarantinedPaths != expected.QuarantinedPaths {
		errs = append(errs, "quarantined_paths must match cases")
	}
	if fixture.SafeToAccept != expected.SafeToAccept {
		errs = append(errs, "safe_to_accept must match cases")
	}
	if fixture.AllRejectedAcceptancePathsQuarantined != expected.AllRejectedAcceptancePathsQuarantined {
		errs = append(errs, "all_rejected_acceptance_paths_quarantined must match cases")
	}
	for i, item := range fixture.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", item.ID)
		requireField(&errs, prefix+".command_validation_path", item.CommandValidationPath)
		validateRejectedTicketDigest(&errs, prefix+".ticket_sha256", item.TicketSHA256)
		if item.CovenantDecision != "rejected" && item.CovenantDecision != "accepted" {
			errs = append(errs, prefix+".covenant_decision must be accepted or rejected")
		}
		if item.CovenantDecision == "rejected" {
			requireField(&errs, prefix+".covenant_native_reason", item.CovenantNativeReason)
		}
		if item.QuarantineDecision != "quarantined" && item.QuarantineDecision != "not_required" {
			errs = append(errs, prefix+".quarantine_decision must be quarantined or not_required")
		}
		if item.CovenantDecision == "rejected" && item.CommandWouldAcceptTicket && item.QuarantineDecision != "quarantined" {
			errs = append(errs, prefix+".quarantine_decision must quarantine rejected tickets Command would accept")
		}
		if item.QuarantineDecision == "quarantined" && item.SafeToAccept {
			errs = append(errs, prefix+".safe_to_accept must be false when quarantined")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeCommandCovenantQuarantine(cases []AtlasCommandCovenantQuarantineCase) AtlasCommandCovenantQuarantineFixture {
	fixtureCases := make([]AtlasCommandCovenantQuarantineFixtureCase, 0, len(cases))
	for _, item := range cases {
		quarantineDecision := "not_required"
		safeToAccept := !item.CommandWouldAcceptTicket || item.CovenantDecision != "rejected"
		if item.CovenantDecision == "rejected" && item.CommandWouldAcceptTicket {
			quarantineDecision = "quarantined"
			safeToAccept = false
		}
		fixtureCases = append(fixtureCases, AtlasCommandCovenantQuarantineFixtureCase{
			ID:                       item.ID,
			CommandValidationPath:    item.CommandValidationPath,
			TicketSHA256:             DigestBytes([]byte(item.TicketCanonicalJSON)),
			CovenantDecision:         item.CovenantDecision,
			CommandWouldAcceptTicket: item.CommandWouldAcceptTicket,
			CovenantNativeReason:     item.CovenantNativeReason,
			QuarantineDecision:       quarantineDecision,
			SafeToAccept:             safeToAccept,
		})
	}
	return summarizeCommandCovenantQuarantineCases(fixtureCases)
}

func summarizeCommandCovenantQuarantineCases(cases []AtlasCommandCovenantQuarantineFixtureCase) AtlasCommandCovenantQuarantineFixture {
	fixture := AtlasCommandCovenantQuarantineFixture{
		CaseCount:                             len(cases),
		SafeToAccept:                          true,
		AllRejectedAcceptancePathsQuarantined: true,
		Cases:                                 append([]AtlasCommandCovenantQuarantineFixtureCase(nil), cases...),
	}
	for _, item := range cases {
		rejectedAcceptance := item.CovenantDecision == "rejected" && item.CommandWouldAcceptTicket
		if rejectedAcceptance {
			fixture.RejectedCommandAcceptancePaths++
			fixture.SafeToAccept = false
			if item.QuarantineDecision == "quarantined" && !item.SafeToAccept {
				fixture.QuarantinedPaths++
			} else {
				fixture.AllRejectedAcceptancePathsQuarantined = false
			}
		}
	}
	return fixture
}
