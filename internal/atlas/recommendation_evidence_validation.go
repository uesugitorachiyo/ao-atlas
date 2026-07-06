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
