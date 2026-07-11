package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CanonicalJSONVectorSmokeChecksCoverGoAndRust(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-14")
	sourceVectorsPath := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-10", "canonical-json-vectors.json")
	recordedPath := filepath.Join(nodeDir, "canonical-json-vector-smoke-checks.json")
	outPath := filepath.Join(t.TempDir(), "canonical-json-vector-smoke-checks.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "canonical-json-vector-smoke-checks",
		"--source-vectors", sourceVectorsPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("canonical-json-vector-smoke-checks command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=smoke_checks_ready",
		"language_count=2",
		"vector_count=5",
		"smoke_check_count=10",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("smoke check output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("canonical JSON vector smoke checks changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["go_dependency_free"] != true ||
		generated["rust_dependency_free"] != true ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("smoke checks lost dependency-free or authority state: %#v", generated)
	}
}

func TestMonth3CanonicalJSONVectorSmokeChecksUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-14", "canonical-json-vector-smoke-checks.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.canonical-json-vector-smoke-checks.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:canonical-json-vector-smoke-checks" {
		t.Fatalf("expected typed canonical JSON vector smoke checks validator, got %s", validator)
	}
}

func TestMonth3CanonicalJSONVectorSmokeChecksRejectsMissingRustCoverage(t *testing.T) {
	fixture, err := BuildAtlasCanonicalJSONVectorSmokeChecks(AtlasCanonicalJSONVectors{
		Schema:          AtlasCanonicalJSONVectorsContract,
		Status:          "canonical_json_vectors_ready",
		DigestAlgorithm: "sha256.canonical-json.v1",
		VectorCount:     1,
		LanguageCount:   2,
		Languages:       []string{"go", "rust"},
		Vectors: []AtlasCanonicalJSONVector{{
			ID:            "mission-record-minimal",
			RecordClass:   "mission",
			CanonicalJSON: `{"id":"mission-1","status":"active"}`,
			Digest:        "sha256:d0e1611079d26711d73586278cd2ae20dccd1c7cc51a949882b3ec4289a6f864",
			Consumers:     []string{"go"},
		}},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	})
	if err == nil || !strings.Contains(err.Error(), "consumers must cover every language") {
		t.Fatalf("expected missing Rust consumer rejection, got fixture=%#v err=%v", fixture, err)
	}
}
