package atlas

import (
	"os"
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

func TestFeatureDepthWaveV02ClosureRollupArtifactsUseTypedEvidenceValidation(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	reportPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-05", "feature-depth-evidence-validation-report.json")
	recorded := mustLoadJSON[AtlasRecommendationEvidenceValidationReport](t, reportPath)
	checkpointRoot := evidenceValidationReportCheckpointRoot(t, waveRoot, recorded)

	report, err := BuildAtlasRecommendationEvidenceValidationReport(checkpointRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationEvidenceValidationReport(recorded); err != nil {
		t.Fatal(err)
	}
	report.EvidenceRoot = recorded.EvidenceRoot
	report.NodeRoot = recorded.NodeRoot
	if digestValue(report) != digestValue(recorded) {
		t.Fatalf("v02 evidence validation report drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(report))
	}
	if recorded.Status != "passed" ||
		recorded.NodeCount != 5 ||
		!recorded.RequiredFilenamesCovered ||
		len(recorded.MissingRequiredFiles) != 0 ||
		len(recorded.MissingSchemaFiles) != 0 ||
		len(recorded.FailedFiles) != 0 {
		t.Fatalf("v02 evidence validation report must pass across the first five completed nodes: %#v", recorded)
	}
	for schema, validator := range map[string]string{
		"ao.atlas.command-readback.v0.1":       "typed:command-readback",
		"ao.atlas.promoter-no-promotion.v0.1":  "typed:promoter-no-promotion",
		"ao.atlas.sentinel-public-safety.v0.1": "typed:sentinel-public-safety",
	} {
		schemaCount := recorded.SchemaCounts[schema]
		if schemaCount == 0 {
			t.Fatalf("expected schema count for %s in v02 evidence report", schema)
		}
		if recorded.Validators[validator] != schemaCount {
			t.Fatalf("%s should validate every %s file, got validator count %d schema count %d", validator, schema, recorded.Validators[validator], schemaCount)
		}
	}
}

func evidenceValidationReportCheckpointRoot(t *testing.T, waveRoot string, recorded AtlasRecommendationEvidenceValidationReport) string {
	t.Helper()
	checkpointRoot := filepath.Join(t.TempDir(), "evidence")
	for _, entry := range recorded.Entries {
		src := filepath.Join(waveRoot, filepath.FromSlash(entry.Path))
		dst := filepath.Join(checkpointRoot, filepath.FromSlash(entry.Path))
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return checkpointRoot
}
