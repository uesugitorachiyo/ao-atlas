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
	schemaHealthFilterState := missionDashboardSchemaHealthFilterState(schemaHealthStatus)
	schemaHealthFilterActionable := missionDashboardSchemaHealthFilterActionable(schemaHealthStatus)
	filters = append(filters,
		buildMissionDashboardStateFilter("recommendation_track", "Track", missionDashboardRecommendationTrackStatus(readback), false),
		buildMissionDashboardSchemaHealthFilter(schemaHealthStatus, schemaHealthFilterActionable),
		buildMissionDashboardStateFilter("ci_state", "CI State", missionDashboardCIStateStatus(readback), missionDashboardCIStateActionable(readback)),
		buildMissionDashboardStateFilter("cleanup_state", "Cleanup State", missionDashboardCleanupStateStatus(readback), missionDashboardCleanupStateActionable(readback)),
	)

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
		fixture.SchemaHealthFilterState = schemaHealthFilterState
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
	if fixture.TotalNodes <= 0 {
		errs = append(errs, "total_nodes must be greater than zero")
	}
	if fixture.CompletedNodes < 0 || fixture.ReadyNodes < 0 || fixture.BlockedNodes < 0 || fixture.FailedNodes < 0 || fixture.ExecutableReadyNodes < 0 {
		errs = append(errs, "source node counts must not be negative")
	}
	if fixture.TotalNodes != fixture.CompletedNodes+fixture.ReadyNodes+fixture.BlockedNodes+fixture.FailedNodes {
		errs = append(errs, "total_nodes must equal completed, ready, blocked, and failed node counts")
	}
	if fixture.ExecutableReadyNodes > fixture.ReadyNodes {
		errs = append(errs, "executable_ready_nodes must not exceed ready_nodes")
	}
	if fixture.ExecutableReadyNodes > 0 {
		requireField(&errs, "first_executable_node", fixture.FirstExecutableNode)
	}
	checkPublicPath(&errs, "first_executable_node", fixture.FirstExecutableNode, true)
	expectedActiveFilter := missionDashboardActiveFilterKeyFromCounts(fixture.ReadyNodes, fixture.BlockedNodes, fixture.FailedNodes)
	if fixture.ActiveFilterKey != expectedActiveFilter {
		errs = append(errs, "active_filter_key must match source node counts")
	}
	hasSchemaHealthStatus := strings.TrimSpace(fixture.SchemaHealthStatus) != ""
	expectedFilterCount := 8
	if fixture.FilterCount != len(fixture.Filters) || fixture.FilterCount != expectedFilterCount {
		errs = append(errs, fmt.Sprintf("filter_count must be %d", expectedFilterCount))
	}
	validateMissionDashboardSchemaHealthFilterBinding(&errs, fixture, hasSchemaHealthStatus)
	validateMissionDashboardCompactFilterRows(&errs, fixture)
	readyFilter := missionDashboardCompactFilterByKey(fixture.Filters, "ready")
	if fixture.ReadyFilterActionable != (readyFilter.Actionable && readyFilter.Count == fixture.ReadyNodes) {
		errs = append(errs, "ready_filter_actionable must match ready filter row")
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
	if fixture.FinalResponseAllowed && (fixture.ReadyNodes > 0 || fixture.BlockedNodes > 0 || fixture.FailedNodes > 0) {
		errs = append(errs, "final_response_allowed requires no ready, blocked, or failed nodes")
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
	status = strings.TrimSpace(status)
	if status == "" {
		return AtlasMissionDashboardCompactFilter{
			Key:             "schema_health",
			Label:           "Schema Health",
			Count:           0,
			PreviewNodeIDs:  []string{},
			Actionable:      false,
			Empty:           true,
			DashboardStatus: "schema_health_not_reported",
		}
	}
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

func buildMissionDashboardStateFilter(key, label, status string, actionable bool) AtlasMissionDashboardCompactFilter {
	return AtlasMissionDashboardCompactFilter{
		Key:             key,
		Label:           label,
		Count:           1,
		PreviewNodeIDs:  []string{},
		Actionable:      actionable,
		Empty:           false,
		DashboardStatus: status,
	}
}

func missionDashboardRecommendationTrackStatus(readback AtlasRecommendationReadback) string {
	source := strings.ToLower(readback.TargetInstance + " " + readback.MissionID)
	switch {
	case strings.Contains(source, "feature-depth"):
		return "track_feature_depth"
	case strings.Contains(source, "refactoring"):
		return "track_refactoring"
	case strings.Contains(source, "hardening"):
		return "track_hardening"
	case strings.Contains(source, "final-closure") || strings.Contains(source, "final_closure"):
		return "track_final_closure"
	default:
		return "track_unknown"
	}
}

func missionDashboardCIStateStatus(readback AtlasRecommendationReadback) string {
	if readback.FailedNodes > 0 || strings.Contains(readback.PublicSafetyScanStatus, "failed") {
		return "ci_state_failed"
	}
	if strings.Contains(readback.FoundryRollupStatus, "completed") &&
		strings.Contains(readback.CommandReadbackStatus, "compact") &&
		readback.PublicSafetyScanStatus == "passed" {
		return "ci_state_passed"
	}
	return "ci_state_pending_remote_lifecycle"
}

func missionDashboardCIStateActionable(readback AtlasRecommendationReadback) bool {
	return missionDashboardCIStateStatus(readback) != "ci_state_passed"
}

func missionDashboardCleanupStateStatus(readback AtlasRecommendationReadback) string {
	if readback.FinalResponseAllowed && readback.ReadyNodes == 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		return "cleanup_state_complete"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		return "cleanup_state_blocked"
	}
	return "cleanup_state_pending_ready_work"
}

func missionDashboardCleanupStateActionable(readback AtlasRecommendationReadback) bool {
	return missionDashboardCleanupStateStatus(readback) != "cleanup_state_complete"
}

func missionDashboardSchemaHealthFilterActionable(status string) bool {
	state := missionDashboardSchemaHealthFilterState(status)
	return state == "failed" || state == "pending"
}

func missionDashboardSchemaHealthFilterState(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}
	if strings.HasPrefix(status, "failed") || strings.Contains(status, "missing") {
		return "failed"
	}
	if strings.HasPrefix(status, "ready") ||
		strings.HasPrefix(status, "passed") ||
		strings.Contains(status, "complete") {
		return "ready"
	}
	return "pending"
}

