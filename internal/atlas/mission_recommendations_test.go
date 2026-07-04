package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsImportBuildsDoubleSizeWaveAndWorkgraph(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 20, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--min-tasks", "20",
		"--node-budget", "20",
		"--estimated-minutes", "90",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "recommendation_tasks=20") ||
		!strings.Contains(out.String(), "estimated_minutes=90") {
		t.Fatalf("import output missing long-run counts: %s", out.String())
	}

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(outDir, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "mission-long-wave" || wave.TotalTasks != 20 || wave.EstimatedMinutes != 90 || wave.NodeBudget != 20 {
		t.Fatalf("bad recommendation wave summary: %#v", wave)
	}
	if wave.SafeToExecute || wave.SchedulesWork || wave.ExecutesWork || wave.ApprovesWork {
		t.Fatalf("recommendation wave widened authority: %#v", wave)
	}
	if !strings.Contains(wave.NextRecommendedPrompt, "at least 20 bounded Atlas nodes") ||
		!strings.Contains(wave.NextRecommendedPrompt, "Return only after") {
		t.Fatalf("wave missing continuation prompt: %q", wave.NextRecommendedPrompt)
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(outDir, "recommendation-workgraph.json"))
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 20 {
		t.Fatalf("expected 20 recommendation nodes, got %d", len(workgraph.Nodes))
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ExecutableReadyNodeIDs) != 1 || state.ExecutableReadyNodeIDs[0] != "mission-recommendation-next-01" {
		t.Fatalf("expected exactly one executable-ready node, got %#v", state.ExecutableReadyNodeIDs)
	}
	for _, node := range workgraph.Nodes {
		if node.FactoryTask.TargetFactoryRepo != "ao-atlas" {
			t.Fatalf("recommendation task should be Atlas-owned: %+v", node.FactoryTask)
		}
		for _, want := range []string{"node_gate", "candidate_record", "rollback_record", "tests", "verification"} {
			if !containsString(node.FactoryTask.RequiredGates, want) {
				t.Fatalf("task %s missing required gate %q: %#v", node.FactoryTask.ID, want, node.FactoryTask.RequiredGates)
			}
		}
		if !containsString(node.FactoryTask.SafetyLimits, "no provider calls") ||
			!containsString(node.FactoryTask.SafetyLimits, "no credential inspection") ||
			!containsString(node.FactoryTask.SafetyLimits, "no direct main mutation") {
			t.Fatalf("task %s missing safety limits: %#v", node.FactoryTask.ID, node.FactoryTask.SafetyLimits)
		}
	}
	prompt, err := os.ReadFile(filepath.Join(outDir, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prompt), "You are AO Atlas") ||
		!strings.Contains(string(prompt), "Double the previous short batch") {
		t.Fatalf("next prompt missing operator-ready continuation text:\n%s", string(prompt))
	}
}

func TestMissionRecommendationsRejectShallowAndUnsafeBundles(t *testing.T) {
	dir := t.TempDir()
	shallowPath := filepath.Join(dir, "shallow.json")
	unsafePath := filepath.Join(dir, "unsafe.json")
	writeFeatureDepthBundle(t, shallowPath, 3, false)
	writeFeatureDepthBundle(t, unsafePath, 20, true)

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "shallow", path: shallowPath, want: "at least 20 tasks"},
		{name: "unsafe", path: unsafePath, want: "safe_to_execute must be false"},
	} {
		var out bytes.Buffer
		code := Run([]string{
			"mission", "recommendations", "import",
			"--recommendations", tc.path,
			"--target-instance", "demo-stack",
			"--min-tasks", "20",
			"--node-budget", "20",
			"--estimated-minutes", "90",
			"--out", filepath.Join(dir, tc.name+"-out"),
		}, &out, &out)
		if code == 0 {
			t.Fatalf("%s bundle was accepted", tc.name)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatalf("%s error missing %q: %s", tc.name, tc.want, out.String())
		}
	}
}

