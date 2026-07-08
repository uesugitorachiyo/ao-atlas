package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AtlasSchemaHealthRepairPromptOptions struct {
	NodeID             string
	SourceCoveragePath string
	PromptPath         string
}

func BuildAtlasSchemaHealthRepairPrompt(coverage AtlasRecommendationEvidenceSchemaRegistryCoverage, options AtlasSchemaHealthRepairPromptOptions) (AtlasSchemaHealthRepairPrompt, string, error) {
	sourceCoveragePath := strings.TrimSpace(options.SourceCoveragePath)
	promptPath := strings.TrimSpace(options.PromptPath)
	fixture := AtlasSchemaHealthRepairPrompt{
		Schema:                 AtlasSchemaHealthRepairPromptContract,
		NodeID:                 strings.TrimSpace(options.NodeID),
		Status:                 "repair_prompt_generated",
		SourceCoveragePath:     publicArtifactRef(sourceCoveragePath),
		SourceCoverageDigest:   digestValue(coverage),
		PromptPath:             publicArtifactRef(promptPath),
		CoverageStatus:         coverage.Status,
		ValidationReportStatus: coverage.ValidationReportStatus,
		FailureReasons:         append([]string{}, coverage.FailureReasons...),
		MissingSchemaCount:     len(coverage.MissingSchemas),
		MissingSchemas:         append([]string{}, coverage.MissingSchemas...),
		MissingValidatorCount:  len(coverage.MissingValidators),
		MissingValidators:      append([]string{}, coverage.MissingValidators...),
		RepairActions: []string{
			"Add missing recommendation control-plane evidence artifacts or typed validators for uncovered registry entries.",
			"Rerun mission recommendations schema-registry-health and keep promotion denied.",
		},
		ExactNextAction:        "Repair recommendation schema-health coverage, rerun schema-registry-health, and keep promotion denied.",
		PlanningOnly:           true,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
		PromotionRequested:     false,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       coverage.RSIRemainsDenied,
	}
	prompt := buildAtlasSchemaHealthRepairPromptMarkdown(fixture)
	if err := ValidateAtlasSchemaHealthRepairPrompt(fixture); err != nil {
		return AtlasSchemaHealthRepairPrompt{}, "", err
	}
	return fixture, prompt, nil
}

func buildAtlasSchemaHealthRepairPromptMarkdown(fixture AtlasSchemaHealthRepairPrompt) string {
	var b strings.Builder
	b.WriteString("You are AO Atlas, repairing recommendation schema-health coverage.\n\n")
	b.WriteString("Current schema-health state:\n")
	b.WriteString(fmt.Sprintf("- Coverage artifact: `%s`\n", fixture.SourceCoveragePath))
	b.WriteString(fmt.Sprintf("- Coverage status: `%s`\n", fixture.CoverageStatus))
	b.WriteString(fmt.Sprintf("- Validation report status: `%s`\n", fixture.ValidationReportStatus))
	b.WriteString(fmt.Sprintf("- Missing schemas: `%d`\n", fixture.MissingSchemaCount))
	b.WriteString(fmt.Sprintf("- Missing validators: `%d`\n", fixture.MissingValidatorCount))
	if len(fixture.FailureReasons) != 0 {
		b.WriteString(fmt.Sprintf("- Failure reasons: `%s`\n", strings.Join(fixture.FailureReasons, ",")))
	}
	b.WriteString("\nRepair actions:\n")
	for _, action := range fixture.RepairActions {
		b.WriteString(fmt.Sprintf("- %s\n", action))
	}
	b.WriteString("\nExecution rules:\n")
	b.WriteString("- Keep this repair planning-only until a bounded implementation node is separately selected.\n")
	b.WriteString("- Rerun `mission recommendations schema-registry-health` after repair.\n")
	b.WriteString("- No promotion is requested.\n")
	b.WriteString("- Do not claim authority advance.\n")
	b.WriteString("- RSI remains denied.\n")
	return b.String()
}

func ValidateAtlasSchemaHealthRepairPrompt(fixture AtlasSchemaHealthRepairPrompt) error {
	var errs []string
	requireContract(&errs, "schema_health_repair_prompt", fixture.Schema, AtlasSchemaHealthRepairPromptContract)
	if fixture.Status != "repair_prompt_generated" {
		errs = append(errs, "status must be repair_prompt_generated")
	}
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	requireField(&errs, "source_coverage_path", fixture.SourceCoveragePath)
	checkPublicPath(&errs, "source_coverage_path", fixture.SourceCoveragePath, true)
	if !digestPattern.MatchString(fixture.SourceCoverageDigest) {
		errs = append(errs, "source_coverage_digest must be sha256 digest")
	}
	requireField(&errs, "prompt_path", fixture.PromptPath)
	checkPublicPath(&errs, "prompt_path", fixture.PromptPath, true)
	if fixture.CoverageStatus != "failed" {
		errs = append(errs, "coverage_status must be failed")
	}
	requireField(&errs, "validation_report_status", fixture.ValidationReportStatus)
	if fixture.MissingSchemaCount != len(fixture.MissingSchemas) {
		errs = append(errs, "missing_schema_count must match missing_schemas")
	}
	if fixture.MissingValidatorCount != len(fixture.MissingValidators) {
		errs = append(errs, "missing_validator_count must match missing_validators")
	}
	if fixture.MissingSchemaCount == 0 && fixture.MissingValidatorCount == 0 && len(fixture.FailureReasons) == 0 {
		errs = append(errs, "repair prompt requires missing schema, validator, or failure reason evidence")
	}
	if !schemaHealthRepairActionsContain(fixture.RepairActions, "Add missing recommendation control-plane evidence artifacts or typed validators for uncovered registry entries.") {
		errs = append(errs, "repair_actions must include missing artifact or validator repair")
	}
	if !schemaHealthRepairActionsContain(fixture.RepairActions, "Rerun mission recommendations schema-registry-health and keep promotion denied.") {
		errs = append(errs, "repair_actions must include schema health rerun with promotion denied")
	}
	requireField(&errs, "exact_next_action", fixture.ExactNextAction)
	checkPublicPath(&errs, "exact_next_action", fixture.ExactNextAction, true)
	if !fixture.PlanningOnly {
		errs = append(errs, "planning_only must be true")
	}
	if fixture.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if fixture.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if fixture.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if fixture.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if fixture.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if fixture.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if fixture.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func schemaHealthRepairActionsContain(actions []string, want string) bool {
	for _, action := range actions {
		if action == want {
			return true
		}
	}
	return false
}

func WriteAtlasSchemaHealthRepairPrompt(promptPath, fixturePath string, fixture AtlasSchemaHealthRepairPrompt, prompt string) error {
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return err
	}
	return WriteJSON(fixturePath, fixture)
}
