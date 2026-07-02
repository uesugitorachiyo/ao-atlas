package atlas

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
		artifacts = buildBlueprintBlockedCompileArtifacts(artifacts, record, sourceLoad)
		return artifacts, blueprintBlockedCompileError(artifacts.Record)
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
