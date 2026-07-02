package atlas

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func BuildBlueprintImport(paths BlueprintImportPaths) (BlueprintImportResult, error) {
	result := BlueprintImportResult{}
	if strings.TrimSpace(paths.OutDir) == "" {
		return result, errors.New("--out is required")
	}
	artifacts, compileErr := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	result = blueprintCompileArtifactsToResult(artifacts)
	if artifacts.Record.Status == "blocked" {
		if err := writeBlueprintBlockedArtifacts(paths.OutDir, artifacts.Record, artifacts.Request); err != nil {
			return result, err
		}
		return result, compileErr
	}
	if compileErr != nil {
		return result, compileErr
	}
	if len(artifacts.ContextPacks) != 1 {
		return result, fmt.Errorf("blueprint compiler must emit exactly one context pack")
	}
	if err := writeBlueprintReadyArtifacts(paths.OutDir, artifacts.Record, artifacts.Intake, artifacts.Candidate, artifacts.ContextPacks[0], artifacts.Workgraph, artifacts.FoundryImport, artifacts.Handoff); err != nil {
		return result, err
	}
	return result, nil
}

func writeBlueprintBlockedArtifacts(outDir string, record BlueprintImport, request BlueprintRequest) error {
	if err := ValidateBlueprintRequest(request); err != nil {
		return err
	}
	if err := ValidateBlueprintImport(record); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "blueprint-import.json"), record); err != nil {
		return err
	}
	return WriteJSON(filepath.Join(outDir, "blueprint-request.json"), request)
}
