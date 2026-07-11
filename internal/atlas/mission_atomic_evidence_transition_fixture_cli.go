package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsAtomicEvidenceTransitionFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations atomic-evidence-transition-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "atomic evidence transition fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasAtomicEvidenceTransitionFixture()
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
	fmt.Fprintf(stdout, "status=%s\nscenario_count=%d\natomic_transitions_required=%t\nduplicate_ingest_idempotent=%t\natomic_evidence_transition_fixture=%s\n",
		fixture.Status,
		fixture.ScenarioCount,
		fixture.AtomicTransitionsRequired,
		fixture.DuplicateIngestIdempotent,
		filepath.ToSlash(*outPath),
	)
	return nil
}
