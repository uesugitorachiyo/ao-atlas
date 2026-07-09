package atlas

import (
	"bytes"
	"encoding/json"
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
		"registry_schema_count=9",
		"missing_schemas=8",
		"missing_validators=8",
		"failure_reasons=missing_registry_schemas,missing_registry_validators",
		"rsi_remains_denied=true",
		"run_ledger_count=3",
		"all_outputs_have_run_ledgers=true",
		"operator_summary=failed: validation report passed; 8 registry schemas missing; 8 registry validators missing; 3 run ledgers written; RSI remains denied",
		"exact_next_action=Add missing recommendation control-plane evidence artifacts, rerun schema-registry-health, and keep promotion denied.",
		"recommendation_evidence_schema_registry=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-schema-registry.json")),
		"recommendation_evidence_validation_report=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-validation-report.json")),
		"recommendation_evidence_schema_registry_coverage=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-evidence-schema-registry-coverage.json")),
		"schema_registry_run_ledger=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-schema-registry-run-ledger.json")),
		"validation_report_run_ledger=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-validation-report-run-ledger.json")),
		"schema_registry_coverage_run_ledger=" + filepath.ToSlash(filepath.Join(outDir, "recommendation-schema-registry-coverage-run-ledger.json")),
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema-registry-health output missing %q: %s", want, out.String())
		}
	}

	registryPath := filepath.Join(outDir, "recommendation-evidence-schema-registry.json")
	reportPath := filepath.Join(outDir, "recommendation-evidence-validation-report.json")
	coveragePath := filepath.Join(outDir, "recommendation-evidence-schema-registry-coverage.json")
	registryLedgerPath := filepath.Join(outDir, "recommendation-schema-registry-run-ledger.json")
	reportLedgerPath := filepath.Join(outDir, "recommendation-validation-report-run-ledger.json")
	coverageLedgerPath := filepath.Join(outDir, "recommendation-schema-registry-coverage-run-ledger.json")
	for path, schema := range map[string]string{
		registryPath:       AtlasRecommendationEvidenceSchemaRegistryContract,
		reportPath:         AtlasRecommendationEvidenceValidationReportContract,
		coveragePath:       AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
		registryLedgerPath: AtlasRecommendationCommandRunLedgerContract,
		reportLedgerPath:   AtlasRecommendationCommandRunLedgerContract,
		coverageLedgerPath: AtlasRecommendationCommandRunLedgerContract,
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
		len(coverage.MissingSchemas) != 8 ||
		len(coverage.MissingValidators) != 8 ||
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
	for path, want := range map[string]struct {
		command      string
		outputStatus string
	}{
		registryLedgerPath: {"schema-registry", "ready"},
		reportLedgerPath:   {"validate-evidence", "passed"},
		coverageLedgerPath: {"schema-registry-coverage", "failed"},
	} {
		ledger := mustLoadJSON[AtlasRecommendationCommandRunLedger](t, path)
		if ledger.Command != want.command ||
			ledger.OutputStatus != want.outputStatus ||
			!ledger.NoPromotionRequested ||
			ledger.PromotionGranted ||
			ledger.ClaimsAuthorityAdvance ||
			!ledger.RSIRemainsDenied ||
			ledger.SafeToExecute ||
			ledger.SchedulesWork ||
			ledger.ExecutesWork ||
			ledger.ApprovesWork ||
			ledger.MutatesRepositories {
			t.Fatalf("schema registry health ledger %s did not preserve safe command traceability: %#v", path, ledger)
		}
	}
}

func mustDecodeJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, string(data))
	}
	return value
}

