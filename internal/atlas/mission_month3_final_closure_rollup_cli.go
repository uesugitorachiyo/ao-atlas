package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3FinalClosureRollup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-final-closure-rollup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "final closure rollup node id")
	sourceReadbackPath := fs.String("source-readback", "", "terminal source recommendation readback path")
	matrixPath := fs.String("readiness-matrix", "", "terminal golden-path readiness matrix path")
	promoterPath := fs.String("promoter", "", "terminal promoter no-promotion evidence path")
	commandPath := fs.String("command", "", "terminal command readback evidence path")
	publicSafetyPath := fs.String("public-safety", "", "terminal public-safety scan path")
	outPath := fs.String("out", "", "final closure rollup output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--readiness-matrix": *matrixPath,
		"--promoter":         *promoterPath,
		"--command":          *commandPath,
		"--public-safety":    *publicSafetyPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *matrixPath, *promoterPath, *commandPath, *publicSafetyPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	rollup, err := BuildAtlasMonth3FinalClosureRollup(*nodeID, *sourceReadbackPath, *matrixPath, *promoterPath, *commandPath, *publicSafetyPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3FinalClosureRollup(*outPath, rollup); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, rollup)
	}
	fmt.Fprintf(stdout, "status=%s\ncompleted_nodes=%d\nmatrix_recommendations=%d\npromoter_status=%s\ncommand_status=%s\npublic_safety_status=%s\npromotion_requested=%t\nrsi_remains_denied=%t\nmonth3_final_closure_rollup=%s\n",
		rollup.Status,
		rollup.SourceCompletedNodes,
		rollup.MatrixRecommendationCount,
		rollup.PromoterStatus,
		rollup.CommandStatus,
		rollup.PublicSafetyStatus,
		rollup.PromotionRequested,
		rollup.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
