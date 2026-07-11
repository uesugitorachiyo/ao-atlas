package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3GoldenPathReadinessMatrix(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-40")
	recordedPath := filepath.Join(nodeDir, "golden-path-readiness-matrix.json")
	outPath := filepath.Join(t.TempDir(), "golden-path-readiness-matrix.json")

	var out bytes.Buffer
	code := Run([]string{"mission", "recommendations", "golden-path-readiness-matrix", "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("golden-path-readiness-matrix command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=golden_path_readiness_matrix_ready",
		"completed_nodes=40",
		"ranked_recommendations=40",
		"promotion_requested=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("readiness matrix output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("golden path readiness matrix changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["completed_nodes"].(float64) != 40 ||
		generated["ranked_recommendation_count"].(float64) != 40 ||
		generated["promotion_requested"] != false ||
		generated["no_promotion_status"] != "no_promotion_requested" ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("readiness matrix lost terminal safety state: %#v", generated)
	}
}

func TestMonth3GoldenPathReadinessMatrixUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-40", "golden-path-readiness-matrix.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.golden-path-readiness-matrix.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:golden-path-readiness-matrix" {
		t.Fatalf("expected typed golden path readiness matrix validator, got %s", validator)
	}
}

func TestMonth3GoldenPathReadinessMatrixRequiresFortyRecommendations(t *testing.T) {
	matrix, err := BuildAtlasGoldenPathReadinessMatrix()
	if err != nil {
		t.Fatal(err)
	}
	matrix.RankedRecommendations = matrix.RankedRecommendations[:39]
	matrix.RankedRecommendationCount = len(matrix.RankedRecommendations)
	if err := ValidateAtlasGoldenPathReadinessMatrix(matrix); err == nil || !strings.Contains(err.Error(), "ranked_recommendation_count must be at least 40") {
		t.Fatalf("expected recommendation count rejection, got %v", err)
	}
}

func TestMonth3GoldenPathReadinessMatrixUsesDomainSpecificRecommendations(t *testing.T) {
	matrix, err := BuildAtlasGoldenPathReadinessMatrix()
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range matrix.RankedRecommendations {
		if strings.Contains(rec.Task, "golden-path-followup-") {
			t.Fatalf("readiness matrix must not use placeholder recommendation labels: %#v", rec)
		}
		if len(strings.Fields(rec.Task)) < 6 {
			t.Fatalf("readiness matrix recommendation must be actionable: %#v", rec)
		}
	}
	for _, want := range []string{
		"aggregate Promoter Command public-safety closure rollup",
		"non-AO repository dry-run replay binding",
		"provider and model provenance",
	} {
		found := false
		for _, rec := range matrix.RankedRecommendations {
			if strings.Contains(rec.Task, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("readiness matrix missing domain recommendation containing %q: %#v", want, matrix.RankedRecommendations)
		}
	}
}
