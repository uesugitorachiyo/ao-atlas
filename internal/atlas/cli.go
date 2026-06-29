package atlas

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	var err error
	switch args[0] {
	case "instance":
		err = runInstance(args[1:], stdout)
	case "intake":
		err = runIntake(args[1:], stdout)
	case "mission":
		err = runMission(args[1:], stdout)
	case "blueprint-request":
		err = runBlueprintRequest(args[1:], stdout)
	case "workgraph":
		err = runWorkgraph(args[1:], stdout)
	case "factory-task":
		err = runFactoryTask(args[1:], stdout)
	case "factory":
		err = runFactory(args[1:], stdout)
	case "context-pack":
		err = runContextPack(args[1:], stdout)
	case "foundry":
		err = runFoundry(args[1:], stdout)
	case "run-link":
		err = runRunLink(args[1:], stdout)
	default:
		err = fmt.Errorf("unknown command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "atlas <instance|intake|mission|blueprint-request|workgraph|factory-task|factory|context-pack|foundry|run-link> ...")
}

func runInstance(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("instance requires subcommand")
	}
	switch args[0] {
	case "init":
		fs := flag.NewFlagSet("instance init", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		id := fs.String("id", "", "instance id")
		stateRoot := fs.String("state-root", "", "state root")
		toolchainRoot := fs.String("toolchain-root", "", "toolchain root")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		instance := DefaultInstance(*id, *stateRoot, *toolchainRoot)
		if err := ValidateInstance(instance); err != nil {
			return err
		}
		if *out == "" {
			return fmt.Errorf("--out is required")
		}
		if err := WriteJSON(*out, instance); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\ninstance=%s\n", *out)
		return nil
	case "validate":
		fs := flag.NewFlagSet("instance validate", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("instance", "", "instance path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		instance, err := LoadJSON[Instance](*path)
		if err != nil {
			return err
		}
		if err := ValidateInstance(instance); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "status=valid")
		return nil
	case "registry":
		if len(args) < 2 || args[1] != "emit" {
			return fmt.Errorf("instance registry requires emit")
		}
		fs := flag.NewFlagSet("instance registry emit", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("instance", "", "instance path")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		instance, err := LoadJSON[Instance](*path)
		if err != nil {
			return err
		}
		if err := ValidateInstance(instance); err != nil {
			return err
		}
		registry := AtlasRegistry{
			ContractVersion: AtlasRegistryContract,
			InstanceID:      instance.ID,
			ToolchainRoot:   instance.ToolchainRoot,
			Roots:           instance.Roots,
			SchedulesWork:   false,
			ExecutesWork:    false,
			ApprovesWork:    false,
		}
		if err := ValidateAtlasRegistry(registry); err != nil {
			return err
		}
		if err := WriteJSON(*out, registry); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\nregistry=%s\n", *out)
		return nil
	case "inspect":
		fs := flag.NewFlagSet("instance inspect", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("instance", "", "instance path")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		instance, err := LoadJSON[Instance](*path)
		if err != nil {
			return err
		}
		if err := ValidateInstance(instance); err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(stdout, instance)
		}
		fmt.Fprintf(stdout, "id=%s\nstate_root=%s\ntoolchain_root=%s\n", instance.ID, instance.StateRoot, instance.ToolchainRoot)
		return nil
	case "doctor":
		fs := flag.NewFlagSet("instance doctor", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("instance", "", "instance path")
		registryPath := fs.String("registry", "", "registry path")
		out := fs.String("out", "", "output path")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		instance, err := LoadJSON[Instance](*path)
		if err != nil {
			return err
		}
		registry := AtlasRegistry{
			ContractVersion: AtlasRegistryContract,
			InstanceID:      instance.ID,
			ToolchainRoot:   instance.ToolchainRoot,
			Roots:           instance.Roots,
			SchedulesWork:   false,
			ExecutesWork:    false,
			ApprovesWork:    false,
		}
		if strings.TrimSpace(*registryPath) != "" {
			registry, err = LoadJSON[AtlasRegistry](*registryPath)
			if err != nil {
				return err
			}
		}
		report, err := BuildInstanceDoctorReport(instance, registry)
		if *out != "" {
			if samePath(*path, *out) || (strings.TrimSpace(*registryPath) != "" && samePath(*registryPath, *out)) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
			if writeErr := WriteJSON(*out, report); writeErr != nil {
				return writeErr
			}
		}
		if *jsonOut {
			if printErr := printJSON(stdout, report); printErr != nil {
				return printErr
			}
		} else {
			fmt.Fprintf(stdout, "status=%s\ninstance=%s\n", report.Status, report.InstanceID)
		}
		return err
	default:
		return fmt.Errorf("unknown instance subcommand %q", args[0])
	}
}

func runIntake(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("intake requires validate")
	}
	fs := flag.NewFlagSet("intake validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("intake", "", "intake path")
	out := fs.String("out-blueprint-request", "", "blueprint request output")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	intake, err := LoadJSON[Intake](*path)
	if err != nil {
		return err
	}
	request, err := ValidateIntake(intake)
	if err != nil {
		return err
	}
	if request.Status == "blueprint_required" {
		if *out != "" {
			if err := WriteJSON(*out, request); err != nil {
				return err
			}
		}
		if *jsonOut {
			return printJSON(stdout, request)
		}
		fmt.Fprintln(stdout, "status=blueprint_required")
		return nil
	}
	fmt.Fprintln(stdout, "status=ready")
	return nil
}

func runMission(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "status" {
		return fmt.Errorf("mission requires status")
	}
	fs := flag.NewFlagSet("mission status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	intakePath := fs.String("intake", "", "intake path")
	workgraphPath := fs.String("workgraph", "", "workgraph path")
	runLinkFlags := stringListFlag{}
	fs.Var(&runLinkFlags, "run-link", "run link path")
	out := fs.String("out", "", "output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *out != "" && (samePath(*intakePath, *out) || samePath(*workgraphPath, *out)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	intake, err := LoadJSON[Intake](*intakePath)
	if err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*workgraphPath)
	if err != nil {
		return err
	}
	links := []RunLink{}
	for _, path := range runLinkFlags {
		if *out != "" && samePath(path, *out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
		link, err := LoadJSON[RunLink](path)
		if err != nil {
			return err
		}
		links = append(links, link)
	}
	status, err := BuildMissionStatus(intake, workgraph, links)
	if err != nil {
		return err
	}
	if *out != "" {
		if err := WriteJSON(*out, status); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, status)
	}
	fmt.Fprintf(stdout, "status=%s\nintake=%s\nworkgraph=%s\n", status.CompletionStatus, status.IntakeID, status.WorkgraphID)
	return nil
}

func runBlueprintRequest(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("blueprint-request requires validate")
	}
	fs := flag.NewFlagSet("blueprint-request validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("request", "", "blueprint request path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	request, err := LoadJSON[BlueprintRequest](*path)
	if err != nil {
		return err
	}
	if err := ValidateBlueprintRequest(request); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "status=valid")
	return nil
}

func runWorkgraph(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("workgraph requires subcommand")
	}
	fs := flag.NewFlagSet("workgraph "+args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("workgraph", "", "workgraph path")
	runLinkPath := fs.String("run-link", "", "run link path")
	jsonOut := fs.Bool("json", false, "json output")
	out := fs.String("out", "", "output directory")
	dryRun := fs.Bool("dry-run", false, "write a dry-run skeleton without scheduling or executing")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*path)
	if err != nil {
		return err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return err
	}
	switch args[0] {
	case "validate":
		fmt.Fprintln(stdout, "status=valid")
	case "next":
		node, ok := NextReadyNode(workgraph)
		if !ok {
			fmt.Fprintln(stdout, "status=no_ready_task")
			return nil
		}
		if *jsonOut {
			return printJSON(stdout, node)
		}
		fmt.Fprintf(stdout, "status=ready\nnode=%s\ntask=%s\n", node.ID, node.FactoryTask.ID)
	case "status":
		counts := map[string]int{"ready": 0, "blocked": 0, "completed": 0}
		for _, node := range workgraph.Nodes {
			counts[node.Status]++
		}
		if *jsonOut {
			return printJSON(stdout, counts)
		}
		fmt.Fprintf(stdout, "ready=%d\nblocked=%d\ncompleted=%d\n", counts["ready"], counts["blocked"], counts["completed"])
	case "materialize-next":
		if !*dryRun {
			return fmt.Errorf("--dry-run is required for v0.1 workgraph materialization")
		}
		node, ok := NextReadyNode(workgraph)
		if !ok {
			fmt.Fprintln(stdout, "status=no_ready_task")
			return nil
		}
		materialization, err := MaterializeFactoryDryRun(node.FactoryTask, *out)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\nnode=%s\ntask=%s\nmode=%s\nexecutes_work=false\nschedules_work=false\n", node.ID, node.FactoryTask.ID, materialization.Mode)
	case "complete":
		if *out == "" {
			return fmt.Errorf("--out is required")
		}
		if samePath(*path, *out) {
			return fmt.Errorf("refusing to overwrite input workgraph")
		}
		link, err := LoadJSON[RunLink](*runLinkPath)
		if err != nil {
			return err
		}
		completed, nodeID, err := CompleteWorkgraph(workgraph, link)
		if err != nil {
			return err
		}
		if err := WriteJSON(*out, completed); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\nnode=%s\ntask=%s\n", nodeID, link.TaskID)
	case "repair-plan":
		if *out == "" {
			return fmt.Errorf("--out is required")
		}
		if samePath(*path, *out) || samePath(*runLinkPath, *out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
		link, err := LoadJSON[RunLink](*runLinkPath)
		if err != nil {
			return err
		}
		plan, err := BuildWorkgraphRepairPlan(workgraph, link)
		if err != nil {
			return err
		}
		if err := WriteJSON(*out, plan); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\ntask=%s\nrepair_tasks=%d\n", plan.TaskID, len(plan.RepairTasks))
	default:
		return fmt.Errorf("unknown workgraph subcommand %q", args[0])
	}
	return nil
}

func runFactoryTask(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("factory-task requires validate")
	}
	fs := flag.NewFlagSet("factory-task validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("task", "", "task path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	task, err := LoadJSON[FactoryTask](*path)
	if err != nil {
		return err
	}
	if err := ValidateFactoryTask(task); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "status=valid")
	return nil
}

func runFactory(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "materialize" {
		return fmt.Errorf("factory requires materialize")
	}
	fs := flag.NewFlagSet("factory materialize", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("task", "", "factory task path")
	out := fs.String("out", "", "output directory")
	dryRun := fs.Bool("dry-run", false, "write a dry-run skeleton without scheduling or executing")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if !*dryRun {
		return fmt.Errorf("--dry-run is required for v0.1 factory materialization")
	}
	task, err := LoadJSON[FactoryTask](*path)
	if err != nil {
		return err
	}
	materialization, err := MaterializeFactoryDryRun(task, *out)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "status=written\nmode=%s\nexecutes_work=false\nschedules_work=false\n", materialization.Mode)
	return nil
}

func runContextPack(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("context-pack requires subcommand")
	}
	switch args[0] {
	case "validate":
		fs := flag.NewFlagSet("context-pack validate", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("pack", "", "context pack path")
		budget := fs.Int("budget", 0, "override budget")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		pack, err := LoadJSON[ContextPack](*path)
		if err != nil {
			return err
		}
		if err := ValidateContextPack(pack, *budget); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "status=valid")
		return nil
	case "repack":
		fs := flag.NewFlagSet("context-pack repack", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		taskPath := fs.String("task", "", "factory task path")
		runLinkPath := fs.String("run-link", "", "run link path")
		sourceRef := fs.String("source-ref", "", "public-safe source reference")
		sourceDigest := fs.String("source-digest", "", "source digest")
		budget := fs.Int("budget", 4096, "context budget bytes")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *out == "" {
			return fmt.Errorf("--out is required")
		}
		if samePath(*taskPath, *out) || samePath(*runLinkPath, *out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
		task, err := LoadJSON[FactoryTask](*taskPath)
		if err != nil {
			return err
		}
		link, err := LoadJSON[RunLink](*runLinkPath)
		if err != nil {
			return err
		}
		pack, err := BuildContextRepack(task, link, *sourceRef, *sourceDigest, *budget)
		if err != nil {
			return err
		}
		if err := WriteJSON(*out, pack); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\ncontext_pack=%s\ntask=%s\n", *out, pack.TaskID)
		return nil
	default:
		return fmt.Errorf("unknown context-pack subcommand %q", args[0])
	}
}

func runFoundry(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("foundry requires subcommand")
	}
	switch {
	case len(args) >= 2 && strings.Join(args[:2], " ") == "handoff emit":
		fs := flag.NewFlagSet("foundry handoff emit", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("workgraph", "", "workgraph path")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		workgraph, err := LoadJSON[Workgraph](*path)
		if err != nil {
			return err
		}
		if err := ValidateWorkgraph(workgraph); err != nil {
			return err
		}
		handoff := BuildFoundryHandoff(workgraph)
		if err := ValidateFoundryHandoff(handoff); err != nil {
			return err
		}
		if err := WriteJSON(*out, handoff); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\nhandoff=%s\n", *out)
		return nil
	case args[0] == "import":
		fs := flag.NewFlagSet("foundry import", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("workgraph", "", "workgraph path")
		instancePath := fs.String("instance", "", "stack instance path")
		nodeID := fs.String("node", "", "optional workgraph node id")
		out := fs.String("out", "", "output directory")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*instancePath) == "" {
			return fmt.Errorf("--instance is required")
		}
		if *out == "" && !*jsonOut {
			return fmt.Errorf("--out or --json is required")
		}
		if *out != "" && (samePath(*path, *out) || samePath(*instancePath, *out)) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
		workgraph, err := LoadJSON[Workgraph](*path)
		if err != nil {
			return err
		}
		instance, err := LoadJSON[Instance](*instancePath)
		if err != nil {
			return err
		}
		if err := ValidateInstance(instance); err != nil {
			return err
		}
		if instance.ID != workgraph.TargetInstance {
			return fmt.Errorf("stack instance id must match workgraph target_instance")
		}
		registry := AtlasRegistry{
			ContractVersion: AtlasRegistryContract,
			InstanceID:      instance.ID,
			ToolchainRoot:   instance.ToolchainRoot,
			Roots:           instance.Roots,
			SchedulesWork:   false,
			ExecutesWork:    false,
			ApprovesWork:    false,
		}
		if _, err := BuildInstanceDoctorReport(instance, registry); err != nil {
			return err
		}
		selected := []string{}
		if strings.TrimSpace(*nodeID) != "" {
			selected = append(selected, strings.TrimSpace(*nodeID))
		}
		sourceArtifacts, err := sourceArtifactsForPaths(*path, *instancePath)
		if err != nil {
			return err
		}
		foundryImport, err := BuildFoundryImportForNodes(workgraph, selected, sourceArtifacts)
		if err != nil {
			return err
		}
		if err := validateFoundryImportContextPacks(foundryImport); err != nil {
			return err
		}
		if *out != "" {
			for _, fixture := range foundryImport.Tasks {
				if err := WriteJSON(filepath.Join(*out, fixture.Path), fixture.Task); err != nil {
					return err
				}
			}
			manifestPath := filepath.Join(*out, "foundry-import.json")
			if err := WriteJSON(manifestPath, foundryImport); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "status=written\nfoundry_import=%s\ntasks=%d\n", manifestPath, len(foundryImport.Tasks))
		}
		if *jsonOut {
			return printJSON(stdout, foundryImport)
		}
		return nil
	default:
		return fmt.Errorf("foundry requires handoff emit or import")
	}
}

func runRunLink(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("run-link requires subcommand")
	}
	switch args[0] {
	case "validate":
		fs := flag.NewFlagSet("run-link validate", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		path := fs.String("run-link", "", "run link path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		link, err := LoadJSON[RunLink](*path)
		if err != nil {
			return err
		}
		if err := ValidateRunLink(link); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "status=valid")
		return nil
	case "attach":
		fs := flag.NewFlagSet("run-link attach", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		taskID := fs.String("task-id", "", "task id")
		status := fs.String("status", "completed", "task completion status")
		out := fs.String("out", "", "output path")
		evidenceFlags := stringListFlag{}
		fs.Var(&evidenceFlags, "evidence", "evidence link as key=path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		evidence, err := parseEvidenceFlags(evidenceFlags)
		if err != nil {
			return err
		}
		link, err := BuildRunLink(*taskID, *status, evidence)
		if err != nil {
			return err
		}
		if *out == "" {
			return fmt.Errorf("--out is required")
		}
		if err := WriteJSON(*out, link); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\ntask=%s\ndigest=%s\n", link.TaskID, link.Digest)
		return nil
	default:
		return fmt.Errorf("unknown run-link subcommand %q", args[0])
	}
}

func printJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func fileDigest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func sourceArtifactsForPaths(paths ...string) ([]SourceRef, error) {
	artifacts := []SourceRef{}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		var errs []string
		checkPublicPath(&errs, "source_artifact", path, true)
		if err := joinErrors(errs); err != nil {
			return nil, err
		}
		digest, err := fileDigest(path)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, SourceRef{Ref: filepath.ToSlash(filepath.Clean(path)), Digest: digest})
	}
	return artifacts, nil
}

func validateFoundryImportContextPacks(foundryImport FoundryImport) error {
	for _, fixture := range foundryImport.Tasks {
		for _, ref := range fixture.Task.ContextPackRefs {
			resolved, err := resolveRepoRelativePath(ref)
			if err != nil {
				return fmt.Errorf("context pack %s: %w", ref, err)
			}
			pack, err := LoadJSON[ContextPack](resolved)
			if err != nil {
				return fmt.Errorf("context pack %s: %w", ref, err)
			}
			if err := ValidateContextPack(pack, 0); err != nil {
				return fmt.Errorf("context pack %s: %w", ref, err)
			}
			if pack.TaskID != fixture.TaskID {
				return fmt.Errorf("context pack %s task_id must match %s", ref, fixture.TaskID)
			}
		}
	}
	return nil
}

func resolveRepoRelativePath(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(cwd, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			resolved := filepath.Join(cwd, path)
			if _, err := os.Stat(resolved); err != nil {
				return "", err
			}
			return resolved, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf("repo root not found")
		}
		cwd = parent
	}
}

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func parseEvidenceFlags(values []string) (map[string]string, error) {
	evidence := map[string]string{}
	for _, value := range values {
		key, path, ok := strings.Cut(value, "=")
		if !ok {
			return nil, fmt.Errorf("--evidence must use key=path")
		}
		key = strings.TrimSpace(key)
		path = strings.TrimSpace(path)
		if key == "" || path == "" {
			return nil, fmt.Errorf("--evidence must use non-empty key=path")
		}
		evidence[key] = path
	}
	return evidence, nil
}

func samePath(left, right string) bool {
	if strings.TrimSpace(left) == "" || strings.TrimSpace(right) == "" {
		return false
	}
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr == nil && rightErr == nil {
		return filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
	}
	return filepath.Clean(left) == filepath.Clean(right)
}
