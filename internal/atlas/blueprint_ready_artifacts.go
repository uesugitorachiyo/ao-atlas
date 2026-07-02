package atlas

import "path/filepath"

func writeBlueprintReadyArtifacts(outDir string, record BlueprintImport, intake Intake, candidate BlueprintCandidateSelection, contextPack ContextPack, workgraph Workgraph, foundryImport FoundryImport, handoff FoundryContinuationHandoff) error {
	if err := WriteJSON(filepath.Join(outDir, "intake.json"), intake); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "candidate-selection.json"), candidate); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "context-packs", contextPack.ID+".json"), contextPack); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "workgraph.json"), workgraph); err != nil {
		return err
	}
	if err := writeBlueprintFoundryImportArtifacts(outDir, foundryImport, handoff); err != nil {
		return err
	}
	return WriteJSON(filepath.Join(outDir, "blueprint-import.json"), record)
}

func writeBlueprintFoundryImportArtifacts(outDir string, foundryImport FoundryImport, handoff FoundryContinuationHandoff) error {
	for _, fixture := range foundryImport.Tasks {
		if err := WriteJSON(filepath.Join(outDir, "foundry-import", fixture.Path), fixture.Task); err != nil {
			return err
		}
	}
	if err := WriteJSON(filepath.Join(outDir, "foundry-import", "foundry-import.json"), foundryImport); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "foundry-import", "foundry-continuation-handoff.json"), handoff); err != nil {
		return err
	}
	return WriteFoundryContinuationPrompt(filepath.Join(outDir, "foundry-import", "foundry-continuation-prompt.md"), handoff)
}
