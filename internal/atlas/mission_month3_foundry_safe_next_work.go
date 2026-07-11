package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3FoundrySafeNextWorkFixture(nodeID, sourceReadbackPath, sourceWorkgraphPath, expectedSelectedNode, expectedNextNode string, readback AtlasRecommendationReadback, workgraph Workgraph) (AtlasMonth3FoundrySafeNextWorkFixture, error) {
	selectedNodeID := strings.TrimSpace(expectedSelectedNode)
	if selectedNodeID == "" {
		selectedNodeID = readback.FirstExecutableNode
	}
	var selected *WorkgraphNode
	for i := range workgraph.Nodes {
		if workgraph.Nodes[i].ID == selectedNodeID {
			selected = &workgraph.Nodes[i]
			break
		}
	}
	if selected == nil {
		return AtlasMonth3FoundrySafeNextWorkFixture{}, fmt.Errorf("selected node %q not found in workgraph", selectedNodeID)
	}
	fixture := AtlasMonth3FoundrySafeNextWorkFixture{
		Schema:                          AtlasMonth3FoundrySafeNextWorkFixtureContract,
		NodeID:                          strings.TrimSpace(nodeID),
		Status:                          "safe_next_work_selected",
		SourceReadbackPath:              publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:            digestValue(readback),
		SourceWorkgraphPath:             publicArtifactRef(sourceWorkgraphPath),
		SourceWorkgraphDigest:           digestValue(workgraph),
		CompletedNodesBefore:            readback.CompletedNodes,
		ReadyNodesBefore:                readback.ReadyNodes,
		BlockedNodesBefore:              readback.BlockedNodes,
		FailedNodesBefore:               readback.FailedNodes,
		SelectedNode:                    selected.ID,
		SelectedTaskID:                  selected.FactoryTask.ID,
		SelectedMutationClass:           selected.FactoryTask.MutationClass,
		SelectedAuthorityBoundary:       selected.FactoryTask.AuthorityBoundary,
		SelectedWriteScope:              selected.FactoryTask.WriteScope,
		RequiredGateCount:               len(selected.FactoryTask.RequiredGates),
		RequiredEvidenceCount:           len(selected.FactoryTask.RequiredEvidence),
		SingleActiveTask:                readback.FirstExecutableNode == selected.ID && readback.ExecutableReadyNodes == 1,
		TerminalReadinessBound:          readback.ReadyNodes > 0 && !readback.FinalResponseAllowed && strings.TrimSpace(readback.ExactNextAction) != "",
		FinalResponseAllowedBefore:      readback.FinalResponseAllowed,
		ExactNextActionBefore:           readback.ExactNextAction,
		ExpectedNextNodeAfterCompletion: strings.TrimSpace(expectedNextNode),
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasMonth3FoundrySafeNextWorkFixture(fixture); err != nil {
		return AtlasMonth3FoundrySafeNextWorkFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasMonth3FoundrySafeNextWorkFixture(fixture AtlasMonth3FoundrySafeNextWorkFixture) error {
	var errs []string
	requireContract(&errs, "month3_foundry_safe_next_work_fixture", fixture.Schema, AtlasMonth3FoundrySafeNextWorkFixtureContract)
	for field, value := range map[string]string{
		"node_id":                             fixture.NodeID,
		"status":                              fixture.Status,
		"source_readback_path":                fixture.SourceReadbackPath,
		"source_readback_digest":              fixture.SourceReadbackDigest,
		"source_workgraph_path":               fixture.SourceWorkgraphPath,
		"source_workgraph_digest":             fixture.SourceWorkgraphDigest,
		"selected_node":                       fixture.SelectedNode,
		"selected_task_id":                    fixture.SelectedTaskID,
		"selected_mutation_class":             fixture.SelectedMutationClass,
		"selected_authority_boundary":         fixture.SelectedAuthorityBoundary,
		"exact_next_action_before":            fixture.ExactNextActionBefore,
		"expected_next_node_after_completion": fixture.ExpectedNextNodeAfterCompletion,
	} {
		requireField(&errs, field, value)
	}
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	checkPublicPath(&errs, "source_workgraph_path", fixture.SourceWorkgraphPath, true)
	checkPublicPath(&errs, "selected_node", fixture.SelectedNode, true)
	checkPublicPath(&errs, "selected_task_id", fixture.SelectedTaskID, true)
	checkPublicPath(&errs, "expected_next_node_after_completion", fixture.ExpectedNextNodeAfterCompletion, true)
	checkOptionalDigest(&errs, "source_readback_digest", fixture.SourceReadbackDigest)
	checkOptionalDigest(&errs, "source_workgraph_digest", fixture.SourceWorkgraphDigest)
	if fixture.Status != "safe_next_work_selected" {
		errs = append(errs, "status must be safe_next_work_selected")
	}
	if fixture.CompletedNodesBefore <= 0 || fixture.ReadyNodesBefore <= 0 {
		errs = append(errs, "completed_nodes_before and ready_nodes_before must be positive")
	}
	if fixture.BlockedNodesBefore != 0 || fixture.FailedNodesBefore != 0 {
		errs = append(errs, "blocked_nodes_before and failed_nodes_before must be zero")
	}
	if fixture.SelectedMutationClass != "low_risk_code" {
		errs = append(errs, "selected_mutation_class must be low_risk_code")
	}
	if fixture.SelectedAuthorityBoundary != "atlas_recommendation_planning_only" {
		errs = append(errs, "selected_authority_boundary must be atlas_recommendation_planning_only")
	}
	if !containsStringValue(fixture.SelectedWriteScope, "docs/evidence") {
		errs = append(errs, "selected_write_scope must include docs/evidence")
	}
	if fixture.RequiredGateCount < 8 {
		errs = append(errs, "required_gate_count must cover all node evidence gates")
	}
	if fixture.RequiredEvidenceCount < 3 {
		errs = append(errs, "required_evidence_count must include source digest, recommendation, and task digest")
	}
	if !fixture.SingleActiveTask {
		errs = append(errs, "single_active_task must be true")
	}
	if !fixture.TerminalReadinessBound {
		errs = append(errs, "terminal_readiness_bound must be true")
	}
	if fixture.FinalResponseAllowedBefore {
		errs = append(errs, "final_response_allowed_before must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3FoundrySafeNextWorkFixture(path string, fixture AtlasMonth3FoundrySafeNextWorkFixture) error {
	return WriteJSON(path, fixture)
}
