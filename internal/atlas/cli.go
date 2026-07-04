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
	case "blueprint":
		err = runBlueprint(args[1:], stdout)
	case "mission":
		err = runMission(args[1:], stdout)
	case "blueprint-request":
		err = runBlueprintRequest(args[1:], stdout)
	case "workgraph":
		err = runWorkgraph(args[1:], stdout)
	case "mutation-classes":
		err = runMutationClasses(args[1:], stdout)
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
	fmt.Fprintln(w, "atlas <instance|intake|blueprint|mission|blueprint-request|workgraph|mutation-classes|factory-task|factory|context-pack|foundry|run-link> ...")
}

func runBlueprint(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "import" {
		return fmt.Errorf("blueprint requires import")
	}
	fs := flag.NewFlagSet("blueprint import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	packPath := fs.String("pack", "", "AO Blueprint pack directory")
	candidateRulesPath := fs.String("candidate-rules", "", "Atlas-owned candidate rules path")
	authorizationPath := fs.String("authorization", "", "AO Blueprint build authorization packet")
	instancePath := fs.String("instance", "", "Atlas stack instance path")
	mutationClassesPath := fs.String("mutation-classes", "", "Atlas mutation class model path")
	outDir := fs.String("out", "", "output directory")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*packPath) == "" {
		return fmt.Errorf("--pack is required")
	}
	if strings.TrimSpace(*instancePath) == "" {
		return fmt.Errorf("--instance is required")
	}
	if strings.TrimSpace(*mutationClassesPath) == "" {
		return fmt.Errorf("--mutation-classes is required")
	}
	if strings.TrimSpace(*outDir) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if *outDir != "" {
		for _, input := range []string{*packPath, *authorizationPath, *instancePath, *mutationClassesPath} {
			if samePath(input, *outDir) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
		}
	}
	result, err := BuildBlueprintImport(BlueprintImportPaths{
		PackPath:            *packPath,
		CandidateRulesPath:  *candidateRulesPath,
		AuthorizationPath:   *authorizationPath,
		InstancePath:        *instancePath,
		MutationClassesPath: *mutationClassesPath,
		OutDir:              *outDir,
	})
	if *jsonOut {
		if printErr := printJSON(stdout, result.Record); printErr != nil {
			return printErr
		}
	} else {
		fmt.Fprintf(stdout, "status=%s\nblueprint_import=%s\nready_for_foundry=%t\n", result.Record.Status, filepath.ToSlash(filepath.Join(*outDir, "blueprint-import.json")), result.Record.ReadyForFoundry)
		if result.Record.ReadyForFoundry {
			handoffPath := filepath.ToSlash(filepath.Join(*outDir, "foundry-import", "foundry-continuation-handoff.json"))
			promptPath := filepath.ToSlash(filepath.Join(*outDir, "foundry-import", "foundry-continuation-prompt.md"))
			fmt.Fprintf(stdout, "foundry_continuation_handoff=%s\nfoundry_continuation_prompt=%s\nnext_recommended_action=%s\nMove to %s\nRun %s\nPaste this prompt\n", handoffPath, promptPath, result.Handoff.NextRecommendedAction, result.Handoff.TargetFolder, result.Handoff.Command)
		}
	}
	return err
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
	if len(args) == 0 {
		return fmt.Errorf("mission requires status, import, workgraph-metadata, provenance, or recommendations")
	}
	if args[0] == "import" {
		return runMissionImport(args[1:], stdout)
	}
	if args[0] == "workgraph-metadata" {
		return runMissionWorkgraphMetadata(args[1:], stdout)
	}
	if args[0] == "provenance" {
		return runMissionProvenance(args[1:], stdout)
	}
	if args[0] == "recommendations" {
		return runMissionRecommendations(args[1:], stdout)
	}
	if args[0] != "status" {
		return fmt.Errorf("mission requires status, import, workgraph-metadata, provenance, or recommendations")
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

func runMissionRecommendations(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "import" {
		return fmt.Errorf("mission recommendations requires import")
	}
	fs := flag.NewFlagSet("mission recommendations import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	recommendationsPath := fs.String("recommendations", "", "AO Mission Feature Depth Recommendations path")
	targetInstance := fs.String("target-instance", "", "Atlas target instance id")
	minTasks := fs.Int("min-tasks", 20, "minimum Atlas recommendation tasks")
	nodeBudget := fs.Int("node-budget", 20, "Atlas node budget")
	estimatedMinutes := fs.Int("estimated-minutes", 90, "estimated long-run minutes")
	outDir := fs.String("out", "", "output directory")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*recommendationsPath) == "" {
		return fmt.Errorf("--recommendations is required")
	}
	if strings.TrimSpace(*targetInstance) == "" {
		return fmt.Errorf("--target-instance is required")
	}
	if strings.TrimSpace(*outDir) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if *outDir != "" && samePath(*recommendationsPath, *outDir) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: *recommendationsPath,
		TargetInstance:      *targetInstance,
		MinTasks:            *minTasks,
		NodeBudget:          *nodeBudget,
		EstimatedMinutes:    *estimatedMinutes,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outDir) != "" {
		if err := WriteAtlasRecommendationWaveArtifacts(*outDir, result); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, result.Wave)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nrecommendation_tasks=%d\nnode_budget=%d\nestimated_minutes=%d\nrecommendation_wave=%s\nrecommendation_workgraph=%s\nnext_recommended_prompt=%s\n",
		result.Wave.Status,
		result.Wave.MissionID,
		result.Wave.TotalTasks,
		result.Wave.NodeBudget,
		result.Wave.EstimatedMinutes,
		filepath.ToSlash(filepath.Join(*outDir, "recommendation-wave.json")),
		filepath.ToSlash(filepath.Join(*outDir, "recommendation-workgraph.json")),
		filepath.ToSlash(filepath.Join(*outDir, "next-recommended-prompt.md")),
	)
	return nil
}

