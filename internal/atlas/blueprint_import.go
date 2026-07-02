package atlas

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateBlueprintCandidateRules(rules BlueprintCandidateRules) error {
	var errs []string
	if rules.SchemaVersion != BlueprintCandidateRulesContract {
		errs = append(errs, "schema_version must be "+BlueprintCandidateRulesContract)
	}
	requireField(&errs, "project_id", rules.ProjectID)
	requireField(&errs, "target_instance", rules.TargetInstance)
	requireField(&errs, "workgraph_id", rules.WorkgraphID)
	requireField(&errs, "candidate_id", rules.CandidateID)
	requireField(&errs, "mutation_class", rules.MutationClass)
	if !requiredMutationClassNames()[rules.MutationClass] {
		errs = append(errs, "mutation_class must be one of the required mutation classes")
	}
	requireField(&errs, "target_factory_repo", rules.TargetFactoryRepo)
	requireField(&errs, "factory_folder", rules.FactoryFolder)
	requireField(&errs, "objective", rules.Objective)
	requireList(&errs, "acceptance_criteria", rules.Acceptance)
	requireList(&errs, "non_goals", rules.NonGoals)
	requireList(&errs, "write_scope", rules.WriteScope)
	requireList(&errs, "rollback_scope", rules.RollbackScope)
	requireList(&errs, "required_gates", rules.RequiredGates)
	requireList(&errs, "verification_commands", rules.Verification)
	requireList(&errs, "required_evidence", rules.RequiredEvidence)
	requireList(&errs, "safety_limits", rules.SafetyLimits)
	requireField(&errs, "authority_boundary", rules.AuthorityBoundary)
	requireList(&errs, "context_refs", rules.ContextRefs)
	checkPublicPath(&errs, "target_factory_repo", rules.TargetFactoryRepo, false)
	checkPublicPath(&errs, "factory_folder", rules.FactoryFolder, false)
	checkPublicStrings(&errs, "write_scope", rules.WriteScope, true)
	checkPublicStrings(&errs, "rollback_scope", rules.RollbackScope, true)
	checkPublicStrings(&errs, "required_gates", rules.RequiredGates, true)
	checkPublicStrings(&errs, "verification_commands", rules.Verification, true)
	checkPublicStrings(&errs, "required_evidence", rules.RequiredEvidence, true)
	checkPublicStrings(&errs, "safety_limits", rules.SafetyLimits, true)
	checkPublicStrings(&errs, "context_refs", rules.ContextRefs, true)
	return joinErrors(errs)
}

func BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error) {
	result := BlueprintImportResult{}
	if strings.TrimSpace(paths.OutDir) == "" {
		return result, errors.New("--out is required")
	}
	artifacts, compileErr := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	result = blueprintCompileArtifactsToResult(artifacts)
	if artifacts.Record.Status == "blocked" {
		if err := writeBlueprintBlockedArtifacts(paths.OutDir, artifacts.Record, artifacts.Request); err != nil {
			return result, err
		}
		return result, compileErr
	}
	if compileErr != nil {
		return result, compileErr
	}
	if len(artifacts.ContextPacks) != 1 {
		return result, fmt.Errorf("blueprint compiler must emit exactly one context pack")
	}
	if err := writeBlueprintReadyArtifacts(paths.OutDir, artifacts.Record, artifacts.Intake, artifacts.Candidate, artifacts.ContextPacks[0], artifacts.Workgraph, artifacts.FoundryImport, artifacts.Handoff); err != nil {
		return result, err
	}
	return result, nil
}

func writeBlueprintBlockedArtifacts(outDir string, record BlueprintImport, request BlueprintRequest) error {
	if err := ValidateBlueprintRequest(request); err != nil {
		return err
	}
	if err := ValidateBlueprintImport(record); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "blueprint-import.json"), record); err != nil {
		return err
	}
	return WriteJSON(filepath.Join(outDir, "blueprint-request.json"), request)
}
