package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsContractCompatibilityInventory(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations contract-compatibility-inventory", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "contract compatibility inventory output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	inventory, err := BuildAtlasContractCompatibilityInventory()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, inventory); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, inventory)
	}
	fmt.Fprintf(stdout, "status=%s\ncontract_count=%d\ngate_critical_count=%d\nconsumer_test_count=%d\nmissing_owner_count=%d\nmissing_consumer_test_count=%d\ncontract_compatibility_inventory=%s\n",
		inventory.Status,
		inventory.ContractCount,
		inventory.GateCriticalCount,
		inventory.ConsumerTestCount,
		inventory.MissingOwnerCount,
		inventory.MissingConsumerTestCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}
