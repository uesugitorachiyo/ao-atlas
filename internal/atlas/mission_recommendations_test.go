package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func containsStringPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

type requiredFieldSchema struct {
	Required   []string                       `json:"required"`
	Properties map[string]requiredFieldSchema `json:"properties"`
}

func assertSchemaRequiredFieldsPresent(t *testing.T, schemaPath, artifactPath string) {
	t.Helper()
	schema := mustLoadJSON[requiredFieldSchema](t, schemaPath)
	if len(schema.Required) == 0 {
		t.Fatalf("schema %s has no required fields", schemaPath)
	}
	artifact := mustLoadJSON[map[string]any](t, artifactPath)
	for _, field := range schema.Required {
		if _, ok := artifact[field]; !ok {
			t.Fatalf("%s missing required schema field %q from %s", artifactPath, field, schemaPath)
		}
	}
}

func assertSchemaRequiresField(t *testing.T, schemaPath, field string) {
	t.Helper()
	schema := mustLoadJSON[requiredFieldSchema](t, schemaPath)
	for _, required := range schema.Required {
		if required == field {
			return
		}
	}
	t.Fatalf("schema %s does not require %q", schemaPath, field)
}

func assertNestedSchemaRequiresField(t *testing.T, schemaPath, property, field string) {
	t.Helper()
	schema := mustLoadJSON[requiredFieldSchema](t, schemaPath)
	nested, ok := schema.Properties[property]
	if !ok {
		t.Fatalf("schema %s missing nested property %q", schemaPath, property)
	}
	for _, required := range nested.Required {
		if required == field {
			return
		}
	}
	t.Fatalf("schema %s nested property %q does not require %q", schemaPath, property, field)
}

func TestRecommendationDerivedReadbackSchemasRequireLeaseHealthStatus(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas")
	for _, schemaName := range []string{
		"recommendation-checkpoint-readback.schema.json",
		"recommendation-command-readback.schema.json",
		"recommendation-promoter-readback.schema.json",
		"recommendation-foundry-rollup.schema.json",
		"recommendation-reconciliation-packet.schema.json",
	} {
		assertSchemaRequiresField(t, filepath.Join(root, schemaName), "lease_health_status")
	}
	assertNestedSchemaRequiresField(t, filepath.Join(root, "recommendation-command-readback.schema.json"), "command_timeline_binding", "lease_health_status")
}

func TestRecommendationDerivedReadbackSchemasRequireCheckpointFreshnessStatus(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas")
	for _, schemaName := range []string{
		"recommendation-checkpoint-readback.schema.json",
		"recommendation-command-readback.schema.json",
		"recommendation-promoter-readback.schema.json",
		"recommendation-foundry-rollup.schema.json",
		"recommendation-reconciliation-packet.schema.json",
	} {
		assertSchemaRequiresField(t, filepath.Join(root, schemaName), "checkpoint_freshness_status")
	}
	assertNestedSchemaRequiresField(t, filepath.Join(root, "recommendation-command-readback.schema.json"), "command_timeline_binding", "checkpoint_freshness_status")
}

func TestRecommendationReconciliationSchemaRequiresStaleRouteDecisionStatus(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-reconciliation-packet.schema.json"), "stale_route_decision_status")
}

func TestRecommendationReadbackSchemaRequiresFoundryTerminalExamples(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "foundry_terminal_status_examples")
}