func validateMissionDashboardSchemaHealthFilterBinding(errs *[]string, fixture AtlasMissionDashboardCompactFilters, hasSchemaHealthStatus bool) {
	for field, value := range map[string]string{
		"schema_health_filter_key":    fixture.SchemaHealthFilterKey,
		"schema_health_filter_status": fixture.SchemaHealthFilterStatus,
		"schema_health_filter_state":  fixture.SchemaHealthFilterState,
	} {
		checkPublicPath(errs, field, value, true)
	}
	if !hasSchemaHealthStatus {
		if fixture.SchemaHealthFilterKey != "" || fixture.SchemaHealthFilterStatus != "" || fixture.SchemaHealthFilterState != "" || fixture.SchemaHealthFilterActionable {
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
	if fixture.SchemaHealthFilterState != missionDashboardSchemaHealthFilterState(fixture.SchemaHealthStatus) {
		*errs = append(*errs, "schema_health_filter_state must classify schema_health_status")
	}
	if fixture.SchemaHealthFilterActionable != missionDashboardSchemaHealthFilterActionable(fixture.SchemaHealthStatus) {
		*errs = append(*errs, "schema_health_filter_actionable must match schema health status")
	}
}

func validateMissionDashboardCompactFilterRows(errs *[]string, fixture AtlasMissionDashboardCompactFilters) {
	expected := []struct {
		key             string
		label           string
		count           int
		actionable      bool
		empty           bool
		dashboardStatus string
		previewLimit    int
	}{
		{"ready", "Ready", fixture.ReadyNodes, fixture.ReadyNodes > 0, fixture.ReadyNodes == 0, missionDashboardStatusFilterStatus("ready", fixture.ReadyNodes), 5},
		{"blocked", "Blocked", fixture.BlockedNodes, false, fixture.BlockedNodes == 0, missionDashboardStatusFilterStatus("blocked", fixture.BlockedNodes), 5},
		{"failed", "Failed", fixture.FailedNodes, false, fixture.FailedNodes == 0, missionDashboardStatusFilterStatus("failed", fixture.FailedNodes), 5},
		{"completed", "Completed", fixture.CompletedNodes, false, fixture.CompletedNodes == 0, "completed_history", 3},
		{"recommendation_track", "Track", 1, false, false, "", 0},
	}
	schemaHealthExpectedStatus := "schema_health_not_reported"
	schemaHealthExpectedCount := 0
	schemaHealthExpectedEmpty := true
	if fixture.SchemaHealthStatus != "" {
		schemaHealthExpectedStatus = fixture.SchemaHealthStatus
		schemaHealthExpectedCount = 1
		schemaHealthExpectedEmpty = false
	}
	expected = append(expected,
		struct {
			key             string
			label           string
			count           int
			actionable      bool
			empty           bool
			dashboardStatus string
			previewLimit    int
		}{"schema_health", "Schema Health", schemaHealthExpectedCount, missionDashboardSchemaHealthFilterActionable(fixture.SchemaHealthStatus), schemaHealthExpectedEmpty, schemaHealthExpectedStatus, 0},
		struct {
			key             string
			label           string
			count           int
			actionable      bool
			empty           bool
			dashboardStatus string
			previewLimit    int
		}{"ci_state", "CI State", 1, false, false, "", 0},
		struct {
			key             string
			label           string
			count           int
			actionable      bool
			empty           bool
			dashboardStatus string
			previewLimit    int
		}{"cleanup_state", "Cleanup State", 1, false, false, missionDashboardCleanupStateStatusFromCounts(fixture), 0},
	)
	if len(fixture.Filters) != len(expected) {
		*errs = append(*errs, "filters must contain ready, blocked, failed, completed, recommendation track, schema health, ci state, and cleanup state rows")
		return
	}
	for i, want := range expected {
		filter := fixture.Filters[i]
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
			if want.key == "ci_state" {
				if filter.Actionable != (filter.DashboardStatus != "ci_state_passed") {
					*errs = append(*errs, prefix+".actionable must match CI state")
				}
			} else if want.key == "cleanup_state" {
				if filter.Actionable != (filter.DashboardStatus != "cleanup_state_complete") {
					*errs = append(*errs, prefix+".actionable must match cleanup state")
				}
			} else {
				*errs = append(*errs, prefix+".actionable must match expected compact filter behavior")
			}
		}
		if filter.Empty != want.empty {
			*errs = append(*errs, prefix+".empty must match count")
		}
		if want.key == "recommendation_track" {
			if !strings.HasPrefix(filter.DashboardStatus, "track_") {
				*errs = append(*errs, prefix+".dashboard_status must describe recommendation track")
			}
		} else if want.key == "ci_state" {
			if !oneOf(filter.DashboardStatus, "ci_state_failed", "ci_state_pending_remote_lifecycle", "ci_state_passed") {
				*errs = append(*errs, prefix+".dashboard_status must describe CI state")
			}
		} else if filter.DashboardStatus != want.dashboardStatus {
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
		expectedOmittedNodeCount := 0
		if want.previewLimit > 0 {
			expectedOmittedNodeCount = filter.Count - len(filter.PreviewNodeIDs)
			if expectedOmittedNodeCount < 0 {
				expectedOmittedNodeCount = 0
			}
		}
		if filter.OmittedNodeCount != expectedOmittedNodeCount {
			*errs = append(*errs, prefix+".omitted_node_count must match compact preview")
		}
		if filter.Count == 0 && (len(filter.PreviewNodeIDs) != 0 || filter.OmittedNodeCount != 0 || filter.FirstNodeID != "" || filter.LastNodeID != "") {
			*errs = append(*errs, prefix+" empty filters must not carry node ids")
		}
		if filter.Count > 0 && want.previewLimit > 0 && (filter.FirstNodeID == "" || filter.LastNodeID == "") {
			*errs = append(*errs, prefix+" populated status filters must carry first and last node ids")
		}
	}
}

func missionDashboardCompactFilterByKey(filters []AtlasMissionDashboardCompactFilter, key string) AtlasMissionDashboardCompactFilter {
	for _, filter := range filters {
		if filter.Key == key {
			return filter
		}
	}
	return AtlasMissionDashboardCompactFilter{}
}

func missionDashboardStatusFilterStatus(key string, count int) string {
	switch key {
	case "ready":
		if count == 0 {
			return "no_ready_nodes"
		}
		return "ready_work_remains"
	case "blocked":
		if count == 0 {
			return "no_blocked_nodes"
		}
		return "blocked_nodes_present"
	case "failed":
		if count == 0 {
			return "no_failed_nodes"
		}
		return "failed_nodes_present"
	default:
		return ""
	}
}

func missionDashboardCleanupStateStatusFromCounts(fixture AtlasMissionDashboardCompactFilters) string {
	if fixture.FinalResponseAllowed && fixture.ReadyNodes == 0 && fixture.BlockedNodes == 0 && fixture.FailedNodes == 0 {
		return "cleanup_state_complete"
	}
	if fixture.BlockedNodes > 0 || fixture.FailedNodes > 0 {
		return "cleanup_state_blocked"
	}
	return "cleanup_state_pending_ready_work"
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
	return missionDashboardActiveFilterKeyFromCounts(readback.ReadyNodes, readback.BlockedNodes, readback.FailedNodes)
}

func missionDashboardActiveFilterKeyFromCounts(readyNodes, blockedNodes, failedNodes int) string {
	if readyNodes > 0 {
		return "ready"
	}
	if blockedNodes > 0 {
		return "blocked"
	}
	if failedNodes > 0 {
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
