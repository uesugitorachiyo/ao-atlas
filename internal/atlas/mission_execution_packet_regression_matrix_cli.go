package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsExecutionPacketRegressionMatrix(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations execution-packet-regression-matrix", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "execution packet regression matrix output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	matrix, err := BuildAtlasExecutionPacketRegressionMatrix()
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
	fmt.Fprintf(stdout, "status=%s\ncases=%d\nprovider_invocation_allowed=%t\nsilent_changed_result_allowed=%t\nexecution_packet_regression_matrix=%s\n",
		matrix.Status,
		matrix.CaseCount,
		matrix.ProviderInvocationAllowed,
		matrix.SilentChangedResultAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}
