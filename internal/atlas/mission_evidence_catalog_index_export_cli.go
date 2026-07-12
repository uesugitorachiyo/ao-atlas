package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsEvidenceCatalogIndexExport(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations evidence-catalog-index-export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "evidence catalog index export output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasEvidenceCatalogIndexExport()
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
	fmt.Fprintf(stdout, "status=%s\nindex_entry_count=%d\nbulk_campaign_artifacts_cataloged=%t\nsource_artifacts_retained=%t\nuploads_artifacts=%t\ndeletes_source_artifacts=%t\nevidence_catalog_index_export=%s\n",
		fixture.Status,
		fixture.IndexEntryCount,
		fixture.BulkCampaignArtifactsCataloged,
		fixture.SourceArtifactsRetained,
		fixture.UploadsArtifacts,
		fixture.DeletesSourceArtifacts,
		filepath.ToSlash(*outPath),
	)
	return nil
}
