package atlas

import "fmt"

func BuildAtlasDurableStateMigrationMetadata() (AtlasDurableStateMigrationMetadata, error) {
	migrations := []AtlasDurableStateMigrationStep{
		{
			Version:     1,
			Name:        "initial_durable_state_metadata",
			Description: "Records durable state schema version and fail-closed unknown-version handling.",
			Reversible:  true,
		},
	}
	metadata := AtlasDurableStateMigrationMetadata{
		Schema:                 AtlasDurableStateMigrationMetadataContract,
		Status:                 "durable_state_migration_metadata_ready",
		CurrentVersion:         1,
		MinimumSupportedVersion: 1,
		UnknownVersionHandling: "fail_closed",
		Migrations:             migrations,
		MigrationCount:         len(migrations),
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasDurableStateMigrationMetadata(metadata); err != nil {
		return AtlasDurableStateMigrationMetadata{}, err
	}
	return metadata, nil
}

func ValidateAtlasDurableStateMigrationMetadata(metadata AtlasDurableStateMigrationMetadata) error {
	var errs []string
	requireContract(&errs, "durable_state_migration_metadata", metadata.Schema, AtlasDurableStateMigrationMetadataContract)
	if metadata.Status != "durable_state_migration_metadata_ready" {
		errs = append(errs, "status must be durable_state_migration_metadata_ready")
	}
	if metadata.CurrentVersion <= 0 {
		errs = append(errs, "current_version must be positive")
	}
	if metadata.MinimumSupportedVersion <= 0 {
		errs = append(errs, "minimum_supported_version must be positive")
	}
	if metadata.MinimumSupportedVersion > metadata.CurrentVersion {
		errs = append(errs, "minimum_supported_version must not exceed current_version")
	}
	if metadata.UnknownVersionHandling != "fail_closed" {
		errs = append(errs, "unknown_version_handling must be fail_closed")
	}
	if metadata.MigrationCount != len(metadata.Migrations) {
		errs = append(errs, "migration_count must match migrations")
	}
	if metadata.MigrationCount == 0 {
		errs = append(errs, "migration_count must be greater than zero")
	}
	for i, migration := range metadata.Migrations {
		prefix := fmt.Sprintf("migrations[%d]", i)
		if migration.Version <= 0 {
			errs = append(errs, prefix+".version must be positive")
		}
		requireField(&errs, prefix+".name", migration.Name)
		requireField(&errs, prefix+".description", migration.Description)
		if !migration.Reversible {
			errs = append(errs, prefix+".reversible must be true")
		}
	}
	validateNoAuthorityEffects(&errs, metadata.SchedulesWork, metadata.ExecutesWork, metadata.ApprovesWork, metadata.ClaimsAuthorityAdvance, metadata.RSIRemainsDenied)
	return joinErrors(errs)
}
