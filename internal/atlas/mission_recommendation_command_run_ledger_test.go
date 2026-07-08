package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsCommandRunLedgerRecordsTypedArtifactOutput(t *testing.T) {
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
	registryPath := filepath.Join(tempDir, "recommendation-track-registry.json")
	ledgerPath := filepath.Join(tempDir, "recommendation-command-run-ledger.json")

	var registryOut bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "track-registry",
		"--out", registryPath,
	}, &registryOut, &registryOut)
	if code != 0 {
		t.Fatalf("track-registry failed: %s", registryOut.String())
	}

	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "run-ledger",
		"--command", "track-registry",
		"--artifact", registryPath,
		"--out", ledgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger failed: %s", out.String())
	}
	for _, want := range []string{
		"status=recorded",
		"command=track-registry",
		"artifact_schema=ao.atlas.recommendation-track-registry.v0.1",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger output missing %q: %s", want, out.String())
		}
	}

	ledger := mustLoadJSON[map[string]any](t, ledgerPath)
	if ledger["schema"] != "ao.atlas.recommendation-command-run-ledger.v0.1" ||
		ledger["status"] != "recorded" ||
		ledger["command"] != "track-registry" ||
		ledger["artifact_path"] != filepath.ToSlash(registryPath) ||
		ledger["artifact_schema"] != "ao.atlas.recommendation-track-registry.v0.1" ||
		ledger["typed_validator"] != "typed:recommendation-track-registry" ||
		ledger["output_status"] != "ready" ||
		ledger["records_invocation"] != true ||
		ledger["safe_to_execute"] != false ||
		ledger["schedules_work"] != false ||
		ledger["executes_work"] != false ||
		ledger["approves_work"] != false ||
		ledger["mutates_repositories"] != false ||
		ledger["no_promotion_requested"] != true ||
		ledger["promotion_granted"] != false ||
		ledger["claims_authority_advance"] != false ||
		ledger["rsi_remains_denied"] != true {
		t.Fatalf("run ledger did not record the typed command output safely: %#v", ledger)
	}
	if digest, ok := ledger["artifact_digest"].(string); !ok || !digestPattern.MatchString(digest) {
		t.Fatalf("run ledger missing artifact digest: %#v", ledger["artifact_digest"])
	}
	validator, err := validateRecommendationEvidenceTypedFile(ledgerPath, "ao.atlas.recommendation-command-run-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-command-run-ledger" {
		t.Fatalf("expected typed recommendation command run ledger validator, got %s", validator)
	}
}

func TestMissionRecommendationsCommandRunLedgerRecordsSchemaRegistryOutput(t *testing.T) {
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
	ledgerPath := filepath.Join(tempDir, "recommendation-schema-registry-run-ledger.json")
	var registryOut bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-registry",
		"--out", registryPath,
	}, &registryOut, &registryOut)
	if code != 0 {
		t.Fatalf("schema-registry failed: %s", registryOut.String())
	}

	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "run-ledger",
		"--command", "schema-registry",
		"--artifact", registryPath,
		"--out", ledgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger failed: %s", out.String())
	}
	for _, want := range []string{
		"status=recorded",
		"command=schema-registry",
		"artifact_schema=ao.atlas.recommendation-evidence-schema-registry.v0.1",
		"typed_validator=typed:recommendation-evidence-schema-registry",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger output missing %q: %s", want, out.String())
		}
	}

	ledger := mustLoadJSON[map[string]any](t, ledgerPath)
	if ledger["schema"] != "ao.atlas.recommendation-command-run-ledger.v0.1" ||
		ledger["status"] != "recorded" ||
		ledger["command"] != "schema-registry" ||
		ledger["artifact_schema"] != "ao.atlas.recommendation-evidence-schema-registry.v0.1" ||
		ledger["typed_validator"] != "typed:recommendation-evidence-schema-registry" ||
		ledger["no_promotion_requested"] != true ||
		ledger["promotion_granted"] != false ||
		ledger["claims_authority_advance"] != false ||
		ledger["rsi_remains_denied"] != true {
		t.Fatalf("run ledger did not record schema registry safely: %#v", ledger)
	}
	validator, err := validateRecommendationEvidenceTypedFile(ledgerPath, "ao.atlas.recommendation-command-run-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-command-run-ledger" {
		t.Fatalf("expected typed recommendation command run ledger validator, got %s", validator)
	}
}

func TestMissionRecommendationsCommandRunLedgerRecordsFailedSchemaRegistryCoverageOutput(t *testing.T) {
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
	coveragePath := filepath.Join(tempDir, "recommendation-schema-registry-coverage.json")
	ledgerPath := filepath.Join(tempDir, "recommendation-schema-registry-coverage-run-ledger.json")
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
		FailureReasons:               []string{"validation_report_failed"},
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
	if err := WriteJSON(coveragePath, coverage); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-ledger",
		"--command", "schema-registry-coverage",
		"--artifact", coveragePath,
		"--out", ledgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger failed: %s", out.String())
	}
	for _, want := range []string{
		"status=recorded",
		"command=schema-registry-coverage",
		"artifact_schema=ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1",
		"typed_validator=typed:recommendation-evidence-schema-registry-coverage",
		"output_status=failed",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger output missing %q: %s", want, out.String())
		}
	}

	ledger := mustLoadJSON[map[string]any](t, ledgerPath)
	if ledger["schema"] != "ao.atlas.recommendation-command-run-ledger.v0.1" ||
		ledger["status"] != "recorded" ||
		ledger["command"] != "schema-registry-coverage" ||
		ledger["artifact_schema"] != "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1" ||
		ledger["typed_validator"] != "typed:recommendation-evidence-schema-registry-coverage" ||
		ledger["output_status"] != "failed" ||
		ledger["records_invocation"] != true ||
		ledger["no_promotion_requested"] != true ||
		ledger["promotion_granted"] != false ||
		ledger["claims_authority_advance"] != false ||
		ledger["rsi_remains_denied"] != true ||
		ledger["safe_to_execute"] != false ||
		ledger["schedules_work"] != false ||
		ledger["executes_work"] != false ||
		ledger["approves_work"] != false ||
		ledger["mutates_repositories"] != false {
		t.Fatalf("run ledger did not record failed schema-registry coverage safely: %#v", ledger)
	}
	validator, err := validateRecommendationEvidenceTypedFile(ledgerPath, "ao.atlas.recommendation-command-run-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-command-run-ledger" {
		t.Fatalf("expected typed recommendation command run ledger validator, got %s", validator)
	}
}
