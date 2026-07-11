package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3RealRunAcceptance(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-real-run-acceptance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "real-run acceptance node id")
	readinessMatrixPath := fs.String("readiness-matrix", "", "golden-path readiness matrix path")
	nonAOReplayPath := fs.String("non-ao-replay", "", "Month 3 non-AO dry-run replay binding path")
	outPath := fs.String("out", "", "real-run acceptance criteria output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--readiness-matrix": *readinessMatrixPath,
		"--non-ao-replay":    *nonAOReplayPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*readinessMatrixPath, *nonAOReplayPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	criteria, err := BuildAtlasMonth3RealRunAcceptanceCriteria(*nodeID, *readinessMatrixPath, *nonAOReplayPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3RealRunAcceptanceCriteria(*outPath, criteria); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, criteria)
	}
	fmt.Fprintf(stdout, "status=%s\nexternal_repo_count=%d\ncriteria_per_repo=%d\nnon_ao_replay_bound=%t\npromotion_requested=%t\nexecutes_work=%t\nrsi_remains_denied=%t\nmonth3_real_run_acceptance=%s\n",
		criteria.Status,
		criteria.ExternalRepoCount,
		criteria.CriteriaPerRepo,
		criteria.NonAOReplayBound,
		criteria.PromotionRequested,
		criteria.ExecutesWork,
		criteria.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
