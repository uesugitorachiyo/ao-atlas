package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const AtlasRecommendationEvidenceValidationReportContract = "ao.atlas.recommendation-evidence-validation-report.v0.1"

type AtlasRecommendationEvidenceValidationReport struct {
	Schema                    string                                       `json:"schema"`
	Status                    string                                       `json:"status"`
	EvidenceRoot              string                                       `json:"evidence_root"`
	NodeRoot                  string                                       `json:"node_root"`
	NodeCount                 int                                          `json:"node_count"`
	JSONFileCount             int                                          `json:"json_file_count"`
	ValidatedJSONFiles        int                                          `json:"validated_json_files"`
	SchemaBoundFiles          int                                          `json:"schema_bound_files"`
	TypedValidatorFiles       int                                          `json:"typed_validator_files"`
	GenericSchemaFiles        int                                          `json:"generic_schema_files"`
	MissingSchemaFiles        []string                                     `json:"missing_schema_files"`
	FailedFiles               []string                                     `json:"failed_files"`
	MissingRequiredFiles      []string                                     `json:"missing_required_files"`
	RequiredFilenames         []string                                     `json:"required_filenames"`
	RequiredFilenamesCovered  bool                                         `json:"required_filenames_covered"`
	StrictSchemaRegistry      bool                                         `json:"strict_schema_registry,omitempty"`
	RequireProvenanceFields   bool                                         `json:"require_provenance_fields,omitempty"`
	UnknownSchemaFiles        []string                                     `json:"unknown_schema_files,omitempty"`
	MissingSourceDigestFiles  []string                                     `json:"missing_source_digest_files,omitempty"`
	MissingEvidenceClassFiles []string                                     `json:"missing_evidence_class_files,omitempty"`
	SchemaCounts              map[string]int                               `json:"schema_counts"`
	Validators                map[string]int                               `json:"validators"`
	Entries                   []AtlasRecommendationEvidenceValidationEntry `json:"entries"`
}

