package atlas

import (
	"fmt"
	"path/filepath"
)

func BuildAOMissionImport(recordPath, commandStatusPath, artifactManifestPath string) (AOMissionImport, error) {
	var record map[string]any
	if err := readJSONIfPossible(recordPath, &record); err != nil {
		return AOMissionImport{}, err
	}
	var commandStatus map[string]any
	if err := readJSONIfPossible(commandStatusPath, &commandStatus); err != nil {
		return AOMissionImport{}, err
	}
	var manifest map[string]any
	if err := readJSONIfPossible(artifactManifestPath, &manifest); err != nil {
		return AOMissionImport{}, err
	}
	missionID, _ := record["mission_id"].(string)
	if missionID == "" {
		return AOMissionImport{}, fmt.Errorf("mission record requires mission_id")
	}
	if commandMissionID, _ := commandStatus["mission_id"].(string); commandMissionID != missionID {
		return AOMissionImport{}, fmt.Errorf("command status mission_id mismatch")
	}
	if manifestMissionID, _ := manifest["mission_id"].(string); manifestMissionID != missionID {
		return AOMissionImport{}, fmt.Errorf("artifact manifest mission_id mismatch")
	}
	for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
		if value, ok := commandStatus[field].(bool); ok && value {
			return AOMissionImport{}, fmt.Errorf("command status %s must be false", field)
		}
	}
	for _, field := range []string{"executes_work", "approves_work"} {
		if value, ok := manifest[field].(bool); ok && value {
			return AOMissionImport{}, fmt.Errorf("artifact manifest %s must be false", field)
		}
	}
	sources, err := aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath)
	if err != nil {
		return AOMissionImport{}, err
	}
	status, _ := record["status"].(string)
	route, _ := record["current_route"].(string)
	return AOMissionImport{
		ContractVersion: AOMissionImportContract,
		MissionID:       missionID,
		Status:          status,
		CurrentRoute:    route,
		SourceArtifacts: sources,
		NextAction:      "compile AO Mission context into Atlas workgraph before Foundry import",
		SafeToExecute:   false,
		SchedulesWork:   false,
		ExecutesWork:    false,
		ApprovesWork:    false,
	}, nil
}

func aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath string) ([]AOMissionSourceArtifact, error) {
	inputs := []struct {
		name string
		path string
	}{
		{name: "mission_record", path: recordPath},
		{name: "command_status", path: commandStatusPath},
		{name: "artifact_manifest", path: artifactManifestPath},
	}
	sources := make([]AOMissionSourceArtifact, 0, len(inputs))
	for _, input := range inputs {
		digest, err := digestFile(input.path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, AOMissionSourceArtifact{Name: input.name, Path: filepath.ToSlash(input.path), SHA256: digest})
	}
	return sources, nil
}
