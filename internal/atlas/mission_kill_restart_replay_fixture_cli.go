package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsKillRestartReplayFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations kill-restart-replay-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "kill restart replay fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasKillRestartReplayFixture()
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
	fmt.Fprintf(stdout, "status=%s\nkilled_run_replayed=%t\nno_lost_evidence=%t\nfalse_completion_detected=%t\nkill_restart_replay_fixture=%s\n",
		fixture.Status,
		fixture.KilledRunReplayed,
		fixture.NoLostEvidence,
		fixture.FalseCompletionDetected,
		filepath.ToSlash(*outPath),
	)
	return nil
}