type AtlasRecommendationEvidenceValidationEntry struct {
	Path          string `json:"path"`
	NodeID        string `json:"node_id"`
	Filename      string `json:"filename"`
	Schema        string `json:"schema"`
	SourceDigest  string `json:"source_digest,omitempty"`
	EvidenceClass string `json:"evidence_class,omitempty"`
	Validator     string `json:"validator"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

type AtlasRecommendationEvidenceValidationOptions struct {
	StrictSchemaRegistry    bool
	RequireProvenanceFields bool
}

func BuildAtlasRecommendationEvidenceValidationReport(evidenceRoot string) (AtlasRecommendationEvidenceValidationReport, error) {
	return BuildAtlasRecommendationEvidenceValidationReportWithOptions(evidenceRoot, false)
}

func BuildAtlasRecommendationEvidenceValidationReportWithOptions(evidenceRoot string, strictSchemaRegistry bool) (AtlasRecommendationEvidenceValidationReport, error) {
	return BuildAtlasRecommendationEvidenceValidationReportWithValidationOptions(evidenceRoot, AtlasRecommendationEvidenceValidationOptions{StrictSchemaRegistry: strictSchemaRegistry})
}

func BuildAtlasRecommendationEvidenceValidationReportWithValidationOptions(evidenceRoot string, options AtlasRecommendationEvidenceValidationOptions) (AtlasRecommendationEvidenceValidationReport, error) {
	evidenceRoot = strings.TrimSpace(evidenceRoot)
	report := AtlasRecommendationEvidenceValidationReport{
		Schema:                    AtlasRecommendationEvidenceValidationReportContract,
		Status:                    "passed",
		EvidenceRoot:              filepath.ToSlash(evidenceRoot),
		NodeRoot:                  filepath.ToSlash(filepath.Join(evidenceRoot, "nodes")),
		SchemaCounts:              map[string]int{},
		Validators:                map[string]int{},
		RequiredFilenames:         requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered:  true,
		StrictSchemaRegistry:      options.StrictSchemaRegistry,
		RequireProvenanceFields:   options.RequireProvenanceFields,
		UnknownSchemaFiles:        []string{},
		MissingSourceDigestFiles:  []string{},
		MissingEvidenceClassFiles: []string{},
		MissingSchemaFiles:        []string{},
		FailedFiles:               []string{},
		MissingRequiredFiles:      []string{},
		Entries:                   []AtlasRecommendationEvidenceValidationEntry{},
	}
	if evidenceRoot == "" {
		report.Status = "failed"
		return report, fmt.Errorf("evidence root is required")
	}
	nodeRoot := filepath.Join(evidenceRoot, "nodes")
	if info, err := os.Stat(nodeRoot); err != nil {
		report.Status = "failed"
		return report, err
	} else if !info.IsDir() {
		report.Status = "failed"
		return report, fmt.Errorf("node evidence root must be a directory: %s", nodeRoot)
	}

	paths := []string{}
	if err := filepath.WalkDir(nodeRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		report.Status = "failed"
		return report, err
	}
	sort.Strings(paths)

	nodeFiles := map[string]map[string]bool{}
	for _, path := range paths {
		entry := validateRecommendationEvidenceJSONFileWithValidationOptions(evidenceRoot, nodeRoot, path, options)
		report.Entries = append(report.Entries, entry)
		report.JSONFileCount++
		if entry.NodeID != "" {
			if _, ok := nodeFiles[entry.NodeID]; !ok {
				nodeFiles[entry.NodeID] = map[string]bool{}
			}
			nodeFiles[entry.NodeID][entry.Filename] = true
		}
		if entry.Status == "passed" {
			report.ValidatedJSONFiles++
		} else {
			report.FailedFiles = append(report.FailedFiles, entry.Path)
			if entry.Validator == "strict:unknown-schema" {
				report.UnknownSchemaFiles = append(report.UnknownSchemaFiles, entry.Path)
			}
			if strings.Contains(entry.Error, "missing source_digest") {
				report.MissingSourceDigestFiles = append(report.MissingSourceDigestFiles, entry.Path)
			}
			if strings.Contains(entry.Error, "missing evidence_class") {
				report.MissingEvidenceClassFiles = append(report.MissingEvidenceClassFiles, entry.Path)
			}
		}
		if entry.Schema == "" {
			report.MissingSchemaFiles = append(report.MissingSchemaFiles, entry.Path)
		} else {
			report.SchemaBoundFiles++
			report.SchemaCounts[entry.Schema]++
		}
		validatorName := strings.TrimPrefix(entry.Validator, "strict:")
		if strings.HasPrefix(validatorName, "typed:") {
			report.TypedValidatorFiles++
		} else if validatorName == "generic:schema-marker" || validatorName == "strict:generic:schema-marker" {
			report.GenericSchemaFiles++
		}
		if entry.Validator != "" {
			report.Validators[entry.Validator]++
		}
	}
	report.NodeCount = len(nodeFiles)

	for _, nodeID := range sortedMapKeys(nodeFiles) {
		for _, filename := range report.RequiredFilenames {
			if !nodeFiles[nodeID][filename] {
				report.MissingRequiredFiles = append(report.MissingRequiredFiles, filepath.ToSlash(filepath.Join(nodeID, filename)))
			}
		}
	}
	if len(report.MissingRequiredFiles) != 0 {
		report.RequiredFilenamesCovered = false
	}
	if report.JSONFileCount == 0 || report.ValidatedJSONFiles != report.JSONFileCount || len(report.MissingSchemaFiles) != 0 || len(report.FailedFiles) != 0 || !report.RequiredFilenamesCovered {
		report.Status = "failed"
		return report, fmt.Errorf("recommendation evidence validation failed")
	}
	return report, nil
}

func ValidateAtlasRecommendationEvidenceValidationReport(report AtlasRecommendationEvidenceValidationReport) error {
	var errs []string
	requireContract(&errs, "recommendation_evidence_validation_report", report.Schema, AtlasRecommendationEvidenceValidationReportContract)
	if !oneOf(report.Status, "passed", "failed") {
		errs = append(errs, "status must be passed or failed")
	}
	requireField(&errs, "evidence_root", report.EvidenceRoot)
	requireField(&errs, "node_root", report.NodeRoot)
	if report.NodeCount < 0 {
		errs = append(errs, "node_count must be non-negative")
	}
	if report.JSONFileCount < 0 || report.ValidatedJSONFiles < 0 || report.SchemaBoundFiles < 0 || report.TypedValidatorFiles < 0 || report.GenericSchemaFiles < 0 {
		errs = append(errs, "file counts must be non-negative")
	}
	if report.ValidatedJSONFiles > report.JSONFileCount {
		errs = append(errs, "validated_json_files must not exceed json_file_count")
	}
	if report.SchemaBoundFiles > report.JSONFileCount {
		errs = append(errs, "schema_bound_files must not exceed json_file_count")
	}
	if report.TypedValidatorFiles+report.GenericSchemaFiles > report.JSONFileCount {
		errs = append(errs, "validator file counts must not exceed json_file_count")
	}
	if len(report.RequiredFilenames) == 0 {
		errs = append(errs, "required_filenames must not be empty")
	}
	if report.RequiredFilenamesCovered != (len(report.MissingRequiredFiles) == 0) {
		errs = append(errs, "required_filenames_covered must match missing_required_files")
	}
	if report.Status == "passed" {
		if report.JSONFileCount == 0 {
			errs = append(errs, "passed status requires json files")
		}
		if report.ValidatedJSONFiles != report.JSONFileCount {
			errs = append(errs, "passed status requires every json file to validate")
		}
		if len(report.MissingSchemaFiles) != 0 {
			errs = append(errs, "passed status requires no missing schema files")
		}
		if len(report.FailedFiles) != 0 {
			errs = append(errs, "passed status requires no failed files")
		}
		if !report.RequiredFilenamesCovered {
			errs = append(errs, "passed status requires required filenames to be covered")
		}
	}
	for _, entry := range report.Entries {
		requireField(&errs, "validation_entry.path", entry.Path)
		requireField(&errs, "validation_entry.filename", entry.Filename)
		if !oneOf(entry.Status, "passed", "failed") {
			errs = append(errs, "validation entry status must be passed or failed")
		}
		if entry.Status == "passed" {
			requireField(&errs, "validation_entry.schema", entry.Schema)
			requireField(&errs, "validation_entry.validator", entry.Validator)
		}
		if entry.Status == "failed" && strings.TrimSpace(entry.Error) == "" {
			errs = append(errs, "failed validation entry requires error")
		}
	}
	return joinErrors(errs)
}

func validateRecommendationEvidenceJSONFile(evidenceRoot, nodeRoot, path string) AtlasRecommendationEvidenceValidationEntry {
	return validateRecommendationEvidenceJSONFileWithOptions(evidenceRoot, nodeRoot, path, false)
}

func validateRecommendationEvidenceJSONFileWithOptions(evidenceRoot, nodeRoot, path string, strictSchemaRegistry bool) AtlasRecommendationEvidenceValidationEntry {
	return validateRecommendationEvidenceJSONFileWithValidationOptions(evidenceRoot, nodeRoot, path, AtlasRecommendationEvidenceValidationOptions{StrictSchemaRegistry: strictSchemaRegistry})
}

func validateRecommendationEvidenceJSONFileWithValidationOptions(evidenceRoot, nodeRoot, path string, options AtlasRecommendationEvidenceValidationOptions) AtlasRecommendationEvidenceValidationEntry {
	rel, err := filepath.Rel(evidenceRoot, path)
	if err != nil {
		rel = path
	}
	nodeRel, err := filepath.Rel(nodeRoot, path)
	if err != nil {
		nodeRel = path
	}
	parts := strings.Split(filepath.ToSlash(nodeRel), "/")
	nodeID := ""
	if len(parts) > 0 {
		nodeID = parts[0]
	}
	entry := AtlasRecommendationEvidenceValidationEntry{
		Path:     filepath.ToSlash(rel),
		NodeID:   nodeID,
		Filename: filepath.Base(path),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
		return entry
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
		return entry
	}
	entry.Schema = evidenceSchemaMarker(raw)
	if entry.Schema == "" {
		entry.Status = "failed"
		entry.Error = "missing schema or contract_version"
		return entry
	}
	validator, err := validateRecommendationEvidenceTypedFile(path, entry.Schema)
	if options.StrictSchemaRegistry {
		validator, err = validateRecommendationEvidenceTypedFileStrict(path, entry.Schema)
	}
	entry.Validator = validator
	if err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
		return entry
	}
	if options.RequireProvenanceFields {
		entry.SourceDigest, _ = raw["source_digest"].(string)
		entry.SourceDigest = strings.TrimSpace(entry.SourceDigest)
		entry.EvidenceClass, _ = raw["evidence_class"].(string)
		entry.EvidenceClass = strings.TrimSpace(entry.EvidenceClass)
		var errs []string
		if !strings.HasPrefix(entry.SourceDigest, "sha256:") || len(entry.SourceDigest) <= len("sha256:") {
			errs = append(errs, "missing source_digest")
		}
		if entry.EvidenceClass == "" {
			errs = append(errs, "missing evidence_class")
		}
		if len(errs) != 0 {
			entry.Status = "failed"
			entry.Error = strings.Join(errs, "; ")
			return entry
		}
	}
	entry.Status = "passed"
	return entry
}

func evidenceSchemaMarker(raw map[string]any) string {
	for _, key := range []string{"schema", "contract_version"} {
		value, _ := raw[key].(string)
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func validateRecommendationEvidenceTypedFile(path, schema string) (string, error) {
	switch schema {
	case WorkgraphContract:
		value, err := LoadJSON[Workgraph](path)
		if err != nil {
			return "typed:workgraph", err
		}
		return "typed:workgraph", ValidateWorkgraph(value)
	case RunLinkContract:
		value, err := LoadJSON[RunLink](path)
		if err != nil {
			return "typed:run-link", err
		}
		return "typed:run-link", ValidateRunLink(value)
	case FactoryTaskContract:
		value, err := LoadJSON[FactoryTask](path)
		if err != nil {
			return "typed:factory-task", err
		}
		return "typed:factory-task", ValidateFactoryTask(value)
	case FoundryImportContract:
		value, err := LoadJSON[FoundryImport](path)
		if err != nil {
			return "typed:foundry-import", err
		}
		return "typed:foundry-import", ValidateFoundryImport(value)
	case FoundryContinuationHandoffContract:
		value, err := LoadJSON[FoundryContinuationHandoff](path)
		if err != nil {
			return "typed:foundry-continuation-handoff", err
		}
		return "typed:foundry-continuation-handoff", ValidateFoundryContinuationHandoff(value)
	case ConsolidationRepositoryBaselineContract:
		value, err := LoadJSON[ConsolidationRepositoryBaseline](path)
		if err != nil {
			return "typed:consolidation-repository-baseline", err
		}
		return "typed:consolidation-repository-baseline", ValidateConsolidationRepositoryBaseline(value)
	case AtlasRecommendationReadbackContract:
		value, err := LoadJSON[AtlasRecommendationReadback](path)
		if err != nil {
			return "typed:recommendation-readback", err
		}
		return "typed:recommendation-readback", ValidateAtlasRecommendationReadback(value)
	case MissionLifecycleMetricsEvidenceContract:
		value, err := LoadJSON[MissionLifecycleMetricsEvidence](path)
		if err != nil {
			return "typed:mission-lifecycle-metrics", err
		}
		return "typed:mission-lifecycle-metrics", ValidateMissionLifecycleMetricsEvidence(value)
	case AtlasMissionReadbackDeltaContract:
		value, err := LoadJSON[AtlasMissionReadbackDelta](path)
		if err != nil {
			return "typed:mission-readback-delta", err
		}
		return "typed:mission-readback-delta", ValidateAtlasMissionReadbackDelta(value)
	case AtlasMissionReadbackDiffFixtureContract:
		value, err := LoadJSON[AtlasMissionReadbackDiffFixture](path)
		if err != nil {
			return "typed:mission-readback-diff-fixture", err
		}
		return "typed:mission-readback-diff-fixture", ValidateAtlasMissionReadbackDiffFixture(value)
	case AtlasMissionStaleCheckpointRejectionContract:
		value, err := LoadJSON[AtlasMissionStaleCheckpointRejection](path)
		if err != nil {
			return "typed:mission-stale-checkpoint-rejection", err
		}
		return "typed:mission-stale-checkpoint-rejection", ValidateAtlasMissionStaleCheckpointRejection(value)
	case AtlasMissionOperatorSummaryCheckContract:
		value, err := LoadJSON[AtlasMissionOperatorSummaryCheck](path)
		if err != nil {
			return "typed:mission-operator-summary-check", err
		}
		return "typed:mission-operator-summary-check", ValidateAtlasMissionOperatorSummaryCheck(value)
	case AtlasNodeCommandReadbackContract:
		value, err := LoadJSON[AtlasNodeCommandReadbackEvidence](path)
		if err != nil {
			return "typed:command-readback", err
		}
		return "typed:command-readback", ValidateAtlasNodeCommandReadbackEvidence(value)
	case AtlasNodePromoterNoPromotionContract:
		value, err := LoadJSON[AtlasNodePromoterNoPromotionEvidence](path)
		if err != nil {
			return "typed:promoter-no-promotion", err
		}
		return "typed:promoter-no-promotion", ValidateAtlasNodePromoterNoPromotionEvidence(value)
	case AtlasNodeSentinelPublicSafetyContract:
		value, err := LoadJSON[AtlasNodeSentinelPublicSafetyEvidence](path)
		if err != nil {
			return "typed:sentinel-public-safety", err
		}
		return "typed:sentinel-public-safety", ValidateAtlasNodeSentinelPublicSafetyEvidence(value)
	case AtlasRunLinkSchemaCoverageContract:
		value, err := LoadJSON[AtlasRunLinkSchemaCoverage](path)
		if err != nil {
			return "typed:run-link-schema-coverage", err
		}
		return "typed:run-link-schema-coverage", ValidateAtlasRunLinkSchemaCoverage(value)
	case AtlasSchemaValidatorDriftContract:
		value, err := LoadJSON[AtlasSchemaValidatorDriftEvidence](path)
		if err != nil {
			return "typed:schema-validator-drift", err
		}
		return "typed:schema-validator-drift", ValidateAtlasSchemaValidatorDriftEvidence(value)
	case AtlasPRCITimingSummaryContract:
		value, err := LoadJSON[AtlasPRCITimingSummary](path)
		if err != nil {
			return "typed:pr-ci-timing-summary", err
		}
		return "typed:pr-ci-timing-summary", ValidateAtlasPRCITimingSummary(value)
	case AtlasPRCIWindowsThresholdEvidenceContract:
		value, err := LoadJSON[AtlasPRCIWindowsThresholdEvidence](path)
		if err != nil {
			return "typed:pr-ci-windows-threshold-evidence", err
		}
		return "typed:pr-ci-windows-threshold-evidence", ValidateAtlasPRCIWindowsThresholdEvidence(value)
	case AtlasP0BWindowsCIWaitTelemetryContract:
		value, err := LoadJSON[AtlasP0BWindowsCIWaitTelemetry](path)
		if err != nil {
			return "typed:p0b-windows-ci-wait-telemetry", err
		}
		return "typed:p0b-windows-ci-wait-telemetry", ValidateAtlasP0BWindowsCIWaitTelemetry(value)
	case AtlasP0BRollbackAuditContract:
		value, err := LoadJSON[AtlasP0BRollbackAudit](path)
		if err != nil {
			return "typed:p0b-rollback-audit", err
		}
		return "typed:p0b-rollback-audit", ValidateAtlasP0BRollbackAudit(value)
	case AtlasP0BCommandPromoterAgreementContract:
		value, err := LoadJSON[AtlasP0BCommandPromoterAgreement](path)
		if err != nil {
			return "typed:p0b-command-promoter-agreement", err
		}
		return "typed:p0b-command-promoter-agreement", ValidateAtlasP0BCommandPromoterAgreement(value)
	case AtlasFailedCheckReplayFixtureContract:
		value, err := LoadJSON[AtlasFailedCheckReplayFixture](path)
		if err != nil {
			return "typed:failed-check-replay-fixture", err
		}
		return "typed:failed-check-replay-fixture", ValidateAtlasFailedCheckReplayFixture(value)
	case AtlasCommandCovenantRejectedTicketFixtureContract:
		value, err := LoadJSON[AtlasCommandCovenantRejectedTicketFixture](path)
		if err != nil {
			return "typed:command-covenant-rejected-ticket-fixture", err
		}
		return "typed:command-covenant-rejected-ticket-fixture", ValidateAtlasCommandCovenantRejectedTicketFixture(value)
	case AtlasCommandCovenantQuarantineFixtureContract:
		value, err := LoadJSON[AtlasCommandCovenantQuarantineFixture](path)
		if err != nil {
			return "typed:command-covenant-quarantine-fixture", err
		}
		return "typed:command-covenant-quarantine-fixture", ValidateAtlasCommandCovenantQuarantineFixture(value)
	case AtlasCommandTicketBytePreservationFixtureContract:
		value, err := LoadJSON[AtlasCommandTicketBytePreservationFixture](path)
		if err != nil {
			return "typed:command-ticket-byte-preservation-fixture", err
		}
		return "typed:command-ticket-byte-preservation-fixture", ValidateAtlasCommandTicketBytePreservationFixture(value)
	case AtlasTicketDigestReadbackBindingFixtureContract:
		value, err := LoadJSON[AtlasTicketDigestReadbackBindingFixture](path)
		if err != nil {
			return "typed:ticket-digest-readback-binding-fixture", err
		}
		return "typed:ticket-digest-readback-binding-fixture", ValidateAtlasTicketDigestReadbackBindingFixture(value)
	case AtlasPolicyHashMismatchRejectionFixtureContract:
		value, err := LoadJSON[AtlasPolicyHashMismatchRejectionFixture](path)
		if err != nil {
			return "typed:policy-hash-mismatch-rejection-fixture", err
		}
		return "typed:policy-hash-mismatch-rejection-fixture", ValidateAtlasPolicyHashMismatchRejectionFixture(value)
	case AtlasPolicyVersionReplayRejectionFixtureContract:
		value, err := LoadJSON[AtlasPolicyVersionReplayRejectionFixture](path)
		if err != nil {
			return "typed:policy-version-replay-rejection-fixture", err
		}
		return "typed:policy-version-replay-rejection-fixture", ValidateAtlasPolicyVersionReplayRejectionFixture(value)
	case AtlasCovenantEvidenceDigestReadbackFixtureContract:
		value, err := LoadJSON[AtlasCovenantEvidenceDigestReadbackFixture](path)
		if err != nil {
			return "typed:covenant-evidence-digest-readback-fixture", err
		}
		return "typed:covenant-evidence-digest-readback-fixture", ValidateAtlasCovenantEvidenceDigestReadbackFixture(value)
	case AtlasCommandCompactRejectionReasonFixtureContract:
		value, err := LoadJSON[AtlasCommandCompactRejectionReasonFixture](path)
		if err != nil {
			return "typed:command-compact-rejection-reason-fixture", err
		}
		return "typed:command-compact-rejection-reason-fixture", ValidateAtlasCommandCompactRejectionReasonFixture(value)
	case AtlasBlueprintTicketSchemaCompatibilityLedgerContract:
		value, err := LoadJSON[AtlasBlueprintTicketSchemaCompatibilityLedger](path)
		if err != nil {
			return "typed:blueprint-ticket-schema-compatibility-ledger", err
		}
		return "typed:blueprint-ticket-schema-compatibility-ledger", ValidateAtlasBlueprintTicketSchemaCompatibilityLedger(value)
	case AtlasTicketSchemaCompatibilityLedgerContract:
		value, err := LoadJSON[AtlasTicketSchemaCompatibilityLedger](path)
		if err != nil {
			return "typed:atlas-ticket-schema-compatibility-ledger", err
		}
		return "typed:atlas-ticket-schema-compatibility-ledger", ValidateAtlasTicketSchemaCompatibilityLedger(value)
	case AtlasFoundryTicketSchemaCompatibilityLedgerContract:
		value, err := LoadJSON[AtlasFoundryTicketSchemaCompatibilityLedger](path)
		if err != nil {
			return "typed:foundry-ticket-schema-compatibility-ledger", err
		}
		return "typed:foundry-ticket-schema-compatibility-ledger", ValidateAtlasFoundryTicketSchemaCompatibilityLedger(value)
	case AtlasCommandTicketSchemaCompatibilityLedgerContract:
		value, err := LoadJSON[AtlasCommandTicketSchemaCompatibilityLedger](path)
		if err != nil {
			return "typed:command-ticket-schema-compatibility-ledger", err
		}
		return "typed:command-ticket-schema-compatibility-ledger", ValidateAtlasCommandTicketSchemaCompatibilityLedger(value)
	case AtlasCovenantTicketSchemaAuthorityLedgerContract:
		value, err := LoadJSON[AtlasCovenantTicketSchemaAuthorityLedger](path)
		if err != nil {
			return "typed:covenant-ticket-schema-authority-ledger", err
		}
		return "typed:covenant-ticket-schema-authority-ledger", ValidateAtlasCovenantTicketSchemaAuthorityLedger(value)
	case AtlasPolicyTicketPublicSafetyScanContract:
		value, err := LoadJSON[AtlasPolicyTicketPublicSafetyScan](path)
		if err != nil {
			return "typed:policy-ticket-public-safety-scan", err
		}
		return "typed:policy-ticket-public-safety-scan", ValidateAtlasPolicyTicketPublicSafetyScan(value)
	case AtlasP0BPRCILedgerContract:
		value, err := LoadJSON[AtlasP0BPRCILedger](path)
		if err != nil {
			return "typed:p0b-pr-ci-ledger", err
		}
		return "typed:p0b-pr-ci-ledger", ValidateAtlasP0BPRCILedger(value)
	case AtlasP0CReadinessCriteriaContract:
		value, err := LoadJSON[AtlasP0CReadinessCriteria](path)
		if err != nil {
			return "typed:p0c-readiness-criteria", err
		}
		return "typed:p0c-readiness-criteria", ValidateAtlasP0CReadinessCriteria(value)
	case AtlasP0CMissionFoundryHandoffCheckContract:
		value, err := LoadJSON[AtlasP0CMissionFoundryHandoffCheck](path)
		if err != nil {
			return "typed:p0c-mission-foundry-handoff-check", err
		}
		return "typed:p0c-mission-foundry-handoff-check", ValidateAtlasP0CMissionFoundryHandoffCheck(value)
	case AtlasMergeCheckBindingContract:
		value, err := LoadJSON[AtlasMergeCheckBinding](path)
		if err != nil {
			return "typed:merge-check-binding", err
		}
		return "typed:merge-check-binding", ValidateAtlasMergeCheckBinding(value)
	case AtlasPostMergeBranchDeletionReadbackContract:
		value, err := LoadJSON[AtlasPostMergeBranchDeletionReadback](path)
		if err != nil {
			return "typed:post-merge-branch-deletion-readback", err
		}
		return "typed:post-merge-branch-deletion-readback", ValidateAtlasPostMergeBranchDeletionReadback(value)
	case AtlasStaleRemoteBranchRepairContract:
		value, err := LoadJSON[AtlasStaleRemoteBranchRepair](path)
		if err != nil {
			return "typed:stale-remote-branch-repair", err
		}
		return "typed:stale-remote-branch-repair", ValidateAtlasStaleRemoteBranchRepair(value)
	case AtlasLocalMainSyncReadbackContract:
		value, err := LoadJSON[AtlasLocalMainSyncReadback](path)
		if err != nil {
			return "typed:local-main-sync-readback", err
		}
		return "typed:local-main-sync-readback", ValidateAtlasLocalMainSyncReadback(value)
	case AtlasBranchCleanupHandoffSummaryContract:
		value, err := LoadJSON[AtlasBranchCleanupHandoffSummary](path)
		if err != nil {
			return "typed:branch-cleanup-handoff-summary", err
		}
		return "typed:branch-cleanup-handoff-summary", ValidateAtlasBranchCleanupHandoffSummary(value)
	case AtlasCompactionResumePromptContract:
		value, err := LoadJSON[AtlasCompactionResumePrompt](path)
		if err != nil {
			return "typed:compaction-resume-prompt", err
		}
		return "typed:compaction-resume-prompt", ValidateAtlasCompactionResumePrompt(value)
	case AtlasCompactionResumeRegressionContract:
		value, err := LoadJSON[AtlasCompactionResumeRegression](path)
		if err != nil {
			return "typed:compaction-resume-regression", err
		}
		return "typed:compaction-resume-regression", ValidateAtlasCompactionResumeRegression(value)
	case AtlasResumeDenialEvidenceContract:
		value, err := LoadJSON[AtlasResumeDenialEvidence](path)
		if err != nil {
			return "typed:resume-denial-evidence", err
		}
		return "typed:resume-denial-evidence", ValidateAtlasResumeDenialEvidence(value)
	case AtlasPublicSafetyReadbackBindingContract:
		value, err := LoadJSON[AtlasPublicSafetyReadbackBinding](path)
		if err != nil {
			return "typed:public-safety-readback-binding", err
		}
		return "typed:public-safety-readback-binding", ValidateAtlasPublicSafetyReadbackBinding(value)
	case AtlasScopedPublicSafetyScanContract:
		value, err := LoadJSON[AtlasScopedPublicSafetyScan](path)
		if err != nil {
			return "typed:scoped-public-safety-scan", err
		}
		return "typed:scoped-public-safety-scan", ValidateAtlasScopedPublicSafetyScan(value)
	case AtlasAuthorityPromotionNegativeFixturesContract:
		value, err := LoadJSON[AtlasAuthorityPromotionNegativeFixtures](path)
		if err != nil {
			return "typed:authority-promotion-negative-fixtures", err
		}
		return "typed:authority-promotion-negative-fixtures", ValidateAtlasAuthorityPromotionNegativeFixtures(value)
	case AtlasPublicSafetyCoverageRollupContract:
		value, err := LoadJSON[AtlasPublicSafetyCoverageRollup](path)
		if err != nil {
			return "typed:public-safety-coverage-rollup", err
		}
		return "typed:public-safety-coverage-rollup", ValidateAtlasPublicSafetyCoverageRollup(value)
	case AtlasPromoterNoPromotionRollupContract:
		value, err := LoadJSON[AtlasPromoterNoPromotionRollup](path)
		if err != nil {
			return "typed:promoter-no-promotion-rollup", err
		}
		return "typed:promoter-no-promotion-rollup", ValidateAtlasPromoterNoPromotionRollup(value)
	case AtlasCommandPromoterAgreementRollupContract:
		value, err := LoadJSON[AtlasCommandPromoterAgreementRollup](path)
		if err != nil {
			return "typed:command-promoter-agreement-rollup", err
		}
		return "typed:command-promoter-agreement-rollup", ValidateAtlasCommandPromoterAgreementRollup(value)
	case AtlasPromoterRollupCountMismatchRegressionContract:
		value, err := LoadJSON[AtlasPromoterRollupCountMismatchRegression](path)
		if err != nil {
			return "typed:promoter-rollup-count-mismatch-regression", err
		}
		return "typed:promoter-rollup-count-mismatch-regression", ValidateAtlasPromoterRollupCountMismatchRegression(value)
	case AtlasCommandPromoterDisagreementDenialContract:
		value, err := LoadJSON[AtlasCommandPromoterDisagreementDenial](path)
		if err != nil {
			return "typed:command-promoter-disagreement-denial", err
		}
		return "typed:command-promoter-disagreement-denial", ValidateAtlasCommandPromoterDisagreementDenial(value)
	case AtlasFoundryImportReadinessBindingContract:
		value, err := LoadJSON[AtlasFoundryImportReadinessBinding](path)
		if err != nil {
			return "typed:foundry-import-readiness-binding", err
		}
		return "typed:foundry-import-readiness-binding", ValidateAtlasFoundryImportReadinessBinding(value)
	case AtlasRunLinkDigestCheckContract:
		value, err := LoadJSON[AtlasRunLinkDigestCheck](path)
		if err != nil {
			return "typed:run-link-digest-check", err
		}
		return "typed:run-link-digest-check", ValidateAtlasRunLinkDigestCheck(value)
	case AtlasFoundryHandoffReplayFixtureContract:
		value, err := LoadJSON[AtlasFoundryHandoffReplayFixture](path)
		if err != nil {
			return "typed:foundry-handoff-replay-fixture", err
		}
		return "typed:foundry-handoff-replay-fixture", ValidateAtlasFoundryHandoffReplayFixture(value)
	case AtlasFoundryTerminalStatusExamplesContract:
		value, err := LoadJSON[AtlasFoundryTerminalStatusExamplesValidation](path)
		if err != nil {
			return "typed:foundry-terminal-status-examples", err
		}
		return "typed:foundry-terminal-status-examples", ValidateAtlasFoundryTerminalStatusExamplesValidation(value)
	case AtlasMissionDashboardClosureBindingContract:
		value, err := LoadJSON[AtlasMissionDashboardClosureBinding](path)
		if err != nil {
			return "typed:mission-dashboard-closure-binding", err
		}
		return "typed:mission-dashboard-closure-binding", ValidateAtlasMissionDashboardClosureBinding(value)
	case AtlasMissionDashboardProvenanceLinksContract:
		value, err := LoadJSON[AtlasMissionDashboardProvenanceLinks](path)
		if err != nil {
			return "typed:mission-dashboard-provenance-links", err
		}
		return "typed:mission-dashboard-provenance-links", ValidateAtlasMissionDashboardProvenanceLinks(value)
	case AtlasMissionDashboardFreshnessChecksContract:
		value, err := LoadJSON[AtlasMissionDashboardFreshnessChecks](path)
		if err != nil {
			return "typed:mission-dashboard-freshness-checks", err
		}
		return "typed:mission-dashboard-freshness-checks", ValidateAtlasMissionDashboardFreshnessChecks(value)
	case AtlasMissionDashboardCompactFiltersContract:
		value, err := LoadJSON[AtlasMissionDashboardCompactFilters](path)
		if err != nil {
			return "typed:mission-dashboard-compact-filters", err
		}
		return "typed:mission-dashboard-compact-filters", ValidateAtlasMissionDashboardCompactFilters(value)
	case AtlasBoundedSignerContractFixtureContract:
		value, err := LoadJSON[AtlasBoundedSignerContractFixture](path)
		if err != nil {
			return "typed:bounded-signer-contract-fixture", err
		}
		return "typed:bounded-signer-contract-fixture", ValidateAtlasBoundedSignerContractFixture(value)
	case AtlasCanonicalContractRegistryManifestContract:
		value, err := LoadJSON[AtlasCanonicalContractRegistryManifest](path)
		if err != nil {
			return "typed:canonical-contract-registry-manifest", err
		}
		return "typed:canonical-contract-registry-manifest", ValidateAtlasCanonicalContractRegistryManifest(value)
	case AtlasContractCompatibilityInventoryContract:
		value, err := LoadJSON[AtlasContractCompatibilityInventory](path)
		if err != nil {
			return "typed:contract-compatibility-inventory", err
		}
		return "typed:contract-compatibility-inventory", ValidateAtlasContractCompatibilityInventory(value)
	case AtlasCanonicalJSONVectorsContract:
		value, err := LoadJSON[AtlasCanonicalJSONVectors](path)
		if err != nil {
			return "typed:canonical-json-vectors", err
		}
		return "typed:canonical-json-vectors", ValidateAtlasCanonicalJSONVectors(value)
	case AtlasCanonicalJSONVectorSmokeChecksContract:
		value, err := LoadJSON[AtlasCanonicalJSONVectorSmokeChecks](path)
		if err != nil {
			return "typed:canonical-json-vector-smoke-checks", err
		}
		return "typed:canonical-json-vector-smoke-checks", ValidateAtlasCanonicalJSONVectorSmokeChecks(value)
	case AtlasSentinelHostedCIWorkflowFixtureContract:
		value, err := LoadJSON[AtlasSentinelHostedCIWorkflowFixture](path)
		if err != nil {
			return "typed:sentinel-hosted-ci-workflow-fixture", err
		}
		return "typed:sentinel-hosted-ci-workflow-fixture", ValidateAtlasSentinelHostedCIWorkflowFixture(value)
	case AtlasSentinelSignalStateFixtureContract:
		value, err := LoadJSON[AtlasSentinelSignalStateFixture](path)
		if err != nil {
			return "typed:sentinel-signal-state-fixture", err
		}
		return "typed:sentinel-signal-state-fixture", ValidateAtlasSentinelSignalStateFixture(value)
	case AtlasSignedAssuranceDryRunFixtureContract:
		value, err := LoadJSON[AtlasSignedAssuranceDryRunFixture](path)
		if err != nil {
			return "typed:signed-assurance-dry-run-fixture", err
		}
		return "typed:signed-assurance-dry-run-fixture", ValidateAtlasSignedAssuranceDryRunFixture(value)
	case AtlasPromoterNoActivationBoundaryFixtureContract:
		value, err := LoadJSON[AtlasPromoterNoActivationBoundaryFixture](path)
		if err != nil {
			return "typed:promoter-no-activation-boundary-fixture", err
		}
		return "typed:promoter-no-activation-boundary-fixture", ValidateAtlasPromoterNoActivationBoundaryFixture(value)
	case AtlasWorkspaceRootPreflightFixtureContract:
		value, err := LoadJSON[AtlasWorkspaceRootPreflightFixture](path)
		if err != nil {
			return "typed:workspace-root-preflight-fixture", err
		}
		return "typed:workspace-root-preflight-fixture", ValidateAtlasWorkspaceRootPreflightFixture(value)
	case AtlasBoundedExecutionPacketFixtureContract:
		value, err := LoadJSON[AtlasBoundedExecutionPacketFixture](path)
		if err != nil {
			return "typed:bounded-execution-packet-fixture", err
		}
		return "typed:bounded-execution-packet-fixture", ValidateAtlasBoundedExecutionPacketFixture(value)
	case AtlasForgeGoalRunEvidenceFixtureContract:
		value, err := LoadJSON[AtlasForgeGoalRunEvidenceFixture](path)
		if err != nil {
			return "typed:forge-goalrun-evidence-fixture", err
		}
		return "typed:forge-goalrun-evidence-fixture", ValidateAtlasForgeGoalRunEvidenceFixture(value)
	case AtlasExecutionPacketRegressionMatrixContract:
		value, err := LoadJSON[AtlasExecutionPacketRegressionMatrix](path)
		if err != nil {
			return "typed:execution-packet-regression-matrix", err
		}
		return "typed:execution-packet-regression-matrix", ValidateAtlasExecutionPacketRegressionMatrix(value)
	case AtlasDurableStateMigrationMetadataContract:
		value, err := LoadJSON[AtlasDurableStateMigrationMetadata](path)
		if err != nil {
			return "typed:durable-state-migration-metadata", err
		}
		return "typed:durable-state-migration-metadata", ValidateAtlasDurableStateMigrationMetadata(value)
	case AtlasExactlyOnceResumeAccountingFixtureContract:
		value, err := LoadJSON[AtlasExactlyOnceResumeAccountingFixture](path)
		if err != nil {
			return "typed:exactly-once-resume-accounting-fixture", err
		}
		return "typed:exactly-once-resume-accounting-fixture", ValidateAtlasExactlyOnceResumeAccountingFixture(value)
	case AtlasReplayableStatePacketFixtureContract:
		value, err := LoadJSON[AtlasReplayableStatePacketFixture](path)
		if err != nil {
			return "typed:replayable-state-packet-fixture", err
		}
		return "typed:replayable-state-packet-fixture", ValidateAtlasReplayableStatePacketFixture(value)
	case AtlasIndexedEventQueryFixtureContract:
		value, err := LoadJSON[AtlasIndexedEventQueryFixture](path)
		if err != nil {
			return "typed:indexed-event-query-fixture", err
		}
		return "typed:indexed-event-query-fixture", ValidateAtlasIndexedEventQueryFixture(value)
	case AtlasAtomicEvidenceTransitionFixtureContract:
		value, err := LoadJSON[AtlasAtomicEvidenceTransitionFixture](path)
		if err != nil {
			return "typed:atomic-evidence-transition-fixture", err
		}
		return "typed:atomic-evidence-transition-fixture", ValidateAtlasAtomicEvidenceTransitionFixture(value)
	case AtlasLocalBackupRestoreFixtureContract:
		value, err := LoadJSON[AtlasLocalBackupRestoreFixture](path)
		if err != nil {
			return "typed:local-backup-restore-fixture", err
		}
		return "typed:local-backup-restore-fixture", ValidateAtlasLocalBackupRestoreFixture(value)
	case AtlasCommandReadbackAdapterBoundaryFixtureContract:
		value, err := LoadJSON[AtlasCommandReadbackAdapterBoundaryFixture](path)
		if err != nil {
			return "typed:command-readback-adapter-boundary-fixture", err
		}
		return "typed:command-readback-adapter-boundary-fixture", ValidateAtlasCommandReadbackAdapterBoundaryFixture(value)
	case AtlasCompactTimelineFilterFixtureContract:
		value, err := LoadJSON[AtlasCompactTimelineFilterFixture](path)
		if err != nil {
			return "typed:compact-timeline-filter-fixture", err
		}
		return "typed:compact-timeline-filter-fixture", ValidateAtlasCompactTimelineFilterFixture(value)
	case AtlasAuthorityReadinessInventoryFixtureContract:
		value, err := LoadJSON[AtlasAuthorityReadinessInventoryFixture](path)
		if err != nil {
			return "typed:authority-readiness-inventory-fixture", err
		}
		return "typed:authority-readiness-inventory-fixture", ValidateAtlasAuthorityReadinessInventoryFixture(value)
	case AtlasContentAddressedEvidenceManifestFixtureContract:
		value, err := LoadJSON[AtlasContentAddressedEvidenceManifestFixture](path)
		if err != nil {
			return "typed:content-addressed-evidence-manifest-fixture", err
		}
		return "typed:content-addressed-evidence-manifest-fixture", ValidateAtlasContentAddressedEvidenceManifestFixture(value)
	case AtlasFoundryEvidenceSizeBoundaryFixtureContract:
		value, err := LoadJSON[AtlasFoundryEvidenceSizeBoundaryFixture](path)
		if err != nil {
			return "typed:foundry-evidence-size-boundary-fixture", err
		}
		return "typed:foundry-evidence-size-boundary-fixture", ValidateAtlasFoundryEvidenceSizeBoundaryFixture(value)
	case AtlasRepeatedTaskResultLedgerFixtureContract:
		value, err := LoadJSON[AtlasRepeatedTaskResultLedgerFixture](path)
		if err != nil {
			return "typed:repeated-task-result-ledger-fixture", err
		}
		return "typed:repeated-task-result-ledger-fixture", ValidateAtlasRepeatedTaskResultLedgerFixture(value)
	case AtlasFailureInjectionFuzzingFixtureContract:
		value, err := LoadJSON[AtlasFailureInjectionFuzzingFixture](path)
		if err != nil {
			return "typed:failure-injection-fuzzing-fixture", err
		}
		return "typed:failure-injection-fuzzing-fixture", ValidateAtlasFailureInjectionFuzzingFixture(value)
	case AtlasLocalPlatformFixtureContract:
		value, err := LoadJSON[AtlasLocalPlatformFixture](path)
		if err != nil {
			return "typed:local-platform-fixture", err
		}
		return "typed:local-platform-fixture", ValidateAtlasLocalPlatformFixture(value)
	case AtlasNonAOReplayBindingFixtureContract:
		value, err := LoadJSON[AtlasNonAOReplayBindingFixture](path)
		if err != nil {
			return "typed:non-ao-replay-binding-fixture", err
		}
		return "typed:non-ao-replay-binding-fixture", ValidateAtlasNonAOReplayBindingFixture(value)
	case AtlasKillRestartReplayFixtureContract:
		value, err := LoadJSON[AtlasKillRestartReplayFixture](path)
		if err != nil {
			return "typed:kill-restart-replay-fixture", err
		}
		return "typed:kill-restart-replay-fixture", ValidateAtlasKillRestartReplayFixture(value)
	case AtlasRollbackTerminalReadbackFixtureContract:
		value, err := LoadJSON[AtlasRollbackTerminalReadbackFixture](path)
		if err != nil {
			return "typed:rollback-terminal-readback-fixture", err
		}
		return "typed:rollback-terminal-readback-fixture", ValidateAtlasRollbackTerminalReadbackFixture(value)
	case AtlasGoldenPathReadinessMatrixContract:
		value, err := LoadJSON[AtlasGoldenPathReadinessMatrix](path)
		if err != nil {
			return "typed:golden-path-readiness-matrix", err
		}
		return "typed:golden-path-readiness-matrix", ValidateAtlasGoldenPathReadinessMatrix(value)
	case AtlasMonth3FinalClosureRollupContract:
		value, err := LoadJSON[AtlasMonth3FinalClosureRollup](path)
		if err != nil {
			return "typed:month3-final-closure-rollup", err
		}
		return "typed:month3-final-closure-rollup", ValidateAtlasMonth3FinalClosureRollup(value)
	case AtlasMonth3TerminalDigestBindingContract:
		value, err := LoadJSON[AtlasMonth3TerminalDigestBinding](path)
		if err != nil {
			return "typed:month3-terminal-digest-binding", err
		}
		return "typed:month3-terminal-digest-binding", ValidateAtlasMonth3TerminalDigestBinding(value)
	case AtlasMonth3NonAODryRunReplayBindingContract:
		value, err := LoadJSON[AtlasMonth3NonAODryRunReplayBinding](path)
		if err != nil {
			return "typed:month3-non-ao-dry-run-replay", err
		}
		return "typed:month3-non-ao-dry-run-replay", ValidateAtlasMonth3NonAODryRunReplayBinding(value)
	case AtlasMonth3RealRunAcceptanceCriteriaContract:
		value, err := LoadJSON[AtlasMonth3RealRunAcceptanceCriteria](path)
		if err != nil {
			return "typed:month3-real-run-acceptance-criteria", err
		}
		return "typed:month3-real-run-acceptance-criteria", ValidateAtlasMonth3RealRunAcceptanceCriteria(value)
	case AtlasMonth3ControlPlaneObserverBindingContract:
		value, err := LoadJSON[AtlasMonth3ControlPlaneObserverBinding](path)
		if err != nil {
			return "typed:month3-control-plane-observer-binding", err
		}
		return "typed:month3-control-plane-observer-binding", ValidateAtlasMonth3ControlPlaneObserverBinding(value)
	case AtlasMonth3SchemaOwnerRegistryProposalContract:
		value, err := LoadJSON[AtlasMonth3SchemaOwnerRegistryProposal](path)
		if err != nil {
			return "typed:month3-schema-owner-registry-proposal", err
		}
		return "typed:month3-schema-owner-registry-proposal", ValidateAtlasMonth3SchemaOwnerRegistryProposal(value)
	case AtlasMonth3EvidenceExternalizationPlanContract:
		value, err := LoadJSON[AtlasMonth3EvidenceExternalizationPlan](path)
		if err != nil {
			return "typed:month3-evidence-externalization-plan", err
		}
		return "typed:month3-evidence-externalization-plan", ValidateAtlasMonth3EvidenceExternalizationPlan(value)
	case AtlasMonth3CrossRepoCIMatrixContract:
		value, err := LoadJSON[AtlasMonth3CrossRepoCIMatrix](path)
		if err != nil {
			return "typed:month3-cross-repo-ci-matrix", err
		}
		return "typed:month3-cross-repo-ci-matrix", ValidateAtlasMonth3CrossRepoCIMatrix(value)
	case AtlasMonth3OperatorDashboardReadbackContract:
		value, err := LoadJSON[AtlasMonth3OperatorDashboardReadback](path)
		if err != nil {
			return "typed:month3-operator-dashboard-readback", err
		}
		return "typed:month3-operator-dashboard-readback", ValidateAtlasMonth3OperatorDashboardReadback(value)
	case AtlasMonth3RestartResumeSoakContract:
		value, err := LoadJSON[AtlasMonth3RestartResumeSoak](path)
		if err != nil {
			return "typed:month3-restart-resume-soak", err
		}
		return "typed:month3-restart-resume-soak", ValidateAtlasMonth3RestartResumeSoak(value)
	case AtlasBlueprintCanonicalPreservationFixtureContract:
		value, err := LoadJSON[AtlasBlueprintCanonicalPreservationFixture](path)
		if err != nil {
			return "typed:blueprint-canonical-preservation-fixture", err
		}
		return "typed:blueprint-canonical-preservation-fixture", ValidateAtlasBlueprintCanonicalPreservationFixture(value)
	case AtlasFoundryCanonicalImportFixtureContract:
		value, err := LoadJSON[AtlasFoundryCanonicalImportFixture](path)
		if err != nil {
			return "typed:foundry-canonical-import-fixture", err
		}
		return "typed:foundry-canonical-import-fixture", ValidateAtlasFoundryCanonicalImportFixture(value)
	case AtlasCommandCovenantFieldParityFixtureContract:
		value, err := LoadJSON[AtlasCommandCovenantFieldParityFixture](path)
		if err != nil {
			return "typed:command-covenant-field-parity-fixture", err
		}
		return "typed:command-covenant-field-parity-fixture", ValidateAtlasCommandCovenantFieldParityFixture(value)
	case AOMissionRefactoringRecommendationsContract:
		validator, _ := recommendationControlPlaneTypedValidator(AOMissionRefactoringRecommendationsContract)
		value, err := LoadJSON[AOMissionRefactoringRecommendations](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasNextWaveRefactoringRecommendations(value, value.MinimumTasks)
	case AtlasRecommendationNextTrackDecisionContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationNextTrackDecisionContract)
		value, err := LoadJSON[AtlasRecommendationNextTrackDecision](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationNextTrackDecision(value)
	case AtlasConsumedRecommendationLedgerContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasConsumedRecommendationLedgerContract)
		value, err := LoadJSON[AtlasConsumedRecommendationLedger](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasConsumedRecommendationLedger(value)
	case AtlasRecommendationTrackRegistryContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationTrackRegistryContract)
		value, err := LoadJSON[AtlasRecommendationTrackRegistry](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationTrackRegistry(value)
	case AtlasRecommendationCommandRunLedgerContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationCommandRunLedgerContract)
		value, err := LoadJSON[AtlasRecommendationCommandRunLedger](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationCommandRunLedger(value)
	case AtlasRecommendationCommandRunLedgerRollupContract:
		value, err := LoadJSON[AtlasRecommendationCommandRunLedgerRollup](path)
		if err != nil {
			return "typed:recommendation-command-run-ledger-rollup", err
		}
		return "typed:recommendation-command-run-ledger-rollup", ValidateAtlasRecommendationCommandRunLedgerRollup(value)
	case AtlasRecommendationRunLedgerCoverageCheckContract:
		value, err := LoadJSON[AtlasRecommendationRunLedgerCoverageCheck](path)
		if err != nil {
			return "typed:recommendation-run-ledger-coverage-check", err
		}
		return "typed:recommendation-run-ledger-coverage-check", ValidateAtlasRecommendationRunLedgerCoverageCheck(value)
	case AtlasRecommendationFinalResponseGatesContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationFinalResponseGatesContract)
		value, err := LoadJSON[AtlasRecommendationFinalResponseGates](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationFinalResponseGates(value)
	case AtlasRecommendationEvidenceValidationReportContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationEvidenceValidationReportContract)
		value, err := LoadJSON[AtlasRecommendationEvidenceValidationReport](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationEvidenceValidationReport(value)
	case AtlasRecommendationEvidenceSchemaRegistryContract:
		value, err := LoadJSON[AtlasRecommendationEvidenceSchemaRegistry](path)
		if err != nil {
			return "typed:recommendation-evidence-schema-registry", err
		}
		return "typed:recommendation-evidence-schema-registry", ValidateAtlasRecommendationEvidenceSchemaRegistry(value)
	case AtlasRecommendationEvidenceSchemaRegistryCoverageContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasRecommendationEvidenceSchemaRegistryCoverageContract)
		value, err := LoadJSON[AtlasRecommendationEvidenceSchemaRegistryCoverage](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasRecommendationEvidenceSchemaRegistryCoverage(value)
	case AtlasSchemaHealthRepairPromptContract:
		validator, _ := recommendationControlPlaneTypedValidator(AtlasSchemaHealthRepairPromptContract)
		value, err := LoadJSON[AtlasSchemaHealthRepairPrompt](path)
		if err != nil {
			return validator, err
		}
		return validator, ValidateAtlasSchemaHealthRepairPrompt(value)
	case "ao.atlas.recommendation-checkpoint-readback.v0.1":
		value, err := LoadJSON[AtlasRecommendationCheckpointReadback](path)
		if err != nil {
			return "typed:recommendation-checkpoint-readback", err
		}
		return "typed:recommendation-checkpoint-readback", ValidateAtlasRecommendationCheckpointReadback(value)
	case "ao.atlas.long-recommendation-wave-execution.v0.3":
		value, err := LoadJSON[AtlasRecommendationExecutionReadback](path)
		if err != nil {
			return "typed:recommendation-execution-readback", err
		}
		readback, err := LoadJSON[AtlasRecommendationReadback](pairedRecommendationReadbackPath(path))
		if err != nil {
			return "typed:recommendation-execution-readback", err
		}
		return "typed:recommendation-execution-readback", ValidateAtlasRecommendationExecutionReadback(value, readback)
	default:
		return "generic:schema-marker", nil
	}
}

func pairedRecommendationReadbackPath(executionPath string) string {
	dir := filepath.Dir(executionPath)
	filename := filepath.Base(executionPath)
	if strings.HasSuffix(filename, "execution-readback.json") {
		candidate := filepath.Join(dir, strings.TrimSuffix(filename, "execution-readback.json")+"recommendation-readback.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join(dir, "recommendation-readback-after.json")
}

func requiredRecommendationEvidenceFilenames() []string {
	return []string{
		"node_gate.json",
		"candidate_record.json",
		"rollback_record.json",
		"tests.json",
		"verification.json",
		"sentinel_public_safety.json",
		"promoter_no_promotion.json",
		"command_readback.json",
		"run-link.json",
		"recommendation-readback-after.json",
		"checkpoint-readback-after.json",
	}
}

func sortedMapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
