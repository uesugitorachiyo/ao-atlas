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

	authorizationResult := loadBlueprintAuthorization(paths, record, rules, packDigest, digests)
	record = authorizationResult.Record
	authDigest := authorizationResult.AuthDigest
	missing = append(missing, authorizationResult.Missing...)
	blockers = append(blockers, authorizationResult.Blockers...)

	if len(missing) > 0 {
		blockedRequest := buildBlueprintBlockedRequest(record, missing, blockers)
		record = blockedRequest.Record
		artifacts.Record = blockedRequest.Record
		artifacts.Request = blockedRequest.Request
		return artifacts, fmt.Errorf("blueprint import blocked: %s", strings.Join(record.BlockingNextActions, "; "))
	}

	readyArtifacts, err := buildBlueprintReadyMaterial(blueprintReadyMaterialInputs{
		Paths:      paths,
		Record:     record,
		Rules:      rules,
		Digests:    digests,
		PackDigest: packDigest,
		AuthDigest: authDigest,
	})
	if err != nil {
		return artifacts, err
	}
	return readyArtifacts, nil
}
