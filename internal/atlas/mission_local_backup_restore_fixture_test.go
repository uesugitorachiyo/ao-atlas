package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3LocalBackupRestoreFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-28")
	recordedPath := filepath.Join(nodeDir, "local-backup-restore-fixture.json")
	outPath := filepath.Join(t.TempDir(), "local-backup-restore-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "local-backup-restore-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("local-backup-restore-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=local_backup_restore_ready",
		"backup_target=local_filesystem",
		"digest_verification_required=true",
		"readback_continuity_required=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("local backup restore output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("local backup restore fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["digest_verification_required"] != true ||
		generated["readback_continuity_required"] != true ||
		generated["external_storage_required"] != false ||
		generated["credentials_required"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("local backup restore fixture lost safety state: %#v", generated)
	}
}

func TestMonth3LocalBackupRestoreFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-28", "local-backup-restore-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.local-backup-restore-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:local-backup-restore-fixture" {
		t.Fatalf("expected typed local backup restore validator, got %s", validator)
	}
}

func TestMonth3LocalBackupRestoreFixtureRejectsExternalStorage(t *testing.T) {
	fixture, err := BuildAtlasLocalBackupRestoreFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.ExternalStorageRequired = true
	if err := ValidateAtlasLocalBackupRestoreFixture(fixture); err == nil || !strings.Contains(err.Error(), "external_storage_required must be false") {
		t.Fatalf("expected external storage rejection, got %v", err)
	}
}
