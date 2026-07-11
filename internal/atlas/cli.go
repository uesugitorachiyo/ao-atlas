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
		return fmt.Errorf("mission requires status, import, final-synthesis, workgraph-metadata, provenance, or recommendations")
	}
	if args[0] == "import" {
		return runMissionImport(args[1:], stdout)
	}
	if args[0] == "final-synthesis" {
		return runMissionFinalSynthesis(args[1:], stdout)
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
		return fmt.Errorf("mission requires status, import, final-synthesis, workgraph-metadata, provenance, or recommendations")
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

func runMissionFinalSynthesis(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "import" {
		return fmt.Errorf("mission final-synthesis requires import")
	}
	fs := flag.NewFlagSet("mission final-synthesis import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	synthesisPath := fs.String("synthesis", "", "AO Mission final synthesis path")
	outPath := fs.String("out", "", "output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*synthesisPath) == "" {
		return fmt.Errorf("--synthesis is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if *outPath != "" && samePath(*synthesisPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := BuildAOMissionFinalSynthesisReadback(*synthesisPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, readback); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, readback)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\ncompleted_nodes=%d\nready_nodes=%d\nfinal_response_allowed=%t\nao_mission_final_synthesis_readback=%s\n",
		readback.Status,
		readback.MissionID,
		readback.CompletedNodes,
		readback.ReadyNodes,
		readback.FinalResponseAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}

type missionRecommendationCommand struct {
	name             string
	run              func([]string, io.Writer) error
	commandClass     string
	recordsRunLedger bool
}

const (
	missionRecommendationCommandClassPlanningOnly    = "planning_only"
	missionRecommendationCommandClassMutationCapable = "mutation_capable"
)

func missionRecommendationCommandRegistry() []missionRecommendationCommand {
	return []missionRecommendationCommand{
		{name: "import", run: runMissionRecommendationsImport, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "export-next-wave", run: runMissionRecommendationsExportNextWave, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "export-refactoring-wave", run: runMissionRecommendationsExportRefactoringWave, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "next-track", run: runMissionRecommendationsNextTrack, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "consumed-ledger", run: runMissionRecommendationsConsumedLedger, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "track-registry", run: runMissionRecommendationsTrackRegistry, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "run-ledger", run: runMissionRecommendationsRunLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-ledger-rollup", run: runMissionRecommendationsRunLedgerRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-ledger-coverage-check", run: runMissionRecommendationsRunLedgerCoverageCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "final-response-gates", run: runMissionRecommendationsFinalResponseGates, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-registry", run: runMissionRecommendationsSchemaRegistry, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-registry-health", run: runMissionRecommendationsSchemaRegistryHealth, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "schema-registry-coverage", run: runMissionRecommendationsSchemaRegistryCoverage, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-health-repair-prompt", run: runMissionRecommendationsSchemaHealthRepairPrompt, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback", run: runMissionRecommendationsReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback-delta", run: runMissionRecommendationsReadbackDelta, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback-diff-fixture", run: runMissionRecommendationsReadbackDiffFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "stale-checkpoint-rejection", run: runMissionRecommendationsStaleCheckpointRejection, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "operator-summary-check", run: runMissionRecommendationsOperatorSummaryCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-link-schema-coverage", run: runMissionRecommendationsRunLinkSchemaCoverage, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "schema-validator-drift", run: runMissionRecommendationsSchemaValidatorDrift, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "pr-ci-timing-summary", run: runMissionRecommendationsPRCITimingSummary, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "pr-ci-windows-threshold", run: runMissionRecommendationsPRCIWindowsThreshold, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "failed-check-replay", run: runMissionRecommendationsFailedCheckReplay, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-rejected-ticket-fixture", run: runMissionRecommendationsCommandCovenantRejectedTicketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-quarantine-fixture", run: runMissionRecommendationsCommandCovenantQuarantineFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-ticket-byte-preservation-fixture", run: runMissionRecommendationsCommandTicketBytePreservationFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "ticket-digest-readback-binding-fixture", run: runMissionRecommendationsTicketDigestReadbackBindingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-hash-mismatch-rejection-fixture", run: runMissionRecommendationsPolicyHashMismatchRejectionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-version-replay-rejection-fixture", run: runMissionRecommendationsPolicyVersionReplayRejectionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "covenant-evidence-digest-readback-fixture", run: runMissionRecommendationsCovenantEvidenceDigestReadbackFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-compact-rejection-reason-fixture", run: runMissionRecommendationsCommandCompactRejectionReasonFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "blueprint-ticket-schema-compatibility-ledger", run: runMissionRecommendationsBlueprintTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "atlas-ticket-schema-compatibility-ledger", run: runMissionRecommendationsAtlasTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-ticket-schema-compatibility-ledger", run: runMissionRecommendationsFoundryTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-ticket-schema-compatibility-ledger", run: runMissionRecommendationsCommandTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "covenant-ticket-schema-authority-ledger", run: runMissionRecommendationsCovenantTicketSchemaAuthorityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-ticket-public-safety-scan", run: runMissionRecommendationsPolicyTicketPublicSafetyScan, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "merge-check-binding", run: runMissionRecommendationsMergeCheckBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "post-merge-branch-deletion-readback", run: runMissionRecommendationsPostMergeBranchDeletionReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "stale-remote-branch-repair", run: runMissionRecommendationsStaleRemoteBranchRepair, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-main-sync-readback", run: runMissionRecommendationsLocalMainSyncReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "branch-cleanup-handoff-summary", run: runMissionRecommendationsBranchCleanupHandoffSummary, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compaction-resume-prompt", run: runMissionRecommendationsCompactionResumePrompt, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compaction-resume-regression", run: runMissionRecommendationsCompactionResumeRegression, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "resume-denial-evidence", run: runMissionRecommendationsResumeDenialEvidence, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "public-safety-readback-binding", run: runMissionRecommendationsPublicSafetyReadbackBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "scoped-public-safety-scan", run: runMissionRecommendationsScopedPublicSafetyScan, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "authority-promotion-negative-fixtures", run: runMissionRecommendationsAuthorityPromotionNegativeFixtures, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "public-safety-coverage-rollup", run: runMissionRecommendationsPublicSafetyCoverageRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-no-promotion-rollup", run: runMissionRecommendationsPromoterNoPromotionRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-promoter-agreement-rollup", run: runMissionRecommendationsCommandPromoterAgreementRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-rollup-count-mismatch-regression", run: runMissionRecommendationsPromoterRollupCountMismatchRegression, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-promoter-disagreement-denial", run: runMissionRecommendationsCommandPromoterDisagreementDenial, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-import-readiness-binding", run: runMissionRecommendationsFoundryImportReadinessBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-link-digest-check", run: runMissionRecommendationsRunLinkDigestCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-handoff-replay-fixture", run: runMissionRecommendationsFoundryHandoffReplayFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-terminal-status-examples", run: runMissionRecommendationsFoundryTerminalStatusExamples, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-closure-binding", run: runMissionRecommendationsMissionDashboardClosureBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-provenance-links", run: runMissionRecommendationsMissionDashboardProvenanceLinks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-freshness-checks", run: runMissionRecommendationsMissionDashboardFreshnessChecks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-compact-filters", run: runMissionRecommendationsMissionDashboardCompactFilters, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "bounded-signer-contract-fixture", run: runMissionRecommendationsBoundedSignerContractFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-contract-registry-manifest", run: runMissionRecommendationsCanonicalContractRegistryManifest, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "contract-compatibility-inventory", run: runMissionRecommendationsContractCompatibilityInventory, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-json-vectors", run: runMissionRecommendationsCanonicalJSONVectors, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-json-vector-smoke-checks", run: runMissionRecommendationsCanonicalJSONVectorSmokeChecks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "sentinel-hosted-ci-workflow-fixture", run: runMissionRecommendationsSentinelHostedCIWorkflowFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "sentinel-signal-state-fixture", run: runMissionRecommendationsSentinelSignalStateFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "signed-assurance-dry-run-fixture", run: runMissionRecommendationsSignedAssuranceDryRunFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-no-activation-boundary-fixture", run: runMissionRecommendationsPromoterNoActivationBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "workspace-root-preflight-fixture", run: runMissionRecommendationsWorkspaceRootPreflightFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "bounded-execution-packet-fixture", run: runMissionRecommendationsBoundedExecutionPacketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "forge-goalrun-evidence-fixture", run: runMissionRecommendationsForgeGoalRunEvidenceFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "execution-packet-regression-matrix", run: runMissionRecommendationsExecutionPacketRegressionMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "durable-state-migration-metadata", run: runMissionRecommendationsDurableStateMigrationMetadata, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "exactly-once-resume-accounting-fixture", run: runMissionRecommendationsExactlyOnceResumeAccountingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "replayable-state-packet-fixture", run: runMissionRecommendationsReplayableStatePacketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "indexed-event-query-fixture", run: runMissionRecommendationsIndexedEventQueryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "atomic-evidence-transition-fixture", run: runMissionRecommendationsAtomicEvidenceTransitionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-backup-restore-fixture", run: runMissionRecommendationsLocalBackupRestoreFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-readback-adapter-boundary-fixture", run: runMissionRecommendationsCommandReadbackAdapterBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compact-timeline-filter-fixture", run: runMissionRecommendationsCompactTimelineFilterFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "authority-readiness-inventory-fixture", run: runMissionRecommendationsAuthorityReadinessInventoryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "content-addressed-evidence-manifest-fixture", run: runMissionRecommendationsContentAddressedEvidenceManifestFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-evidence-size-boundary-fixture", run: runMissionRecommendationsFoundryEvidenceSizeBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "repeated-task-result-ledger-fixture", run: runMissionRecommendationsRepeatedTaskResultLedgerFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "failure-injection-fuzzing-fixture", run: runMissionRecommendationsFailureInjectionFuzzingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-platform-fixture", run: runMissionRecommendationsLocalPlatformFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "non-ao-replay-binding-fixture", run: runMissionRecommendationsNonAOReplayBindingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "kill-restart-replay-fixture", run: runMissionRecommendationsKillRestartReplayFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "rollback-terminal-readback-fixture", run: runMissionRecommendationsRollbackTerminalReadbackFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "golden-path-readiness-matrix", run: runMissionRecommendationsGoldenPathReadinessMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "blueprint-canonical-preservation-fixture", run: runMissionRecommendationsBlueprintCanonicalPreservationFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-canonical-import-fixture", run: runMissionRecommendationsFoundryCanonicalImportFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-field-parity-fixture", run: runMissionRecommendationsCommandCovenantFieldParityFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "complete-node", run: runMissionRecommendationsCompleteNode, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "resume", run: runMissionRecommendationsResume, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "validate-evidence", run: runMissionRecommendationsValidateEvidence, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
	}
}

func missionRecommendationCommandNames() []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.name)
	}
	return names
}

func missionRecommendationRunLedgerCommandNames() []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		if command.recordsRunLedger {
			names = append(names, command.name)
		}
	}
	return names
}

func missionRecommendationPlanningOnlyCommandNames() []string {
	return missionRecommendationCommandNamesByClass(missionRecommendationCommandClassPlanningOnly)
}

func missionRecommendationMutationCapableCommandNames() []string {
	return missionRecommendationCommandNamesByClass(missionRecommendationCommandClassMutationCapable)
}

func missionRecommendationCommandNamesByClass(commandClass string) []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		if command.commandClass == commandClass {
			names = append(names, command.name)
		}
	}
	return names
}

func missionRecommendationsUsageError() error {
	return fmt.Errorf("mission recommendations requires %s", formatCommandList(missionRecommendationCommandNames()))
}

func formatCommandList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}

func runMissionRecommendations(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return missionRecommendationsUsageError()
	}
	for _, command := range missionRecommendationCommandRegistry() {
		if args[0] == command.name {
			return command.run(args[1:], stdout)
		}
	}
	return missionRecommendationsUsageError()
}

