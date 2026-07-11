package atlas

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func runMissionRecommendationsMonth3ControlPlaneObserver(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations month3-control-plane-observer", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "control-plane observer node id")
	adapterFixturePath := fs.String("adapter-fixture", "", "Command readback adapter boundary fixture path")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "control-plane observer binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":         *nodeID,
		"--adapter-fixture": *adapterFixturePath,
		"--source-readback": *sourceReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*adapterFixturePath, *sourceReadbackPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasMonth3ControlPlaneObserverBinding(*nodeID, *adapterFixturePath, *sourceReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMonth3ControlPlaneObserverBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nadapter_delegates_to_control_plane=%t\nobserver_read_only=%t\nduplicates_domain_decisions=%t\ncompleted_nodes=%d\nready_nodes=%d\nrsi_remains_denied=%t\nmonth3_control_plane_observer=%s\n",
		binding.Status,
		binding.AdapterDelegatesToControlPlane,
		binding.ObserverReadOnly,
		binding.DuplicatesDomainDecisions,
		binding.CompletedNodes,
		binding.ReadyNodes,
		binding.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}
