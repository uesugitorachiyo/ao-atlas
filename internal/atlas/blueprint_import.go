package atlas

import (
	"errors"
	"strings"
)

func BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error) {
	result := BlueprintImportResult{}
	if strings.TrimSpace(paths.OutDir) == "" {
		return result, errors.New("--out is required")
	}
	artifacts, compileErr := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	result = blueprintCompileArtifactsToResult(artifacts)
	if err := persistBlueprintImportArtifacts(paths, artifacts, compileErr); err != nil {
		return result, err
	}
	return result, nil
}
