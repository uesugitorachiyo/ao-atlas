package atlas

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLongRunHardeningWaveLeaseSeedAndNodeOneReadback(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	seed := mustLoadJSON[map[string]any](t, filepath.Join(root, "source-seed.json"))
	seededFrom, ok := seed["seeded_from"].(map[string]any)
	if !ok || seededFrom["mission_id"] != "ao-mission-doubled-wave-v01" || seededFrom["completed_nodes"] != float64(50) {
		t.Fatalf("hardening wave seed must bind to completed 50-node doubled wave: %#v", seed["seeded_from"])
	}
	target, ok := seed["target"].(map[string]any)
	if !ok ||
		target["min_nodes"] != float64(30) ||
		target["node_budget"] != float64(40) ||
		target["min_minutes"] != float64(120) ||
		target["max_minutes"] != float64(180) ||
		target["continue_if_fast_target"] != float64(40) {
		t.Fatalf("hardening wave seed lost 2-3 hour budget: %#v", seed["target"])
	}

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "ao-atlas-long-run-hardening-wave-v01" ||
		wave.TotalTasks != 40 ||
		wave.MinimumTasks != 30 ||
		wave.NodeBudget != 40 ||
		wave.EstimatedMinutes != 120 ||
		wave.Supervisor == nil ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 ||
		wave.FinalResponseAllowed {
		t.Fatalf("hardening wave lost long-run lease settings: %#v", wave)
	}

	lease := mustLoadJSON[AtlasRecommendationLeaseStart](t, filepath.Join(root, "lease-start.json"))
	if err := ValidateAtlasRecommendationLeaseStart(lease); err != nil {
		t.Fatal(err)
	}
	if lease.MinMinutes != 120 ||
		lease.MaxMinutes != 180 ||
		lease.ContinueIfFastTarget != 40 ||
		lease.FinalResponseAllowed ||
		lease.MutatesRepositories ||
		lease.CallsProviders ||
		lease.ClaimsAuthorityAdvance {
		t.Fatalf("hardening lease widened authority or lost budget: %#v", lease)
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(root, "recommendation-workgraph.json"))
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 40 ||
		len(state.ExecutableReadyNodeIDs) != 1 ||
		state.ExecutableReadyNodeIDs[0] != "mission-recommendation-hardening-01" {
		t.Fatalf("hardening workgraph must expose exactly one executable node: nodes=%d executable=%#v", len(workgraph.Nodes), state.ExecutableReadyNodeIDs)
	}

	foundryImport := mustLoadJSON[FoundryImport](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-01", "foundry-import.json"))
	if err := ValidateFoundryImport(foundryImport); err != nil {
		t.Fatal(err)
	}
	if len(foundryImport.Tasks) != 1 || foundryImport.Tasks[0].NodeID != "mission-recommendation-hardening-01" {
		t.Fatalf("node 1 Foundry import must contain exactly the active node: %#v", foundryImport.Tasks)
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-01", "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 1 ||
		readback.ReadyNodes != 39 ||
		readback.ExecutableReadyNodes != 1 ||
		readback.FirstExecutableNode != "mission-recommendation-hardening-02" ||
		readback.FinalResponseAllowed ||
		readback.LeaseHealthStatus != "minimum_unmet" ||
		readback.EarlyReturnRiskStatus != "blocked_final_response_ready_nodes_remain" ||
		!strings.Contains(readback.ExactNextAction, "mission-recommendation-hardening-02") {
		t.Fatalf("node 1 completion readback must continue to node 2 without final response: %#v", readback)
	}
}

func TestLongRunHardeningWaveUntilDoneContinuesAfterOneHandoff(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-01", "recommendation-readback-after.json"))
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-02", "until-done-one-handoff-fixture.json"))
	if fixture["schema"] != "ao.atlas.until-done-governed-handoff-fixture.v0.1" ||
		fixture["status"] != "continuation_required" ||
		fixture["mode"] != "continue_until_done" ||
		fixture["governed_handoffs_recorded"] != float64(1) ||
		fixture["completed_nodes"] != float64(nodeOneReadback.CompletedNodes) ||
		fixture["ready_nodes_after_handoff"] != float64(nodeOneReadback.ReadyNodes) ||
		fixture["first_executable_node"] != nodeOneReadback.FirstExecutableNode ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("until-done fixture must bind one governed handoff to continuation state: %#v", fixture)
	}
	exactNextAction, _ := fixture["exact_next_action"].(string)
	if exactNextAction != nodeOneReadback.ExactNextAction ||
		!strings.Contains(exactNextAction, "mission-recommendation-hardening-02") {
		t.Fatalf("until-done fixture must preserve node 2 exact next action: fixture=%q readback=%q", exactNextAction, nodeOneReadback.ExactNextAction)
	}
	returnGate, _ := fixture["return_gate_status"].(string)
	if returnGate != "blocked_ready_nodes_remain" || nodeOneReadback.FinalResponseAllowed {
		t.Fatalf("until-done fixture must block final response while ready nodes remain: fixture=%#v readback=%#v", fixture, nodeOneReadback)
	}
	if stopReason, _ := fixture["premature_stop_reason"].(string); !strings.Contains(stopReason, "one governed handoff") {
		t.Fatalf("until-done fixture must explain why one handoff is insufficient: %#v", fixture)
	}

	nodeTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-02", "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwoReadback); err != nil {
		t.Fatal(err)
	}
	if nodeTwoReadback.CompletedNodes != 2 ||
		nodeTwoReadback.ReadyNodes != 38 ||
		nodeTwoReadback.FirstExecutableNode != "mission-recommendation-hardening-03" ||
		nodeTwoReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwoReadback.ExactNextAction, "mission-recommendation-hardening-03") {
		t.Fatalf("node 2 readback must continue to node 3 without final response: %#v", nodeTwoReadback)
	}
}

func TestLongRunHardeningWaveCommandReadbackDeniesFinalWithExactNextAction(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-02", "recommendation-readback-after.json"))
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-03", "command-final-denial-fixture.json"))
	if fixture["schema"] != "ao.command.final-response-denial.v0.1" ||
		fixture["status"] != "continuation_required" ||
		fixture["source"] != "command_readback" ||
		fixture["completed_nodes"] != float64(nodeTwoReadback.CompletedNodes) ||
		fixture["ready_nodes"] != float64(nodeTwoReadback.ReadyNodes) ||
		fixture["first_executable_node"] != nodeTwoReadback.FirstExecutableNode ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("Command final denial fixture must mirror node 2 continuation readback: %#v", fixture)
	}
	exactNextAction, _ := fixture["exact_next_action"].(string)
	if exactNextAction != nodeTwoReadback.ExactNextAction ||
		!strings.Contains(exactNextAction, "mission-recommendation-hardening-03") {
		t.Fatalf("Command final denial fixture must preserve node 3 exact next action: fixture=%q readback=%q", exactNextAction, nodeTwoReadback.ExactNextAction)
	}
	if denialGate, _ := fixture["final_response_denial_gate"].(string); denialGate != "deny_ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("Command fixture must use exact final-response denial gate: %#v", fixture)
	}
	if reason, _ := fixture["command_denial_reason"].(string); !strings.Contains(reason, "exact next action") || !strings.Contains(reason, "ready nodes") {
		t.Fatalf("Command fixture must explain both exact next action and ready-node denial: %#v", fixture)
	}

	nodeThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-03", "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThreeReadback); err != nil {
		t.Fatal(err)
	}
	if nodeThreeReadback.CompletedNodes != 3 ||
		nodeThreeReadback.ReadyNodes != 37 ||
		nodeThreeReadback.FirstExecutableNode != "mission-recommendation-hardening-04" ||
		nodeThreeReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThreeReadback.ExactNextAction, "mission-recommendation-hardening-04") {
		t.Fatalf("node 3 readback must continue to node 4 without final response: %#v", nodeThreeReadback)
	}
}

func TestLongRunHardeningWaveResumeBundleRequiresFreshCheckpoint(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-03", "recommendation-readback-after.json"))
	nodeThreeCheckpoint := mustLoadJSON[AtlasRecommendationCheckpointReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-03", "checkpoint-readback-after.json"))
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-04", "resume-fresh-checkpoint-fixture.json"))
	if fixture["schema"] != "ao.atlas.resume-fresh-checkpoint-fixture.v0.1" ||
		fixture["status"] != "continuation_required" ||
		fixture["source_checkpoint_status"] != nodeThreeCheckpoint.Status ||
		fixture["checkpoint_freshness_status"] != nodeThreeCheckpoint.CheckpointFreshnessStatus ||
		fixture["completed_nodes"] != float64(nodeThreeReadback.CompletedNodes) ||
		fixture["ready_nodes"] != float64(nodeThreeReadback.ReadyNodes) ||
		fixture["first_executable_node"] != nodeThreeReadback.FirstExecutableNode ||
		fixture["resume_uses_latest_checkpoint"] != true ||
		fixture["requires_fresh_checkpoint_before_final_answer"] != true ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("resume freshness fixture must bind node 3 checkpoint to continuation state: %#v", fixture)
	}
	exactNextAction, _ := fixture["exact_next_action"].(string)
	if exactNextAction != nodeThreeReadback.ExactNextAction ||
		!strings.Contains(exactNextAction, "mission-recommendation-hardening-04") {
		t.Fatalf("resume freshness fixture must preserve node 4 exact next action: fixture=%q readback=%q", exactNextAction, nodeThreeReadback.ExactNextAction)
	}
	if policy, _ := fixture["checkpoint_policy"].(string); policy != "after_each_node_or_timed_interval" {
		t.Fatalf("resume freshness fixture must require the long-run checkpoint policy: %#v", fixture)
	}

	nodeFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-04", "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeFourReadback); err != nil {
		t.Fatal(err)
	}
	if nodeFourReadback.CompletedNodes != 4 ||
		nodeFourReadback.ReadyNodes != 36 ||
		nodeFourReadback.FirstExecutableNode != "mission-recommendation-hardening-05" ||
		nodeFourReadback.FinalResponseAllowed ||
		!strings.Contains(nodeFourReadback.ExactNextAction, "mission-recommendation-hardening-05") {
		t.Fatalf("node 4 readback must continue to node 5 without final response: %#v", nodeFourReadback)
	}
}

func TestLongRunHardeningWaveRouteReconciliationStaysFreshAcrossArtifacts(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-04", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-05")
	routeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "route-recommendation-readback.json"))
	command := mustLoadJSON[AtlasRecommendationCommandReadback](t, filepath.Join(nodeDir, "route-command-readback.json"))
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, filepath.Join(nodeDir, "route-promoter-readback.json"))
	foundry := mustLoadJSON[AtlasRecommendationFoundryRollup](t, filepath.Join(nodeDir, "route-foundry-rollup.json"))
	reconciliation := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, filepath.Join(nodeDir, "route-reconciliation-packet.json"))
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "route-reconciliation-fixture.json"))

	if err := ValidateAtlasRecommendationClosureArtifacts(routeReadback, command, promoter, foundry); err != nil {
		t.Fatalf("route closure artifacts should agree: %v", err)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(routeReadback, command, promoter, foundry, reconciliation); err != nil {
		t.Fatalf("route reconciliation packet should agree: %v", err)
	}
	if routeReadback.StaleRouteDecisionStatus != nodeFourReadback.StaleRouteDecisionStatus ||
		reconciliation.StaleRouteDecisionStatus != routeReadback.StaleRouteDecisionStatus ||
		fixture["stale_route_decision_status"] != routeReadback.StaleRouteDecisionStatus ||
		fixture["schema"] != "ao.atlas.route-reconciliation-fixture.v0.1" ||
		fixture["status"] != "reconciled" ||
		fixture["artifact_agreement"] != true ||
		fixture["continuation_reason_agreement"] != true ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("route reconciliation fixture must bind fresh route status across artifacts: fixture=%#v readback=%#v reconciliation=%#v", fixture, routeReadback, reconciliation)
	}
	if command.ExactNextAction != routeReadback.ExactNextAction ||
		foundry.ExactNextAction != routeReadback.ExactNextAction ||
		reconciliation.ExactNextAction != routeReadback.ExactNextAction ||
		!strings.Contains(routeReadback.ExactNextAction, "mission-recommendation-hardening-05") {
		t.Fatalf("route artifacts must preserve node 5 exact next action: command=%q foundry=%q reconciliation=%q readback=%q", command.ExactNextAction, foundry.ExactNextAction, reconciliation.ExactNextAction, routeReadback.ExactNextAction)
	}
	if command.ContinuationContractReason != routeReadback.ContinuationContract.Reason ||
		promoter.ContinuationContractReason != routeReadback.ContinuationContract.Reason ||
		foundry.ContinuationContractReason != routeReadback.ContinuationContract.Reason ||
		reconciliation.ContinuationContractReason != routeReadback.ContinuationContract.Reason {
		t.Fatalf("route artifacts must agree on continuation reason")
	}

	nodeFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeFiveReadback); err != nil {
		t.Fatal(err)
	}
	if nodeFiveReadback.CompletedNodes != 5 ||
		nodeFiveReadback.ReadyNodes != 35 ||
		nodeFiveReadback.FirstExecutableNode != "mission-recommendation-hardening-06" ||
		nodeFiveReadback.FinalResponseAllowed ||
		!strings.Contains(nodeFiveReadback.ExactNextAction, "mission-recommendation-hardening-06") {
		t.Fatalf("node 5 readback must continue to node 6 without final response: %#v", nodeFiveReadback)
	}
}

func TestLongRunHardeningWaveEventIndexBindsEvidenceSlots(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-05", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-06")
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "event-index-bindings-fixture.json"))
	if fixture["schema"] != "ao.atlas.event-index-bindings-fixture.v0.1" ||
		fixture["status"] != "indexed" ||
		fixture["completed_nodes"] != float64(nodeFiveReadback.CompletedNodes) ||
		fixture["ready_nodes"] != float64(nodeFiveReadback.ReadyNodes) ||
		fixture["first_executable_node"] != nodeFiveReadback.FirstExecutableNode ||
		fixture["exact_next_action"] != nodeFiveReadback.ExactNextAction ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("event index fixture must bind node 5 continuation readback: %#v", fixture)
	}
	events, ok := fixture["events"].([]any)
	if !ok {
		t.Fatalf("event index fixture missing events array: %#v", fixture)
	}
	seen := map[string]bool{}
	for _, raw := range events {
		event, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("event index entry is not an object: %#v", raw)
		}
		slot, _ := event["slot"].(string)
		path, _ := event["evidence_path"].(string)
		if slot == "" || path == "" || strings.HasPrefix(path, "/") {
			t.Fatalf("event index entry must have slot and relative evidence_path: %#v", event)
		}
		seen[slot] = true
	}
	for _, want := range []string{"route", "node", "pull_request", "ci", "rollup", "blocker", "next_action"} {
		if !seen[want] {
			t.Fatalf("event index fixture missing %s slot: %#v", want, fixture)
		}
	}

	nodeSixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeSixReadback); err != nil {
		t.Fatal(err)
	}
	if nodeSixReadback.CompletedNodes != 6 ||
		nodeSixReadback.ReadyNodes != 34 ||
		nodeSixReadback.FirstExecutableNode != "mission-recommendation-hardening-07" ||
		nodeSixReadback.FinalResponseAllowed ||
		!strings.Contains(nodeSixReadback.ExactNextAction, "mission-recommendation-hardening-07") {
		t.Fatalf("node 6 readback must continue to node 7 without final response: %#v", nodeSixReadback)
	}
}

func TestLongRunHardeningWaveFoundryImportKeepsOneActiveNode(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeSixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-06", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-07")
	foundryImport := mustLoadJSON[FoundryImport](t, filepath.Join(nodeDir, "foundry-import.json"))
	if err := ValidateFoundryImport(foundryImport); err != nil {
		t.Fatal(err)
	}
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "single-active-foundry-import-fixture.json"))
	if len(foundryImport.Tasks) != 1 ||
		foundryImport.Tasks[0].NodeID != "mission-recommendation-hardening-07" ||
		foundryImport.Tasks[0].Task.ID != "mission-recommendation-hardening-07-task" ||
		fixture["schema"] != "ao.atlas.single-active-foundry-import-fixture.v0.1" ||
		fixture["status"] != "single_active_node_confirmed" ||
		fixture["active_node"] != foundryImport.Tasks[0].NodeID ||
		fixture["active_task"] != foundryImport.Tasks[0].Task.ID ||
		fixture["foundry_task_count"] != float64(1) ||
		fixture["completed_nodes"] != float64(nodeSixReadback.CompletedNodes) ||
		fixture["ready_nodes"] != float64(nodeSixReadback.ReadyNodes) ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("Foundry import must bind exactly one active node: fixture=%#v import=%#v", fixture, foundryImport)
	}
	if exactNextAction, _ := fixture["exact_next_action"].(string); exactNextAction != nodeSixReadback.ExactNextAction ||
		!strings.Contains(exactNextAction, "mission-recommendation-hardening-07") {
		t.Fatalf("single-active fixture must preserve node 7 exact next action: %#v", fixture)
	}

	nodeSevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeSevenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeSevenReadback.CompletedNodes != 7 ||
		nodeSevenReadback.ReadyNodes != 33 ||
		nodeSevenReadback.FirstExecutableNode != "mission-recommendation-hardening-08" ||
		nodeSevenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeSevenReadback.ExactNextAction, "mission-recommendation-hardening-08") {
		t.Fatalf("node 7 readback must continue to node 8 without final response: %#v", nodeSevenReadback)
	}
}

