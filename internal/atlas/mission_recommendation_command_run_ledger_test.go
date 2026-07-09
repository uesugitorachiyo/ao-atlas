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
		"export-refactoring-wave",
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
		"required_command_count=9",
		"covered_command_count=9",
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
		check.RequiredCommandCount != 9 ||
		check.CoveredCommandCount != 9 ||
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

func TestMissionRecommendationsArtifactSummaryBindsRunLedgerRollupAndCoverageCheck(t *testing.T) {
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
	rollupPath := filepath.Join(tempDir, "recommendation-command-run-ledger-rollup.json")
	checkPath := filepath.Join(tempDir, "recommendation-command-run-ledger-coverage-check.json")
	if code := Run([]string{"mission", "recommendations", "schema-registry", "--out", registryPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("schema-registry failed")
	}

	registrySummary, err := BuildAtlasRecommendationArtifactSummary(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := BuildAtlasRecommendationCommandRunLedger("schema-registry", registryPath)
	if err != nil {
		t.Fatal(err)
	}
	if ledger.ArtifactPath != registrySummary.Path ||
		ledger.ArtifactDigest != registrySummary.Digest ||
		ledger.ArtifactSchema != registrySummary.Schema ||
		ledger.TypedValidator != registrySummary.TypedValidator ||
		ledger.OutputStatus != registrySummary.OutputStatus {
		t.Fatalf("run-ledger did not bind artifact summary: ledger=%#v summary=%#v", ledger, registrySummary)
	}
	if err := WriteAtlasRecommendationCommandRunLedger(ledgerPath, ledger); err != nil {
		t.Fatal(err)
	}

	ledgerSummary, err := BuildAtlasRecommendationArtifactSummary(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	rollup, err := BuildAtlasRecommendationCommandRunLedgerRollup([]string{ledgerPath})
	if err != nil {
		t.Fatal(err)
	}
	if len(rollup.Ledgers) != 1 {
		t.Fatalf("expected one rollup ledger entry: %#v", rollup)
	}
	entry := rollup.Ledgers[0]
	if entry.LedgerPath != ledgerSummary.PublicPath ||
		entry.LedgerDigest != ledgerSummary.Digest ||
		entry.ArtifactPath != registrySummary.PublicPath ||
		entry.ArtifactDigest != registrySummary.Digest ||
		entry.ArtifactSchema != registrySummary.Schema ||
		entry.TypedValidator != registrySummary.TypedValidator ||
		entry.OutputStatus != registrySummary.OutputStatus {
		t.Fatalf("rollup entry did not bind ledger and artifact summaries: entry=%#v ledger_summary=%#v artifact_summary=%#v", entry, ledgerSummary, registrySummary)
	}
	if err := WriteAtlasRecommendationCommandRunLedgerRollup(rollupPath, rollup); err != nil {
		t.Fatal(err)
	}

	rollupSummary, err := BuildAtlasRecommendationArtifactSummary(rollupPath)
	if err != nil {
		t.Fatal(err)
	}
	check, err := BuildAtlasRecommendationRunLedgerCoverageCheck(registryPath, rollupPath)
	if err != nil {
		t.Fatal(err)
	}
	if check.SourceRegistryPath != registrySummary.PublicPath ||
		check.SourceRegistryDigest != registrySummary.Digest ||
		check.SourceRollupPath != rollupSummary.PublicPath ||
		check.SourceRollupDigest != rollupSummary.Digest ||
		check.PromotionGranted ||
		check.ClaimsAuthorityAdvance ||
		!check.RSIRemainsDenied {
		t.Fatalf("coverage check did not bind source artifact summaries safely: check=%#v registry_summary=%#v rollup_summary=%#v", check, registrySummary, rollupSummary)
	}
	if err := WriteJSON(checkPath, check); err != nil {
		t.Fatal(err)
	}
}

func TestRecommendationRunLedgerOutputStatusClassificationCoversPassReadyFailedBlocked(t *testing.T) {
	cases := []struct {
		status       string
		category     string
		countsFailed bool
	}{
		{"passed", "pass", false},
		{"ready", "ready", false},
		{"failed", "failed", true},
		{"blocked", "blocked", true},
		{"blocked_hard_blocker", "blocked", true},
		{"unknown_status", "failed", true},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			classification := ClassifyAtlasRecommendationRunLedgerOutputStatus(tc.status)
			if classification.OutputStatus != tc.status ||
				classification.Category != tc.category ||
				classification.CountsAsFailedOutput != tc.countsFailed {
				t.Fatalf("unexpected classification for %s: %#v", tc.status, classification)
			}
		})
	}
}

func TestMissionRecommendationsRunLedgerRecordsRefactoringExporterAndRoutingArtifacts(t *testing.T) {
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

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-wave-v01"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	sourceAssertion := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json"
	tempDir := t.TempDir()
	decisionPath := filepath.Join(tempDir, "next-track-decision.json")
	consumedLedgerPath := filepath.Join(tempDir, "consumed-recommendation-ledger.json")
	refactoringPath := filepath.Join(tempDir, "next-wave-refactoring-recommendations.json")
	refactoringLedgerPath := filepath.Join(tempDir, "refactoring-export-run-ledger.json")

	if code := Run([]string{
		"mission", "recommendations", "next-track",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--out", decisionPath,
	}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("next-track failed")
	}
	if code := Run([]string{
		"mission", "recommendations", "consumed-ledger",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--next-track-decision", decisionPath,
		"--out", consumedLedgerPath,
	}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("consumed-ledger failed")
	}
	if code := Run([]string{
		"mission", "recommendations", "export-refactoring-wave",
		"--mission-id", "ao-atlas-refactoring-wave-v01",
		"--source-evidence-root", sourceRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--next-track-decision", decisionPath,
		"--consumed-ledger", consumedLedgerPath,
		"--min-tasks", "40",
		"--out", refactoringPath,
	}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("export-refactoring-wave failed")
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-ledger",
		"--command", "export-refactoring-wave",
		"--artifact", refactoringPath,
		"--out", refactoringLedgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-ledger export-refactoring-wave failed: %s", out.String())
	}
	for _, want := range []string{
		"command=export-refactoring-wave",
		"artifact_schema=ao.mission.refactoring-recommendations.v0.1",
		"typed_validator=typed:recommendation-refactoring-recommendations",
		"output_status=ready",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("run-ledger export-refactoring-wave output missing %q: %s", want, out.String())
		}
	}
	ledger := mustLoadJSON[AtlasRecommendationCommandRunLedger](t, refactoringLedgerPath)
	if ledger.Command != "export-refactoring-wave" ||
		ledger.ArtifactSchema != "ao.mission.refactoring-recommendations.v0.1" ||
		ledger.TypedValidator != "typed:recommendation-refactoring-recommendations" ||
		ledger.OutputStatus != "ready" ||
		ledger.PromotionGranted ||
		ledger.ClaimsAuthorityAdvance ||
		!ledger.RSIRemainsDenied {
		t.Fatalf("refactoring exporter ledger lost safe routing coverage: %#v", ledger)
	}
}

