package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3TerminalDigestBinding(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-terminal-digest-binding", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "terminal digest binding node id")
	sourceReadbackPath := fs.String("source-readback", "", "terminal source recommendation readback path")
	matrixPath := fs.String("readiness-matrix", "", "terminal golden-path readiness matrix path")
	outPath := fs.String("out", "", "terminal digest binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--readiness-matrix": *matrixPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *matrixPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasMonth3TerminalDigestBinding(*nodeID, *sourceReadbackPath, *matrixPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3TerminalDigestBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nreadback_completed_nodes=%d\nmatrix_completed_nodes=%d\nnode_counts_match=%t\nfinal_response_allowed=%t\npromotion_requested=%t\nrsi_remains_denied=%t\nmonth3_terminal_digest_binding=%s\n",
		binding.Status,
		binding.ReadbackCompletedNodes,
		binding.MatrixCompletedNodes,
		binding.NodeCountsMatch,
		binding.FinalResponseAllowed,
		binding.PromotionRequested,
		binding.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
