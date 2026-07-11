package atlas

import "fmt"

func BuildAtlasLocalBackupRestoreFixture() (AtlasLocalBackupRestoreFixture, error) {
	artifacts := []string{
		"mission_event_ledger",
		"workgraph_checkpoint",
		"recommendation_readback",
		"run_link_digest",
	}
	fixture := AtlasLocalBackupRestoreFixture{
		Schema:                     AtlasLocalBackupRestoreFixtureContract,
		Status:                     "local_backup_restore_ready",
		BackupTarget:               "local_filesystem",
		RestoreSource:              "local_filesystem",
		RestoredArtifacts:          artifacts,
		RestoredArtifactCount:      len(artifacts),
		DigestVerificationRequired: true,
		ReadbackContinuityRequired: true,
		ExternalStorageRequired:    false,
		CredentialsRequired:        false,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           true,
	}
	if err := ValidateAtlasLocalBackupRestoreFixture(fixture); err != nil {
		return AtlasLocalBackupRestoreFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasLocalBackupRestoreFixture(fixture AtlasLocalBackupRestoreFixture) error {
	var errs []string
	requireContract(&errs, "local_backup_restore_fixture", fixture.Schema, AtlasLocalBackupRestoreFixtureContract)
	if fixture.Status != "local_backup_restore_ready" {
		errs = append(errs, "status must be local_backup_restore_ready")
	}
	if fixture.BackupTarget != "local_filesystem" {
		errs = append(errs, "backup_target must be local_filesystem")
	}
	if fixture.RestoreSource != "local_filesystem" {
		errs = append(errs, "restore_source must be local_filesystem")
	}
	if fixture.RestoredArtifactCount != len(fixture.RestoredArtifacts) {
		errs = append(errs, "restored_artifact_count must match restored_artifacts")
	}
	if fixture.RestoredArtifactCount != 4 {
		errs = append(errs, "restored_artifact_count must be 4")
	}
	for i, artifact := range fixture.RestoredArtifacts {
		requireField(&errs, fmt.Sprintf("restored_artifacts[%d]", i), artifact)
	}
	if !fixture.DigestVerificationRequired {
		errs = append(errs, "digest_verification_required must be true")
	}
	if !fixture.ReadbackContinuityRequired {
		errs = append(errs, "readback_continuity_required must be true")
	}
	if fixture.ExternalStorageRequired {
		errs = append(errs, "external_storage_required must be false")
	}
	if fixture.CredentialsRequired {
		errs = append(errs, "credentials_required must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
