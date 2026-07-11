package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3RestartResumeSoak(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-restart-resume-soak", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "restart resume soak node id")
	exactlyOncePath := fs.String("exactly-once", "", "exactly-once resume accounting fixture path")
	killRestartPath := fs.String("kill-restart", "", "kill restart replay fixture path")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	dashboardReadbackPath := fs.String("dashboard-readback", "", "operator dashboard readback path")
	outPath := fs.String("out", "", "restart resume soak output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":            *nodeID,
		"--exactly-once":       *exactlyOncePath,
		"--kill-restart":       *killRestartPath,
		"--source-readback":    *sourceReadbackPath,
		"--dashboard-readback": *dashboardReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*exactlyOncePath, *killRestartPath, *sourceReadbackPath, *dashboardReadbackPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	fixture, err := BuildAtlasMonth3RestartResumeSoak(*nodeID, *exactlyOncePath, *killRestartPath, *sourceReadbackPath, *dashboardReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3RestartResumeSoak(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nscenario_count=%d\ncheckpoint_recovery_bound=%t\nno_lost_evidence=%t\nfinal_response_allowed=%t\nrsi_remains_denied=%t\nmonth3_restart_resume_soak=%s\n",
		fixture.Status,
		fixture.ScenarioCount,
		fixture.CheckpointRecoveryBound,
		fixture.NoLostEvidence,
		fixture.FinalResponseAllowed,
		fixture.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
