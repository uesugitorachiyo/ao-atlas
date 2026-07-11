package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3ProviderModelProvenance(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-provider-model-provenance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "provider model provenance node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "provider model provenance output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceReadbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasMonth3ProviderModelProvenance(*nodeID, *sourceReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3ProviderModelProvenance(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nrun_record_count=%d\nevery_run_has_provider=%t\nevery_run_has_model=%t\nlive_provider_call_count=%d\nrsi_remains_denied=%t\nmonth3_provider_model_provenance=%s\n",
		fixture.Status,
		fixture.RunRecordCount,
		fixture.EveryRunHasProvider,
		fixture.EveryRunHasModel,
		fixture.LiveProviderCallCount,
		fixture.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
