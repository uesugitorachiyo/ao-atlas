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
		JSONFileCount:            8,
		ValidatedJSONFiles:       8,
		SchemaBoundFiles:         8,
		TypedValidatorFiles:      8,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationFinalResponseGatesContract:             1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-final-response-gates":              1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
		"registry_schema_count=9",
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
		coverage["registry_schema_count"] != float64(9) ||
		coverage["covered_schema_count"] != float64(9) ||
		coverage["registry_validator_count"] != float64(9) ||
		coverage["covered_validator_count"] != float64(9) ||
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
		t.Fatalf("schema registry coverage did not emit safe coverage state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}

func assertSchemaRegistryCoverageFailureOutput(t *testing.T, output string, expected []string) {
	t.Helper()
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Fatalf("schema-registry-coverage failure output missing %q: %s", want, output)
		}
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
		JSONFileCount:            7,
		ValidatedJSONFiles:       7,
		SchemaBoundFiles:         7,
		TypedValidatorFiles:      8,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-final-response-gates":              1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
	assertSchemaRegistryCoverageFailureOutput(t, out.String(), []string{
		"status=failed",
		"missing_schemas=1",
		"missing_validators=0",
		"rsi_remains_denied=true",
	})

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

func TestMissionRecommendationsSchemaRegistryCoverageDetectsStaleRegistryEntries(t *testing.T) {
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
		JSONFileCount:            7,
		ValidatedJSONFiles:       7,
		SchemaBoundFiles:         7,
		TypedValidatorFiles:      7,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
		t.Fatalf("schema-registry-coverage should fail when a registry entry is stale: %s", out.String())
	}
	assertSchemaRegistryCoverageFailureOutput(t, out.String(), []string{
		"status=failed",
		"missing_schemas=1",
		"missing_validators=1",
		"stale_registry_entries=1",
		"failure_reasons=missing_registry_schemas,missing_registry_validators,stale_registry_entries",
		"rsi_remains_denied=true",
	})

	coverage := mustLoadJSON[AtlasRecommendationEvidenceSchemaRegistryCoverage](t, coveragePath)
	if coverage.StaleRegistryEntryCount != 1 ||
		len(coverage.StaleRegistryEntries) != 1 ||
		coverage.StaleRegistryEntries[0] != AtlasRecommendationFinalResponseGatesContract ||
		coverage.AllRegistryEntriesFresh ||
		!containsString(coverage.FailureReasons, "stale_registry_entries") ||
		coverage.NoPromotionRequested != true ||
		coverage.PromotionGranted ||
		coverage.ClaimsAuthorityAdvance ||
		!coverage.RSIRemainsDenied {
		t.Fatalf("schema registry coverage did not report stale entry safely: %#v", coverage)
	}
}

