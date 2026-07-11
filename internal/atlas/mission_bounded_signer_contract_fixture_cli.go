package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsBoundedSignerContractFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations bounded-signer-contract-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "bounded signer contract fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasBoundedSignerContractFixture()
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
	fmt.Fprintf(stdout, "status=%s\nsigner_count=%d\nrotation_boundary=%s\nrevocation_boundary=%s\nbounded_signer_contract_fixture=%s\n",
		fixture.Status,
		fixture.SignerCount,
		fixture.RotationBoundary,
		fixture.RevocationBoundary,
		filepath.ToSlash(*outPath),
	)
	return nil
}
