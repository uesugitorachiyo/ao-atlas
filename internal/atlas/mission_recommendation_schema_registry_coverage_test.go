package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsSchemaRegistryCoverageAuditsValidationReport(t *testing.T) {
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

	tempDir := t.TempDir()
	registryPath := filepath.Join(tempDir, "recommendation-evidence-schema-registry.json")
	registryOut := bytes.Buffer{}
	code := Run([]string{
		"mission", "recommendations", "schema-registry",
		"--out", registryPath,
	}, &registryOut, &registryOut)
	if code != 0 {
		t.Fatalf("schema-registry failed: %s", registryOut.String())
	}

	reportPath := filepath.Join(tempDir, "recommendation-evidence-validation-report.json")
	report := AtlasRecommendationEvidenceValidationReport{
		Schema:                   AtlasRecommendationEvidenceValidationReportContract,
		Status:                   "passed",
		EvidenceRoot:             "docs/evidence/example",
		NodeRoot:                 "docs/evidence/example/nodes",
		NodeCount:                1,
		JSONFileCount:            6,
		ValidatedJSONFiles:       6,
		SchemaBoundFiles:         6,
		TypedValidatorFiles:      6,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationFinalResponseGatesContract:             1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
		},
		Validators: map[string]int{
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-final-response-gates":              1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
		},
		Entries: []AtlasRecommendationEvidenceValidationEntry{},
	}
	if err := WriteJSON(reportPath, report); err != nil {
		t.Fatal(err)
	}

	coveragePath := filepath.Join(tempDir, "recommendation-schema-registry-coverage.json")
	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "schema-registry-coverage",
		"--registry", registryPath,
		"--validation-report", reportPath,
		"--out", coveragePath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("schema-registry-coverage failed: %s", out.String())
	}
	for _, want := range []string{
		"status=passed",
		"registry_schema_count=6",
		"missing_schemas=0",
		"missing_validators=0",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema-registry-coverage output missing %q: %s", want, out.String())
		}
	}

	coverage := mustLoadJSON[map[string]any](t, coveragePath)
	if coverage["schema"] != "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1" ||
		coverage["status"] != "passed" ||
		coverage["registry_schema_count"] != float64(6) ||
		coverage["covered_schema_count"] != float64(6) ||
		coverage["registry_validator_count"] != float64(6) ||
		coverage["covered_validator_count"] != float64(6) ||
		coverage["all_registry_schemas_covered"] != true ||
		coverage["all_registry_validators_covered"] != true ||
		coverage["no_promotion_requested"] != true ||
		coverage["promotion_granted"] != false ||
		coverage["claims_authority_advance"] != false ||
		coverage["rsi_remains_denied"] != true ||
		coverage["safe_to_execute"] != false ||
		coverage["schedules_work"] != false ||
		coverage["executes_work"] != false ||
		coverage["approves_work"] != false ||
		coverage["mutates_repositories"] != false {
		t.Fatalf("schema registry coverage did not publish safe coverage state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}

func TestMissionRecommendationsSchemaRegistryCoverageRecordsMissingSchemaFailureReason(t *testing.T) {
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

	tempDir := t.TempDir()
	registryPath := filepath.Join(tempDir, "recommendation-evidence-schema-registry.json")
	registryOut := bytes.Buffer{}
	code := Run([]string{
		"mission", "recommendations", "schema-registry",
		"--out", registryPath,
	}, &registryOut, &registryOut)
	if code != 0 {
		t.Fatalf("schema-registry failed: %s", registryOut.String())
	}

	reportPath := filepath.Join(tempDir, "recommendation-evidence-validation-report.json")
	report := AtlasRecommendationEvidenceValidationReport{
		Schema:                   AtlasRecommendationEvidenceValidationReportContract,
		Status:                   "passed",
		EvidenceRoot:             "docs/evidence/example",
		NodeRoot:                 "docs/evidence/example/nodes",
		NodeCount:                1,
		JSONFileCount:            5,
		ValidatedJSONFiles:       5,
		SchemaBoundFiles:         5,
		TypedValidatorFiles:      6,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
		},
		Validators: map[string]int{
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-final-response-gates":              1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
		},
		Entries: []AtlasRecommendationEvidenceValidationEntry{},
	}
	if err := WriteJSON(reportPath, report); err != nil {
		t.Fatal(err)
	}

	coveragePath := filepath.Join(tempDir, "recommendation-schema-registry-coverage.json")
	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "schema-registry-coverage",
		"--registry", registryPath,
		"--validation-report", reportPath,
		"--out", coveragePath,
	}, &out, &out)
	if code == 0 {
		t.Fatalf("schema-registry-coverage should fail when a registry schema is missing: %s", out.String())
	}
	for _, want := range []string{
		"status=failed",
		"missing_schemas=1",
		"missing_validators=0",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema-registry-coverage failure output missing %q: %s", want, out.String())
		}
	}

	coverage := mustLoadJSON[map[string]any](t, coveragePath)
	missingSchemas, ok := coverage["missing_schemas"].([]any)
	if !ok || len(missingSchemas) != 1 || missingSchemas[0] != AtlasRecommendationFinalResponseGatesContract {
		t.Fatalf("missing schemas should identify final-response gates: %#v", coverage["missing_schemas"])
	}
	failureReasons, ok := coverage["failure_reasons"].([]any)
	if !ok || len(failureReasons) != 1 || failureReasons[0] != "missing_registry_schemas" {
		t.Fatalf("failure reasons should identify missing registry schemas: %#v", coverage["failure_reasons"])
	}
	if coverage["status"] != "failed" ||
		coverage["all_registry_schemas_covered"] != false ||
		coverage["all_registry_validators_covered"] != true ||
		coverage["no_promotion_requested"] != true ||
		coverage["promotion_granted"] != false ||
		coverage["claims_authority_advance"] != false ||
		coverage["rsi_remains_denied"] != true ||
		coverage["safe_to_execute"] != false ||
		coverage["schedules_work"] != false ||
		coverage["executes_work"] != false ||
		coverage["approves_work"] != false ||
		coverage["mutates_repositories"] != false {
		t.Fatalf("failed schema coverage should preserve safe no-promotion state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}
