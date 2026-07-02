package atlas

type blueprintReadyMaterialInputs struct {
	Paths      BlueprintImportPaths
	Record     BlueprintImport
	Rules      BlueprintCandidateRules
	Digests    map[string]string
	PackDigest string
	AuthDigest string
}

func buildBlueprintReadyMaterial(inputs blueprintReadyMaterialInputs) (BlueprintCompileArtifacts, error) {
	contextPack, err := buildBlueprintContextPack(inputs.Paths.PackPath, inputs.Paths.CandidateRulesPath, inputs.Rules, inputs.Digests)
	if err != nil {
		return BlueprintCompileArtifacts{}, err
	}
	task := buildBlueprintFactoryTask(inputs.Rules, contextPack)
	workgraph, err := buildBlueprintWorkgraph(inputs.Rules, task)
	if err != nil {
		return BlueprintCompileArtifacts{}, err
	}
	inputs.Digests["context_pack"] = digestValue(contextPack)
	inputs.Digests["workgraph"] = digestValue(workgraph)
	candidate := buildBlueprintCandidateSelection(inputs.Rules, workgraph.Nodes[0], inputs.Digests)
	inputs.Digests["candidate_selection"] = digestValue(candidate)

	sourceArtifacts := buildBlueprintFoundrySourceArtifacts(inputs.Paths, contextPack, inputs.PackDigest, inputs.AuthDigest, inputs.Digests)
	downstreamFoundry, err := buildBlueprintDownstreamFoundry(workgraph, sourceArtifacts, inputs.Paths)
	if err != nil {
		return BlueprintCompileArtifacts{}, err
	}
	foundryImport := downstreamFoundry.FoundryImport
	handoff := downstreamFoundry.Handoff
	inputs.Digests["downstream_foundry_import"] = digestValue(foundryImport)
	inputs.Digests["downstream_foundry_continuation_handoff"] = digestValue(handoff)
	record, err := buildBlueprintReadyRecord(blueprintReadyRecordInputs{
		Record:        inputs.Record,
		Candidate:     candidate,
		Digests:       inputs.Digests,
		FoundryDigest: inputs.Digests["downstream_foundry_import"],
		HandoffDigest: inputs.Digests["downstream_foundry_continuation_handoff"],
	})
	if err != nil {
		return BlueprintCompileArtifacts{}, err
	}
	return BlueprintCompileArtifacts{
		Record:        record,
		Intake:        buildBlueprintIntake(inputs.Rules),
		Candidate:     candidate,
		ContextPacks:  []ContextPack{contextPack},
		Workgraph:     workgraph,
		FoundryImport: foundryImport,
		Handoff:       handoff,
	}, nil
}
