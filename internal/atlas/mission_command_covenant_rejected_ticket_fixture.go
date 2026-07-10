package atlas

import "fmt"

func BuildAtlasCommandCovenantRejectedTicketFixture(inputPath string) (AtlasCommandCovenantRejectedTicketFixture, error) {
	input, err := LoadJSON[AtlasCommandCovenantRejectedTicketInput](inputPath)
	if err != nil {
		return AtlasCommandCovenantRejectedTicketFixture{}, err
	}
	if err := ValidateAtlasCommandCovenantRejectedTicketInput(input); err != nil {
		return AtlasCommandCovenantRejectedTicketFixture{}, err
	}
	fixture := AtlasCommandCovenantRejectedTicketFixture{
		Schema:                 AtlasCommandCovenantRejectedTicketFixtureContract,
		Status:                 "rejected_ticket_reason_preserved",
		SourceInputPath:        publicArtifactRef(inputPath),
		SourceInputDigest:      digestValue(input),
		RequestSHA256:          DigestBytes([]byte(input.RequestCanonicalJSON)),
		TicketSHA256:           DigestBytes([]byte(input.TicketCanonicalJSON)),
		CovenantDecision:       input.CovenantDecision,
		CommandAcceptsTicket:   input.CommandAcceptsTicket,
		CovenantNativeReason:   input.CovenantNativeReason,
		CommandReadbackReason:  input.CommandReadbackReason,
		ReasonPreserved:        input.CovenantNativeReason == input.CommandReadbackReason,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasCommandCovenantRejectedTicketFixture(fixture); err != nil {
		return AtlasCommandCovenantRejectedTicketFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandCovenantRejectedTicketInput(input AtlasCommandCovenantRejectedTicketInput) error {
	var errs []string
	requireContract(&errs, "command_covenant_rejected_ticket_input", input.Schema, AtlasCommandCovenantRejectedTicketInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	requireField(&errs, "request_canonical_json", input.RequestCanonicalJSON)
	requireField(&errs, "ticket_canonical_json", input.TicketCanonicalJSON)
	if input.CovenantDecision != "rejected" {
		errs = append(errs, "covenant_decision must be rejected")
	}
	if input.CommandAcceptsTicket {
		errs = append(errs, "command_accepts_ticket must be false for rejected ticket fixtures")
	}
	requireField(&errs, "covenant_native_reason", input.CovenantNativeReason)
	requireField(&errs, "command_readback_reason", input.CommandReadbackReason)
	if input.CovenantNativeReason != input.CommandReadbackReason {
		errs = append(errs, "command_readback_reason must preserve covenant_native_reason")
	}
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCommandCovenantRejectedTicketFixture(fixture AtlasCommandCovenantRejectedTicketFixture) error {
	var errs []string
	requireContract(&errs, "command_covenant_rejected_ticket_fixture", fixture.Schema, AtlasCommandCovenantRejectedTicketFixtureContract)
	if fixture.Status != "rejected_ticket_reason_preserved" {
		errs = append(errs, "status must be rejected_ticket_reason_preserved")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validateRejectedTicketDigest(&errs, "request_sha256", fixture.RequestSHA256)
	validateRejectedTicketDigest(&errs, "ticket_sha256", fixture.TicketSHA256)
	if fixture.CovenantDecision != "rejected" {
		errs = append(errs, "covenant_decision must be rejected")
	}
	if fixture.CommandAcceptsTicket {
		errs = append(errs, "command_accepts_ticket must be false")
	}
	requireField(&errs, "covenant_native_reason", fixture.CovenantNativeReason)
	requireField(&errs, "command_readback_reason", fixture.CommandReadbackReason)
	if !fixture.ReasonPreserved {
		errs = append(errs, "reason_preserved must be true")
	}
	if fixture.CovenantNativeReason != fixture.CommandReadbackReason {
		errs = append(errs, "command_readback_reason must match covenant_native_reason")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateRejectedTicketDigest(errs *[]string, field, value string) {
	if !digestPattern.MatchString(value) {
		*errs = append(*errs, fmt.Sprintf("%s must be sha256 digest", field))
	}
}