func TestMissionRecommendationsDefaultToTwoToThreeHourSupervisorWave(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(outDir, "recommendation-wave.json"))
	if wave.TotalTasks != 40 || wave.NodeBudget != 40 {
		t.Fatalf("default wave should generate 40 nodes for continue-if-fast policy: %#v", wave)
	}
	if wave.MinimumTasks != 30 || wave.EstimatedMinutes != 120 {
		t.Fatalf("default wave should require 30 nodes and 120 minute floor: %#v", wave)
	}
	if wave.Supervisor == nil {
		t.Fatalf("default wave missing long-run supervisor: %#v", wave)
	}
	if wave.Supervisor.MinNodes != 30 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 ||
		wave.Supervisor.ReturnOnlyWhen != "all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker" ||
		wave.Supervisor.CheckpointPolicy != "after_each_node_or_timed_interval" ||
		wave.Supervisor.EvidencePolicy != "node_gate_candidate_rollback_tests_verification_public_safety_promoter_command" ||
		wave.Supervisor.FinalReportContract != "ao.atlas.long-run-final-report.v0.2" {
		t.Fatalf("bad long-run supervisor: %#v", wave.Supervisor)
	}
	if wave.FinalResponseAllowed || wave.FinalResponseReason != "ready nodes or exact next actions remain" {
		t.Fatalf("default wave must deny final response while ready nodes remain: %#v", wave)
	}
	if wave.PromoterReadbackStatus != "required_not_bound" || wave.CommandReadbackStatus != "required_not_bound" || wave.PublicSafetyScanStatus != "required_pending_verification" {
		t.Fatalf("wave should require promoter, command, and public-safety readbacks: %#v", wave)
	}
	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(outDir, "recommendation-workgraph.json"))
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 40 || len(state.ExecutableReadyNodeIDs) != 1 {
		t.Fatalf("expected 40 dependency-chained nodes with one executable-ready node, nodes=%d ready=%#v", len(workgraph.Nodes), state.ExecutableReadyNodeIDs)
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(outDir, "recommendation-readback.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.TotalNodes != 40 || readback.CompletedNodes != 0 || readback.ReadyNodes != 40 || readback.ExecutableReadyNodes != 1 {
		t.Fatalf("bad default readback node counts: %#v", readback)
	}
	if readback.LeaseHealthStatus != "minimum_unmet" ||
		readback.CheckpointFreshnessStatus != "fresh_checkpoint_required_after_each_node_or_timed_interval" ||
		readback.StaleRouteDecisionStatus != "fresh_atlas_supervises_foundry_owns_one_active_node" ||
		readback.EarlyReturnRiskStatus != "blocked_final_response_ready_nodes_remain" {
		t.Fatalf("readback missing long-run health statuses: %#v", readback)
	}
	if readback.FinalResponseAllowed || readback.ExactNextAction != "Emit Foundry import for mission-recommendation-next-01 and execute exactly one active node." {
		t.Fatalf("readback must deny final response with exact next node: %#v", readback)
	}
	if readback.FoundryTerminalStatusReadback["promoted"] == "" ||
		readback.FoundryTerminalStatusReadback["denied"] == "" ||
		readback.FoundryTerminalStatusReadback["blocked"] == "" ||
		readback.PromoterNoPromotionStatus == "" ||
		readback.CommandTimelineStatus == "" {
		t.Fatalf("readback missing terminal-state, promoter, or command summaries: %#v", readback)
	}
	if len(readback.NodeEvidence) != 40 || readback.NodeEvidence[0].NodeGate != "recorded" || readback.NodeEvidence[0].RollbackRecord != "recorded" {
		t.Fatalf("readback missing per-node evidence: %#v", readback.NodeEvidence[:1])
	}
	if len(readback.FeatureDepthRecommendations) < 10 {
		t.Fatalf("readback must carry at least 10 next recommendations: %#v", readback.FeatureDepthRecommendations)
	}
	prompt := wave.NextRecommendedPrompt
	for _, want := range []string{
		"Current state:",
		"Problem:",
		"Goal:",
		"Minimum work budget:",
		"Safety boundaries:",
		"Required work:",
		"Per-node requirements:",
		"Regression tests:",
		"Verification:",
		"Final response only after completion or true hard blocker:",
		"Target 2-3 hours",
		"Complete at least 30 bounded implementation/evidence nodes",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v0.2 prompt missing section %q:\n%s", want, prompt)
		}
	}
}