func TestLongRunHardeningWaveFinalStateReconciliationBindsClosureArtifacts(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeSevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-07", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-08")
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "final-state-recommendation-readback.json"))
	command := mustLoadJSON[AtlasRecommendationCommandReadback](t, filepath.Join(nodeDir, "final-state-command-readback.json"))
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, filepath.Join(nodeDir, "final-state-promoter-readback.json"))
	foundry := mustLoadJSON[AtlasRecommendationFoundryRollup](t, filepath.Join(nodeDir, "final-state-foundry-rollup.json"))
	reconciliation := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, filepath.Join(nodeDir, "final-state-reconciliation-packet.json"))
	fixture := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "final-state-reconciliation-fixture.json"))

	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err != nil {
		t.Fatalf("final-state closure artifacts should agree: %v", err)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, reconciliation); err != nil {
		t.Fatalf("final-state reconciliation packet should agree: %v", err)
	}
	finalState := reconciliation.FinalStateReconciliation
	if finalState.ContractVersion != "ao.atlas.final-state-reconciliation.v0.1" ||
		finalState.Status != reconciliation.Status ||
		finalState.WorkgraphStatus != readback.Status ||
		finalState.FoundryRollupStatus != foundry.Status ||
		finalState.PromoterVerdictStatus != promoter.Status ||
		finalState.CommandReadbackStatus != command.Status ||
		finalState.ExactNextAction != readback.ExactNextAction ||
		finalState.ContinuationReason != readback.ContinuationContract.Reason ||
		!finalState.ContinuationAgreement ||
		finalState.SchedulesWork ||
		finalState.ExecutesWork ||
		finalState.ApprovesWork {
		t.Fatalf("embedded final-state reconciliation must bind workgraph, Foundry, Promoter, and Command state: %#v", finalState)
	}
	if fixture["schema"] != "ao.atlas.final-state-reconciliation-fixture.v0.1" ||
		fixture["status"] != "continuation_required" ||
		fixture["artifacts_agree"] != true ||
		fixture["continuation_reason_agreement"] != true ||
		fixture["completed_nodes"] != float64(nodeSevenReadback.CompletedNodes) ||
		fixture["ready_nodes"] != float64(nodeSevenReadback.ReadyNodes) ||
		fixture["exact_next_action"] != nodeSevenReadback.ExactNextAction ||
		fixture["final_response_allowed"] != false {
		t.Fatalf("final-state fixture must bind node 7 continuation state: %#v", fixture)
	}

	nodeEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeEightReadback); err != nil {
		t.Fatal(err)
	}
	if nodeEightReadback.CompletedNodes != 8 ||
		nodeEightReadback.ReadyNodes != 32 ||
		nodeEightReadback.FirstExecutableNode != "mission-recommendation-hardening-09" ||
		nodeEightReadback.FinalResponseAllowed ||
		!strings.Contains(nodeEightReadback.ExactNextAction, "mission-recommendation-hardening-09") {
		t.Fatalf("node 8 readback must continue to node 9 without final response: %#v", nodeEightReadback)
	}
}

func TestLongRunHardeningWaveCommandTimelineSummarizesDoubledWave(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-08", "recommendation-readback-after.json"))
	sourceSummary := mustLoadJSON[struct {
		Schema                               string `json:"schema"`
		MissionID                            string `json:"mission_id"`
		TargetNodes                          int    `json:"target_nodes"`
		CompletedNodesAfterNode50Merge       int    `json:"completed_nodes_after_node_50_merge"`
		ReadyNodesAfterNode50Merge           int    `json:"ready_nodes_after_node_50_merge"`
		BlockedNodesAfterNode50Merge         int    `json:"blocked_nodes_after_node_50_merge"`
		FinalResponseAllowedAfterNode50Merge bool   `json:"final_response_allowed_after_node_50_merge"`
		ExactNextActionAfterNode50Merge      string `json:"exact_next_action_after_node_50_merge"`
	}](t, filepath.Join(repoRoot(t), "docs", "evidence", "ao-mission-doubled-wave-v01", "final-summary.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-09")
	timeline := mustLoadJSON[struct {
		Schema                     string `json:"schema"`
		NodeID                     string `json:"node_id"`
		Status                     string `json:"status"`
		SourceMissionID            string `json:"source_mission_id"`
		SourceFinalSummary         string `json:"source_final_summary"`
		TargetNodes                int    `json:"target_nodes"`
		CompletedNodes             int    `json:"completed_nodes"`
		ReadyNodes                 int    `json:"ready_nodes"`
		BlockedNodes               int    `json:"blocked_nodes"`
		FinalResponseAllowed       bool   `json:"final_response_allowed"`
		ExactNextAction            string `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		NodeCoverage struct {
			FirstNode    int   `json:"first_node"`
			LastNode     int   `json:"last_node"`
			TotalCovered int   `json:"total_covered"`
			CoveredNodes []int `json:"covered_nodes"`
		} `json:"node_coverage"`
		TimelineSegments []struct {
			Range    string `json:"range"`
			Summary  string `json:"summary"`
			Evidence string `json:"evidence"`
		} `json:"timeline_segments"`
	}](t, filepath.Join(nodeDir, "doubled-wave-command-timeline.json"))

	if sourceSummary.Schema != "ao.atlas.doubled-wave-final-summary.v0.1" ||
		sourceSummary.MissionID != "ao-mission-doubled-wave-v01" ||
		sourceSummary.TargetNodes != 50 ||
		sourceSummary.CompletedNodesAfterNode50Merge != 50 ||
		sourceSummary.ReadyNodesAfterNode50Merge != 0 ||
		sourceSummary.BlockedNodesAfterNode50Merge != 0 ||
		!sourceSummary.FinalResponseAllowedAfterNode50Merge {
		t.Fatalf("source doubled-wave summary must describe a completed 50-node wave: %#v", sourceSummary)
	}
	if timeline.Schema != "ao.atlas.command-compact-timeline.v0.1" ||
		timeline.NodeID != "mission-recommendation-hardening-09" ||
		timeline.Status != "recorded" ||
		timeline.SourceMissionID != sourceSummary.MissionID ||
		timeline.SourceFinalSummary != "docs/evidence/ao-mission-doubled-wave-v01/final-summary.json" ||
		timeline.TargetNodes != sourceSummary.TargetNodes ||
		timeline.CompletedNodes != sourceSummary.CompletedNodesAfterNode50Merge ||
		timeline.ReadyNodes != sourceSummary.ReadyNodesAfterNode50Merge ||
		timeline.BlockedNodes != sourceSummary.BlockedNodesAfterNode50Merge ||
		timeline.FinalResponseAllowed != sourceSummary.FinalResponseAllowedAfterNode50Merge ||
		timeline.ExactNextAction != sourceSummary.ExactNextActionAfterNode50Merge {
		t.Fatalf("Command timeline must bind doubled-wave final summary: %#v", timeline)
	}
	if timeline.CurrentHardeningCheckpoint.CompletedNodes != nodeEightReadback.CompletedNodes ||
		timeline.CurrentHardeningCheckpoint.ReadyNodes != nodeEightReadback.ReadyNodes ||
		timeline.CurrentHardeningCheckpoint.FirstExecutableNode != nodeEightReadback.FirstExecutableNode ||
		timeline.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeEightReadback.FinalResponseAllowed ||
		timeline.CurrentHardeningCheckpoint.ExactNextAction != nodeEightReadback.ExactNextAction {
		t.Fatalf("Command timeline must bind current hardening checkpoint: %#v", timeline.CurrentHardeningCheckpoint)
	}
	if timeline.NodeCoverage.FirstNode != 1 ||
		timeline.NodeCoverage.LastNode != 50 ||
		timeline.NodeCoverage.TotalCovered != 50 ||
		len(timeline.NodeCoverage.CoveredNodes) != 50 ||
		timeline.NodeCoverage.CoveredNodes[0] != 1 ||
		timeline.NodeCoverage.CoveredNodes[len(timeline.NodeCoverage.CoveredNodes)-1] != 50 {
		t.Fatalf("Command timeline must explicitly cover nodes 1 through 50: %#v", timeline.NodeCoverage)
	}
	if len(timeline.TimelineSegments) != 5 {
		t.Fatalf("Command timeline should summarize five 10-node segments, got %#v", timeline.TimelineSegments)
	}
	for _, segment := range timeline.TimelineSegments {
		if segment.Range == "" || segment.Summary == "" || segment.Evidence == "" {
			t.Fatalf("timeline segment must include range, summary, and evidence: %#v", segment)
		}
	}

	nodeNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeNineReadback); err != nil {
		t.Fatal(err)
	}
	if nodeNineReadback.CompletedNodes != 9 ||
		nodeNineReadback.ReadyNodes != 31 ||
		nodeNineReadback.FirstExecutableNode != "mission-recommendation-hardening-10" ||
		nodeNineReadback.FinalResponseAllowed ||
		!strings.Contains(nodeNineReadback.ExactNextAction, "mission-recommendation-hardening-10") {
		t.Fatalf("node 9 readback must continue to node 10 without final response: %#v", nodeNineReadback)
	}
}

func TestLongRunHardeningWavePromoterNoPromotionSummaryDeniesAuthorityAdvance(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-09", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-10")
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, filepath.Join(nodeDir, "supervisor-promoter-readback.json"))
	summary := mustLoadJSON[struct {
		Schema                                        string `json:"schema"`
		NodeID                                        string `json:"node_id"`
		Status                                        string `json:"status"`
		MissionID                                     string `json:"mission_id"`
		CompletedNodesBeforeNode                      int    `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode                          int    `json:"ready_nodes_before_node"`
		PromotionClaimed                              bool   `json:"promotion_claimed"`
		RSIRemainsDenied                              bool   `json:"rsi_remains_denied"`
		NextDeniedClass                               string `json:"next_denied_class"`
		SupervisorHardeningWithoutCapabilityPromotion bool   `json:"supervisor_hardening_without_capability_promotion"`
		FinalResponseAllowed                          bool   `json:"final_response_allowed"`
		ContinuationContractReason                    string `json:"continuation_contract_reason"`
		ExactNextAction                               string `json:"exact_next_action"`
		NoPromotionReasonSummary                      string `json:"no_promotion_reason_summary"`
		SchedulesWork                                 bool   `json:"schedules_work"`
		ExecutesWork                                  bool   `json:"executes_work"`
		ApprovesWork                                  bool   `json:"approves_work"`
		ClaimsAuthorityAdvance                        bool   `json:"claims_authority_advance"`
	}](t, filepath.Join(nodeDir, "promoter-no-promotion-summary.json"))

	if err := ValidateAtlasRecommendationReadback(nodeNineReadback); err != nil {
		t.Fatal(err)
	}
	if promoter.Schema != "ao.atlas.recommendation-promoter-readback.v0.1" ||
		promoter.Status != "no_promotion" ||
		promoter.PromotionClaimed ||
		!promoter.RSIRemainsDenied ||
		promoter.NextDeniedClass != "RSI" ||
		promoter.FinalResponseAllowed ||
		promoter.ContinuationContractReason != nodeNineReadback.ContinuationContract.Reason ||
		promoter.ClaimsAuthorityAdvance {
		t.Fatalf("promoter readback must deny promotion and authority advance: %#v", promoter)
	}
	if summary.Schema != "ao.atlas.promoter-no-promotion-summary.v0.1" ||
		summary.NodeID != "mission-recommendation-hardening-10" ||
		summary.Status != "no_promotion" ||
		summary.MissionID != nodeNineReadback.MissionID ||
		summary.CompletedNodesBeforeNode != nodeNineReadback.CompletedNodes ||
		summary.ReadyNodesBeforeNode != nodeNineReadback.ReadyNodes ||
		summary.PromotionClaimed ||
		!summary.RSIRemainsDenied ||
		summary.NextDeniedClass != "RSI" ||
		!summary.SupervisorHardeningWithoutCapabilityPromotion ||
		summary.FinalResponseAllowed ||
		summary.ContinuationContractReason != nodeNineReadback.ContinuationContract.Reason ||
		summary.ExactNextAction != nodeNineReadback.ExactNextAction ||
		summary.SchedulesWork ||
		summary.ExecutesWork ||
		summary.ApprovesWork ||
		summary.ClaimsAuthorityAdvance {
		t.Fatalf("no-promotion summary must bind continuation state without authority advance: %#v", summary)
	}
	if !strings.Contains(summary.NoPromotionReasonSummary, "ready_nodes=31") ||
		!strings.Contains(summary.NoPromotionReasonSummary, "final_response_allowed=false") {
		t.Fatalf("no-promotion summary must include exact denial counts and final gate: %q", summary.NoPromotionReasonSummary)
	}

	nodeTenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeTenReadback.CompletedNodes != 10 ||
		nodeTenReadback.ReadyNodes != 30 ||
		nodeTenReadback.FirstExecutableNode != "mission-recommendation-hardening-11" ||
		nodeTenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTenReadback.ExactNextAction, "mission-recommendation-hardening-11") {
		t.Fatalf("node 10 readback must continue to node 11 without final response: %#v", nodeTenReadback)
	}
}

func TestLongRunHardeningWaveSentinelScanCoversGeneratedDocsAndReadbacks(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-10", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-11")
	scan := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		ScannedScope               []string `json:"scanned_scope"`
		EvidenceRoots              []string `json:"evidence_roots"`
		NegativeScanTermsRedacted  bool     `json:"negative_scan_terms_redacted"`
		UnsafeMatchCount           int      `json:"unsafe_match_count"`
		PublicDocsScanPassed       bool     `json:"public_docs_scan_passed"`
		GeneratedReadbacksPassed   bool     `json:"generated_readbacks_scan_passed"`
		RSIRemainsDenied           bool     `json:"rsi_remains_denied"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
	}](t, filepath.Join(nodeDir, "sentinel-wording-scan.json"))

	if scan.Schema != "ao.atlas.sentinel-wording-scan.v0.1" ||
		scan.NodeID != "mission-recommendation-hardening-11" ||
		scan.Status != "passed" ||
		!scan.NegativeScanTermsRedacted ||
		scan.UnsafeMatchCount != 0 ||
		!scan.PublicDocsScanPassed ||
		!scan.GeneratedReadbacksPassed ||
		!scan.RSIRemainsDenied ||
		scan.SchedulesWork ||
		scan.ExecutesWork ||
		scan.ApprovesWork ||
		scan.ClaimsAuthorityAdvance {
		t.Fatalf("Sentinel scan summary must pass without authority effects: %#v", scan)
	}
	scopeSeen := map[string]bool{}
	for _, scope := range scan.ScannedScope {
		scopeSeen[scope] = true
	}
	if !scopeSeen["generated_docs"] || !scopeSeen["generated_readbacks"] {
		t.Fatalf("Sentinel scan must cover generated docs and readbacks: %#v", scan.ScannedScope)
	}
	for _, root := range scan.EvidenceRoots {
		if root == "" || strings.HasPrefix(root, "/") {
			t.Fatalf("Sentinel scan evidence roots must be relative: %#v", scan.EvidenceRoots)
		}
	}
	if scan.CurrentHardeningCheckpoint.CompletedNodes != nodeTenReadback.CompletedNodes ||
		scan.CurrentHardeningCheckpoint.ReadyNodes != nodeTenReadback.ReadyNodes ||
		scan.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTenReadback.FirstExecutableNode ||
		scan.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTenReadback.FinalResponseAllowed ||
		scan.CurrentHardeningCheckpoint.ExactNextAction != nodeTenReadback.ExactNextAction {
		t.Fatalf("Sentinel scan must bind current hardening checkpoint: %#v", scan.CurrentHardeningCheckpoint)
	}

	nodeElevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeElevenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeElevenReadback.CompletedNodes != 11 ||
		nodeElevenReadback.ReadyNodes != 29 ||
		nodeElevenReadback.FirstExecutableNode != "mission-recommendation-hardening-12" ||
		nodeElevenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeElevenReadback.ExactNextAction, "mission-recommendation-hardening-12") {
		t.Fatalf("node 11 readback must continue to node 12 without final response: %#v", nodeElevenReadback)
	}
}

func TestLongRunHardeningWaveUnsafePromptBlocksForbiddenActionCategories(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeElevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-11", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-12")
	fixture := mustLoadJSON[struct {
		Schema                     string          `json:"schema"`
		NodeID                     string          `json:"node_id"`
		Status                     string          `json:"status"`
		PromptEncoding             string          `json:"prompt_encoding"`
		BlockedActionCategories    map[string]bool `json:"blocked_action_categories"`
		UnsafeLiteralStored        bool            `json:"unsafe_literal_stored"`
		FinalResponseAllowed       bool            `json:"final_response_allowed"`
		ExactNextAction            string          `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
	}](t, filepath.Join(nodeDir, "unsafe-prompt-blocks-fixture.json"))

	if fixture.Schema != "ao.atlas.unsafe-prompt-blocks-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-12" ||
		fixture.Status != "blocked" ||
		fixture.PromptEncoding != "category_only_no_unsafe_literal" ||
		fixture.UnsafeLiteralStored ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeElevenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance {
		t.Fatalf("unsafe prompt fixture must block without execution or authority effects: %#v", fixture)
	}
	for _, category := range []string{
		"provider_call",
		"token_or_secret_inspection",
		"main_branch_mutation",
		"release_deploy_publish_upload_tag",
		"auth_policy_config_widening",
		"hidden_instruction_mutation",
		"broad_rsi_claim",
	} {
		if !fixture.BlockedActionCategories[category] {
			t.Fatalf("unsafe prompt fixture must block %s: %#v", category, fixture.BlockedActionCategories)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeElevenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeElevenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeElevenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeElevenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeElevenReadback.ExactNextAction {
		t.Fatalf("unsafe prompt fixture must bind current hardening checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeTwelveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwelveReadback); err != nil {
		t.Fatal(err)
	}
	if nodeTwelveReadback.CompletedNodes != 12 ||
		nodeTwelveReadback.ReadyNodes != 28 ||
		nodeTwelveReadback.FirstExecutableNode != "mission-recommendation-hardening-13" ||
		nodeTwelveReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwelveReadback.ExactNextAction, "mission-recommendation-hardening-13") {
		t.Fatalf("node 12 readback must continue to node 13 without final response: %#v", nodeTwelveReadback)
	}
}

func TestLongRunHardeningWaveFoundryRollupNormalizesTerminalStatuses(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwelveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-12", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-13")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		SourceCompletedNodes int    `json:"source_completed_nodes"`
		SourceReadyNodes     int    `json:"source_ready_nodes"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		TerminalStatuses     []struct {
			Status                       string `json:"status"`
			NormalizedStatus             string `json:"normalized_status"`
			Terminal                     bool   `json:"terminal"`
			ClosesTask                   bool   `json:"closes_task"`
			ClosesMission                bool   `json:"closes_mission"`
			RequiresCommandAgreement     bool   `json:"requires_command_agreement"`
			ExactMissingEvidenceRequired bool   `json:"exact_missing_evidence_required"`
			BlockerDetailsRequired       bool   `json:"blocker_details_required"`
			SafeNextActionRequired       bool   `json:"safe_next_action_required"`
		} `json:"terminal_statuses"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "foundry-terminal-normalization-fixture.json"))

	if fixture.Schema != "ao.atlas.foundry-terminal-normalization-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-13" ||
		fixture.Status != "normalized" ||
		fixture.SourceCompletedNodes != nodeTwelveReadback.CompletedNodes ||
		fixture.SourceReadyNodes != nodeTwelveReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwelveReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("terminal normalization fixture must bind node 12 checkpoint without execution or authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwelveReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwelveReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwelveReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwelveReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwelveReadback.ExactNextAction {
		t.Fatalf("terminal normalization fixture must preserve the active checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}
	statuses := map[string]struct {
		NormalizedStatus             string
		ClosesMission                bool
		RequiresCommandAgreement     bool
		ExactMissingEvidenceRequired bool
		BlockerDetailsRequired       bool
		SafeNextActionRequired       bool
	}{}
	for _, terminal := range fixture.TerminalStatuses {
		if !terminal.Terminal || !terminal.ClosesTask {
			t.Fatalf("terminal rollup status must close the task as a terminal Foundry result: %#v", terminal)
		}
		statuses[terminal.Status] = struct {
			NormalizedStatus             string
			ClosesMission                bool
			RequiresCommandAgreement     bool
			ExactMissingEvidenceRequired bool
			BlockerDetailsRequired       bool
			SafeNextActionRequired       bool
		}{
			NormalizedStatus:             terminal.NormalizedStatus,
			ClosesMission:                terminal.ClosesMission,
			RequiresCommandAgreement:     terminal.RequiresCommandAgreement,
			ExactMissingEvidenceRequired: terminal.ExactMissingEvidenceRequired,
			BlockerDetailsRequired:       terminal.BlockerDetailsRequired,
			SafeNextActionRequired:       terminal.SafeNextActionRequired,
		}
	}
	for _, status := range []string{"completed", "promoted", "denied", "blocked"} {
		if _, ok := statuses[status]; !ok {
			t.Fatalf("terminal normalization fixture missing %s status: %#v", status, statuses)
		}
	}
	if statuses["completed"].NormalizedStatus != "completed" ||
		statuses["completed"].ClosesMission ||
		statuses["promoted"].NormalizedStatus != "completed" ||
		statuses["promoted"].ClosesMission ||
		!statuses["promoted"].RequiresCommandAgreement ||
		statuses["denied"].NormalizedStatus != "denied" ||
		statuses["denied"].ClosesMission ||
		!statuses["denied"].ExactMissingEvidenceRequired ||
		!statuses["denied"].SafeNextActionRequired ||
		statuses["blocked"].NormalizedStatus != "blocked" ||
		statuses["blocked"].ClosesMission ||
		!statuses["blocked"].BlockerDetailsRequired ||
		!statuses["blocked"].SafeNextActionRequired {
		t.Fatalf("terminal statuses must normalize promoted/denied/blocked with exact closure requirements: %#v", statuses)
	}

	nodeThirteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirteenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeThirteenReadback.CompletedNodes != 13 ||
		nodeThirteenReadback.ReadyNodes != 27 ||
		nodeThirteenReadback.FirstExecutableNode != "mission-recommendation-hardening-14" ||
		nodeThirteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirteenReadback.ExactNextAction, "mission-recommendation-hardening-14") {
		t.Fatalf("node 13 readback must continue to node 14 without final response: %#v", nodeThirteenReadback)
	}
}

