package atlas

func buildBlueprintReadyCompileArtifacts(paths BlueprintImportPaths, record BlueprintImport, digests map[string]string, packDigest string, sourceLoad blueprintCompileSourceLoad) (BlueprintCompileArtifacts, error) {
	return buildBlueprintReadyMaterial(blueprintReadyMaterialInputs{
		Paths:      paths,
		Record:     record,
		Rules:      sourceLoad.Rules,
		Digests:    digests,
		PackDigest: packDigest,
		AuthDigest: sourceLoad.AuthDigest,
	})
}
