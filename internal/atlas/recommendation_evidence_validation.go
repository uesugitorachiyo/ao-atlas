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
	Schema                   string                                       `json:"schema"`
	Status                   string                                       `json:"status"`
	EvidenceRoot             string                                       `json:"evidence_root"`
	NodeRoot                 string                                       `json:"node_root"`
	NodeCount                int                                          `json:"node_count"`
	JSONFileCount            int                                          `json:"json_file_count"`
	ValidatedJSONFiles       int                                          `json:"validated_json_files"`
	SchemaBoundFiles         int                                          `json:"schema_bound_files"`
	TypedValidatorFiles      int                                          `json:"typed_validator_files"`
	GenericSchemaFiles       int                                          `json:"generic_schema_files"`
	MissingSchemaFiles       []string                                     `json:"missing_schema_files"`
	FailedFiles              []string                                     `json:"failed_files"`
	MissingRequiredFiles     []string                                     `json:"missing_required_files"`
	RequiredFilenames        []string                                     `json:"required_filenames"`
	RequiredFilenamesCovered bool                                         `json:"required_filenames_covered"`
	SchemaCounts             map[string]int                               `json:"schema_counts"`
	Validators               map[string]int                               `json:"validators"`
	Entries                  []AtlasRecommendationEvidenceValidationEntry `json:"entries"`
}

type AtlasRecommendationEvidenceValidationEntry struct {
	Path      string `json:"path"`
	NodeID    string `json:"node_id"`
	Filename  string `json:"filename"`
	Schema    string `json:"schema"`
	Validator string `json:"validator"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

func BuildAtlasRecommendationEvidenceValidationReport(evidenceRoot string) (AtlasRecommendationEvidenceValidationReport, error) {
	evidenceRoot = strings.TrimSpace(evidenceRoot)
	report := AtlasRecommendationEvidenceValidationReport{
		Schema:                   AtlasRecommendationEvidenceValidationReportContract,
		Status:                   "passed",
		EvidenceRoot:             filepath.ToSlash(evidenceRoot),
		NodeRoot:                 filepath.ToSlash(filepath.Join(evidenceRoot, "nodes")),
		SchemaCounts:             map[string]int{},
		Validators:               map[string]int{},
		RequiredFilenames:        requiredRecommendationEvidenceFilenames(),
		RequiredFilenamesCovered: true,
		MissingSchemaFiles:       []string{},
		FailedFiles:              []string{},
		MissingRequiredFiles:     []string{},
		Entries:                  []AtlasRecommendationEvidenceValidationEntry{},
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
		entry := validateRecommendationEvidenceJSONFile(evidenceRoot, nodeRoot, path)
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
		}
		if entry.Schema == "" {
			report.MissingSchemaFiles = append(report.MissingSchemaFiles, entry.Path)
		} else {
			report.SchemaBoundFiles++
			report.SchemaCounts[entry.Schema]++
		}
		if strings.HasPrefix(entry.Validator, "typed:") {
			report.TypedValidatorFiles++
		} else if entry.Validator == "generic:schema-marker" {
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

func validateRecommendationEvidenceJSONFile(evidenceRoot, nodeRoot, path string) AtlasRecommendationEvidenceValidationEntry {
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
	entry.Validator = validator
	if err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
		return entry
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
	case AtlasRecommendationReadbackContract:
		value, err := LoadJSON[AtlasRecommendationReadback](path)
		if err != nil {
			return "typed:recommendation-readback", err
		}
		return "typed:recommendation-readback", ValidateAtlasRecommendationReadback(value)
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
	case AtlasFailedCheckReplayFixtureContract:
		value, err := LoadJSON[AtlasFailedCheckReplayFixture](path)
		if err != nil {
			return "typed:failed-check-replay-fixture", err
		}
		return "typed:failed-check-replay-fixture", ValidateAtlasFailedCheckReplayFixture(value)
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
