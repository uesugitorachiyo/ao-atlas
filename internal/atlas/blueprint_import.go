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
	"time"
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
		if record.CandidateSelection.ContractVersion != BlueprintCandidateSelectionContract {
			errs = append(errs, "candidate_selection contract_version must be "+BlueprintCandidateSelectionContract)
		}
		if record.Digests["downstream_foundry_import"] == "" {
			errs = append(errs, "digests.downstream_foundry_import must not be empty when ready")
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
	packDigest, packErr := digestDirectory(paths.PackPath)
	if packErr != nil {
		packDigest = DigestBytes([]byte("missing-blueprint-pack:" + paths.PackPath))
	}
	digests := map[string]string{"blueprint_pack": packDigest}
	record := BlueprintImport{
		ContractVersion:         BlueprintImportContract,
		ID:                      "blueprint-import-blocked",
		ProjectID:               "unknown-project",
		Status:                  "blocked",
		Reason:                  "AO Atlas cannot compile Blueprint material until authorization and candidate rules are ready.",
		BlueprintPack:           SourceRef{Ref: publicArtifactRef(paths.PackPath), Digest: packDigest},
		Digests:                 digests,
		ReadyForFoundry:         false,
		SafeToExecute:           false,
		LiveExecutionProven:     false,
		SchedulesWork:           false,
		ExecutesWork:            false,
		ApprovesWork:            false,
		MutatesRepositories:     false,
		CallsProviders:          false,
		ReleaseOrPublishAllowed: false,
	}
	missing := []string{}
	blockers := []string{}
	if packErr != nil {
		missing = append(missing, "blueprint_pack")
		blockers = append(blockers, "provide a readable AO Blueprint pack")
	}

	var rules BlueprintCandidateRules
	rulesPath := filepath.Join(paths.PackPath, "candidate-rules.json")
	if strings.TrimSpace(paths.CandidateRulesPath) != "" {
		rulesPath = paths.CandidateRulesPath
	}
	if err := readJSONIfPossible(rulesPath, &rules); err != nil {
		missing = append(missing, "candidate_rules")
		blockers = append(blockers, "add candidate-rules.json to the Blueprint pack")
	} else if err := ValidateBlueprintCandidateRules(rules); err != nil {
		missing = append(missing, "candidate_rules")
		blockers = append(blockers, "repair candidate-rules.json: "+err.Error())
	} else {
		record.ID = rules.WorkgraphID + "-blueprint-import"
		record.ProjectID = rules.ProjectID
		record.TargetInstance = rules.TargetInstance
		record.WorkgraphID = rules.WorkgraphID
		record.MutationClass = rules.MutationClass
		record.SafetyLimits = append([]string(nil), rules.SafetyLimits...)
		if digest, err := digestFile(rulesPath); err == nil {
			digests["candidate_rules"] = digest
		}
	}

	for name, path := range map[string]string{
		"implementation_spec": filepath.Join(paths.PackPath, "implementation-spec.md"),
		"quality_profile":     filepath.Join(paths.PackPath, "quality-profile.md"),
	} {
		digest, err := digestFile(path)
		if err != nil {
			missing = append(missing, name)
			blockers = append(blockers, "add "+filepath.Base(path)+" to the Blueprint pack")
			continue
		}
		digests[name] = digest
	}

	var instance Instance
	if err := readJSONIfPossible(paths.InstancePath, &instance); err != nil {
		missing = append(missing, "stack_instance")
		blockers = append(blockers, "provide an AO Atlas stack instance")
	} else if err := ValidateInstance(instance); err != nil {
		missing = append(missing, "stack_instance")
		blockers = append(blockers, "repair stack instance: "+err.Error())
	} else if rules.TargetInstance != "" && instance.ID != rules.TargetInstance {
		missing = append(missing, "stack_instance_scope")
		blockers = append(blockers, "stack instance id must match candidate target_instance")
	} else if digest, err := digestFile(paths.InstancePath); err == nil {
		digests["stack_instance"] = digest
	}

	var mutationModel MutationClassModel
	if err := readJSONIfPossible(paths.MutationClassesPath, &mutationModel); err != nil {
		missing = append(missing, "mutation_class_model")
		blockers = append(blockers, "provide the AO Atlas mutation class model")
	} else if err := ValidateMutationClassModel(mutationModel); err != nil {
		missing = append(missing, "mutation_class_model")
		blockers = append(blockers, "repair mutation class model: "+err.Error())
	} else if !mutationModelIncludes(mutationModel, rules.MutationClass) {
		missing = append(missing, "mutation_class_scope")
		blockers = append(blockers, "mutation class model must include "+rules.MutationClass)
	} else if digest, err := digestFile(paths.MutationClassesPath); err == nil {
		digests["mutation_class_model"] = digest
	}

	var authorization BlueprintBuildAuthorization
	authDigest := ""
	if strings.TrimSpace(paths.AuthorizationPath) == "" {
		missing = append(missing, "build_authorization")
		blockers = append(blockers, "provide AO Blueprint build authorization")
	} else if err := readJSONIfPossible(paths.AuthorizationPath, &authorization); err != nil {
		missing = append(missing, "build_authorization")
		blockers = append(blockers, "provide readable AO Blueprint build authorization")
	} else {
		authDigest, _ = digestFile(paths.AuthorizationPath)
		digests["build_authorization"] = authDigest
		record.BuildAuthorization = SourceRef{Ref: publicArtifactRef(paths.AuthorizationPath), Digest: authDigest}
		authMissing, authBlockers := validateBlueprintAuthorization(authorization, rules, packDigest)
		missing = append(missing, authMissing...)
		blockers = append(blockers, authBlockers...)
	}

	if len(missing) > 0 {
		request := BlueprintRequest{
			ContractVersion: BlueprintRequestContract,
			IntakeID:        firstNonEmpty(record.ProjectID, "blueprint-import") + "-intake",
			Status:          "blueprint_required",
			Missing:         uniqueStrings(missing),
			Reason:          "AO Atlas cannot emit a ready workgraph until Blueprint authorization is present, current, digest-bound, and scoped to this work.",
		}
		record.BlockingNextActions = uniqueStrings(blockers)
		if len(record.BlockingNextActions) == 0 {
			record.BlockingNextActions = []string{"return to AO Blueprint for build authorization"}
		}
		if err := writeBlueprintBlockedArtifacts(paths.OutDir, record, request); err != nil {
			return result, err
		}
		result.Record = record
		result.Request = request
		return result, fmt.Errorf("blueprint import blocked: %s", strings.Join(record.BlockingNextActions, "; "))
	}

	intake := buildBlueprintIntake(rules)
	contextPack, err := buildBlueprintContextPack(paths.PackPath, paths.CandidateRulesPath, rules, digests)
	if err != nil {
		return result, err
	}
	task := buildBlueprintFactoryTask(rules, contextPack)
	workgraph := Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              rules.WorkgraphID,
		TargetInstance:  rules.TargetInstance,
		Nodes: []WorkgraphNode{{
			ID:           rules.CandidateID + "-node",
			Status:       "ready",
			FactoryTask:  task,
			Dependencies: []string{},
			Blockers:     []string{},
			StitchTask:   false,
		}},
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return result, err
	}
	digests["context_pack"] = digestValue(contextPack)
	digests["workgraph"] = digestValue(workgraph)
	candidate := buildBlueprintCandidateSelection(rules, workgraph.Nodes[0], digests)
	digests["candidate_selection"] = digestValue(candidate)

	sourceArtifacts := []SourceRef{
		{Ref: publicArtifactRef(paths.PackPath), Digest: packDigest},
		{Ref: publicArtifactRef(paths.AuthorizationPath), Digest: authDigest},
		{Ref: "candidate-selection.json", Digest: digests["candidate_selection"]},
		{Ref: "context-packs/" + contextPack.ID + ".json", Digest: digests["context_pack"]},
		{Ref: "workgraph.json", Digest: digests["workgraph"]},
	}
	foundryImport, err := BuildFoundryImportForNodes(workgraph, nil, sourceArtifacts)
	if err != nil {
		return result, err
	}
	digests["downstream_foundry_import"] = digestValue(foundryImport)
	record.Status = "ready"
	record.Reason = "Blueprint authorization is ready and Atlas compiled digest-bound Foundry import material."
	record.CandidateSelection = candidate
	record.DownstreamFoundryImport = SourceRef{Ref: "foundry-import/foundry-import.json", Digest: digests["downstream_foundry_import"]}
	record.Digests = digests
	record.ReadyForFoundry = true
	if err := ValidateBlueprintImport(record); err != nil {
		return result, err
	}
	if err := writeBlueprintReadyArtifacts(paths.OutDir, record, intake, candidate, contextPack, workgraph, foundryImport); err != nil {
		return result, err
	}
	result.Record = record
	result.Intake = intake
	result.Candidate = candidate
	result.ContextPacks = []ContextPack{contextPack}
	result.Workgraph = workgraph
	result.FoundryImport = foundryImport
	return result, nil
}

