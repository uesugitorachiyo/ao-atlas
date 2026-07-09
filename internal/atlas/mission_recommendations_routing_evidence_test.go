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
	rawWave := mustLoadJSON[map[string]any](t, filepath.Join(outDir, "recommendation-wave.json"))
	rawTasks, ok := rawWave["tasks"].([]any)
	if !ok || len(rawTasks) != 20 {
		t.Fatalf("recommendation wave missing raw tasks: %#v", rawWave["tasks"])
	}
	for i, raw := range rawTasks {
		task, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("task %d is not an object: %#v", i, raw)
		}
		digest, ok := task["source_task_digest"].(string)
		if !ok || !digestPattern.MatchString(digest) {
			t.Fatalf("task %d missing source_task_digest: %#v", i, task)
		}
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
		if !containsStringPrefix(node.FactoryTask.RequiredEvidence, "source_task_digest:sha256:") {
			t.Fatalf("task %s missing digest-bound source evidence: %#v", node.FactoryTask.ID, node.FactoryTask.RequiredEvidence)
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

func TestMissionRecommendationsRejectShallowFeatureDepthVariants(t *testing.T) {
	dir := t.TempDir()

	lowMinimumPath := filepath.Join(dir, "low-minimum.json")
	writeFeatureDepthBundle(t, lowMinimumPath, 20, false)
	lowMinimum := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, lowMinimumPath)
	lowMinimum.MinimumTasks = 3
	if err := WriteJSON(lowMinimumPath, lowMinimum); err != nil {
		t.Fatal(err)
	}

	tooFewTasksPath := filepath.Join(dir, "too-few-tasks.json")
	writeFeatureDepthBundle(t, tooFewTasksPath, 3, false)
	tooFewTasks := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, tooFewTasksPath)
	tooFewTasks.MinimumTasks = 20
	if err := WriteJSON(tooFewTasksPath, tooFewTasks); err != nil {
		t.Fatal(err)
	}

	filteredPath := filepath.Join(dir, "owner-filtered.json")
	writeFeatureDepthBundle(t, filteredPath, 20, false)
	filtered := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, filteredPath)
	filtered.Tasks[19].Owner = "ao-foundry"
	if err := WriteJSON(filteredPath, filtered); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "low minimum_tasks", path: lowMinimumPath, want: "minimum_tasks must be at least 20"},
		{name: "too few task entries", path: tooFewTasksPath, want: "tasks must include at least 20 tasks"},
		{name: "too few Atlas-owned tasks", path: filteredPath, want: "requires at least 20 tasks"},
	} {
		var out bytes.Buffer
		code := Run([]string{
			"mission", "recommendations", "import",
			"--recommendations", tc.path,
			"--target-instance", "demo-stack",
			"--min-tasks", "20",
			"--node-budget", "20",
			"--continue-if-fast-target", "20",
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

func TestMissionRecommendationsRejectUnsafeFeatureDepthAuthorityClaims(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		name   string
		mutate func(*AOMissionFeatureDepthRecommendations)
		want   string
	}{
		{
			name: "safe_to_execute",
			mutate: func(bundle *AOMissionFeatureDepthRecommendations) {
				bundle.SafeToExecute = true
			},
			want: "safe_to_execute must be false",
		},
		{
			name: "schedules_work",
			mutate: func(bundle *AOMissionFeatureDepthRecommendations) {
				bundle.SchedulesWork = true
			},
			want: "schedules_work must be false",
		},
		{
			name: "executes_work",
			mutate: func(bundle *AOMissionFeatureDepthRecommendations) {
				bundle.ExecutesWork = true
			},
			want: "executes_work must be false",
		},
		{
			name: "approves_work",
			mutate: func(bundle *AOMissionFeatureDepthRecommendations) {
				bundle.ApprovesWork = true
			},
			want: "approves_work must be false",
		},
		{
			name: "mutates_repositories",
			mutate: func(bundle *AOMissionFeatureDepthRecommendations) {
				bundle.MutatesRepositories = true
			},
			want: "mutates_repositories must be false",
		},
	} {
		path := filepath.Join(dir, tc.name+".json")
		writeFeatureDepthBundle(t, path, 20, false)
		bundle := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, path)
		tc.mutate(&bundle)
		if err := WriteJSON(path, bundle); err != nil {
			t.Fatal(err)
		}

		var out bytes.Buffer
		code := Run([]string{
			"mission", "recommendations", "import",
			"--recommendations", path,
			"--target-instance", "demo-stack",
			"--min-tasks", "20",
			"--node-budget", "20",
			"--estimated-minutes", "90",
			"--out", filepath.Join(dir, tc.name+"-out"),
		}, &out, &out)
		if code == 0 {
			t.Fatalf("%s authority claim was accepted", tc.name)
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
	readinessPacket := mustLoadJSON[AtlasRecommendationWorkgraphReadinessPacket](t, filepath.Join(outDir, "workgraph-readiness-packet.json"))
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(readinessPacket, readback); err != nil {
		t.Fatal(err)
	}
	if readback.TotalNodes != 40 || readback.CompletedNodes != 0 || readback.ReadyNodes != 40 || readback.ExecutableReadyNodes != 1 {
		t.Fatalf("bad default readback node counts: %#v", readback)
	}
	if readinessPacket.Status != "continuation_required" ||
		readinessPacket.TotalNodes != 40 ||
		readinessPacket.MinimumNodes != 30 ||
		readinessPacket.NodeBudget != 40 ||
		readinessPacket.ContinueIfFastTarget != 40 ||
		readinessPacket.ReadyNodes != 40 ||
		readinessPacket.ExecutableReadyNodes != 1 ||
		readinessPacket.FirstExecutableNode != "mission-recommendation-next-01" ||
		readinessPacket.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		readinessPacket.ContinuationBudgetStatus != "minimum_nodes_unmet_continue_to_40_node_budget" ||
		!readinessPacket.OneExecutableMutationNodeActive ||
		!readinessPacket.RefusesFinalResponse ||
		readinessPacket.FinalResponseAllowed ||
		!strings.Contains(readinessPacket.ExactNextAction, readinessPacket.FirstExecutableNode) ||
		!readinessPacket.RSIRemainsDenied {
		t.Fatalf("readiness packet lost 40-node continuation budget: %#v", readinessPacket)
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
	if readback.ContinuationContract.ContractVersion != "ao.atlas.continuation-contract.v0.1" ||
		readback.ContinuationContract.Status != "continuation_required" ||
		!readback.ContinuationContract.RefusesFinalResponse ||
		readback.ContinuationContract.ReadyNodes != readback.ReadyNodes ||
		readback.ContinuationContract.ExactNextAction != readback.ExactNextAction ||
		readback.ContinuationContract.FinalResponseAllowed != readback.FinalResponseAllowed {
		t.Fatalf("readback missing Atlas continuation contract: %#v", readback.ContinuationContract)
	}
	rawReadback := mustLoadJSON[map[string]any](t, filepath.Join(outDir, "recommendation-readback.json"))
	continuationContract, ok := rawReadback["continuation_contract"].(map[string]any)
	if !ok ||
		continuationContract["status"] != "continuation_required" ||
		continuationContract["refuses_final_response"] != true ||
		continuationContract["ready_nodes"] != float64(readback.ReadyNodes) ||
		continuationContract["exact_next_action"] != readback.ExactNextAction {
		t.Fatalf("raw readback missing continuation contract: %#v", rawReadback["continuation_contract"])
	}
	exactNextActionReadback, ok := rawReadback["exact_next_action_readback"].(map[string]any)
	if !ok ||
		exactNextActionReadback["action"] != readback.ExactNextAction ||
		exactNextActionReadback["next_executable_node"] != readback.FirstExecutableNode ||
		exactNextActionReadback["return_gate_status"] != readback.ReturnGateStatus ||
		exactNextActionReadback["final_response_allowed"] != false ||
		exactNextActionReadback["source"] != "recommendation_readback" {
		t.Fatalf("readback missing structured exact next action binding: %#v", rawReadback["exact_next_action_readback"])
	}
	if rawReadback["final_response_denial_gate"] != "deny_ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("readback missing final response denial gate: %#v", rawReadback["final_response_denial_gate"])
	}
	timelinePlaceholders, ok := rawReadback["command_timeline_placeholders"].([]any)
	if !ok || len(timelinePlaceholders) < 3 {
		t.Fatalf("readback missing Command timeline placeholders: %#v", rawReadback["command_timeline_placeholders"])
	}
	timelineBySlot := map[string]map[string]any{}
	for _, item := range timelinePlaceholders {
		placeholder, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("bad Command timeline placeholder: %#v", item)
		}
		slot, _ := placeholder["slot"].(string)
		timelineBySlot[slot] = placeholder
	}
	if timelineBySlot["checkpoint"]["source"] != "recommendation_readback" ||
		timelineBySlot["exact_next_action"]["status"] != "pending_command_timeline" ||
		timelineBySlot["return_gate"]["required_before_final_response"] != true {
		t.Fatalf("Command timeline placeholders do not bind checkpoint/action/return gate: %#v", timelineBySlot)
	}
	promoterPlaceholders, ok := rawReadback["promoter_no_promotion_placeholders"].([]any)
	if !ok || len(promoterPlaceholders) < 3 {
		t.Fatalf("readback missing Promoter no-promotion placeholders: %#v", rawReadback["promoter_no_promotion_placeholders"])
	}
	promoterBySlot := map[string]map[string]any{}
	for _, item := range promoterPlaceholders {
		placeholder, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("bad Promoter no-promotion placeholder: %#v", item)
		}
		slot, _ := placeholder["slot"].(string)
		promoterBySlot[slot] = placeholder
	}
	if promoterBySlot["promotion_claim"]["source"] != "recommendation_readback" ||
		promoterBySlot["rsi_boundary"]["status"] != "pending_promoter_no_promotion" ||
		promoterBySlot["authority_advance"]["required_before_final_response"] != true {
		t.Fatalf("Promoter no-promotion placeholders do not bind promotion, RSI, and authority boundaries: %#v", promoterBySlot)
	}
	if readback.ReturnGateStatus != "blocked_ready_nodes_remain" || readback.CheckpointCount != 0 {
		t.Fatalf("readback missing return gate status or checkpoint count: %#v", readback)
	}
	if readback.FoundryTerminalStatusReadback["promoted"] == "" ||
		readback.FoundryTerminalStatusReadback["denied"] == "" ||
		readback.FoundryTerminalStatusReadback["blocked"] == "" ||
		readback.PromoterNoPromotionStatus == "" ||
		readback.CommandTimelineStatus == "" {
		t.Fatalf("readback missing terminal-state, promoter, or command summaries: %#v", readback)
	}
	terminalExamples, ok := rawReadback["foundry_terminal_status_examples"].([]any)
	if !ok || len(terminalExamples) != 4 {
		t.Fatalf("readback missing structured Foundry terminal examples: %#v", rawReadback["foundry_terminal_status_examples"])
	}
	terminalByStatus := map[string]map[string]any{}
	for _, item := range terminalExamples {
		example, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("bad terminal example: %#v", item)
		}
		status, _ := example["source_status"].(string)
		terminalByStatus[status] = example
	}
	if terminalByStatus["promoted"]["normalized_status"] != "completed" ||
		terminalByStatus["promoted"]["can_close_mission"] != true ||
		terminalByStatus["promoted"]["required_readback"] != "Promoter and Command agree promotion is terminal, RSI remains denied, and no ready nodes remain." ||
		terminalByStatus["denied"]["can_close_mission"] != true ||
		terminalByStatus["blocked"]["can_close_mission"] != false {
		t.Fatalf("structured terminal examples do not describe promoted/denied/blocked closure: %#v", terminalByStatus)
	}
	deniedExamples, ok := rawReadback["foundry_denied_terminal_examples"].([]any)
	if !ok || len(deniedExamples) < 3 {
		t.Fatalf("readback missing structured Foundry denied terminal examples: %#v", rawReadback["foundry_denied_terminal_examples"])
	}
	deniedByReason := map[string]map[string]any{}
	for _, item := range deniedExamples {
		example, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("bad denied terminal example: %#v", item)
		}
		reason, _ := example["denial_reason"].(string)
		deniedByReason[reason] = example
	}
	if deniedByReason["missing_node_evidence"]["requires_exact_missing_evidence"] != true ||
		deniedByReason["missing_stop_gate_evidence"]["can_close_mission"] != true ||
		deniedByReason["forbidden_surface_or_rsi_claim"]["rsi_remains_denied"] != true ||
		deniedByReason["forbidden_surface_or_rsi_claim"]["authority_advance_claimed"] != false {
		t.Fatalf("denied terminal examples do not describe exact blocker and RSI-safe denial: %#v", deniedByReason)
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
		"`early_return_risk_status`",
		"If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
		"If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v0.2 prompt missing section %q:\n%s", want, prompt)
		}
	}
}

