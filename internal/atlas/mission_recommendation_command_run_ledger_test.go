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

func TestMissionRecommendationsCommandRunLedgerRecordsValidationReportOutput(t *testing.T) {
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
	reportPath := filepath.Join(tempDir, "recommendation-evidence-validation-report.json")
	ledgerPath := filepath.Join(tempDir, "recommendation-validation-report-run-ledger.json")
	var reportOut bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "validate-evidence",
		"--evidence-root", filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"),
		"--out", reportPath,
	}, &reportOut, &reportOut)
	if code != 0 {
		t.Fatalf("validate-evidence failed: %s", reportOut.String())
	}

	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "run-ledger",
		"--command", "validate-evidence",
		"--artifact", reportPath,
		"--out", ledgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger failed: %s", out.String())
	}
	for _, want := range []string{
		"status=recorded",
		"command=validate-evidence",
		"artifact_schema=ao.atlas.recommendation-evidence-validation-report.v0.1",
		"typed_validator=typed:recommendation-evidence-validation-report",
		"output_status=passed",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger output missing %q: %s", want, out.String())
		}
	}

	ledger := mustLoadJSON[map[string]any](t, ledgerPath)
	if ledger["schema"] != "ao.atlas.recommendation-command-run-ledger.v0.1" ||
		ledger["status"] != "recorded" ||
		ledger["command"] != "validate-evidence" ||
		ledger["artifact_schema"] != "ao.atlas.recommendation-evidence-validation-report.v0.1" ||
		ledger["typed_validator"] != "typed:recommendation-evidence-validation-report" ||
		ledger["output_status"] != "passed" ||
		ledger["no_promotion_requested"] != true ||
		ledger["promotion_granted"] != false ||
		ledger["claims_authority_advance"] != false ||
		ledger["rsi_remains_denied"] != true ||
		ledger["safe_to_execute"] != false ||
		ledger["schedules_work"] != false ||
		ledger["executes_work"] != false ||
		ledger["approves_work"] != false ||
		ledger["mutates_repositories"] != false {
		t.Fatalf("run ledger did not record validation report output safely: %#v", ledger)
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
		StaleRegistryEntryCount:      0,
		StaleRegistryEntries:         []string{},
		AllRegistryEntriesFresh:      true,
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

func TestMissionRecommendationsCommandRunLedgerRollupAggregatesControlPlaneLedgers(t *testing.T) {
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
	reportPath := filepath.Join(tempDir, "recommendation-evidence-validation-report.json")
	coveragePath := filepath.Join(tempDir, "recommendation-evidence-schema-registry-coverage.json")
	registryLedgerPath := filepath.Join(tempDir, "recommendation-schema-registry-run-ledger.json")
	reportLedgerPath := filepath.Join(tempDir, "recommendation-validation-report-run-ledger.json")
	coverageLedgerPath := filepath.Join(tempDir, "recommendation-schema-registry-coverage-run-ledger.json")
	rollupPath := filepath.Join(tempDir, "recommendation-command-run-ledger-rollup.json")

	if code := Run([]string{"mission", "recommendations", "schema-registry", "--out", registryPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("schema-registry failed")
	}
	if code := Run([]string{"mission", "recommendations", "validate-evidence", "--evidence-root", filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"), "--out", reportPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("validate-evidence failed")
	}
	var coverageOut bytes.Buffer
	if code := Run([]string{"mission", "recommendations", "schema-registry-coverage", "--registry", registryPath, "--validation-report", reportPath, "--out", coveragePath}, &coverageOut, &coverageOut); code == 0 {
		t.Fatalf("schema-registry-coverage should fail for rollup fixture setup: %s", coverageOut.String())
	}
	for _, spec := range []struct {
		command  string
		artifact string
		out      string
	}{
		{"schema-registry", registryPath, registryLedgerPath},
		{"validate-evidence", reportPath, reportLedgerPath},
		{"schema-registry-coverage", coveragePath, coverageLedgerPath},
	} {
		var out bytes.Buffer
		code := Run([]string{"mission", "recommendations", "run-ledger", "--command", spec.command, "--artifact", spec.artifact, "--out", spec.out}, &out, &out)
		if code != 0 {
			t.Fatalf("run-ledger %s failed: %s", spec.command, out.String())
		}
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-ledger-rollup",
		"--ledger", registryLedgerPath,
		"--ledger", reportLedgerPath,
		"--ledger", coverageLedgerPath,
		"--out", rollupPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger-rollup failed: %s", out.String())
	}
	for _, want := range []string{
		"status=rolled_up",
		"ledger_count=3",
		"failed_output_count=1",
		"all_ledgers_record_invocation=true",
		"all_outputs_no_promotion=true",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger-rollup output missing %q: %s", want, out.String())
		}
	}

	rollup := mustLoadJSON[AtlasRecommendationCommandRunLedgerRollup](t, rollupPath)
	if rollup.Schema != AtlasRecommendationCommandRunLedgerRollupContract ||
		rollup.Status != "rolled_up" ||
		rollup.LedgerCount != 3 ||
		len(rollup.Ledgers) != 3 ||
		rollup.FailedOutputCount != 1 ||
		len(rollup.FailedCommands) != 1 ||
		rollup.FailedCommands[0] != "schema-registry-coverage" ||
		rollup.OutputStatusCounts["ready"] != 1 ||
		rollup.OutputStatusCounts["passed"] != 1 ||
		rollup.OutputStatusCounts["failed"] != 1 ||
		!rollup.AllLedgersRecordInvocation ||
		!rollup.AllOutputsNoPromotion ||
		rollup.PromotionGranted ||
		rollup.ClaimsAuthorityAdvance ||
		!rollup.RSIRemainsDenied ||
		rollup.SafeToExecute ||
		rollup.SchedulesWork ||
		rollup.ExecutesWork ||
		rollup.ApprovesWork ||
		rollup.MutatesRepositories {
		t.Fatalf("run-ledger rollup lost aggregate safety state: %#v", rollup)
	}
	validator, err := validateRecommendationEvidenceTypedFile(rollupPath, AtlasRecommendationCommandRunLedgerRollupContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-command-run-ledger-rollup" {
		t.Fatalf("expected typed recommendation command run ledger rollup validator, got %s", validator)
	}
}

func TestMissionRecommendationsRunLedgerCoverageCheckRequiresEveryControlPlaneCommand(t *testing.T) {
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
	rollupPath := filepath.Join(tempDir, "recommendation-command-run-ledger-rollup.json")
	checkPath := filepath.Join(tempDir, "recommendation-command-run-ledger-coverage-check.json")
	if code := Run([]string{"mission", "recommendations", "schema-registry", "--out", registryPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("schema-registry failed")
	}

	commands := []string{
		"next-track",
		"consumed-ledger",
		"track-registry",
		"final-response-gates",
		"schema-registry",
		"validate-evidence",
		"schema-registry-coverage",
		"schema-health-repair-prompt",
	}
	entries := []AtlasRecommendationCommandRunLedgerRollupEntry{}
	digest := "sha256:" + strings.Repeat("a", 64)
	for _, command := range commands {
		entries = append(entries, AtlasRecommendationCommandRunLedgerRollupEntry{
			LedgerPath:             "docs/evidence/example/" + command + "-run-ledger.json",
			LedgerDigest:           digest,
			Command:                command,
			ArtifactSchema:         "ao.atlas." + command + ".v0.1",
			TypedValidator:         "typed:" + command,
			OutputStatus:           "passed",
			ArtifactPath:           "docs/evidence/example/" + command + ".json",
			ArtifactDigest:         digest,
			RecordsInvocation:      true,
			NoPromotionRequested:   true,
			PromotionGranted:       false,
			ClaimsAuthorityAdvance: false,
			RSIRemainsDenied:       true,
		})
	}
	rollup := AtlasRecommendationCommandRunLedgerRollup{
		Schema:                     AtlasRecommendationCommandRunLedgerRollupContract,
		Status:                     "rolled_up",
		LedgerCount:                len(entries),
		Ledgers:                    entries,
		Commands:                   commands,
		OutputStatusCounts:         map[string]int{"passed": len(commands)},
		FailedOutputCount:          0,
		FailedCommands:             []string{},
		AllLedgersRecordInvocation: true,
		AllOutputsNoPromotion:      true,
		PromotionGranted:           false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           true,
		SafeToExecute:              false,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		MutatesRepositories:        false,
	}
	if err := WriteJSON(rollupPath, rollup); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-ledger-coverage-check",
		"--registry", registryPath,
		"--rollup", rollupPath,
		"--out", checkPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger-coverage-check failed: %s", out.String())
	}
	for _, want := range []string{
		"status=passed",
		"required_command_count=8",
		"covered_command_count=8",
		"missing_command_count=0",
		"all_control_plane_commands_covered=true",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger-coverage-check output missing %q: %s", want, out.String())
		}
	}

	check := mustLoadJSON[AtlasRecommendationRunLedgerCoverageCheck](t, checkPath)
	if check.Schema != AtlasRecommendationRunLedgerCoverageCheckContract ||
		check.Status != "passed" ||
		check.RequiredCommandCount != 8 ||
		check.CoveredCommandCount != 8 ||
		check.MissingCommandCount != 0 ||
		len(check.MissingCommands) != 0 ||
		!check.AllControlPlaneCommandsCovered ||
		!check.AllOutputsNoPromotion ||
		check.PromotionGranted ||
		check.ClaimsAuthorityAdvance ||
		!check.RSIRemainsDenied {
		t.Fatalf("run ledger coverage check lost closure safety state: %#v", check)
	}
	validator, err := validateRecommendationEvidenceTypedFile(checkPath, AtlasRecommendationRunLedgerCoverageCheckContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-run-ledger-coverage-check" {
		t.Fatalf("expected typed run ledger coverage check validator, got %s", validator)
	}
}