func runMissionProvenance(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "render" {
		return fmt.Errorf("mission provenance requires render")
	}
	fs := flag.NewFlagSet("mission provenance render", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	metadataPath := fs.String("metadata", "", "AO Mission workgraph metadata path")
	outPath := fs.String("out", "", "output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*metadataPath) == "" {
		return fmt.Errorf("--metadata is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if samePath(*metadataPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	metadata, err := LoadJSON[AOMissionWorkgraphMetadata](*metadataPath)
	if err != nil {
		return err
	}
	render, err := BuildAOMissionProvenanceRender(metadata)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, render); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, render)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nprimary_mission_provenance=%s\nprovenance_summary=%s\n", render.Status, render.MissionID, render.PrimaryMissionProvenance, render.ProvenanceSummary)
	return nil
}

func runMissionImport(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	recordPath := fs.String("record", "", "AO Mission record path")
	commandStatusPath := fs.String("command-status", "", "AO Command mission status path")
	artifactManifestPath := fs.String("artifact-manifest", "", "AO Mission artifact manifest path")
	routeHistoryPath := fs.String("route-history", "", "optional AO Mission route history path")
	schedulerRecoveryPath := fs.String("scheduler-recovery", "", "optional AO Mission scheduler recovery readback path")
	ledgerCompactionPath := fs.String("ledger-compaction", "", "optional AO Mission ledger compaction readback path")
	timelineCompactionPath := fs.String("timeline-compaction", "", "optional AO Mission timeline compaction readback path")
	missionArchivePath := fs.String("mission-archive", "", "optional AO Mission archive path")
	gatewayReadinessRollupPath := fs.String("gateway-readiness-rollup", "", "optional AO Mission gateway readiness rollup path")
	outPath := fs.String("out", "", "output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*recordPath) == "" || strings.TrimSpace(*commandStatusPath) == "" || strings.TrimSpace(*artifactManifestPath) == "" {
		return fmt.Errorf("--record, --command-status, and --artifact-manifest are required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if *outPath != "" {
		for _, input := range []string{*recordPath, *commandStatusPath, *artifactManifestPath, *routeHistoryPath, *schedulerRecoveryPath, *ledgerCompactionPath, *timelineCompactionPath, *missionArchivePath, *gatewayReadinessRollupPath} {
			if samePath(input, *outPath) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
		}
	}
	importRecord, err := BuildAOMissionImportWithTimelineCompaction(*recordPath, *commandStatusPath, *artifactManifestPath, *routeHistoryPath, *schedulerRecoveryPath, *ledgerCompactionPath, *timelineCompactionPath, *missionArchivePath, *gatewayReadinessRollupPath)
	if err != nil {
		return err
	}
	if *outPath != "" {
		if err := WriteJSON(*outPath, importRecord); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, importRecord)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nao_mission_import=%s\n", importRecord.Status, importRecord.MissionID, *outPath)
	return nil
}

func runMissionWorkgraphMetadata(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission workgraph-metadata", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	importPath := fs.String("import", "", "AO Mission import path")
	workgraphPath := fs.String("workgraph", "", "Atlas workgraph path")
	outPath := fs.String("out", "", "output path")
	provenanceWorkgraphOut := fs.String("provenance-workgraph-out", "", "optional output path for workgraph with AO Mission provenance nodes")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*importPath) == "" || strings.TrimSpace(*workgraphPath) == "" {
		return fmt.Errorf("--import and --workgraph are required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if *outPath != "" {
		for _, input := range []string{*importPath, *workgraphPath} {
			if samePath(input, *outPath) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
		}
	}
	if *provenanceWorkgraphOut != "" {
		for _, input := range []string{*importPath, *workgraphPath} {
			if samePath(input, *provenanceWorkgraphOut) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
		}
	}
	metadata, err := BuildAOMissionWorkgraphMetadata(*importPath, *workgraphPath)
	if err != nil {
		return err
	}
	if *provenanceWorkgraphOut != "" {
		importRecord, err := LoadJSON[AOMissionImport](*importPath)
		if err != nil {
			return err
		}
		workgraph, err := LoadJSON[Workgraph](*workgraphPath)
		if err != nil {
			return err
		}
		provenanceWorkgraph, err := BuildAOMissionProvenanceWorkgraph(importRecord, workgraph)
		if err != nil {
			return err
		}
		if err := WriteJSON(*provenanceWorkgraphOut, provenanceWorkgraph); err != nil {
			return err
		}
	}
	if *outPath != "" {
		if err := WriteJSON(*outPath, metadata); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, metadata)
	}
	fmt.Fprintf(stdout, "status=ready\nmission_id=%s\nworkgraph=%s\nao_mission_workgraph_metadata=%s\n", metadata.MissionID, metadata.WorkgraphID, *outPath)
	if *provenanceWorkgraphOut != "" {
		fmt.Fprintf(stdout, "ao_mission_provenance_workgraph=%s\n", *provenanceWorkgraphOut)
	}
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

func runMutationClasses(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("mutation-classes requires validate")
	}
	fs := flag.NewFlagSet("mutation-classes validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("model", "", "mutation class model path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	model, err := LoadJSON[MutationClassModel](*path)
	if err != nil {
		return err
	}
	if err := ValidateMutationClassModel(model); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "status=valid")
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
		aoMissionMetadataPath := fs.String("ao-mission-metadata", "", "optional AO Mission workgraph metadata source artifact")
		blueprintPackPath := fs.String("blueprint-pack", "", "optional Blueprint pack path for Foundry continuation handoff")
		atlasImportPath := fs.String("atlas-import", "", "optional Atlas import path for Foundry continuation handoff")
		missionContinuationPath := fs.String("mission-continuation", "", "optional mission continuation evidence path for Foundry continuation handoff")
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
		if *out != "" && (samePath(*path, *out) || samePath(*instancePath, *out) || samePath(*aoMissionMetadataPath, *out)) {
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
		if strings.TrimSpace(*aoMissionMetadataPath) != "" {
			metadata, err := LoadJSON[AOMissionWorkgraphMetadata](*aoMissionMetadataPath)
			if err != nil {
				return err
			}
			if err := ValidateAOMissionWorkgraphMetadata(metadata, workgraph); err != nil {
				return err
			}
		}
		selected := []string{}
		if strings.TrimSpace(*nodeID) != "" {
			selected = append(selected, strings.TrimSpace(*nodeID))
		}
		sourceArtifacts, err := sourceArtifactsForPaths(*path, *instancePath, *aoMissionMetadataPath)
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
			continuation, err := BuildFoundryContinuationHandoff(workgraph, foundryImport, FoundryContinuationHandoffInputs{
				BlueprintPackPath:               *blueprintPackPath,
				AtlasImportPath:                 *atlasImportPath,
				WorkgraphPath:                   *path,
				FoundryImportPath:               manifestPath,
				MissionContinuationEvidencePath: *missionContinuationPath,
			})
			if err != nil {
				return err
			}
			continuationPath := filepath.Join(*out, "foundry-continuation-handoff.json")
			if err := WriteJSON(continuationPath, continuation); err != nil {
				return err
			}
			promptPath := filepath.Join(*out, "foundry-continuation-prompt.md")
			if err := WriteFoundryContinuationPrompt(promptPath, continuation); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "status=written\nfoundry_import=%s\nfoundry_continuation_handoff=%s\nfoundry_continuation_prompt=%s\ntasks=%d\nnext_recommended_action=%s\nMove to %s\nRun %s\nPaste this prompt\n", manifestPath, continuationPath, promptPath, len(foundryImport.Tasks), continuation.NextRecommendedAction, continuation.TargetFolder, continuation.Command)
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
