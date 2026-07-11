package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsLocalPlatformFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations local-platform-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "local platform fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasLocalPlatformFixture()
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
	fmt.Fprintf(stdout, "status=%s\nplatform_count=%d\nline_ending_modes=%d\nlive_provider_calls=%t\nlocal_platform_fixture=%s\n",
		fixture.Status,
		fixture.PlatformCount,
		fixture.LineEndingModeCount,
		fixture.LiveProviderCalls,
		filepath.ToSlash(*outPath),
	)
	return nil
}
