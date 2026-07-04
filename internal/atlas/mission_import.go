package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildAOMissionImport(recordPath, commandStatusPath, artifactManifestPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, "")
}

func BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, "", "")
}

func BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithMissionArchive(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, "")
}

func BuildAOMissionImportWithMissionArchive(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithGatewayReadiness(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath, "")
}

func BuildAOMissionImportWithGatewayReadiness(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithTimelineCompaction(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, "", missionArchivePath, gatewayReadinessRollupPath)
}

func BuildAOMissionImportWithTimelineCompaction(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) (AOMissionImport, error) {
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
	if strings.TrimSpace(timelineCompactionPath) != "" {
		if err := validateAOMissionReadback(timelineCompactionPath, missionID, "ao.mission.timeline-compaction-readback.v0.1", "timeline compaction"); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(missionArchivePath) != "" {
		if err := validateAOMissionArchive(missionArchivePath, missionID); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(gatewayReadinessRollupPath) != "" {
		if err := validateAOMissionReadback(gatewayReadinessRollupPath, missionID, "ao.mission.gateway-readiness-rollup.v0.1", "gateway readiness rollup"); err != nil {
			return AOMissionImport{}, err
		}
	}
	sources, err := aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath)
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
	provenance := aoMissionProvenanceCounts(importRecord)
	return AOMissionWorkgraphMetadata{
		ContractVersion:          AOMissionWorkgraphMetadataContract,
		MissionID:                importRecord.MissionID,
		WorkgraphID:              workgraph.ID,
		TargetInstance:           workgraph.TargetInstance,
		CurrentRoute:             importRecord.CurrentRoute,
		NodeCounts:               aoMissionWorkgraphNodeCounts(workgraph),
		MissionProvenance:        provenance,
		ProvenanceNodes:          sortedMissionProvenanceKeys(provenance),
		PrimaryMissionProvenance: primaryMissionProvenance(provenance),
		ProvenanceDiagnostics:    missionProvenanceDiagnostics(provenance),
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

func primaryMissionProvenance(counts map[string]int) string {
	keys := sortedMissionProvenanceKeys(counts)
	if len(keys) == 0 {
		return ""
	}
	best := keys[0]
	for _, key := range keys[1:] {
		if counts[key] > counts[best] {
			best = key
		}
	}
	return best
}

func missionProvenanceDiagnostics(counts map[string]int) string {
	keys := sortedMissionProvenanceKeys(counts)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, ",")
}

func sortedMissionProvenanceKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) ([]AOMissionSourceArtifact, error) {
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
	if strings.TrimSpace(timelineCompactionPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "timeline_compaction", path: timelineCompactionPath})
	}
	if strings.TrimSpace(missionArchivePath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "mission_archive", path: missionArchivePath})
	}
	if strings.TrimSpace(gatewayReadinessRollupPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "gateway_readiness_rollup", path: gatewayReadinessRollupPath})
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

func validateAOMissionArchive(path, missionID string) error {
	var archive map[string]any
	if err := readJSONIfPossible(path, &archive); err != nil {
		return err
	}
	if schema, _ := archive["schema"].(string); schema != "ao.mission.archive.v0.1" {
		return fmt.Errorf("mission archive schema must be ao.mission.archive.v0.1")
	}
	if archiveMissionID, _ := archive["mission_id"].(string); archiveMissionID != missionID {
		return fmt.Errorf("mission archive mission_id mismatch")
	}
	if strings.TrimSpace(fmt.Sprint(archive["archive_digest"])) == "" {
		return fmt.Errorf("mission archive requires archive_digest")
	}
	for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
		if value, ok := archive[field].(bool); ok && value {
			return fmt.Errorf("mission archive %s must be false", field)
		}
	}
	return nil
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