func TestMissionRecommendationsReadbackCLIMatchesGeneratedArtifacts(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	importDir := filepath.Join(dir, "recommendations-out")
	readbackPath := filepath.Join(dir, "readback.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", importDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	out.Reset()
	code = Run([]string{
		"mission", "recommendations", "readback",
		"--wave", filepath.Join(importDir, "recommendation-wave.json"),
		"--workgraph", filepath.Join(importDir, "recommendation-workgraph.json"),
		"--out", readbackPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation readback failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "final_response_allowed=false") ||
		!strings.Contains(out.String(), "exact_next_action=Emit Foundry import for mission-recommendation-next-01 and execute exactly one active node.") {
		t.Fatalf("readback output missing final gate and exact action: %s", out.String())
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)
	if readback.WaveDigest == "" || readback.WorkgraphDigest == "" {
		t.Fatalf("readback must carry artifact digests: %#v", readback)
	}
}

func TestMissionRecommendationsReadbackFinalGateTransitions(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "demo-stack",
	})
	if err != nil {
		t.Fatal(err)
	}

	partial := completeRecommendationNodes(result.Workgraph, 30)
	partialReadback, err := BuildAtlasRecommendationReadback(result.Wave, partial, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if partialReadback.LeaseHealthStatus != "minimum_met_continue_if_fast" ||
		partialReadback.FinalResponseAllowed ||
		partialReadback.ExactNextAction != "Emit Foundry import for mission-recommendation-next-31 and execute exactly one active node." {
		t.Fatalf("partial readback must continue after minimum while ready nodes remain: %#v", partialReadback)
	}

	completed := completeRecommendationNodes(result.Workgraph, 40)
	completedReadback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !completedReadback.FinalResponseAllowed ||
		completedReadback.FinalResponseReason != "all generated nodes complete and no ready nodes remain" ||
		completedReadback.LeaseHealthStatus != "all_generated_nodes_complete" ||
		completedReadback.ExactNextAction != "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks." {
		t.Fatalf("completed readback must allow closure: %#v", completedReadback)
	}
	if completedReadback.FoundryRollupStatus != "completed_all_node_run_links_recorded" ||
		completedReadback.PromoterReadbackStatus != "no_promotion_recorded" ||
		completedReadback.PromoterNoPromotionStatus != "recorded_no_promotion_for_recommendation_wave" ||
		completedReadback.CommandReadbackStatus != "compact_timeline_recorded" ||
		completedReadback.CommandTimelineStatus != "recorded_compact_timeline_for_completed_wave" {
		t.Fatalf("completed readback missing closure bindings: %#v", completedReadback)
	}
}

func TestRecommendationCompleteNodeRejectsMissingGateEvidence(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "demo-stack",
	})
	if err != nil {
		t.Fatal(err)
	}
	node := result.Workgraph.Nodes[0]
	link := recommendationRunLink(t, node.FactoryTask.ID, map[string]string{
		"node_gate": recommendationEvidenceFile(t, "missing-gate", node.ID, "node-gate.json"),
	})

	_, _, err = CompleteAtlasRecommendationNodeWithRunLink(result.Wave, result.Workgraph, link, AtlasRecommendationCompleteNodeOptions{
		ExpectedNodeID: node.ID,
		EvidenceRoot:   repoRoot(t),
	})
	if err == nil || !strings.Contains(err.Error(), "missing evidence candidate_record") {
		t.Fatalf("expected missing candidate record evidence, got %v", err)
	}
}