func TestRecommendationReadbackSchemaRequiresFoundryDeniedTerminalExamples(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "foundry_denied_terminal_examples")
}

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
	rawReadback := mustLoadJSON[map[string]any](t, filepath.Join(outDir, "recommendation-readback.json"))
	if rawReadback["final_response_denial_gate"] != "deny_ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("readback missing final response denial gate: %#v", rawReadback["final_response_denial_gate"])
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
	if leaseStart.WaveDigest != waveDigest || leaseStart.WorkgraphDigest != workgraphDigest {
		t.Fatalf("lease start digests do not bind generated artifacts: lease=%#v wave=%s workgraph=%s", leaseStart, waveDigest, workgraphDigest)
	}
	if readback.WaveDigest != waveDigest || readback.WorkgraphDigest != workgraphDigest {
		t.Fatalf("recommendation readback digests do not bind generated artifacts: readback=%#v wave=%s workgraph=%s", readback, waveDigest, workgraphDigest)
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
		!strings.Contains(out.String(), "return_gate_status=blocked_ready_nodes_remain") ||
		!strings.Contains(out.String(), "checkpoint_count=0") ||
		!strings.Contains(out.String(), "exact_next_action=Emit Foundry import for mission-recommendation-next-01 and execute exactly one active node.") {
		t.Fatalf("readback output missing final gate and exact action: %s", out.String())
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)
	if readback.WaveDigest == "" || readback.WorkgraphDigest == "" {
		t.Fatalf("readback must carry artifact digests: %#v", readback)
	}
	if readback.ReturnGateStatus != "blocked_ready_nodes_remain" || readback.CheckpointCount != 0 {
		t.Fatalf("readback must expose return gate status and checkpoint count: %#v", readback)
	}
}

