package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildAOMissionImport(recordPath, commandStatusPath, artifactManifestPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, "")
}

func BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, "", "")
}

func BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath string) (AOMissionImport, error) {
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
	if err := validateAOMissionManifestRefs(manifest, artifactManifestPath); err != nil {
		return AOMissionImport{}, err
	}
	if strings.TrimSpace(routeHistoryPath) != "" {
		if err := validateAOMissionRouteHistory(routeHistoryPath, missionID); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(schedulerRecoveryPath) != "" {
		if err := validateAOMissionReadback(schedulerRecoveryPath, missionID, "ao.mission.scheduler-recovery-readback.v0.1", "scheduler recovery"); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(ledgerCompactionPath) != "" {
		if err := validateAOMissionReadback(ledgerCompactionPath, missionID, "ao.mission.ledger-compaction-readback.v0.1", "ledger compaction"); err != nil {
			return AOMissionImport{}, err
		}
	}
	sources, err := aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath)
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

func BuildAOMissionWorkgraphMetadata(importPath, workgraphPath string) (AOMissionWorkgraphMetadata, error) {
	importRecord, err := LoadJSON[AOMissionImport](importPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	if importRecord.ContractVersion != AOMissionImportContract {
		return AOMissionWorkgraphMetadata{}, fmt.Errorf("invalid AO Mission import contract_version")
	}
	if importRecord.SafeToExecute || importRecord.SchedulesWork || importRecord.ExecutesWork || importRecord.ApprovesWork {
		return AOMissionWorkgraphMetadata{}, fmt.Errorf("AO Mission import must not claim execution, scheduling, or approval authority")
	}
	workgraph, err := LoadJSON[Workgraph](workgraphPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	importDigest, err := digestFile(importPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	workgraphDigest, err := digestFile(workgraphPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	return AOMissionWorkgraphMetadata{
		ContractVersion:   AOMissionWorkgraphMetadataContract,
		MissionID:         importRecord.MissionID,
		WorkgraphID:       workgraph.ID,
		TargetInstance:    workgraph.TargetInstance,
		CurrentRoute:      importRecord.CurrentRoute,
		NodeCounts:        aoMissionWorkgraphNodeCounts(workgraph),
		MissionProvenance: aoMissionProvenanceCounts(importRecord),
		SourceArtifacts: map[string]string{
			"ao_mission_import": importDigest,
			"workgraph":         workgraphDigest,
		},
		NextAction:    "send the first safe Atlas workgraph node to AO Foundry import",
		SafeToExecute: false,
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}, nil
}

func ValidateAOMissionWorkgraphMetadata(metadata AOMissionWorkgraphMetadata, workgraph Workgraph) error {
	if metadata.ContractVersion != AOMissionWorkgraphMetadataContract {
		return fmt.Errorf("invalid AO Mission workgraph metadata contract_version")
	}
	if strings.TrimSpace(metadata.MissionID) == "" {
		return fmt.Errorf("AO Mission workgraph metadata requires mission_id")
	}
	if metadata.WorkgraphID != workgraph.ID {
		return fmt.Errorf("AO Mission workgraph metadata workgraph_id must match workgraph")
	}
	if metadata.TargetInstance != workgraph.TargetInstance {
		return fmt.Errorf("AO Mission workgraph metadata target_instance must match workgraph")
	}
	if metadata.SafeToExecute || metadata.SchedulesWork || metadata.ExecutesWork || metadata.ApprovesWork {
		return fmt.Errorf("AO Mission workgraph metadata must not claim execution, scheduling, or approval authority")
	}
	if metadata.NodeCounts["total"] != len(workgraph.Nodes) {
		return fmt.Errorf("AO Mission workgraph metadata node_counts.total must match workgraph")
	}
	for _, node := range workgraph.Nodes {
		if metadata.NodeCounts[node.Status] == 0 {
			return fmt.Errorf("AO Mission workgraph metadata missing node_counts for status %q", node.Status)
		}
	}
	if len(metadata.SourceArtifacts) == 0 {
		return fmt.Errorf("AO Mission workgraph metadata requires source_artifacts")
	}
	if len(metadata.MissionProvenance) == 0 {
		return fmt.Errorf("AO Mission workgraph metadata requires mission_provenance")
	}
	return nil
}

func aoMissionWorkgraphNodeCounts(workgraph Workgraph) map[string]int {
	counts := map[string]int{"total": len(workgraph.Nodes)}
	for _, node := range workgraph.Nodes {
		counts[node.Status]++
	}
	return counts
}

func aoMissionProvenanceCounts(importRecord AOMissionImport) map[string]int {
	counts := map[string]int{}
	for _, source := range importRecord.SourceArtifacts {
		counts[source.Name]++
	}
	return counts
}

func aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath string) ([]AOMissionSourceArtifact, error) {
	inputs := []struct {
		name string
		path string
	}{
		{name: "mission_record", path: recordPath},
		{name: "command_status", path: commandStatusPath},
		{name: "artifact_manifest", path: artifactManifestPath},
	}
	if strings.TrimSpace(routeHistoryPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "route_history", path: routeHistoryPath})
	}
	if strings.TrimSpace(schedulerRecoveryPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "scheduler_recovery", path: schedulerRecoveryPath})
	}
	if strings.TrimSpace(ledgerCompactionPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "ledger_compaction", path: ledgerCompactionPath})
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

func validateAOMissionRouteHistory(path, missionID string) error {
	var history []map[string]any
	if err := readJSONIfPossible(path, &history); err != nil {
		return err
	}
	if len(history) == 0 {
		return fmt.Errorf("route history requires at least one item")
	}
	for i, item := range history {
		if schema, _ := item["schema"].(string); schema != "ao.mission.route-decision.v0.1" {
			return fmt.Errorf("route history item %d schema must be ao.mission.route-decision.v0.1", i)
		}
		if got, _ := item["mission_id"].(string); got != missionID {
			return fmt.Errorf("route history item %d mission_id mismatch", i)
		}
		for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
			if value, ok := item[field].(bool); ok && value {
				return fmt.Errorf("route history must not claim execution, approval, or repository mutation authority")
			}
		}
	}
	return nil
}