func TestRecommendationCompleteNodeRejectsOutOfOrderRunLink(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "demo-stack",
	})
	if err != nil {
		t.Fatal(err)
	}
	firstNode := result.Workgraph.Nodes[0]
	secondNode := result.Workgraph.Nodes[1]
	link := recommendationRunLink(t, secondNode.FactoryTask.ID, recommendationEvidenceFiles(t, "out-of-order", secondNode.ID))

	_, _, err = CompleteAtlasRecommendationNodeWithRunLink(result.Wave, result.Workgraph, link, AtlasRecommendationCompleteNodeOptions{
		ExpectedNodeID: firstNode.ID,
		EvidenceRoot:   repoRoot(t),
	})
	if err == nil || !strings.Contains(err.Error(), "run-link task_id must match executable node") {
		t.Fatalf("expected out-of-order rejection, got %v", err)
	}
}

func TestRecommendationCompleteNodeAdvancesReadbackAndExecutionLedger(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "demo-stack",
	})
	if err != nil {
		t.Fatal(err)
	}
	firstNode := result.Workgraph.Nodes[0]
	link := recommendationRunLink(t, firstNode.FactoryTask.ID, recommendationEvidenceFiles(t, "advance", firstNode.ID))

	updated, completedNodeID, err := CompleteAtlasRecommendationNodeWithRunLink(result.Wave, result.Workgraph, link, AtlasRecommendationCompleteNodeOptions{
		ExpectedNodeID: firstNode.ID,
		EvidenceRoot:   repoRoot(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	if completedNodeID != "mission-recommendation-next-01" {
		t.Fatalf("completed wrong node: %s", completedNodeID)
	}
	readback, err := BuildAtlasRecommendationReadback(result.Wave, updated, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 1 ||
		readback.ReadyNodes != 39 ||
		readback.FirstExecutableNode != "mission-recommendation-next-02" ||
		readback.FinalResponseAllowed {
		t.Fatalf("bad readback after first completion: %#v", readback)
	}
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("execution ledger should match readback: %v", err)
	}
	if execution.CompletedRecommendationNodes != 1 || execution.GeneratedWorkgraph.ExecutableReadyNodes != 1 {
		t.Fatalf("bad execution ledger after first completion: %#v", execution)
	}
}

func TestMissionRecommendationsCompleteNodeCLIWritesUpdatedArtifacts(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	importDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", importDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(importDir, "recommendation-workgraph.json"))
	firstNode := workgraph.Nodes[0]
	link := recommendationRunLink(t, firstNode.FactoryTask.ID, recommendationEvidenceFiles(t, "cli", firstNode.ID))
	linkPath := filepath.Join(dir, "run-link.json")
	if err := WriteJSON(linkPath, link); err != nil {
		t.Fatal(err)
	}
	updatedWorkgraphPath := filepath.Join(dir, "updated-workgraph.json")
	readbackPath := filepath.Join(dir, "updated-readback.json")
	executionPath := filepath.Join(dir, "updated-execution-readback.json")

	out.Reset()
	code = Run([]string{
		"mission", "recommendations", "complete-node",
		"--wave", filepath.Join(importDir, "recommendation-wave.json"),
		"--workgraph", filepath.Join(importDir, "recommendation-workgraph.json"),
		"--run-link", linkPath,
		"--expected-node", firstNode.ID,
		"--evidence-root", repoRoot(t),
		"--readback-evidence-root", "docs/evidence/test-wave",
		"--out-workgraph", updatedWorkgraphPath,
		"--out-readback", readbackPath,
		"--out-execution-readback", executionPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("complete-node failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "completed_nodes=1") ||
		!strings.Contains(out.String(), "next_executable_node=mission-recommendation-next-02") {
		t.Fatalf("complete-node output missing progress readback: %s", out.String())
	}
	updated := mustLoadJSON[Workgraph](t, updatedWorkgraphPath)
	if updated.Nodes[0].Status != "completed" || updated.Nodes[1].Status != "ready" {
		t.Fatalf("bad updated workgraph statuses: %#v", updated.Nodes[:2])
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)
	if readback.CompletedNodes != 1 || readback.EvidenceRoot != "docs/evidence/test-wave" {
		t.Fatalf("bad readback artifact: %#v", readback)
	}
	execution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, executionPath)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("bad execution artifact: %v", err)
	}
}

func TestMissionRecommendationsReadbackRejectsMismatchedWaveAndWorkgraph(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "demo-stack",
	})
	if err != nil {
		t.Fatal(err)
	}
	result.Workgraph.TargetInstance = "other-stack"
	if _, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{}); err == nil || !strings.Contains(err.Error(), "target_instance") {
		t.Fatalf("expected target_instance mismatch rejection, got %v", err)
	}
}

