package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsNonAOReplayBindingFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations non-ao-replay-binding-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "non-AO replay binding fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasNonAOReplayBindingFixture()
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
	fmt.Fprintf(stdout, "status=%s\nreviewed_pr_evidence=%t\nobserver_readback_bound=%t\npromotion_requested=%t\nnon_ao_replay_binding_fixture=%s\n",
		fixture.Status,
		fixture.ReviewedPREvidence,
		fixture.ObserverReadbackBound,
		fixture.PromotionRequested,
		filepath.ToSlash(*outPath),
	)
	return nil
}
