package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasFoundryImportReadinessBinding(nodeID, sourceReadbackPath, sourceWorkgraphPath, foundryImportPath, foundryHandoffPath string) (AtlasFoundryImportReadinessBinding, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	sourceWorkgraphPath = strings.TrimSpace(sourceWorkgraphPath)
	foundryImportPath = strings.TrimSpace(foundryImportPath)
	foundryHandoffPath = strings.TrimSpace(foundryHandoffPath)
	for name, value := range map[string]string{
		"node id":               nodeID,
		"source readback path":  sourceReadbackPath,
		"source workgraph path": sourceWorkgraphPath,
		"foundry import path":   foundryImportPath,
		"foundry handoff path":  foundryHandoffPath,
	} {
		if value == "" {
			return AtlasFoundryImportReadinessBinding{}, fmt.Errorf("%s is required", name)
		}
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	workgraph, err := LoadJSON[Workgraph](sourceWorkgraphPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	foundryImport, err := LoadJSON[FoundryImport](foundryImportPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	if err := ValidateFoundryImport(foundryImport); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	if err := ValidateFoundryImportMatchesWorkgraph(workgraph, foundryImport); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	handoff, err := LoadJSON[FoundryContinuationHandoff](foundryHandoffPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	workgraphDigest, err := digestTextFileWithNormalizedLineEndings(sourceWorkgraphPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	importDigest, err := digestTextFileWithNormalizedLineEndings(foundryImportPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	handoffDigest, err := digestTextFileWithNormalizedLineEndings(foundryHandoffPath)
	if err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}

	next, ok := NextReadyNode(workgraph)
	if !ok {
		return AtlasFoundryImportReadinessBinding{}, fmt.Errorf("source workgraph has no next ready node")
	}
	activeNodeID := ""
	activeTaskID := ""
	if len(foundryImport.Tasks) == 1 {
		activeNodeID = foundryImport.Tasks[0].NodeID
		activeTaskID = foundryImport.Tasks[0].TaskID
	}
	dependenciesCompleted := true
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	for _, dep := range next.Dependencies {
		if statusByID[dep] != "completed" {
			dependenciesCompleted = false
			break
		}
	}
	binding := AtlasFoundryImportReadinessBinding{
		Schema:                      AtlasFoundryImportReadinessBindingContract,
		NodeID:                      nodeID,
		Status:                      "single_active_foundry_import_ready",
		SourceReadbackPath:          publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:        readbackDigest,
		SourceWorkgraphPath:         publicArtifactRef(sourceWorkgraphPath),
		SourceWorkgraphDigest:       workgraphDigest,
		FoundryImportPath:           publicArtifactRef(foundryImportPath),
		FoundryImportDigest:         importDigest,
		FoundryHandoffPath:          publicArtifactRef(foundryHandoffPath),
		FoundryHandoffDigest:        handoffDigest,
		FoundryTaskCount:            len(foundryImport.Tasks),
		ActiveNodeID:                activeNodeID,
		ActiveTaskID:                activeTaskID,
		WorkgraphNextReadyNode:      next.ID,
		ReadbackFirstExecutableNode: readback.FirstExecutableNode,
		HandoffFirstSafeNode:        handoff.FirstSafeNode,
		DependenciesCompleted:       dependenciesCompleted,
		MatchesWorkgraph:            activeNodeID == next.ID && activeTaskID == next.FactoryTask.ID,
		MatchesReadbackNextNode:     activeNodeID == readback.FirstExecutableNode,
		HandoffMatchesImport:        handoff.FirstSafeNode == activeNodeID && samePortablePathSuffix(foundryImportPath, handoff.FoundryImportPath),
		FinalResponseAllowed:        readback.FinalResponseAllowed,
		ExactNextAction:             readback.ExactNextAction,
		SchedulesWork:               false,
		ExecutesWork:                false,
		ApprovesWork:                false,
		ClaimsAuthorityAdvance:      false,
		RSIRemainsDenied:            readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !binding.DependenciesCompleted || !binding.MatchesWorkgraph || !binding.MatchesReadbackNextNode || !binding.HandoffMatchesImport || binding.FoundryTaskCount != 1 {
		binding.Status = "foundry_import_readiness_binding_failed"
	}
	if err := ValidateAtlasFoundryImportReadinessBinding(binding); err != nil {
		return AtlasFoundryImportReadinessBinding{}, err
	}
	return binding, nil
}

func filepathSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func samePortablePathSuffix(left, right string) bool {
	left = filepathSlash(left)
	right = filepathSlash(right)
	return strings.HasSuffix(left, right) || strings.HasSuffix(right, left)
}

func ValidateAtlasFoundryImportReadinessBinding(binding AtlasFoundryImportReadinessBinding) error {
	var errs []string
	requireContract(&errs, "foundry_import_readiness_binding", binding.Schema, AtlasFoundryImportReadinessBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if !oneOf(binding.Status, "single_active_foundry_import_ready", "foundry_import_readiness_binding_failed") {
		errs = append(errs, "status must be single_active_foundry_import_ready or foundry_import_readiness_binding_failed")
	}
	for field, value := range map[string]string{
		"source_readback_path":           binding.SourceReadbackPath,
		"source_workgraph_path":          binding.SourceWorkgraphPath,
		"foundry_import_path":            binding.FoundryImportPath,
		"foundry_handoff_path":           binding.FoundryHandoffPath,
		"active_node_id":                 binding.ActiveNodeID,
		"active_task_id":                 binding.ActiveTaskID,
		"workgraph_next_ready_node":      binding.WorkgraphNextReadyNode,
		"readback_first_executable_node": binding.ReadbackFirstExecutableNode,
		"handoff_first_safe_node":        binding.HandoffFirstSafeNode,
		"exact_next_action":              binding.ExactNextAction,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  binding.SourceReadbackDigest,
		"source_workgraph_digest": binding.SourceWorkgraphDigest,
		"foundry_import_digest":   binding.FoundryImportDigest,
		"foundry_handoff_digest":  binding.FoundryHandoffDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if binding.FoundryTaskCount != 1 {
		errs = append(errs, "foundry_task_count must be 1")
	}
	if binding.ActiveNodeID != binding.WorkgraphNextReadyNode {
		errs = append(errs, "active_node_id must match workgraph_next_ready_node")
	}
	if binding.ActiveNodeID != binding.ReadbackFirstExecutableNode {
		errs = append(errs, "active_node_id must match readback_first_executable_node")
	}
	if binding.ActiveNodeID != binding.HandoffFirstSafeNode {
		errs = append(errs, "active_node_id must match handoff_first_safe_node")
	}
	if !strings.HasSuffix(binding.ActiveTaskID, "-task") {
		errs = append(errs, "active_task_id must be a task id")
	}
	if !binding.DependenciesCompleted {
		errs = append(errs, "dependencies_completed must be true")
	}
	if !binding.MatchesWorkgraph {
		errs = append(errs, "matches_workgraph must be true")
	}
	if !binding.MatchesReadbackNextNode {
		errs = append(errs, "matches_readback_next_node must be true")
	}
	if !binding.HandoffMatchesImport {
		errs = append(errs, "handoff_matches_import must be true")
	}
	if binding.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	if !strings.Contains(binding.ExactNextAction, binding.ActiveNodeID) || !strings.Contains(binding.ExactNextAction, "exactly one active node") {
		errs = append(errs, "exact_next_action must name the active node and exactly one active node")
	}
	if binding.Status == "single_active_foundry_import_ready" && (!binding.DependenciesCompleted || !binding.MatchesWorkgraph || !binding.MatchesReadbackNextNode || !binding.HandoffMatchesImport || binding.FoundryTaskCount != 1) {
		errs = append(errs, "ready status requires one active Foundry task bound to workgraph, readback, and handoff")
	}
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasFoundryImportReadinessBinding(path string, binding AtlasFoundryImportReadinessBinding) error {
	return WriteJSON(path, binding)
}
