package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsRollbackTerminalReadbackFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations rollback-terminal-readback-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "rollback terminal readback fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasRollbackTerminalReadbackFixture()
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
	fmt.Fprintf(stdout, "status=%s\nrollback_receipt_replayed=%t\nreadback_agreement_count=%d\nterminal_state=%s\nrollback_terminal_readback_fixture=%s\n",
		fixture.Status,
		fixture.RollbackReceiptReplayed,
		fixture.ReadbackAgreementCount,
		fixture.TerminalState,
		filepath.ToSlash(*outPath),
	)
	return nil
}
