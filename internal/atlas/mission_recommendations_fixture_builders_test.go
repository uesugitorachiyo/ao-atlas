package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecommendationTestFixtureBuildersCoverWaveNodeAndReadbackDomains(t *testing.T) {
	bundlePath := filepath.Join(t.TempDir(), "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, bundlePath, 2, false)
	bundle := mustLoadJSON[map[string]any](t, bundlePath)
	if bundle["recommendation_count"].(float64) != 2 || bundle["safe_to_execute"].(bool) {
		t.Fatalf("feature-depth bundle builder did not preserve bounded defaults: %#v", bundle)
	}

	workgraph := Workgraph{
		Nodes: []WorkgraphNode{
			{ID: "mission-recommendation-next-01", Status: "ready"},
			{ID: "mission-recommendation-next-02", Status: "ready"},
		},
	}
	completed := completeRecommendationNodes(workgraph, 1)
	if completed.Nodes[0].Status != "completed" || completed.Nodes[1].Status != "ready" || workgraph.Nodes[0].Status != "ready" {
		t.Fatalf("completeRecommendationNodes should complete a copy without mutating the source: %#v source=%#v", completed.Nodes, workgraph.Nodes)
	}

	evidence := recommendationEvidenceFiles(t, "fixture-builder-regression", "mission-recommendation-next-01")
	runLink := recommendationRunLink(t, "mission-recommendation-next-01-task", evidence)
	if runLink.TaskID != "mission-recommendation-next-01-task" ||
		runLink.Status != "completed" ||
		len(evidence) != 11 ||
		evidence["node_gate"] == "" ||
		runLink.Evidence["node_gate"] != evidence["node_gate"] ||
		!strings.HasPrefix(runLink.Digest, "sha256:") {
		t.Fatalf("recommendation evidence builders did not produce node/run-link coverage: evidence=%#v runLink=%#v", evidence, runLink)
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
