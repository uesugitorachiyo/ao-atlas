package atlas

func BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error) {
	result := BlueprintImportResult{}
	if err := validateBlueprintImportInputs(paths); err != nil {
		return result, err
	}
	compilation := compileBlueprintImportArtifacts(paths)
	if err := persistBlueprintImportArtifacts(paths, compilation.Artifacts, compilation.CompileErr); err != nil {
		return compilation.Result, err
	}
	return compilation.Result, nil
}
