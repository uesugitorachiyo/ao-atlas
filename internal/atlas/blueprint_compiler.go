package atlas

import (
	"fmt"
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

	requiredResult := loadBlueprintRequiredArtifacts(paths, digests)
	missing = append(missing, requiredResult.Missing...)
	blockers = append(blockers, requiredResult.Blockers...)

	instanceResult := loadBlueprintInstance(paths, rules, digests)
	missing = append(missing, instanceResult.Missing...)
	blockers = append(blockers, instanceResult.Blockers...)

	mutationModelResult := loadBlueprintMutationModel(paths, rules, digests)
	missing = append(missing, mutationModelResult.Missing...)
	blockers = append(blockers, mutationModelResult.Blockers...)

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
