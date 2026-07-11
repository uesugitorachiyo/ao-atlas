package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsContentAddressedEvidenceManifestFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations content-addressed-evidence-manifest-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "content addressed evidence manifest fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasContentAddressedEvidenceManifestFixture()
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
	fmt.Fprintf(stdout, "status=%s\nbulk_evidence_externalized=%t\nsmall_replayable_fixtures_retained=%t\ncontent_addressed_evidence_manifest_fixture=%s\n",
		fixture.Status,
		fixture.BulkEvidenceExternalized,
		fixture.SmallReplayableFixturesRetained,
		filepath.ToSlash(*outPath),
	)
	return nil
}
