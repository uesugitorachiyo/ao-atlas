package atlas

import (
	"path/filepath"
	"testing"
)

func TestStrictRecommendationEvidenceSchemaRegistryRejectsUnknownSchema(t *testing.T) {
	unknownPath := filepath.Join(repoRoot(t), "examples", "invalid", "recommendation-evidence-unknown-schema.json")
	validator, err := validateRecommendationEvidenceTypedFileStrict(unknownPath, "ao.atlas.consolidation-unknown.v0.1")
	if err == nil {
		t.Fatal("strict evidence validation accepted an unknown schema")
	}
	if validator != "strict:unknown-schema" {
		t.Fatalf("unknown schema used validator %q", validator)
	}
	if got := err.Error(); got != `unknown recommendation evidence schema "ao.atlas.consolidation-unknown.v0.1"` {
		t.Fatalf("unexpected unknown schema error: %s", got)
	}
}

func TestStrictRecommendationEvidenceSchemaRegistryAcceptsKnownConsolidationSchema(t *testing.T) {
	path := filepath.Join(repoRoot(t), "docs", "evidence", "ao-stack-consolidation-month1-wave-v01", "nodes", "mission-recommendation-consolidation-month1-11", "foundry_import.json")
	validator, err := validateRecommendationEvidenceTypedFileStrict(path, "ao.atlas.consolidation-foundry-import.v0.1")
	if err != nil {
		t.Fatalf("known consolidation schema was rejected: %v", err)
	}
	if validator != "strict:generic:schema-marker" {
		t.Fatalf("known consolidation schema used validator %q", validator)
	}
}

func TestStrictRecommendationEvidenceSchemaRegistryAcceptsEvidenceVolumeBaseline(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "valid", "evidence-volume-baseline.json")
	validator, err := validateRecommendationEvidenceTypedFileStrict(path, "ao.atlas.consolidation-evidence-volume-baseline.v0.1")
	if err != nil {
		t.Fatalf("evidence volume baseline schema was rejected: %v", err)
	}
	if validator != "strict:generic:schema-marker" {
		t.Fatalf("evidence volume baseline used validator %q", validator)
	}
}

func TestStrictRecommendationEvidenceSchemaRegistryAcceptsEvidenceCatalogPlan(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "valid", "evidence-catalog-plan.json")
	validator, err := validateRecommendationEvidenceTypedFileStrict(path, "ao.atlas.consolidation-evidence-catalog-plan.v0.1")
	if err != nil {
		t.Fatalf("evidence catalog plan schema was rejected: %v", err)
	}
	if validator != "strict:generic:schema-marker" {
		t.Fatalf("evidence catalog plan used validator %q", validator)
	}
}
