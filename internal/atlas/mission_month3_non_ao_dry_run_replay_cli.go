package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3NonAODryRunReplay(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-non-ao-dry-run-replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "non-AO dry-run replay node id")
	sourceFixturePath := fs.String("source-fixture", "", "source non-AO replay fixture path")
	terminalBindingPath := fs.String("terminal-binding", "", "terminal digest binding path")
	outPath := fs.String("out", "", "non-AO dry-run replay output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-fixture":   *sourceFixturePath,
		"--terminal-binding": *terminalBindingPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceFixturePath, *terminalBindingPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasMonth3NonAODryRunReplayBinding(*nodeID, *sourceFixturePath, *terminalBindingPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3NonAODryRunReplayBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nreplay_repo=%s\nfixture_only_execution_evidence=%t\nterminal_digest_binding_bound=%t\npromotion_requested=%t\nlive_provider_calls=%t\nrsi_remains_denied=%t\nmonth3_non_ao_dry_run_replay=%s\n",
		binding.Status,
		binding.ReplayRepo,
		binding.FixtureOnlyExecutionEvidence,
		binding.TerminalDigestBindingBound,
		binding.PromotionRequested,
		binding.LiveProviderCalls,
		binding.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
