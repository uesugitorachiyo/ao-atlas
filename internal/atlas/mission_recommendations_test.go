package atlas

import (
	"path/filepath"
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

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
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

func assertSchemaHasProperty(t *testing.T, schemaPath, field string) {
	t.Helper()
	schema := mustLoadJSON[requiredFieldSchema](t, schemaPath)
	if _, ok := schema.Properties[field]; !ok {
		t.Fatalf("schema %s does not define property %q", schemaPath, field)
	}
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
