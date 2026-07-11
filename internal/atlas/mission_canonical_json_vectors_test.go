package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CanonicalJSONVectorsRecordDigestFixturesForGateRecords(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-10")
	recordedPath := filepath.Join(nodeDir, "canonical-json-vectors.json")
	outPath := filepath.Join(t.TempDir(), "canonical-json-vectors.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "canonical-json-vectors",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("canonical-json-vectors command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=canonical_json_vectors_ready",
		"vector_count=5",
		"language_count=2",
		"digest_algorithm=sha256.canonical-json.v1",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("canonical JSON vectors output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("canonical JSON vectors fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["vector_count"] != float64(5) ||
		generated["language_count"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("canonical JSON vectors fixture lost safety state: %#v", generated)
	}
}

func TestMonth3CanonicalJSONVectorsUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-10", "canonical-json-vectors.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.canonical-json-vectors.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:canonical-json-vectors" {
		t.Fatalf("expected typed canonical JSON vectors validator, got %s", validator)
	}
}

func TestMonth3CanonicalJSONVectorsRejectsDigestDrift(t *testing.T) {
	fixture, err := BuildAtlasCanonicalJSONVectors()
	if err != nil {
		t.Fatal(err)
	}
	fixture.Vectors[0].CanonicalJSON = `{"id":"mission-2","status":"active"}`
	if err := ValidateAtlasCanonicalJSONVectors(fixture); err == nil || !strings.Contains(err.Error(), "digest must match canonical_json") {
		t.Fatalf("expected digest drift rejection, got %v", err)
	}
}
