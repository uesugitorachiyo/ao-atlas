package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsAuthorityReadinessInventoryFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations authority-readiness-inventory-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "authority readiness inventory fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasAuthorityReadinessInventoryFixture()
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
	fmt.Fprintf(stdout, "status=%s\ninput_count=%d\ngenerated_from_inputs=%t\ncopied_campaign_prose_allowed=%t\nauthority_readiness_inventory_fixture=%s\n",
		fixture.Status,
		fixture.InputCount,
		fixture.GeneratedFromInputs,
		fixture.CopiedCampaignProseAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}
