package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsFoundryCanonicalImportFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-canonical-import-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	workgraphPath := fs.String("workgraph", "", "source Atlas workgraph path")
	expectedNode := fs.String("expected-node", "", "expected ready node id")
	outPath := fs.String("out", "", "Foundry canonical import fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--workgraph":     *workgraphPath,
		"--expected-node": *expectedNode,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasFoundryCanonicalImportFixture(*workgraphPath, *expectedNode)
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
	fmt.Fprintf(stdout, "status=%s\naccepted_canonical_import=%t\nrejected_alias_count=%d\nexpected_node=%s\nfoundry_canonical_import_fixture=%s\n",
		fixture.Status,
		fixture.AcceptedCanonicalImport,
		fixture.RejectedAliasCount,
		fixture.ExpectedNode,
		filepath.ToSlash(*outPath),
	)
	return nil
}
