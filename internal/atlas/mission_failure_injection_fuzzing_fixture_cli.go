package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsFailureInjectionFuzzingFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations failure-injection-fuzzing-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "failure injection fuzzing fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasFailureInjectionFuzzingFixture()
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
	fmt.Fprintf(stdout, "status=%s\ncase_count=%d\ndeterministic_fuzzing=%t\nlive_provider_calls=%t\nfailure_injection_fuzzing_fixture=%s\n",
		fixture.Status,
		fixture.CaseCount,
		fixture.DeterministicFuzzing,
		fixture.LiveProviderCalls,
		filepath.ToSlash(*outPath),
	)
	return nil
}
