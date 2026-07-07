package atlas

import (
	"fmt"
	"sort"
	"strings"
)

func BuildAtlasSchemaValidatorDriftEvidence(sourceReportPath, targetReportPath string) (AtlasSchemaValidatorDriftEvidence, error) {
	source, err := LoadJSON[AtlasRecommendationEvidenceValidationReport](sourceReportPath)
	if err != nil {
		return AtlasSchemaValidatorDriftEvidence{}, err
	}
	target, err := LoadJSON[AtlasRecommendationEvidenceValidationReport](targetReportPath)
	if err != nil {
		return AtlasSchemaValidatorDriftEvidence{}, err
	}
	if source.Status != "passed" {
		return AtlasSchemaValidatorDriftEvidence{}, fmt.Errorf("source report status must be passed")
	}
	if target.Status != "passed" {
		return AtlasSchemaValidatorDriftEvidence{}, fmt.Errorf("target report status must be passed")
	}
	schemaDeltas, addedSchemas, lostSchemas := countDeltas(source.SchemaCounts, target.SchemaCounts)
	validatorDeltas, addedValidators, lostValidators := countDeltas(source.Validators, target.Validators)
	unexpectedLoss := len(lostSchemas) != 0 || len(lostValidators) != 0
	status := "recorded_no_unexpected_loss"
	if unexpectedLoss {
		status = "blocked_unexpected_loss"
	}
	drift := AtlasSchemaValidatorDriftEvidence{
		Schema:                 AtlasSchemaValidatorDriftContract,
		Status:                 status,
		SourceReportPath:       publicArtifactRef(sourceReportPath),
		TargetReportPath:       publicArtifactRef(targetReportPath),
		SourceReportDigest:     digestValue(source),
		TargetReportDigest:     digestValue(target),
		SourceNodeCount:        source.NodeCount,
		TargetNodeCount:        target.NodeCount,
		JSONFileDelta:          target.JSONFileCount - source.JSONFileCount,
		TypedValidatorDelta:    target.TypedValidatorFiles - source.TypedValidatorFiles,
		GenericSchemaDelta:     target.GenericSchemaFiles - source.GenericSchemaFiles,
		SchemaCountDeltas:      schemaDeltas,
		ValidatorCountDeltas:   validatorDeltas,
		AddedSchemas:           addedSchemas,
		LostSchemas:            lostSchemas,
		AddedValidators:        addedValidators,
		LostValidators:         lostValidators,
		UnexpectedLossDetected: unexpectedLoss,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasSchemaValidatorDriftEvidence(drift); err != nil {
		return AtlasSchemaValidatorDriftEvidence{}, err
	}
	return drift, nil
}

func ValidateAtlasSchemaValidatorDriftEvidence(drift AtlasSchemaValidatorDriftEvidence) error {
	var errs []string
	requireContract(&errs, "schema_validator_drift", drift.Schema, AtlasSchemaValidatorDriftContract)
	if !oneOf(drift.Status, "recorded_no_unexpected_loss", "blocked_unexpected_loss") {
		errs = append(errs, "status must be recorded_no_unexpected_loss or blocked_unexpected_loss")
	}
	requireField(&errs, "source_report_path", drift.SourceReportPath)
	checkPublicPath(&errs, "source_report_path", drift.SourceReportPath, true)
	requireField(&errs, "target_report_path", drift.TargetReportPath)
	checkPublicPath(&errs, "target_report_path", drift.TargetReportPath, true)
	if !digestPattern.MatchString(drift.SourceReportDigest) {
		errs = append(errs, "source_report_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(drift.TargetReportDigest) {
		errs = append(errs, "target_report_digest must be sha256 digest")
	}
	if drift.SourceNodeCount <= 0 {
		errs = append(errs, "source_node_count must be greater than zero")
	}
	if drift.TargetNodeCount < drift.SourceNodeCount {
		errs = append(errs, "target_node_count must be greater than or equal to source_node_count")
	}
	if drift.SchemaCountDeltas == nil {
		errs = append(errs, "schema_count_deltas is required")
	}
	if drift.ValidatorCountDeltas == nil {
		errs = append(errs, "validator_count_deltas is required")
	}
	if drift.UnexpectedLossDetected != (len(drift.LostSchemas) != 0 || len(drift.LostValidators) != 0) {
		errs = append(errs, "unexpected_loss_detected must match lost schema or validator entries")
	}
	if drift.UnexpectedLossDetected && drift.Status != "blocked_unexpected_loss" {
		errs = append(errs, "unexpected loss requires blocked_unexpected_loss status")
	}
	if !drift.UnexpectedLossDetected && drift.Status != "recorded_no_unexpected_loss" {
		errs = append(errs, "no unexpected loss requires recorded_no_unexpected_loss status")
	}
	for _, key := range append([]string{}, drift.AddedSchemas...) {
		if strings.TrimSpace(key) == "" || drift.SchemaCountDeltas[key] <= 0 {
			errs = append(errs, "added_schemas must reference positive schema deltas")
		}
	}
	for _, key := range append([]string{}, drift.LostSchemas...) {
		if strings.TrimSpace(key) == "" || drift.SchemaCountDeltas[key] >= 0 {
			errs = append(errs, "lost_schemas must reference negative schema deltas")
		}
	}
	for _, key := range append([]string{}, drift.AddedValidators...) {
		if strings.TrimSpace(key) == "" || drift.ValidatorCountDeltas[key] <= 0 {
			errs = append(errs, "added_validators must reference positive validator deltas")
		}
	}
	for _, key := range append([]string{}, drift.LostValidators...) {
		if strings.TrimSpace(key) == "" || drift.ValidatorCountDeltas[key] >= 0 {
			errs = append(errs, "lost_validators must reference negative validator deltas")
		}
	}
	if drift.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if drift.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if drift.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if drift.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !drift.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func countDeltas(source, target map[string]int) (map[string]int, []string, []string) {
	deltas := map[string]int{}
	added := []string{}
	lost := []string{}
	keys := map[string]bool{}
	for key := range source {
		keys[key] = true
	}
	for key := range target {
		keys[key] = true
	}
	for key := range keys {
		delta := target[key] - source[key]
		deltas[key] = delta
		if source[key] == 0 && delta > 0 {
			added = append(added, key)
		}
		if delta < 0 {
			lost = append(lost, key)
		}
	}
	sort.Strings(added)
	sort.Strings(lost)
	return deltas, added, lost
}
