package atlas

import "fmt"

func persistBlueprintImportArtifacts(paths BlueprintImportPaths, artifacts BlueprintCompileArtifacts, compileErr error) error {
	if artifacts.Record.Status == "blocked" {
		if err := writeBlueprintBlockedArtifacts(paths.OutDir, artifacts.Record, artifacts.Request); err != nil {
			return err
		}
		return compileErr
	}
	if compileErr != nil {
		return compileErr
	}
	if len(artifacts.ContextPacks) != 1 {
		return fmt.Errorf("blueprint compiler must emit exactly one context pack")
	}
	return writeBlueprintReadyArtifacts(paths.OutDir, artifacts.Record, artifacts.Intake, artifacts.Candidate, artifacts.ContextPacks[0], artifacts.Workgraph, artifacts.FoundryImport, artifacts.Handoff)
}
