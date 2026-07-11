package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsGoldenPathReadinessMatrix(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations golden-path-readiness-matrix", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "golden path readiness matrix output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	matrix, err := BuildAtlasGoldenPathReadinessMatrix()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, matrix); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, matrix)
	}
	fmt.Fprintf(stdout, "status=%s\ncompleted_nodes=%d\nranked_recommendations=%d\npromotion_requested=%t\ngolden_path_readiness_matrix=%s\n",
		matrix.Status,
		matrix.CompletedNodes,
		matrix.RankedRecommendationCount,
		matrix.PromotionRequested,
		filepath.ToSlash(*outPath),
	)
	return nil
}
