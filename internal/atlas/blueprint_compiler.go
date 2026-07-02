package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

type BlueprintCompileInputs struct {
	Paths BlueprintImportPaths
}

type BlueprintCompileArtifacts struct {
	Record        BlueprintImport
	Request       BlueprintRequest
	Intake        Intake
	Candidate     BlueprintCandidateSelection
	ContextPacks  []ContextPack
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Handoff       FoundryContinuationHandoff
}

type BlueprintCompiler struct {
	Inputs BlueprintCompileInputs
}

func blueprintCompileArtifactsToResult(artifacts BlueprintCompileArtifacts) BlueprintImportResult {
	return BlueprintImportResult{
		Record:        artifacts.Record,
		Request:       artifacts.Request,
		Intake:        artifacts.Intake,
		Candidate:     artifacts.Candidate,
		ContextPacks:  artifacts.ContextPacks,
		Workgraph:     artifacts.Workgraph,
		FoundryImport: artifacts.FoundryImport,
		Handoff:       artifacts.Handoff,
	}
}

func (compiler BlueprintCompiler) Compile() (BlueprintCompileArtifacts, error) {
	paths := compiler.Inputs.Paths
	state := newBlockedBlueprintCompileState(paths)
	artifacts := state.Artifacts
	record := state.Record
	digests := state.Digests
	packDigest := state.PackDigest
	packErr := state.PackErr
	missing := []string{}
	blockers := []string{}
	if packErr != nil {
		missing = append(missing, "blueprint_pack")
		blockers = append(blockers, "provide a readable AO Blueprint pack")
	}

	rulesResult := loadBlueprintCandidateRules(paths, record, digests)
	rules := rulesResult.Rules
	record = rulesResult.Record
	missing = append(missing, rulesResult.Missing...)
	blockers = append(blockers, rulesResult.Blockers...)

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
		artifacts.Record = record
		artifacts.Request = request
		return artifacts, fmt.Errorf("blueprint import blocked: %s", strings.Join(record.BlockingNextActions, "; "))
	}

	intake := buildBlueprintIntake(rules)
	contextPack, err := buildBlueprintContextPack(paths.PackPath, paths.CandidateRulesPath, rules, digests)
	if err != nil {
		return artifacts, err
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
		return artifacts, err
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
		return artifacts, err
	}
	digests["downstream_foundry_import"] = digestValue(foundryImport)
	handoff, err := BuildFoundryContinuationHandoff(workgraph, foundryImport, FoundryContinuationHandoffInputs{
		BlueprintPackPath: publicArtifactRef(paths.PackPath),
		AtlasImportPath:   "blueprint-import.json",
		WorkgraphPath:     "workgraph.json",
		FoundryImportPath: "foundry-import/foundry-import.json",
	})
	if err != nil {
		return artifacts, err
	}
	digests["downstream_foundry_continuation_handoff"] = digestValue(handoff)
	record.Status = "ready"
	record.Reason = "Blueprint authorization is ready and Atlas compiled digest-bound Foundry import material."
	record.CandidateSelection = candidate
	record.DownstreamFoundryImport = SourceRef{Ref: "foundry-import/foundry-import.json", Digest: digests["downstream_foundry_import"]}
	record.DownstreamFoundryContinuationHandoff = SourceRef{Ref: "foundry-import/foundry-continuation-handoff.json", Digest: digests["downstream_foundry_continuation_handoff"]}
	record.Digests = digests
	record.ReadyForFoundry = true
	if err := ValidateBlueprintImport(record); err != nil {
		return artifacts, err
	}
	artifacts.Record = record
	artifacts.Intake = intake
	artifacts.Candidate = candidate
	artifacts.ContextPacks = []ContextPack{contextPack}
	artifacts.Workgraph = workgraph
	artifacts.FoundryImport = foundryImport
	artifacts.Handoff = handoff
	return artifacts, nil
}
