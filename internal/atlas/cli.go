package atlas

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
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
	fmt.Fprintln(w, "atlas <instance|intake|blueprint-request|workgraph|factory-task|factory|context-pack|foundry|run-link> ...")
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
		registry := map[string]any{
			"contract_version": "ao.atlas.foundry-registry.v0.1",
			"instance_id":      instance.ID,
			"toolchain_root":   instance.ToolchainRoot,
			"roots":            instance.Roots,
			"schedules_work":   false,
			"executes_work":    false,
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
	jsonOut := fs.Bool("json", false, "json output")
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
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("context-pack requires validate")
	}
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
}

func runFoundry(args []string, stdout io.Writer) error {
	if len(args) < 2 || strings.Join(args[:2], " ") != "handoff emit" {
		return fmt.Errorf("foundry requires handoff emit")
	}
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
}

func runRunLink(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("run-link requires validate")
	}
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
