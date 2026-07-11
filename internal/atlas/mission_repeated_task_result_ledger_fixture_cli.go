package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsRepeatedTaskResultLedgerFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations repeated-task-result-ledger-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "repeated task result ledger fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasRepeatedTaskResultLedgerFixture()
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
	fmt.Fprintf(stdout, "status=%s\nattempt_count=%d\nreplayable_result_ledger=%t\nlive_provider_calls=%t\nrepeated_task_result_ledger_fixture=%s\n",
		fixture.Status,
		fixture.AttemptCount,
		fixture.ReplayableResultLedger,
		fixture.LiveProviderCalls,
		filepath.ToSlash(*outPath),
	)
	return nil
}