func TestRecommendationExecutionReadbackRejectsFalseCompletedNodes(t *testing.T) {
	readback := AtlasRecommendationReadback{
		ContractVersion:             AtlasRecommendationReadbackContract,
		MissionID:                   "mission-long-wave",
		TargetInstance:              "demo-stack",
		Status:                      "ready",
		SourceDigest:                "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TotalNodes:                  40,
		MinimumNodes:                30,
		CompletedNodes:              0,
		ReadyNodes:                  40,
		ExecutableReadyNodes:        1,
		LeaseHealthStatus:           "minimum_unmet",
		CheckpointFreshnessStatus:   "fresh_checkpoint_required_after_each_node_or_timed_interval",
		StaleRouteDecisionStatus:    "fresh_atlas_supervises_foundry_owns_one_active_node",
		EarlyReturnRiskStatus:       "blocked_final_response_ready_nodes_remain",
		FoundryRollupStatus:         "required_pending_first_node_import",
		PromoterReadbackStatus:      "required_not_bound",
		CommandReadbackStatus:       "required_not_bound",
		PublicSafetyScanStatus:      "required_pending_verification",
		FinalResponseReason:         "ready nodes or exact next actions remain",
		ExactNextAction:             "Emit Foundry import for mission-recommendation-next-01 and execute exactly one active node.",
		NodeEvidence:                []AtlasRecommendationNodeEvidence{{NodeID: "mission-recommendation-next-01", TaskID: "mission-recommendation-next-01-task", Status: "ready", NodeGate: "recorded", CandidateRecord: "recorded", RollbackRecord: "recorded", ImplementationEvidence: "recorded", Tests: "recorded", Verification: "recorded", PublicSafetyWording: "recorded", PromoterReadback: "recorded", CommandReadback: "recorded", RequiredGates: []string{"node_gate"}, VerificationCommands: []string{"go test ./... -count=1"}}},
		FeatureDepthRecommendations: []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"},
		SafetyBoundaries:            map[string]bool{"provider_calls": false},
	}
	execution := AtlasRecommendationExecutionReadback{
		Schema:                       "ao.atlas.long-recommendation-wave-execution.v0.3",
		Status:                       "completed",
		MissionID:                    "mission-long-wave",
		CompletedRecommendationNodes: 40,
		TotalRecommendationNodes:     40,
		GeneratedWorkgraph: AtlasRecommendationGeneratedWorkgraphReadback{
			TotalNodes:           40,
			ReadyNodes:           40,
			ExecutableReadyNodes: 1,
			FinalResponseAllowed: false,
		},
	}

	err := ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "completed_recommendation_nodes must match recommendation readback completed_nodes") {
		t.Fatalf("expected false completion rejection, got %v", err)
	}
}

func TestRecommendationExecutionReadbackArtifactsStayConsistent(t *testing.T) {
	root := repoRoot(t)
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "docs", "evidence", "ao-atlas-long-recommendation-wave-v03", "recommendation-readback.json"))
	execution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, filepath.Join(root, "docs", "evidence", "ao-atlas-long-recommendation-wave-v03", "execution-readback.json"))
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("v0.3 execution ledger is inconsistent with recommendation readback: %v", err)
	}
}

