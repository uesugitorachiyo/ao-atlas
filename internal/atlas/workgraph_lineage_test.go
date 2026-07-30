package atlas

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteWorkgraphPreservesCampaignLineage(t *testing.T) {
	dir := t.TempDir()
	inputPath := writeCampaignLineageWorkgraphFixture(t, dir)
	workgraph, err := LoadJSON[Workgraph](inputPath)
	if err != nil {
		t.Fatal(err)
	}
	_, link := buildCampaignLineageRunLink(t, dir)

	completed, nodeID, err := CompleteWorkgraph(workgraph, link)
	if err != nil {
		t.Fatal(err)
	}
	if nodeID != "readiness-ready" {
		t.Fatalf("completed node = %q, want readiness-ready", nodeID)
	}
	if workgraph.Nodes[1].Status != "ready" {
		t.Fatalf("input workgraph was mutated: status=%q", workgraph.Nodes[1].Status)
	}
	if completed.Nodes[1].Status != "completed" {
		t.Fatalf("completed workgraph status=%q, want completed", completed.Nodes[1].Status)
	}
	body, err := json.Marshal(completed)
	if err != nil {
		t.Fatal(err)
	}
	assertCampaignLineage(t, body)
}

func TestWorkgraphCompleteCLIPreservesCampaignLineage(t *testing.T) {
	dir := t.TempDir()
	inputPath := writeCampaignLineageWorkgraphFixture(t, dir)
	evidenceRoot, link := buildCampaignLineageRunLink(t, dir)
	runLinkPath := filepath.Join(dir, "run-link.json")
	if err := WriteJSON(runLinkPath, link); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "workgraph-after.json")
	var output bytes.Buffer
	code := Run([]string{
		"workgraph", "complete",
		"--workgraph", inputPath,
		"--run-link", runLinkPath,
		"--evidence-root", evidenceRoot,
		"--evidence-root-id", "evidence-lineage-regression",
		"--out", outputPath,
	}, &output, &output)
	if code != 0 {
		t.Fatalf("workgraph complete failed: %s", output.String())
	}
	completedBody, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	assertCampaignLineage(t, completedBody)

	inputBody, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	assertCampaignLineage(t, inputBody)
	var input Workgraph
	if err := json.Unmarshal(inputBody, &input); err != nil {
		t.Fatal(err)
	}
	if input.Nodes[1].Status != "ready" {
		t.Fatalf("CLI mutated input workgraph: status=%q", input.Nodes[1].Status)
	}
}

func writeCampaignLineageWorkgraphFixture(t *testing.T, dir string) string {
	t.Helper()
	sourcePath := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	body, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	for key, value := range campaignLineageValues() {
		document[key] = value
	}
	inputPath := filepath.Join(dir, "workgraph-before.json")
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(inputPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	return inputPath
}

func buildCampaignLineageRunLink(t *testing.T, dir string) (string, RunLink) {
	t.Helper()
	evidenceRoot := filepath.Join(dir, "evidence")
	evidencePath := filepath.Join(evidenceRoot, "nodes", "readiness-ready.json")
	if err := os.MkdirAll(filepath.Dir(evidencePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidencePath, []byte("{\"status\":\"passed\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link, err := BuildEvidenceBoundRunLink(
		"atlas-readiness-task",
		"completed",
		map[string]string{"node": "nodes/readiness-ready.json"},
		evidenceRoot,
		"evidence-lineage-regression",
	)
	if err != nil {
		t.Fatal(err)
	}
	return evidenceRoot, link
}

func assertCampaignLineage(t *testing.T, body []byte) {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	for key, want := range campaignLineageValues() {
		if got, ok := document[key].(string); !ok || got != want {
			t.Errorf("%s=%q, want %q", key, got, want)
		}
	}
}

func campaignLineageValues() map[string]string {
	return map[string]string{
		"mission_id":             "mission-lineage-regression",
		"objective_id":           "objective-lineage-regression",
		"objective_digest":       "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"correlation_id":         "correlation-lineage-regression",
		"soak_id":                "soak-lineage-regression",
		"plan_id":                "plan-lineage-regression",
		"policy_digest":          "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		"activation_id":          "activation-lineage-regression",
		"evidence_root_identity": "evidence-lineage-regression",
		"mission_source_head":    "3333333333333333333333333333333333333333",
		"atlas_source_head":      "4444444444444444444444444444444444444444",
		"mission_binary_sha256":  "sha256:5555555555555555555555555555555555555555555555555555555555555555",
		"atlas_binary_sha256":    "sha256:6666666666666666666666666666666666666666666666666666666666666666",
	}
}