func TestLongRunHardeningWavePromotedFoundryRollupRequiresCommandAgreement(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-13", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-14")
	fixture := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		SourceStatus               string   `json:"source_status"`
		NormalizedStatus           string   `json:"normalized_status"`
		ClosesTask                 bool     `json:"closes_task"`
		ClosesMission              bool     `json:"closes_mission"`
		CommandAgreementRequired   bool     `json:"command_agreement_required"`
		CommandAgreementStatus     string   `json:"command_agreement_status"`
		DisagreementBlocksClosure  bool     `json:"disagreement_blocks_closure"`
		ClosureConditions          []string `json:"closure_conditions"`
		CompletedNodesBeforeNode   int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool     `json:"final_response_allowed"`
		ExactNextAction            string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "foundry-promoted-command-agreement-fixture.json"))

	if fixture.Schema != "ao.atlas.foundry-promoted-command-agreement-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-14" ||
		fixture.Status != "command_agreement_required" ||
		fixture.SourceStatus != "promoted" ||
		fixture.NormalizedStatus != "completed" ||
		!fixture.ClosesTask ||
		fixture.ClosesMission ||
		!fixture.CommandAgreementRequired ||
		fixture.CommandAgreementStatus != "required_before_mission_closure" ||
		!fixture.DisagreementBlocksClosure ||
		fixture.CompletedNodesBeforeNode != nodeThirteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeThirteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("promoted rollup fixture must require Command agreement before closure: %#v", fixture)
	}
	requiredConditions := map[string]bool{}
	for _, condition := range fixture.ClosureConditions {
		requiredConditions[condition] = true
	}
	for _, condition := range []string{
		"command_readback_agrees",
		"zero_ready_nodes",
		"all_required_closure_evidence_exists",
		"no_forbidden_surface",
		"rsi_remains_denied",
	} {
		if !requiredConditions[condition] {
			t.Fatalf("promoted rollup fixture missing closure condition %s: %#v", condition, fixture.ClosureConditions)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeThirteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeThirteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeThirteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeThirteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeThirteenReadback.ExactNextAction {
		t.Fatalf("promoted rollup fixture must bind node 13 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeFourteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeFourteenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeFourteenReadback.CompletedNodes != 14 ||
		nodeFourteenReadback.ReadyNodes != 26 ||
		nodeFourteenReadback.FirstExecutableNode != "mission-recommendation-hardening-15" ||
		nodeFourteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeFourteenReadback.ExactNextAction, "mission-recommendation-hardening-15") {
		t.Fatalf("node 14 readback must continue to node 15 without final response: %#v", nodeFourteenReadback)
	}
}

func TestLongRunHardeningWaveDeniedFoundryRollupReportsExactMissingEvidence(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeFourteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-14", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-15")
	fixture := mustLoadJSON[struct {
		Schema                       string   `json:"schema"`
		NodeID                       string   `json:"node_id"`
		Status                       string   `json:"status"`
		SourceStatus                 string   `json:"source_status"`
		NormalizedStatus             string   `json:"normalized_status"`
		ClosesTask                   bool     `json:"closes_task"`
		ClosesMission                bool     `json:"closes_mission"`
		ExactMissingEvidenceRequired bool     `json:"exact_missing_evidence_required"`
		GenericDenialAllowed         bool     `json:"generic_denial_allowed"`
		SafeNextActionRequired       bool     `json:"safe_next_action_required"`
		MissingEvidenceCategories    []string `json:"missing_evidence_categories"`
		DeniedReasonReadback         string   `json:"denied_reason_readback"`
		CompletedNodesBeforeNode     int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode         int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed         bool     `json:"final_response_allowed"`
		ExactNextAction              string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint   struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "foundry-denied-exact-evidence-fixture.json"))

	if fixture.Schema != "ao.atlas.foundry-denied-exact-evidence-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-15" ||
		fixture.Status != "exact_missing_evidence_required" ||
		fixture.SourceStatus != "denied" ||
		fixture.NormalizedStatus != "denied" ||
		!fixture.ClosesTask ||
		fixture.ClosesMission ||
		!fixture.ExactMissingEvidenceRequired ||
		fixture.GenericDenialAllowed ||
		!fixture.SafeNextActionRequired ||
		fixture.DeniedReasonReadback != "required_exact_missing_evidence_not_generic" ||
		fixture.CompletedNodesBeforeNode != nodeFourteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeFourteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeFourteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("denied rollup fixture must require exact missing evidence without authority effects: %#v", fixture)
	}
	categories := map[string]bool{}
	for _, category := range fixture.MissingEvidenceCategories {
		categories[category] = true
	}
	for _, category := range []string{
		"node_evidence",
		"stop_gate",
		"ci_pr_merge",
		"command_readback_agreement",
	} {
		if !categories[category] {
			t.Fatalf("denied rollup fixture missing evidence category %s: %#v", category, fixture.MissingEvidenceCategories)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeFourteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeFourteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeFourteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeFourteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeFourteenReadback.ExactNextAction {
		t.Fatalf("denied rollup fixture must bind node 14 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeFifteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeFifteenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeFifteenReadback.CompletedNodes != 15 ||
		nodeFifteenReadback.ReadyNodes != 25 ||
		nodeFifteenReadback.FirstExecutableNode != "mission-recommendation-hardening-16" ||
		nodeFifteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeFifteenReadback.ExactNextAction, "mission-recommendation-hardening-16") {
		t.Fatalf("node 15 readback must continue to node 16 without final response: %#v", nodeFifteenReadback)
	}
}

func TestLongRunHardeningWaveBlockedFoundryRollupPreservesBlockerDetails(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeFifteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-15", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-16")
	fixture := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		SourceStatus               string   `json:"source_status"`
		NormalizedStatus           string   `json:"normalized_status"`
		Terminal                   bool     `json:"terminal"`
		ClosesTask                 bool     `json:"closes_task"`
		ClosesMission              bool     `json:"closes_mission"`
		BlockerDetailsRequired     bool     `json:"blocker_details_required"`
		GenericBlockerAllowed      bool     `json:"generic_blocker_allowed"`
		SafeNextActionRequired     bool     `json:"safe_next_action_required"`
		ResumeCheckpointRequired   bool     `json:"resume_checkpoint_required"`
		BlockerDetailCategories    []string `json:"blocker_detail_categories"`
		CompletedNodesBeforeNode   int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool     `json:"final_response_allowed"`
		ExactNextAction            string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "foundry-blocked-safe-next-action-fixture.json"))

	if fixture.Schema != "ao.atlas.foundry-blocked-safe-next-action-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-16" ||
		fixture.Status != "blocker_details_required" ||
		fixture.SourceStatus != "blocked" ||
		fixture.NormalizedStatus != "blocked" ||
		!fixture.Terminal ||
		!fixture.ClosesTask ||
		fixture.ClosesMission ||
		!fixture.BlockerDetailsRequired ||
		fixture.GenericBlockerAllowed ||
		!fixture.SafeNextActionRequired ||
		!fixture.ResumeCheckpointRequired ||
		fixture.CompletedNodesBeforeNode != nodeFifteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeFifteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeFifteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("blocked rollup fixture must preserve blocker details and safe next action: %#v", fixture)
	}
	categories := map[string]bool{}
	for _, category := range fixture.BlockerDetailCategories {
		categories[category] = true
	}
	for _, category := range []string{
		"blocked_node_id",
		"blocker_reason",
		"repair_attempts",
		"safe_next_action",
		"resume_checkpoint",
	} {
		if !categories[category] {
			t.Fatalf("blocked rollup fixture missing blocker detail category %s: %#v", category, fixture.BlockerDetailCategories)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeFifteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeFifteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeFifteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeFifteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeFifteenReadback.ExactNextAction {
		t.Fatalf("blocked rollup fixture must bind node 15 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeSixteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeSixteenReadback); err != nil {
		t.Fatal(err)
	}
	if nodeSixteenReadback.CompletedNodes != 16 ||
		nodeSixteenReadback.ReadyNodes != 24 ||
		nodeSixteenReadback.FirstExecutableNode != "mission-recommendation-hardening-17" ||
		nodeSixteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeSixteenReadback.ExactNextAction, "mission-recommendation-hardening-17") {
		t.Fatalf("node 16 readback must continue to node 17 without final response: %#v", nodeSixteenReadback)
	}
}

func TestLongRunHardeningWaveFeatureDepthDefaultsToTwentyTasks(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "feature-depth-default-20",
	})
	if err != nil {
		t.Fatal(err)
	}
	readback, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(readback.FeatureDepthRecommendations) < 20 {
		t.Fatalf("readback must carry at least 20 feature-depth recommendations by default, got %d: %#v", len(readback.FeatureDepthRecommendations), readback.FeatureDepthRecommendations)
	}

	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeSixteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-16", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-17")
	fixture := mustLoadJSON[struct {
		Schema                      string `json:"schema"`
		NodeID                      string `json:"node_id"`
		Status                      string `json:"status"`
		DefaultRecommendationFloor  int    `json:"default_recommendation_floor"`
		ObservedRecommendationCount int    `json:"observed_recommendation_count"`
		ActionableTaskFloorMet      bool   `json:"actionable_task_floor_met"`
		CompletedNodesBeforeNode    int    `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode        int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed        bool   `json:"final_response_allowed"`
		ExactNextAction             string `json:"exact_next_action"`
		CurrentHardeningCheckpoint  struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "feature-depth-default-20-fixture.json"))

	if fixture.Schema != "ao.atlas.feature-depth-default-20-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-17" ||
		fixture.Status != "floor_met" ||
		fixture.DefaultRecommendationFloor != 20 ||
		fixture.ObservedRecommendationCount < 20 ||
		!fixture.ActionableTaskFloorMet ||
		fixture.CompletedNodesBeforeNode != nodeSixteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeSixteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeSixteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("feature-depth default fixture must prove a 20-task floor without authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeSixteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeSixteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeSixteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeSixteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeSixteenReadback.ExactNextAction {
		t.Fatalf("feature-depth default fixture must bind node 16 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeSeventeenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeSeventeenReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeSeventeenReadback.FeatureDepthRecommendations) < 20 ||
		nodeSeventeenReadback.CompletedNodes != 17 ||
		nodeSeventeenReadback.ReadyNodes != 23 ||
		nodeSeventeenReadback.FirstExecutableNode != "mission-recommendation-hardening-18" ||
		nodeSeventeenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeSeventeenReadback.ExactNextAction, "mission-recommendation-hardening-18") {
		t.Fatalf("node 17 readback must carry 20 recommendations and continue to node 18: %#v", nodeSeventeenReadback)
	}
}

func TestLongRunHardeningWaveDoubledFeatureDepthReturnsFortyTasks(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "feature-depth-doubled-40",
	})
	if err != nil {
		t.Fatal(err)
	}
	readback, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(readback.FeatureDepthRecommendations) < 40 {
		t.Fatalf("doubled readback must carry all 40 feature-depth recommendations, got %d: %#v", len(readback.FeatureDepthRecommendations), readback.FeatureDepthRecommendations)
	}
	if !strings.Contains(readback.FeatureDepthRecommendations[39], "next-40") {
		t.Fatalf("doubled readback must include the 40th concrete task: %#v", readback.FeatureDepthRecommendations[39])
	}

	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeSeventeenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-17", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-18")
	fixture := mustLoadJSON[struct {
		Schema                      string `json:"schema"`
		NodeID                      string `json:"node_id"`
		Status                      string `json:"status"`
		DoubledRecommendationFloor  int    `json:"doubled_recommendation_floor"`
		ObservedRecommendationCount int    `json:"observed_recommendation_count"`
		IncludesFortiethTask        bool   `json:"includes_fortieth_task"`
		CompletedNodesBeforeNode    int    `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode        int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed        bool   `json:"final_response_allowed"`
		ExactNextAction             string `json:"exact_next_action"`
		CurrentHardeningCheckpoint  struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "feature-depth-doubled-40-fixture.json"))

	if fixture.Schema != "ao.atlas.feature-depth-doubled-40-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-18" ||
		fixture.Status != "floor_met" ||
		fixture.DoubledRecommendationFloor != 40 ||
		fixture.ObservedRecommendationCount < 40 ||
		!fixture.IncludesFortiethTask ||
		fixture.CompletedNodesBeforeNode != nodeSeventeenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeSeventeenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeSeventeenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("doubled feature-depth fixture must prove a 40-task floor without authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeSeventeenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeSeventeenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeSeventeenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeSeventeenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeSeventeenReadback.ExactNextAction {
		t.Fatalf("doubled feature-depth fixture must bind node 17 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeEighteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeEighteenReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeEighteenReadback.FeatureDepthRecommendations) < 40 ||
		nodeEighteenReadback.CompletedNodes != 18 ||
		nodeEighteenReadback.ReadyNodes != 22 ||
		nodeEighteenReadback.FirstExecutableNode != "mission-recommendation-hardening-19" ||
		nodeEighteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeEighteenReadback.ExactNextAction, "mission-recommendation-hardening-19") {
		t.Fatalf("node 18 readback must carry 40 recommendations and continue to node 19: %#v", nodeEighteenReadback)
	}
}

