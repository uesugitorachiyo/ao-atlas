package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCommandReadbackAdapterBoundaryFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-readback-adapter-boundary-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "command readback adapter boundary fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasCommandReadbackAdapterBoundaryFixture()
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
	fmt.Fprintf(stdout, "status=%s\nadapter_count=%d\nduplicates_domain_decisions=%t\ndelegates_domain_decisions=%t\ncommand_readback_adapter_boundary_fixture=%s\n",
		fixture.Status,
		fixture.AdapterCount,
		fixture.DuplicatesDomainDecisions,
		fixture.DelegatesDomainDecisions,
		filepath.ToSlash(*outPath),
	)
	return nil
}
