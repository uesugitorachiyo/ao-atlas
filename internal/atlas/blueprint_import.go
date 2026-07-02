package atlas

func BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error) {
	result := BlueprintImportResult{}
	if err := validateBlueprintImportInputs(paths); err != nil {
		return result, err
	}
	artifacts, compileErr := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	result = blueprintCompileArtifactsToResult(artifacts)
	if err := persistBlueprintImportArtifacts(paths, artifacts, compileErr); err != nil {
		return result, err
	}
	return result, nil
}
