package atlas

import (
	"fmt"
	"strings"
)

type AtlasFoundryHandoffReplayFixtureOptions struct {
	NodeID              string
	SourceReadbackPath  string
	SourceWorkgraphPath string
	FoundryImportPath   string
	FoundryHandoffPath  string
}

func BuildAtlasFoundryHandoffReplayFixture(options AtlasFoundryHandoffReplayFixtureOptions) (AtlasFoundryHandoffReplayFixture, error) {
	nodeID := strings.TrimSpace(options.NodeID)
	sourceReadbackPath := strings.TrimSpace(options.SourceReadbackPath)
	sourceWorkgraphPath := strings.TrimSpace(options.SourceWorkgraphPath)
	foundryImportPath := strings.TrimSpace(options.FoundryImportPath)
	foundryHandoffPath := strings.TrimSpace(options.FoundryHandoffPath)
	for name, value := range map[string]string{
		"node id":               nodeID,
		"source readback path":  sourceReadbackPath,
		"source workgraph path": sourceWorkgraphPath,
		"foundry import path":   foundryImportPath,
		"foundry handoff path":  foundryHandoffPath,
	} {
		if value == "" {
			return AtlasFoundryHandoffReplayFixture{}, fmt.Errorf("%s is required", name)
		}
	}

	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	workgraph, err := LoadJSON[Workgraph](sourceWorkgraphPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	foundryImport, err := LoadJSON[FoundryImport](foundryImportPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	if err := ValidateFoundryImport(foundryImport); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	if err := ValidateFoundryImportMatchesWorkgraph(workgraph, foundryImport); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	handoff, err := LoadJSON[FoundryContinuationHandoff](foundryHandoffPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	next, ok := NextReadyNode(workgraph)
	if !ok {
		return AtlasFoundryHandoffReplayFixture{}, fmt.Errorf("source workgraph has no next ready node")
	}

	activeNodeID := ""
	activeTaskID := ""
	mutationClass := ""
	if len(foundryImport.Tasks) == 1 {
		task := foundryImport.Tasks[0]
		activeNodeID = task.NodeID
		activeTaskID = task.TaskID
		mutationClass = task.MutationClass
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	workgraphDigest, err := digestTextFileWithNormalizedLineEndings(sourceWorkgraphPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	importDigest, err := digestTextFileWithNormalizedLineEndings(foundryImportPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	handoffDigest, err := digestTextFileWithNormalizedLineEndings(foundryHandoffPath)
	if err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}

	fixture := AtlasFoundryHandoffReplayFixture{
		Schema:                         AtlasFoundryHandoffReplayFixtureContract,
		NodeID:                         nodeID,
		Status:                         "replay_guarded",
		SourceReadbackPath:             publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:           readbackDigest,
		SourceWorkgraphPath:            publicArtifactRef(sourceWorkgraphPath),
		SourceWorkgraphDigest:          workgraphDigest,
		FoundryImportPath:              publicArtifactRef(foundryImportPath),
		FoundryImportDigest:            importDigest,
		FoundryHandoffPath:             publicArtifactRef(foundryHandoffPath),
		FoundryHandoffDigest:           handoffDigest,
		ResumedFirstExecutableNode:     readback.FirstExecutableNode,
		ResumedExactNextAction:         readback.ExactNextAction,
		CompletedNodesBefore:           readback.CompletedNodes,
		ReadyNodesBefore:               readback.ReadyNodes,
		FinalResponseAllowed:           readback.FinalResponseAllowed,
		ActiveNodeID:                   activeNodeID,
		ActiveTaskID:                   activeTaskID,
		FoundryTaskCount:               len(foundryImport.Tasks),
		HandoffFirstSafeNode:           handoff.FirstSafeNode,
		WorkgraphNextReadyNode:         next.ID,
		MutationClass:                  mutationClass,
		SingleActiveImportTask:         len(foundryImport.Tasks) == 1,
		HandoffMatchesResumedReadback:  handoff.FirstSafeNode == readback.FirstExecutableNode,
		ImportMatchesResumedReadback:   activeNodeID == readback.FirstExecutableNode,
		HandoffMatchesWorkgraph:        handoff.FirstSafeNode == next.ID && activeTaskID == next.FactoryTask.ID && samePortablePathSuffix(foundryImportPath, handoff.FoundryImportPath),
		BoundedMutationClass:           mutationClass == "low_risk_code",
		ExactNextActionNamesActiveNode: strings.Contains(readback.ExactNextAction, activeNodeID) && strings.Contains(readback.ExactNextAction, "exactly one active node"),
		PromptPreservesActiveNode:      strings.Contains(handoff.Prompt, "first safe node: "+activeNodeID) && strings.Contains(handoff.Prompt, "do not stop after one node"),
		ReplayAssertions: []string{
			"single_active_foundry_task",
			"handoff_matches_resumed_readback",
			"low_risk_code_boundary_preserved",
			"final_response_denied_while_ready",
			"atlas_no_execution_boundary_preserved",
			"rsi_denial_preserved",
		},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if err := ValidateAtlasFoundryHandoffReplayFixture(fixture); err != nil {
		return AtlasFoundryHandoffReplayFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasFoundryHandoffReplayFixture(fixture AtlasFoundryHandoffReplayFixture) error {
	var errs []string
	requireContract(&errs, "foundry_handoff_replay_fixture", fixture.Schema, AtlasFoundryHandoffReplayFixtureContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "replay_guarded" {
		errs = append(errs, "status must be replay_guarded")
	}
	for field, value := range map[string]string{
		"source_readback_path":          fixture.SourceReadbackPath,
		"source_workgraph_path":         fixture.SourceWorkgraphPath,
		"foundry_import_path":           fixture.FoundryImportPath,
		"foundry_handoff_path":          fixture.FoundryHandoffPath,
		"resumed_first_executable_node": fixture.ResumedFirstExecutableNode,
		"resumed_exact_next_action":     fixture.ResumedExactNextAction,
		"active_node_id":                fixture.ActiveNodeID,
		"active_task_id":                fixture.ActiveTaskID,
		"handoff_first_safe_node":       fixture.HandoffFirstSafeNode,
		"workgraph_next_ready_node":     fixture.WorkgraphNextReadyNode,
		"mutation_class":                fixture.MutationClass,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_readback_digest":  fixture.SourceReadbackDigest,
		"source_workgraph_digest": fixture.SourceWorkgraphDigest,
		"foundry_import_digest":   fixture.FoundryImportDigest,
		"foundry_handoff_digest":  fixture.FoundryHandoffDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if fixture.CompletedNodesBefore <= 0 || fixture.ReadyNodesBefore <= 0 {
		errs = append(errs, "completed_nodes_before and ready_nodes_before must be positive")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	if fixture.FoundryTaskCount != 1 {
		errs = append(errs, "foundry_task_count must be 1")
	}
	if !fixture.SingleActiveImportTask {
		errs = append(errs, "single_active_import_task must be true")
	}
	if fixture.ActiveNodeID != fixture.ResumedFirstExecutableNode {
		errs = append(errs, "active_node_id must match resumed_first_executable_node")
	}
	if fixture.HandoffFirstSafeNode != fixture.ResumedFirstExecutableNode {
		errs = append(errs, "handoff_first_safe_node must match resumed_first_executable_node")
	}
	if fixture.WorkgraphNextReadyNode != fixture.ResumedFirstExecutableNode {
		errs = append(errs, "workgraph_next_ready_node must match resumed_first_executable_node")
	}
	if !strings.HasSuffix(fixture.ActiveTaskID, "-task") {
		errs = append(errs, "active_task_id must be a task id")
	}
	if fixture.MutationClass != "low_risk_code" {
		errs = append(errs, "mutation_class must be low_risk_code")
	}
	if !fixture.HandoffMatchesResumedReadback {
		errs = append(errs, "handoff_matches_resumed_readback must be true")
	}
	if !fixture.ImportMatchesResumedReadback {
		errs = append(errs, "import_matches_resumed_readback must be true")
	}
	if !fixture.HandoffMatchesWorkgraph {
		errs = append(errs, "handoff_matches_workgraph must be true")
	}
	if !fixture.BoundedMutationClass {
		errs = append(errs, "bounded_mutation_class must be true")
	}
	if !fixture.ExactNextActionNamesActiveNode {
		errs = append(errs, "exact_next_action_names_active_node must be true")
	}
	if !fixture.PromptPreservesActiveNode {
		errs = append(errs, "prompt_preserves_active_node must be true")
	}
	for _, assertion := range []string{
		"single_active_foundry_task",
		"handoff_matches_resumed_readback",
		"low_risk_code_boundary_preserved",
		"final_response_denied_while_ready",
		"atlas_no_execution_boundary_preserved",
		"rsi_denial_preserved",
	} {
		if !containsStringValue(fixture.ReplayAssertions, assertion) {
			errs = append(errs, "replay_assertions missing "+assertion)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasFoundryHandoffReplayFixture(path string, fixture AtlasFoundryHandoffReplayFixture) error {
	return WriteJSON(path, fixture)
}
