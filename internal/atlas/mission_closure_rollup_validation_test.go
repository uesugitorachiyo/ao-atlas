package atlas

import (
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveClosureRollupArtifactsUseTypedEvidenceValidation(t *testing.T) {
	waveRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-feature-depth-wave-v01")

	report, err := BuildAtlasRecommendationEvidenceValidationReport(waveRoot)
	if err != nil {
		t.Fatal(err)
	}
	for schema, validator := range map[string]string{
		"ao.atlas.command-readback.v0.1":       "typed:command-readback",
		"ao.atlas.promoter-no-promotion.v0.1":  "typed:promoter-no-promotion",
		"ao.atlas.sentinel-public-safety.v0.1": "typed:sentinel-public-safety",
	} {
		schemaCount := report.SchemaCounts[schema]
		if schemaCount == 0 {
			t.Fatalf("expected schema count for %s in Feature Depth evidence report", schema)
		}
		if report.Validators[validator] != schemaCount {
			t.Fatalf("%s should validate every %s file, got validator count %d schema count %d", validator, schema, report.Validators[validator], schemaCount)
		}
	}
}
