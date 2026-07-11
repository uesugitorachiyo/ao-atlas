package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsSentinelHostedCIWorkflowFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations sentinel-hosted-ci-workflow-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Sentinel hosted CI workflow fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasSentinelHostedCIWorkflowFixture()
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
	fmt.Fprintf(stdout, "status=%s\npermissions=%s\ndeterministic_fixture_commands=%d\nsentinel_hosted_ci_workflow_fixture=%s\n",
		fixture.Status,
		fixture.Permissions,
		fixture.CommandCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
