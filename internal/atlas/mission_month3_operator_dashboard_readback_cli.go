package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3OperatorDashboardReadback(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-operator-dashboard-readback", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "operator dashboard readback node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	ciMatrixPath := fs.String("ci-matrix", "", "cross-repo CI matrix path")
	outPath := fs.String("out", "", "operator dashboard readback output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":         *nodeID,
		"--source-readback": *sourceReadbackPath,
		"--ci-matrix":       *ciMatrixPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *ciMatrixPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	fixture, err := BuildAtlasMonth3OperatorDashboardReadback(*nodeID, *sourceReadbackPath, *ciMatrixPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3OperatorDashboardReadback(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\ncompleted_nodes=%d\nready_nodes=%d\nblocker_count=%d\nci_matrix_bound=%t\nfinal_response_allowed=%t\nrsi_remains_denied=%t\nmonth3_operator_dashboard_readback=%s\n",
		fixture.Status,
		fixture.CompletedNodes,
		fixture.ReadyNodes,
		fixture.BlockerCount,
		fixture.CIMatrixBound,
		fixture.FinalResponseAllowed,
		fixture.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