func validateBlueprintAuthorization(auth BlueprintBuildAuthorization, rules BlueprintCandidateRules, packDigest string) ([]string, []string) {
	missing := []string{}
	blockers := []string{}
	if auth.SchemaVersion != "ao.blueprint.build-authorization.v0.1" {
		missing = append(missing, "build_authorization_schema")
		blockers = append(blockers, "Blueprint authorization schema must be ao.blueprint.build-authorization.v0.1")
	}
	if auth.Status != "ready" {
		missing = append(missing, "build_authorization_ready")
		blockers = append(blockers, "Blueprint authorization status must be ready")
	}
	if auth.Score < 100 {
		missing = append(missing, "build_authorization_score")
		blockers = append(blockers, "Blueprint authorization score must be 100")
	}
	if !auth.ApprovedByUser {
		missing = append(missing, "user_approval")
		blockers = append(blockers, "Blueprint authorization must be approved by user")
	}
	if len(auth.BlockingAssumptions) > 0 {
		missing = append(missing, "blocking_assumptions")
		blockers = append(blockers, "Blueprint authorization must not contain blocking assumptions")
	}
	if auth.ProjectID != rules.ProjectID {
		missing = append(missing, "authorization_scope")
		blockers = append(blockers, "Blueprint authorization project_id must match candidate rules")
	}
	if auth.MutationClass != "" && auth.MutationClass != rules.MutationClass {
		missing = append(missing, "authorization_mutation_class")
		blockers = append(blockers, "Blueprint authorization mutation_class must match candidate rules")
	}
	if auth.BlueprintPackDigest != "" && auth.BlueprintPackDigest != packDigest {
		missing = append(missing, "blueprint_pack_digest")
		blockers = append(blockers, "Blueprint authorization blueprint_pack_digest is stale or mismatched")
	}
	if auth.NextAllowedAction != "" && !oneOf(auth.NextAllowedAction, "ao-atlas", "ao-foundry") {
		missing = append(missing, "next_allowed_action")
		blockers = append(blockers, "Blueprint authorization next_allowed_action must route to AO Atlas or legacy AO Foundry intake")
	}
	if rules.MutationClass == "low_risk_code" {
		if !strings.Contains(auth.Scope, "low_risk_code") || !strings.Contains(auth.Scope, "dry_run") {
			missing = append(missing, "authorization_scope")
			blockers = append(blockers, "low_risk_code authorization must be scoped to low_risk_code dry_run")
		}
	}
	if strings.TrimSpace(auth.ExpiresAtUTC) != "" {
		expires, err := time.Parse(time.RFC3339, auth.ExpiresAtUTC)
		if err != nil || !expires.After(time.Now().UTC()) {
			missing = append(missing, "build_authorization_freshness")
			blockers = append(blockers, "Blueprint authorization is stale or has invalid expires_at_utc")
		}
	}
	return uniqueStrings(missing), uniqueStrings(blockers)
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

func writeBlueprintReadyArtifacts(outDir string, record BlueprintImport, intake Intake, candidate BlueprintCandidateSelection, contextPack ContextPack, workgraph Workgraph, foundryImport FoundryImport) error {
	if err := WriteJSON(filepath.Join(outDir, "intake.json"), intake); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "candidate-selection.json"), candidate); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "context-packs", contextPack.ID+".json"), contextPack); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "workgraph.json"), workgraph); err != nil {
		return err
	}
	for _, fixture := range foundryImport.Tasks {
		if err := WriteJSON(filepath.Join(outDir, "foundry-import", fixture.Path), fixture.Task); err != nil {
			return err
		}
	}
	if err := WriteJSON(filepath.Join(outDir, "foundry-import", "foundry-import.json"), foundryImport); err != nil {
		return err
	}
	return WriteJSON(filepath.Join(outDir, "blueprint-import.json"), record)
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

func mutationModelIncludes(model MutationClassModel, class string) bool {
	for _, definition := range model.Classes {
		if definition.Name == class {
			return true
		}
	}
	return false
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