func TestRecommendationRunLedgerRollupBindsOperatorSummaryWithoutSelfReference(t *testing.T) {
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
	rollupPath := filepath.Join(tempDir, "recommendation-command-run-ledger-rollup.json")
	summaryPath := filepath.Join(tempDir, "operator-summary.md")
	summaryCheckPath := filepath.Join(tempDir, "operator-summary-check.json")
	readbackPath := filepath.Join("docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04", "nodes", "mission-recommendation-feature-depth-next-wave-39", "recommendation-readback-after.json")

	if code := Run([]string{"mission", "recommendations", "schema-registry", "--out", registryPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("schema-registry failed")
	}
	if code := Run([]string{"mission", "recommendations", "run-ledger", "--command", "schema-registry", "--artifact", registryPath, "--out", ledgerPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("run-ledger failed")
	}
	if code := Run([]string{"mission", "recommendations", "run-ledger-rollup", "--ledger", ledgerPath, "--out", rollupPath}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("run-ledger-rollup failed")
	}
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)
	if err := WriteAtlasMissionOperatorSummary(summaryPath, readback); err != nil {
		t.Fatal(err)
	}
	summaryCheck, err := BuildAtlasMissionOperatorSummaryCheck(readbackPath, summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(summaryCheckPath, summaryCheck); err != nil {
		t.Fatal(err)
	}

	binding, err := BuildAtlasRecommendationRunLedgerOperatorSummaryBinding(rollupPath, summaryCheckPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationRunLedgerOperatorSummaryBinding(binding); err != nil {
		t.Fatal(err)
	}
	if binding.Schema != AtlasRecommendationRunLedgerOperatorSummaryBindingContract ||
		binding.Status != "bound" ||
		binding.SourceRollupPath != publicArtifactRef(rollupPath) ||
		binding.SourceOperatorSummaryCheckPath != publicArtifactRef(summaryCheckPath) ||
		binding.RollupLedgerCount != 1 ||
		binding.SummaryRequiresOwnRunLedger ||
		binding.RollupRequiresSummaryRunLedger ||
		binding.SelfReferentialLedgerRequirement ||
		!binding.AllOutputsNoPromotion ||
		binding.PromotionGranted ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute ||
		binding.SchedulesWork ||
		binding.ExecutesWork ||
		binding.ApprovesWork ||
		binding.MutatesRepositories {
		t.Fatalf("rollup operator summary binding lost no-self-reference safety: %#v", binding)
	}
}

func TestRecommendationRunLedgerRetryFixturePackCoversRetriesAndResumedSessions(t *testing.T) {
	pack, err := BuildAtlasRecommendationRunLedgerRetryFixturePack("refactoring-next-wave-20", []AtlasRecommendationRunLedgerRetryAttempt{
		{Command: "schema-registry-coverage", SessionID: "session-1", Attempt: 1, OutputStatus: "failed", RetryReason: "missing_registry_schema"},
		{Command: "schema-registry-coverage", SessionID: "session-1", Attempt: 2, OutputStatus: "passed", RetryReason: "safe_retry_after_fixture_repair"},
		{Command: "validate-evidence", SessionID: "session-2", Attempt: 1, OutputStatus: "blocked_hard_blocker", RetryReason: "resume_after_compaction"},
		{Command: "validate-evidence", SessionID: "session-3", Attempt: 2, OutputStatus: "passed", RetryReason: "resumed_session_replay"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRecommendationRunLedgerRetryFixturePack(pack); err != nil {
		t.Fatal(err)
	}
	if pack.Schema != AtlasRecommendationRunLedgerRetryFixturePackContract ||
		pack.Status != "covered" ||
		pack.NodeID != "refactoring-next-wave-20" ||
		pack.AttemptCount != 4 ||
		pack.RetryCommandCount != 2 ||
		pack.ResumedSessionCount != 2 ||
		pack.FailedOrBlockedAttemptCount != 2 ||
		!pack.AllAttemptsClassified ||
		!pack.RetryReplayPlanningOnly ||
		pack.PromotionGranted ||
		pack.ClaimsAuthorityAdvance ||
		!pack.RSIRemainsDenied ||
		pack.SafeToExecute ||
		pack.SchedulesWork ||
		pack.ExecutesWork ||
		pack.ApprovesWork ||
		pack.MutatesRepositories {
		t.Fatalf("retry fixture pack lost retry/resume safety state: %#v", pack)
	}
	categories := map[string]int{}
	for _, attempt := range pack.Attempts {
		categories[attempt.OutputCategory]++
		if attempt.StatusClassification.OutputStatus != attempt.OutputStatus ||
			attempt.StatusClassification.Category != attempt.OutputCategory {
			t.Fatalf("attempt classification drifted: %#v", attempt)
		}
	}
	if categories["pass"] != 2 || categories["failed"] != 1 || categories["blocked"] != 1 {
		t.Fatalf("retry fixture pack did not cover pass/failed/blocked attempts: %#v", categories)
	}
}
