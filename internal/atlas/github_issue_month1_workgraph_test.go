package atlas

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGitHubIssueMonth1WorkgraphKeepsClosureLockedUntilRequiredNodesPass(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "examples", "valid", "github-issue-month1-workgraph.json"))
	if err != nil {
		t.Fatal(err)
	}
	var graph map[string]any
	if err := json.Unmarshal(body, &graph); err != nil {
		t.Fatal(err)
	}
	if graph["schema"] != "ao.atlas.github-issue-to-draft-pr-month1-workgraph.v0.1" ||
		graph["status"] != "ready" ||
		graph["month"].(float64) != 1 {
		t.Fatalf("unexpected GitHub issue Month 1 workgraph identity: %#v", graph)
	}
	nodes := graph["nodes"].([]any)
	if len(nodes) != 10 {
		t.Fatalf("Month 1 workgraph must have exactly 10 required nodes, got %d", len(nodes))
	}
	seen := map[string]bool{}
	for _, raw := range nodes {
		node := raw.(map[string]any)
		if node["required"] != true {
			t.Fatalf("node must be required: %#v", node)
		}
		seen[node["id"].(string)] = true
		if node["id"] == "month1-closure-reconcile" {
			if node["status"] != "locked" {
				t.Fatalf("closure reconcile must remain locked before dependency proof: %#v", node)
			}
			dependencies := node["dependencies"].([]any)
			if len(dependencies) != 9 {
				t.Fatalf("closure reconcile must depend on the other 9 nodes: %#v", node)
			}
		}
	}
	for _, id := range []string{
		"month1-architecture-contracts",
		"month1-ao2-url-intake",
		"month1-covenant-policy",
		"month1-control-plane-observation",
		"month1-command-readback",
		"month1-sentinel-wording",
		"month1-promoter-no-promotion",
		"month1-blueprint-bounded-claim",
		"month1-mission-supervision",
		"month1-closure-reconcile",
	} {
		if !seen[id] {
			t.Fatalf("missing Month 1 node %s", id)
		}
	}
	locks := graph["stage_locks"].(map[string]any)
	if locks["month2_locked_until_month1_closure"] != true ||
		locks["feature_generated_pr_merge_node_exists"] != false {
		t.Fatalf("Month 1 stage locks widened: %#v", locks)
	}
	boundaries := graph["boundaries"].(map[string]any)
	for key, value := range boundaries {
		if value != false {
			t.Fatalf("boundary %s = %#v, want false", key, value)
		}
	}
}
