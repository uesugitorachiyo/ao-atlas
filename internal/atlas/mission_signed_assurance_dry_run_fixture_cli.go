package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsSignedAssuranceDryRunFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations signed-assurance-dry-run-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "signed assurance dry-run fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasSignedAssuranceDryRunFixture()
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
	fmt.Fprintf(stdout, "status=%s\nrequired_check_count=%d\npromotion_decision_enabled=%t\nsigned_assurance_dry_run_fixture=%s\n",
		fixture.Status,
		fixture.RequiredCheckCount,
		fixture.PromotionDecisionEnabled,
		filepath.ToSlash(*outPath),
	)
	return nil
}