func TestMissionRecommendationsSchemaRegistryHealthJSONReportsLedgerCompleteness(t *testing.T) {
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
	var stderr bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-registry-health",
		"--evidence-root", filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"),
		"--out-dir", outDir,
		"--json",
	}, &out, &stderr)
	if code == 0 {
		t.Fatalf("schema-registry-health JSON should fail when registry artifacts are not covered: %s", out.String())
	}
	report := mustDecodeJSON[map[string]any](t, out.Bytes())
	if report["status"] != "failed" ||
		report["validation_report_status"] != "passed" ||
		report["run_ledger_count"] != float64(3) ||
		report["all_outputs_have_run_ledgers"] != true ||
		report["operator_summary"] != "failed: validation report passed; 8 registry schemas missing; 8 registry validators missing; 3 run ledgers written; RSI remains denied" ||
		report["exact_next_action"] != "Add missing recommendation control-plane evidence artifacts, rerun schema-registry-health, and keep promotion denied." ||
		report["rsi_remains_denied"] != true ||
		report["schema_registry_run_ledger"] != filepath.ToSlash(filepath.Join(outDir, "recommendation-schema-registry-run-ledger.json")) ||
		report["validation_report_run_ledger"] != filepath.ToSlash(filepath.Join(outDir, "recommendation-validation-report-run-ledger.json")) ||
		report["schema_registry_coverage_run_ledger"] != filepath.ToSlash(filepath.Join(outDir, "recommendation-schema-registry-coverage-run-ledger.json")) {
		t.Fatalf("schema registry health JSON did not report ledger completeness: %#v", report)
	}
}

func TestMissionRecommendationsSchemaHealthRepairPromptGeneratesPlanningOnlyRepair(t *testing.T) {
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

	healthDir := t.TempDir()
	var healthOut bytes.Buffer
	healthCode := Run([]string{
		"mission", "recommendations", "schema-registry-health",
		"--evidence-root", filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"),
		"--out-dir", healthDir,
	}, &healthOut, &healthOut)
	if healthCode == 0 {
		t.Fatalf("schema-registry-health should fail for repair prompt setup: %s", healthOut.String())
	}

	outDir := t.TempDir()
	fixturePath := filepath.Join(outDir, "schema-health-repair-prompt.json")
	promptPath := filepath.Join(outDir, "schema-health-repair-prompt.md")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-health-repair-prompt",
		"--coverage", filepath.Join(healthDir, "recommendation-evidence-schema-registry-coverage.json"),
		"--node-id", "mission-recommendation-schema-health-repair",
		"--prompt-out", promptPath,
		"--fixture-out", fixturePath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("schema-health-repair-prompt command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=repair_prompt_generated",
		"missing_schemas=8",
		"missing_validators=8",
		"safe_to_execute=false",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema health repair output missing %q: %s", want, out.String())
		}
	}

	fixture := mustLoadJSON[AtlasSchemaHealthRepairPrompt](t, fixturePath)
	if fixture.Schema != "ao.atlas.schema-health-repair-prompt.v0.1" ||
		fixture.Status != "repair_prompt_generated" ||
		fixture.MissingSchemaCount != 8 ||
		fixture.MissingValidatorCount != 8 ||
		!fixture.PlanningOnly ||
		fixture.SafeToExecute ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("schema health repair prompt lost planning-only safety state: %#v", fixture)
	}
	if !containsString(fixture.RepairActions, "Add missing recommendation control-plane evidence artifacts or typed validators for uncovered registry entries.") ||
		!containsString(fixture.RepairActions, "Rerun mission recommendations schema-registry-health and keep promotion denied.") {
		t.Fatalf("schema health repair prompt missing expected repair actions: %#v", fixture.RepairActions)
	}
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, repairing recommendation schema-health coverage.",
		"Missing schemas: `8`",
		"Missing validators: `8`",
		"Rerun `mission recommendations schema-registry-health` after repair.",
		"No promotion is requested.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("schema health repair prompt missing %q:\n%s", want, prompt)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(fixturePath, "ao.atlas.schema-health-repair-prompt.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:schema-health-repair-prompt" {
		t.Fatalf("expected typed schema health repair prompt validator, got %s", validator)
	}
}
