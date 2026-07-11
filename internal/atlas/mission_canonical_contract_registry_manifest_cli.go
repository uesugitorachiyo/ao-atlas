package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsCanonicalContractRegistryManifest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations canonical-contract-registry-manifest", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "canonical contract registry manifest output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	manifest, err := BuildAtlasCanonicalContractRegistryManifest()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, manifest); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, manifest)
	}
	fmt.Fprintf(stdout, "status=%s\ncontract_count=%d\ngate_critical_count=%d\nconsumer_count=%d\ncanonical_contract_registry_manifest=%s\n",
		manifest.Status,
		manifest.ContractCount,
		manifest.GateCriticalCount,
		manifest.ConsumerCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
