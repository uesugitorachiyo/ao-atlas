package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsReplayableStatePacketFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations replayable-state-packet-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "replayable state packet fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasReplayableStatePacketFixture()
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
	fmt.Fprintf(stdout, "status=%s\nstate_count=%d\nhandoff_counts_as_completed=%t\nreplayable=%t\nreplayable_state_packet_fixture=%s\n",
		fixture.Status,
		fixture.StateCount,
		fixture.HandoffCountsAsCompleted,
		fixture.Replayable,
		filepath.ToSlash(*outPath),
	)
	return nil
}
