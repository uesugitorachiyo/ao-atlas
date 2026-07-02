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
	sourceLoad := loadBlueprintCompileSources(blueprintCompileSourceInputs{
		Paths:      paths,
		Record:     record,
		Digests:    digests,
		PackDigest: packDigest,
		PackErr:    state.PackErr,
	})
	record = sourceLoad.Record

	if len(sourceLoad.Missing) > 0 {
		blockedRequest := buildBlueprintBlockedRequest(record, sourceLoad.Missing, sourceLoad.Blockers)
		record = blockedRequest.Record
		artifacts.Record = blockedRequest.Record
		artifacts.Request = blockedRequest.Request
		return artifacts, fmt.Errorf("blueprint import blocked: %s", strings.Join(record.BlockingNextActions, "; "))
	}

	readyArtifacts, err := buildBlueprintReadyMaterial(blueprintReadyMaterialInputs{
		Paths:      paths,
		Record:     record,
		Rules:      sourceLoad.Rules,
		Digests:    digests,
		PackDigest: packDigest,
		AuthDigest: sourceLoad.AuthDigest,
	})
	if err != nil {
		return artifacts, err
	}
	return readyArtifacts, nil
}
