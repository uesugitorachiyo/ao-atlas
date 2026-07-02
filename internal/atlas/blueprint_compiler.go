package atlas

import (
	"fmt"
	"strings"
)

type BlueprintCompileInputs struct {
	Paths BlueprintImportPaths
}

type BlueprintCompiler struct {
	Inputs BlueprintCompileInputs
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
