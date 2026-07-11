package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCommandCovenantFieldParityFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-covenant-field-parity-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Command/Covenant field parity fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasCommandCovenantFieldParityFixture()
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
	fmt.Fprintf(stdout, "status=%s\npolicy_field_count=%d\napproval_field_count=%d\nrejected_extra_field_count=%d\ncommand_covenant_field_parity_fixture=%s\n",
		fixture.Status,
		fixture.PolicyFieldCount,
		fixture.ApprovalFieldCount,
		fixture.RejectedExtraFieldCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