func TestLongRunHardeningWavePromptGeneratorCoversDurationStopGatesAndSafetyBoundaries(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "prompt-generator-coverage",
	})
	if err != nil {
		t.Fatal(err)
	}
	prompt := result.Prompt
	for _, want := range []string{
		"Target duration: 120 to 180 minutes.",
		"Node floor stop gate: complete at least 30 nodes before final response unless a true hard blocker remains.",
		"Lease floor stop gate: do not return before min_minutes=120 unless a true hard blocker remains.",
		"Continue-if-fast stop gate: if 30 nodes finish quickly and no blocker remains, continue through 40 nodes.",
		"Ready-work stop gate: if ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
		"Checkpoint stop gate: record a checkpoint after each node or timed interval before evaluating final response.",
		"No provider calls.",
		"No credential or token inspection.",
		"No direct main mutation.",
		"No release, deploy, publish, upload, or tag.",
		"No dependency updates unless separately authorized.",
		"No auth, policy, or config widening.",
		"No hidden instruction mutation.",
		"No broad RSI claim.",
		"RSI remains denied.",
		"Feature Depth Recommendations, at least 40 tasks",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("generated prompt missing required long-run coverage %q:\n%s", want, prompt)
		}
	}

	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeEighteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-18", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-19")
	fixture := mustLoadJSON[struct {
		Schema                     string          `json:"schema"`
		NodeID                     string          `json:"node_id"`
		Status                     string          `json:"status"`
		PromptSource               string          `json:"prompt_source"`
		TargetDurationMinutesMin   int             `json:"target_duration_minutes_min"`
		TargetDurationMinutesMax   int             `json:"target_duration_minutes_max"`
		MinimumNodes               int             `json:"minimum_nodes"`
		ContinueIfFastTarget       int             `json:"continue_if_fast_target"`
		FeatureDepthTaskFloor      int             `json:"feature_depth_task_floor"`
		PromptCoverage             map[string]bool `json:"prompt_coverage"`
		CompletedNodesBeforeNode   int             `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int             `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool            `json:"final_response_allowed"`
		ExactNextAction            string          `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "prompt-generator-coverage-fixture.json"))

	if fixture.Schema != "ao.atlas.prompt-generator-coverage-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-19" ||
		fixture.Status != "prompt_coverage_recorded" ||
		fixture.PromptSource != "buildAtlasRecommendationPrompt" ||
		fixture.TargetDurationMinutesMin != 120 ||
		fixture.TargetDurationMinutesMax != 180 ||
		fixture.MinimumNodes != 30 ||
		fixture.ContinueIfFastTarget != 40 ||
		fixture.FeatureDepthTaskFloor != 40 ||
		fixture.CompletedNodesBeforeNode != nodeEighteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeEighteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeEighteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("prompt generator coverage fixture must bind prompt floors and safety without authority effects: %#v", fixture)
	}
	for _, key := range []string{
		"target_duration",
		"node_floor_stop_gate",
		"lease_floor_stop_gate",
		"continue_if_fast_stop_gate",
		"ready_work_stop_gate",
		"checkpoint_stop_gate",
		"safety_boundaries",
		"rsi_denial",
		"feature_depth_40_floor",
	} {
		if !fixture.PromptCoverage[key] {
			t.Fatalf("prompt generator fixture missing coverage key %s: %#v", key, fixture.PromptCoverage)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeEighteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeEighteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeEighteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeEighteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeEighteenReadback.ExactNextAction {
		t.Fatalf("prompt generator fixture must bind node 18 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeNineteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeNineteenReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeNineteenReadback.FeatureDepthRecommendations) < 40 ||
		nodeNineteenReadback.CompletedNodes != 19 ||
		nodeNineteenReadback.ReadyNodes != 21 ||
		nodeNineteenReadback.FirstExecutableNode != "mission-recommendation-hardening-20" ||
		nodeNineteenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeNineteenReadback.ExactNextAction, "mission-recommendation-hardening-20") {
		t.Fatalf("node 19 readback must carry prompt coverage and continue to node 20: %#v", nodeNineteenReadback)
	}
}

func TestLongRunHardeningWaveCommandReadbackFinalGateRequiresZeroReadyAndLeaseMinimum(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "command-final-gate-coverage",
	})
	if err != nil {
		t.Fatal(err)
	}

	readyWorkgraph := completeRecommendationNodes(result.Workgraph, 30)
	readyReadback, err := BuildAtlasRecommendationReadback(result.Wave, readyWorkgraph, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:00-07:00",
		CompletedAt:     "2026-07-04T09:20:00-07:00",
		ElapsedMinutes:  120,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	readyCommand := BuildAtlasRecommendationCommandReadback(readyReadback)
	if readyCommand.FinalResponseAllowed ||
		readyCommand.ReadyNodes == 0 ||
		!readyCommand.MinMinutesMet ||
		readyCommand.NodeCompletionStatus != "nodes_in_progress" {
		t.Fatalf("Command must deny final response while ready nodes remain even after min_minutes: %#v", readyCommand)
	}
	for _, want := range []string{
		"ready_nodes=10",
		"min_minutes=120",
		"min_minutes_met=true",
		"node_completion_status=nodes_in_progress",
		"final_response_allowed=false",
	} {
		if !strings.Contains(readyCommand.CompactTimeline, want) {
			t.Fatalf("ready-node Command timeline missing %q: %s", want, readyCommand.CompactTimeline)
		}
	}

	shortCompletedWorkgraph := completeRecommendationNodes(result.Workgraph, 40)
	shortReadback, err := BuildAtlasRecommendationReadback(result.Wave, shortCompletedWorkgraph, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:20-07:00",
		CompletedAt:     "2026-07-04T07:42:06-07:00",
		ElapsedMinutes:  22,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	shortCommand := BuildAtlasRecommendationCommandReadback(shortReadback)
	if shortCommand.FinalResponseAllowed ||
		shortCommand.ReadyNodes != 0 ||
		shortCommand.MinMinutesMet ||
		shortCommand.NodeCompletionStatus != "all_nodes_complete" {
		t.Fatalf("Command must deny final response when all nodes complete but min_minutes is unmet: %#v", shortCommand)
	}
	for _, want := range []string{
		"ready_nodes=0",
		"min_minutes=120",
		"min_minutes_met=false",
		"node_completion_status=all_nodes_complete",
		"final_response_allowed=false",
	} {
		if !strings.Contains(shortCommand.CompactTimeline, want) {
			t.Fatalf("short-lease Command timeline missing %q: %s", want, shortCommand.CompactTimeline)
		}
	}

	completeReadback, err := BuildAtlasRecommendationReadback(result.Wave, shortCompletedWorkgraph, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:00-07:00",
		CompletedAt:     "2026-07-04T09:20:00-07:00",
		ElapsedMinutes:  120,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	completeCommand := BuildAtlasRecommendationCommandReadback(completeReadback)
	if !completeCommand.FinalResponseAllowed ||
		completeCommand.ReadyNodes != 0 ||
		!completeCommand.MinMinutesMet ||
		completeCommand.NodeCompletionStatus != "all_nodes_complete" {
		t.Fatalf("Command must allow final response only with zero ready nodes and min_minutes met: %#v", completeCommand)
	}
	for _, want := range []string{
		"ready_nodes=0",
		"min_minutes=120",
		"min_minutes_met=true",
		"node_completion_status=all_nodes_complete",
		"final_response_allowed=true",
	} {
		if !strings.Contains(completeCommand.CompactTimeline, want) {
			t.Fatalf("final-allowed Command timeline missing %q: %s", want, completeCommand.CompactTimeline)
		}
	}
	if err := ValidateAtlasRecommendationClosureArtifacts(
		completeReadback,
		completeCommand,
		BuildAtlasRecommendationPromoterReadback(completeReadback),
		BuildAtlasRecommendationFoundryRollup(completeReadback),
	); err != nil {
		t.Fatalf("Command final gate should agree with closure artifacts when zero ready nodes and min_minutes are met: %v", err)
	}

	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeNineteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-19", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-20")
	fixture := mustLoadJSON[struct {
		Schema                     string            `json:"schema"`
		NodeID                     string            `json:"node_id"`
		Status                     string            `json:"status"`
		CommandGateCases           map[string]string `json:"command_gate_cases"`
		AllowsFinalOnlyWhen        []string          `json:"allows_final_only_when"`
		CompactTimelineFields      []string          `json:"compact_timeline_fields"`
		CompletedNodesBeforeNode   int               `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int               `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool              `json:"final_response_allowed"`
		ExactNextAction            string            `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "command-final-gate-fixture.json"))

	if fixture.Schema != "ao.atlas.command-final-gate-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-20" ||
		fixture.Status != "command_final_gate_recorded" ||
		fixture.CommandGateCases["ready_nodes_remain_min_minutes_met"] != "denied" ||
		fixture.CommandGateCases["all_nodes_complete_min_minutes_unmet"] != "denied" ||
		fixture.CommandGateCases["zero_ready_nodes_min_minutes_met"] != "allowed" ||
		fixture.CompletedNodesBeforeNode != nodeNineteenReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeNineteenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeNineteenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("Command final gate fixture must bind final response gating without authority effects: %#v", fixture)
	}
	for _, want := range []string{"zero_ready_nodes", "min_minutes_met", "all_nodes_complete"} {
		if !containsString(fixture.AllowsFinalOnlyWhen, want) {
			t.Fatalf("Command final gate fixture missing allow condition %s: %#v", want, fixture.AllowsFinalOnlyWhen)
		}
	}
	for _, want := range []string{"ready_nodes", "min_minutes", "min_minutes_met", "node_completion_status", "final_response_allowed"} {
		if !containsString(fixture.CompactTimelineFields, want) {
			t.Fatalf("Command final gate fixture missing compact timeline field %s: %#v", want, fixture.CompactTimelineFields)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeNineteenReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeNineteenReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeNineteenReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeNineteenReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeNineteenReadback.ExactNextAction {
		t.Fatalf("Command final gate fixture must bind node 19 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeTwentyReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyReadback.CompletedNodes != 20 ||
		nodeTwentyReadback.ReadyNodes != 20 ||
		nodeTwentyReadback.FirstExecutableNode != "mission-recommendation-hardening-21" ||
		nodeTwentyReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyReadback.ExactNextAction, "mission-recommendation-hardening-21") {
		t.Fatalf("node 20 readback must carry Command final gate coverage and continue to node 21: %#v", nodeTwentyReadback)
	}
}

func TestRecommendationFinalResponseGateEvaluationCentralizesReadbackClosureFields(t *testing.T) {
	cases := []struct {
		name                 string
		finalAllowed         bool
		nodesComplete        bool
		leaseTiming          atlasRecommendationLeaseTiming
		ready                int
		blocked              int
		failed               int
		exactNextAction      string
		firstExecutableNode  string
		wantReturnGate       string
		wantDenialGate       string
		wantContractStatus   string
		wantRefusesFinal     bool
		wantExactActionState string
	}{
		{
			name:                 "ready nodes deny final response",
			ready:                1,
			exactNextAction:      "Emit Foundry import for mission-recommendation-next-01 and execute exactly one active node.",
			firstExecutableNode:  "mission-recommendation-next-01",
			wantReturnGate:       "blocked_ready_nodes_remain",
			wantDenialGate:       "deny_ready_nodes_or_exact_next_action_remain",
			wantContractStatus:   "continuation_required",
			wantRefusesFinal:     true,
			wantExactActionState: "continuation_required",
		},
		{
			name:                 "hard blocker preserves blocker gate",
			blocked:              1,
			exactNextAction:      "Resolve blocked or failed recommendation node with exact repair evidence.",
			wantReturnGate:       "blocked_hard_blocker",
			wantDenialGate:       "blocked_hard_blocker",
			wantContractStatus:   "continuation_required",
			wantRefusesFinal:     true,
			wantExactActionState: "continuation_required",
		},
		{
			name:          "lease timing missing denies final response",
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "lease_timing_missing",
			},
			exactNextAction:      "Record started_at, completed_at, and elapsed_minutes before evaluating final response for the long-run lease.",
			wantReturnGate:       "blocked_lease_timing_missing",
			wantDenialGate:       "deny_ready_nodes_or_exact_next_action_remain",
			wantContractStatus:   "continuation_required",
			wantRefusesFinal:     true,
			wantExactActionState: "continuation_required",
		},
		{
			name:          "minimum minutes unmet denies final response",
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "elapsed_recorded",
				MinMinutesMet:   false,
			},
			exactNextAction:      "Generate and execute the next useful Atlas recommendation wave until elapsed_minutes meets supervisor.min_minutes, or record a true hard blocker.",
			wantReturnGate:       "blocked_minimum_minutes_unmet",
			wantDenialGate:       "deny_ready_nodes_or_exact_next_action_remain",
			wantContractStatus:   "continuation_required",
			wantRefusesFinal:     true,
			wantExactActionState: "continuation_required",
		},
		{
			name:          "completed lease allows final response",
			finalAllowed:  true,
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "elapsed_recorded",
				MinMinutesMet:   true,
			},
			exactNextAction:      "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks.",
			wantReturnGate:       "final_response_allowed",
			wantDenialGate:       "allow_final_response",
			wantContractStatus:   "ready_for_final_response",
			wantRefusesFinal:     false,
			wantExactActionState: "ready_for_final_response",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gate := recommendationFinalResponseGateEvaluation(tc.finalAllowed, tc.nodesComplete, tc.leaseTiming, tc.ready, tc.blocked, tc.failed, tc.exactNextAction, tc.firstExecutableNode)
			if gate.ReturnGateStatus != tc.wantReturnGate ||
				gate.FinalResponseDenialGate != tc.wantDenialGate ||
				gate.ContinuationContract.Status != tc.wantContractStatus ||
				gate.ContinuationContract.RefusesFinalResponse != tc.wantRefusesFinal ||
				gate.ExactNextActionReadback.Status != tc.wantExactActionState ||
				gate.ExactNextActionReadback.ReturnGateStatus != tc.wantReturnGate ||
				gate.ExactNextActionReadback.FinalResponseAllowed != tc.finalAllowed {
				t.Fatalf("final response gate evaluation drifted: %#v", gate)
			}
		})
	}
}

func TestRecommendationReadbackTransitionPreservesExactNextActionDecisions(t *testing.T) {
	cases := []struct {
		name                string
		finalAllowed        bool
		nodesComplete       bool
		leaseTiming         atlasRecommendationLeaseTiming
		completed           int
		minimum             int
		ready               int
		blocked             int
		failed              int
		firstExecutableNode string
		wantReason          string
		wantActionContains  string
		wantLeaseHealth     string
		wantEarlyReturnRisk string
	}{
		{
			name:                "ready node preserves first executable exact action",
			completed:           20,
			minimum:             15,
			ready:               1,
			firstExecutableNode: "mission-recommendation-hardening-21",
			wantReason:          "ready nodes or exact next actions remain",
			wantActionContains:  "mission-recommendation-hardening-21",
			wantLeaseHealth:     "minimum_met_continue_if_fast",
			wantEarlyReturnRisk: "blocked_final_response_ready_nodes_remain",
		},
		{
			name:                "blocked node preserves repair action",
			blocked:             1,
			wantReason:          "true hard blocker requires exact repair evidence",
			wantActionContains:  "Resolve blocked or failed recommendation node",
			wantLeaseHealth:     "hard_blocker_requires_repair",
			wantEarlyReturnRisk: "hard_blocker_requires_exact_missing_evidence",
		},
		{
			name:          "lease timing missing preserves timing action",
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "lease_timing_missing",
			},
			wantReason:          "minimum lease timing evidence missing",
			wantActionContains:  "Record started_at, completed_at, and elapsed_minutes",
			wantLeaseHealth:     "minimum_minutes_timing_missing",
			wantEarlyReturnRisk: "blocked_final_response_minimum_timing_missing",
		},
		{
			name:          "minimum minutes unmet preserves next wave action",
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "elapsed_recorded",
				MinMinutesMet:   false,
			},
			wantReason:          "minimum lease minutes unmet",
			wantActionContains:  "Generate and execute the next useful Atlas recommendation wave",
			wantLeaseHealth:     "minimum_minutes_unmet_continue_next_wave",
			wantEarlyReturnRisk: "blocked_final_response_minimum_minutes_unmet",
		},
		{
			name:          "final allowed preserves closure action",
			finalAllowed:  true,
			nodesComplete: true,
			leaseTiming: atlasRecommendationLeaseTiming{
				LeaseTimeStatus: "elapsed_recorded",
				MinMinutesMet:   true,
			},
			wantReason:          "all generated nodes complete and no ready nodes remain",
			wantActionContains:  "Finalize AO Atlas long-run wave",
			wantLeaseHealth:     "all_generated_nodes_complete",
			wantEarlyReturnRisk: "cleared_no_ready_nodes_remain",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transition := recommendationReadbackTransition(tc.finalAllowed, tc.nodesComplete, tc.leaseTiming, tc.completed, tc.minimum, tc.ready, tc.blocked, tc.failed, tc.firstExecutableNode)
			if transition.FinalResponseReason != tc.wantReason ||
				!strings.Contains(transition.ExactNextAction, tc.wantActionContains) ||
				transition.LeaseHealthStatus != tc.wantLeaseHealth ||
				transition.EarlyReturnRiskStatus != tc.wantEarlyReturnRisk {
				t.Fatalf("recommendation transition drifted: %#v", transition)
			}
		})
	}
}

func TestRecommendationCompactReadbackStatusNormalizesNodeStates(t *testing.T) {
	cases := []struct {
		name                     string
		completed                int
		total                    int
		ready                    int
		blocked                  int
		failed                   int
		finalAllowed             bool
		wantReadbackStatus       string
		wantNodeCompletionStatus string
	}{
		{
			name:                     "ready wave",
			total:                    4,
			ready:                    4,
			wantReadbackStatus:       "ready",
			wantNodeCompletionStatus: "nodes_in_progress",
		},
		{
			name:                     "in progress wave",
			completed:                2,
			total:                    4,
			ready:                    2,
			wantReadbackStatus:       "in_progress",
			wantNodeCompletionStatus: "nodes_in_progress",
		},
		{
			name:                     "completed final allowed wave",
			completed:                4,
			total:                    4,
			finalAllowed:             true,
			wantReadbackStatus:       "completed",
			wantNodeCompletionStatus: "all_nodes_complete",
		},
		{
			name:                     "blocked wave",
			completed:                2,
			total:                    4,
			blocked:                  1,
			wantReadbackStatus:       "blocked",
			wantNodeCompletionStatus: "blocked_or_failed_nodes_present",
		},
		{
			name:                     "failed wave",
			completed:                2,
			total:                    4,
			failed:                   1,
			wantReadbackStatus:       "blocked",
			wantNodeCompletionStatus: "blocked_or_failed_nodes_present",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status := recommendationCompactReadbackStatus(tc.completed, tc.total, tc.ready, tc.blocked, tc.failed, tc.finalAllowed)
			if status.ReadbackStatus != tc.wantReadbackStatus || status.NodeCompletionStatus != tc.wantNodeCompletionStatus {
				t.Fatalf("compact readback status drifted: %#v", status)
			}
		})
	}
}

func TestRecommendationReturnGateDenialReasonsAreStructured(t *testing.T) {
	cases := []struct {
		name                    string
		returnGateStatus        string
		finalAllowed            bool
		wantCode                string
		wantFinalResponseGate   string
		wantSummaryContains     string
		wantAllowsFinalResponse bool
	}{
		{
			name:                    "final response allowed",
			returnGateStatus:        "final_response_allowed",
			finalAllowed:            true,
			wantCode:                "allow_final_response",
			wantFinalResponseGate:   "allow_final_response",
			wantSummaryContains:     "final response allowed",
			wantAllowsFinalResponse: true,
		},
		{
			name:                  "hard blocker",
			returnGateStatus:      "blocked_hard_blocker",
			wantCode:              "hard_blocker",
			wantFinalResponseGate: "blocked_hard_blocker",
			wantSummaryContains:   "hard blocker",
		},
		{
			name:                  "lease timing missing",
			returnGateStatus:      "blocked_lease_timing_missing",
			wantCode:              "lease_timing_missing",
			wantFinalResponseGate: "deny_ready_nodes_or_exact_next_action_remain",
			wantSummaryContains:   "lease timing",
		},
		{
			name:                  "minimum minutes unmet",
			returnGateStatus:      "blocked_minimum_minutes_unmet",
			wantCode:              "minimum_minutes_unmet",
			wantFinalResponseGate: "deny_ready_nodes_or_exact_next_action_remain",
			wantSummaryContains:   "minimum minutes",
		},
		{
			name:                  "ready nodes remain",
			returnGateStatus:      "blocked_ready_nodes_remain",
			wantCode:              "ready_nodes_remain",
			wantFinalResponseGate: "deny_ready_nodes_or_exact_next_action_remain",
			wantSummaryContains:   "ready nodes",
		},
		{
			name:                  "no executable ready node",
			returnGateStatus:      "blocked_no_executable_ready_node",
			wantCode:              "no_executable_ready_node",
			wantFinalResponseGate: "deny_ready_nodes_or_exact_next_action_remain",
			wantSummaryContains:   "executable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason := recommendationReturnGateDenialReason(tc.finalAllowed, tc.returnGateStatus)
			if reason.Code != tc.wantCode ||
				reason.FinalResponseDenialGate != tc.wantFinalResponseGate ||
				reason.AllowsFinalResponse != tc.wantAllowsFinalResponse ||
				!strings.Contains(reason.Summary, tc.wantSummaryContains) {
				t.Fatalf("structured return gate denial reason drifted: %#v", reason)
			}
		})
	}
}

