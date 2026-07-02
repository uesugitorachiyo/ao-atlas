package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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

func buildBlueprintIntake(rules BlueprintCandidateRules) Intake {
	return Intake{
		ContractVersion: IntakeContract,
		ID:              rules.ProjectID + "-intake",
		TargetInstance:  rules.TargetInstance,
		BroadPrompt:     rules.Objective + " The work must pass through AO Atlas before AO Foundry gates and must remain dry-run readback only.",
		InstructionRefs: []string{"implementation-spec.md", "quality-profile.md", "candidate-rules.json"},
		FolderRefs:      []string{"excluded/"},
		Constraints:     append([]string{"Blueprint -> Atlas -> Foundry is mandatory"}, rules.SafetyLimits...),
	}
}

func buildBlueprintContextPack(packPath, candidateRulesPath string, rules BlueprintCandidateRules, digests map[string]string) (ContextPack, error) {
	sourceRefs := []SourceRef{}
	for _, ref := range rules.ContextRefs {
		key := contextDigestKey(ref)
		digest := digests[key]
		sourceRef := filepath.ToSlash(ref)
		if digest == "" {
			path := filepath.Join(packPath, ref)
			var err error
			digest, err = digestFile(path)
			if err != nil {
				return ContextPack{}, fmt.Errorf("context ref %s: %w", ref, err)
			}
		}
		if key == "candidate_rules" && strings.TrimSpace(candidateRulesPath) != "" {
			sourceRef = publicArtifactRef(candidateRulesPath)
		}
		sourceRefs = append(sourceRefs, SourceRef{Ref: sourceRef, Digest: digest})
	}
	pack := ContextPack{
		ContractVersion: ContextPackContract,
		ID:              rules.CandidateID + "-context-pack",
		TaskID:          rules.CandidateID + "-task",
		BudgetBytes:     8192,
		SourceRefs:      sourceRefs,
		Summaries: []string{
			"Blueprint pack is authorized for Atlas compile into Foundry import material.",
			"Mutation class " + rules.MutationClass + " remains dry-run/readback only.",
		},
		Assumptions:     []string{"AO Blueprint owns requirements authorization.", "AO Atlas does not schedule or execute work."},
		Exclusions:      []string{"live code mutation", "provider calls", "private local paths outside excluded placeholders"},
		MissingProtocol: "Return to AO Blueprint for missing requirements or authorization before widening scope.",
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		return ContextPack{}, err
	}
	return pack, nil
}

func buildBlueprintFactoryTask(rules BlueprintCandidateRules, pack ContextPack) FactoryTask {
	return FactoryTask{
		ContractVersion:   FactoryTaskContract,
		ID:                rules.CandidateID + "-task",
		Objective:         rules.Objective,
		TargetFactoryRepo: rules.TargetFactoryRepo,
		FactoryFolder:     rules.FactoryFolder,
		MutationClass:     rules.MutationClass,
		Acceptance:        append([]string(nil), rules.Acceptance...),
		NonGoals:          append([]string(nil), rules.NonGoals...),
		WriteScope:        append([]string(nil), rules.WriteScope...),
		RequiredGates:     append([]string(nil), rules.RequiredGates...),
		RollbackScope:     append([]string(nil), rules.RollbackScope...),
		Verification:      append([]string(nil), rules.Verification...),
		RequiredEvidence:  append([]string(nil), rules.RequiredEvidence...),
		SafetyLimits:      append([]string(nil), rules.SafetyLimits...),
		AuthorityBoundary: rules.AuthorityBoundary,
		DependencyRefs:    append([]string(nil), rules.DependencyRefs...),
		ContextPackRefs:   []string{filepath.ToSlash(filepath.Join("context-packs", pack.ID+".json"))},
	}
}

func buildBlueprintCandidateSelection(rules BlueprintCandidateRules, node WorkgraphNode, digests map[string]string) BlueprintCandidateSelection {
	return BlueprintCandidateSelection{
		ContractVersion:     BlueprintCandidateSelectionContract,
		ID:                  rules.CandidateID + "-selection",
		ProjectID:           rules.ProjectID,
		Status:              "ready",
		SelectedCandidateID: rules.CandidateID,
		MutationClass:       rules.MutationClass,
		TargetFactoryRepo:   rules.TargetFactoryRepo,
		WorkgraphID:         rules.WorkgraphID,
		NodeID:              node.ID,
		TaskID:              node.FactoryTask.ID,
		RequiredGates:       append([]string(nil), rules.RequiredGates...),
		RequiredEvidence:    append([]string(nil), rules.RequiredEvidence...),
		SafetyLimits:        append([]string(nil), rules.SafetyLimits...),
		Digests:             copyStringMap(digests),
		SchedulesWork:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		SafeToExecute:       false,
		LiveExecutionProven: false,
	}
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

func readJSONIfPossible(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func digestFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func digestDirectory(root string) (string, error) {
	hash := sha256.New()
	cleanRoot := filepath.Clean(root)
	err := filepath.WalkDir(cleanRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipBlueprintDigestDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(cleanRoot, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash.Write([]byte(filepath.ToSlash(rel)))
		hash.Write([]byte{0})
		hash.Write(body)
		hash.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func digestValue(value any) string {
	data, _ := json.Marshal(value)
	return DigestBytes(data)
}

func shouldSkipBlueprintDigestDir(name string) bool {
	switch name {
	case ".git", "tmp", "target", ".idea", ".vscode", "__pycache__":
		return true
	default:
		return false
	}
}

func contextDigestKey(ref string) string {
	switch filepath.ToSlash(ref) {
	case "implementation-spec.md":
		return "implementation_spec"
	case "quality-profile.md":
		return "quality_profile"
	case "candidate-rules.json":
		return "candidate_rules"
	default:
		return ""
	}
}

func publicArtifactRef(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	clean := filepath.Clean(path)
	abs, err := filepath.Abs(clean)
	if err == nil {
		if root, rootErr := findRepoRoot(); rootErr == nil {
			if rel, relErr := filepath.Rel(root, abs); relErr == nil && !strings.HasPrefix(rel, "..") {
				return filepath.ToSlash(rel)
			}
		}
	}
	return filepath.ToSlash(filepath.Join("excluded", "local-artifacts", filepath.Base(clean)))
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", errors.New("repo root not found")
		}
		cwd = parent
	}
}

func copyStringMap(values map[string]string) map[string]string {
	result := map[string]string{}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result[key] = values[key]
	}
	return result
}
