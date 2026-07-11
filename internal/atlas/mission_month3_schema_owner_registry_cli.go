package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3SchemaOwnerRegistry(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-schema-owner-registry", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "schema owner registry node id")
	registryManifestPath := fs.String("registry-manifest", "", "canonical contract registry manifest path")
	outPath := fs.String("out", "", "schema owner registry proposal output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":           *nodeID,
		"--registry-manifest": *registryManifestPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*registryManifestPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	proposal, err := BuildAtlasMonth3SchemaOwnerRegistryProposal(*nodeID, *registryManifestPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3SchemaOwnerRegistryProposal(*outPath, proposal); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, proposal)
	}
	fmt.Fprintf(stdout, "status=%s\nregistry_authority_owner=%s\ncontract_count=%d\nconsumer_compatibility_check_count=%d\ncovenant_owns_registry=%t\nexecutes_work=%t\nrsi_remains_denied=%t\nmonth3_schema_owner_registry=%s\n",
		proposal.Status,
		proposal.RegistryAuthorityOwner,
		proposal.ContractCount,
		proposal.ConsumerCompatibilityCheckCount,
		proposal.CovenantOwnsRegistry,
		proposal.ExecutesWork,
		proposal.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
