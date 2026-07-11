package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsDurableStateMigrationMetadata(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations durable-state-migration-metadata", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "durable state migration metadata output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	metadata, err := BuildAtlasDurableStateMigrationMetadata()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, metadata); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, metadata)
	}
	fmt.Fprintf(stdout, "status=%s\ncurrent_version=%d\nunknown_version_handling=%s\nmigration_count=%d\ndurable_state_migration_metadata=%s\n",
		metadata.Status,
		metadata.CurrentVersion,
		metadata.UnknownVersionHandling,
		metadata.MigrationCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