func validateAOMissionReadback(path, missionID, expectedSchema, label string) error {
	var readback map[string]any
	if err := readJSONIfPossible(path, &readback); err != nil {
		return err
	}
	if schema, _ := readback["schema"].(string); schema != expectedSchema {
		return fmt.Errorf("%s schema must be %s", label, expectedSchema)
	}
	if got, _ := readback["mission_id"].(string); got != missionID {
		return fmt.Errorf("%s mission_id mismatch", label)
	}
	for _, field := range []string{"safe_to_execute", "schedules_work", "executes_work", "approves_work", "mutates_repositories", "provider_calls", "release_or_publish", "credential_use", "direct_main_mutation", "concurrent_mutation"} {
		if value, ok := readback[field].(bool); ok && value {
			return fmt.Errorf("%s %s must be false", label, field)
		}
	}
	return nil
}

func validateAOMissionManifestRefs(manifest map[string]any, manifestPath string) error {
	refs, ok := manifest["artifact_refs"].([]any)
	if !ok {
		return nil
	}
	for i, raw := range refs {
		ref, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("artifact manifest artifact_refs[%d] must be an object", i)
		}
		path, _ := ref["ref"].(string)
		if path == "" {
			path, _ = ref["path"].(string)
		}
		want, _ := ref["digest"].(string)
		if want == "" {
			want, _ = ref["sha256"].(string)
		}
		if strings.TrimSpace(path) == "" || strings.TrimSpace(want) == "" {
			return fmt.Errorf("artifact manifest artifact_refs[%d] requires ref/path and digest/sha256", i)
		}
		if !strings.HasPrefix(want, "sha256:") {
			return fmt.Errorf("artifact manifest artifact_refs[%d] digest must start with sha256:", i)
		}
		actualPath, err := resolveAOMissionManifestRef(manifestPath, path)
		if err != nil {
			return fmt.Errorf("artifact manifest ref %q: %w", path, err)
		}
		got, err := digestFile(actualPath)
		if err != nil {
			return fmt.Errorf("artifact manifest ref %q: %w", path, err)
		}
		if got != want {
			return fmt.Errorf("artifact manifest ref %q digest mismatch", path)
		}
	}
	return nil
}

func resolveAOMissionManifestRef(manifestPath, ref string) (string, error) {
	if filepath.IsAbs(ref) {
		if _, err := os.Stat(ref); err != nil {
			return "", err
		}
		return ref, nil
	}
	if _, err := os.Stat(ref); err == nil {
		return ref, nil
	}
	candidate := filepath.Join(filepath.Dir(manifestPath), ref)
	if _, err := os.Stat(candidate); err != nil {
		return "", err
	}
	return candidate, nil
}
