package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildAtlasRecommendationCommandRunLedger(command, artifactPath string) (AtlasRecommendationCommandRunLedger, error) {
	command = strings.TrimSpace(command)
	artifactPath = filepath.ToSlash(strings.TrimSpace(artifactPath))
	if command == "" {
		return AtlasRecommendationCommandRunLedger{}, fmt.Errorf("command is required")
	}
	if artifactPath == "" {
		return AtlasRecommendationCommandRunLedger{}, fmt.Errorf("artifact path is required")
	}
	if !oneOf(command, "next-track", "consumed-ledger", "track-registry", "final-response-gates", "schema-registry", "schema-registry-coverage") {
		return AtlasRecommendationCommandRunLedger{}, fmt.Errorf("command must be next-track, consumed-ledger, track-registry, final-response-gates, schema-registry, or schema-registry-coverage")
	}

	header, err := LoadJSON[struct {
		Schema string `json:"schema"`
		Status string `json:"status"`
	}](artifactPath)
	if err != nil {
		return AtlasRecommendationCommandRunLedger{}, err
	}
	if strings.TrimSpace(header.Schema) == "" {
		return AtlasRecommendationCommandRunLedger{}, fmt.Errorf("artifact schema is required")
	}
	validator, err := validateRecommendationEvidenceTypedFile(artifactPath, header.Schema)
	if err != nil {
		return AtlasRecommendationCommandRunLedger{}, err
	}
	digest, err := digestFile(artifactPath)
	if err != nil {
		return AtlasRecommendationCommandRunLedger{}, err
	}

	ledger := AtlasRecommendationCommandRunLedger{
		Schema:                 AtlasRecommendationCommandRunLedgerContract,
		Status:                 "recorded",
		Command:                command,
		ArtifactPath:           artifactPath,
		ArtifactDigest:         digest,
		ArtifactSchema:         header.Schema,
		TypedValidator:         validator,
		OutputStatus:           header.Status,
		RecordsInvocation:      true,
		NoPromotionRequested:   true,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
	}
	if err := ValidateAtlasRecommendationCommandRunLedger(ledger); err != nil {
		return AtlasRecommendationCommandRunLedger{}, err
	}
	return ledger, nil
}

func ValidateAtlasRecommendationCommandRunLedger(ledger AtlasRecommendationCommandRunLedger) error {
	var errs []string
	requireContract(&errs, "recommendation_command_run_ledger", ledger.Schema, AtlasRecommendationCommandRunLedgerContract)
	if ledger.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	if !oneOf(ledger.Command, "next-track", "consumed-ledger", "track-registry", "final-response-gates", "schema-registry", "schema-registry-coverage") {
		errs = append(errs, "command must be next-track, consumed-ledger, track-registry, final-response-gates, schema-registry, or schema-registry-coverage")
	}
	requireField(&errs, "artifact_path", ledger.ArtifactPath)
	if !digestPattern.MatchString(ledger.ArtifactDigest) {
		errs = append(errs, "artifact_digest must be sha256 digest")
	}
	if !oneOf(ledger.ArtifactSchema,
		AtlasRecommendationNextTrackDecisionContract,
		AtlasConsumedRecommendationLedgerContract,
		AtlasRecommendationTrackRegistryContract,
		AtlasRecommendationFinalResponseGatesContract,
		AtlasRecommendationEvidenceSchemaRegistryContract,
		AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
	) {
		errs = append(errs, "artifact_schema is not an allowed recommendation command output")
	}
	requireField(&errs, "typed_validator", ledger.TypedValidator)
	if !strings.HasPrefix(ledger.TypedValidator, "typed:") {
		errs = append(errs, "typed_validator must name a typed validator")
	}
	requireField(&errs, "output_status", ledger.OutputStatus)
	if !ledger.RecordsInvocation {
		errs = append(errs, "records_invocation must be true")
	}
	if !ledger.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if ledger.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if ledger.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !ledger.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if ledger.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if ledger.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if ledger.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if ledger.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if ledger.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