func runMissionRecommendationsImport(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	recommendationsPath := fs.String("recommendations", "", "AO Mission Feature Depth Recommendations path")
	targetInstance := fs.String("target-instance", "", "Atlas target instance id")
	minTasks := fs.Int("min-tasks", 0, "minimum Atlas recommendation tasks")
	nodeBudget := fs.Int("node-budget", 0, "Atlas node budget")
	estimatedMinutes := fs.Int("estimated-minutes", 0, "estimated long-run minutes")
	minMinutes := fs.Int("min-minutes", 0, "minimum lease minutes")
	maxMinutes := fs.Int("max-minutes", 0, "maximum lease minutes")
	continueIfFastTarget := fs.Int("continue-if-fast-target", 0, "continue-if-fast node target")
	returnOnlyWhen := fs.String("return-only-when", "", "final response return policy")
	checkpointPolicy := fs.String("checkpoint-policy", "", "checkpoint policy")
	evidencePolicy := fs.String("evidence-policy", "", "evidence policy")
	finalReportContract := fs.String("final-report-contract", "", "final report contract")
	startedAt := fs.String("started-at", "", "long-run lease start time, RFC3339")
	outDir := fs.String("out", "", "output directory")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
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
		RecommendationsPath:  *recommendationsPath,
		TargetInstance:       *targetInstance,
		MinTasks:             *minTasks,
		NodeBudget:           *nodeBudget,
		EstimatedMinutes:     *estimatedMinutes,
		MinMinutes:           *minMinutes,
		MaxMinutes:           *maxMinutes,
		ContinueIfFastTarget: *continueIfFastTarget,
		ReturnOnlyWhen:       *returnOnlyWhen,
		CheckpointPolicy:     *checkpointPolicy,
		EvidencePolicy:       *evidencePolicy,
		FinalReportContract:  *finalReportContract,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outDir) != "" {
		if err := WriteAtlasRecommendationWaveArtifacts(*outDir, result); err != nil {
			return err
		}
		if strings.TrimSpace(*startedAt) != "" {
			leaseStart, err := BuildAtlasRecommendationLeaseStart(result.Wave, result.Workgraph, AtlasRecommendationLeaseStartOptions{
				WavePath:      filepath.Join(*outDir, "recommendation-wave.json"),
				WorkgraphPath: filepath.Join(*outDir, "recommendation-workgraph.json"),
				EvidenceRoot:  filepath.ToSlash(*outDir),
				StartedAt:     *startedAt,
			})
			if err != nil {
				return err
			}
			if err := WriteJSON(filepath.Join(*outDir, "lease-start.json"), leaseStart); err != nil {
				return err
			}
		}
	}
	if *jsonOut {
		return printJSON(stdout, result.Wave)
	}
	minLeaseMinutes := result.Wave.EstimatedMinutes
	maxLeaseMinutes := result.Wave.EstimatedMinutes
	continueTarget := result.Wave.NodeBudget
	if result.Wave.Supervisor != nil {
		minLeaseMinutes = result.Wave.Supervisor.MinMinutes
		maxLeaseMinutes = result.Wave.Supervisor.MaxMinutes
		continueTarget = result.Wave.Supervisor.ContinueIfFastTarget
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nrecommendation_tasks=%d\nnode_budget=%d\nestimated_minutes=%d\nmin_minutes=%d\nmax_minutes=%d\ncontinue_if_fast_target=%d\nfinal_response_allowed=%t\nrecommendation_wave=%s\nrecommendation_workgraph=%s\nlease_start=%s\nrecommendation_readback=%s\nworkgraph_readiness_packet=%s\nnext_recommended_prompt=%s\n",
		result.Wave.Status,
		result.Wave.MissionID,
		result.Wave.TotalTasks,
		result.Wave.NodeBudget,
		result.Wave.EstimatedMinutes,
		minLeaseMinutes,
		maxLeaseMinutes,
		continueTarget,
		result.Wave.FinalResponseAllowed,
		filepath.ToSlash(filepath.Join(*outDir, "recommendation-wave.json")),
		filepath.ToSlash(filepath.Join(*outDir, "recommendation-workgraph.json")),
		filepath.ToSlash(filepath.Join(*outDir, "lease-start.json")),
		filepath.ToSlash(filepath.Join(*outDir, "recommendation-readback.json")),
		filepath.ToSlash(filepath.Join(*outDir, "workgraph-readiness-packet.json")),
		filepath.ToSlash(filepath.Join(*outDir, "next-recommended-prompt.md")),
	)
	return nil
}

