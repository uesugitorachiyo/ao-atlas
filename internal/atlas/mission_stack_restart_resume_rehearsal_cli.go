package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsStackRestartResumeRehearsal(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations stack-restart-resume-rehearsal", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "stack restart resume rehearsal output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasStackRestartResumeRehearsal()
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
	fmt.Fprintf(stdout, "status=%s\ncomponent_count=%d\nmission_checkpoint_bound=%t\natlas_workgraph_bound=%t\nfoundry_safe_next_work_bound=%t\nfinal_response_allowed=%t\nstack_restart_resume_rehearsal=%s\n",
		fixture.Status,
		fixture.ComponentCount,
		fixture.MissionCheckpointBound,
		fixture.AtlasWorkgraphBound,
		fixture.FoundrySafeNextWorkBound,
		fixture.FinalResponseAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}