func TestLongRunHardeningWaveProductionReadinessSummaryBindsVerificationCIMergeCleanupAndEvidenceRoots(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-20", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-21")
	fixture := mustLoadJSON[struct {
		Schema                        string   `json:"schema"`
		NodeID                        string   `json:"node_id"`
		Status                        string   `json:"status"`
		LocalVerificationCommands     []string `json:"local_verification_commands"`
		ProductionReadinessSummary    string   `json:"production_readiness_summary"`
		CIRequiredBeforeMerge         bool     `json:"ci_required_before_merge"`
		MergeRequiredBeforeCompletion bool     `json:"merge_required_before_completion"`
		RemoteBranchCleanupRequired   bool     `json:"remote_branch_cleanup_required"`
		LocalBranchCleanupRequired    bool     `json:"local_branch_cleanup_required"`
		EvidenceRoots                 []string `json:"evidence_roots"`
		PriorMergedPRs                []int    `json:"prior_merged_prs"`
		PriorMergedPRCIPassed         bool     `json:"prior_merged_pr_ci_passed"`
		PriorMergedPRBranchCleanup    bool     `json:"prior_merged_pr_branch_cleanup"`
		CompletedNodesBeforeNode      int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode          int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed          bool     `json:"final_response_allowed"`
		ExactNextAction               string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint    struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "production-readiness-summary-fixture.json"))

	if fixture.Schema != "ao.atlas.production-readiness-summary-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-21" ||
		fixture.Status != "production_readiness_summary_recorded" ||
		fixture.ProductionReadinessSummary != "status=ready; score=100/100" ||
		!fixture.CIRequiredBeforeMerge ||
		!fixture.MergeRequiredBeforeCompletion ||
		!fixture.RemoteBranchCleanupRequired ||
		!fixture.LocalBranchCleanupRequired ||
		!fixture.PriorMergedPRCIPassed ||
		!fixture.PriorMergedPRBranchCleanup ||
		fixture.CompletedNodesBeforeNode != nodeTwentyReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("production readiness summary fixture must bind verification, CI, merge cleanup, and evidence roots without authority effects: %#v", fixture)
	}
	for _, want := range []string{
		"go test ./... -count=1",
		"go vet ./...",
		"go build ./cmd/atlas",
		"scripts/production-readiness.sh",
		"scripts/atlas-foundry-roundtrip-smoke.sh",
		"git diff --check",
	} {
		if !containsString(fixture.LocalVerificationCommands, want) {
			t.Fatalf("production readiness fixture missing verification command %s: %#v", want, fixture.LocalVerificationCommands)
		}
	}
	for _, want := range []string{
		"docs/evidence/ao-atlas-long-run-hardening-wave-v01",
		"docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-21",
		"target/production-readiness/summary.json",
		"target/atlas-foundry-roundtrip/summary.json",
	} {
		if !containsString(fixture.EvidenceRoots, want) {
			t.Fatalf("production readiness fixture missing evidence root %s: %#v", want, fixture.EvidenceRoots)
		}
	}
	for _, want := range []int{276, 277, 278, 279, 280, 281, 282, 283} {
		found := false
		for _, pr := range fixture.PriorMergedPRs {
			if pr == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("production readiness fixture missing prior merged PR #%d: %#v", want, fixture.PriorMergedPRs)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyReadback.ExactNextAction {
		t.Fatalf("production readiness fixture must bind node 20 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeTwentyOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyOneReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyOneReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyOneReadback.CompletedNodes != 21 ||
		nodeTwentyOneReadback.ReadyNodes != 19 ||
		nodeTwentyOneReadback.FirstExecutableNode != "mission-recommendation-hardening-22" ||
		nodeTwentyOneReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyOneReadback.ExactNextAction, "mission-recommendation-hardening-22") {
		t.Fatalf("node 21 readback must carry production readiness summary and continue to node 22: %#v", nodeTwentyOneReadback)
	}
}

func TestLongRunHardeningWaveEvidenceDigestSummaryUsesRelativeRouteAndPromptArtifacts(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-21", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-22")
	fixture := mustLoadJSON[struct {
		Schema                     string            `json:"schema"`
		NodeID                     string            `json:"node_id"`
		Status                     string            `json:"status"`
		DigestAlgorithm            string            `json:"digest_algorithm"`
		ArtifactPaths              map[string]string `json:"artifact_paths"`
		ArtifactDigests            map[string]string `json:"artifact_digests"`
		NoAbsolutePaths            bool              `json:"no_absolute_paths"`
		CompletedNodesBeforeNode   int               `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int               `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool              `json:"final_response_allowed"`
		ExactNextAction            string            `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "evidence-digest-summary-fixture.json"))

	if fixture.Schema != "ao.atlas.evidence-digest-summary-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-22" ||
		fixture.Status != "digest_summary_recorded" ||
		fixture.DigestAlgorithm != "sha256_normalized_line_endings" ||
		!fixture.NoAbsolutePaths ||
		fixture.CompletedNodesBeforeNode != nodeTwentyOneReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyOneReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyOneReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("digest summary fixture must bind relative evidence paths without authority effects: %#v", fixture)
	}
	requiredRefs := []string{
		"route_recommendation_readback",
		"route_command_readback",
		"route_reconciliation_packet",
		"root_next_recommended_prompt",
		"node_22_foundry_continuation_prompt",
	}
	for _, ref := range requiredRefs {
		artifactPath := fixture.ArtifactPaths[ref]
		if artifactPath == "" {
			t.Fatalf("digest summary fixture missing path ref %s: %#v", ref, fixture.ArtifactPaths)
		}
		forbiddenPathMarkers := []string{
			string(filepath.Separator) + "Users" + string(filepath.Separator),
			string(filepath.Separator) + "home" + string(filepath.Separator),
			string(filepath.Separator) + "private" + string(filepath.Separator),
			"file" + "://",
		}
		hasForbiddenPathMarker := false
		for _, marker := range forbiddenPathMarkers {
			if strings.Contains(artifactPath, marker) {
				hasForbiddenPathMarker = true
				break
			}
		}
		if filepath.IsAbs(artifactPath) || hasForbiddenPathMarker {
			t.Fatalf("digest summary path must be relative and public-safe for %s: %s", ref, artifactPath)
		}
		digest := fixture.ArtifactDigests[ref]
		if !strings.HasPrefix(digest, "sha256:") {
			t.Fatalf("digest summary missing sha256 digest for %s: %#v", ref, fixture.ArtifactDigests)
		}
		actual, err := digestFileWithNormalizedLineEndings(filepath.Join(repoRoot(t), filepath.FromSlash(artifactPath)))
		if err != nil {
			t.Fatal(err)
		}
		if digest != actual {
			t.Fatalf("digest summary mismatch for %s: fixture=%s actual=%s", ref, digest, actual)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyOneReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyOneReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyOneReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyOneReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyOneReadback.ExactNextAction {
		t.Fatalf("digest summary fixture must bind node 21 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeTwentyTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyTwoReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyTwoReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyTwoReadback.CompletedNodes != 22 ||
		nodeTwentyTwoReadback.ReadyNodes != 18 ||
		nodeTwentyTwoReadback.FirstExecutableNode != "mission-recommendation-hardening-23" ||
		nodeTwentyTwoReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyTwoReadback.ExactNextAction, "mission-recommendation-hardening-23") {
		t.Fatalf("node 22 readback must carry evidence digest summary and continue to node 23: %#v", nodeTwentyTwoReadback)
	}
}

func TestLongRunHardeningWaveArtifactAgreementTiesPromptCommandAndWorkgraphStatus(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-22", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-23")
	fixture := mustLoadJSON[struct {
		Schema                     string `json:"schema"`
		NodeID                     string `json:"node_id"`
		Status                     string `json:"status"`
		GeneratedPromptPath        string `json:"generated_prompt_path"`
		CommandReadbackPath        string `json:"command_readback_path"`
		WorkgraphAfterPath         string `json:"workgraph_after_path"`
		SourceReadbackPath         string `json:"source_readback_path"`
		PromptFirstSafeNode        string `json:"prompt_first_safe_node"`
		PromptTotalNodes           int    `json:"prompt_total_nodes"`
		PromptCompletedNodes       int    `json:"prompt_completed_nodes"`
		PromptReadyNodes           int    `json:"prompt_ready_nodes"`
		CommandExactNextAction     string `json:"command_exact_next_action"`
		WorkgraphNodeStatus        string `json:"workgraph_node_status"`
		WorkgraphNextReadyNode     string `json:"workgraph_next_ready_node"`
		CompletedNodesBeforeNode   int    `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool   `json:"final_response_allowed"`
		ExactNextAction            string `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "artifact-agreement-fixture.json"))

	if fixture.Schema != "ao.atlas.artifact-agreement-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-23" ||
		fixture.Status != "artifact_agreement_recorded" ||
		fixture.PromptFirstSafeNode != nodeTwentyTwoReadback.FirstExecutableNode ||
		fixture.PromptTotalNodes != nodeTwentyTwoReadback.TotalNodes ||
		fixture.PromptCompletedNodes != nodeTwentyTwoReadback.CompletedNodes ||
		fixture.PromptReadyNodes != nodeTwentyTwoReadback.ReadyNodes ||
		fixture.CommandExactNextAction != nodeTwentyTwoReadback.ExactNextAction ||
		fixture.CompletedNodesBeforeNode != nodeTwentyTwoReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyTwoReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyTwoReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("artifact agreement fixture must bind node 22 prompt and command state without authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyTwoReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyTwoReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyTwoReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyTwoReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyTwoReadback.ExactNextAction {
		t.Fatalf("artifact agreement fixture must bind node 22 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}
	for _, artifactPath := range []string{
		fixture.GeneratedPromptPath,
		fixture.CommandReadbackPath,
		fixture.WorkgraphAfterPath,
		fixture.SourceReadbackPath,
	} {
		if artifactPath == "" || filepath.IsAbs(artifactPath) {
			t.Fatalf("artifact agreement paths must be non-empty and relative: %#v", fixture)
		}
		forbiddenPathMarkers := []string{
			string(filepath.Separator) + "Users" + string(filepath.Separator),
			string(filepath.Separator) + "home" + string(filepath.Separator),
			string(filepath.Separator) + "private" + string(filepath.Separator),
			"file" + "://",
		}
		for _, marker := range forbiddenPathMarkers {
			if strings.Contains(artifactPath, marker) {
				t.Fatalf("artifact agreement path must be public-safe: %s", artifactPath)
			}
		}
		if _, err := os.Stat(filepath.Join(repoRoot(t), filepath.FromSlash(artifactPath))); err != nil {
			t.Fatal(err)
		}
	}

	promptBytes, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(fixture.GeneratedPromptPath)))
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	wantPromptSnippets := []string{
		"first safe node: " + nodeTwentyTwoReadback.FirstExecutableNode,
		"total nodes: " + strconv.Itoa(nodeTwentyTwoReadback.TotalNodes),
		"completed nodes: " + strconv.Itoa(nodeTwentyTwoReadback.CompletedNodes),
		"ready nodes: " + strconv.Itoa(nodeTwentyTwoReadback.ReadyNodes),
	}
	for _, want := range wantPromptSnippets {
		if !strings.Contains(prompt, want) {
			t.Fatalf("generated prompt missing agreement snippet %q:\n%s", want, prompt)
		}
	}

	command := mustLoadJSON[struct {
		Status               string `json:"status"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
	}](t, filepath.Join(repoRoot(t), filepath.FromSlash(fixture.CommandReadbackPath)))
	if command.Status != "artifact_agreement_recorded" ||
		command.CompletedNodesBefore != nodeTwentyTwoReadback.CompletedNodes ||
		command.ReadyNodesBefore != nodeTwentyTwoReadback.ReadyNodes ||
		command.FinalResponseAllowed ||
		command.ExactNextAction != nodeTwentyTwoReadback.ExactNextAction {
		t.Fatalf("artifact agreement command readback disagrees with node 22 readback: %#v", command)
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(repoRoot(t), filepath.FromSlash(fixture.WorkgraphAfterPath)))
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	nodeStatus := map[string]string{}
	for _, node := range workgraph.Nodes {
		nodeStatus[node.ID] = node.Status
	}
	if fixture.WorkgraphNodeStatus != "completed" ||
		nodeStatus["mission-recommendation-hardening-23"] != fixture.WorkgraphNodeStatus ||
		fixture.WorkgraphNextReadyNode != "mission-recommendation-hardening-24" ||
		nodeStatus[fixture.WorkgraphNextReadyNode] != "ready" {
		t.Fatalf("artifact agreement workgraph status mismatch: fixture=%#v status=%#v", fixture, nodeStatus)
	}

	nodeTwentyThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyThreeReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyThreeReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyThreeReadback.CompletedNodes != 23 ||
		nodeTwentyThreeReadback.ReadyNodes != 17 ||
		nodeTwentyThreeReadback.FirstExecutableNode != "mission-recommendation-hardening-24" ||
		nodeTwentyThreeReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyThreeReadback.ExactNextAction, "mission-recommendation-hardening-24") {
		t.Fatalf("node 23 readback must carry artifact agreement and continue to node 24: %#v", nodeTwentyThreeReadback)
	}
}

func TestLongRunHardeningWaveRollbackBoundaryForPromptOnlyNodes(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-23", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-24")
	fixture := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		PromptOnlyNode             bool     `json:"prompt_only_node"`
		NoDataLossBoundary         bool     `json:"no_data_loss_boundary"`
		DestructiveRollbackAllowed bool     `json:"destructive_rollback_allowed"`
		ReleaseActionRequired      bool     `json:"release_action_required"`
		RollbackCommand            string   `json:"rollback_command"`
		RollbackScope              []string `json:"rollback_scope"`
		RestoresPreviousCheckpoint string   `json:"restores_previous_checkpoint"`
		CompletedNodesBeforeNode   int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool     `json:"final_response_allowed"`
		ExactNextAction            string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "rollback-boundary-fixture.json"))
	rollback := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		RollbackScope              []string `json:"rollback_scope"`
		RollbackCommand            string   `json:"rollback_command"`
		RestoresPreviousCheckpoint string   `json:"restores_previous_checkpoint"`
		RequiresReleaseAction      bool     `json:"requires_release_action"`
	}](t, filepath.Join(nodeDir, "rollback_record.json"))

	if fixture.Schema != "ao.atlas.rollback-boundary-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-24" ||
		fixture.Status != "rollback_boundary_recorded" ||
		!fixture.PromptOnlyNode ||
		!fixture.NoDataLossBoundary ||
		fixture.DestructiveRollbackAllowed ||
		fixture.ReleaseActionRequired ||
		fixture.CompletedNodesBeforeNode != nodeTwentyThreeReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyThreeReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyThreeReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("rollback boundary fixture must bind prompt-only rollback without authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyThreeReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyThreeReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyThreeReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyThreeReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyThreeReadback.ExactNextAction {
		t.Fatalf("rollback boundary fixture must bind node 23 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}
	if rollback.Schema != "ao.atlas.rollback-record.v0.1" ||
		rollback.NodeID != fixture.NodeID ||
		rollback.Status != "available" ||
		rollback.RollbackCommand != fixture.RollbackCommand ||
		rollback.RestoresPreviousCheckpoint != fixture.RestoresPreviousCheckpoint ||
		rollback.RequiresReleaseAction ||
		len(rollback.RollbackScope) != len(fixture.RollbackScope) {
		t.Fatalf("rollback record must agree with rollback boundary fixture: rollback=%#v fixture=%#v", rollback, fixture)
	}
	for i, scope := range rollback.RollbackScope {
		if scope != fixture.RollbackScope[i] {
			t.Fatalf("rollback scope %d disagrees: rollback=%q fixture=%q", i, scope, fixture.RollbackScope[i])
		}
	}
	for _, value := range append([]string{rollback.RollbackCommand}, rollback.RollbackScope...) {
		lower := strings.ToLower(value)
		for _, forbidden := range []string{"reset --hard", "rm -rf", "drop database", "delete production data", "release", "deploy", "publish", "upload", "tag"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("prompt-only rollback must not include destructive or release action %q in %q", forbidden, value)
			}
		}
	}

	nodeTwentyFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyFourReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyFourReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyFourReadback.CompletedNodes != 24 ||
		nodeTwentyFourReadback.ReadyNodes != 16 ||
		nodeTwentyFourReadback.FirstExecutableNode != "mission-recommendation-hardening-25" ||
		nodeTwentyFourReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyFourReadback.ExactNextAction, "mission-recommendation-hardening-25") {
		t.Fatalf("node 24 readback must carry rollback boundary and continue to node 25: %#v", nodeTwentyFourReadback)
	}
}

func TestLongRunHardeningWaveSupportEvidenceNodeGateCannotWidenAuthority(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-24", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-25")
	fixture := mustLoadJSON[struct {
		Schema                       string   `json:"schema"`
		NodeID                       string   `json:"node_id"`
		Status                       string   `json:"status"`
		SupportEvidenceNode          bool     `json:"support_evidence_node"`
		AuthorityBoundary            string   `json:"authority_boundary"`
		NodeGateAuthorityBoundary    string   `json:"node_gate_authority_boundary"`
		FoundryTaskAuthorityBoundary string   `json:"foundry_task_authority_boundary"`
		AuthorityWideningAllowed     bool     `json:"authority_widening_allowed"`
		AllowedWriteScopes           []string `json:"allowed_write_scopes"`
		ForbiddenBoundaryClaims      []string `json:"forbidden_boundary_claims"`
		CompletedNodesBeforeNode     int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode         int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed         bool     `json:"final_response_allowed"`
		ExactNextAction              string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint   struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "authority-boundary-gate-fixture.json"))
	nodeGate := mustLoadJSON[struct {
		Schema                    string `json:"schema"`
		NodeID                    string `json:"node_id"`
		Status                    string `json:"status"`
		SupportEvidenceNode       bool   `json:"support_evidence_node"`
		AuthorityBoundary         string `json:"authority_boundary"`
		AuthorityWideningAllowed  bool   `json:"authority_widening_allowed"`
		OneExecutableMutationNode bool   `json:"one_executable_mutation_node_active"`
		EntryReadiness            struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
		} `json:"entry_readiness"`
	}](t, filepath.Join(nodeDir, "node_gate.json"))
	foundryImport := mustLoadJSON[FoundryImport](t, filepath.Join(nodeDir, "foundry-import.json"))

	if fixture.Schema != "ao.atlas.authority-boundary-gate-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-25" ||
		fixture.Status != "authority_boundary_recorded" ||
		!fixture.SupportEvidenceNode ||
		fixture.AuthorityBoundary != "atlas_recommendation_planning_only" ||
		fixture.NodeGateAuthorityBoundary != fixture.AuthorityBoundary ||
		fixture.FoundryTaskAuthorityBoundary != fixture.AuthorityBoundary ||
		fixture.AuthorityWideningAllowed ||
		fixture.CompletedNodesBeforeNode != nodeTwentyFourReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyFourReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyFourReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("authority boundary fixture must bind support evidence gate without widening authority: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyFourReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyFourReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyFourReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyFourReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyFourReadback.ExactNextAction {
		t.Fatalf("authority boundary fixture must bind node 24 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}
	if nodeGate.Schema != "ao.atlas.node-gate.v0.1" ||
		nodeGate.NodeID != fixture.NodeID ||
		nodeGate.Status != "opened" ||
		!nodeGate.SupportEvidenceNode ||
		nodeGate.AuthorityBoundary != fixture.AuthorityBoundary ||
		nodeGate.AuthorityWideningAllowed ||
		!nodeGate.OneExecutableMutationNode ||
		nodeGate.EntryReadiness.CompletedNodes != nodeTwentyFourReadback.CompletedNodes ||
		nodeGate.EntryReadiness.ReadyNodes != nodeTwentyFourReadback.ReadyNodes ||
		nodeGate.EntryReadiness.FirstExecutableNode != nodeTwentyFourReadback.FirstExecutableNode ||
		nodeGate.EntryReadiness.FinalResponseAllowed {
		t.Fatalf("node gate must preserve authority boundary and current readiness: %#v", nodeGate)
	}
	if len(foundryImport.Tasks) != 1 ||
		foundryImport.Tasks[0].NodeID != fixture.NodeID ||
		foundryImport.Tasks[0].AuthorityBoundary != fixture.AuthorityBoundary {
		t.Fatalf("Foundry import task must preserve node authority boundary: %#v", foundryImport.Tasks)
	}
	for _, scope := range fixture.AllowedWriteScopes {
		if scope != "internal/atlas" && scope != "schemas" && scope != "examples" && scope != "docs/evidence" {
			t.Fatalf("authority boundary fixture contains unexpected write scope %q", scope)
		}
	}
	for _, claim := range fixture.ForbiddenBoundaryClaims {
		if claim == "" || strings.Contains(strings.ToLower(claim), "allowed") {
			t.Fatalf("forbidden boundary claim must be explicit denial text, got %q", claim)
		}
	}

	nodeTwentyFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyFiveReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyFiveReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyFiveReadback.CompletedNodes != 25 ||
		nodeTwentyFiveReadback.ReadyNodes != 15 ||
		nodeTwentyFiveReadback.FirstExecutableNode != "mission-recommendation-hardening-26" ||
		nodeTwentyFiveReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyFiveReadback.ExactNextAction, "mission-recommendation-hardening-26") {
		t.Fatalf("node 25 readback must carry authority boundary gate and continue to node 26: %#v", nodeTwentyFiveReadback)
	}
}