func TestRecommendationLongRunSupervisorExamplesCoverLeaseFields(t *testing.T) {
	root := filepath.Join(repoRoot(t), "examples", "valid")
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, "recommendation-wave-long-run-supervisor.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.Supervisor == nil ||
		wave.MinimumTasks != 30 ||
		wave.NodeBudget != 40 ||
		wave.EstimatedMinutes != 120 ||
		wave.Supervisor.MinNodes != 30 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 ||
		wave.Supervisor.ReturnOnlyWhen != "all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker" ||
		wave.Supervisor.CheckpointPolicy != "after_each_node_or_timed_interval" ||
		wave.Supervisor.FinalReportContract != "ao.atlas.long-run-final-report.v0.2" {
		t.Fatalf("long-run wave example missing supervisor lease fields: %#v", wave)
	}
	if wave.FinalResponseAllowed || wave.SafeToExecute || wave.SchedulesWork || wave.ExecutesWork || wave.ApprovesWork {
		t.Fatalf("long-run wave example widened authority: %#v", wave)
	}

	lease := mustLoadJSON[AtlasRecommendationLeaseStart](t, filepath.Join(root, "recommendation-lease-start-long-run.json"))
	if err := ValidateAtlasRecommendationLeaseStart(lease); err != nil {
		t.Fatal(err)
	}
	if lease.MinMinutes != 120 ||
		lease.MaxMinutes != 180 ||
		lease.ContinueIfFastTarget != 40 ||
		lease.CheckpointPolicy != "after_each_node_or_timed_interval" ||
		lease.FinalResponseAllowed ||
		lease.SchedulesWork ||
		lease.ExecutesWork ||
		lease.ApprovesWork ||
		lease.MutatesRepositories ||
		lease.CallsProviders ||
		lease.ClaimsAuthorityAdvance {
		t.Fatalf("long-run lease example missing lease or safety fields: %#v", lease)
	}
}