func TestMissionRecommendationsImportPersistsLeaseStartAndResumeUsesIt(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	importDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--started-at", "2026-07-04T08:00:00-07:00",
		"--out", importDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "lease_start=") {
		t.Fatalf("import output missing lease start artifact: %s", out.String())
	}
	leaseStartPath := filepath.Join(importDir, "lease-start.json")
	leaseStart := mustLoadJSON[AtlasRecommendationLeaseStart](t, leaseStartPath)
	if err := ValidateAtlasRecommendationLeaseStart(leaseStart); err != nil {
		t.Fatal(err)
	}
	if leaseStart.StartedAt != "2026-07-04T08:00:00-07:00" ||
		leaseStart.MinMinutes != 120 ||
		leaseStart.MaxMinutes != 180 ||
		leaseStart.FinalResponseAllowed {
		t.Fatalf("bad lease start marker: %#v", leaseStart)
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(importDir, "recommendation-workgraph.json"))
	firstNode := workgraph.Nodes[0]
	linkPath := filepath.Join(dir, "run-link.json")
	if err := WriteJSON(linkPath, recommendationRunLink(t, firstNode.FactoryTask.ID, recommendationEvidenceFiles(t, "lease-start", firstNode.ID))); err != nil {
		t.Fatal(err)
	}
	updatedWorkgraphPath := filepath.Join(dir, "updated-workgraph.json")
	readbackPath := filepath.Join(dir, "updated-readback.json")
	executionPath := filepath.Join(dir, "updated-execution-readback.json")
	checkpointPath := filepath.Join(dir, "checkpoint-readback.json")

	out.Reset()
	code = Run([]string{
		"mission", "recommendations", "complete-node",
		"--wave", filepath.Join(importDir, "recommendation-wave.json"),
		"--workgraph", filepath.Join(importDir, "recommendation-workgraph.json"),
		"--run-link", linkPath,
		"--expected-node", firstNode.ID,
		"--evidence-root", repoRoot(t),
		"--readback-evidence-root", "docs/evidence/test-wave",
		"--lease-start", leaseStartPath,
		"--completed-at", "2026-07-04T08:17:00-07:00",
		"--out-workgraph", updatedWorkgraphPath,
		"--out-readback", readbackPath,
		"--out-execution-readback", executionPath,
		"--out-checkpoint-readback", checkpointPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("complete-node failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "checkpoint_readback=") ||
		!strings.Contains(out.String(), "elapsed_minutes=17") {
		t.Fatalf("complete-node output missing checkpoint timing: %s", out.String())
	}
	checkpoint := mustLoadJSON[AtlasRecommendationCheckpointReadback](t, checkpointPath)
	if err := ValidateAtlasRecommendationCheckpointReadback(checkpoint); err != nil {
		t.Fatal(err)
	}
	if checkpoint.StartedAt != leaseStart.StartedAt ||
		checkpoint.ElapsedMinutes != 17 ||
		checkpoint.LeaseHealthStatus != "minimum_unmet" ||
		checkpoint.CompletedNodes != 1 ||
		checkpoint.ReadyNodes != 39 ||
		checkpoint.FinalResponseAllowed {
		t.Fatalf("bad checkpoint readback: %#v", checkpoint)
	}

	resumeReadbackPath := filepath.Join(dir, "resume-readback.json")
	resumeExecutionPath := filepath.Join(dir, "resume-execution-readback.json")
	commandPath := filepath.Join(dir, "command-readback.json")
	promoterPath := filepath.Join(dir, "promoter-readback.json")
	foundryPath := filepath.Join(dir, "foundry-rollup.json")
	reconciliationPath := filepath.Join(dir, "reconciliation-packet.json")
	nextPromptPath := filepath.Join(dir, "next-recommended-prompt.md")
	out.Reset()
	code = Run([]string{
		"mission", "recommendations", "resume",
		"--wave", filepath.Join(importDir, "recommendation-wave.json"),
		"--workgraph", updatedWorkgraphPath,
		"--lease-start", leaseStartPath,
		"--completed-at", "2026-07-04T08:25:00-07:00",
		"--evidence-root", "docs/evidence/test-wave",
		"--out-readback", resumeReadbackPath,
		"--out-execution-readback", resumeExecutionPath,
		"--out-command-readback", commandPath,
		"--out-promoter-readback", promoterPath,
		"--out-foundry-rollup", foundryPath,
		"--out-reconciliation-packet", reconciliationPath,
		"--out-next-prompt", nextPromptPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("resume failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "started_at=2026-07-04T08:00:00-07:00") ||
		!strings.Contains(out.String(), "elapsed_minutes=25") ||
		!strings.Contains(out.String(), "next_recommended_prompt=") {
		t.Fatalf("resume output did not preserve lease timing: %s", out.String())
	}
	resumeReadback := mustLoadJSON[AtlasRecommendationReadback](t, resumeReadbackPath)
	if resumeReadback.StartedAt != leaseStart.StartedAt || resumeReadback.ElapsedMinutes != 25 {
		t.Fatalf("resume readback lost lease start: %#v", resumeReadback)
	}
	resumeExecution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, resumeExecutionPath)
	if resumeExecution.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.GeneratedWorkgraph.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.GeneratedWorkgraph.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.CompletedRunLinks != resumeReadback.CompletedNodes ||
		resumeExecution.FoundryRunLinkReadinessSummary.RequiredRunLinks != resumeReadback.TotalNodes ||
		resumeExecution.FoundryRunLinkReadinessSummary.NextExecutableNode != resumeReadback.FirstExecutableNode ||
		!hasSourceArtifact(resumeExecution.SourceArtifacts, "foundry_run_link_readiness_summary") {
		t.Fatalf("execution readback missing Foundry run-link readiness source artifact: %#v", resumeExecution)
	}
	command := mustLoadJSON[AtlasRecommendationCommandReadback](t, commandPath)
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, promoterPath)
	foundry := mustLoadJSON[AtlasRecommendationFoundryRollup](t, foundryPath)
	reconciliation := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, reconciliationPath)
	if err := ValidateAtlasRecommendationClosureArtifacts(resumeReadback, command, promoter, foundry); err != nil {
		t.Fatalf("closure artifacts should agree with resumed readback: %v", err)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(resumeReadback, command, promoter, foundry, reconciliation); err != nil {
		t.Fatalf("reconciliation packet should agree with resumed closure artifacts: %v", err)
	}
	rawPromoter := mustLoadJSON[map[string]any](t, promoterPath)
	if rawPromoter["no_promotion_summary"] != "No mutation authority promotion claimed; RSI remains denied." ||
		rawPromoter["next_denied_class"] != "RSI" {
		t.Fatalf("promoter readback missing no-promotion summary fields: %#v", rawPromoter)
	}
	rawCommand := mustLoadJSON[map[string]any](t, commandPath)
	binding, ok := rawCommand["command_timeline_binding"].(map[string]any)
	if !ok ||
		binding["summary"] != command.CompactTimeline ||
		binding["lease_health_status"] != resumeReadback.LeaseHealthStatus ||
		binding["checkpoint_freshness_status"] != resumeReadback.CheckpointFreshnessStatus ||
		binding["first_executable_node"] != resumeReadback.FirstExecutableNode ||
		binding["exact_next_action"] != resumeReadback.ExactNextAction ||
		binding["return_gate_status"] != resumeReadback.ReturnGateStatus {
		t.Fatalf("command readback missing structured timeline binding: %#v", rawCommand)
	}
	if command.ElapsedMinutes != 25 || command.FinalResponseAllowed ||
		command.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		command.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		command.CommandTimelineBinding.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		command.CommandTimelineBinding.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		promoter.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		promoter.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		promoter.PromotionClaimed || !promoter.RSIRemainsDenied ||
		foundry.NodeCompletionStatus != "nodes_in_progress" ||
		foundry.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		foundry.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		foundry.LeaseCompletionStatus != "minimum_minutes_unmet" {
		t.Fatalf("bad resume closure artifacts: command=%#v promoter=%#v foundry=%#v", command, promoter, foundry)
	}
	if reconciliation.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		reconciliation.CheckpointCount != 1 ||
		reconciliation.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		reconciliation.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		reconciliation.StaleRouteDecisionStatus != resumeReadback.StaleRouteDecisionStatus ||
		!reconciliation.ArtifactsAgree ||
		reconciliation.CommandReturnGateStatus != resumeReadback.ReturnGateStatus ||
		reconciliation.FoundryReturnGateStatus != resumeReadback.ReturnGateStatus ||
		reconciliation.PromotionClaimed {
		t.Fatalf("bad reconciliation packet: %#v", reconciliation)
	}
	nextPrompt, err := os.ReadFile(nextPromptPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Current status:",
		"Completed nodes: 1 / 40",
		"Early-return risk: `" + resumeReadback.EarlyReturnRiskStatus + "`",
		"Next executable node: `mission-recommendation-next-02`",
		"Exact next action:",
		"Emit Foundry import for mission-recommendation-next-02 and execute exactly one active node.",
		"If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.",
		"If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(string(nextPrompt), want) {
			t.Fatalf("resume next prompt missing %q:\n%s", want, string(nextPrompt))
		}
	}
	schemaRoot := filepath.Join(repoRoot(t), "schemas")
	for _, tc := range []struct {
		schemaPath   string
		artifactPath string
	}{
		{filepath.Join(schemaRoot, "recommendation-readback.schema.json"), resumeReadbackPath},
		{filepath.Join(schemaRoot, "recommendation-checkpoint-readback.schema.json"), checkpointPath},
		{filepath.Join(schemaRoot, "recommendation-command-readback.schema.json"), commandPath},
		{filepath.Join(schemaRoot, "recommendation-promoter-readback.schema.json"), promoterPath},
		{filepath.Join(schemaRoot, "recommendation-foundry-rollup.schema.json"), foundryPath},
		{filepath.Join(schemaRoot, "recommendation-reconciliation-packet.schema.json"), reconciliationPath},
	} {
		assertSchemaRequiredFieldsPresent(t, tc.schemaPath, tc.artifactPath)
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
	completedReadback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:00-07:00",
		CompletedAt:     "2026-07-04T09:20:00-07:00",
		ElapsedMinutes:  120,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !completedReadback.FinalResponseAllowed ||
		completedReadback.FinalResponseReason != "all generated nodes complete and no ready nodes remain" ||
		completedReadback.LeaseHealthStatus != "all_generated_nodes_complete" ||
		!completedReadback.MinMinutesMet ||
		completedReadback.LeaseTimeStatus != "minimum_minutes_met" ||
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

func TestMissionRecommendationsDenyFinalResponseWhenLeaseMinutesUnmet(t *testing.T) {
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

	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:20-07:00",
		CompletedAt:     "2026-07-04T07:42:06-07:00",
		ElapsedMinutes:  22,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.FinalResponseAllowed {
		t.Fatalf("completed nodes cannot close before min_minutes: %#v", readback)
	}
	if readback.Status != "in_progress" ||
		readback.MinMinutesMet ||
		readback.LeaseTimeStatus != "minimum_minutes_unmet" ||
		readback.LeaseHealthStatus != "minimum_minutes_unmet_continue_next_wave" ||
		readback.EarlyReturnRiskStatus != "blocked_final_response_minimum_minutes_unmet" ||
		readback.FinalResponseReason != "minimum lease minutes unmet" {
		t.Fatalf("readback did not report unmet lease timing: %#v", readback)
	}
	if !strings.Contains(readback.ExactNextAction, "Generate and execute the next useful Atlas recommendation wave") {
		t.Fatalf("readback missing continuation action after short run: %#v", readback)
	}
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("execution ledger should stay consistent for early timing denial: %v", err)
	}
	if execution.Status == "completed" {
		t.Fatalf("execution ledger cannot be completed before min_minutes: %#v", execution)
	}
}

func TestMissionRecommendationsDenyFinalResponseWhenLeaseTimingMissing(t *testing.T) {
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

	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if readback.FinalResponseAllowed ||
		readback.MinMinutesMet ||
		readback.LeaseTimeStatus != "lease_timing_missing" ||
		readback.FinalResponseReason != "minimum lease timing evidence missing" {
		t.Fatalf("completed long-run wave without timing must deny final response: %#v", readback)
	}
	if !strings.Contains(readback.ExactNextAction, "Record started_at, completed_at, and elapsed_minutes") {
		t.Fatalf("missing timing denial should ask for timing evidence: %#v", readback)
	}
}

func TestMissionRecommendationsDeriveElapsedMinutesFromTimestamps(t *testing.T) {
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

	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T08:00:00-07:00",
		CompletedAt:     "2026-07-04T10:00:01-07:00",
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.ElapsedMinutes != 121 ||
		!readback.MinMinutesMet ||
		readback.LeaseTimeStatus != "minimum_minutes_met" ||
		!readback.FinalResponseAllowed {
		t.Fatalf("readback did not derive elapsed lease minutes: %#v", readback)
	}
}

func TestMissionRecommendationsRejectInvalidLeaseTimestamps(t *testing.T) {
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

	completed := completeRecommendationNodes(result.Workgraph, 40)
	_, err = BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:   "not-a-time",
		CompletedAt: "2026-07-04T08:00:00-07:00",
	})
	if err == nil || !strings.Contains(err.Error(), "started_at must be RFC3339") {
		t.Fatalf("expected invalid started_at rejection, got %v", err)
	}
	_, err = BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:   "2026-07-04T09:00:00-07:00",
		CompletedAt: "2026-07-04T08:00:00-07:00",
	})
	if err == nil || !strings.Contains(err.Error(), "completed_at must be greater than or equal to started_at") {
		t.Fatalf("expected reversed timestamp rejection, got %v", err)
	}
}

