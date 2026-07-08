package atlas

import "fmt"

func DefaultAtlasRecommendationEvidenceSchemaRegistry() (AtlasRecommendationEvidenceSchemaRegistry, error) {
	registry := AtlasRecommendationEvidenceSchemaRegistry{
		Schema:          AtlasRecommendationEvidenceSchemaRegistryContract,
		Status:          "ready",
		RegistryPurpose: "recommendation_control_plane_typed_artifact_coverage",
		Schemas: []AtlasRecommendationEvidenceSchemaRegistryEntry{
			{
				Schema:         AtlasRecommendationNextTrackDecisionContract,
				Artifact:       "recommendation-next-track-decision",
				Command:        "next-track",
				TypedValidator: "typed:recommendation-next-track-decision",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasConsumedRecommendationLedgerContract,
				Artifact:       "consumed-recommendation-ledger",
				Command:        "consumed-ledger",
				TypedValidator: "typed:consumed-recommendation-ledger",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasRecommendationTrackRegistryContract,
				Artifact:       "recommendation-track-registry",
				Command:        "track-registry",
				TypedValidator: "typed:recommendation-track-registry",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasRecommendationCommandRunLedgerContract,
				Artifact:       "recommendation-command-run-ledger",
				Command:        "run-ledger",
				TypedValidator: "typed:recommendation-command-run-ledger",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasRecommendationEvidenceValidationReportContract,
				Artifact:       "recommendation-evidence-validation-report",
				Command:        "validate-evidence",
				TypedValidator: "typed:recommendation-evidence-validation-report",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasRecommendationFinalResponseGatesContract,
				Artifact:       "recommendation-final-response-gates",
				Command:        "final-response-gates",
				TypedValidator: "typed:recommendation-final-response-gates",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
			{
				Schema:         AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
				Artifact:       "recommendation-evidence-schema-registry-coverage",
				Command:        "schema-registry-coverage",
				TypedValidator: "typed:recommendation-evidence-schema-registry-coverage",
				StatusField:    "status",
				SafetyClass:    "planning_readback_no_execution",
				PlanningOnly:   true,
			},
		},
		SchemaCount:                    7,
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

func ValidateAtlasRecommendationEvidenceSchemaRegistry(registry AtlasRecommendationEvidenceSchemaRegistry) error {
	var errs []string
	requireContract(&errs, "recommendation_evidence_schema_registry", registry.Schema, AtlasRecommendationEvidenceSchemaRegistryContract)
	if registry.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if registry.RegistryPurpose != "recommendation_control_plane_typed_artifact_coverage" {
		errs = append(errs, "registry_purpose must be recommendation_control_plane_typed_artifact_coverage")
	}
	expected := []AtlasRecommendationEvidenceSchemaRegistryEntry{
		{Schema: AtlasRecommendationNextTrackDecisionContract, Artifact: "recommendation-next-track-decision", Command: "next-track", TypedValidator: "typed:recommendation-next-track-decision", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasConsumedRecommendationLedgerContract, Artifact: "consumed-recommendation-ledger", Command: "consumed-ledger", TypedValidator: "typed:consumed-recommendation-ledger", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasRecommendationTrackRegistryContract, Artifact: "recommendation-track-registry", Command: "track-registry", TypedValidator: "typed:recommendation-track-registry", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasRecommendationCommandRunLedgerContract, Artifact: "recommendation-command-run-ledger", Command: "run-ledger", TypedValidator: "typed:recommendation-command-run-ledger", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasRecommendationEvidenceValidationReportContract, Artifact: "recommendation-evidence-validation-report", Command: "validate-evidence", TypedValidator: "typed:recommendation-evidence-validation-report", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasRecommendationFinalResponseGatesContract, Artifact: "recommendation-final-response-gates", Command: "final-response-gates", TypedValidator: "typed:recommendation-final-response-gates", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
		{Schema: AtlasRecommendationEvidenceSchemaRegistryCoverageContract, Artifact: "recommendation-evidence-schema-registry-coverage", Command: "schema-registry-coverage", TypedValidator: "typed:recommendation-evidence-schema-registry-coverage", StatusField: "status", SafetyClass: "planning_readback_no_execution", PlanningOnly: true},
	}
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
