package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3ArchitectureSourceTruth(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-architecture-source-truth", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "recommendation node id")
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "architecture source-truth checklist output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":         *nodeID,
		"--source-readback": *sourceReadback,
		"--out":             *outPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if samePath(*sourceReadback, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](*sourceReadback)
	if err != nil {
		return err
	}
	checklist, err := BuildAtlasMonth3ArchitectureSourceTruthChecklist(*nodeID, *sourceReadback, readback)
	if err != nil {
		return err
	}
	if err := WriteAtlasMonth3ArchitectureSourceTruthChecklist(*outPath, checklist); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes=%d\nready_nodes=%d\ncorrections=%d\narchitecture_source_truth_checklist=%s\n",
		checklist.Status,
		checklist.NodeID,
		checklist.CompletedNodes,
		checklist.ReadyNodes,
		len(checklist.Checklist),
		filepath.ToSlash(*outPath),
	)
	return nil
}
