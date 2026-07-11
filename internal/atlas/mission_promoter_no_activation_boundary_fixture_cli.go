package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsPromoterNoActivationBoundaryFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations promoter-no-activation-boundary-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "Promoter no-activation boundary fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasPromoterNoActivationBoundaryFixture()
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
	fmt.Fprintf(stdout, "status=%s\nno_promotion_decision_supported=%t\nactivation_execution_owned=%t\nrelease_execution_owned=%t\npromoter_no_activation_boundary_fixture=%s\n",
		fixture.Status,
		fixture.NoPromotionDecisionSupported,
		fixture.ActivationExecutionOwned,
		fixture.ReleaseExecutionOwned,
		filepath.ToSlash(*outPath),
	)
	return nil
}
