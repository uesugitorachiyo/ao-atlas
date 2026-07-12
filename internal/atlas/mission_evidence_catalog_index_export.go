package atlas

import "fmt"

func BuildAtlasEvidenceCatalogIndexExport() (AtlasEvidenceCatalogIndexExport, error) {
	entries := []AtlasEvidenceCatalogIndexEntry{
		{
			Name:          "final_closure_consolidation_wave",
			EvidenceRoot:  "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01",
			ArtifactClass: "bulk_campaign_output",
			DigestMode:    "directory_digest_required_before_externalization",
			IndexOnly:     true,
		},
		{
			Name:          "long_run_hardening_wave",
			EvidenceRoot:  "docs/evidence/ao-atlas-long-run-hardening-wave-v01",
			ArtifactClass: "bulk_campaign_output",
			DigestMode:    "directory_digest_required_before_externalization",
			IndexOnly:     true,
		},
		{
			Name:          "month5_beta_operations",
			EvidenceRoot:  "docs/evidence/ao-stack-month5-beta-operations-v01",
			ArtifactClass: "bulk_campaign_output",
			DigestMode:    "directory_digest_required_before_externalization",
			IndexOnly:     true,
		},
	}
	fixture := AtlasEvidenceCatalogIndexExport{
		Schema:                          AtlasEvidenceCatalogIndexExportContract,
		Status:                          "evidence_catalog_index_export_ready",
		Wave:                            "ao-stack-month6-recommendations",
		IndexEntries:                    entries,
		IndexEntryCount:                 len(entries),
		BulkCampaignArtifactsCataloged:  true,
		SourceArtifactsRetained:         true,
		UploadsArtifacts:                false,
		DeletesSourceArtifacts:          false,
		ContentAddressedExportRequired:  true,
		SmallReplayableFixturesRetained: true,
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasEvidenceCatalogIndexExport(fixture); err != nil {
		return AtlasEvidenceCatalogIndexExport{}, err
	}
	return fixture, nil
}

func ValidateAtlasEvidenceCatalogIndexExport(fixture AtlasEvidenceCatalogIndexExport) error {
	var errs []string
	requireContract(&errs, "evidence_catalog_index_export", fixture.Schema, AtlasEvidenceCatalogIndexExportContract)
	if fixture.Status != "evidence_catalog_index_export_ready" {
		errs = append(errs, "status must be evidence_catalog_index_export_ready")
	}
	requireField(&errs, "wave", fixture.Wave)
	if fixture.IndexEntryCount != len(fixture.IndexEntries) {
		errs = append(errs, "index_entry_count must match index_entries")
	}
	if fixture.IndexEntryCount < 3 {
		errs = append(errs, "index_entry_count must be at least 3")
	}
	if !fixture.BulkCampaignArtifactsCataloged {
		errs = append(errs, "bulk_campaign_artifacts_cataloged must be true")
	}
	if !fixture.SourceArtifactsRetained {
		errs = append(errs, "source_artifacts_retained must be true")
	}
	if fixture.UploadsArtifacts {
		errs = append(errs, "uploads_artifacts must be false")
	}
	if fixture.DeletesSourceArtifacts {
		errs = append(errs, "deletes_source_artifacts must be false")
	}
	if !fixture.ContentAddressedExportRequired {
		errs = append(errs, "content_addressed_export_required must be true")
	}
	if !fixture.SmallReplayableFixturesRetained {
		errs = append(errs, "small_replayable_fixtures_retained must be true")
	}
	for i, entry := range fixture.IndexEntries {
		prefix := fmt.Sprintf("index_entries[%d]", i)
		requireField(&errs, prefix+".name", entry.Name)
		requireField(&errs, prefix+".evidence_root", entry.EvidenceRoot)
		requireField(&errs, prefix+".artifact_class", entry.ArtifactClass)
		requireField(&errs, prefix+".digest_mode", entry.DigestMode)
		if !entry.IndexOnly {
			errs = append(errs, prefix+".index_only must be true")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
