package atlas

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
