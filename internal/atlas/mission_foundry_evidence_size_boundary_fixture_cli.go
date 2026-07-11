package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsFoundryEvidenceSizeBoundaryFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-evidence-size-boundary-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "foundry evidence size boundary fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasFoundryEvidenceSizeBoundaryFixture()
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
	fmt.Fprintf(stdout, "status=%s\nevidence_reference_count=%d\nimplementation_state_separate=%t\ngenerated_campaign_bulk_separate=%t\nfoundry_evidence_size_boundary_fixture=%s\n",
		fixture.Status,
		fixture.EvidenceReferenceCount,
		fixture.ImplementationStateSeparate,
		fixture.GeneratedCampaignBulkSeparate,
		filepath.ToSlash(*outPath),
	)
	return nil
}
