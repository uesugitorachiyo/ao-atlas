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

func assertSchemaEnumContains(t *testing.T, schemaPath, property string, values ...string) {
	t.Helper()
	schema := mustLoadJSON[map[string]any](t, schemaPath)
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema %s missing properties", schemaPath)
	}
	field, ok := properties[property].(map[string]any)
	if !ok {
		t.Fatalf("schema %s missing property %q", schemaPath, property)
	}
	rawEnum, ok := field["enum"].([]any)
	if !ok || len(rawEnum) == 0 {
		t.Fatalf("schema %s property %q missing enum", schemaPath, property)
	}
	enum := map[string]bool{}
	for _, value := range rawEnum {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("schema %s property %q has non-string enum value %#v", schemaPath, property, value)
		}
		enum[text] = true
	}
	for _, value := range values {
		if !enum[value] {
			t.Fatalf("schema %s property %q enum missing %q", schemaPath, property, value)
		}
	}
}

func assertNestedSchemaEnumContains(t *testing.T, schemaPath, property, nestedProperty string, values ...string) {
	t.Helper()
	schema := mustLoadJSON[map[string]any](t, schemaPath)
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema %s missing properties", schemaPath)
	}
	parent, ok := properties[property].(map[string]any)
	if !ok {
		t.Fatalf("schema %s missing property %q", schemaPath, property)
	}
	parentProperties, ok := parent["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema %s property %q missing nested properties", schemaPath, property)
	}
	nested, ok := parentProperties[nestedProperty].(map[string]any)
	if !ok {
		t.Fatalf("schema %s property %q missing nested property %q", schemaPath, property, nestedProperty)
	}
	rawEnum, ok := nested["enum"].([]any)
	if !ok || len(rawEnum) == 0 {
		t.Fatalf("schema %s nested property %q.%q missing enum", schemaPath, property, nestedProperty)
	}
	enum := map[string]bool{}
	for _, value := range rawEnum {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("schema %s nested property %q.%q has non-string enum value %#v", schemaPath, property, nestedProperty, value)
		}
		enum[text] = true
	}
	for _, value := range values {
		if !enum[value] {
			t.Fatalf("schema %s nested property %q.%q enum missing %q", schemaPath, property, nestedProperty, value)
		}
	}
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

func TestRecommendationReconciliationSchemaRequiresFinalStateReconciliation(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-reconciliation-packet.schema.json"), "final_state_reconciliation")
}

func TestRecommendationReconciliationSchemaRequiresContinuationReasonAgreement(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-reconciliation-packet.schema.json")
	for _, field := range []string{
		"continuation_contract_reason",
		"command_continuation_contract_reason",
		"promoter_continuation_contract_reason",
		"foundry_continuation_contract_reason",
		"continuation_reason_agreement",
	} {
		assertSchemaRequiresField(t, root, field)
	}
	assertSchemaEnumContains(t, root, "continuation_contract_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationReconciliationSchemaRequiresFinalStateContinuationReason(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-reconciliation-packet.schema.json")
	assertNestedSchemaRequiresField(t, root, "final_state_reconciliation", "continuation_contract_reason")
	assertNestedSchemaRequiresField(t, root, "final_state_reconciliation", "continuation_reason_agreement")
	assertNestedSchemaEnumContains(t, root, "final_state_reconciliation", "continuation_contract_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationReadbackSchemaRequiresFoundryTerminalExamples(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "foundry_terminal_status_examples")
}

func TestRecommendationReadbackSchemaRequiresFoundryDeniedTerminalExamples(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "foundry_denied_terminal_examples")
}

func TestRecommendationReadbackSchemaRequiresExactNextActionReadback(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "exact_next_action_readback")
}

func TestRecommendationReadbackSchemaRequiresContinuationContract(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "continuation_contract")
}

