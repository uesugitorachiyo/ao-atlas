package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth2ContractConvergenceInputPreservesSourceMetadata(t *testing.T) {
	path := filepath.Join(repoRoot(t), "docs/evidence/ao-stack-contract-convergence-month2-wave-v01/month2-contract-convergence-input.json")
	bundle := mustLoadJSON[map[string]any](t, path)

	if bundle["schema"] != "ao.mission.feature-depth-recommendations.v0.3" {
		t.Fatalf("bridge must use the canonical Mission recommendation schema: %#v", bundle["schema"])
	}
	if bundle["source_contract_schema"] != "ao.atlas.consolidation-implementation-evidence.v0.1" {
		t.Fatalf("bridge lost source contract schema: %#v", bundle["source_contract_schema"])
	}
	if bundle["source_recommendations_path"] != "docs/evidence/ao-stack-consolidation-month1-wave-v01/nodes/mission-recommendation-consolidation-month1-34/month2-contract-convergence-recommendations.json" {
		t.Fatalf("bridge lost source recommendations path: %#v", bundle["source_recommendations_path"])
	}
	if bundle["source_readback_path"] != "docs/evidence/ao-stack-consolidation-month1-wave-v01/nodes/mission-recommendation-consolidation-month1-34/recommendation-readback-after.json" {
		t.Fatalf("bridge must bind the readback that generated the source recommendations: %#v", bundle["source_readback_path"])
	}
	if bundle["initial_executable_node"] != "mission-recommendation-month2-contract-convergence-1" || bundle["one_executable_node_active"] != true {
		t.Fatalf("bridge lost the serialized initial-node invariant: %#v", bundle)
	}

	tasks, ok := bundle["tasks"].([]any)
	if !ok || len(tasks) != 40 {
		t.Fatalf("bridge must preserve all 40 tasks: %#v", bundle["tasks"])
	}
	for i, raw := range tasks {
		task, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("task %d is not an object: %#v", i, raw)
		}
		for _, field := range []string{
			"source_recommendation_id",
			"source_owner",
			"source_title",
			"source_theme",
			"source_mutation_class",
			"requires_blueprint_authorization",
			"source_safety_limits",
			"source_exit_evidence",
		} {
			if _, exists := task[field]; !exists {
				t.Fatalf("task %d lost source field %q: %#v", i, field, task)
			}
		}
		if task["owner"] != "ao-atlas" {
			t.Fatalf("factory owner must remain Atlas while source owner is preserved separately: %#v", task)
		}
	}
}
