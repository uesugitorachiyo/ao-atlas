package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsExactlyOnceResumeAccountingFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations exactly-once-resume-accounting-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "exactly-once resume accounting fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasExactlyOnceResumeAccountingFixture()
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
	fmt.Fprintf(stdout, "status=%s\nscenario_count=%d\nexactly_once_node_accounting=%t\nduplicate_handoff_double_count_allowed=%t\nexactly_once_resume_accounting_fixture=%s\n",
		fixture.Status,
		fixture.ScenarioCount,
		fixture.ExactlyOnceNodeAccounting,
		fixture.DuplicateHandoffDoubleCountAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}
