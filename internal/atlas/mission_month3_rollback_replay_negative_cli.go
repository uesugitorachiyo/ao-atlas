package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3RollbackReplayNegative(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-rollback-replay-negative", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "rollback replay negative node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "rollback replay negative output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceReadbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasMonth3RollbackReplayNegative(*nodeID, *sourceReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3RollbackReplayNegative(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\ncase_count=%d\naccepted_case_count=%d\nstale_base_commit_rejected=%t\nreceipt_digest_mismatch_rejected=%t\nrsi_remains_denied=%t\nmonth3_rollback_replay_negative=%s\n",
		fixture.Status,
		fixture.CaseCount,
		fixture.AcceptedCaseCount,
		fixture.StaleBaseCommitRejected,
		fixture.ReceiptDigestMismatchRejected,
		fixture.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