func TestMissionRecommendationsRejectMixedOwnerDefaultWaveWithExactReadback(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	var bundle AOMissionFeatureDepthRecommendations
	if err := readJSONIfPossible(recommendationsPath, &bundle); err != nil {
		t.Fatal(err)
	}
	bundle.Tasks[39].Owner = "ao-foundry"
	if err := WriteJSON(recommendationsPath, bundle); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", filepath.Join(dir, "out"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("mixed-owner default wave was accepted")
	}
	if !strings.Contains(out.String(), "requires at least 30 AO Atlas-owned tasks and 40 tasks for continue-if-fast target") {
		t.Fatalf("mixed-owner error did not report exact readback: %s", out.String())
	}
}

func TestProductionReadinessExercisesMissionRecommendationsImport(t *testing.T) {
	root := repoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"mission recommendations import",
		"--recommendations examples/valid/ao-mission/feature-depth-recommendations.json",
		"--min-tasks 30",
		"--min-minutes 120",
		"--max-minutes 180",
		"--continue-if-fast-target 40",
		"recommendation-workgraph.json",
		"recommendation-readback.json",
		"mission recommendations readback",
		"mission recommendations complete-node",
		"--out-execution-readback",
		"completed_recommendation_nodes",
		"recommendation-ledger-consistency",
		"next-recommended-prompt.md",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing recommendation coverage %q", want)
		}
	}
}

func writeFeatureDepthBundle(t *testing.T, path string, taskCount int, unsafe bool) {
	t.Helper()
	tasks := make([]map[string]string, 0, taskCount)
	for i := 1; i <= taskCount; i++ {
		tasks = append(tasks, map[string]string{
			"id":    "next-" + twoDigit(i),
			"owner": "ao-atlas",
			"task":  "Implement Atlas long-run recommendation node " + twoDigit(i) + " with tests, readback evidence, and continuation prompt coverage.",
		})
	}
	if err := WriteJSON(path, map[string]any{
		"schema":               "ao.mission.feature-depth-recommendations.v0.3",
		"mission_id":           "mission-long-wave",
		"status":               "ready",
		"minimum_tasks":        taskCount,
		"recommendation_count": taskCount,
		"tasks":                tasks,
		"safe_to_execute":      unsafe,
		"executes_work":        false,
		"approves_work":        false,
	}); err != nil {
		t.Fatal(err)
	}
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return "10"[:0] + string(rune('0'+value/10)) + string(rune('0'+value%10))
}

func completeRecommendationNodes(workgraph Workgraph, count int) Workgraph {
	updated := workgraph
	updated.Nodes = append([]WorkgraphNode(nil), workgraph.Nodes...)
	for i := range updated.Nodes {
		if i < count {
			updated.Nodes[i].Status = "completed"
		}
	}
	return updated
}

func recommendationRunLink(t *testing.T, taskID string, evidence map[string]string) RunLink {
	t.Helper()
	link, err := BuildRunLink(taskID, "completed", evidence)
	if err != nil {
		t.Fatal(err)
	}
	return link
}

func recommendationEvidenceFiles(t *testing.T, scenario, nodeID string) map[string]string {
	t.Helper()
	keys := []string{
		"node_gate",
		"candidate_record",
		"rollback_record",
		"implementation_evidence",
		"tests",
		"verification",
		"sentinel_public_safety",
		"promoter_no_promotion",
		"command_readback",
		"foundry_import",
		"checkpoint_bundle",
	}
	evidence := map[string]string{}
	for _, key := range keys {
		evidence[key] = recommendationEvidenceFile(t, scenario, nodeID, key+".json")
	}
	return evidence
}

func recommendationEvidenceFile(t *testing.T, scenario, nodeID, name string) string {
	t.Helper()
	rel := filepath.ToSlash(filepath.Join("target", "recommendation-node-evidence-test", scenario, nodeID, name))
	abs := filepath.Join(repoRoot(t), rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(`{"status":"recorded"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return rel
}
