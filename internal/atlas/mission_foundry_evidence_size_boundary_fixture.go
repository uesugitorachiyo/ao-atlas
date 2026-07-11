package atlas

import "fmt"

func BuildAtlasFoundryEvidenceSizeBoundaryFixture() (AtlasFoundryEvidenceSizeBoundaryFixture, error) {
	refs := []AtlasFoundryEvidenceSizeBoundaryReference{
		{Name: "implementation_state", Boundary: "repository_small_fixture", MaxBytesClass: "small"},
		{Name: "generated_campaign_bulk", Boundary: "content_addressed_external", MaxBytesClass: "bulk"},
	}
	fixture := AtlasFoundryEvidenceSizeBoundaryFixture{
		Schema:                        AtlasFoundryEvidenceSizeBoundaryFixtureContract,
		Status:                        "foundry_evidence_size_boundary_ready",
		EvidenceReferences:            refs,
		EvidenceReferenceCount:        len(refs),
		ImplementationStateSeparate:   true,
		GeneratedCampaignBulkSeparate: true,
		SizeChecksRequired:            true,
		SchedulesWork:                 false,
		ExecutesWork:                  false,
		ApprovesWork:                  false,
		ClaimsAuthorityAdvance:        false,
		RSIRemainsDenied:              true,
	}
	if err := ValidateAtlasFoundryEvidenceSizeBoundaryFixture(fixture); err != nil {
		return AtlasFoundryEvidenceSizeBoundaryFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasFoundryEvidenceSizeBoundaryFixture(fixture AtlasFoundryEvidenceSizeBoundaryFixture) error {
	var errs []string
	requireContract(&errs, "foundry_evidence_size_boundary_fixture", fixture.Schema, AtlasFoundryEvidenceSizeBoundaryFixtureContract)
	if fixture.Status != "foundry_evidence_size_boundary_ready" {
		errs = append(errs, "status must be foundry_evidence_size_boundary_ready")
	}
	if fixture.EvidenceReferenceCount != len(fixture.EvidenceReferences) {
		errs = append(errs, "evidence_reference_count must match evidence_references")
	}
	if fixture.EvidenceReferenceCount != 2 {
		errs = append(errs, "evidence_reference_count must be 2")
	}
	if !fixture.ImplementationStateSeparate {
		errs = append(errs, "implementation_state_separate must be true")
	}
	if !fixture.GeneratedCampaignBulkSeparate {
		errs = append(errs, "generated_campaign_bulk_separate must be true")
	}
	if !fixture.SizeChecksRequired {
		errs = append(errs, "size_checks_required must be true")
	}
	for i, ref := range fixture.EvidenceReferences {
		prefix := fmt.Sprintf("evidence_references[%d]", i)
		requireField(&errs, prefix+".name", ref.Name)
		requireField(&errs, prefix+".boundary", ref.Boundary)
		requireField(&errs, prefix+".max_bytes_class", ref.MaxBytesClass)
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
