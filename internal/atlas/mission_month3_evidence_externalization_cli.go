package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3EvidenceExternalization(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-evidence-externalization", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "evidence externalization node id")
	contentManifestPath := fs.String("content-manifest", "", "content-addressed evidence manifest fixture path")
	outPath := fs.String("out", "", "evidence externalization plan output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--content-manifest": *contentManifestPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*contentManifestPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	plan, err := BuildAtlasMonth3EvidenceExternalizationPlan(*nodeID, *contentManifestPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3EvidenceExternalizationPlan(*outPath, plan); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, plan)
	}
	fmt.Fprintf(stdout, "status=%s\nexternalized_class_count=%d\nretained_fixture_class_count=%d\ncontent_manifest_bound=%t\nexecutes_work=%t\nrsi_remains_denied=%t\nmonth3_evidence_externalization=%s\n",
		plan.Status,
		plan.ExternalizedClassCount,
		plan.RetainedFixtureClassCount,
		plan.ContentManifestBound,
		plan.ExecutesWork,
		plan.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
