package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsWorkspaceRootPreflightFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations workspace-root-preflight-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "workspace-root preflight fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasWorkspaceRootPreflightFixture()
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
	fmt.Fprintf(stdout, "status=%s\nrepository_identity_validated=%t\nobjective_digest_validated=%t\nworktree_boundary_validated=%t\nsafe_next_node_selected=%t\nworkspace_root_preflight_fixture=%s\n",
		fixture.Status,
		fixture.RepositoryIdentityValidated,
		fixture.ObjectiveDigestValidated,
		fixture.WorktreeBoundaryValidated,
		fixture.SafeNextNodeSelected,
		filepath.ToSlash(*outPath),
	)
	return nil
}
