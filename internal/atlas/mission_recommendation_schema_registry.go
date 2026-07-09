package atlas

import "fmt"

func DefaultAtlasRecommendationEvidenceSchemaRegistry() (AtlasRecommendationEvidenceSchemaRegistry, error) {
	entries := defaultAtlasRecommendationEvidenceSchemaRegistryEntries()
	registry := AtlasRecommendationEvidenceSchemaRegistry{
		Schema:                         AtlasRecommendationEvidenceSchemaRegistryContract,
		Status:                         "ready",
		RegistryPurpose:                "recommendation_control_plane_typed_artifact_coverage",
		Schemas:                        entries,
		SchemaCount:                    len(entries),
		TypedValidatorCoverageComplete: true,
		NoPromotionRequested:           true,
		PromotionGranted:               false,
		ClaimsAuthorityAdvance:         false,
		RSIRemainsDenied:               true,
		SafeToExecute:                  false,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		MutatesRepositories:            false,
	}
	if err := ValidateAtlasRecommendationEvidenceSchemaRegistry(registry); err != nil {
		return AtlasRecommendationEvidenceSchemaRegistry{}, err
	}
	return registry, nil
}

func defaultAtlasRecommendationEvidenceSchemaRegistryEntries() []AtlasRecommendationEvidenceSchemaRegistryEntry {
	return []AtlasRecommendationEvidenceSchemaRegistryEntry{
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationNextTrackDecisionContract, "recommendation-next-track-decision", "next-track", "typed:recommendation-next-track-decision"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasConsumedRecommendationLedgerContract, "consumed-recommendation-ledger", "consumed-ledger", "typed:consumed-recommendation-ledger"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationTrackRegistryContract, "recommendation-track-registry", "track-registry", "typed:recommendation-track-registry"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationCommandRunLedgerContract, "recommendation-command-run-ledger", "run-ledger", "typed:recommendation-command-run-ledger"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationEvidenceValidationReportContract, "recommendation-evidence-validation-report", "validate-evidence", "typed:recommendation-evidence-validation-report"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationFinalResponseGatesContract, "recommendation-final-response-gates", "final-response-gates", "typed:recommendation-final-response-gates"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasRecommendationEvidenceSchemaRegistryCoverageContract, "recommendation-evidence-schema-registry-coverage", "schema-registry-coverage", "typed:recommendation-evidence-schema-registry-coverage"),
		newAtlasRecommendationEvidenceSchemaRegistryEntry(AtlasSchemaHealthRepairPromptContract, "schema-health-repair-prompt", "schema-health-repair-prompt", "typed:schema-health-repair-prompt"),
	}
}

func newAtlasRecommendationEvidenceSchemaRegistryEntry(schema, artifact, command, typedValidator string) AtlasRecommendationEvidenceSchemaRegistryEntry {
	return AtlasRecommendationEvidenceSchemaRegistryEntry{
		Schema:         schema,
		Artifact:       artifact,
		Command:        command,
		TypedValidator: typedValidator,
		StatusField:    "status",
		SafetyClass:    "planning_readback_no_execution",
		PlanningOnly:   true,
	}
}

func ValidateAtlasRecommendationEvidenceSchemaRegistry(registry AtlasRecommendationEvidenceSchemaRegistry) error {
	var errs []string
	requireContract(&errs, "recommendation_evidence_schema_registry", registry.Schema, AtlasRecommendationEvidenceSchemaRegistryContract)
	if registry.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if registry.RegistryPurpose != "recommendation_control_plane_typed_artifact_coverage" {
		errs = append(errs, "registry_purpose must be recommendation_control_plane_typed_artifact_coverage")
	}
	expected := defaultAtlasRecommendationEvidenceSchemaRegistryEntries()
	if registry.SchemaCount != len(registry.Schemas) {
		errs = append(errs, "schema_count must equal schemas length")
	}
	if registry.SchemaCount != len(expected) {
		errs = append(errs, fmt.Sprintf("schema_count must be %d", len(expected)))
	}
	if len(registry.Schemas) != len(expected) {
		errs = append(errs, fmt.Sprintf("schemas must include %d entries", len(expected)))
	}
	for index, entry := range registry.Schemas {
		requireField(&errs, "schema_registry_entry.schema", entry.Schema)
		requireField(&errs, "schema_registry_entry.artifact", entry.Artifact)
		requireField(&errs, "schema_registry_entry.command", entry.Command)
		requireField(&errs, "schema_registry_entry.typed_validator", entry.TypedValidator)
		requireField(&errs, "schema_registry_entry.status_field", entry.StatusField)
		requireField(&errs, "schema_registry_entry.safety_class", entry.SafetyClass)
		if !entry.PlanningOnly {
			errs = append(errs, fmt.Sprintf("schema registry entry %s planning_only must be true", entry.Schema))
		}
		if index < len(expected) && entry != expected[index] {
			errs = append(errs, fmt.Sprintf("schema registry entry %d must describe %s", index, expected[index].Schema))
		}
	}
	if !registry.TypedValidatorCoverageComplete {
		errs = append(errs, "typed_validator_coverage_complete must be true")
	}
	if !registry.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if registry.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if registry.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !registry.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if registry.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if registry.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if registry.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if registry.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if registry.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
