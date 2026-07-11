package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsBoundedExecutionPacketFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations bounded-execution-packet-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "bounded execution packet fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasBoundedExecutionPacketFixture()
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
	fmt.Fprintf(stdout, "status=%s\nisolated_worktree_required=%t\nexact_digest_approval_required=%t\nrollback_receipt_required=%t\nbounded_execution_packet_fixture=%s\n",
		fixture.Status,
		fixture.IsolatedWorktreeRequired,
		fixture.ExactDigestApprovalRequired,
		fixture.RollbackReceiptRequired,
		filepath.ToSlash(*outPath),
	)
	return nil
}
