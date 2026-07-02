package atlas

type blueprintDownstreamFoundryResult struct {
	FoundryImport FoundryImport
	Handoff       FoundryContinuationHandoff
}

func buildBlueprintDownstreamFoundry(workgraph Workgraph, sourceArtifacts []SourceRef, paths BlueprintImportPaths) (blueprintDownstreamFoundryResult, error) {
	foundryImport, err := BuildFoundryImportForNodes(workgraph, nil, sourceArtifacts)
	if err != nil {
		return blueprintDownstreamFoundryResult{}, err
	}
	handoff, err := BuildFoundryContinuationHandoff(workgraph, foundryImport, FoundryContinuationHandoffInputs{
		BlueprintPackPath: publicArtifactRef(paths.PackPath),
		AtlasImportPath:   "blueprint-import.json",
		WorkgraphPath:     "workgraph.json",
		FoundryImportPath: "foundry-import/foundry-import.json",
	})
	if err != nil {
		return blueprintDownstreamFoundryResult{}, err
	}
	return blueprintDownstreamFoundryResult{FoundryImport: foundryImport, Handoff: handoff}, nil
}