func TestLongRunHardeningWaveBranchCleanupEvidenceRequiresLocalAndRemoteCodexRemoval(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-25", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-26")
	fixture := mustLoadJSON[struct {
		Schema                        string   `json:"schema"`
		NodeID                        string   `json:"node_id"`
		Status                        string   `json:"status"`
		CleanupScope                  string   `json:"cleanup_scope"`
		CurrentNodeBranch             string   `json:"current_node_branch"`
		LocalCodexBranchesBeforeNode  []string `json:"local_codex_branches_before_node"`
		RemoteCodexBranchesBeforeNode []string `json:"remote_codex_branches_before_node"`
		LocalCodexBranchCountBefore   int      `json:"local_codex_branch_count_before_node"`
		RemoteCodexBranchCountBefore  int      `json:"remote_codex_branch_count_before_node"`
		PostMergeCleanupRequired      bool     `json:"post_merge_cleanup_required"`
		LocalBranchCleanupCommand     string   `json:"local_branch_cleanup_command"`
		RemoteBranchCleanupCommand    string   `json:"remote_branch_cleanup_command"`
		DirectMainMutation            bool     `json:"direct_main_mutation"`
		CompletedNodesBeforeNode      int      `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode          int      `json:"ready_nodes_before_node"`
		FinalResponseAllowed          bool     `json:"final_response_allowed"`
		ExactNextAction               string   `json:"exact_next_action"`
		CurrentHardeningCheckpoint    struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "branch-cleanup-evidence-fixture.json"))

	if fixture.Schema != "ao.atlas.branch-cleanup-evidence-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-26" ||
		fixture.Status != "branch_cleanup_recorded" ||
		fixture.CleanupScope != "after_previous_node_merge_before_current_node" ||
		fixture.CurrentNodeBranch != "codex/hardening-wave-node-26-branch-cleanup" ||
		len(fixture.LocalCodexBranchesBeforeNode) != 0 ||
		len(fixture.RemoteCodexBranchesBeforeNode) != 0 ||
		fixture.LocalCodexBranchCountBefore != 0 ||
		fixture.RemoteCodexBranchCountBefore != 0 ||
		!fixture.PostMergeCleanupRequired ||
		fixture.DirectMainMutation ||
		fixture.CompletedNodesBeforeNode != nodeTwentyFiveReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentyFiveReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyFiveReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("branch cleanup fixture must prove prior codex branch cleanup without authority effects: %#v", fixture)
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentyFiveReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentyFiveReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentyFiveReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentyFiveReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentyFiveReadback.ExactNextAction {
		t.Fatalf("branch cleanup fixture must bind node 25 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}
	if !strings.Contains(fixture.LocalBranchCleanupCommand, fixture.CurrentNodeBranch) ||
		!strings.Contains(fixture.RemoteBranchCleanupCommand, "delete-branch") {
		t.Fatalf("branch cleanup fixture must name local and remote cleanup commands: %#v", fixture)
	}

	nodeTwentySixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentySixReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentySixReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentySixReadback.CompletedNodes != 26 ||
		nodeTwentySixReadback.ReadyNodes != 14 ||
		nodeTwentySixReadback.FirstExecutableNode != "mission-recommendation-hardening-27" ||
		nodeTwentySixReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentySixReadback.ExactNextAction, "mission-recommendation-hardening-27") {
		t.Fatalf("node 26 readback must carry branch cleanup evidence and continue to node 27: %#v", nodeTwentySixReadback)
	}
}

func TestLongRunHardeningWavePRLedgerFixtureBindsMergeCIAndCleanupEvidence(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentySixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-26", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-27")
	fixture := mustLoadJSON[struct {
		Schema                string `json:"schema"`
		NodeID                string `json:"node_id"`
		Status                string `json:"status"`
		LedgerScope           string `json:"ledger_scope"`
		PreviousNodeID        string `json:"previous_node_id"`
		PreviousPRNumber      int    `json:"previous_pr_number"`
		PreviousPRURL         string `json:"previous_pr_url"`
		PreviousHeadBranch    string `json:"previous_head_branch"`
		PreviousMergeCommit   string `json:"previous_merge_commit"`
		PreviousPRState       string `json:"previous_pr_state"`
		PreviousCIStatus      string `json:"previous_ci_status"`
		PreviousBranchCleanup struct {
			LocalCodexBranchesAfterMerge  []string `json:"local_codex_branches_after_merge"`
			RemoteCodexBranchesAfterMerge []string `json:"remote_codex_branches_after_merge"`
			LocalCodexBranchCountAfter    int      `json:"local_codex_branch_count_after_merge"`
			RemoteCodexBranchCountAfter   int      `json:"remote_codex_branch_count_after_merge"`
		} `json:"previous_branch_cleanup"`
		CIChecks []struct {
			Name       string `json:"name"`
			Workflow   string `json:"workflow"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"ci_checks"`
		DirectMainMutation         bool   `json:"direct_main_mutation"`
		CompletedNodesBeforeNode   int    `json:"completed_nodes_before_node"`
		ReadyNodesBeforeNode       int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed       bool   `json:"final_response_allowed"`
		ExactNextAction            string `json:"exact_next_action"`
		CurrentHardeningCheckpoint struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "github-pr-ledger-fixture.json"))

	if fixture.Schema != "ao.atlas.github-pr-ledger-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-27" ||
		fixture.Status != "pr_ledger_recorded" ||
		fixture.LedgerScope != "previous_node_merge_ci_and_cleanup" ||
		fixture.PreviousNodeID != "mission-recommendation-hardening-26" ||
		fixture.PreviousPRNumber != 289 ||
		!strings.Contains(fixture.PreviousPRURL, "/pull/289") ||
		fixture.PreviousHeadBranch != "codex/hardening-wave-node-26-branch-cleanup" ||
		fixture.PreviousMergeCommit != "b6f5ee71d716070e24201ff3c7d5e1e1d7b0f905" ||
		fixture.PreviousPRState != "MERGED" ||
		fixture.PreviousCIStatus != "pass" ||
		fixture.DirectMainMutation ||
		fixture.CompletedNodesBeforeNode != nodeTwentySixReadback.CompletedNodes ||
		fixture.ReadyNodesBeforeNode != nodeTwentySixReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentySixReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("PR ledger fixture must bind previous PR, CI, cleanup, and no-authority state: %#v", fixture)
	}
	if len(fixture.PreviousBranchCleanup.LocalCodexBranchesAfterMerge) != 0 ||
		len(fixture.PreviousBranchCleanup.RemoteCodexBranchesAfterMerge) != 0 ||
		fixture.PreviousBranchCleanup.LocalCodexBranchCountAfter != 0 ||
		fixture.PreviousBranchCleanup.RemoteCodexBranchCountAfter != 0 {
		t.Fatalf("PR ledger fixture must prove local and remote codex branch cleanup: %#v", fixture.PreviousBranchCleanup)
	}
	if len(fixture.CIChecks) < 9 {
		t.Fatalf("PR ledger fixture must include full CI/readiness check rollup, got %d checks", len(fixture.CIChecks))
	}
	for _, check := range fixture.CIChecks {
		if check.Status != "COMPLETED" || check.Conclusion != "SUCCESS" || check.Name == "" || check.Workflow == "" {
			t.Fatalf("PR ledger fixture check must be completed and successful: %#v", check)
		}
	}
	if fixture.CurrentHardeningCheckpoint.CompletedNodes != nodeTwentySixReadback.CompletedNodes ||
		fixture.CurrentHardeningCheckpoint.ReadyNodes != nodeTwentySixReadback.ReadyNodes ||
		fixture.CurrentHardeningCheckpoint.FirstExecutableNode != nodeTwentySixReadback.FirstExecutableNode ||
		fixture.CurrentHardeningCheckpoint.FinalResponseAllowed != nodeTwentySixReadback.FinalResponseAllowed ||
		fixture.CurrentHardeningCheckpoint.ExactNextAction != nodeTwentySixReadback.ExactNextAction {
		t.Fatalf("PR ledger fixture must bind node 26 checkpoint: %#v", fixture.CurrentHardeningCheckpoint)
	}

	nodeTwentySevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentySevenReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentySevenReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentySevenReadback.CompletedNodes != 27 ||
		nodeTwentySevenReadback.ReadyNodes != 13 ||
		nodeTwentySevenReadback.FirstExecutableNode != "mission-recommendation-hardening-28" ||
		nodeTwentySevenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentySevenReadback.ExactNextAction, "mission-recommendation-hardening-28") {
		t.Fatalf("node 27 readback must carry PR ledger evidence and continue to node 28: %#v", nodeTwentySevenReadback)
	}
}

func TestLongRunHardeningWaveCIReadbackFixtureDistinguishesLocalPendingPassFailureStates(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentySevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-27", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-28")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		ReadbackScope        string `json:"readback_scope"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		CurrentCheckpoint    struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		States []struct {
			Name                    string `json:"name"`
			LocalVerificationStatus string `json:"local_verification_status"`
			GitHubCIStatus          string `json:"github_ci_status"`
			MergeAllowed            bool   `json:"merge_allowed"`
			NodeClosureAllowed      bool   `json:"node_closure_allowed"`
			FinalResponseAllowed    bool   `json:"final_response_allowed"`
			RequiresRepair          bool   `json:"requires_repair"`
			ExactNextAction         string `json:"exact_next_action"`
		} `json:"states"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "ci-readback-state-fixture.json"))

	if fixture.Schema != "ao.atlas.ci-readback-state-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-28" ||
		fixture.Status != "ci_readback_states_recorded" ||
		fixture.ReadbackScope != "local_verification_and_remote_ci_lifecycle" ||
		fixture.CompletedNodesBefore != nodeTwentySevenReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeTwentySevenReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentySevenReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("CI readback fixture must bind checkpoint state without authority effects: %#v", fixture)
	}
	if fixture.CurrentCheckpoint.CompletedNodes != nodeTwentySevenReadback.CompletedNodes ||
		fixture.CurrentCheckpoint.ReadyNodes != nodeTwentySevenReadback.ReadyNodes ||
		fixture.CurrentCheckpoint.FirstExecutableNode != nodeTwentySevenReadback.FirstExecutableNode ||
		fixture.CurrentCheckpoint.FinalResponseAllowed != nodeTwentySevenReadback.FinalResponseAllowed ||
		fixture.CurrentCheckpoint.ExactNextAction != nodeTwentySevenReadback.ExactNextAction {
		t.Fatalf("CI readback fixture must bind node 27 checkpoint: %#v", fixture.CurrentCheckpoint)
	}

	states := map[string]struct {
		local       string
		ci          string
		merge       bool
		closeNode   bool
		final       bool
		repair      bool
		actionMatch string
	}{
		"local_pass": {"passed", "not_started", false, false, false, false, "open PR"},
		"ci_pending": {"passed", "pending", false, false, false, false, "wait for CI"},
		"ci_pass":    {"passed", "passed", true, true, false, false, "merge PR"},
		"ci_failure": {"passed", "failed", false, false, false, true, "repair failing CI"},
	}
	if len(fixture.States) != len(states) {
		t.Fatalf("CI readback fixture must contain exactly %d states, got %d", len(states), len(fixture.States))
	}
	for _, state := range fixture.States {
		want, ok := states[state.Name]
		if !ok {
			t.Fatalf("unexpected CI readback state %q", state.Name)
		}
		if state.LocalVerificationStatus != want.local ||
			state.GitHubCIStatus != want.ci ||
			state.MergeAllowed != want.merge ||
			state.NodeClosureAllowed != want.closeNode ||
			state.FinalResponseAllowed != want.final ||
			state.RequiresRepair != want.repair ||
			!strings.Contains(state.ExactNextAction, want.actionMatch) {
			t.Fatalf("CI readback state %q mismatch: %#v", state.Name, state)
		}
		delete(states, state.Name)
	}
	if len(states) != 0 {
		t.Fatalf("missing CI readback states: %#v", states)
	}

	nodeTwentyEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyEightReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyEightReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyEightReadback.CompletedNodes != 28 ||
		nodeTwentyEightReadback.ReadyNodes != 12 ||
		nodeTwentyEightReadback.FirstExecutableNode != "mission-recommendation-hardening-29" ||
		nodeTwentyEightReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyEightReadback.ExactNextAction, "mission-recommendation-hardening-29") {
		t.Fatalf("node 28 readback must carry CI readback states and continue to node 29: %#v", nodeTwentyEightReadback)
	}
}

func TestLongRunHardeningWaveRouteDecisionReadbackExplainsBlueprintBypassForFoundryImplementation(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-28", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-29")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		ReadbackScope        string `json:"readback_scope"`
		SelectedRoute        string `json:"selected_route"`
		BlueprintRouteStatus string `json:"blueprint_route_status"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		CurrentCheckpoint    struct {
			CompletedNodes       int    `json:"completed_nodes"`
			ReadyNodes           int    `json:"ready_nodes"`
			FirstExecutableNode  string `json:"first_executable_node"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ExactNextAction      string `json:"exact_next_action"`
		} `json:"current_hardening_checkpoint"`
		RouteDecisions []struct {
			Owner                 string `json:"owner"`
			Status                string `json:"status"`
			Reason                string `json:"reason"`
			RequiresAuthorization bool   `json:"requires_new_authorization"`
			RequiresGovernedPlan  bool   `json:"requires_new_governed_plan"`
			ExactlyOneActiveNode  bool   `json:"exactly_one_active_node"`
		} `json:"route_decisions"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "route-decision-readback-fixture.json"))

	if fixture.Schema != "ao.atlas.route-decision-readback-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-29" ||
		fixture.Status != "route_decision_recorded" ||
		fixture.ReadbackScope != "atlas_to_foundry_without_blueprint_for_ready_bounded_implementation" ||
		fixture.SelectedRoute != "ao-foundry" ||
		fixture.BlueprintRouteStatus != "not_required_for_ready_bounded_implementation" ||
		fixture.CompletedNodesBefore != nodeTwentyEightReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeTwentyEightReadback.ReadyNodes ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyEightReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("route decision fixture must bind checkpoint state without authority effects: %#v", fixture)
	}
	if fixture.CurrentCheckpoint.CompletedNodes != nodeTwentyEightReadback.CompletedNodes ||
		fixture.CurrentCheckpoint.ReadyNodes != nodeTwentyEightReadback.ReadyNodes ||
		fixture.CurrentCheckpoint.FirstExecutableNode != nodeTwentyEightReadback.FirstExecutableNode ||
		fixture.CurrentCheckpoint.FinalResponseAllowed != nodeTwentyEightReadback.FinalResponseAllowed ||
		fixture.CurrentCheckpoint.ExactNextAction != nodeTwentyEightReadback.ExactNextAction {
		t.Fatalf("route decision fixture must bind node 28 checkpoint: %#v", fixture.CurrentCheckpoint)
	}

	decisions := map[string]struct {
		status          string
		reasonContains  string
		newAuth         bool
		newGovernedPlan bool
	}{
		"ao-atlas":     {"coordinate", "workgraph", false, false},
		"ao-foundry":   {"selected", "ready bounded implementation", false, false},
		"ao-blueprint": {"bypassed", "new authorization", false, false},
	}
	if len(fixture.RouteDecisions) != len(decisions) {
		t.Fatalf("route decision fixture must contain exactly %d decisions, got %d", len(decisions), len(fixture.RouteDecisions))
	}
	for _, decision := range fixture.RouteDecisions {
		want, ok := decisions[decision.Owner]
		if !ok {
			t.Fatalf("unexpected route decision owner %q", decision.Owner)
		}
		if decision.Status != want.status ||
			!strings.Contains(decision.Reason, want.reasonContains) ||
			decision.RequiresAuthorization != want.newAuth ||
			decision.RequiresGovernedPlan != want.newGovernedPlan ||
			!decision.ExactlyOneActiveNode {
			t.Fatalf("route decision %q mismatch: %#v", decision.Owner, decision)
		}
		delete(decisions, decision.Owner)
	}
	if len(decisions) != 0 {
		t.Fatalf("missing route decisions: %#v", decisions)
	}

	nodeTwentyNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeTwentyNineReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeTwentyNineReadback.FeatureDepthRecommendations) < 40 ||
		nodeTwentyNineReadback.CompletedNodes != 29 ||
		nodeTwentyNineReadback.ReadyNodes != 11 ||
		nodeTwentyNineReadback.FirstExecutableNode != "mission-recommendation-hardening-30" ||
		nodeTwentyNineReadback.FinalResponseAllowed ||
		!strings.Contains(nodeTwentyNineReadback.ExactNextAction, "mission-recommendation-hardening-30") {
		t.Fatalf("node 29 readback must carry route decision evidence and continue to node 30: %#v", nodeTwentyNineReadback)
	}
}

