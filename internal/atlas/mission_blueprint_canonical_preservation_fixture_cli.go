package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsBlueprintCanonicalPreservationFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations blueprint-canonical-preservation-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Blueprint canonical preservation fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasBlueprintCanonicalPreservationFixture()
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
	fmt.Fprintf(stdout, "status=%s\nworkspace_root_ref=%s\ndigest_preserved=%t\ncanonical_file_count=%d\nblueprint_canonical_preservation_fixture=%s\n",
		fixture.Status,
		fixture.WorkspaceRootRef,
		fixture.DigestPreserved,
		fixture.CanonicalFileCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
