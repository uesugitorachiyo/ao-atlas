package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3FoundrySafeNextWork(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-foundry-safe-next-work", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "recommendation node id")
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	sourceWorkgraph := fs.String("source-workgraph", "", "source workgraph path")
	expectedSelectedNode := fs.String("expected-selected-node", "", "expected selected node")
	expectedNextNode := fs.String("expected-next-node-after-completion", "", "expected next node after this node completes")
	outPath := fs.String("out", "", "safe-next-work fixture output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":                             *nodeID,
		"--source-readback":                     *sourceReadback,
		"--source-workgraph":                    *sourceWorkgraph,
		"--expected-selected-node":              *expectedSelectedNode,
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
	fixture, err := BuildAtlasMonth3FoundrySafeNextWorkFixture(*nodeID, *sourceReadback, *sourceWorkgraph, *expectedSelectedNode, *expectedNextNode, readback, workgraph)
	if err != nil {
		return err
	}
	if err := WriteAtlasMonth3FoundrySafeNextWorkFixture(*outPath, fixture); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "status=%s\nselected_node=%s\nsingle_active_task=%t\nterminal_readiness_bound=%t\nexecutes_work=%t\nfoundry_safe_next_work_fixture=%s\n",
		fixture.Status,
		fixture.SelectedNode,
		fixture.SingleActiveTask,
		fixture.TerminalReadinessBound,
		fixture.ExecutesWork,
		filepath.ToSlash(*outPath),
	)
	return nil
}
