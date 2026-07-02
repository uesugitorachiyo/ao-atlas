package atlas

import (
	"flag"
	"fmt"
	"io"
)

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
