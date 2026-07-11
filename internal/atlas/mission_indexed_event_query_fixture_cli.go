package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsIndexedEventQueryFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations indexed-event-query-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "indexed event query fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasIndexedEventQueryFixture()
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
	fmt.Fprintf(stdout, "status=%s\nevent_type_count=%d\nmigration_required=%t\nquery_index_required=%t\nindexed_event_query_fixture=%s\n",
		fixture.Status,
		fixture.EventTypeCount,
		fixture.MigrationRequired,
		fixture.QueryIndexRequired,
		filepath.ToSlash(*outPath),
	)
	return nil
}
