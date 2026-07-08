package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMissionDashboardCompactFilters(nodeID, sourceReadbackPath, sourceWorkgraphPath string) (AtlasMissionDashboardCompactFilters, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	sourceWorkgraphPath = strings.TrimSpace(sourceWorkgraphPath)
	for name, value := range map[string]string{
		"node id":               nodeID,
		"source readback path":  sourceReadbackPath,
		"source workgraph path": sourceWorkgraphPath,
	} {
		if value == "" {
			return AtlasMissionDashboardCompactFilters{}, fmt.Errorf("%s is required", name)
		}
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	workgraph, err := LoadJSON[Workgraph](sourceWorkgraphPath)
	if err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	workgraphDigest, err := digestTextFileWithNormalizedLineEndings(sourceWorkgraphPath)
	if err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}

	statusNodeIDs := missionDashboardNodeIDsByStatus(workgraph)
	filters := []AtlasMissionDashboardCompactFilter{
		buildMissionDashboardCompactFilter("ready", "Ready", statusNodeIDs["ready"], "ready_work_remains", true),
		buildMissionDashboardCompactFilter("blocked", "Blocked", statusNodeIDs["blocked"], "no_blocked_nodes", false),
		buildMissionDashboardCompactFilter("failed", "Failed", statusNodeIDs["failed"], "no_failed_nodes", false),
		buildMissionDashboardCompactFilter("completed", "Completed", statusNodeIDs["completed"], "completed_history", false),
	}
	schemaHealthStatus := strings.TrimSpace(readback.SchemaHealthStatus)
	schemaHealthFilterActionable := missionDashboardSchemaHealthFilterActionable(schemaHealthStatus)
	if schemaHealthStatus != "" {
		filters = append(filters, buildMissionDashboardSchemaHealthFilter(schemaHealthStatus, schemaHealthFilterActionable))
	}

	fixture := AtlasMissionDashboardCompactFilters{
		Schema:                             AtlasMissionDashboardCompactFiltersContract,
		NodeID:                             nodeID,
		Status:                             "compact_dashboard_filters_bound",
		SourceReadbackPath:                 publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:               readbackDigest,
		SourceWorkgraphPath:                publicArtifactRef(sourceWorkgraphPath),
		SourceWorkgraphDigest:              workgraphDigest,
		TotalNodes:                         readback.TotalNodes,
		CompletedNodes:                     readback.CompletedNodes,
		ReadyNodes:                         readback.ReadyNodes,
		BlockedNodes:                       readback.BlockedNodes,
		FailedNodes:                        readback.FailedNodes,
		ExecutableReadyNodes:               readback.ExecutableReadyNodes,
		FirstExecutableNode:                readback.FirstExecutableNode,
		ExactNextAction:                    readback.ExactNextAction,
		ReturnGateStatus:                   readback.ReturnGateStatus,
		SchemaHealthStatus:                 schemaHealthStatus,
		ActiveFilterKey:                    missionDashboardActiveFilterKey(readback),
		FilterCount:                        len(filters),
		Filters:                            filters,
		ReadyFilterActionable:              filters[0].Actionable && filters[0].Count == readback.ReadyNodes,
		BlockedFilterEmpty:                 filters[1].Empty,
		FailedFilterEmpty:                  filters[2].Empty,
		CompletedHistoryAvailable:          filters[3].Count > 0,
		ReadbackCountsMatchWorkgraphCounts: missionDashboardCompactCountsMatch(readback, state),
		FinalResponseAllowed:               readback.FinalResponseAllowed,
		SchedulesWork:                      false,
		ExecutesWork:                       false,
		ApprovesWork:                       false,
		ClaimsAuthorityAdvance:             false,
		RSIRemainsDenied:                   readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if schemaHealthStatus != "" {
		fixture.SchemaHealthFilterKey = "schema_health"
		fixture.SchemaHealthFilterStatus = schemaHealthStatus
		fixture.SchemaHealthFilterActionable = schemaHealthFilterActionable
	}
	if err := ValidateAtlasMissionDashboardCompactFilters(fixture); err != nil {
		return AtlasMissionDashboardCompactFilters{}, err
	}
	return fixture, nil
}

func ValidateAtlasMissionDashboardCompactFilters(fixture AtlasMissionDashboardCompactFilters) error {
	var errs []string
	requireContract(&errs, "mission_dashboard_compact_filters", fixture.Schema, AtlasMissionDashboardCompactFiltersContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "compact_dashboard_filters_bound" {
		errs = append(errs, "status must be compact_dashboard_filters_bound")
	}
	for field, value := range map[string]string{
		"source_readback_path":  fixture.SourceReadbackPath,
		"source_workgraph_path": fixture.SourceWorkgraphPath,
		"first_executable_node": fixture.FirstExecutableNode,
		"exact_next_action":     fixture.ExactNextAction,
		"return_gate_status":    fixture.ReturnGateStatus,
		"schema_health_status":  fixture.SchemaHealthStatus,
		"active_filter_key":     fixture.ActiveFilterKey,
	} {
		if field != "schema_health_status" {
			requireField(&errs, field, value)
		}
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  fixture.SourceReadbackDigest,
		"source_workgraph_digest": fixture.SourceWorkgraphDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if fixture.TotalNodes != 40 || fixture.CompletedNodes != 35 || fixture.ReadyNodes != 5 || fixture.BlockedNodes != 0 || fixture.FailedNodes != 0 {
		errs = append(errs, "source node counts must be total_nodes=40 completed_nodes=35 ready_nodes=5 blocked_nodes=0 failed_nodes=0")
	}
	if fixture.ExecutableReadyNodes != 1 {
		errs = append(errs, "executable_ready_nodes must be 1")
	}
	if fixture.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-36" {
		errs = append(errs, "first_executable_node must be mission-recommendation-feature-depth-next-wave-36")
	}
	if fixture.ActiveFilterKey != "ready" {
		errs = append(errs, "active_filter_key must be ready while ready work remains")
	}
	if fixture.ReturnGateStatus != "blocked_ready_nodes_remain" {
		errs = append(errs, "return_gate_status must be blocked_ready_nodes_remain")
	}
	hasSchemaHealthStatus := strings.TrimSpace(fixture.SchemaHealthStatus) != ""
	expectedFilterCount := 4
	if hasSchemaHealthStatus {
		expectedFilterCount = 5
	}
	if fixture.FilterCount != len(fixture.Filters) || fixture.FilterCount != expectedFilterCount {
		errs = append(errs, fmt.Sprintf("filter_count must be %d", expectedFilterCount))
	}
	validateMissionDashboardSchemaHealthFilterBinding(&errs, fixture, hasSchemaHealthStatus)
	validateMissionDashboardCompactFilterRows(&errs, fixture.Filters, fixture.SchemaHealthStatus)
	if !fixture.ReadyFilterActionable {
		errs = append(errs, "ready_filter_actionable must be true")
	}
	if !fixture.BlockedFilterEmpty {
		errs = append(errs, "blocked_filter_empty must be true")
	}
	if !fixture.FailedFilterEmpty {
		errs = append(errs, "failed_filter_empty must be true")
	}
	if !fixture.CompletedHistoryAvailable {
		errs = append(errs, "completed_history_available must be true")
	}
	if !fixture.ReadbackCountsMatchWorkgraphCounts {
		errs = append(errs, "readback_counts_match_workgraph_counts must be true")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func buildMissionDashboardCompactFilter(key, label string, ids []string, populatedStatus string, actionableWhenPopulated bool) AtlasMissionDashboardCompactFilter {
	preview := missionDashboardCompactPreviewIDs(key, ids)
	firstNode, lastNode := missionDashboardFirstLastNodeIDs(ids)
	count := len(ids)
	omitted := count - len(preview)
	if omitted < 0 {
		omitted = 0
	}
	status := populatedStatus
	if count == 0 && key == "ready" {
		status = "no_ready_nodes"
	}
	if count > 0 && key == "blocked" {
		status = "blocked_nodes_present"
	}
	if count > 0 && key == "failed" {
		status = "failed_nodes_present"
	}
	return AtlasMissionDashboardCompactFilter{
		Key:              key,
		Label:            label,
		Count:            count,
		PreviewNodeIDs:   preview,
		OmittedNodeCount: omitted,
		FirstNodeID:      firstNode,
		LastNodeID:       lastNode,
		Actionable:       actionableWhenPopulated && count > 0,
		Empty:            count == 0,
		DashboardStatus:  status,
	}
}

func buildMissionDashboardSchemaHealthFilter(status string, actionable bool) AtlasMissionDashboardCompactFilter {
	return AtlasMissionDashboardCompactFilter{
		Key:             "schema_health",
		Label:           "Schema Health",
		Count:           1,
		PreviewNodeIDs:  []string{},
		Actionable:      actionable,
		Empty:           false,
		DashboardStatus: status,
	}
}

func missionDashboardSchemaHealthFilterActionable(status string) bool {
	status = strings.TrimSpace(status)
	if status == "" {
		return false
	}
	return strings.HasPrefix(status, "failed") ||
		strings.HasPrefix(status, "pending") ||
		strings.Contains(status, "missing")
}

func validateMissionDashboardSchemaHealthFilterBinding(errs *[]string, fixture AtlasMissionDashboardCompactFilters, hasSchemaHealthStatus bool) {
	for field, value := range map[string]string{
		"schema_health_filter_key":    fixture.SchemaHealthFilterKey,
		"schema_health_filter_status": fixture.SchemaHealthFilterStatus,
	} {
		checkPublicPath(errs, field, value, true)
	}
	if !hasSchemaHealthStatus {
		if fixture.SchemaHealthFilterKey != "" || fixture.SchemaHealthFilterStatus != "" || fixture.SchemaHealthFilterActionable {
			*errs = append(*errs, "schema health filter fields require schema_health_status")
		}
		return
	}
	if fixture.SchemaHealthFilterKey != "schema_health" {
		*errs = append(*errs, "schema_health_filter_key must be schema_health")
	}
	if fixture.SchemaHealthFilterStatus != fixture.SchemaHealthStatus {
		*errs = append(*errs, "schema_health_filter_status must match schema_health_status")
	}
	if fixture.SchemaHealthFilterActionable != missionDashboardSchemaHealthFilterActionable(fixture.SchemaHealthStatus) {
		*errs = append(*errs, "schema_health_filter_actionable must match schema health status")
	}
}

func validateMissionDashboardCompactFilterRows(errs *[]string, filters []AtlasMissionDashboardCompactFilter, schemaHealthStatus string) {
	expected := []struct {
		key              string
		label            string
		count            int
		actionable       bool
		empty            bool
		dashboardStatus  string
		previewLimit     int
		firstNodeID      string
		lastNodeID       string
		omittedNodeCount int
	}{
		{"ready", "Ready", 5, true, false, "ready_work_remains", 5, "mission-recommendation-feature-depth-next-wave-36", "mission-recommendation-feature-depth-next-wave-40", 0},
		{"blocked", "Blocked", 0, false, true, "no_blocked_nodes", 5, "", "", 0},
		{"failed", "Failed", 0, false, true, "no_failed_nodes", 5, "", "", 0},
		{"completed", "Completed", 35, false, false, "completed_history", 3, "mission-recommendation-feature-depth-next-wave-01", "mission-recommendation-feature-depth-next-wave-35", 32},
	}
	if schemaHealthStatus != "" {
		expected = append(expected, struct {
			key              string
			label            string
			count            int
			actionable       bool
			empty            bool
			dashboardStatus  string
			previewLimit     int
			firstNodeID      string
			lastNodeID       string
			omittedNodeCount int
		}{"schema_health", "Schema Health", 1, missionDashboardSchemaHealthFilterActionable(schemaHealthStatus), false, schemaHealthStatus, 0, "", "", 0})
	}
	if len(filters) != len(expected) {
		*errs = append(*errs, "filters must contain ready, blocked, failed, completed, and optional schema health rows")
		return
	}
	for i, want := range expected {
		filter := filters[i]
		prefix := fmt.Sprintf("filters[%d]", i)
		requireField(errs, prefix+".key", filter.Key)
		requireField(errs, prefix+".label", filter.Label)
		requireField(errs, prefix+".dashboard_status", filter.DashboardStatus)
		checkPublicPath(errs, prefix+".key", filter.Key, true)
		checkPublicPath(errs, prefix+".label", filter.Label, true)
		checkPublicPath(errs, prefix+".dashboard_status", filter.DashboardStatus, true)
		if filter.Key != want.key || filter.Label != want.label {
			*errs = append(*errs, prefix+" key and label must match compact filter order")
		}
		if filter.Count != want.count {
			*errs = append(*errs, prefix+".count must match source workgraph status count")
		}
		if filter.Actionable != want.actionable {
			*errs = append(*errs, prefix+".actionable must match expected compact filter behavior")
		}
		if filter.Empty != want.empty {
			*errs = append(*errs, prefix+".empty must match count")
		}
		if filter.DashboardStatus != want.dashboardStatus {
			*errs = append(*errs, prefix+".dashboard_status must be "+want.dashboardStatus)
		}
		if len(filter.PreviewNodeIDs) > want.previewLimit {
			*errs = append(*errs, prefix+".preview_node_ids exceeds compact preview limit")
		}
		for j, id := range filter.PreviewNodeIDs {
			if strings.TrimSpace(id) == "" {
				*errs = append(*errs, fmt.Sprintf("%s.preview_node_ids[%d] is required", prefix, j))
			}
			checkPublicPath(errs, fmt.Sprintf("%s.preview_node_ids[%d]", prefix, j), id, true)
		}
		if filter.OmittedNodeCount != want.omittedNodeCount {
			*errs = append(*errs, prefix+".omitted_node_count must match compact preview")
		}
		if filter.FirstNodeID != want.firstNodeID || filter.LastNodeID != want.lastNodeID {
			*errs = append(*errs, prefix+" first and last node ids must match source workgraph order")
		}
		if filter.Count == 0 && (len(filter.PreviewNodeIDs) != 0 || filter.OmittedNodeCount != 0 || filter.FirstNodeID != "" || filter.LastNodeID != "") {
			*errs = append(*errs, prefix+" empty filters must not carry node ids")
		}
	}
}

func missionDashboardNodeIDsByStatus(workgraph Workgraph) map[string][]string {
	ids := map[string][]string{
		"ready":     {},
		"blocked":   {},
		"failed":    {},
		"completed": {},
	}
	for _, node := range workgraph.Nodes {
		if _, ok := ids[node.Status]; ok {
			ids[node.Status] = append(ids[node.Status], node.ID)
		}
	}
	return ids
}

func missionDashboardCompactPreviewIDs(key string, ids []string) []string {
	limit := 5
	if key == "completed" {
		limit = 3
	}
	if len(ids) <= limit {
		return append([]string(nil), ids...)
	}
	if key == "completed" {
		return append([]string(nil), ids[len(ids)-limit:]...)
	}
	return append([]string(nil), ids[:limit]...)
}

func missionDashboardFirstLastNodeIDs(ids []string) (string, string) {
	if len(ids) == 0 {
		return "", ""
	}
	return ids[0], ids[len(ids)-1]
}

func missionDashboardActiveFilterKey(readback AtlasRecommendationReadback) string {
	if readback.ReadyNodes > 0 {
		return "ready"
	}
	if readback.BlockedNodes > 0 {
		return "blocked"
	}
	if readback.FailedNodes > 0 {
		return "failed"
	}
	return "completed"
}

func missionDashboardCompactCountsMatch(readback AtlasRecommendationReadback, state WorkgraphState) bool {
	return readback.CompletedNodes == state.NodeCounts["completed"] &&
		readback.ReadyNodes == state.NodeCounts["ready"] &&
		readback.BlockedNodes == state.NodeCounts["blocked"] &&
		readback.FailedNodes == state.NodeCounts["failed"]
}

func WriteAtlasMissionDashboardCompactFilters(path string, fixture AtlasMissionDashboardCompactFilters) error {
	return WriteJSON(path, fixture)
}
