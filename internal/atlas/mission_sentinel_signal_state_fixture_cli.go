package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsSentinelSignalStateFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations sentinel-signal-state-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Sentinel signal state fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasSentinelSignalStateFixture()
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
	fmt.Fprintf(stdout, "status=%s\nsignal_count=%d\nstate_count=%d\nsentinel_signal_state_fixture=%s\n",
		fixture.Status,
		fixture.SignalCount,
		fixture.StateCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
