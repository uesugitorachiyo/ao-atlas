package atlas

type blueprintImportCompilation struct {
	Artifacts  BlueprintCompileArtifacts
	Result     BlueprintImportResult
	CompileErr error
}

func compileBlueprintImportArtifacts(paths BlueprintImportPaths) blueprintImportCompilation {
	artifacts, compileErr := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	return blueprintImportCompilation{
		Artifacts:  artifacts,
		Result:     blueprintCompileArtifactsToResult(artifacts),
		CompileErr: compileErr,
	}
}
