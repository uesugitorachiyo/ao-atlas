package atlas

import "fmt"

func BuildAtlasContentAddressedEvidenceManifestFixture() (AtlasContentAddressedEvidenceManifestFixture, error) {
	entries := []AtlasContentAddressedEvidenceManifestEntry{
		{Name: "bulk_campaign_evidence", Boundary: "external_content_addressed_manifest", Addressing: "sha256"},
		{Name: "small_replayable_fixture", Boundary: "repository_fixture", Addressing: "git"},
	}
	fixture := AtlasContentAddressedEvidenceManifestFixture{
		Schema:                          AtlasContentAddressedEvidenceManifestFixtureContract,
		Status:                          "content_addressed_evidence_manifest_ready",
		ManifestEntries:                 entries,
		ManifestEntryCount:              len(entries),
		BulkEvidenceExternalized:        true,
		SmallReplayableFixturesRetained: true,
		ContentAddressingRequired:       true,
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasContentAddressedEvidenceManifestFixture(fixture); err != nil {
		return AtlasContentAddressedEvidenceManifestFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasContentAddressedEvidenceManifestFixture(fixture AtlasContentAddressedEvidenceManifestFixture) error {
	var errs []string
	requireContract(&errs, "content_addressed_evidence_manifest_fixture", fixture.Schema, AtlasContentAddressedEvidenceManifestFixtureContract)
	if fixture.Status != "content_addressed_evidence_manifest_ready" {
		errs = append(errs, "status must be content_addressed_evidence_manifest_ready")
	}
	if fixture.ManifestEntryCount != len(fixture.ManifestEntries) {
		errs = append(errs, "manifest_entry_count must match manifest_entries")
	}
	if fixture.ManifestEntryCount != 2 {
		errs = append(errs, "manifest_entry_count must be 2")
	}
	if !fixture.BulkEvidenceExternalized {
		errs = append(errs, "bulk_evidence_externalized must be true")
	}
	if !fixture.SmallReplayableFixturesRetained {
		errs = append(errs, "small_replayable_fixtures_retained must be true")
	}
	if !fixture.ContentAddressingRequired {
		errs = append(errs, "content_addressing_required must be true")
	}
	for i, entry := range fixture.ManifestEntries {
		prefix := fmt.Sprintf("manifest_entries[%d]", i)
		requireField(&errs, prefix+".name", entry.Name)
		requireField(&errs, prefix+".boundary", entry.Boundary)
		requireField(&errs, prefix+".addressing", entry.Addressing)
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
