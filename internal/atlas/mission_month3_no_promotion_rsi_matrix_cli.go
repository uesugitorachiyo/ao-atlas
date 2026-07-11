package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3NoPromotionRSIMatrix(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-no-promotion-rsi-matrix", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "recommendation node id")
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	sourceWorkgraph := fs.String("source-workgraph", "", "source workgraph path")
	evidenceRoot := fs.String("evidence-root", "", "evidence root")
	expectedNextNode := fs.String("expected-next-node-after-completion", "", "expected next node after this node completes")
	outPath := fs.String("out", "", "no-promotion RSI matrix output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":                             *nodeID,
		"--source-readback":                     *sourceReadback,
		"--source-workgraph":                    *sourceWorkgraph,
		"--evidence-root":                       *evidenceRoot,
		"--expected-next-node-after-completion": *expectedNextNode,
		"--out":                                 *outPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if samePath(*sourceReadback, *outPath) || samePath(*sourceWorkgraph, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](*sourceReadback)
	if err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*sourceWorkgraph)
	if err != nil {
		return err
	}
	matrix, err := BuildAtlasMonth3NoPromotionRSIMatrix(*nodeID, *sourceReadback, *sourceWorkgraph, *evidenceRoot, *expectedNextNode, readback, workgraph)
	if err != nil {
		return err
	}
	if err := WriteAtlasMonth3NoPromotionRSIMatrix(*outPath, matrix); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes=%d\npromoter_no_promotion_files=%d\nno_promotion_invariant_holds=%t\nrsi_denial_invariant_holds=%t\nmonth3_no_promotion_rsi_matrix=%s\n",
		matrix.Status,
		matrix.NodeID,
		matrix.CompletedNodes,
		matrix.PromoterNoPromotionFiles,
		matrix.NoPromotionInvariantHolds,
		matrix.RSIDenialInvariantHolds,
		filepath.ToSlash(*outPath),
	)
	return nil
}