func TestLongRunHardeningWaveCompactionResumePromptSkipsCompletedNodes(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-29", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-30")
	fixture := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		ResumePromptPath           string   `json:"resume_prompt_path"`
		SourceCheckpointReadback   string   `json:"source_checkpoint_readback"`
		CompletedNodesBefore       int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore           int      `json:"ready_nodes_before_node"`
		FirstExecutableNode        string   `json:"first_executable_node"`
		FinalResponseAllowed       bool     `json:"final_response_allowed"`
		ExactNextAction            string   `json:"exact_next_action"`
		ResumeUsesLatestCheckpoint bool     `json:"resume_uses_latest_checkpoint"`
		CompletedNodesReadOnly     bool     `json:"completed_nodes_read_only"`
		RerunCompletedNodes        bool     `json:"rerun_completed_nodes"`
		FirstNodeToExecute         string   `json:"first_node_to_execute"`
		PromptRequiredPhrases      []string `json:"prompt_required_phrases"`
		SchedulesWork              bool     `json:"schedules_work"`
		ExecutesWork               bool     `json:"executes_work"`
		ApprovesWork               bool     `json:"approves_work"`
		ClaimsAuthorityAdvance     bool     `json:"claims_authority_advance"`
		RSIRemainsDenied           bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "compaction-resume-prompt-fixture.json"))

	if fixture.Schema != "ao.atlas.compaction-resume-prompt-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-30" ||
		fixture.Status != "resume_prompt_recorded" ||
		fixture.SourceCheckpointReadback != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-29/recommendation-readback-after.json" ||
		fixture.CompletedNodesBefore != nodeTwentyNineReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeTwentyNineReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeTwentyNineReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeTwentyNineReadback.ExactNextAction ||
		!fixture.ResumeUsesLatestCheckpoint ||
		!fixture.CompletedNodesReadOnly ||
		fixture.RerunCompletedNodes ||
		fixture.FirstNodeToExecute != "mission-recommendation-hardening-30" ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("compaction resume fixture must bind node 29 checkpoint and skip completed nodes: %#v", fixture)
	}

	prompt, err := os.ReadFile(filepath.Join(repoRoot(t), fixture.ResumePromptPath))
	if err != nil {
		t.Fatal(err)
	}
	promptText := string(prompt)
	for _, phrase := range fixture.PromptRequiredPhrases {
		if !strings.Contains(promptText, phrase) {
			t.Fatalf("compaction resume prompt missing required phrase %q in %s", phrase, fixture.ResumePromptPath)
		}
	}
	if strings.Contains(promptText, "Start from mission-recommendation-hardening-01") ||
		strings.Contains(promptText, "rerun completed nodes") {
		t.Fatalf("compaction resume prompt must not restart or rerun completed nodes: %s", fixture.ResumePromptPath)
	}

	nodeThirtyReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyReadback.CompletedNodes != 30 ||
		nodeThirtyReadback.ReadyNodes != 10 ||
		nodeThirtyReadback.FirstExecutableNode != "mission-recommendation-hardening-31" ||
		nodeThirtyReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyReadback.ExactNextAction, "mission-recommendation-hardening-31") {
		t.Fatalf("node 30 readback must carry compaction resume prompt evidence and continue to node 31: %#v", nodeThirtyReadback)
	}
}

func TestLongRunHardeningWaveOperatorRoutingGuideCoversAORoles(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-30", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-31")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		GuidePath            string `json:"guide_path"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FirstExecutableNode  string `json:"first_executable_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		BlueprintBoundary    string `json:"blueprint_boundary"`
		FoundryBoundary      string `json:"foundry_boundary"`
		Roles                []struct {
			Name     string `json:"name"`
			UseWhen  string `json:"use_when"`
			Boundary string `json:"boundary"`
		} `json:"roles"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "operator-routing-guide-fixture.json"))

	if fixture.Schema != "ao.atlas.operator-routing-guide-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-31" ||
		fixture.Status != "operator_routing_documented" ||
		fixture.CompletedNodesBefore != nodeThirtyReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyReadback.ExactNextAction ||
		!strings.Contains(fixture.BlueprintBoundary, "new authorization") ||
		!strings.Contains(fixture.FoundryBoundary, "bounded implementation") ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("operator routing fixture must bind node 30 checkpoint without authority effects: %#v", fixture)
	}

	expected := map[string]string{
		"AO Mission":      "operator-facing loop",
		"AO Atlas":        "workgraph state",
		"AO Blueprint":    "new authorization",
		"AO Foundry":      "bounded implementation",
		"AO Promoter":     "promotion",
		"AO Command":      "readback",
		"AO Sentinel":     "public-safety",
		"AO Architecture": "capability map",
	}
	if len(fixture.Roles) != len(expected) {
		t.Fatalf("operator routing fixture must cover all AO roles, got %d roles: %#v", len(fixture.Roles), fixture.Roles)
	}
	guideData, err := os.ReadFile(filepath.Join(repoRoot(t), fixture.GuidePath))
	if err != nil {
		t.Fatal(err)
	}
	guide := string(guideData)
	for _, role := range fixture.Roles {
		want, ok := expected[role.Name]
		if !ok {
			t.Fatalf("unexpected AO role in routing guide fixture: %#v", role)
		}
		if !strings.Contains(role.UseWhen, want) ||
			role.Boundary == "" ||
			!strings.Contains(guide, "## "+role.Name) {
			t.Fatalf("role %q missing expected routing guidance: %#v", role.Name, role)
		}
		delete(expected, role.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing AO roles in routing guide fixture: %#v", expected)
	}
	for _, phrase := range []string{
		"Blueprint only when new authorization or governed planning is required",
		"Foundry for ready bounded implementation",
		"Atlas owns workgraph state and next-node continuation",
		"Mission is the operator-facing loop",
	} {
		if !strings.Contains(guide, phrase) {
			t.Fatalf("operator routing guide missing %q", phrase)
		}
	}

	nodeThirtyOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyOneReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyOneReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyOneReadback.CompletedNodes != 31 ||
		nodeThirtyOneReadback.ReadyNodes != 9 ||
		nodeThirtyOneReadback.FirstExecutableNode != "mission-recommendation-hardening-32" ||
		nodeThirtyOneReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyOneReadback.ExactNextAction, "mission-recommendation-hardening-32") {
		t.Fatalf("node 31 readback must carry operator routing evidence and continue to node 32: %#v", nodeThirtyOneReadback)
	}
}

func TestLongRunHardeningWavePrematureReturnGuideRejectsShortLoops(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-31", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-32")
	fixture := mustLoadJSON[struct {
		Schema                           string   `json:"schema"`
		NodeID                           string   `json:"node_id"`
		Status                           string   `json:"status"`
		GuidePath                        string   `json:"guide_path"`
		CompletedNodesBefore             int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore                 int      `json:"ready_nodes_before_node"`
		FirstExecutableNode              string   `json:"first_executable_node"`
		FinalResponseAllowed             bool     `json:"final_response_allowed"`
		ExactNextAction                  string   `json:"exact_next_action"`
		PrematureLoopLowerMinutes        int      `json:"premature_loop_lower_minutes"`
		PrematureLoopUpperMinutes        int      `json:"premature_loop_upper_minutes"`
		RequiredMinimumMinutes           int      `json:"required_minimum_minutes"`
		RequiredTargetMinutes            int      `json:"required_target_minutes"`
		FinalResponseDeniedWithReadyWork bool     `json:"final_response_denied_with_ready_work"`
		Reasons                          []string `json:"reasons"`
		SchedulesWork                    bool     `json:"schedules_work"`
		ExecutesWork                     bool     `json:"executes_work"`
		ApprovesWork                     bool     `json:"approves_work"`
		ClaimsAuthorityAdvance           bool     `json:"claims_authority_advance"`
		RSIRemainsDenied                 bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "premature-return-guide-fixture.json"))

	if fixture.Schema != "ao.atlas.premature-return-guide-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-32" ||
		fixture.Status != "premature_return_documented" ||
		fixture.CompletedNodesBefore != nodeThirtyOneReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyOneReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyOneReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyOneReadback.ExactNextAction ||
		fixture.PrematureLoopLowerMinutes != 14 ||
		fixture.PrematureLoopUpperMinutes != 20 ||
		fixture.RequiredMinimumMinutes != 120 ||
		fixture.RequiredTargetMinutes != 180 ||
		!fixture.FinalResponseDeniedWithReadyWork ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("premature return fixture must bind node 31 checkpoint and deny short-loop closure: %#v", fixture)
	}
	reasons := map[string]bool{}
	for _, reason := range fixture.Reasons {
		reasons[reason] = true
	}
	for _, reason := range []string{
		"minimum_minutes_unmet",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"single_pr_is_not_mission_completion",
	} {
		if !reasons[reason] {
			t.Fatalf("premature return fixture missing reason %q: %#v", reason, fixture.Reasons)
		}
	}

	guideData, err := os.ReadFile(filepath.Join(repoRoot(t), fixture.GuidePath))
	if err != nil {
		t.Fatal(err)
	}
	guide := string(guideData)
	for _, phrase := range []string{
		"14 to 20 minute loops are premature returns",
		"2 to 3 hour workgraph",
		"ready_nodes > 0",
		"exact_next_action",
		"final_response_allowed=false",
		"one PR merge is not mission completion",
	} {
		if !strings.Contains(guide, phrase) {
			t.Fatalf("premature return guide missing %q", phrase)
		}
	}

	nodeThirtyTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyTwoReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyTwoReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyTwoReadback.CompletedNodes != 32 ||
		nodeThirtyTwoReadback.ReadyNodes != 8 ||
		nodeThirtyTwoReadback.FirstExecutableNode != "mission-recommendation-hardening-33" ||
		nodeThirtyTwoReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyTwoReadback.ExactNextAction, "mission-recommendation-hardening-33") {
		t.Fatalf("node 32 readback must carry premature-return guide evidence and continue to node 33: %#v", nodeThirtyTwoReadback)
	}
}

func TestLongRunHardeningWaveDoctorReadinessFixtureCoversContinuationRisks(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyTwoReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-32", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-33")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FirstExecutableNode  string `json:"first_executable_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		FinalResponseDenied  bool   `json:"final_response_denied"`
		SafeNextAction       string `json:"safe_next_action"`
		DoctorChecks         []struct {
			Name             string `json:"name"`
			Status           string `json:"status"`
			BlocksIfStale    bool   `json:"blocks_if_stale"`
			RequiredReadback string `json:"required_readback"`
		} `json:"doctor_checks"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "doctor-readiness-fixture.json"))

	if fixture.Schema != "ao.atlas.doctor-readiness-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-33" ||
		fixture.Status != "doctor_readiness_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtyTwoReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyTwoReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyTwoReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyTwoReadback.ExactNextAction ||
		!fixture.FinalResponseDenied ||
		!strings.Contains(fixture.SafeNextAction, "mission-recommendation-hardening-33") ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("doctor readiness fixture must bind node 32 checkpoint without authority effects: %#v", fixture)
	}
	expected := map[string]string{
		"lease_health":            "lease_health_status",
		"checkpoint_freshness":    "checkpoint_freshness_status",
		"stale_routes":            "stale_route_decision_status",
		"shallow_recommendations": "feature_depth_recommendations",
		"early_return_risk":       "final_response_allowed",
	}
	if len(fixture.DoctorChecks) != len(expected) {
		t.Fatalf("doctor readiness fixture must include all checks, got %d: %#v", len(fixture.DoctorChecks), fixture.DoctorChecks)
	}
	for _, check := range fixture.DoctorChecks {
		want, ok := expected[check.Name]
		if !ok {
			t.Fatalf("unexpected doctor readiness check: %#v", check)
		}
		if check.Status != "ready" ||
			!check.BlocksIfStale ||
			check.RequiredReadback != want {
			t.Fatalf("doctor readiness check %q mismatch: %#v", check.Name, check)
		}
		delete(expected, check.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing doctor readiness checks: %#v", expected)
	}

	nodeThirtyThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyThreeReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyThreeReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyThreeReadback.CompletedNodes != 33 ||
		nodeThirtyThreeReadback.ReadyNodes != 7 ||
		nodeThirtyThreeReadback.FirstExecutableNode != "mission-recommendation-hardening-34" ||
		nodeThirtyThreeReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyThreeReadback.ExactNextAction, "mission-recommendation-hardening-34") {
		t.Fatalf("node 33 readback must carry doctor readiness evidence and continue to node 34: %#v", nodeThirtyThreeReadback)
	}
}

func TestLongRunHardeningWaveEarlyReturnRiskDeniesUnmetMinimums(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-33", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-34")
	fixture := mustLoadJSON[struct {
		Schema               string `json:"schema"`
		NodeID               string `json:"node_id"`
		Status               string `json:"status"`
		CompletedNodesBefore int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore     int    `json:"ready_nodes_before_node"`
		FirstExecutableNode  string `json:"first_executable_node"`
		FinalResponseAllowed bool   `json:"final_response_allowed"`
		ExactNextAction      string `json:"exact_next_action"`
		Cases                []struct {
			Name                 string `json:"name"`
			CompletedNodes       int    `json:"completed_nodes"`
			ElapsedMinutes       int    `json:"elapsed_minutes"`
			ReadyNodes           int    `json:"ready_nodes"`
			ExactNextAction      string `json:"exact_next_action"`
			MinNodesMet          bool   `json:"min_nodes_met"`
			MinMinutesMet        bool   `json:"min_minutes_met"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
			ReturnGateStatus     string `json:"return_gate_status"`
		} `json:"cases"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "early-return-risk-regression-fixture.json"))

	if fixture.Schema != "ao.atlas.early-return-risk-regression-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-34" ||
		fixture.Status != "early_return_risk_cases_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtyThreeReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyThreeReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyThreeReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyThreeReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("early-return fixture must bind node 33 checkpoint without authority effects: %#v", fixture)
	}
	cases := map[string]string{
		"min_nodes_unmet":                          "blocked_minimum_nodes_unmet",
		"min_minutes_unmet":                        "blocked_minimum_minutes_unmet",
		"ready_nodes_remain_after_minimums":        "blocked_ready_nodes_remain",
		"exact_next_action_remains_after_minimums": "blocked_exact_next_action_remains",
	}
	if len(fixture.Cases) != len(cases) {
		t.Fatalf("early-return fixture must include all cases, got %d: %#v", len(fixture.Cases), fixture.Cases)
	}
	for _, item := range fixture.Cases {
		want, ok := cases[item.Name]
		if !ok {
			t.Fatalf("unexpected early-return case: %#v", item)
		}
		if item.FinalResponseAllowed || item.ReturnGateStatus != want {
			t.Fatalf("early-return case %q should deny final response with %q: %#v", item.Name, want, item)
		}
		delete(cases, item.Name)
	}
	if len(cases) != 0 {
		t.Fatalf("missing early-return cases: %#v", cases)
	}

	nodeThirtyFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyFourReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyFourReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyFourReadback.CompletedNodes != 34 ||
		nodeThirtyFourReadback.ReadyNodes != 6 ||
		nodeThirtyFourReadback.FirstExecutableNode != "mission-recommendation-hardening-35" ||
		nodeThirtyFourReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyFourReadback.ExactNextAction, "mission-recommendation-hardening-35") {
		t.Fatalf("node 34 readback must carry early-return regression evidence and continue to node 35: %#v", nodeThirtyFourReadback)
	}
}

func TestLongRunHardeningWaveExactNextActionPropagatesToSummaryAndPrompt(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-34", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-35")
	fixture := mustLoadJSON[struct {
		Schema                 string   `json:"schema"`
		NodeID                 string   `json:"node_id"`
		Status                 string   `json:"status"`
		SummaryPath            string   `json:"summary_path"`
		PromptPath             string   `json:"prompt_path"`
		CompletedNodesBefore   int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore       int      `json:"ready_nodes_before_node"`
		FirstExecutableNode    string   `json:"first_executable_node"`
		FinalResponseAllowed   bool     `json:"final_response_allowed"`
		ExactNextAction        string   `json:"exact_next_action"`
		RequiredArtifacts      []string `json:"required_artifacts"`
		SchedulesWork          bool     `json:"schedules_work"`
		ExecutesWork           bool     `json:"executes_work"`
		ApprovesWork           bool     `json:"approves_work"`
		ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
		RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "exact-next-action-propagation-fixture.json"))

	if fixture.Schema != "ao.atlas.exact-next-action-propagation-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-35" ||
		fixture.Status != "exact_next_action_propagated" ||
		fixture.CompletedNodesBefore != nodeThirtyFourReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyFourReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyFourReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyFourReadback.ExactNextAction ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("exact next action fixture must bind node 34 checkpoint without authority effects: %#v", fixture)
	}
	if !containsString(fixture.RequiredArtifacts, "summary") || !containsString(fixture.RequiredArtifacts, "generated_prompt") {
		t.Fatalf("exact next action fixture must require summary and generated prompt artifacts: %#v", fixture.RequiredArtifacts)
	}
	for _, path := range []string{fixture.SummaryPath, fixture.PromptPath} {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), path))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if !strings.Contains(text, fixture.ExactNextAction) ||
			!strings.Contains(text, "mission-recommendation-hardening-35") ||
			!strings.Contains(text, "final_response_allowed=false") {
			t.Fatalf("%s must carry exact next action and final-response denial", path)
		}
	}

	nodeThirtyFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyFiveReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyFiveReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyFiveReadback.CompletedNodes != 35 ||
		nodeThirtyFiveReadback.ReadyNodes != 5 ||
		nodeThirtyFiveReadback.FirstExecutableNode != "mission-recommendation-hardening-36" ||
		nodeThirtyFiveReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyFiveReadback.ExactNextAction, "mission-recommendation-hardening-36") {
		t.Fatalf("node 35 readback must carry exact-next-action propagation evidence and continue to node 36: %#v", nodeThirtyFiveReadback)
	}
}

func TestLongRunHardeningWavePublicClaimGuardRejectsUnevidencedPromotionClaims(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-35", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-36")
	fixture := mustLoadJSON[struct {
		Schema                          string   `json:"schema"`
		NodeID                          string   `json:"node_id"`
		Status                          string   `json:"status"`
		CompletedNodesBefore            int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore                int      `json:"ready_nodes_before_node"`
		FirstExecutableNode             string   `json:"first_executable_node"`
		FinalResponseAllowed            bool     `json:"final_response_allowed"`
		ExactNextAction                 string   `json:"exact_next_action"`
		PublicDocsScanPath              string   `json:"public_docs_scan_path"`
		PublicDocPaths                  []string `json:"public_doc_paths"`
		ForbiddenClaimTokens            []string `json:"forbidden_claim_tokens"`
		ForbiddenClaimMatches           int      `json:"forbidden_claim_matches"`
		EvidenceRequiredBeforePromotion bool     `json:"evidence_required_before_promotion"`
		PromoterNoPromotionRequired     bool     `json:"promoter_no_promotion_required"`
		CommandReadbackRequired         bool     `json:"command_readback_required"`
		SchedulesWork                   bool     `json:"schedules_work"`
		ExecutesWork                    bool     `json:"executes_work"`
		ApprovesWork                    bool     `json:"approves_work"`
		ClaimsAuthorityAdvance          bool     `json:"claims_authority_advance"`
		RSIRemainsDenied                bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "public-claim-guard-fixture.json"))

	if fixture.Schema != "ao.atlas.public-claim-guard-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-36" ||
		fixture.Status != "public_claim_guard_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtyFiveReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyFiveReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyFiveReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyFiveReadback.ExactNextAction ||
		fixture.ForbiddenClaimMatches != 0 ||
		!fixture.EvidenceRequiredBeforePromotion ||
		!fixture.PromoterNoPromotionRequired ||
		!fixture.CommandReadbackRequired ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("public claim guard must bind node 35 checkpoint without authority effects: %#v", fixture)
	}
	if !containsString(fixture.ForbiddenClaimTokens, "broad_rsi_proof_claim") ||
		!containsString(fixture.ForbiddenClaimTokens, "unsupervised_promotion_live_claim") ||
		!containsString(fixture.ForbiddenClaimTokens, "highest_live_unsupervised_claim") {
		t.Fatalf("public claim guard must record all forbidden claim tokens: %#v", fixture.ForbiddenClaimTokens)
	}
	scan := mustLoadJSON[struct {
		Schema                string   `json:"schema"`
		NodeID                string   `json:"node_id"`
		Status                string   `json:"status"`
		ScannedPaths          []string `json:"scanned_paths"`
		ForbiddenClaimMatches int      `json:"forbidden_claim_matches"`
		PublicDocsScanPassed  bool     `json:"public_docs_scan_passed"`
		PromoterStatus        string   `json:"promoter_status"`
		CommandStatus         string   `json:"command_status"`
		RSIRemainsDenied      bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(repoRoot(t), fixture.PublicDocsScanPath))
	if scan.Schema != "ao.atlas.public-claim-guard-scan.v0.1" ||
		scan.NodeID != fixture.NodeID ||
		scan.Status != "passed" ||
		scan.ForbiddenClaimMatches != 0 ||
		!scan.PublicDocsScanPassed ||
		scan.PromoterStatus != "no_promotion_requested" ||
		scan.CommandStatus != "readback_agrees_no_promotion" ||
		!scan.RSIRemainsDenied {
		t.Fatalf("public claim guard scan must pass without promotion effects: %#v", scan)
	}
	if len(scan.ScannedPaths) != len(fixture.PublicDocPaths) {
		t.Fatalf("public claim guard scan path count mismatch: scan=%#v fixture=%#v", scan.ScannedPaths, fixture.PublicDocPaths)
	}

	forbiddenClaims := []string{
		strings.Join([]string{"RSI", "is", "proven"}, " "),
		strings.Join([]string{"fully_unsupervised_complex_mutation", "is", "live-proven"}, " "),
		"highest proven live class: " + "fully_unsupervised_complex_mutation",
	}
	for _, path := range fixture.PublicDocPaths {
		if !strings.HasPrefix(path, "docs/") {
			t.Fatalf("public doc scan path must stay under docs/: %s", path)
		}
		data, err := os.ReadFile(filepath.Join(repoRoot(t), path))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, claim := range forbiddenClaims {
			if strings.Contains(text, claim) {
				t.Fatalf("%s contains forbidden unevidenced public claim %q", path, claim)
			}
		}
	}

	nodeThirtySixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtySixReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtySixReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtySixReadback.CompletedNodes != 36 ||
		nodeThirtySixReadback.ReadyNodes != 4 ||
		nodeThirtySixReadback.FirstExecutableNode != "mission-recommendation-hardening-37" ||
		nodeThirtySixReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtySixReadback.ExactNextAction, "mission-recommendation-hardening-37") {
		t.Fatalf("node 36 readback must carry public-claim guard evidence and continue to node 37: %#v", nodeThirtySixReadback)
	}
}

func TestLongRunHardeningWaveClosureGateRequiresEvidenceCIAndCleanup(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtySixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-36", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-37")
	fixture := mustLoadJSON[struct {
		Schema                   string `json:"schema"`
		NodeID                   string `json:"node_id"`
		Status                   string `json:"status"`
		CompletedNodesBefore     int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore         int    `json:"ready_nodes_before_node"`
		FirstExecutableNode      string `json:"first_executable_node"`
		FinalResponseAllowed     bool   `json:"final_response_allowed"`
		ExactNextAction          string `json:"exact_next_action"`
		ClosureAllowed           bool   `json:"closure_allowed"`
		CompleteEvidenceRequired bool   `json:"complete_evidence_required"`
		CIPassRequired           bool   `json:"ci_pass_required"`
		BranchCleanupRequired    bool   `json:"branch_cleanup_required"`
		PublicSafetyRequired     bool   `json:"public_safety_required"`
		DenialCases              []struct {
			Name             string `json:"name"`
			MissingEvidence  bool   `json:"missing_evidence"`
			CIPending        bool   `json:"ci_pending"`
			CodexBranchFound bool   `json:"codex_branch_found"`
			ForbiddenSurface bool   `json:"forbidden_surface"`
			ClosureAllowed   bool   `json:"closure_allowed"`
			Reason           string `json:"reason"`
		} `json:"denial_cases"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "closure-gate-fixture.json"))

	if fixture.Schema != "ao.atlas.closure-gate-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-37" ||
		fixture.Status != "closure_gate_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtySixReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtySixReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtySixReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtySixReadback.ExactNextAction ||
		fixture.ClosureAllowed ||
		!fixture.CompleteEvidenceRequired ||
		!fixture.CIPassRequired ||
		!fixture.BranchCleanupRequired ||
		!fixture.PublicSafetyRequired ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("closure gate must bind node 36 checkpoint without authority effects: %#v", fixture)
	}
	wantReasons := map[string]string{
		"missing_evidence":       "blocked_missing_evidence",
		"ci_pending":             "blocked_ci_not_green",
		"codex_branch_remaining": "blocked_branch_cleanup",
		"forbidden_surface":      "blocked_public_safety",
	}
	if len(fixture.DenialCases) != len(wantReasons) {
		t.Fatalf("closure gate must include all denial cases, got %d: %#v", len(fixture.DenialCases), fixture.DenialCases)
	}
	for _, item := range fixture.DenialCases {
		want, ok := wantReasons[item.Name]
		if !ok {
			t.Fatalf("unexpected closure denial case: %#v", item)
		}
		if item.ClosureAllowed || item.Reason != want {
			t.Fatalf("closure denial case %q should deny closure with %q: %#v", item.Name, want, item)
		}
		switch item.Name {
		case "missing_evidence":
			if !item.MissingEvidence {
				t.Fatalf("missing evidence case must mark missing evidence: %#v", item)
			}
		case "ci_pending":
			if !item.CIPending {
				t.Fatalf("CI pending case must mark CI pending: %#v", item)
			}
		case "codex_branch_remaining":
			if !item.CodexBranchFound {
				t.Fatalf("branch cleanup case must mark codex branch found: %#v", item)
			}
		case "forbidden_surface":
			if !item.ForbiddenSurface {
				t.Fatalf("public-safety case must mark forbidden surface: %#v", item)
			}
		}
		delete(wantReasons, item.Name)
	}
	if len(wantReasons) != 0 {
		t.Fatalf("missing closure denial cases: %#v", wantReasons)
	}

	nodeThirtySevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtySevenReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtySevenReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtySevenReadback.CompletedNodes != 37 ||
		nodeThirtySevenReadback.ReadyNodes != 3 ||
		nodeThirtySevenReadback.FirstExecutableNode != "mission-recommendation-hardening-38" ||
		nodeThirtySevenReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtySevenReadback.ExactNextAction, "mission-recommendation-hardening-38") {
		t.Fatalf("node 37 readback must carry closure-gate evidence and continue to node 38: %#v", nodeThirtySevenReadback)
	}
}

