package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsSchemaRegistryHealthChainsValidationAndCoverage(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	outDir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-registry-health",
		"--evidence-root", filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"),
		"--out-dir", outDir,
	}, &out, &out)
	if code == 0 {
		t.Fatalf("schema-registry-health should fail when registry artifacts are not covered: %s", out.String())
	}
	for _, want := range []string{
		"status=failed",
		"validation_report_status=passed",
		"registry_schema_count=7",
		"missing_schemas=6",
		"missing_validators=6",
		"failure_reasons=missing_registry_schemas,missing_registry_validators",
		"rsi_remains_denied=true",
		"recommendation_evidence_schema_registry=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-schema-registry.json")),
		"recommendation_evidence_validation_report=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-validation-report.json")),
		"recommendation_evidence_schema_registry_coverage=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-schema-registry-coverage.json")),
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema-registry-health output missing %q: %s", want, out.String())
		}
	}

	registryPath := filepath.Join(outDir, "recommendation-evidence-schema-registry.json")
	reportPath := filepath.Join(outDir, "recommendation-evidence-validation-report.json")
	coveragePath := filepath.Join(outDir, "recommendation-evidence-schema-registry-coverage.json")
	for path, schema := range map[string]string{
		registryPath: AtlasRecommendationEvidenceSchemaRegistryContract,
		reportPath:   AtlasRecommendationEvidenceValidationReportContract,
		coveragePath: AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
	} {
		validator, err := validateRecommendationEvidenceTypedFile(path, schema)
		if err != nil {
			t.Fatalf("validate generated artifact %s: %v", path, err)
		}
		if !strings.HasPrefix(validator, "typed:") {
			t.Fatalf("generated artifact %s should use typed validator, got %s", path, validator)
		}
	}
	coverage := mustLoadJSON[AtlasRecommendationEvidenceSchemaRegistryCoverage](t, coveragePath)
	if coverage.Status != "failed" ||
		coverage.ValidationReportStatus != "passed" ||
		len(coverage.MissingSchemas) != 6 ||
		len(coverage.MissingValidators) != 6 ||
		!coverage.NoPromotionRequested ||
		coverage.PromotionGranted ||
		coverage.ClaimsAuthorityAdvance ||
		!coverage.RSIRemainsDenied ||
		coverage.SafeToExecute ||
		coverage.SchedulesWork ||
		coverage.ExecutesWork ||
		coverage.ApprovesWork ||
		coverage.MutatesRepositories {
		t.Fatalf("schema registry health coverage did not preserve safe failure metadata: %#v", coverage)
	}
}