func TestRecommendationReadbackSchemaBindsContinuationContractReasonEnum(t *testing.T) {
	assertNestedSchemaEnumContains(t,
		filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"),
		"continuation_contract",
		"reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationReadbackSchemaRequiresCommandTimelinePlaceholders(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "command_timeline_placeholders")
}

func TestRecommendationCommandReadbackSchemaRequiresContinuationReasonBinding(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-command-readback.schema.json")
	assertSchemaRequiresField(t, root, "continuation_contract_reason")
	assertNestedSchemaRequiresField(t, root, "command_timeline_binding", "continuation_contract_reason")
}

func TestRecommendationFoundryRollupSchemaRequiresContinuationReason(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-foundry-rollup.schema.json"), "continuation_contract_reason")
}

func TestRecommendationPromoterReadbackSchemaRequiresContinuationReason(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-promoter-readback.schema.json")
	assertSchemaRequiresField(t, root, "continuation_contract_reason")
	assertSchemaRequiresField(t, root, "no_promotion_reason_summary")
	assertSchemaEnumContains(t, root, "continuation_contract_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationCheckpointReadbackSchemaRequiresContinuationReason(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-checkpoint-readback.schema.json")
	assertSchemaRequiresField(t, root, "continuation_contract_reason")
	assertSchemaEnumContains(t, root, "continuation_contract_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationContinuationReasonCoverageSchemaRequiresAgreementSources(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-continuation-reason-coverage.schema.json")
	for _, field := range []string{
		"schema",
		"status",
		"expected_reason",
		"sources",
		"all_sources_agree",
		"final_response_allowed",
		"refuses_final_response",
		"exact_next_action",
		"claims_authority_advance",
		"rsi_remains_denied",
	} {
		assertSchemaRequiresField(t, root, field)
	}
	for _, field := range []string{
		"recommendation_readback",
		"checkpoint_readback",
		"workgraph_readiness_packet",
		"command_readback",
		"command_timeline_binding",
		"promoter_readback",
		"foundry_rollup",
		"reconciliation_packet",
		"reconciliation_command",
		"reconciliation_promoter",
		"reconciliation_foundry",
		"final_state_reconciliation",
		"resume_prompt",
	} {
		assertNestedSchemaRequiresField(t, root, "sources", field)
	}
	assertSchemaEnumContains(t, root, "expected_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
}

func TestRecommendationReadbackSchemaRequiresPromoterNoPromotionPlaceholders(t *testing.T) {
	assertSchemaRequiresField(t, filepath.Join(repoRoot(t), "schemas", "recommendation-readback.schema.json"), "promoter_no_promotion_placeholders")
}

func TestRecommendationWorkgraphReadinessPacketSchemaRequiresBudgetAndReturnGate(t *testing.T) {
	root := filepath.Join(repoRoot(t), "schemas", "recommendation-workgraph-readiness-packet.schema.json")
	for _, field := range []string{
		"schema",
		"status",
		"mission_id",
		"target_instance",
		"wave_digest",
		"workgraph_digest",
		"readback_digest",
		"total_nodes",
		"minimum_nodes",
		"node_budget",
		"continue_if_fast_target",
		"ready_nodes",
		"executable_ready_nodes",
		"return_gate_status",
		"early_return_risk_status",
		"final_response_allowed",
		"exact_next_action",
		"continuation_contract_reason",
		"one_executable_mutation_node_active",
		"refuses_final_response",
		"rsi_remains_denied",
	} {
		assertSchemaRequiresField(t, root, field)
	}
	assertSchemaEnumContains(t, root, "continuation_contract_reason",
		"ready_nodes_or_exact_next_action_remain",
		"ready_nodes_remain",
		"exact_next_action_remains",
		"final response allowed by recommendation readback",
		"blocked_hard_blocker",
		"blocked_lease_timing_missing",
		"blocked_minimum_minutes_unmet",
		"blocked_ready_nodes_remain",
		"blocked_no_executable_ready_node",
	)
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

func TestMissionRecommendationsReadbackCLIWritesWorkgraphReadinessPacket(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	importDir := filepath.Join(dir, "recommendations-out")
	readbackPath := filepath.Join(dir, "readback.json")
	packetPath := filepath.Join(dir, "workgraph-readiness-packet.json")
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
		"--out-workgraph-readiness-packet", packetPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation readback failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "workgraph_readiness_packet=") {
		t.Fatalf("readback output missing workgraph readiness packet path: %s", out.String())
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)
	packet := mustLoadJSON[AtlasRecommendationWorkgraphReadinessPacket](t, packetPath)
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(packet, readback); err != nil {
		t.Fatalf("bad workgraph readiness packet: %v", err)
	}
	if packet.NodeBudget != 40 ||
		packet.ContinueIfFastTarget != 40 ||
		packet.ReadyNodes != 40 ||
		packet.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		packet.ContinuationContractReason != readback.ContinuationContract.Reason ||
		packet.FinalResponseAllowed {
		t.Fatalf("packet lost 40-node ready workgraph denial: %#v", packet)
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
		checkpoint.ContinuationContractReason != "ready_nodes_or_exact_next_action_remain" ||
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
	readbackExecutionPath := filepath.Join(dir, "readback-execution-readback.json")
	out.Reset()
	code = Run([]string{
		"mission", "recommendations", "readback",
		"--wave", filepath.Join(importDir, "recommendation-wave.json"),
		"--workgraph", updatedWorkgraphPath,
		"--evidence-root", "docs/evidence/test-wave",
		"--started-at", "2026-07-04T08:00:00-07:00",
		"--completed-at", "2026-07-04T08:25:00-07:00",
		"--elapsed-minutes", "25",
		"--lease-timing-mode", "actual",
		"--out-execution-readback", readbackExecutionPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("readback execution output failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "execution_readback="+filepath.ToSlash(readbackExecutionPath)) {
		t.Fatalf("readback output missing execution readback path: %s", out.String())
	}
	readbackExecution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, readbackExecutionPath)
	if readbackExecution.ReasonArtifactAgreementSummary.Status != "agreement" ||
		readbackExecution.ReasonArtifactAgreementSummary.SourceArtifactCount != len(readbackExecution.SourceArtifacts) ||
		!readbackExecution.ReasonArtifactAgreementSummary.SourceArtifactsAgree {
		t.Fatalf("readback command execution output missing agreement summary: %#v", readbackExecution)
	}
	resumeReadback := mustLoadJSON[AtlasRecommendationReadback](t, resumeReadbackPath)
	if resumeReadback.StartedAt != leaseStart.StartedAt || resumeReadback.ElapsedMinutes != 25 {
		t.Fatalf("resume readback lost lease start: %#v", resumeReadback)
	}
	resumeExecution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, resumeExecutionPath)
	if resumeExecution.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.ReturnGateStatus != resumeReadback.ReturnGateStatus ||
		resumeExecution.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		resumeExecution.ExactNextAction != resumeReadback.ExactNextAction ||
		resumeExecution.FinalResponseAllowed != resumeReadback.FinalResponseAllowed ||
		resumeExecution.RefusesFinalResponse != resumeReadback.ContinuationContract.RefusesFinalResponse ||
		resumeExecution.GeneratedWorkgraph.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.GeneratedWorkgraph.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.CompletedRunLinks != resumeReadback.CompletedNodes ||
		resumeExecution.FoundryRunLinkReadinessSummary.RequiredRunLinks != resumeReadback.TotalNodes ||
		resumeExecution.FoundryRunLinkReadinessSummary.NextExecutableNode != resumeReadback.FirstExecutableNode ||
		resumeExecution.FoundryRunLinkReadinessSummary.ReturnGateStatus != resumeReadback.ReturnGateStatus ||
		resumeExecution.FoundryRunLinkReadinessSummary.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		resumeExecution.FoundryRunLinkReadinessSummary.ExactNextAction != resumeReadback.ExactNextAction ||
		resumeExecution.FoundryRunLinkReadinessSummary.RefusesFinalResponse != resumeReadback.ContinuationContract.RefusesFinalResponse ||
		resumeExecution.ContinuationReasonCoverage.ExpectedReason != resumeReadback.ContinuationContract.Reason ||
		resumeExecution.ContinuationReasonCoverage.SourceCount != 13 ||
		!containsString(resumeExecution.ContinuationReasonCoverage.IndexedSources, "checkpoint_readback") ||
		!containsString(resumeExecution.ContinuationReasonCoverage.IndexedSources, "resume_prompt") ||
		resumeExecution.ReasonArtifactAgreementSummary.Status != "agreement" ||
		resumeExecution.ReasonArtifactAgreementSummary.ExpectedReason != resumeReadback.ContinuationContract.Reason ||
		resumeExecution.ReasonArtifactAgreementSummary.SourceCount != resumeExecution.ContinuationReasonCoverage.SourceCount ||
		!resumeExecution.ReasonArtifactAgreementSummary.AllRequiredSourcesIndexed ||
		!resumeExecution.ReasonArtifactAgreementSummary.SourceArtifactsAgree ||
		resumeExecution.ReasonArtifactAgreementSummary.SourceArtifactCount != len(resumeExecution.SourceArtifacts) ||
		!containsString(resumeExecution.ReasonArtifactAgreementSummary.SourceArtifactRefs, "continuation_reason_coverage") ||
		!containsString(resumeExecution.ReasonArtifactAgreementSummary.SourceArtifactRefs, "foundry_run_link_readiness_summary") ||
		!hasSourceArtifact(resumeExecution.SourceArtifacts, "continuation_reason_coverage") ||
		!hasSourceArtifact(resumeExecution.SourceArtifacts, "foundry_run_link_readiness_summary") {
		t.Fatalf("execution readback missing source artifact coverage: %#v", resumeExecution)
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
		!strings.Contains(rawPromoter["no_promotion_reason_summary"].(string), "continuation_contract_reason="+resumeReadback.ContinuationContract.Reason) ||
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
		binding["return_gate_status"] != resumeReadback.ReturnGateStatus ||
		binding["continuation_contract_reason"] != resumeReadback.ContinuationContract.Reason {
		t.Fatalf("command readback missing structured timeline binding: %#v", rawCommand)
	}
	if command.ElapsedMinutes != 25 || command.FinalResponseAllowed ||
		command.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		command.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		command.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		!strings.Contains(command.CompactTimeline, "continuation_contract_reason="+resumeReadback.ContinuationContract.Reason) ||
		!strings.Contains(command.CompactTimeline, "exact_next_action="+resumeReadback.ExactNextAction) ||
		command.CommandTimelineBinding.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		command.CommandTimelineBinding.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		command.CommandTimelineBinding.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		promoter.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		promoter.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		promoter.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		promoter.PromotionClaimed || !promoter.RSIRemainsDenied ||
		foundry.NodeCompletionStatus != "nodes_in_progress" ||
		foundry.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		foundry.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		foundry.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		foundry.LeaseCompletionStatus != "minimum_minutes_unmet" {
		t.Fatalf("bad resume closure artifacts: command=%#v promoter=%#v foundry=%#v", command, promoter, foundry)
	}
	if reconciliation.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		reconciliation.CheckpointCount != 1 ||
		reconciliation.FinalStateReconciliation.ContractVersion != "ao.atlas.final-state-reconciliation.v0.1" ||
		reconciliation.FinalStateReconciliation.Status != reconciliation.Status ||
		reconciliation.FinalStateReconciliation.WorkgraphStatus != resumeReadback.Status ||
		reconciliation.FinalStateReconciliation.FoundryRollupStatus != foundry.Status ||
		reconciliation.FinalStateReconciliation.PromoterVerdictStatus != promoter.Status ||
		reconciliation.FinalStateReconciliation.CommandReadbackStatus != command.Status ||
		reconciliation.FinalStateReconciliation.ExactNextAction != resumeReadback.ExactNextAction ||
		reconciliation.FinalStateReconciliation.ContinuationReason != resumeReadback.ContinuationContract.Reason ||
		!reconciliation.FinalStateReconciliation.ContinuationAgreement ||
		reconciliation.FinalStateReconciliation.SchedulesWork ||
		reconciliation.LeaseHealthStatus != resumeReadback.LeaseHealthStatus ||
		reconciliation.CheckpointFreshnessStatus != resumeReadback.CheckpointFreshnessStatus ||
		reconciliation.StaleRouteDecisionStatus != resumeReadback.StaleRouteDecisionStatus ||
		reconciliation.ContinuationContractReason != resumeReadback.ContinuationContract.Reason ||
		reconciliation.CommandContinuationReason != resumeReadback.ContinuationContract.Reason ||
		reconciliation.PromoterContinuationReason != resumeReadback.ContinuationContract.Reason ||
		reconciliation.FoundryContinuationReason != resumeReadback.ContinuationContract.Reason ||
		!reconciliation.ContinuationReasonAgreement ||
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
		"Continuation contract reason: `" + resumeReadback.ContinuationContract.Reason + "`",
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

func TestRecommendationReadbackRejectsReadyWorkgraphFinalGateDrift(t *testing.T) {
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
	readback, err := BuildAtlasRecommendationReadback(result.Wave, partial, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T07:20:00-07:00",
		CompletedAt:     "2026-07-04T09:20:00-07:00",
		ElapsedMinutes:  120,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.ReadyNodes == 0 || readback.FirstExecutableNode != "mission-recommendation-next-31" || readback.FinalResponseAllowed {
		t.Fatalf("test setup expected ready continuation readback: %#v", readback)
	}

	tampered := readback
	tampered.ReturnGateStatus = "final_response_allowed"
	tampered.ExactNextActionReadback.ReturnGateStatus = tampered.ReturnGateStatus
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "ready nodes require return_gate_status=blocked_ready_nodes_remain") {
		t.Fatalf("expected stale return gate rejection, got %v", err)
	}

	tampered = readback
	tampered.FinalResponseReason = "all generated nodes complete and no ready nodes remain"
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "ready nodes require final_response_reason=ready nodes or exact next actions remain") {
		t.Fatalf("expected stale final reason rejection, got %v", err)
	}

	tampered = readback
	tampered.ExactNextAction = "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks."
	tampered.ExactNextActionReadback.Action = tampered.ExactNextAction
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "ready nodes require exact_next_action to name first_executable_node") {
		t.Fatalf("expected stale exact next action rejection, got %v", err)
	}
}

func TestRecommendationReadbackRejectsCompletedWorkgraphFinalAllowanceDrift(t *testing.T) {
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
		StartedAt:       "2026-07-04T07:20:00-07:00",
		CompletedAt:     "2026-07-04T09:20:00-07:00",
		ElapsedMinutes:  120,
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !readback.FinalResponseAllowed || readback.ReadyNodes != 0 || readback.ReturnGateStatus != "final_response_allowed" {
		t.Fatalf("test setup expected completed final-allowed readback: %#v", readback)
	}

	tampered := readback
	tampered.Status = "in_progress"
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "final_response_allowed requires status=completed") {
		t.Fatalf("expected stale completed status rejection, got %v", err)
	}

	tampered = readback
	tampered.ReturnGateStatus = "blocked_ready_nodes_remain"
	tampered.ExactNextActionReadback.ReturnGateStatus = tampered.ReturnGateStatus
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "final_response_allowed requires return_gate_status=final_response_allowed") {
		t.Fatalf("expected stale return gate rejection, got %v", err)
	}

	tampered = readback
	tampered.FinalResponseReason = "ready nodes or exact next actions remain"
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "final_response_allowed requires final_response_reason=all generated nodes complete and no ready nodes remain") {
		t.Fatalf("expected stale final reason rejection, got %v", err)
	}

	tampered = readback
	tampered.ExactNextAction = "Emit Foundry import for mission-recommendation-next-40 and execute exactly one active node."
	tampered.ExactNextActionReadback.Action = tampered.ExactNextAction
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "final_response_allowed requires final exact_next_action") {
		t.Fatalf("expected stale exact next action rejection, got %v", err)
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
	if readback.ReadyNodes != 0 ||
		readback.ContinuationContract.Status != "continuation_required" ||
		!readback.ContinuationContract.RefusesFinalResponse ||
		readback.ContinuationContract.Reason != "exact_next_action_remains" ||
		readback.ContinuationContract.ExactNextAction != readback.ExactNextAction {
		t.Fatalf("exact next action must keep final response denied after a short completed wave: %#v", readback.ContinuationContract)
	}
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("execution ledger should stay consistent for early timing denial: %v", err)
	}
	if execution.Status == "completed" {
		t.Fatalf("execution ledger cannot be completed before min_minutes: %#v", execution)
	}

	tampered := readback
	tampered.ContinuationContract.Reason = "ready_nodes_or_exact_next_action_remain"
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil ||
		!strings.Contains(err.Error(), "continuation_contract.reason must be exact_next_action_remains") {
		t.Fatalf("expected exact-next-action continuation reason rejection, got %v", err)
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
	command.Status = "completed"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback status disagrees") {
		t.Fatalf("expected stale command status rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	foundry.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "foundry rollup continuation_contract_reason disagrees") {
		t.Fatalf("expected stale foundry continuation reason rejection, got %v", err)
	}
	foundry = BuildAtlasRecommendationFoundryRollup(readback)
	command = BuildAtlasRecommendationCommandReadback(readback)
	promoter.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "promoter readback continuation_contract_reason disagrees") {
		t.Fatalf("expected stale promoter continuation reason rejection, got %v", err)
	}
	promoter = BuildAtlasRecommendationPromoterReadback(readback)
	promoter.NoPromotionReasonSummary = "No authority promotion claimed; RSI remains denied."
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "promoter readback no_promotion_reason_summary must include continuation_contract_reason") {
		t.Fatalf("expected stale promoter reason summary rejection, got %v", err)
	}
	promoter = BuildAtlasRecommendationPromoterReadback(readback)
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CommandTimelineBinding.ExactNextAction = "stale next action"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command timeline binding exact_next_action disagrees") {
		t.Fatalf("expected stale command timeline binding rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback continuation_contract_reason disagrees") {
		t.Fatalf("expected stale command continuation reason rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CommandTimelineBinding.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command timeline binding continuation_contract_reason disagrees") {
		t.Fatalf("expected stale command timeline continuation reason rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CompactTimeline = strings.Replace(command.CompactTimeline, "; continuation_contract_reason="+readback.ContinuationContract.Reason, "", 1)
	command.CommandTimelineBinding.Summary = command.CompactTimeline
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command compact timeline continuation_contract_reason missing") {
		t.Fatalf("expected stale command compact continuation reason rejection, got %v", err)
	}
	command = BuildAtlasRecommendationCommandReadback(readback)
	command.CompactTimeline = strings.Replace(command.CompactTimeline, "; exact_next_action="+readback.ExactNextAction, "", 1)
	command.CommandTimelineBinding.Summary = command.CompactTimeline
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command compact timeline exact_next_action missing") {
		t.Fatalf("expected stale command compact exact next action rejection, got %v", err)
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
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.FinalStateReconciliation.CommandReadbackStatus = "stale_command_readback"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "final_state_reconciliation.command_readback_status disagrees") {
		t.Fatalf("expected stale final-state command rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.FinalStateReconciliation.FoundryRollupStatus = "stale_foundry_rollup"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "final_state_reconciliation.foundry_rollup_status disagrees") {
		t.Fatalf("expected stale final-state Foundry rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "reconciliation continuation_contract_reason disagrees") {
		t.Fatalf("expected stale reconciliation continuation reason rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.PromoterContinuationReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "reconciliation promoter_continuation_contract_reason disagrees") {
		t.Fatalf("expected stale reconciliation Promoter continuation reason rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.ContinuationReasonAgreement = false
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "reconciliation continuation_reason_agreement disagrees") {
		t.Fatalf("expected stale reconciliation continuation agreement rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.FinalStateReconciliation.Status = "ready"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "final_state_reconciliation.status must match reconciliation status") {
		t.Fatalf("expected stale final-state status rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.FinalStateReconciliation.ContinuationReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "final_state_reconciliation.continuation_contract_reason disagrees") {
		t.Fatalf("expected stale final-state continuation reason rejection, got %v", err)
	}
	packet = BuildAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry)
	packet.FinalStateReconciliation.ContinuationAgreement = false
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, packet); err == nil ||
		!strings.Contains(err.Error(), "final_state_reconciliation.continuation_reason_agreement disagrees") {
		t.Fatalf("expected stale final-state continuation agreement rejection, got %v", err)
	}
}

func TestRecommendationReconciliationStaleCommandStatusFixture(t *testing.T) {
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
	inProgress := completeRecommendationNodes(result.Workgraph, 1)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, inProgress, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T08:00:00-07:00",
		CompletedAt:     "2026-07-04T08:25:00-07:00",
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	command := BuildAtlasRecommendationCommandReadback(readback)
	promoter := BuildAtlasRecommendationPromoterReadback(readback)
	foundry := BuildAtlasRecommendationFoundryRollup(readback)
	command.Status = "completed"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback status disagrees") {
		t.Fatalf("expected stale command status closure rejection, got %v", err)
	}

	fixture := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, filepath.Join(repoRoot(t), "examples", "invalid", "recommendation-reconciliation-stale-command-status.json"))
	if fixture.Status != "blocked_stale_artifact" ||
		fixture.ArtifactsAgree ||
		fixture.FinalStateReconciliation.CommandReadbackStatus != "completed" ||
		fixture.FinalStateReconciliation.ContinuationReason != readback.ContinuationContract.Reason ||
		!fixture.FinalStateReconciliation.ContinuationAgreement ||
		fixture.CommandFinalResponseAllowed ||
		fixture.FinalResponseAllowed ||
		!fixture.ContinuationReasonAgreement ||
		fixture.ContinuationContractReason != readback.ContinuationContract.Reason ||
		fixture.ReadyNodes != 39 ||
		fixture.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		fixture.ExactNextAction != readback.ExactNextAction ||
		fixture.RSIRemainsDenied != true {
		t.Fatalf("stale Command fixture lost mismatch readback: %#v", fixture)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, fixture); err != nil {
		t.Fatalf("stale Command fixture should validate as blocked stale artifact: %v", err)
	}
}

func TestRecommendationReconciliationStaleContinuationReasonFixture(t *testing.T) {
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
	inProgress := completeRecommendationNodes(result.Workgraph, 1)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, inProgress, AtlasRecommendationReadbackOptions{
		StartedAt:       "2026-07-04T08:00:00-07:00",
		CompletedAt:     "2026-07-04T08:25:00-07:00",
		LeaseTimingMode: "actual",
	})
	if err != nil {
		t.Fatal(err)
	}
	command := BuildAtlasRecommendationCommandReadback(readback)
	promoter := BuildAtlasRecommendationPromoterReadback(readback)
	foundry := BuildAtlasRecommendationFoundryRollup(readback)
	command.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err == nil ||
		!strings.Contains(err.Error(), "command readback continuation_contract_reason disagrees") {
		t.Fatalf("expected stale command continuation reason rejection, got %v", err)
	}

	fixture := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, filepath.Join(repoRoot(t), "examples", "invalid", "recommendation-reconciliation-stale-continuation-reason.json"))
	if fixture.Status != "blocked_stale_artifact" ||
		fixture.ArtifactsAgree ||
		fixture.ContinuationReasonAgreement ||
		fixture.ContinuationContractReason != readback.ContinuationContract.Reason ||
		fixture.CommandContinuationReason != "ready_nodes_remain" ||
		fixture.PromoterContinuationReason != readback.ContinuationContract.Reason ||
		fixture.FoundryContinuationReason != readback.ContinuationContract.Reason ||
		fixture.FinalStateReconciliation.ContinuationAgreement ||
		fixture.FinalStateReconciliation.CommandReadbackStatus != command.Status ||
		fixture.ReadyNodes != 39 ||
		fixture.ExactNextAction != readback.ExactNextAction ||
		fixture.RSIRemainsDenied != true {
		t.Fatalf("stale continuation reason fixture lost mismatch readback: %#v", fixture)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, fixture); err != nil {
		t.Fatalf("stale continuation reason fixture should validate as blocked stale artifact: %v", err)
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

func TestRecommendationReadbackRejectsMissingPromoterNoPromotionPlaceholders(t *testing.T) {
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
	readback, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}

	readback.PromoterNoPromotionPlaceholders = nil
	if err := ValidateAtlasRecommendationReadback(readback); err == nil ||
		!strings.Contains(err.Error(), "promoter_no_promotion_placeholders must include promotion_claim, rsi_boundary, and authority_advance") {
		t.Fatalf("expected missing promoter no-promotion placeholders rejection, got %v", err)
	}

	readback, err = BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	readback.PromoterNoPromotionPlaceholders[0].Status = "stale"
	if err := ValidateAtlasRecommendationReadback(readback); err == nil ||
		!strings.Contains(err.Error(), "promoter_no_promotion_placeholders.promotion_claim status must be pending_promoter_no_promotion") {
		t.Fatalf("expected stale promoter no-promotion placeholder rejection, got %v", err)
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
	execution.ContinuationContractReason = "ready_nodes_remain"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "continuation_contract_reason must match recommendation readback") {
		t.Fatalf("expected stale execution continuation reason rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.ExactNextAction = "stale exact next action"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "exact_next_action must match recommendation readback") {
		t.Fatalf("expected stale execution exact next action rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.RefusesFinalResponse = !readback.ContinuationContract.RefusesFinalResponse
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "refuses_final_response must match recommendation readback") {
		t.Fatalf("expected stale execution final refusal rejection, got %v", err)
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
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.ContinuationContractReason = "ready_nodes_remain"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness continuation_contract_reason must match recommendation readback") {
		t.Fatalf("expected stale Foundry continuation reason rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.ExactNextAction = "stale exact next action"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness exact_next_action must match recommendation readback") {
		t.Fatalf("expected stale Foundry exact next action rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.FoundryRunLinkReadinessSummary.RefusesFinalResponse = !readback.ContinuationContract.RefusesFinalResponse
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "foundry run-link readiness refuses_final_response must match recommendation readback") {
		t.Fatalf("expected stale Foundry refusal rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.ContinuationReasonCoverage.ExpectedReason = "ready_nodes_remain"
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "continuation_reason_coverage.expected_reason must match recommendation readback") {
		t.Fatalf("expected stale continuation reason coverage rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.ContinuationReasonCoverage.IndexedSources = execution.ContinuationReasonCoverage.IndexedSources[:12]
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "continuation_reason_coverage.source_count must match indexed_sources length") {
		t.Fatalf("expected stale continuation source count rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	for i := range execution.SourceArtifacts {
		if execution.SourceArtifacts[i].Ref == "continuation_reason_coverage" {
			execution.SourceArtifacts[i].Digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
		}
	}
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "continuation_reason_coverage source artifact digest disagrees") {
		t.Fatalf("expected stale continuation source artifact digest rejection, got %v", err)
	}
	execution = BuildAtlasRecommendationExecutionReadback(readback)
	execution.ReasonArtifactAgreementSummary.SourceArtifactCount = 1
	err = ValidateAtlasRecommendationExecutionReadback(execution, readback)
	if err == nil || !strings.Contains(err.Error(), "reason_artifact_agreement_summary.source_artifact_count must match source_artifacts length") {
		t.Fatalf("expected stale reason artifact summary rejection, got %v", err)
	}
}

func TestRecommendationWorkgraphReadinessPacketRejectsStaleReadyNodeState(t *testing.T) {
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
	readback, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := BuildAtlasRecommendationWorkgraphReadinessPacket(readback, AtlasRecommendationWorkgraphReadinessPacketOptions{})
	if err != nil {
		t.Fatal(err)
	}

	stale := packet
	stale.ReadyNodes = 0
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(stale, readback); err == nil ||
		!strings.Contains(err.Error(), "ready_nodes must match recommendation readback") {
		t.Fatalf("expected stale ready node rejection, got %v", err)
	}

	stale = packet
	stale.OneExecutableMutationNodeActive = false
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(stale, readback); err == nil ||
		!strings.Contains(err.Error(), "ready nodes require one_executable_mutation_node_active=true") {
		t.Fatalf("expected missing one-active-node rejection, got %v", err)
	}

	stale = packet
	stale.FinalResponseAllowed = true
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(stale, readback); err == nil ||
		!strings.Contains(err.Error(), "final_response_allowed must match recommendation readback") {
		t.Fatalf("expected stale final response rejection, got %v", err)
	}

	stale = packet
	stale.ContinuationContractReason = "ready_nodes_remain"
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(stale, readback); err == nil ||
		!strings.Contains(err.Error(), "continuation_contract_reason must match recommendation readback") {
		t.Fatalf("expected stale continuation reason rejection, got %v", err)
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
	execution := mustLoadJSON[AtlasRecommendationExecutionReadback](t, filepath.Join(root, "execution-readback.json"))
	command := mustLoadJSON[AtlasRecommendationCommandReadback](t, filepath.Join(root, "command-readback.json"))
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, filepath.Join(root, "promoter-readback.json"))
	foundry := mustLoadJSON[AtlasRecommendationFoundryRollup](t, filepath.Join(root, "foundry-rollup.json"))
	reconciliation := mustLoadJSON[AtlasRecommendationReconciliationPacket](t, filepath.Join(root, "reconciliation-packet.json"))
	var synthesis struct {
		CompletedNodes             int    `json:"completed_nodes"`
		TotalNodes                 int    `json:"total_nodes"`
		ReadyNodes                 int    `json:"ready_nodes"`
		CheckpointCount            int    `json:"checkpoint_count"`
		ElapsedMinutes             int    `json:"elapsed_minutes"`
		ReturnGateStatus           string `json:"return_gate_status"`
		FinalResponseAllowed       bool   `json:"final_response_allowed"`
		ExactNextAction            string `json:"exact_next_action"`
		ContinuationContractReason string `json:"continuation_contract_reason"`
		RefusesFinalResponse       bool   `json:"refuses_final_response"`
		NextRecommendedPrompt      string `json:"next_recommended_prompt"`
	}
	synthesis = mustLoadJSON[struct {
		CompletedNodes             int    `json:"completed_nodes"`
		TotalNodes                 int    `json:"total_nodes"`
		ReadyNodes                 int    `json:"ready_nodes"`
		CheckpointCount            int    `json:"checkpoint_count"`
		ElapsedMinutes             int    `json:"elapsed_minutes"`
		ReturnGateStatus           string `json:"return_gate_status"`
		FinalResponseAllowed       bool   `json:"final_response_allowed"`
		ExactNextAction            string `json:"exact_next_action"`
		ContinuationContractReason string `json:"continuation_contract_reason"`
		RefusesFinalResponse       bool   `json:"refuses_final_response"`
		NextRecommendedPrompt      string `json:"next_recommended_prompt"`
	}](t, filepath.Join(root, "final-synthesis.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("final execution readback should agree with lease-resume readback: %v", err)
	}
	if err := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry); err != nil {
		t.Fatalf("final closure artifacts should agree with lease-resume readback: %v", err)
	}
	if err := ValidateAtlasRecommendationReconciliationPacket(readback, command, promoter, foundry, reconciliation); err != nil {
		t.Fatalf("final reconciliation packet should agree with lease-resume readback: %v", err)
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
		synthesis.ExactNextAction != readback.ExactNextAction ||
		synthesis.ContinuationContractReason != readback.ContinuationContract.Reason ||
		synthesis.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse {
		t.Fatalf("final synthesis does not match root readback: synthesis=%#v readback=%#v", synthesis, readback)
	}
	if execution.ContinuationContractReason != readback.ContinuationContract.Reason ||
		execution.RefusesFinalResponse != readback.ContinuationContract.RefusesFinalResponse ||
		execution.FoundryRunLinkReadinessSummary.ContinuationContractReason != readback.ContinuationContract.Reason ||
		execution.ContinuationReasonCoverage.ExpectedReason != readback.ContinuationContract.Reason {
		t.Fatalf("execution readback lost lease-resume continuation reason: execution=%#v readback=%#v", execution, readback)
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
	nextExecutableNode := readback.FirstExecutableNode
	if nextExecutableNode == "" {
		nextExecutableNode = "none"
	}
	for _, want := range []string{
		"Current workgraph: `" + workgraphRef + "`",
		"Completed nodes: " + strconv.Itoa(readback.CompletedNodes) + " / " + strconv.Itoa(readback.TotalNodes),
		"Ready nodes: " + strconv.Itoa(readback.ReadyNodes),
		"Elapsed minutes at latest checkpoint: " + strconv.Itoa(readback.ElapsedMinutes),
		"`final_response_allowed=" + strconv.FormatBool(readback.FinalResponseAllowed) + "`",
		"Return gate: `" + readback.ReturnGateStatus + "`",
		"Continuation contract reason: `" + readback.ContinuationContract.Reason + "`",
		"Early-return risk: `" + readback.EarlyReturnRiskStatus + "`",
		"Checkpoint count: " + strconv.Itoa(readback.CheckpointCount),
		"Next executable node: `" + nextExecutableNode + "`",
		readback.ExactNextAction,
		"If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("continuation prompt missing final-state evidence %q:\n%s", want, prompt)
		}
	}
}

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
		"execution-readback-regenerated.json",
		"reason_artifact_agreement_summary",
		"generated-recommendation-prompt-continuation-reason-negative-scan",
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

func TestProductionReadinessRejectsUnsafeRecommendationPromptContinuationReasonFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "examples", "invalid", "recommendation-prompt-unsafe-continuation-reason.md")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read unsafe continuation reason prompt fixture: %v", err)
	}
	if !strings.Contains(string(fixture), "Continuation contract reason: `fully_unsupervised_complex_mutation is proven`") {
		t.Fatalf("unsafe prompt fixture must poison the continuation reason line:\n%s", string(fixture))
	}

	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"examples/invalid/recommendation-prompt-unsafe-continuation-reason.md",
		"unsafe_recommendation_reason_scan",
		"unsafe generated recommendation prompt continuation reason was accepted",
		"generated recommendation prompt contains unsafe wording",
		"generated-recommendation-prompt-continuation-reason-negative-scan",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing unsafe recommendation prompt fixture coverage %q", want)
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