func TestMissionRecommendationsDetectStaleClosureArtifacts(t *testing.T) {
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
	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T08:00:00-07:00",
		CompletedAt:     "2026-07-04T08:22:00-07:00",
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	command := BuildAtlasRecommendationCommandReadback(readback)
	promoter := BuildAtlasRecommendationPromoterReadback(readback)
	foundry := BuildAtlasRecommendationFoundryRollup(readback)
	command.FinalResponseAllowed = true
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback final_response_allowed disagrees") {
		t.Fatalf("expected stale command readback rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	foundry.Status = "completed"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "foundry rollup completed while recommendation final response is denied") {
		t.Fatalf("expected stale foundry rollup rejection, got %v", err)
	}
	foundry = BuildAtlasRecommendationFoundryRollup(readback)
	command.CommandTimelineBinding.ExactNextAction = "stale next action"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command timeline binding exact_next_action disagrees") {
		t.Fatalf("expected stale command timeline binding rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.LeaseHealthStatus = "stale_lease_health"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback lease_health_status disagrees") {
		t.Fatalf("expected stale command lease health rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CommandTimelineBinding.LeaseHealthStatus = "stale_lease_health"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command timeline binding lease_health_status disagrees") {
		t.Fatalf("expected stale command timeline lease health rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	promoter.LeaseHealthStatus = "stale_lease_health"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "promoter readback lease_health_status disagrees") {
		t.Fatalf("expected stale promoter lease health rejection, got %v", err)
	}
	promoter = BuildAtlasRecommendationPromoterReadback(readback)
	foundry.LeaseHealthStatus = "stale_lease_health"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "foundry rollup lease_health_status disagrees") {
		t.Fatalf("expected stale foundry lease health rejection, got %v", err)
	}
	foundry = BuildAtlasRecommendationFoundryRollup(readback)
	command.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback checkpoint_freshness_status disagrees") {
		t.Fatalf("expected stale command checkpoint freshness rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CommandTimelineBinding.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command timeline binding checkpoint_freshness_status disagrees") {
		t.Fatalf("expected stale command timeline checkpoint freshness rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	promoter.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "promoter readback checkpoint_freshness_status disagrees") {
		t.Fatalf("expected stale promoter checkpoint freshness rejection, got %v", err)
	}
	promoter = BuildAtlasRecommendationPromoterReadback(readback)
	foundry.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "foundry rollup checkpoint_freshness_status disagrees") {
		t.Fatalf("expected stale foundry checkpoint freshness rejection, got %v", err)
	}
	foundry = BuildAtlasRecommendationFoundryRollup(readback)
	packet := BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.StaleRouteDecisionStatus = "stale_route_decision"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "reconciliation stale_route_decision_status disagrees") {
		t.Fatalf("expected stale reconciliation route decision rejection, got %v", err)
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
		!strings.Contains(out.String(), "checkpoint_count=1") ||
		!strings.Contains(out.String(), "return_gate_status=blocked_ready_nodes_remain") ||
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
	if readback.ReturnGateStatus != "blocked_ready_nodes_remain" || readback.CheckpointCount != 1 {
		t.Fatalf("readback must carry node checkpoint count and return gate status: %#v", readback)
	}
	execution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, executionPath)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("bad execution artifact: %v", err)
	}
	if execution.GeneratedWorkgraph.ReturnGateStatus != readback.ReturnGateStatus ||
		execution.GeneratedWorkgraph.CheckpointCount != readback.CheckpointCount {
		t.Fatalf("execution readback missing status gate mirror: %#v", execution.GeneratedWorkgraph)
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
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	execution.Status = "completed"
	execution.CompletedRecommendationNodes = 40

	err := ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "completed_recommendation_nodes must match recommendation readback completed_nodes") {
		t.Fatalf("expected false completion rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.CompletedRunLinks = 40
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness completed_run_links must match recommendation readback completed_nodes") {
		t.Fatalf("expected stale Foundry run-link readiness rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.LeaseHealthStatus = "stale_lease_health"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "lease_health_status must match recommendation readback") {
		t.Fatalf("expected stale execution lease health rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.GeneratedWorkgraph.LeaseHealthStatus = "stale_lease_health"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "generated_workgraph.lease_health_status must match recommendation readback") {
		t.Fatalf("expected stale generated workgraph lease health rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.LeaseHealthStatus = "stale_lease_health"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness lease_health_status must match recommendation readback") {
		t.Fatalf("expected stale Foundry lease health rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "checkpoint_freshness_status must match recommendation readback") {
		t.Fatalf("expected stale execution checkpoint freshness rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.GeneratedWorkgraph.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "generated_workgraph.checkpoint_freshness_status must match recommendation readback") {
		t.Fatalf("expected stale generated workgraph checkpoint freshness rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.CheckpointFreshnessStatus = "stale_checkpoint_freshness"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness checkpoint_freshness_status must match recommendation readback") {
		t.Fatalf("expected stale Foundry checkpoint freshness rejection, got %v", err)
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

func TestLeaseResumeWaveFinalStateEvidenceMatchesPrompt(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-lease-resume-wave-v01")
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, "recommendation-wave.json"))
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "recommendation-readback.json"))
	var synthesis struct {
		CompletedNodes        int    `json:"completed_nodes"`
		TotalNodes            int    `json:"total_nodes"`
		ReadyNodes            int    `json:"ready_nodes"`
		CheckpointCount       int    `json:"checkpoint_count"`
		ElapsedMinutes        int    `json:"elapsed_minutes"`
		ReturnGateStatus      string `json:"return_gate_status"`
		FinalResponseAllowed  bool   `json:"final_response_allowed"`
		ExactNextAction       string `json:"exact_next_action"`
		NextRecommendedPrompt string `json:"next_recommended_prompt"`
	}
	synthesis = mustLoadJSON[struct {
		CompletedNodes        int    `json:"completed_nodes"`
		TotalNodes            int    `json:"total_nodes"`
		ReadyNodes            int    `json:"ready_nodes"`
		CheckpointCount       int    `json:"checkpoint_count"`
		ElapsedMinutes        int    `json:"elapsed_minutes"`
		ReturnGateStatus      string `json:"return_gate_status"`
		FinalResponseAllowed  bool   `json:"final_response_allowed"`
		ExactNextAction       string `json:"exact_next_action"`
		NextRecommendedPrompt string `json:"next_recommended_prompt"`
	}](t, filepath.Join(root, "final-synthesis.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	workgraphPath := filepath.Join(root, "recommendation-workgraph.json")
	if readback.CompletedNodes > 0 {
		workgraphPath = filepath.Join(root, "nodes", "mission-recommendation-next-"+twoDigit(readback.CompletedNodes), "workgraph-after.json")
	}
	workgraph := mustLoadJSON[Workgraph](t, workgraphPath)
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	workgraphDigest, err := digestFileWithNormalizedLineEndings(workgraphPath)
	if err != nil {
		t.Fatal(err)
	}
	if readback.WorkgraphDigest != workgraphDigest {
		t.Fatalf("root readback workgraph digest does not match latest workgraph: readback=%s latest=%s path=%s", readback.WorkgraphDigest, workgraphDigest, workgraphPath)
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if wave.TotalTasks != readback.TotalNodes ||
		len(workgraph.Nodes) != readback.TotalNodes ||
		state.NodeCounts["completed"] != readback.CompletedNodes ||
		state.NodeCounts["ready"] != readback.ReadyNodes ||
		len(state.ExecutableReadyNodeIDs) != readback.ExecutableReadyNodes {
		t.Fatalf("wave, workgraph, and readback disagree: wave=%#v state=%#v readback=%#v", wave, state, readback)
	}
	if readback.FirstExecutableNode != "" && (len(state.ExecutableReadyNodeIDs) == 0 || state.ExecutableReadyNodeIDs[0] != readback.FirstExecutableNode) {
		t.Fatalf("readback first executable node disagrees with workgraph state: state=%#v readback=%#v", state.ExecutableReadyNodeIDs, readback.FirstExecutableNode)
	}
	if synthesis.CompletedNodes != readback.CompletedNodes ||
		synthesis.TotalNodes != readback.TotalNodes ||
		synthesis.ReadyNodes != readback.ReadyNodes ||
		synthesis.CheckpointCount != readback.CheckpointCount ||
		synthesis.ElapsedMinutes != readback.ElapsedMinutes ||
		synthesis.ReturnGateStatus != readback.ReturnGateStatus ||
		synthesis.FinalResponseAllowed != readback.FinalResponseAllowed ||
		synthesis.ExactNextAction != readback.ExactNextAction {
		t.Fatalf("final synthesis does not match root readback: synthesis=%#v readback=%#v", synthesis, readback)
	}
	promptPath := filepath.Join(root, "next-recommended-prompt.md")
	if synthesis.NextRecommendedPrompt != "docs/evidence/ao-atlas-lease-resume-wave-v01/next-recommended-prompt.md" {
		t.Fatalf("final synthesis points at wrong prompt: %#v", synthesis.NextRecommendedPrompt)
	}
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	workgraphRef := filepath.ToSlash(workgraphPath)
	if idx := strings.Index(workgraphRef, "docs/"); idx >= 0 {
		workgraphRef = workgraphRef[idx:]
	}
	for _, want := range []string{
		"Current workgraph: `" + workgraphRef + "`",
		"Completed nodes: " + strconv.Itoa(readback.CompletedNodes) + " / " + strconv.Itoa(readback.TotalNodes),
		"Ready nodes: " + strconv.Itoa(readback.ReadyNodes),
		"Elapsed minutes at latest checkpoint: " + strconv.Itoa(readback.ElapsedMinutes),
		"`final_response_allowed=" + strconv.FormatBool(readback.FinalResponseAllowed) + "`",
		"Return gate: `" + readback.ReturnGateStatus + "`",
		"Early-return risk: `" + readback.EarlyReturnRiskStatus + "`",
		"Checkpoint count: " + strconv.Itoa(readback.CheckpointCount),
		"Next executable node: `" + readback.FirstExecutableNode + "`",
		readback.ExactNextAction,
		"If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("continuation prompt missing final-state evidence %q:\n%s", want, prompt)
		}
	}
}

func digestFileWithNormalizedLineEndings(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	return DigestBytes(data), nil
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
		"lease-start.json",
		"recommendation-readback.json",
		"mission recommendations readback",
		"mission recommendations complete-node",
		"mission recommendations resume",
		"--lease-start",
		"--elapsed-minutes",
		"--started-at",
		"--completed-at",
		"--lease-timing-mode",
		"--out-checkpoint-readback",
		"checkpoint-readback-after-node-01.json",
		"command-readback-resumed.json",
		"promoter-readback-resumed.json",
		"foundry-rollup-resumed.json",
		"reconciliation-packet-resumed.json",
		"--out-reconciliation-packet",
		"recommendation-reconciliation-packet.schema.json",
		"recommendation-lease-start.schema.json",
		"recommendation-checkpoint-readback.schema.json",
		"recommendation-command-readback.schema.json",
		"recommendation-promoter-readback.schema.json",
		"recommendation-foundry-rollup.schema.json",
		"minimum_minutes_unmet",
		"lease_timing_missing",
		"minimum_minutes_met",
		"--out-execution-readback",
		"completed_recommendation_nodes",
		"checkpoint_count",
		"return_gate_status",
		"blocked_ready_nodes_remain",
		"blocked_minimum_minutes_unmet",
		"blocked_lease_timing_missing",
		"final_response_allowed",
		"min_minutes_met=true",
		"recommendation-ledger-consistency",
		"next-recommended-prompt.md",
		"reject_generated_recommendation_prompt_public_safety",
		"recommendation-prompt-public-safety-scan",
		"lease-resume-wave-public-safety-readback",
		"lease_resume_root=\"docs/evidence/ao-atlas-lease-resume-wave-v01\"",
		"final-synthesis.json",
		"Current workgraph:",
		"early_return_risk_status",
		"Early-return risk:",
		"do not produce a final response",
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

func hasSourceArtifact(sources []SourceRef, ref string) bool {
	for _, source := range sources {
		if source.Ref == ref && strings.HasPrefix(source.Digest, "sha256:") {
			return true
		}
	}
	return false
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
