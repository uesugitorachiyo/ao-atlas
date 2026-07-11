package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3CrossRepoCIMatrix(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-cross-repo-ci-matrix", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "cross-repo CI matrix node id")
	sentinelSignalStatePath := fs.String("sentinel-signal-state", "", "Sentinel signal state fixture path")
	outPath := fs.String("out", "", "cross-repo CI matrix output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":               *nodeID,
		"--sentinel-signal-state": *sentinelSignalStatePath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sentinelSignalStatePath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	matrix, err := BuildAtlasMonth3CrossRepoCIMatrix(*nodeID, *sentinelSignalStatePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3CrossRepoCIMatrix(*outPath, matrix); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, matrix)
	}
	fmt.Fprintf(stdout, "status=%s\nrepo_count=%d\nmatrix_entry_count=%d\nsentinel_signal_state_bound=%t\nrequires_pass_before_merge=%t\nexecutes_work=%t\nrsi_remains_denied=%t\nmonth3_cross_repo_ci_matrix=%s\n",
		matrix.Status,
		matrix.RepoCount,
		matrix.MatrixEntryCount,
		matrix.SentinelSignalStateBound,
		matrix.RequiresPassBeforeMerge,
		matrix.ExecutesWork,
		matrix.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
