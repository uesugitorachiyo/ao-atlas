package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsForgeGoalRunEvidenceFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations forge-goalrun-evidence-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Forge GoalRun evidence fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasForgeGoalRunEvidenceFixture()
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
	fmt.Fprintf(stdout, "status=%s\ngoalrun_start_required=%t\nprovider_execution_allowed=%t\nterminal_receipt_required=%t\nforge_goalrun_evidence_fixture=%s\n",
		fixture.Status,
		fixture.GoalRunStartRequired,
		fixture.ProviderExecutionAllowed,
		fixture.TerminalReceiptRequired,
		filepath.ToSlash(*outPath),
	)
	return nil
}
