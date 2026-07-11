package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCanonicalJSONVectorSmokeChecks(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations canonical-json-vector-smoke-checks", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceVectorsPath := fs.String("source-vectors", "", "source canonical JSON vectors fixture path")
	outPath := fs.String("out", "", "canonical JSON vector smoke checks output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceVectorsPath) == "" {
		return fmt.Errorf("--source-vectors is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceVectorsPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	source, err := LoadJSON[AtlasCanonicalJSONVectors](*sourceVectorsPath)
	if err != nil {
		return err
	}
	fixture, err := BuildAtlasCanonicalJSONVectorSmokeChecks(source)
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
	fmt.Fprintf(stdout, "status=%s\nlanguage_count=%d\nvector_count=%d\nsmoke_check_count=%d\ncanonical_json_vector_smoke_checks=%s\n",
		fixture.Status,
		fixture.LanguageCount,
		fixture.VectorCount,
		fixture.SmokeCheckCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
