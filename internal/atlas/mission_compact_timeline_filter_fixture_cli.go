package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCompactTimelineFilterFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations compact-timeline-filter-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "compact timeline filter fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasCompactTimelineFilterFixture()
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
	fmt.Fprintf(stdout, "status=%s\nfilter_count=%d\nstale_records_distinguished=%t\nduplicate_records_distinguished=%t\ncompact_timeline_filter_fixture=%s\n",
		fixture.Status,
		fixture.FilterCount,
		fixture.StaleRecordsDistinguished,
		fixture.DuplicateRecordsDistinguished,
		filepath.ToSlash(*outPath),
	)
	return nil
}
