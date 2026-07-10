package atlas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecommendationEvidenceValidationRequiresProvenanceFieldsWhenEnabled(t *testing.T) {
	root := t.TempDir()
	nodeRoot := filepath.Join(root, "nodes")
	nodeDir := filepath.Join(nodeRoot, "node-01")
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nodeDir, "node_gate.json")
	if err := os.WriteFile(path, []byte(`{"schema":"ao.atlas.consolidation-node-gate.v0.1"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entry := validateRecommendationEvidenceJSONFileWithValidationOptions(root, nodeRoot, path, AtlasRecommendationEvidenceValidationOptions{
		RequireProvenanceFields: true,
	})
	if entry.Status != "failed" {
		t.Fatalf("missing provenance fields should fail validation: %#v", entry)
	}
	if entry.Error != "missing source_digest; missing evidence_class" {
		t.Fatalf("unexpected provenance error: %q", entry.Error)
	}
}

func TestRecommendationEvidenceValidationAcceptsProvenanceFieldsWhenEnabled(t *testing.T) {
	root := t.TempDir()
	nodeRoot := filepath.Join(root, "nodes")
	nodeDir := filepath.Join(nodeRoot, "node-01")
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nodeDir, "node_gate.json")
	if err := os.WriteFile(path, []byte(`{
  "schema": "ao.atlas.consolidation-node-gate.v0.1",
  "source_digest": "sha256:e776ea250456c59d92eb6758e6fbf3ac1abdc7ad821c08d9cc9b520a390b4b5e",
  "evidence_class": "node_gate"
}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entry := validateRecommendationEvidenceJSONFileWithValidationOptions(root, nodeRoot, path, AtlasRecommendationEvidenceValidationOptions{
		RequireProvenanceFields: true,
	})
	if entry.Status != "passed" {
		t.Fatalf("provenance-bound evidence should pass validation: %#v", entry)
	}
	if entry.SourceDigest == "" || entry.EvidenceClass != "node_gate" {
		t.Fatalf("provenance fields were not recorded: %#v", entry)
	}
}

func TestRecommendationEvidenceValidationReportRecordsMissingProvenanceFields(t *testing.T) {
	root := t.TempDir()
	nodeDir := filepath.Join(root, "nodes", "node-01")
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nodeDir, "node_gate.json")
	if err := os.WriteFile(path, []byte(`{"schema":"ao.atlas.consolidation-node-gate.v0.1"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := BuildAtlasRecommendationEvidenceValidationReportWithValidationOptions(root, AtlasRecommendationEvidenceValidationOptions{
		RequireProvenanceFields: true,
	})
	if err == nil {
		t.Fatal("validation report should fail when provenance fields are missing")
	}
	if report.Status != "failed" || len(report.MissingSourceDigestFiles) != 1 || len(report.MissingEvidenceClassFiles) != 1 {
		t.Fatalf("report did not record missing provenance fields: %#v", report)
	}
}