func TestLongRunHardeningWaveMissionAtlasMultiImportSmokeArtifact(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtySevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-37", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-38")
	fixture := mustLoadJSON[struct {
		Schema                  string `json:"schema"`
		NodeID                  string `json:"node_id"`
		Status                  string `json:"status"`
		Supervisor              string `json:"supervisor"`
		AtlasOwner              string `json:"atlas_owner"`
		CompletedNodesBefore    int    `json:"completed_nodes_before_node"`
		ReadyNodesBefore        int    `json:"ready_nodes_before_node"`
		FirstExecutableNode     string `json:"first_executable_node"`
		FinalResponseAllowed    bool   `json:"final_response_allowed"`
		ExactNextAction         string `json:"exact_next_action"`
		OneExecutableNodeActive bool   `json:"one_executable_node_active"`
		SerializedImportsOnly   bool   `json:"serialized_imports_only"`
		Imports                 []struct {
			NodeID         string `json:"node_id"`
			FoundryImport  string `json:"foundry_import"`
			RunLink        string `json:"run_link"`
			Readback       string `json:"readback"`
			ExactNextNode  string `json:"exact_next_node"`
			ImportComplete bool   `json:"import_complete"`
		} `json:"imports"`
		SchedulesWork          bool `json:"schedules_work"`
		ExecutesWork           bool `json:"executes_work"`
		ApprovesWork           bool `json:"approves_work"`
		ClaimsAuthorityAdvance bool `json:"claims_authority_advance"`
		RSIRemainsDenied       bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "mission-atlas-multi-import-smoke-fixture.json"))

	if fixture.Schema != "ao.atlas.mission-atlas-multi-import-smoke-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-38" ||
		fixture.Status != "multi_import_smoke_recorded" ||
		fixture.Supervisor != "ao-mission" ||
		fixture.AtlasOwner != "ao-atlas" ||
		fixture.CompletedNodesBefore != nodeThirtySevenReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtySevenReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtySevenReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtySevenReadback.ExactNextAction ||
		!fixture.OneExecutableNodeActive ||
		!fixture.SerializedImportsOnly ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("multi-import smoke fixture must bind node 37 checkpoint without authority effects: %#v", fixture)
	}
	if len(fixture.Imports) < 3 {
		t.Fatalf("multi-import smoke fixture must reference at least three serialized imports: %#v", fixture.Imports)
	}
	seenNodes := map[string]bool{}
	for _, item := range fixture.Imports {
		if seenNodes[item.NodeID] {
			t.Fatalf("multi-import smoke fixture repeats node %s", item.NodeID)
		}
		seenNodes[item.NodeID] = true
		if !item.ImportComplete || item.FoundryImport == "" || item.RunLink == "" || item.Readback == "" || item.ExactNextNode == "" {
			t.Fatalf("multi-import smoke item must bind import, run-link, readback, and next node: %#v", item)
		}
		for _, path := range []string{item.FoundryImport, item.RunLink, item.Readback} {
			if !strings.HasPrefix(path, "docs/") {
				t.Fatalf("multi-import smoke artifact path must stay public-safe: %s", path)
			}
			if _, err := os.Stat(filepath.Join(repoRoot(t), path)); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, nodeID := range []string{
		"mission-recommendation-hardening-36",
		"mission-recommendation-hardening-37",
		"mission-recommendation-hardening-38",
	} {
		if !seenNodes[nodeID] {
			t.Fatalf("multi-import smoke fixture missing node %s", nodeID)
		}
	}

	nodeThirtyEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyEightReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyEightReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyEightReadback.CompletedNodes != 38 ||
		nodeThirtyEightReadback.ReadyNodes != 2 ||
		nodeThirtyEightReadback.FirstExecutableNode != "mission-recommendation-hardening-39" ||
		nodeThirtyEightReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyEightReadback.ExactNextAction, "mission-recommendation-hardening-39") {
		t.Fatalf("node 38 readback must carry multi-import smoke evidence and continue to node 39: %#v", nodeThirtyEightReadback)
	}
}

func TestLongRunHardeningWaveVerificationPublicSafetyRollupBindsChangedReadbacks(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-38", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-39")
	fixture := mustLoadJSON[struct {
		Schema                      string   `json:"schema"`
		NodeID                      string   `json:"node_id"`
		Status                      string   `json:"status"`
		CompletedNodesBefore        int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore            int      `json:"ready_nodes_before_node"`
		FirstExecutableNode         string   `json:"first_executable_node"`
		FinalResponseAllowed        bool     `json:"final_response_allowed"`
		ExactNextAction             string   `json:"exact_next_action"`
		ChangedDocsReadbacksChecked bool     `json:"changed_docs_readbacks_checked"`
		LocalVerificationPassed     bool     `json:"local_verification_passed"`
		PublicSafetyScanPassed      bool     `json:"public_safety_scan_passed"`
		ForbiddenClaimMatches       int      `json:"forbidden_claim_matches"`
		VerificationCommands        []string `json:"verification_commands"`
		ChangedDocsAndReadbackPaths []string `json:"changed_docs_and_readback_paths"`
		ScopedPublicSafetyScan      string   `json:"scoped_public_safety_scan"`
		SchedulesWork               bool     `json:"schedules_work"`
		ExecutesWork                bool     `json:"executes_work"`
		ApprovesWork                bool     `json:"approves_work"`
		ClaimsAuthorityAdvance      bool     `json:"claims_authority_advance"`
		RSIRemainsDenied            bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "verification-public-safety-rollup-fixture.json"))

	if fixture.Schema != "ao.atlas.verification-public-safety-rollup-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-39" ||
		fixture.Status != "verification_public_safety_rollup_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtyEightReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyEightReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyEightReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowed ||
		fixture.ExactNextAction != nodeThirtyEightReadback.ExactNextAction ||
		!fixture.ChangedDocsReadbacksChecked ||
		!fixture.LocalVerificationPassed ||
		!fixture.PublicSafetyScanPassed ||
		fixture.ForbiddenClaimMatches != 0 ||
		fixture.ScopedPublicSafetyScan == "" ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("verification rollup fixture must bind node 38 checkpoint without authority effects: %#v", fixture)
	}
	for _, command := range []string{
		"go test ./... -count=1",
		"go vet ./...",
		"go build ./cmd/atlas",
		"scripts/atlas-foundry-roundtrip-smoke.sh",
		"scripts/production-readiness.sh",
		"git diff --check",
	} {
		if !containsString(fixture.VerificationCommands, command) {
			t.Fatalf("verification rollup missing command %q: %#v", command, fixture.VerificationCommands)
		}
	}
	if len(fixture.ChangedDocsAndReadbackPaths) < 4 {
		t.Fatalf("verification rollup must include changed docs and readbacks: %#v", fixture.ChangedDocsAndReadbackPaths)
	}
	for _, path := range fixture.ChangedDocsAndReadbackPaths {
		if !strings.HasPrefix(path, "docs/") && !strings.HasPrefix(path, "internal/") {
			t.Fatalf("verification rollup path must stay scoped to docs or tests: %s", path)
		}
		if _, err := os.Stat(filepath.Join(repoRoot(t), path)); err != nil {
			t.Fatal(err)
		}
	}

	nodeThirtyNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeThirtyNineReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeThirtyNineReadback.FeatureDepthRecommendations) < 40 ||
		nodeThirtyNineReadback.CompletedNodes != 39 ||
		nodeThirtyNineReadback.ReadyNodes != 1 ||
		nodeThirtyNineReadback.FirstExecutableNode != "mission-recommendation-hardening-40" ||
		nodeThirtyNineReadback.FinalResponseAllowed ||
		!strings.Contains(nodeThirtyNineReadback.ExactNextAction, "mission-recommendation-hardening-40") {
		t.Fatalf("node 39 readback must carry verification rollup evidence and continue to node 40: %#v", nodeThirtyNineReadback)
	}
}

func TestLongRunHardeningWaveFinalClosureArtifactsAllowFinalResponseOnlyAtCompletion(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeThirtyNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-hardening-39", "recommendation-readback-after.json"))
	nodeDir := filepath.Join(root, "nodes", "mission-recommendation-hardening-40")
	fixture := mustLoadJSON[struct {
		Schema                     string   `json:"schema"`
		NodeID                     string   `json:"node_id"`
		Status                     string   `json:"status"`
		CompletedNodesBefore       int      `json:"completed_nodes_before_node"`
		ReadyNodesBefore           int      `json:"ready_nodes_before_node"`
		FirstExecutableNode        string   `json:"first_executable_node"`
		FinalResponseAllowedBefore bool     `json:"final_response_allowed_before_node"`
		ExactNextActionBefore      string   `json:"exact_next_action_before_node"`
		ClosureArtifactPaths       []string `json:"closure_artifact_paths"`
		CleanRepoStatusPath        string   `json:"clean_repo_status_path"`
		VerificationSummaryPath    string   `json:"verification_summary_path"`
		PromoterNoPromotionStatus  string   `json:"promoter_no_promotion_status"`
		CommandReadbackStatus      string   `json:"command_readback_status"`
		FinalResponseAllowedAfter  bool     `json:"final_response_allowed_after_node"`
		ReadyNodesAfter            int      `json:"ready_nodes_after_node"`
		BlockedNodesAfter          int      `json:"blocked_nodes_after_node"`
		SchedulesWork              bool     `json:"schedules_work"`
		ExecutesWork               bool     `json:"executes_work"`
		ApprovesWork               bool     `json:"approves_work"`
		ClaimsAuthorityAdvance     bool     `json:"claims_authority_advance"`
		RSIRemainsDenied           bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeDir, "final-closure-artifacts-fixture.json"))

	if fixture.Schema != "ao.atlas.final-closure-artifacts-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-hardening-40" ||
		fixture.Status != "final_closure_artifacts_recorded" ||
		fixture.CompletedNodesBefore != nodeThirtyNineReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeThirtyNineReadback.ReadyNodes ||
		fixture.FirstExecutableNode != nodeThirtyNineReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowedBefore ||
		fixture.ExactNextActionBefore != nodeThirtyNineReadback.ExactNextAction ||
		fixture.PromoterNoPromotionStatus != "no_promotion_requested" ||
		fixture.CommandReadbackStatus != "readback_agrees_no_promotion" ||
		!fixture.FinalResponseAllowedAfter ||
		fixture.ReadyNodesAfter != 0 ||
		fixture.BlockedNodesAfter != 0 ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("final closure fixture must bind node 39 checkpoint and final no-promotion closure: %#v", fixture)
	}
	if !containsString(fixture.ClosureArtifactPaths, fixture.CleanRepoStatusPath) ||
		!containsString(fixture.ClosureArtifactPaths, fixture.VerificationSummaryPath) {
		t.Fatalf("final closure fixture must include clean repo and verification summary artifacts: %#v", fixture.ClosureArtifactPaths)
	}
	for _, path := range fixture.ClosureArtifactPaths {
		if !strings.HasPrefix(path, "docs/") {
			t.Fatalf("final closure artifact path must stay under docs/: %s", path)
		}
		if _, err := os.Stat(filepath.Join(repoRoot(t), path)); err != nil {
			t.Fatal(err)
		}
	}

	nodeFortyReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeDir, "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(nodeFortyReadback); err != nil {
		t.Fatal(err)
	}
	if len(nodeFortyReadback.FeatureDepthRecommendations) < 40 ||
		nodeFortyReadback.CompletedNodes != 40 ||
		nodeFortyReadback.ReadyNodes != 0 ||
		nodeFortyReadback.BlockedNodes != 0 ||
		nodeFortyReadback.FirstExecutableNode != "" ||
		!nodeFortyReadback.FinalResponseAllowed {
		t.Fatalf("node 40 readback must close the wave and allow final response: %#v", nodeFortyReadback)
	}
}
