package atlas

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateBlueprintImport(record BlueprintImport) error {
	var errs []string
	requireContract(&errs, "blueprint_import", record.ContractVersion, BlueprintImportContract)
	requireField(&errs, "id", record.ID)
	requireField(&errs, "project_id", record.ProjectID)
	if !oneOf(record.Status, "ready", "blocked") {
		errs = append(errs, "status must be ready or blocked")
	}
	requireField(&errs, "reason", record.Reason)
	requireField(&errs, "blueprint_pack.ref", record.BlueprintPack.Ref)
	if !digestPattern.MatchString(record.BlueprintPack.Digest) {
		errs = append(errs, "blueprint_pack.digest must be sha256:<64 hex>")
	}
	if record.BuildAuthorization.Ref != "" && !digestPattern.MatchString(record.BuildAuthorization.Digest) {
		errs = append(errs, "build_authorization.digest must be sha256:<64 hex>")
	}
	if len(record.Digests) == 0 {
		errs = append(errs, "digests must not be empty")
	}
	for key, digest := range record.Digests {
		requireField(&errs, "digests."+key, digest)
		if !digestPattern.MatchString(digest) {
			errs = append(errs, "digests."+key+" must be sha256:<64 hex>")
		}
	}
	if record.Status == "ready" {
		if !record.ReadyForFoundry {
			errs = append(errs, "ready_for_foundry must be true when status is ready")
		}
		requireField(&errs, "target_instance", record.TargetInstance)
		requireField(&errs, "workgraph_id", record.WorkgraphID)
		requireField(&errs, "mutation_class", record.MutationClass)
		requireField(&errs, "downstream_foundry_import.ref", record.DownstreamFoundryImport.Ref)
		if !digestPattern.MatchString(record.DownstreamFoundryImport.Digest) {
			errs = append(errs, "downstream_foundry_import.digest must be sha256:<64 hex>")
		}
		requireField(&errs, "downstream_foundry_continuation_handoff.ref", record.DownstreamFoundryContinuationHandoff.Ref)
		if !digestPattern.MatchString(record.DownstreamFoundryContinuationHandoff.Digest) {
			errs = append(errs, "downstream_foundry_continuation_handoff.digest must be sha256:<64 hex>")
		}
		if record.CandidateSelection.ContractVersion != BlueprintCandidateSelectionContract {
			errs = append(errs, "candidate_selection contract_version must be "+BlueprintCandidateSelectionContract)
		}
		if record.Digests["downstream_foundry_import"] == "" {
			errs = append(errs, "digests.downstream_foundry_import must not be empty when ready")
		}
		if record.Digests["downstream_foundry_continuation_handoff"] == "" {
			errs = append(errs, "digests.downstream_foundry_continuation_handoff must not be empty when ready")
		}
	} else {
		if record.ReadyForFoundry {
			errs = append(errs, "ready_for_foundry must be false when status is blocked")
		}
		requireList(&errs, "blocking_next_actions", record.BlockingNextActions)
	}
	if record.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if record.LiveExecutionProven {
		errs = append(errs, "live_execution_proven must be false")
	}
	if record.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if record.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if record.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if record.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if record.CallsProviders {
		errs = append(errs, "calls_providers must be false")
	}
	if record.ReleaseOrPublishAllowed {
		errs = append(errs, "release_or_publish_allowed must be false")
	}
	checkPublicPath(&errs, "blueprint_pack.ref", record.BlueprintPack.Ref, true)
	checkPublicPath(&errs, "build_authorization.ref", record.BuildAuthorization.Ref, true)
	checkPublicPath(&errs, "downstream_foundry_import.ref", record.DownstreamFoundryImport.Ref, true)
	checkPublicPath(&errs, "downstream_foundry_continuation_handoff.ref", record.DownstreamFoundryContinuationHandoff.Ref, true)
	checkPublicStrings(&errs, "safety_limits", record.SafetyLimits, true)
	checkPublicStrings(&errs, "blocking_next_actions", record.BlockingNextActions, true)
	return joinErrors(errs)
}

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
