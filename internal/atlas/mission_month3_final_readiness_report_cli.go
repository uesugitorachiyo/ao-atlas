package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3FinalReadinessReport(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-final-readiness-report", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "final readiness report node id")
	sourceReadbackPath := fs.String("source-readback", "", "terminal source recommendation readback path")
	matrixPath := fs.String("readiness-matrix", "", "terminal golden-path readiness matrix path")
	closureRollupPath := fs.String("closure-rollup", "", "final closure rollup path")
	closureReadbackPath := fs.String("closure-readback", "", "closure wave readback before final report")
	outPath := fs.String("out", "", "final readiness report output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--readiness-matrix": *matrixPath,
		"--closure-rollup":   *closureRollupPath,
		"--closure-readback": *closureReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *matrixPath, *closureRollupPath, *closureReadbackPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	report, err := BuildAtlasMonth3FinalReadinessReport(*nodeID, *sourceReadbackPath, *matrixPath, *closureRollupPath, *closureReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3FinalReadinessReport(*outPath, report); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, report)
	}
	fmt.Fprintf(stdout, "status=%s\nsource_completed_nodes=%d\nclosure_completed_nodes_before_report=%d\nclosure_next_executable_node=%s\nproven_capability_count=%d\nunresolved_blocker_count=%d\nrecommended_next_action_count=%d\npromotion_requested=%t\nrsi_remains_denied=%t\nmonth3_final_readiness_report=%s\n",
		report.Status,
		report.SourceCompletedNodes,
		report.ClosureCompletedNodesBeforeReport,
		report.ClosureNextExecutableNode,
		report.ProvenCapabilityCount,
		report.UnresolvedBlockerCount,
		report.RecommendedNextActionCount,
		report.PromotionRequested,
		report.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