func TestMissionRecommendationsSchemaRegistryCoverageRecordsMissingValidatorFailureReason(t *testing.T) {
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
		JSONFileCount:            8,
		ValidatedJSONFiles:       8,
		SchemaBoundFiles:         8,
		TypedValidatorFiles:      7,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationFinalResponseGatesContract:             1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
		t.Fatalf("schema-registry-coverage should fail when a registry validator is missing: %s", out.String())
	}
	assertSchemaRegistryCoverageFailureOutput(t, out.String(), []string{
		"status=failed",
		"missing_schemas=0",
		"missing_validators=1",
		"failure_reasons=missing_registry_validators",
		"rsi_remains_denied=true",
	})

	coverage := mustLoadJSON[map[string]any](t, coveragePath)
	missingValidators, ok := coverage["missing_validators"].([]any)
	if !ok || len(missingValidators) != 1 || missingValidators[0] != "typed:recommendation-final-response-gates" {
		t.Fatalf("missing validators should identify final-response gates: %#v", coverage["missing_validators"])
	}
	failureReasons, ok := coverage["failure_reasons"].([]any)
	if !ok || len(failureReasons) != 1 || failureReasons[0] != "missing_registry_validators" {
		t.Fatalf("failure reasons should identify missing registry validators: %#v", coverage["failure_reasons"])
	}
	if coverage["status"] != "failed" ||
		coverage["all_registry_schemas_covered"] != true ||
		coverage["all_registry_validators_covered"] != false ||
		coverage["no_promotion_requested"] != true ||
		coverage["promotion_granted"] != false ||
		coverage["claims_authority_advance"] != false ||
		coverage["rsi_remains_denied"] != true ||
		coverage["safe_to_execute"] != false ||
		coverage["schedules_work"] != false ||
		coverage["executes_work"] != false ||
		coverage["approves_work"] != false ||
		coverage["mutates_repositories"] != false {
		t.Fatalf("failed validator coverage should preserve safe no-promotion state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}

func TestMissionRecommendationsSchemaRegistryCoverageRecordsCombinedMissingSchemaAndValidatorFailureReasons(t *testing.T) {
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
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
		t.Fatalf("schema-registry-coverage should fail when registry schemas and validators are missing: %s", out.String())
	}
	assertSchemaRegistryCoverageFailureOutput(t, out.String(), []string{
		"status=failed",
		"missing_schemas=2",
		"missing_validators=2",
		"stale_registry_entries=2",
		"failure_reasons=missing_registry_schemas,missing_registry_validators,stale_registry_entries",
		"rsi_remains_denied=true",
	})

	coverage := mustLoadJSON[map[string]any](t, coveragePath)
	missingSchemas, ok := coverage["missing_schemas"].([]any)
	if !ok || len(missingSchemas) != 2 ||
		missingSchemas[0] != AtlasRecommendationCommandRunLedgerContract ||
		missingSchemas[1] != AtlasRecommendationFinalResponseGatesContract {
		t.Fatalf("missing schemas should identify command ledger and final-response gates: %#v", coverage["missing_schemas"])
	}
	missingValidators, ok := coverage["missing_validators"].([]any)
	if !ok || len(missingValidators) != 2 ||
		missingValidators[0] != "typed:recommendation-command-run-ledger" ||
		missingValidators[1] != "typed:recommendation-final-response-gates" {
		t.Fatalf("missing validators should identify command ledger and final-response gates: %#v", coverage["missing_validators"])
	}
	failureReasons, ok := coverage["failure_reasons"].([]any)
	if !ok || len(failureReasons) != 3 ||
		failureReasons[0] != "missing_registry_schemas" ||
		failureReasons[1] != "missing_registry_validators" ||
		failureReasons[2] != "stale_registry_entries" {
		t.Fatalf("failure reasons should preserve combined schema and validator causes: %#v", coverage["failure_reasons"])
	}
	staleEntries, ok := coverage["stale_registry_entries"].([]any)
	if !ok || len(staleEntries) != 2 ||
		staleEntries[0] != AtlasRecommendationCommandRunLedgerContract ||
		staleEntries[1] != AtlasRecommendationFinalResponseGatesContract ||
		coverage["stale_registry_entry_count"] != float64(2) ||
		coverage["all_registry_entries_fresh"] != false {
		t.Fatalf("stale registry entries should identify fully missing registry entries: %#v", coverage)
	}
	if coverage["status"] != "failed" ||
		coverage["all_registry_schemas_covered"] != false ||
		coverage["all_registry_validators_covered"] != false ||
		coverage["no_promotion_requested"] != true ||
		coverage["promotion_granted"] != false ||
		coverage["claims_authority_advance"] != false ||
		coverage["rsi_remains_denied"] != true ||
		coverage["safe_to_execute"] != false ||
		coverage["schedules_work"] != false ||
		coverage["executes_work"] != false ||
		coverage["approves_work"] != false ||
		coverage["mutates_repositories"] != false {
		t.Fatalf("combined failed coverage should preserve safe no-promotion state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}

func TestMissionRecommendationsSchemaRegistryCoverageRecordsFailedValidationReportFailureReason(t *testing.T) {
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
		Status:                   "failed",
		EvidenceRoot:             "docs/evidence/example",
		NodeRoot:                 "docs/evidence/example/nodes",
		NodeCount:                1,
		JSONFileCount:            8,
		ValidatedJSONFiles:       6,
		SchemaBoundFiles:         8,
		TypedValidatorFiles:      8,
		GenericSchemaFiles:       0,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{"docs/evidence/example/nodes/node-01/command_readback.json"},
		MissingRequiredFiles:     []string{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		SchemaCounts: map[string]int{
			AOMissionRefactoringRecommendationsContract:               1,
			AtlasRecommendationNextTrackDecisionContract:              1,
			AtlasConsumedRecommendationLedgerContract:                 1,
			AtlasRecommendationTrackRegistryContract:                  1,
			AtlasRecommendationCommandRunLedgerContract:               1,
			AtlasRecommendationEvidenceValidationReportContract:       1,
			AtlasRecommendationFinalResponseGatesContract:             1,
			AtlasRecommendationEvidenceSchemaRegistryCoverageContract: 1,
			AtlasSchemaHealthRepairPromptContract:                     1,
		},
		Validators: map[string]int{
			"typed:recommendation-refactoring-recommendations":       1,
			"typed:recommendation-next-track-decision":               1,
			"typed:consumed-recommendation-ledger":                   1,
			"typed:recommendation-track-registry":                    1,
			"typed:recommendation-command-run-ledger":                1,
			"typed:recommendation-evidence-validation-report":        1,
			"typed:recommendation-final-response-gates":              1,
			"typed:recommendation-evidence-schema-registry-coverage": 1,
			"typed:schema-health-repair-prompt":                      1,
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
		t.Fatalf("schema-registry-coverage should fail when validation report failed: %s", out.String())
	}
	assertSchemaRegistryCoverageFailureOutput(t, out.String(), []string{
		"status=failed",
		"validation_report_status=failed",
		"missing_schemas=0",
		"missing_validators=0",
		"failure_reasons=validation_report_failed",
		"rsi_remains_denied=true",
	})

	coverage := mustLoadJSON[map[string]any](t, coveragePath)
	failureReasons, ok := coverage["failure_reasons"].([]any)
	if !ok || len(failureReasons) != 1 || failureReasons[0] != "validation_report_failed" {
		t.Fatalf("failure reasons should identify failed validation report: %#v", coverage["failure_reasons"])
	}
	if coverage["status"] != "failed" ||
		coverage["validation_report_status"] != "failed" ||
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
		t.Fatalf("failed report coverage should preserve safe no-promotion state: %#v", coverage)
	}
	validator, err := validateRecommendationEvidenceTypedFile(coveragePath, "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry-coverage" {
		t.Fatalf("expected typed recommendation schema registry coverage validator, got %s", validator)
	}
}

func TestValidateAtlasRecommendationEvidenceSchemaRegistryCoverageRejectsInvalidFailureReasonWithAllowedValues(t *testing.T) {
	coverage := AtlasRecommendationEvidenceSchemaRegistryCoverage{
		Schema:                       AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
		Status:                       "failed",
		RegistryPath:                 "docs/evidence/example/recommendation-evidence-schema-registry.json",
		ValidationReportPath:         "docs/evidence/example/recommendation-evidence-validation-report.json",
		ValidationReportStatus:       "failed",
		RegistrySchemaCount:          6,
		CoveredSchemaCount:           6,
		MissingSchemas:               []string{},
		RegistryValidatorCount:       6,
		CoveredValidatorCount:        6,
		MissingValidators:            []string{},
		FailureReasons:               []string{"unexpected_failure_reason"},
		AllRegistrySchemasCovered:    true,
		AllRegistryValidatorsCovered: true,
		NoPromotionRequested:         true,
		PromotionGranted:             false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
		SafeToExecute:                false,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		MutatesRepositories:          false,
	}

	err := ValidateAtlasRecommendationEvidenceSchemaRegistryCoverage(coverage)
	if err == nil {
		t.Fatal("expected invalid failure reason to be rejected")
	}
	for _, want := range []string{
		"failure_reasons contains invalid reason unexpected_failure_reason",
		"allowed: validation_report_failed, missing_registry_schemas, missing_registry_validators",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("invalid failure reason error missing %q: %v", want, err)
		}
	}
}