func TestMissionRecommendationsImportArtifactsAreDigestBound(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--started-at", "2026-07-04T08:00:00-07:00",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	wavePath := filepath.Join(outDir, "recommendation-wave.json")
	workgraphPath := filepath.Join(outDir, "recommendation-workgraph.json")
	waveDigest, err := digestFile(wavePath)
	if err != nil {
		t.Fatal(err)
	}
	workgraphDigest, err := digestFileWithNormalizedLineEndings(workgraphPath)
	if err != nil {
		t.Fatal(err)
	}

	leaseStart := mustLoadJSON[AtlasRecommendationLeaseStart](t, filepath.Join(outDir, "lease-start.json"))
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(outDir, "recommendation-readback.json"))
	readinessPacket := mustLoadJSON[AtlasRecommendationWorkgraphReadinessPacket](t, filepath.Join(outDir, "workgraph-readiness-packet.json"))
	if leaseStart.WaveDigest != waveDigest || leaseStart.WorkgraphDigest != workgraphDigest {
		t.Fatalf("lease start digests do not bind generated artifacts: lease=%#v wave=%s workgraph=%s", leaseStart, waveDigest, workgraphDigest)
	}
	if readback.WaveDigest != waveDigest || readback.WorkgraphDigest != workgraphDigest {
		t.Fatalf("recommendation readback digests do not bind generated artifacts: readback=%#v wave=%s workgraph=%s", readback, waveDigest, workgraphDigest)
	}
	readbackDigest, err := digestFile(filepath.Join(outDir, "recommendation-readback.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readinessPacket.WaveDigest != waveDigest ||
		readinessPacket.WorkgraphDigest != workgraphDigest ||
		readinessPacket.ReadbackDigest != readbackDigest {
		t.Fatalf("readiness packet does not bind source digests: packet=%#v wave=%s workgraph=%s readback=%s", readinessPacket, waveDigest, workgraphDigest, readbackDigest)
	}
}

func TestMissionRecommendationsFirstNodeFoundryImportSmoke(t *testing.T) {
	scratchRel := filepath.Join("..", "..", "target", "mission-recommendations-first-node-foundry-import-smoke")
	scratchAbs := filepath.Join(repoRoot(t), "target", "mission-recommendations-first-node-foundry-import-smoke")
	if err := os.RemoveAll(scratchAbs); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(scratchAbs)
	})
	recommendationsPath := filepath.Join(scratchRel, "feature-depth-recommendations.json")
	recommendationsOut := filepath.Join(scratchRel, "recommendations-out")
	foundryOut := filepath.Join(scratchRel, "foundry-import")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--started-at", "2026-07-04T08:00:00-07:00",
		"--out", recommendationsOut,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	workgraphPath := filepath.Join(recommendationsOut, "recommendation-workgraph.json")
	leaseStartPath := filepath.Join(recommendationsOut, "lease-start.json")

	out.Reset()
	code = Run([]string{
		"foundry", "import",
		"--workgraph", workgraphPath,
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--node", "mission-recommendation-next-01",
		"--mission-continuation", leaseStartPath,
		"--out", foundryOut,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "tasks=1") ||
		!strings.Contains(out.String(), "next_recommended_action=Move to ../ao-foundry") {
		t.Fatalf("foundry import output missing single-node continuation readback: %s", out.String())
	}

	workgraph := mustLoadJSON[Workgraph](t, workgraphPath)
	manifestPath := filepath.Join(foundryOut, "foundry-import.json")
	manifest := mustLoadJSON[FoundryImport](t, manifestPath)
	if err := ValidateFoundryImport(manifest); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImportMatchesWorkgraph(workgraph, manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 1 {
		t.Fatalf("recommendation first-node import should emit exactly one task fixture: %#v", manifest.Tasks)
	}
	task := manifest.Tasks[0]
	if task.NodeID != "mission-recommendation-next-01" ||
		task.TaskID != "mission-recommendation-next-01-task" ||
		task.Task.ID != task.TaskID ||
		task.Path != "tasks/mission-recommendation-next-01-task.json" {
		t.Fatalf("bad first recommendation task fixture: %#v", task)
	}
	if task.MutationClass != "low_risk_code" ||
		task.AuthorityBoundary != "atlas_recommendation_planning_only" ||
		!containsString(task.RequiredGates, "node_gate") ||
		!containsStringPrefix(task.RequiredEvidence, "source_task_digest:sha256:") {
		t.Fatalf("recommendation task fixture lost gate, authority, or source digest binding: %#v", task)
	}
	if _, err := os.Stat(filepath.Join(foundryOut, task.Path)); err != nil {
		t.Fatal(err)
	}
	if len(manifest.SourceArtifacts) != 2 {
		t.Fatalf("expected workgraph and instance source artifacts, got %#v", manifest.SourceArtifacts)
	}
	for _, source := range manifest.SourceArtifacts {
		if !digestPattern.MatchString(source.Digest) {
			t.Fatalf("source artifact missing digest binding: %#v", manifest.SourceArtifacts)
		}
	}
	if manifest.SchedulesWork || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("Foundry import smoke must remain fixture-only: %#v", manifest)
	}

	handoff := mustLoadJSON[FoundryContinuationHandoff](t, filepath.Join(foundryOut, "foundry-continuation-handoff.json"))
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		t.Fatal(err)
	}
	if handoff.FirstSafeNode != "mission-recommendation-next-01" ||
		handoff.TotalNodeCount != 40 ||
		handoff.ReadyNodeCount != 40 ||
		handoff.CompletedNodeCount != 0 ||
		handoff.FoundryImportPath != filepath.ToSlash(manifestPath) {
		t.Fatalf("handoff lost generated recommendation node readback: %#v", handoff)
	}
	prompt, err := os.ReadFile(filepath.Join(foundryOut, "foundry-continuation-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"first safe node: mission-recommendation-next-01",
		"do not stop after import validation",
		"do not stop after one node",
		"RSI remains denied",
		filepath.ToSlash(workgraphPath),
		filepath.ToSlash(manifestPath),
	} {
		if !strings.Contains(string(prompt), want) {
			t.Fatalf("Foundry continuation prompt missing %q:\n%s", want, string(prompt))
		}
	}
	if strings.Contains(string(prompt), "RSI is proven") || strings.Contains(handoff.Prompt, "RSI is proven") {
		t.Fatalf("Foundry continuation prompt must avoid unsafe RSI proof wording:\n%s", string(prompt))
	}
}
