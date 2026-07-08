package atlas

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

var allowedRecommendationSchemaRegistryCoverageFailureReasons = []string{
	"validation_report_failed",
	"missing_registry_schemas",
	"missing_registry_validators",
	"stale_registry_entries",
}

func BuildAtlasRecommendationEvidenceSchemaRegistryCoverage(registryPath, validationReportPath string) (AtlasRecommendationEvidenceSchemaRegistryCoverage, error) {
	registryPath = filepath.ToSlash(strings.TrimSpace(registryPath))
	validationReportPath = filepath.ToSlash(strings.TrimSpace(validationReportPath))
	if registryPath == "" {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, fmt.Errorf("registry path is required")
	}
	if validationReportPath == "" {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, fmt.Errorf("validation report path is required")
	}
	registry, err := LoadJSON[AtlasRecommendationEvidenceSchemaRegistry](registryPath)
	if err != nil {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, err
	}
	if err := ValidateAtlasRecommendationEvidenceSchemaRegistry(registry); err != nil {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, err
	}
	report, err := LoadJSON[AtlasRecommendationEvidenceValidationReport](validationReportPath)
	if err != nil {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, err
	}
	if report.Schema != AtlasRecommendationEvidenceValidationReportContract {
		return AtlasRecommendationEvidenceSchemaRegistryCoverage{}, fmt.Errorf("validation report schema must be %s", AtlasRecommendationEvidenceValidationReportContract)
	}

	missingSchemas := []string{}
	missingValidators := []string{}
	staleRegistryEntries := []string{}
	coveredSchemas := 0
	coveredValidators := 0
	for _, entry := range registry.Schemas {
		schemaCovered := report.SchemaCounts[entry.Schema] > 0
		validatorCovered := report.Validators[entry.TypedValidator] > 0
		if schemaCovered {
			coveredSchemas++
		} else {
			missingSchemas = append(missingSchemas, entry.Schema)
		}
		if validatorCovered {
			coveredValidators++
		} else {
			missingValidators = append(missingValidators, entry.TypedValidator)
		}
		if !schemaCovered && !validatorCovered {
			staleRegistryEntries = append(staleRegistryEntries, entry.Schema)
		}
	}
	sort.Strings(missingSchemas)
	sort.Strings(missingValidators)
	sort.Strings(staleRegistryEntries)
	failureReasons := []string{}
	if report.Status != "passed" {
		failureReasons = append(failureReasons, "validation_report_failed")
	}
	if len(missingSchemas) != 0 {
		failureReasons = append(failureReasons, "missing_registry_schemas")
	}
	if len(missingValidators) != 0 {
		failureReasons = append(failureReasons, "missing_registry_validators")
	}
	if len(staleRegistryEntries) != 0 {
		failureReasons = append(failureReasons, "stale_registry_entries")
	}

	coverage := AtlasRecommendationEvidenceSchemaRegistryCoverage{
		Schema:                       AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
		Status:                       "passed",
		RegistryPath:                 registryPath,
		ValidationReportPath:         validationReportPath,
		ValidationReportStatus:       report.Status,
		RegistrySchemaCount:          len(registry.Schemas),
		CoveredSchemaCount:           coveredSchemas,
		MissingSchemas:               missingSchemas,
		RegistryValidatorCount:       len(registry.Schemas),
		CoveredValidatorCount:        coveredValidators,
		MissingValidators:            missingValidators,
		StaleRegistryEntryCount:      len(staleRegistryEntries),
		StaleRegistryEntries:         staleRegistryEntries,
		AllRegistryEntriesFresh:      len(staleRegistryEntries) == 0,
		FailureReasons:               failureReasons,
		AllRegistrySchemasCovered:    len(missingSchemas) == 0,
		AllRegistryValidatorsCovered: len(missingValidators) == 0,
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
	if report.Status != "passed" || len(missingSchemas) != 0 || len(missingValidators) != 0 || len(staleRegistryEntries) != 0 {
		coverage.Status = "failed"
	}
	if err := ValidateAtlasRecommendationEvidenceSchemaRegistryCoverage(coverage); err != nil {
		return coverage, err
	}
	if coverage.Status != "passed" {
		return coverage, fmt.Errorf("recommendation schema registry coverage failed")
	}
	return coverage, nil
}

func ValidateAtlasRecommendationEvidenceSchemaRegistryCoverage(coverage AtlasRecommendationEvidenceSchemaRegistryCoverage) error {
	var errs []string
	requireContract(&errs, "recommendation_evidence_schema_registry_coverage", coverage.Schema, AtlasRecommendationEvidenceSchemaRegistryCoverageContract)
	if !oneOf(coverage.Status, "passed", "failed") {
		errs = append(errs, "status must be passed or failed")
	}
	requireField(&errs, "registry_path", coverage.RegistryPath)
	requireField(&errs, "validation_report_path", coverage.ValidationReportPath)
	if !oneOf(coverage.ValidationReportStatus, "passed", "failed") {
		errs = append(errs, "validation_report_status must be passed or failed")
	}
	if coverage.RegistrySchemaCount <= 0 {
		errs = append(errs, "registry_schema_count must be positive")
	}
	if coverage.CoveredSchemaCount < 0 || coverage.CoveredSchemaCount > coverage.RegistrySchemaCount {
		errs = append(errs, "covered_schema_count must be within registry_schema_count")
	}
	if coverage.RegistryValidatorCount != coverage.RegistrySchemaCount {
		errs = append(errs, "registry_validator_count must equal registry_schema_count")
	}
	if coverage.CoveredValidatorCount < 0 || coverage.CoveredValidatorCount > coverage.RegistryValidatorCount {
		errs = append(errs, "covered_validator_count must be within registry_validator_count")
	}
	if coverage.AllRegistrySchemasCovered != (len(coverage.MissingSchemas) == 0) {
		errs = append(errs, "all_registry_schemas_covered must match missing_schemas")
	}
	if coverage.AllRegistryValidatorsCovered != (len(coverage.MissingValidators) == 0) {
		errs = append(errs, "all_registry_validators_covered must match missing_validators")
	}
	if coverage.StaleRegistryEntryCount != len(coverage.StaleRegistryEntries) {
		errs = append(errs, "stale_registry_entry_count must match stale_registry_entries")
	}
	if coverage.AllRegistryEntriesFresh != (len(coverage.StaleRegistryEntries) == 0) {
		errs = append(errs, "all_registry_entries_fresh must match stale_registry_entries")
	}
	checkPublicStrings(&errs, "stale_registry_entries", coverage.StaleRegistryEntries, true)
	if coverage.Status == "passed" && (!coverage.AllRegistrySchemasCovered || !coverage.AllRegistryValidatorsCovered || !coverage.AllRegistryEntriesFresh || coverage.ValidationReportStatus != "passed") {
		errs = append(errs, "passed status requires passed report, full registry coverage, and fresh registry entries")
	}
	if coverage.Status == "failed" && len(coverage.FailureReasons) == 0 {
		errs = append(errs, "failed status requires failure_reasons")
	}
	for _, reason := range coverage.FailureReasons {
		if !oneOf(reason, allowedRecommendationSchemaRegistryCoverageFailureReasons...) {
			errs = append(errs, "failure_reasons contains invalid reason "+reason+"; allowed: "+strings.Join(allowedRecommendationSchemaRegistryCoverageFailureReasons, ", "))
		}
	}
	if !coverage.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if coverage.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if coverage.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !coverage.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if coverage.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if coverage.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if coverage.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if coverage.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if coverage.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
