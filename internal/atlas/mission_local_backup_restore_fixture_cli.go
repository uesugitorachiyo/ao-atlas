package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsLocalBackupRestoreFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations local-backup-restore-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "local backup restore fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasLocalBackupRestoreFixture()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nbackup_target=%s\ndigest_verification_required=%t\nreadback_continuity_required=%t\nlocal_backup_restore_fixture=%s\n",
		fixture.Status,
		fixture.BackupTarget,
		fixture.DigestVerificationRequired,
		fixture.ReadbackContinuityRequired,
		filepath.ToSlash(*outPath),
	)
	return nil
}
