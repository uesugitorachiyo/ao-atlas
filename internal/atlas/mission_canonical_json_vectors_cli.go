package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCanonicalJSONVectors(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations canonical-json-vectors", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "canonical JSON vectors output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasCanonicalJSONVectors()
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
	fmt.Fprintf(stdout, "status=%s\nvector_count=%d\nlanguage_count=%d\ndigest_algorithm=%s\ncanonical_json_vectors=%s\n",
		fixture.Status,
		fixture.VectorCount,
		fixture.LanguageCount,
		fixture.DigestAlgorithm,
		filepath.ToSlash(*outPath),
	)
	return nil
}