func runMissionRecommendationsExportNextWave(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations export-next-wave", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	missionID := fs.String("mission-id", "ao-atlas-next-feature-depth-wave-v01", "next wave mission id")
	sourceEvidenceRoot := fs.String("source-evidence-root", "", "source evidence root")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	sourceAssertionPath := fs.String("source-assertion", "", "source no-promotion/no-RSI assertion path")
	minTasks := fs.Int("min-tasks", 40, "minimum ranked Feature Depth tasks")
	outPath := fs.String("out", "", "output Feature Depth recommendations path")
	fixtureOutPath := fs.String("fixture-out", "", "output next-wave exporter fixture path")
	nodeID := fs.String("node-id", "", "exporting recommendation node id")
	expectedNextNode := fs.String("expected-next-node", "", "expected next node after exporter completion")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceEvidenceRoot) == "" {
		return fmt.Errorf("--source-evidence-root is required")
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*sourceAssertionPath) == "" {
		return fmt.Errorf("--source-assertion is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceEvidenceRoot, *sourceReadbackPath, *sourceAssertionPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
		if strings.TrimSpace(*fixtureOutPath) != "" && samePath(input, *fixtureOutPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	bundle, err := BuildAtlasNextWaveFeatureDepthRecommendations(AtlasNextWaveFeatureDepthExportOptions{
		MissionID:           *missionID,
		SourceEvidenceRoot:  *sourceEvidenceRoot,
		SourceReadbackPath:  *sourceReadbackPath,
		SourceAssertionPath: *sourceAssertionPath,
		MinTasks:            *minTasks,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, bundle); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*fixtureOutPath) != "" {
		if strings.TrimSpace(*nodeID) == "" {
			return fmt.Errorf("--node-id is required with --fixture-out")
		}
		if strings.TrimSpace(*expectedNextNode) == "" {
			return fmt.Errorf("--expected-next-node is required with --fixture-out")
		}
		sourceReadback, err := LoadJSON[AtlasRecommendationReadback](*sourceReadbackPath)
		if err != nil {
			return err
		}
		fixture, err := BuildAtlasNextWaveRecommendationExport(bundle, sourceReadback, *nodeID, *expectedNextNode)
		if err != nil {
			return err
		}
		if err := WriteJSON(*fixtureOutPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, bundle)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nminimum_tasks=%d\nrecommendation_count=%d\nranked_tasks=%d\nsafe_to_execute=%t\nfeature_depth_recommendations=%s\nnext_wave_export_fixture=%s\n",
		bundle.Status,
		bundle.MissionID,
		bundle.MinimumTasks,
		bundle.RecommendationCount,
		len(bundle.Tasks),
		bundle.SafeToExecute,
		filepath.ToSlash(*outPath),
		filepath.ToSlash(*fixtureOutPath),
	)
	return nil
}

func runMissionRecommendationsExportRefactoringWave(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations export-refactoring-wave", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	missionID := fs.String("mission-id", "ao-atlas-refactoring-wave-v01", "refactoring wave mission id")
	sourceEvidenceRoot := fs.String("source-evidence-root", "", "source evidence root")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	sourceAssertionPath := fs.String("source-assertion", "", "source no-promotion/no-RSI assertion path")
	nextTrackDecisionPath := fs.String("next-track-decision", "", "next-track decision path")
	consumedLedgerPath := fs.String("consumed-ledger", "", "consumed recommendation ledger path")
	minTasks := fs.Int("min-tasks", 40, "minimum ranked refactoring tasks")
	outPath := fs.String("out", "", "output refactoring recommendations path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for flagName, value := range map[string]string{
		"--source-evidence-root": *sourceEvidenceRoot,
		"--source-readback":      *sourceReadbackPath,
		"--source-assertion":     *sourceAssertionPath,
		"--next-track-decision":  *nextTrackDecisionPath,
		"--consumed-ledger":      *consumedLedgerPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", flagName)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceEvidenceRoot, *sourceReadbackPath, *sourceAssertionPath, *nextTrackDecisionPath, *consumedLedgerPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	bundle, err := BuildAtlasNextWaveRefactoringRecommendations(AtlasNextWaveRefactoringExportOptions{
		MissionID:                  *missionID,
		SourceEvidenceRoot:         *sourceEvidenceRoot,
		SourceReadbackPath:         *sourceReadbackPath,
		SourceAssertionPath:        *sourceAssertionPath,
		NextTrackDecisionPath:      *nextTrackDecisionPath,
		ConsumedRecommendationPath: *consumedLedgerPath,
		MinTasks:                   *minTasks,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, bundle); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, bundle)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\ntrack=%s\nminimum_tasks=%d\nrecommendation_count=%d\nranked_tasks=%d\nrsi_remains_denied=%t\nrefactoring_recommendations=%s\n",
		bundle.Status,
		bundle.MissionID,
		bundle.Track,
		bundle.MinimumTasks,
		bundle.RecommendationCount,
		len(bundle.Tasks),
		bundle.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsNextTrack(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations next-track", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceEvidenceRoot := fs.String("source-evidence-root", "", "source evidence root")
	readbackPath := fs.String("readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "output next-track decision path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceEvidenceRoot) == "" {
		return fmt.Errorf("--source-evidence-root is required")
	}
	if strings.TrimSpace(*readbackPath) == "" {
		return fmt.Errorf("--readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*readbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	decision, err := BuildAtlasRecommendationNextTrackDecision(*sourceEvidenceRoot, *readbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, decision); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, decision)
	}
	fmt.Fprintf(stdout, "status=%s\ncurrent_track=%s\ncurrent_track_status=%s\nrecommended_track=%s\nrsi_track_status=%s\nnext_track_decision=%s\n",
		decision.Status,
		decision.CurrentTrack,
		decision.CurrentTrackStatus,
		decision.RecommendedTrack,
		decision.RSITrackStatus,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsConsumedLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations consumed-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceEvidenceRoot := fs.String("source-evidence-root", "", "source evidence root")
	readbackPath := fs.String("readback", "", "source recommendation readback path")
	nextTrackDecisionPath := fs.String("next-track-decision", "", "next-track decision path")
	outPath := fs.String("out", "", "output consumed recommendation ledger path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceEvidenceRoot) == "" {
		return fmt.Errorf("--source-evidence-root is required")
	}
	if strings.TrimSpace(*readbackPath) == "" {
		return fmt.Errorf("--readback is required")
	}
	if strings.TrimSpace(*nextTrackDecisionPath) == "" {
		return fmt.Errorf("--next-track-decision is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*readbackPath, *outPath) || samePath(*nextTrackDecisionPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasConsumedRecommendationLedger(*sourceEvidenceRoot, *readbackPath, *nextTrackDecisionPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nconsumed_track=%s\nrecommended_track=%s\nrsi_remains_denied=%t\nconsumed_recommendation_ledger=%s\n",
		ledger.Status,
		ledger.ConsumedTrack,
		ledger.RecommendedTrack,
		ledger.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsTrackRegistry(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations track-registry", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "output recommendation track registry path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	registry, err := DefaultAtlasRecommendationTrackRegistry()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, registry); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, registry)
	}
	fmt.Fprintf(stdout, "status=%s\ndefault_track=%s\nsaturated_feature_depth_next_track=%s\nrsi_remains_denied=%t\nrecommendation_track_registry=%s\n",
		registry.Status,
		registry.DefaultTrack,
		registry.SaturatedFeatureDepthNextTrack,
		registry.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsRunLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations run-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	commandName := fs.String("command", "", "recommendation command name")
	artifactPath := fs.String("artifact", "", "recommendation command output artifact path")
	outPath := fs.String("out", "", "output recommendation command run ledger path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*commandName) == "" {
		return fmt.Errorf("--command is required")
	}
	if strings.TrimSpace(*artifactPath) == "" {
		return fmt.Errorf("--artifact is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*artifactPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasRecommendationCommandRunLedger(*commandName, *artifactPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\ncommand=%s\nartifact_schema=%s\ntyped_validator=%s\noutput_status=%s\nrsi_remains_denied=%t\nrecommendation_command_run_ledger=%s\n",
		ledger.Status,
		ledger.Command,
		ledger.ArtifactSchema,
		ledger.TypedValidator,
		ledger.OutputStatus,
		ledger.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsRunLedgerRollup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations run-ledger-rollup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	ledgerPaths := stringListFlag{}
	fs.Var(&ledgerPaths, "ledger", "recommendation command run ledger path")
	outPath := fs.String("out", "", "output recommendation command run ledger rollup path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(ledgerPaths) == 0 {
		return fmt.Errorf("--ledger is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" {
		for _, ledgerPath := range ledgerPaths {
			if samePath(ledgerPath, *outPath) {
				return fmt.Errorf("refusing to overwrite input artifact")
			}
		}
	}
	rollup, err := BuildAtlasRecommendationCommandRunLedgerRollup(ledgerPaths)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasRecommendationCommandRunLedgerRollup(*outPath, rollup); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, rollup)
	}
	fmt.Fprintf(stdout, "status=%s\nledger_count=%d\nfailed_output_count=%d\nall_ledgers_record_invocation=%t\nall_outputs_no_promotion=%t\nrsi_remains_denied=%t\nrecommendation_command_run_ledger_rollup=%s\n",
		rollup.Status,
		rollup.LedgerCount,
		rollup.FailedOutputCount,
		rollup.AllLedgersRecordInvocation,
		rollup.AllOutputsNoPromotion,
		rollup.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsRunLedgerCoverageCheck(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations run-ledger-coverage-check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	registryPath := fs.String("registry", "", "recommendation evidence schema registry path")
	rollupPath := fs.String("rollup", "", "recommendation command run ledger rollup path")
	outPath := fs.String("out", "", "output recommendation run ledger coverage check path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*registryPath) == "" {
		return fmt.Errorf("--registry is required")
	}
	if strings.TrimSpace(*rollupPath) == "" {
		return fmt.Errorf("--rollup is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*registryPath, *outPath) || samePath(*rollupPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	check, err := BuildAtlasRecommendationRunLedgerCoverageCheck(*registryPath, *rollupPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasRecommendationRunLedgerCoverageCheck(*outPath, check); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, check)
	}
	fmt.Fprintf(stdout, "status=%s\nrequired_command_count=%d\ncovered_command_count=%d\nmissing_command_count=%d\nall_control_plane_commands_covered=%t\nall_outputs_no_promotion=%t\nrsi_remains_denied=%t\nrecommendation_run_ledger_coverage_check=%s\n",
		check.Status,
		check.RequiredCommandCount,
		check.CoveredCommandCount,
		check.MissingCommandCount,
		check.AllControlPlaneCommandsCovered,
		check.AllOutputsNoPromotion,
		check.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFinalResponseGates(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations final-response-gates", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "output recommendation final-response gates path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	gates, err := DefaultAtlasRecommendationFinalResponseGates()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, gates); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, gates)
	}
	fmt.Fprintf(stdout, "status=%s\ngate_count=%d\nfinal_response_allowed_requires_all_gates=%t\nrsi_remains_denied=%t\nrecommendation_final_response_gates=%s\n",
		gates.Status,
		len(gates.Gates),
		gates.FinalResponseAllowedRequiresAllGates,
		gates.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsSchemaRegistry(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations schema-registry", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outPath := fs.String("out", "", "output recommendation evidence schema registry path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	registry, err := DefaultAtlasRecommendationEvidenceSchemaRegistry()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, registry); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, registry)
	}
	fmt.Fprintf(stdout, "status=%s\nschema_count=%d\ntyped_validator_coverage_complete=%t\nrsi_remains_denied=%t\nrecommendation_evidence_schema_registry=%s\n",
		registry.Status,
		registry.SchemaCount,
		registry.TypedValidatorCoverageComplete,
		registry.RSIRemainsDenied,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsSchemaRegistryHealth(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations schema-registry-health", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	evidenceRoot := fs.String("evidence-root", "", "Atlas recommendation evidence root")
	outDir := fs.String("out-dir", "", "output directory for generated schema health artifacts")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outDir) == "" && !*jsonOut {
		return fmt.Errorf("--out-dir or --json is required")
	}
	if strings.TrimSpace(*outDir) == "" {
		return fmt.Errorf("--out-dir is required")
	}
	if samePath(*evidenceRoot, *outDir) {
		return fmt.Errorf("refusing to write schema health artifacts over evidence root")
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}

	registryPath := filepath.Join(*outDir, "recommendation-evidence-schema-registry.json")
	reportPath := filepath.Join(*outDir, "recommendation-evidence-validation-report.json")
	coveragePath := filepath.Join(*outDir, "recommendation-evidence-schema-registry-coverage.json")
	registryLedgerPath := filepath.Join(*outDir, "recommendation-schema-registry-run-ledger.json")
	reportLedgerPath := filepath.Join(*outDir, "recommendation-validation-report-run-ledger.json")
	coverageLedgerPath := filepath.Join(*outDir, "recommendation-schema-registry-coverage-run-ledger.json")

	registry, err := DefaultAtlasRecommendationEvidenceSchemaRegistry()
	if err != nil {
		return err
	}
	if err := WriteJSON(registryPath, registry); err != nil {
		return err
	}
	report, reportErr := BuildAtlasRecommendationEvidenceValidationReport(*evidenceRoot)
	if err := WriteJSON(reportPath, report); err != nil {
		return err
	}
	coverage, coverageErr := BuildAtlasRecommendationEvidenceSchemaRegistryCoverage(registryPath, reportPath)
	if err := WriteJSON(coveragePath, coverage); err != nil {
		return err
	}
	registryLedger, err := BuildAtlasRecommendationCommandRunLedger("schema-registry", registryPath)
	if err != nil {
		return err
	}
	if err := WriteJSON(registryLedgerPath, registryLedger); err != nil {
		return err
	}
	reportLedger, err := BuildAtlasRecommendationCommandRunLedger("validate-evidence", reportPath)
	if err != nil {
		return err
	}
	if err := WriteJSON(reportLedgerPath, reportLedger); err != nil {
		return err
	}
	coverageLedger, err := BuildAtlasRecommendationCommandRunLedger("schema-registry-coverage", coveragePath)
	if err != nil {
		return err
	}
	if err := WriteJSON(coverageLedgerPath, coverageLedger); err != nil {
		return err
	}
	runLedgerPaths := []string{registryLedgerPath, reportLedgerPath, coverageLedgerPath}
	runLedgerCount := len(runLedgerPaths)
	allOutputsHaveRunLedgers := runLedgerCount == 3 &&
		registryLedger.ArtifactPath == filepath.ToSlash(registryPath) &&
		reportLedger.ArtifactPath == filepath.ToSlash(reportPath) &&
		coverageLedger.ArtifactPath == filepath.ToSlash(coveragePath)
	operatorSummary := fmt.Sprintf("%s: validation report %s; %d registry schemas missing; %d registry validators missing; %d run ledgers written; RSI remains denied",
		coverage.Status,
		coverage.ValidationReportStatus,
		len(coverage.MissingSchemas),
		len(coverage.MissingValidators),
		runLedgerCount,
	)
	exactNextAction := "Add missing recommendation control-plane evidence artifacts, rerun schema-registry-health, and keep promotion denied."
	if *jsonOut {
		if err := printJSON(stdout, map[string]any{
			"status":                                           coverage.Status,
			"validation_report_status":                         coverage.ValidationReportStatus,
			"registry_schema_count":                            coverage.RegistrySchemaCount,
			"covered_schema_count":                             coverage.CoveredSchemaCount,
			"missing_schemas":                                  len(coverage.MissingSchemas),
			"missing_validators":                               len(coverage.MissingValidators),
			"stale_registry_entries":                           coverage.StaleRegistryEntryCount,
			"failure_reasons":                                  coverage.FailureReasons,
			"rsi_remains_denied":                               coverage.RSIRemainsDenied,
			"run_ledger_count":                                 runLedgerCount,
			"all_outputs_have_run_ledgers":                     allOutputsHaveRunLedgers,
			"operator_summary":                                 operatorSummary,
			"exact_next_action":                                exactNextAction,
			"recommendation_evidence_schema_registry":          filepath.ToSlash(registryPath),
			"recommendation_evidence_validation_report":        filepath.ToSlash(reportPath),
			"recommendation_evidence_schema_registry_coverage": filepath.ToSlash(coveragePath),
			"schema_registry_run_ledger":                       filepath.ToSlash(registryLedgerPath),
			"validation_report_run_ledger":                     filepath.ToSlash(reportLedgerPath),
			"schema_registry_coverage_run_ledger":              filepath.ToSlash(coverageLedgerPath),
		}); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(stdout, "status=%s\nvalidation_report_status=%s\nregistry_schema_count=%d\ncovered_schema_count=%d\nmissing_schemas=%d\nmissing_validators=%d\nstale_registry_entries=%d\nfailure_reasons=%s\nrsi_remains_denied=%t\nrun_ledger_count=%d\nall_outputs_have_run_ledgers=%t\noperator_summary=%s\nexact_next_action=%s\nrecommendation_evidence_schema_registry=%s\nrecommendation_evidence_validation_report=%s\nrecommendation_evidence_schema_registry_coverage=%s\nschema_registry_run_ledger=%s\nvalidation_report_run_ledger=%s\nschema_registry_coverage_run_ledger=%s\n",
			coverage.Status,
			coverage.ValidationReportStatus,
			coverage.RegistrySchemaCount,
			coverage.CoveredSchemaCount,
			len(coverage.MissingSchemas),
			len(coverage.MissingValidators),
			coverage.StaleRegistryEntryCount,
			strings.Join(coverage.FailureReasons, ","),
			coverage.RSIRemainsDenied,
			runLedgerCount,
			allOutputsHaveRunLedgers,
			operatorSummary,
			exactNextAction,
			filepath.ToSlash(registryPath),
			filepath.ToSlash(reportPath),
			filepath.ToSlash(coveragePath),
			filepath.ToSlash(registryLedgerPath),
			filepath.ToSlash(reportLedgerPath),
			filepath.ToSlash(coverageLedgerPath),
		)
	}
	if coverageErr != nil {
		return coverageErr
	}
	return reportErr
}

func runMissionRecommendationsSchemaRegistryCoverage(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations schema-registry-coverage", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	registryPath := fs.String("registry", "", "recommendation evidence schema registry path")
	validationReportPath := fs.String("validation-report", "", "recommendation evidence validation report path")
	outPath := fs.String("out", "", "output recommendation evidence schema registry coverage path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--registry":          *registryPath,
		"--validation-report": *validationReportPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*registryPath, *outPath) || samePath(*validationReportPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	coverage, err := BuildAtlasRecommendationEvidenceSchemaRegistryCoverage(*registryPath, *validationReportPath)
	if strings.TrimSpace(*outPath) != "" {
		if writeErr := WriteJSON(*outPath, coverage); writeErr != nil {
			return writeErr
		}
	}
	if *jsonOut {
		if printErr := printJSON(stdout, coverage); printErr != nil {
			return printErr
		}
	} else {
		fmt.Fprintf(stdout, "status=%s\nvalidation_report_status=%s\nregistry_schema_count=%d\ncovered_schema_count=%d\nmissing_schemas=%d\nmissing_validators=%d\nstale_registry_entries=%d\nfailure_reasons=%s\nrsi_remains_denied=%t\nrecommendation_evidence_schema_registry_coverage=%s\n",
			coverage.Status,
			coverage.ValidationReportStatus,
			coverage.RegistrySchemaCount,
			coverage.CoveredSchemaCount,
			len(coverage.MissingSchemas),
			len(coverage.MissingValidators),
			coverage.StaleRegistryEntryCount,
			strings.Join(coverage.FailureReasons, ","),
			coverage.RSIRemainsDenied,
			filepath.ToSlash(*outPath),
		)
	}
	return err
}

func runMissionRecommendationsSchemaHealthRepairPrompt(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations schema-health-repair-prompt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	coveragePath := fs.String("coverage", "", "schema registry coverage artifact path")
	nodeID := fs.String("node-id", "", "repair prompt node id")
	promptOut := fs.String("prompt-out", "", "schema health repair prompt markdown output path")
	fixtureOut := fs.String("fixture-out", "", "schema health repair prompt fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--coverage":    *coveragePath,
		"--node-id":     *nodeID,
		"--prompt-out":  *promptOut,
		"--fixture-out": *fixtureOut,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if samePath(*coveragePath, *promptOut) || samePath(*coveragePath, *fixtureOut) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	coverage, err := LoadJSON[AtlasRecommendationEvidenceSchemaRegistryCoverage](*coveragePath)
	if err != nil {
		return err
	}
	fixture, prompt, err := BuildAtlasSchemaHealthRepairPrompt(coverage, AtlasSchemaHealthRepairPromptOptions{
		NodeID:             *nodeID,
		SourceCoveragePath: *coveragePath,
		PromptPath:         *promptOut,
	})
	if err != nil {
		return err
	}
	if err := WriteAtlasSchemaHealthRepairPrompt(*promptOut, *fixtureOut, fixture, prompt); err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncoverage_status=%s\nvalidation_report_status=%s\nmissing_schemas=%d\nmissing_validators=%d\nsafe_to_execute=%t\nrsi_remains_denied=%t\nschema_health_repair_prompt=%s\nschema_health_repair_fixture=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.CoverageStatus,
		fixture.ValidationReportStatus,
		fixture.MissingSchemaCount,
		fixture.MissingValidatorCount,
		fixture.SafeToExecute,
		fixture.RSIRemainsDenied,
		filepath.ToSlash(*promptOut),
		filepath.ToSlash(*fixtureOut),
	)
	return nil
}

func runMissionRecommendationsReadbackDelta(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations readback-delta", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	targetReadbackPath := fs.String("target-readback", "", "target recommendation readback path")
	outPath := fs.String("out", "", "mission readback delta output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*targetReadbackPath) == "" {
		return fmt.Errorf("--target-readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*sourceReadbackPath, *outPath) || samePath(*targetReadbackPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	delta, err := BuildAtlasMissionReadbackDelta(*sourceReadbackPath, *targetReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, delta); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, delta)
	}
	fmt.Fprintf(stdout, "status=%s\nchanged_fields=%d\nsource_readback=%s\ntarget_readback=%s\nmission_readback_delta=%s\n",
		delta.Status,
		len(delta.ChangedFields),
		delta.SourceReadbackPath,
		delta.TargetReadbackPath,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsReadbackDiffFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations readback-diff-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	targetReadbackPath := fs.String("target-readback", "", "target recommendation readback path")
	deltaPath := fs.String("delta", "", "mission readback delta path")
	outPath := fs.String("out", "", "resumable readback diff fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*targetReadbackPath) == "" {
		return fmt.Errorf("--target-readback is required")
	}
	if strings.TrimSpace(*deltaPath) == "" {
		return fmt.Errorf("--delta is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" &&
		(samePath(*sourceReadbackPath, *outPath) || samePath(*targetReadbackPath, *outPath) || samePath(*deltaPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasMissionReadbackDiffFixture(*sourceReadbackPath, *targetReadbackPath, *deltaPath)
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
	fmt.Fprintf(stdout, "status=%s\ncompleted_delta=%d\nready_delta=%d\ncheckpoint_delta=%d\nsource_readback=%s\ntarget_readback=%s\nreadback_diff_fixture=%s\n",
		fixture.Status,
		fixture.CompletedNodeTransition.Delta,
		fixture.ReadyNodeTransition.Delta,
		fixture.CheckpointTransition.Delta,
		fixture.SourceReadbackPath,
		fixture.TargetReadbackPath,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsStaleCheckpointRejection(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations stale-checkpoint-rejection", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	staleReadbackPath := fs.String("stale-readback", "", "stale recommendation readback path")
	latestReadbackPath := fs.String("latest-readback", "", "latest recommendation readback path")
	promptReadbackPath := fs.String("prompt-readback", "", "readback path referenced by continuation prompt")
	outPath := fs.String("out", "", "stale checkpoint rejection output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*staleReadbackPath) == "" {
		return fmt.Errorf("--stale-readback is required")
	}
	if strings.TrimSpace(*latestReadbackPath) == "" {
		return fmt.Errorf("--latest-readback is required")
	}
	if strings.TrimSpace(*promptReadbackPath) == "" {
		return fmt.Errorf("--prompt-readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" &&
		(samePath(*staleReadbackPath, *outPath) || samePath(*latestReadbackPath, *outPath) || samePath(*promptReadbackPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasMissionStaleCheckpointRejection(*staleReadbackPath, *latestReadbackPath, *promptReadbackPath)
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
	fmt.Fprintf(stdout, "status=%s\nrejection_reason=%s\nprompt_next_node=%s\nexpected_current_next_node=%s\nstale_readback=%s\nlatest_readback=%s\nstale_checkpoint_rejection=%s\n",
		fixture.Status,
		fixture.RejectionReason,
		fixture.PromptNextExecutableNode,
		fixture.ExpectedCurrentNextExecutableNode,
		fixture.StaleReadbackPath,
		fixture.LatestReadbackPath,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsOperatorSummaryCheck(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations operator-summary-check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	readbackPath := fs.String("readback", "", "source recommendation readback path")
	summaryOutPath := fs.String("summary-out", "", "operator summary markdown output path")
	outPath := fs.String("out", "", "operator summary check output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*readbackPath) == "" {
		return fmt.Errorf("--readback is required")
	}
	if strings.TrimSpace(*summaryOutPath) == "" {
		return fmt.Errorf("--summary-out is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*summaryOutPath) != "" && samePath(*readbackPath, *summaryOutPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*readbackPath, *outPath) || samePath(*summaryOutPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](*readbackPath)
	if err != nil {
		return err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return err
	}
	if err := WriteAtlasMissionOperatorSummary(*summaryOutPath, readback); err != nil {
		return err
	}
	fixture, err := BuildAtlasMissionOperatorSummaryCheck(*readbackPath, *summaryOutPath)
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
	fmt.Fprintf(stdout, "status=%s\nexact_next_action_occurrences=%d\nfirst_executable_node=%s\noperator_summary=%s\noperator_summary_check=%s\n",
		fixture.Status,
		fixture.ExactNextActionOccurrences,
		fixture.FirstExecutableNode,
		filepath.ToSlash(*summaryOutPath),
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsRunLinkSchemaCoverage(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations run-link-schema-coverage", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	evidenceRoot := fs.String("evidence-root", "", "Atlas recommendation evidence root")
	outPath := fs.String("out", "", "run-link schema coverage output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*evidenceRoot, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	coverage, err := BuildAtlasRunLinkSchemaCoverage(*evidenceRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, coverage); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, coverage)
	}
	fmt.Fprintf(stdout, "status=%s\nrun_link_count=%d\ntyped_run_link_validators=%d\nrun_link_schema_coverage=%s\n",
		coverage.Status,
		coverage.RunLinkCount,
		coverage.ValidatorCounts["typed:run-link"],
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsSchemaValidatorDrift(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations schema-validator-drift", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceReportPath := fs.String("source-report", "", "source recommendation evidence validation report path")
	targetReportPath := fs.String("target-report", "", "target recommendation evidence validation report path")
	outPath := fs.String("out", "", "schema validator drift output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--source-report": *sourceReportPath,
		"--target-report": *targetReportPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*sourceReportPath, *outPath) || samePath(*targetReportPath, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	drift, err := BuildAtlasSchemaValidatorDriftEvidence(*sourceReportPath, *targetReportPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, drift); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, drift)
	}
	fmt.Fprintf(stdout, "status=%s\njson_file_delta=%d\ntyped_validator_delta=%d\ngeneric_schema_delta=%d\nlost_schemas=%d\nlost_validators=%d\nschema_validator_drift=%s\n",
		drift.Status,
		drift.JSONFileDelta,
		drift.TypedValidatorDelta,
		drift.GenericSchemaDelta,
		len(drift.LostSchemas),
		len(drift.LostValidators),
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPRCITimingSummary(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations pr-ci-timing-summary", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	ledgerPath := fs.String("ledger", "", "PR/CI timing ledger path")
	outPath := fs.String("out", "", "PR/CI timing summary output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*ledgerPath) == "" {
		return fmt.Errorf("--ledger is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*ledgerPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	summary, err := BuildAtlasPRCITimingSummary(*ledgerPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, summary); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, summary)
	}
	fmt.Fprintf(stdout, "status=%s\nrow_count=%d\nmax_windows_seconds=%d\nmax_check_seconds=%d\npr_ci_timing_summary=%s\n",
		summary.Status,
		summary.RowCount,
		summary.MaxWindowsSeconds,
		summary.MaxCheckSeconds,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPRCIWindowsThreshold(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations pr-ci-windows-threshold", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	summaryPath := fs.String("summary", "", "PR/CI timing summary path")
	thresholdSeconds := fs.Int("threshold-seconds", 0, "Windows long-running check threshold in seconds")
	outPath := fs.String("out", "", "PR/CI Windows threshold evidence output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*summaryPath) == "" {
		return fmt.Errorf("--summary is required")
	}
	if *thresholdSeconds <= 0 {
		return fmt.Errorf("--threshold-seconds must be greater than zero")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*summaryPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	evidence, err := BuildAtlasPRCIWindowsThresholdEvidence(*summaryPath, *thresholdSeconds)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, evidence); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, evidence)
	}
	fmt.Fprintf(stdout, "status=%s\nthreshold_seconds=%d\nrow_count=%d\nlong_running_windows_checks=%d\npr_ci_windows_threshold_evidence=%s\n",
		evidence.Status,
		evidence.ThresholdSeconds,
		evidence.RowCount,
		evidence.LongRunningWindowsChecks,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFailedCheckReplay(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations failed-check-replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "failed check replay input path")
	outPath := fs.String("out", "", "failed check replay fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasFailedCheckReplayFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\ncase_count=%d\nmerge_denied_cases=%d\nretry_allowed_cases=%d\nfailed_check_replay_fixture=%s\n",
		fixture.Status,
		fixture.CaseCount,
		fixture.MergeDeniedCases,
		fixture.RetryAllowedCases,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandCovenantRejectedTicketFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-covenant-rejected-ticket-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "Command/Covenant rejected ticket input path")
	outPath := fs.String("out", "", "Command/Covenant rejected ticket fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasCommandCovenantRejectedTicketFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\ncommand_accepts_ticket=%t\nreason_preserved=%t\ncommand_covenant_rejected_ticket_fixture=%s\n",
		fixture.Status,
		fixture.CommandAcceptsTicket,
		fixture.ReasonPreserved,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandCovenantQuarantineFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-covenant-quarantine-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "Command/Covenant quarantine input path")
	outPath := fs.String("out", "", "Command/Covenant quarantine fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasCommandCovenantQuarantineFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\nquarantined_paths=%d\nsafe_to_accept=%t\ncommand_covenant_quarantine_fixture=%s\n",
		fixture.Status,
		fixture.QuarantinedPaths,
		fixture.SafeToAccept,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandTicketBytePreservationFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-ticket-byte-preservation-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "Command ticket byte preservation input path")
	outPath := fs.String("out", "", "Command ticket byte preservation fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasCommandTicketBytePreservationFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\nbyte_preservation_passed=%t\ncase_count=%d\ncommand_ticket_byte_preservation_fixture=%s\n",
		fixture.Status,
		fixture.BytePreservationPassed,
		fixture.CaseCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsTicketDigestReadbackBindingFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations ticket-digest-readback-binding-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "ticket digest readback binding input path")
	outPath := fs.String("out", "", "ticket digest readback binding fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasTicketDigestReadbackBindingFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\ndigest_binding_passed=%t\ncase_count=%d\nticket_digest_readback_binding_fixture=%s\n",
		fixture.Status,
		fixture.DigestBindingPassed,
		fixture.CaseCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPolicyHashMismatchRejectionFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations policy-hash-mismatch-rejection-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "policy hash mismatch rejection input path")
	outPath := fs.String("out", "", "policy hash mismatch rejection fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasPolicyHashMismatchRejectionFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\nrejected_cases=%d\nsafe_to_accept=%t\npolicy_hash_mismatch_rejection_fixture=%s\n",
		fixture.Status,
		fixture.RejectedCases,
		fixture.SafeToAccept,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPolicyVersionReplayRejectionFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations policy-version-replay-rejection-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "policy version replay rejection input path")
	outPath := fs.String("out", "", "policy version replay rejection fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasPolicyVersionReplayRejectionFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\nrejected_cases=%d\nsafe_to_accept=%t\npolicy_version_replay_rejection_fixture=%s\n",
		fixture.Status,
		fixture.RejectedCases,
		fixture.SafeToAccept,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCovenantEvidenceDigestReadbackFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations covenant-evidence-digest-readback-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "covenant evidence digest readback input path")
	outPath := fs.String("out", "", "covenant evidence digest readback fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasCovenantEvidenceDigestReadbackFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\ndigest_readback_complete=%t\ncase_count=%d\ncovenant_evidence_digest_readback_fixture=%s\n",
		fixture.Status,
		fixture.DigestReadbackComplete,
		fixture.CaseCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandCompactRejectionReasonFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-compact-rejection-reason-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "command compact rejection reason input path")
	outPath := fs.String("out", "", "command compact rejection reason fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasCommandCompactRejectionReasonFixture(*inputPath)
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
	fmt.Fprintf(stdout, "status=%s\nreasons_rendered=%t\ncase_count=%d\ncommand_compact_rejection_reason_fixture=%s\n",
		fixture.Status,
		fixture.ReasonsRendered,
		fixture.CaseCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsBlueprintTicketSchemaCompatibilityLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations blueprint-ticket-schema-compatibility-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "blueprint ticket schema compatibility ledger input path")
	outPath := fs.String("out", "", "blueprint ticket schema compatibility ledger output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasBlueprintTicketSchemaCompatibilityLedger(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nall_entries_compatible=%t\nentry_count=%d\nblueprint_ticket_schema_compatibility_ledger=%s\n",
		ledger.Status,
		ledger.AllEntriesCompatible,
		ledger.EntryCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsAtlasTicketSchemaCompatibilityLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations atlas-ticket-schema-compatibility-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "atlas ticket schema compatibility ledger input path")
	outPath := fs.String("out", "", "atlas ticket schema compatibility ledger output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasTicketSchemaCompatibilityLedger(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nall_entries_compatible=%t\nentry_count=%d\natlas_ticket_schema_compatibility_ledger=%s\n",
		ledger.Status,
		ledger.AllEntriesCompatible,
		ledger.EntryCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFoundryTicketSchemaCompatibilityLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-ticket-schema-compatibility-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "foundry ticket schema compatibility ledger input path")
	outPath := fs.String("out", "", "foundry ticket schema compatibility ledger output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasFoundryTicketSchemaCompatibilityLedger(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nall_entries_compatible=%t\nentry_count=%d\nfoundry_ticket_schema_compatibility_ledger=%s\n",
		ledger.Status,
		ledger.AllEntriesCompatible,
		ledger.EntryCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandTicketSchemaCompatibilityLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-ticket-schema-compatibility-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "command ticket schema compatibility ledger input path")
	outPath := fs.String("out", "", "command ticket schema compatibility ledger output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasCommandTicketSchemaCompatibilityLedger(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nall_entries_compatible=%t\nentry_count=%d\ncommand_ticket_schema_compatibility_ledger=%s\n",
		ledger.Status,
		ledger.AllEntriesCompatible,
		ledger.EntryCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCovenantTicketSchemaAuthorityLedger(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations covenant-ticket-schema-authority-ledger", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "covenant ticket schema authority ledger input path")
	outPath := fs.String("out", "", "covenant ticket schema authority ledger output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	ledger, err := BuildAtlasCovenantTicketSchemaAuthorityLedger(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, ledger); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, ledger)
	}
	fmt.Fprintf(stdout, "status=%s\nall_entries_compatible=%t\nentry_count=%d\ncovenant_ticket_schema_authority_ledger=%s\n",
		ledger.Status,
		ledger.AllEntriesCompatible,
		ledger.EntryCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPolicyTicketPublicSafetyScan(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations policy-ticket-public-safety-scan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "policy ticket public-safety scan input path")
	outPath := fs.String("out", "", "policy ticket public-safety scan output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	scan, err := BuildAtlasPolicyTicketPublicSafetyScan(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, scan); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, scan)
	}
	fmt.Fprintf(stdout, "status=%s\nunsafe_claims_found=%d\nclaim_count=%d\npolicy_ticket_public_safety_scan=%s\n",
		scan.Status,
		scan.UnsafeClaimsFound,
		scan.ClaimCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsMergeCheckBinding(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations merge-check-binding", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "merge check binding input path")
	outPath := fs.String("out", "", "merge check binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	binding, err := BuildAtlasMergeCheckBinding(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nrow_count=%d\npassed_required_check_rows=%d\nunbound_merge_commits=%d\nmerge_check_binding=%s\n",
		binding.Status,
		binding.RowCount,
		binding.PassedRequiredCheckRows,
		binding.UnboundMergeCommits,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPostMergeBranchDeletionReadback(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations post-merge-branch-deletion-readback", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	evidenceRoot := fs.String("evidence-root", "", "Atlas recommendation evidence root")
	outPath := fs.String("out", "", "post-merge branch deletion readback output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	readback, err := BuildAtlasPostMergeBranchDeletionReadback(*evidenceRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, readback); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, readback)
	}
	fmt.Fprintf(stdout, "status=%s\npost_merge_lifecycle_count=%d\nlocal_branch_deleted_count=%d\nremote_branch_deleted_count=%d\nbranches_remaining_total=%d\npost_merge_branch_deletion_readback=%s\n",
		readback.Status,
		readback.PostMergeLifecycleCount,
		readback.LocalBranchDeletedCount,
		readback.RemoteBranchDeletedCount,
		readback.BranchesRemainingTotal,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsStaleRemoteBranchRepair(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations stale-remote-branch-repair", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "stale remote branch repair input path")
	outPath := fs.String("out", "", "stale remote branch repair output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	repair, err := BuildAtlasStaleRemoteBranchRepair(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, repair); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, repair)
	}
	fmt.Fprintf(stdout, "status=%s\ncase_count=%d\nrepair_required_cases=%d\ncleanup_safe_cases=%d\nblocked_cases=%d\nstale_remote_branch_repair=%s\n",
		repair.Status,
		repair.CaseCount,
		repair.RepairRequiredCases,
		repair.CleanupSafeCases,
		repair.BlockedCases,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsLocalMainSyncReadback(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations local-main-sync-readback", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "local main sync readback input path")
	outPath := fs.String("out", "", "local main sync readback output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*inputPath) == "" {
		return fmt.Errorf("--input is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*inputPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := BuildAtlasLocalMainSyncReadback(*inputPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, readback); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, readback)
	}
	fmt.Fprintf(stdout, "status=%s\nlocal_main_synced=%t\nworking_tree_clean=%t\ncodex_branch_cleanup_confirmed=%t\nsafe_to_select_next_node=%t\ndenial_case_count=%d\nlocal_main_sync_readback=%s\n",
		readback.Status,
		readback.LocalMainSynced,
		readback.WorkingTreeClean,
		readback.CodexBranchCleanupConfirmed,
		readback.SafeToSelectNextNode,
		readback.DenialCaseCount,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsBranchCleanupHandoffSummary(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations branch-cleanup-handoff-summary", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	evidenceRoot := fs.String("evidence-root", "", "feature depth evidence root")
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "branch cleanup handoff summary output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*sourceReadback) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	summary, err := BuildAtlasBranchCleanupHandoffSummary(*evidenceRoot, *sourceReadback)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, summary); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, summary)
	}
	fmt.Fprintf(stdout, "status=%s\npost_merge_lifecycle_count=%d\nmerged_and_cleaned_count=%d\npassed_ci_count=%d\ncleanup_complete=%t\noperator_handoff_status=%s\nbranch_cleanup_handoff_summary=%s\n",
		summary.Status,
		summary.PostMergeLifecycleCount,
		summary.MergedAndCleanedCount,
		summary.PassedCICount,
		summary.CleanupComplete,
		summary.OperatorHandoffStatus,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCompactionResumePrompt(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations compaction-resume-prompt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	workgraphPath := fs.String("workgraph", "", "current workgraph path")
	leaseStartPath := fs.String("lease-start", "", "lease start marker path")
	checkpointReadbackPath := fs.String("checkpoint-readback", "", "checkpoint readback path to bind into the resume prompt")
	evidenceRoot := fs.String("evidence-root", "", "portable evidence root")
	nodeID := fs.String("node-id", "", "current resume prompt node id")
	expectedNextNode := fs.String("expected-next-node-after-completion", "", "expected next node after completing the active node")
	promptOut := fs.String("prompt-out", "", "compaction resume prompt markdown output path")
	fixtureOut := fs.String("fixture-out", "", "compaction resume prompt fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--source-readback": *sourceReadback,
		"--workgraph":       *workgraphPath,
		"--lease-start":     *leaseStartPath,
		"--node-id":         *nodeID,
		"--prompt-out":      *promptOut,
		"--fixture-out":     *fixtureOut,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if samePath(*sourceReadback, *promptOut) || samePath(*sourceReadback, *fixtureOut) || samePath(*workgraphPath, *fixtureOut) || samePath(*leaseStartPath, *fixtureOut) || samePath(*checkpointReadbackPath, *fixtureOut) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](*sourceReadback)
	if err != nil {
		return err
	}
	fixture, prompt, err := BuildAtlasCompactionResumePromptFixture(readback, AtlasCompactionResumePromptOptions{
		NodeID:                          *nodeID,
		SourceReadbackPath:              *sourceReadback,
		PromptPath:                      *promptOut,
		LeaseStartPath:                  *leaseStartPath,
		WorkgraphPath:                   *workgraphPath,
		CheckpointReadbackPath:          *checkpointReadbackPath,
		EvidenceRoot:                    *evidenceRoot,
		ExpectedNextNodeAfterCompletion: *expectedNextNode,
	})
	if err != nil {
		return err
	}
	if err := WriteAtlasCompactionResumePrompt(*promptOut, *fixtureOut, fixture, prompt); err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes=%d\nready_nodes=%d\nfirst_executable_node=%s\nelapsed_minutes=%d\nfinal_response_allowed=%t\ncompaction_resume_prompt=%s\ncompaction_resume_fixture=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.CompletedNodes,
		fixture.ReadyNodes,
		fixture.FirstExecutableNode,
		fixture.ElapsedMinutes,
		fixture.FinalResponseAllowed,
		filepath.ToSlash(*promptOut),
		filepath.ToSlash(*fixtureOut),
	)
	return nil
}

func runMissionRecommendationsCompactionResumeRegression(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations compaction-resume-regression", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sourcePromptFixture := fs.String("source-prompt-fixture", "", "source compaction resume prompt fixture path")
	sourcePromptMarkdown := fs.String("source-prompt-markdown", "", "source compaction resume prompt markdown path")
	sourceReadback := fs.String("source-readback", "", "source recommendation readback path")
	nodeID := fs.String("node-id", "", "current regression node id")
	expectedNextNode := fs.String("expected-next-node-after-completion", "", "expected next node after completing the active node")
	outPath := fs.String("out", "", "compaction resume regression output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--source-prompt-fixture":  *sourcePromptFixture,
		"--source-prompt-markdown": *sourcePromptMarkdown,
		"--source-readback":        *sourceReadback,
		"--node-id":                *nodeID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourcePromptFixture, *sourcePromptMarkdown, *sourceReadback} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	regression, err := BuildAtlasCompactionResumeRegression(AtlasCompactionResumeRegressionOptions{
		NodeID:                          *nodeID,
		SourcePromptFixturePath:         *sourcePromptFixture,
		SourcePromptMarkdownPath:        *sourcePromptMarkdown,
		SourceReadbackPath:              *sourceReadback,
		ExpectedNextNodeAfterCompletion: *expectedNextNode,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasCompactionResumeRegression(*outPath, regression); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, regression)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes_before=%d\nready_nodes_before=%d\nfirst_executable_node_before=%s\nexact_next_action_preserved=%t\nfinal_response_allowed_before=%t\ncompaction_resume_regression=%s\n",
		regression.Status,
		regression.NodeID,
		regression.CompletedNodesBefore,
		regression.ReadyNodesBefore,
		regression.FirstExecutableNodeBefore,
		regression.SourcePromptExactActionPreserved,
		regression.FinalResponseAllowedBefore,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsResumeDenialEvidence(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations resume-denial-evidence", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	readbackPath := fs.String("readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "resume denial evidence output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*readbackPath) == "" {
		return fmt.Errorf("--readback is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*readbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	evidence, err := BuildAtlasResumeDenialEvidence(*readbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasResumeDenialEvidence(*outPath, evidence); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, evidence)
	}
	fmt.Fprintf(stdout, "status=%s\nready_nodes=%d\ncurrent_next_executable_node=%s\nfinal_response_allowed=%t\nresume_denial_evidence=%s\n",
		evidence.Status,
		evidence.ReadyNodes,
		evidence.CurrentNextExecutableNode,
		evidence.FinalResponseAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPublicSafetyReadbackBinding(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations public-safety-readback-binding", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	readbackPath := fs.String("readback", "", "source recommendation readback path")
	sentinelPath := fs.String("sentinel", "", "Sentinel public-safety evidence path")
	verificationPath := fs.String("verification", "", "verification summary evidence path")
	nodeID := fs.String("node-id", "", "binding node id")
	outPath := fs.String("out", "", "public-safety readback binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--readback":     *readbackPath,
		"--sentinel":     *sentinelPath,
		"--verification": *verificationPath,
		"--node-id":      *nodeID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*readbackPath, *sentinelPath, *verificationPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasPublicSafetyReadbackBinding(*readbackPath, *sentinelPath, *verificationPath, *nodeID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasPublicSafetyReadbackBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nbound_public_safety_scan_status=%s\nprevious_public_safety_scan_status=%s\nready_nodes_after_binding=%d\nfinal_response_allowed_after_binding=%t\npublic_safety_readback_binding=%s\n",
		binding.Status,
		binding.NodeID,
		binding.BoundPublicSafetyScanStatus,
		binding.PreviousPublicSafetyScanStatus,
		binding.ReadyNodesAfterBinding,
		binding.FinalResponseAllowedAfter,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsScopedPublicSafetyScan(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations scoped-public-safety-scan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "scan node id")
	outPath := fs.String("out", "", "scoped public-safety scan output path")
	jsonOut := fs.Bool("json", false, "json output")
	scopeFlags := stringListFlag{}
	fs.Var(&scopeFlags, "scope", "file or directory scope to scan")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if len(scopeFlags) == 0 {
		return fmt.Errorf("--scope is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, scope := range scopeFlags {
		if strings.TrimSpace(*outPath) != "" && samePath(scope, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	scan, err := BuildAtlasScopedPublicSafetyScan(*nodeID, scopeFlags)
	if err != nil && strings.TrimSpace(scan.Schema) == "" {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if writeErr := WriteAtlasScopedPublicSafetyScan(*outPath, scan); writeErr != nil {
			return writeErr
		}
	}
	if *jsonOut {
		if printErr := printJSON(stdout, scan); printErr != nil {
			return printErr
		}
	} else {
		fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nscanned_files=%d\nchanged_evidence_files=%d\nchanged_prompt_artifacts=%d\nunsafe_match_count=%d\npublic_safety_scan_passed=%t\nscoped_public_safety_scan=%s\n",
			scan.Status,
			scan.NodeID,
			scan.ScannedFileCount,
			scan.ChangedEvidenceFiles,
			scan.ChangedPromptArtifacts,
			scan.UnsafeMatchCount,
			scan.PublicSafetyScanPassed,
			filepath.ToSlash(*outPath),
		)
	}
	return err
}

func runMissionRecommendationsAuthorityPromotionNegativeFixtures(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations authority-promotion-negative-fixtures", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "negative fixture node id")
	outPath := fs.String("out", "", "authority promotion negative fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	fixture, err := BuildAtlasAuthorityPromotionNegativeFixtures(*nodeID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasAuthorityPromotionNegativeFixtures(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncase_count=%d\nfixture_encoding=%s\nunsafe_literal_stored=%t\nexpected_scan_status=%s\nauthority_promotion_negative_fixtures=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.CaseCount,
		fixture.FixtureEncoding,
		fixture.UnsafeLiteralStored,
		fixture.ExpectedScanStatus,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPublicSafetyCoverageRollup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations public-safety-coverage-rollup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "rollup node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	evidenceRoot := fs.String("evidence-root", "", "recommendation evidence root")
	outPath := fs.String("out", "", "public safety coverage rollup output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && (samePath(*sourceReadbackPath, *outPath) || samePath(*evidenceRoot, *outPath)) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	rollup, err := BuildAtlasPublicSafetyCoverageRollup(*nodeID, *sourceReadbackPath, *evidenceRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasPublicSafetyCoverageRollup(*outPath, rollup); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, rollup)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes_before=%d\nsentinel_evidence_count=%d\nscoped_scan_count=%d\nunsafe_match_count_total=%d\npublic_safety_scan_passed=%t\npublic_safety_coverage_rollup=%s\n",
		rollup.Status,
		rollup.NodeID,
		rollup.CompletedNodesBefore,
		rollup.SentinelEvidenceCount,
		rollup.ScopedScanCount,
		rollup.UnsafeMatchCountTotal,
		rollup.PublicSafetyScanPassed,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPromoterNoPromotionRollup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations promoter-no-promotion-rollup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "rollup node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	var evidenceRoots stringListFlag
	fs.Var(&evidenceRoots, "evidence-root", "recommendation evidence root; repeat for multiple roots")
	outPath := fs.String("out", "", "Promoter no-promotion rollup output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nodeID) == "" {
		return fmt.Errorf("--node-id is required")
	}
	if strings.TrimSpace(*sourceReadbackPath) == "" {
		return fmt.Errorf("--source-readback is required")
	}
	if len(evidenceRoots) == 0 {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceReadbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	rollup, err := BuildAtlasPromoterNoPromotionRollup(*nodeID, *sourceReadbackPath, evidenceRoots)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasPromoterNoPromotionRollup(*outPath, rollup); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, rollup)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncompleted_nodes_total=%d\npromoter_no_promotion_files=%d\nmissing_promoter_nodes_total=%d\nno_promotion_invariant_holds=%t\npromoter_no_promotion_rollup=%s\n",
		rollup.Status,
		rollup.NodeID,
		rollup.CompletedNodesTotal,
		rollup.PromoterNoPromotionFiles,
		rollup.MissingPromoterNodesTotal,
		rollup.NoPromotionInvariantHolds,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandPromoterAgreementRollup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-promoter-agreement-rollup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "rollup node id")
	promoterRollupPath := fs.String("promoter-rollup", "", "source Promoter no-promotion rollup path")
	commandReadbackPath := fs.String("command-readback", "", "source Command readback path")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "Command/Promoter agreement rollup output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--promoter-rollup":  *promoterRollupPath,
		"--command-readback": *commandReadbackPath,
		"--source-readback":  *sourceReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*promoterRollupPath, *commandReadbackPath, *sourceReadbackPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	rollup, err := BuildAtlasCommandPromoterAgreementRollup(*nodeID, *promoterRollupPath, *commandReadbackPath, *sourceReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasCommandPromoterAgreementRollup(*outPath, rollup); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, rollup)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncommand_status=%s\npromoter_no_promotion_files=%d\nreadback_completed_nodes=%d\nreadback_ready_nodes=%d\ncommand_agrees_no_promotion=%t\nreadback_agrees_with_command=%t\ncommand_promoter_agreement_rollup=%s\n",
		rollup.Status,
		rollup.NodeID,
		rollup.CommandStatus,
		rollup.PromoterNoPromotionFiles,
		rollup.ReadbackCompletedNodes,
		rollup.ReadbackReadyNodes,
		rollup.CommandAgreesNoPromotion,
		rollup.ReadbackAgreesWithCommand,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsPromoterRollupCountMismatchRegression(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations promoter-rollup-count-mismatch-regression", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "regression node id")
	sourceRollupPath := fs.String("source-rollup", "", "source Promoter no-promotion rollup path")
	outPath := fs.String("out", "", "Promoter rollup count mismatch regression output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":       *nodeID,
		"--source-rollup": *sourceRollupPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceRollupPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	regression, err := BuildAtlasPromoterRollupCountMismatchRegression(*nodeID, *sourceRollupPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasPromoterRollupCountMismatchRegression(*outPath, regression); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, regression)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncase_count=%d\nrejected_cases=%d\npromoter_rollup_count_mismatch_regression=%s\n",
		regression.Status,
		regression.NodeID,
		regression.CaseCount,
		regression.RejectedCases,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsCommandPromoterDisagreementDenial(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations command-promoter-disagreement-denial", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "denial evidence node id")
	sourceAgreementPath := fs.String("source-agreement", "", "source Command/Promoter agreement rollup path")
	outPath := fs.String("out", "", "Command/Promoter disagreement denial output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-agreement": *sourceAgreementPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceAgreementPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	evidence, err := BuildAtlasCommandPromoterDisagreementDenial(*nodeID, *sourceAgreementPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasCommandPromoterDisagreementDenial(*outPath, evidence); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, evidence)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ncase_count=%d\ndenied_cases=%d\nfinal_response_allowed=%t\ncommand_promoter_disagreement_denial=%s\n",
		evidence.Status,
		evidence.NodeID,
		evidence.CaseCount,
		evidence.DeniedCases,
		evidence.FinalResponseAllowed,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFoundryImportReadinessBinding(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-import-readiness-binding", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "readiness binding node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	sourceWorkgraphPath := fs.String("source-workgraph", "", "source workgraph path")
	foundryImportPath := fs.String("foundry-import", "", "Foundry import path")
	foundryHandoffPath := fs.String("foundry-handoff", "", "Foundry continuation handoff path")
	outPath := fs.String("out", "", "Foundry import readiness binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--source-workgraph": *sourceWorkgraphPath,
		"--foundry-import":   *foundryImportPath,
		"--foundry-handoff":  *foundryHandoffPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *sourceWorkgraphPath, *foundryImportPath, *foundryHandoffPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasFoundryImportReadinessBinding(*nodeID, *sourceReadbackPath, *sourceWorkgraphPath, *foundryImportPath, *foundryHandoffPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasFoundryImportReadinessBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nactive_node_id=%s\nfoundry_task_count=%d\nmatches_workgraph=%t\nmatches_readback_next_node=%t\nhandoff_matches_import=%t\nfoundry_import_readiness_binding=%s\n",
		binding.Status,
		binding.NodeID,
		binding.ActiveNodeID,
		binding.FoundryTaskCount,
		binding.MatchesWorkgraph,
		binding.MatchesReadbackNextNode,
		binding.HandoffMatchesImport,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsRunLinkDigestCheck(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations run-link-digest-check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "run-link digest check node id")
	runLinkPath := fs.String("run-link", "", "source run-link path")
	evidenceRoot := fs.String("evidence-root", "", "evidence root used to resolve run-link evidence paths")
	outPath := fs.String("out", "", "run-link digest check output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":       *nodeID,
		"--run-link":      *runLinkPath,
		"--evidence-root": *evidenceRoot,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*runLinkPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	check, err := BuildAtlasRunLinkDigestCheck(*nodeID, *runLinkPath, *evidenceRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasRunLinkDigestCheck(*outPath, check); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, check)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\ntask_id=%s\nevidence_count=%d\ndigest_matches=%t\nrun_link_digest_check=%s\n",
		check.Status,
		check.NodeID,
		check.TaskID,
		check.EvidenceCount,
		check.DigestMatches,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFoundryHandoffReplayFixture(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-handoff-replay-fixture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "Foundry handoff replay fixture node id")
	sourceReadbackPath := fs.String("source-readback", "", "resumed recommendation readback path")
	sourceWorkgraphPath := fs.String("source-workgraph", "", "source workgraph path")
	foundryImportPath := fs.String("foundry-import", "", "Foundry import path")
	foundryHandoffPath := fs.String("foundry-handoff", "", "Foundry continuation handoff path")
	outPath := fs.String("out", "", "Foundry handoff replay fixture output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--source-workgraph": *sourceWorkgraphPath,
		"--foundry-import":   *foundryImportPath,
		"--foundry-handoff":  *foundryHandoffPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *sourceWorkgraphPath, *foundryImportPath, *foundryHandoffPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	fixture, err := BuildAtlasFoundryHandoffReplayFixture(AtlasFoundryHandoffReplayFixtureOptions{
		NodeID:              *nodeID,
		SourceReadbackPath:  *sourceReadbackPath,
		SourceWorkgraphPath: *sourceWorkgraphPath,
		FoundryImportPath:   *foundryImportPath,
		FoundryHandoffPath:  *foundryHandoffPath,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasFoundryHandoffReplayFixture(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nactive_node_id=%s\nsingle_active_import_task=%t\nhandoff_matches_resumed_readback=%t\nfoundry_handoff_replay_fixture=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.ActiveNodeID,
		fixture.SingleActiveImportTask,
		fixture.HandoffMatchesResumedReadback,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsFoundryTerminalStatusExamples(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations foundry-terminal-status-examples", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "terminal status examples validation node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	outPath := fs.String("out", "", "terminal status examples validation output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":         *nodeID,
		"--source-readback": *sourceReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*sourceReadbackPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	fixture, err := BuildAtlasFoundryTerminalStatusExamplesValidation(*nodeID, *sourceReadbackPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasFoundryTerminalStatusExamplesValidation(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nterminal_example_count=%d\ndenied_example_count=%d\nexamples_match_readback_enums=%t\nfoundry_terminal_status_examples=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.TerminalExampleCount,
		fixture.DeniedExampleCount,
		fixture.ExamplesMatchReadbackEnums,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsMissionDashboardClosureBinding(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations mission-dashboard-closure-binding", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "Mission dashboard closure binding node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	sourceNodeDir := fs.String("source-node-dir", "", "source node evidence directory")
	outPath := fs.String("out", "", "Mission dashboard closure binding output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":         *nodeID,
		"--source-readback": *sourceReadbackPath,
		"--source-node-dir": *sourceNodeDir,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *sourceNodeDir} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	binding, err := BuildAtlasMissionDashboardClosureBinding(*nodeID, *sourceReadbackPath, *sourceNodeDir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMissionDashboardClosureBinding(*outPath, binding); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, binding)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nrow_count=%d\nfinal_response_allowed=%t\natlas_closure_evidence_bound=%t\nmission_dashboard_closure_binding=%s\n",
		binding.Status,
		binding.NodeID,
		binding.RowCount,
		binding.FinalResponseAllowed,
		binding.AtlasClosureEvidenceBound,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsMissionDashboardProvenanceLinks(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations mission-dashboard-provenance-links", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "Mission dashboard provenance links node id")
	dashboardBindingPath := fs.String("dashboard-binding", "", "Mission dashboard closure binding path")
	outPath := fs.String("out", "", "Mission dashboard provenance links output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":           *nodeID,
		"--dashboard-binding": *dashboardBindingPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*dashboardBindingPath, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	links, err := BuildAtlasMissionDashboardProvenanceLinks(*nodeID, *dashboardBindingPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMissionDashboardProvenanceLinks(*outPath, links); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, links)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nprovenance_link_count=%d\nall_required_provenance_linked=%t\nevery_link_matches_dashboard=%t\nmission_dashboard_provenance_links=%s\n",
		links.Status,
		links.NodeID,
		links.ProvenanceLinkCount,
		links.AllRequiredProvenanceLinked,
		links.EveryLinkMatchesDashboard,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsMissionDashboardFreshnessChecks(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations mission-dashboard-freshness-checks", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "Mission dashboard freshness checks node id")
	provenanceLinksPath := fs.String("provenance-links", "", "Mission dashboard provenance links path")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	postMergeLifecyclePath := fs.String("post-merge-lifecycle", "", "post-merge lifecycle evidence path")
	outPath := fs.String("out", "", "Mission dashboard freshness checks output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":              *nodeID,
		"--provenance-links":     *provenanceLinksPath,
		"--source-readback":      *sourceReadbackPath,
		"--post-merge-lifecycle": *postMergeLifecyclePath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*provenanceLinksPath, *sourceReadbackPath, *postMergeLifecyclePath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	fixture, err := BuildAtlasMissionDashboardFreshnessChecks(*nodeID, *provenanceLinksPath, *sourceReadbackPath, *postMergeLifecyclePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMissionDashboardFreshnessChecks(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nfreshness_check_count=%d\nall_freshness_checks_passed=%t\nmain_synced_to_merge_commit=%t\nmission_dashboard_freshness_checks=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.FreshnessCheckCount,
		fixture.AllFreshnessChecksPassed,
		fixture.MainSyncedToMergeCommit,
		filepath.ToSlash(*outPath),
	)
	return nil
}

func runMissionRecommendationsMissionDashboardCompactFilters(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations mission-dashboard-compact-filters", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	nodeID := fs.String("node-id", "", "Mission dashboard compact filters node id")
	sourceReadbackPath := fs.String("source-readback", "", "source recommendation readback path")
	sourceWorkgraphPath := fs.String("source-workgraph", "", "source workgraph path")
	outPath := fs.String("out", "", "Mission dashboard compact filters output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--node-id":          *nodeID,
		"--source-readback":  *sourceReadbackPath,
		"--source-workgraph": *sourceWorkgraphPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, input := range []string{*sourceReadbackPath, *sourceWorkgraphPath} {
		if strings.TrimSpace(*outPath) != "" && samePath(input, *outPath) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	fixture, err := BuildAtlasMissionDashboardCompactFilters(*nodeID, *sourceReadbackPath, *sourceWorkgraphPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteAtlasMissionDashboardCompactFilters(*outPath, fixture); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, fixture)
	}
	fmt.Fprintf(stdout, "status=%s\nnode_id=%s\nready_nodes=%d\nblocked_nodes=%d\nactive_filter=%s\nmission_dashboard_compact_filters=%s\n",
		fixture.Status,
		fixture.NodeID,
		fixture.ReadyNodes,
		fixture.BlockedNodes,
		fixture.ActiveFilterKey,
		filepath.ToSlash(*outPath),
	)
	if fixture.SchemaHealthStatus != "" {
		fmt.Fprintf(stdout, "schema_health_status=%s\nschema_health_filter=%s\nschema_health_filter_state=%s\n",
			fixture.SchemaHealthStatus,
			fixture.SchemaHealthFilterKey,
			fixture.SchemaHealthFilterState,
		)
	}
	return nil
}

func runMissionRecommendationsValidateEvidence(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations validate-evidence", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	evidenceRoot := fs.String("evidence-root", "", "Atlas recommendation evidence root")
	outPath := fs.String("out", "", "validation report output path")
	jsonOut := fs.Bool("json", false, "json output")
	strictSchemaRegistry := fs.Bool("strict-schema-registry", false, "reject evidence schemas absent from the consolidation allowlist")
	requireProvenanceFields := fs.Bool("require-provenance-fields", false, "require source_digest and evidence_class on evidence files")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*evidenceRoot) == "" {
		return fmt.Errorf("--evidence-root is required")
	}
	if strings.TrimSpace(*outPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	if strings.TrimSpace(*outPath) != "" && samePath(*evidenceRoot, *outPath) {
		return fmt.Errorf("refusing to overwrite input artifact")
	}
	report, err := BuildAtlasRecommendationEvidenceValidationReportWithValidationOptions(*evidenceRoot, AtlasRecommendationEvidenceValidationOptions{
		StrictSchemaRegistry:    *strictSchemaRegistry,
		RequireProvenanceFields: *requireProvenanceFields,
	})
	if strings.TrimSpace(*outPath) != "" {
		if writeErr := WriteJSON(*outPath, report); writeErr != nil {
			return writeErr
		}
	}
	if *jsonOut {
		if printErr := printJSON(stdout, report); printErr != nil {
			return printErr
		}
	} else {
		fmt.Fprintf(stdout, "status=%s\nnode_count=%d\njson_files=%d\nschema_bound_files=%d\ntyped_validator_files=%d\nstrict_schema_registry=%t\nrequire_provenance_fields=%t\nunknown_schema_files=%d\nmissing_schema_files=%d\nmissing_source_digest_files=%d\nmissing_evidence_class_files=%d\nfailed_files=%d\nrecommendation_evidence_validation_report=%s\n",
			report.Status,
			report.NodeCount,
			report.JSONFileCount,
			report.SchemaBoundFiles,
			report.TypedValidatorFiles,
			report.StrictSchemaRegistry,
			report.RequireProvenanceFields,
			len(report.UnknownSchemaFiles),
			len(report.MissingSchemaFiles),
			len(report.MissingSourceDigestFiles),
			len(report.MissingEvidenceClassFiles),
			len(report.FailedFiles),
			filepath.ToSlash(*outPath),
		)
	}
	return err
}

func runMissionRecommendationsCompleteNode(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations complete-node", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	wavePath := fs.String("wave", "", "Atlas recommendation wave path")
	workgraphPath := fs.String("workgraph", "", "Atlas recommendation workgraph path")
	runLinkPath := fs.String("run-link", "", "completed run-link path")
	expectedNodeID := fs.String("expected-node", "", "expected executable recommendation node id")
	evidenceRoot := fs.String("evidence-root", "", "filesystem root used to verify run-link evidence files")
	readbackEvidenceRoot := fs.String("readback-evidence-root", "", "portable evidence root written into readback")
	leaseStartPath := fs.String("lease-start", "", "lease start marker path")
	startedAt := fs.String("started-at", "", "long-run lease start time, RFC3339")
	completedAt := fs.String("completed-at", "", "long-run lease completion time, RFC3339")
	elapsedMinutes := fs.Int("elapsed-minutes", 0, "long-run lease elapsed minutes")
	leaseTimingMode := fs.String("lease-timing-mode", "", "lease timing evidence mode")
	outWorkgraphPath := fs.String("out-workgraph", "", "updated workgraph output path")
	outReadbackPath := fs.String("out-readback", "", "updated recommendation readback output path")
	outExecutionReadbackPath := fs.String("out-execution-readback", "", "updated execution readback output path")
	outCheckpointReadbackPath := fs.String("out-checkpoint-readback", "", "updated checkpoint readback output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--wave":          *wavePath,
		"--workgraph":     *workgraphPath,
		"--run-link":      *runLinkPath,
		"--out-workgraph": *outWorkgraphPath,
		"--out-readback":  *outReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	for _, out := range []string{*outWorkgraphPath, *outReadbackPath, *outExecutionReadbackPath, *outCheckpointReadbackPath} {
		if strings.TrimSpace(out) == "" {
			continue
		}
		if samePath(*wavePath, out) || samePath(*workgraphPath, out) || samePath(*runLinkPath, out) || samePath(*leaseStartPath, out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	wave, err := LoadJSON[AtlasRecommendationWave](*wavePath)
	if err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*workgraphPath)
	if err != nil {
		return err
	}
	link, err := LoadJSON[RunLink](*runLinkPath)
	if err != nil {
		return err
	}
	updated, completedNodeID, err := CompleteAtlasRecommendationNodeWithRunLink(wave, workgraph, link, AtlasRecommendationCompleteNodeOptions{
		ExpectedNodeID: *expectedNodeID,
		EvidenceRoot:   *evidenceRoot,
	})
	if err != nil {
		return err
	}
	if err := WriteJSON(*outWorkgraphPath, updated); err != nil {
		return err
	}
	readbackOptions, err := recommendationReadbackOptionsFromLeaseStart(*leaseStartPath, AtlasRecommendationReadbackOptions{
		WavePath:               *wavePath,
		WorkgraphPath:          *outWorkgraphPath,
		EvidenceRoot:           *readbackEvidenceRoot,
		StartedAt:              *startedAt,
		CompletedAt:            *completedAt,
		ElapsedMinutes:         *elapsedMinutes,
		LeaseTimingMode:        *leaseTimingMode,
		PublicSafetyScanStatus: publicSafetyStatusFromRunLink(link),
	})
	if err != nil {
		return err
	}
	readback, err := BuildAtlasRecommendationReadback(wave, updated, readbackOptions)
	if err != nil {
		return err
	}
	if err := WriteJSON(*outReadbackPath, readback); err != nil {
		return err
	}
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		return err
	}
	if strings.TrimSpace(*outExecutionReadbackPath) != "" {
		if err := WriteJSON(*outExecutionReadbackPath, execution); err != nil {
			return err
		}
	}
	checkpoint := BuildAtlasRecommendationCheckpointReadback(readback)
	if err := ValidateAtlasRecommendationCheckpointReadback(checkpoint); err != nil {
		return err
	}
	if strings.TrimSpace(*outCheckpointReadbackPath) != "" {
		if err := WriteJSON(*outCheckpointReadbackPath, checkpoint); err != nil {
			return err
		}
	}
	nextExecutable := readback.FirstExecutableNode
	if nextExecutable == "" {
		nextExecutable = "none"
	}
	fmt.Fprintf(stdout, "status=written\ncompleted_node=%s\ncompleted_nodes=%d\nready_nodes=%d\nnext_executable_node=%s\ncheckpoint_count=%d\nreturn_gate_status=%s\nelapsed_minutes=%d\nmin_minutes_met=%t\nlease_time_status=%s\nfinal_response_allowed=%t\nupdated_workgraph=%s\nrecommendation_readback=%s\nexecution_readback=%s\ncheckpoint_readback=%s\n",
		completedNodeID,
		readback.CompletedNodes,
		readback.ReadyNodes,
		nextExecutable,
		readback.CheckpointCount,
		readback.ReturnGateStatus,
		readback.ElapsedMinutes,
		readback.MinMinutesMet,
		readback.LeaseTimeStatus,
		readback.FinalResponseAllowed,
		filepath.ToSlash(*outWorkgraphPath),
		filepath.ToSlash(*outReadbackPath),
		filepath.ToSlash(*outExecutionReadbackPath),
		filepath.ToSlash(*outCheckpointReadbackPath),
	)
	return nil
}

func publicSafetyStatusFromRunLink(link RunLink) string {
	if strings.TrimSpace(link.Evidence["public_safety_readback_binding"]) == "" {
		return ""
	}
	return "passed"
}

func recommendationReadbackOptionsFromLeaseStart(leaseStartPath string, options AtlasRecommendationReadbackOptions) (AtlasRecommendationReadbackOptions, error) {
	if strings.TrimSpace(leaseStartPath) == "" {
		return options, nil
	}
	leaseStart, err := LoadJSON[AtlasRecommendationLeaseStart](leaseStartPath)
	if err != nil {
		return AtlasRecommendationReadbackOptions{}, err
	}
	if err := ValidateAtlasRecommendationLeaseStart(leaseStart); err != nil {
		return AtlasRecommendationReadbackOptions{}, err
	}
	if strings.TrimSpace(options.StartedAt) == "" {
		options.StartedAt = leaseStart.StartedAt
	}
	if strings.TrimSpace(options.EvidenceRoot) == "" {
		options.EvidenceRoot = leaseStart.EvidenceRoot
	}
	if strings.TrimSpace(options.LeaseTimingMode) == "" {
		options.LeaseTimingMode = "actual"
	}
	return options, nil
}

func runMissionRecommendationsResume(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations resume", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	wavePath := fs.String("wave", "", "Atlas recommendation wave path")
	workgraphPath := fs.String("workgraph", "", "Atlas recommendation workgraph path")
	leaseStartPath := fs.String("lease-start", "", "lease start marker path")
	evidenceRoot := fs.String("evidence-root", "", "relative evidence root")
	completedAt := fs.String("completed-at", "", "long-run lease completion time, RFC3339")
	elapsedMinutes := fs.Int("elapsed-minutes", 0, "long-run lease elapsed minutes")
	leaseTimingMode := fs.String("lease-timing-mode", "actual", "lease timing evidence mode")
	outReadbackPath := fs.String("out-readback", "", "resumed recommendation readback output path")
	outExecutionReadbackPath := fs.String("out-execution-readback", "", "resumed execution readback output path")
	outCommandReadbackPath := fs.String("out-command-readback", "", "compact Command readback output path")
	outPromoterReadbackPath := fs.String("out-promoter-readback", "", "Promoter no-promotion readback output path")
	outFoundryRollupPath := fs.String("out-foundry-rollup", "", "Foundry run-link rollup output path")
	outReconciliationPacketPath := fs.String("out-reconciliation-packet", "", "Atlas recommendation reconciliation packet output path")
	outNextPromptPath := fs.String("out-next-prompt", "", "updated Atlas continuation prompt output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--wave":         *wavePath,
		"--workgraph":    *workgraphPath,
		"--lease-start":  *leaseStartPath,
		"--out-readback": *outReadbackPath,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	for _, out := range []string{*outReadbackPath, *outExecutionReadbackPath, *outCommandReadbackPath, *outPromoterReadbackPath, *outFoundryRollupPath, *outReconciliationPacketPath, *outNextPromptPath} {
		if strings.TrimSpace(out) == "" {
			continue
		}
		if samePath(*wavePath, out) || samePath(*workgraphPath, out) || samePath(*leaseStartPath, out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	wave, err := LoadJSON[AtlasRecommendationWave](*wavePath)
	if err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*workgraphPath)
	if err != nil {
		return err
	}
	options, err := recommendationReadbackOptionsFromLeaseStart(*leaseStartPath, AtlasRecommendationReadbackOptions{
		WavePath:        *wavePath,
		WorkgraphPath:   *workgraphPath,
		EvidenceRoot:    *evidenceRoot,
		CompletedAt:     *completedAt,
		ElapsedMinutes:  *elapsedMinutes,
		LeaseTimingMode: *leaseTimingMode,
	})
	if err != nil {
		return err
	}
	readback, err := BuildAtlasRecommendationReadback(wave, workgraph, options)
	if err != nil {
		return err
	}
	if err := WriteJSON(*outReadbackPath, readback); err != nil {
		return err
	}
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		return err
	}
	if strings.TrimSpace(*outExecutionReadbackPath) != "" {
		if err := WriteJSON(*outExecutionReadbackPath, execution); err != nil {
			return err
		}
	}
	command := BuildAtlasRecommendationCommandReadback(readback)
	promoter := BuildAtlasRecommendationPromoterReadback(readback)
	foundry := BuildAtlasRecommendationFoundryRollup(readback)
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err != nil {
		return err
	}
	if strings.TrimSpace(*outCommandReadbackPath) != "" {
		if err := WriteJSON(*outCommandReadbackPath, command); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*outPromoterReadbackPath) != "" {
		if err := WriteJSON(*outPromoterReadbackPath, promoter); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*outFoundryRollupPath) != "" {
		if err := WriteJSON(*outFoundryRollupPath, foundry); err != nil {
			return err
		}
	}
	reconciliation := BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, reconciliation); err != nil {
		return err
	}
	if strings.TrimSpace(*outReconciliationPacketPath) != "" {
		if err := WriteJSON(*outReconciliationPacketPath, reconciliation); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*outNextPromptPath) != "" {
		prompt := BuildAtlasRecommendationResumePrompt(readback, AtlasRecommendationResumePromptOptions{
			EvidenceRoot:   *evidenceRoot,
			LeaseStartPath: *leaseStartPath,
			WorkgraphPath:  *workgraphPath,
			ReadbackPath:   *outReadbackPath,
		})
		if err := os.WriteFile(*outNextPromptPath, []byte(prompt), 0o644); err != nil {
			return err
		}
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\nstarted_at=%s\ncompleted_at=%s\nelapsed_minutes=%d\nmin_minutes_met=%t\nlease_time_status=%s\ncheckpoint_count=%d\nreturn_gate_status=%s\nfinal_response_allowed=%t\nexact_next_action=%s\nrecommendation_readback=%s\nexecution_readback=%s\ncommand_readback=%s\npromoter_readback=%s\nfoundry_rollup=%s\nreconciliation_packet=%s\nnext_recommended_prompt=%s\n",
		readback.Status,
		readback.MissionID,
		readback.StartedAt,
		readback.CompletedAt,
		readback.ElapsedMinutes,
		readback.MinMinutesMet,
		readback.LeaseTimeStatus,
		readback.CheckpointCount,
		readback.ReturnGateStatus,
		readback.FinalResponseAllowed,
		readback.ExactNextAction,
		filepath.ToSlash(*outReadbackPath),
		filepath.ToSlash(*outExecutionReadbackPath),
		filepath.ToSlash(*outCommandReadbackPath),
		filepath.ToSlash(*outPromoterReadbackPath),
		filepath.ToSlash(*outFoundryRollupPath),
		filepath.ToSlash(*outReconciliationPacketPath),
		filepath.ToSlash(*outNextPromptPath),
	)
	return nil
}

func runMissionRecommendationsReadback(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mission recommendations readback", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	wavePath := fs.String("wave", "", "Atlas recommendation wave path")
	workgraphPath := fs.String("workgraph", "", "Atlas recommendation workgraph path")
	evidenceRoot := fs.String("evidence-root", "", "relative evidence root")
	startedAt := fs.String("started-at", "", "long-run lease start time, RFC3339")
	completedAt := fs.String("completed-at", "", "long-run lease completion time, RFC3339")
	elapsedMinutes := fs.Int("elapsed-minutes", 0, "long-run lease elapsed minutes")
	leaseTimingMode := fs.String("lease-timing-mode", "", "lease timing evidence mode")
	schemaHealthStatus := fs.String("schema-health-status", "", "schema registry health readback status")
	outPath := fs.String("out", "", "output path")
	outExecutionReadbackPath := fs.String("out-execution-readback", "", "execution readback output path")
	outWorkgraphReadinessPacketPath := fs.String("out-workgraph-readiness-packet", "", "generated workgraph readiness packet output path")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*wavePath) == "" {
		return fmt.Errorf("--wave is required")
	}
	if strings.TrimSpace(*workgraphPath) == "" {
		return fmt.Errorf("--workgraph is required")
	}
	if strings.TrimSpace(*outPath) == "" && strings.TrimSpace(*outExecutionReadbackPath) == "" && strings.TrimSpace(*outWorkgraphReadinessPacketPath) == "" && !*jsonOut {
		return fmt.Errorf("--out or --json is required")
	}
	for _, out := range []string{*outPath, *outExecutionReadbackPath, *outWorkgraphReadinessPacketPath} {
		if strings.TrimSpace(out) == "" {
			continue
		}
		if samePath(*wavePath, out) || samePath(*workgraphPath, out) {
			return fmt.Errorf("refusing to overwrite input artifact")
		}
	}
	wave, err := LoadJSON[AtlasRecommendationWave](*wavePath)
	if err != nil {
		return err
	}
	workgraph, err := LoadJSON[Workgraph](*workgraphPath)
	if err != nil {
		return err
	}
	readback, err := BuildAtlasRecommendationReadback(wave, workgraph, AtlasRecommendationReadbackOptions{
		WavePath:           *wavePath,
		WorkgraphPath:      *workgraphPath,
		EvidenceRoot:       *evidenceRoot,
		StartedAt:          *startedAt,
		CompletedAt:        *completedAt,
		ElapsedMinutes:     *elapsedMinutes,
		LeaseTimingMode:    *leaseTimingMode,
		SchemaHealthStatus: *schemaHealthStatus,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) != "" {
		if err := WriteJSON(*outPath, readback); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*outExecutionReadbackPath) != "" {
		execution := BuildAtlasRecommendationExecutionReadback(readback)
		if err := WriteJSON(*outExecutionReadbackPath, execution); err != nil {
			return err
		}
	}
	if strings.TrimSpace(*outWorkgraphReadinessPacketPath) != "" {
		packet, err := BuildAtlasRecommendationWorkgraphReadinessPacket(readback, AtlasRecommendationWorkgraphReadinessPacketOptions{
			WavePath:      *wavePath,
			WorkgraphPath: *workgraphPath,
			ReadbackPath:  *outPath,
		})
		if err != nil {
			return err
		}
		if err := WriteJSON(*outWorkgraphReadinessPacketPath, packet); err != nil {
			return err
		}
	}
	if *jsonOut {
		return printJSON(stdout, readback)
	}
	fmt.Fprintf(stdout, "status=%s\nmission_id=%s\ntotal_nodes=%d\ncompleted_nodes=%d\nready_nodes=%d\nexecutable_ready_nodes=%d\nlease_health=%s\ncheckpoint_count=%d\nreturn_gate_status=%s\nschema_health_status=%s\nelapsed_minutes=%d\nmin_minutes_met=%t\nlease_time_status=%s\nfinal_response_allowed=%t\nexact_next_action=%s\nrecommendation_readback=%s\nexecution_readback=%s\nworkgraph_readiness_packet=%s\n",
		readback.Status,
		readback.MissionID,
		readback.TotalNodes,
		readback.CompletedNodes,
		readback.ReadyNodes,
		readback.ExecutableReadyNodes,
		readback.LeaseHealthStatus,
		readback.CheckpointCount,
		readback.ReturnGateStatus,
		readback.SchemaHealthStatus,
		readback.ElapsedMinutes,
		readback.MinMinutesMet,
		readback.LeaseTimeStatus,
		readback.FinalResponseAllowed,
		readback.ExactNextAction,
		filepath.ToSlash(*outPath),
		filepath.ToSlash(*outExecutionReadbackPath),
		filepath.ToSlash(*outWorkgraphReadinessPacketPath),
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
