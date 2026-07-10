package atlas

import "fmt"

func BuildAtlasCovenantEvidenceDigestReadbackFixture(inputPath string) (AtlasCovenantEvidenceDigestReadbackFixture, error) {
	input, err := LoadJSON[AtlasCovenantEvidenceDigestReadbackInput](inputPath)
	if err != nil {
		return AtlasCovenantEvidenceDigestReadbackFixture{}, err
	}
	if err := ValidateAtlasCovenantEvidenceDigestReadbackInput(input); err != nil {
		return AtlasCovenantEvidenceDigestReadbackFixture{}, err
	}
	fixture := summarizeCovenantEvidenceDigestReadback(input.Cases)
	fixture.Schema = AtlasCovenantEvidenceDigestReadbackFixtureContract
	fixture.Status = "covenant_digest_readback_recorded"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasCovenantEvidenceDigestReadbackFixture(fixture); err != nil {
		return AtlasCovenantEvidenceDigestReadbackFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCovenantEvidenceDigestReadbackInput(input AtlasCovenantEvidenceDigestReadbackInput) error {
	var errs []string
	requireContract(&errs, "covenant_evidence_digest_readback_input", input.Schema, AtlasCovenantEvidenceDigestReadbackInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateCovenantEvidenceDigestReadbackCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCovenantEvidenceDigestReadbackFixture(fixture AtlasCovenantEvidenceDigestReadbackFixture) error {
	var errs []string
	requireContract(&errs, "covenant_evidence_digest_readback_fixture", fixture.Schema, AtlasCovenantEvidenceDigestReadbackFixtureContract)
	if fixture.Status != "covenant_digest_readback_recorded" {
		errs = append(errs, "status must be covenant_digest_readback_recorded")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", fixture.SourceInputDigest)
	validateCovenantEvidenceDigestReadbackCases(&errs, fixture.Cases)
	expected := summarizeCovenantEvidenceDigestReadback(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases")
	}
	if fixture.DigestReadbackComplete != expected.DigestReadbackComplete {
		errs = append(errs, "digest_readback_complete must match cases")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateCovenantEvidenceDigestReadbackCases(errs *[]string, cases []AtlasCovenantEvidenceDigestReadbackCase) {
	if len(cases) == 0 {
		*errs = append(*errs, "cases must not be empty")
	}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		requireField(errs, prefix+".covenant_readback_path", item.CovenantReadbackPath)
		checkPublicPath(errs, prefix+".covenant_readback_path", item.CovenantReadbackPath, true)
		validateRejectedTicketDigest(errs, prefix+".policy_sha256", item.PolicySHA256)
		validateRejectedTicketDigest(errs, prefix+".ticket_sha256", item.TicketSHA256)
		validateRejectedTicketDigest(errs, prefix+".decision_sha256", item.DecisionSHA256)
		if item.Decision != "accepted" && item.Decision != "rejected" {
			*errs = append(*errs, prefix+".decision must be accepted or rejected")
		}
	}
}

func summarizeCovenantEvidenceDigestReadback(cases []AtlasCovenantEvidenceDigestReadbackCase) AtlasCovenantEvidenceDigestReadbackFixture {
	fixture := AtlasCovenantEvidenceDigestReadbackFixture{
		CaseCount:              len(cases),
		DigestReadbackComplete: true,
		Cases:                  append([]AtlasCovenantEvidenceDigestReadbackCase(nil), cases...),
	}
	for _, item := range cases {
		if item.PolicySHA256 == "" || item.TicketSHA256 == "" || item.DecisionSHA256 == "" {
			fixture.DigestReadbackComplete = false
			break
		}
	}
	return fixture
}
