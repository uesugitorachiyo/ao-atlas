package atlas

import "fmt"

func BuildAtlasRecommendationRunLedgerCoverageCheck(registryPath, rollupPath string) (AtlasRecommendationRunLedgerCoverageCheck, error) {
	if registryPath == "" {
		return AtlasRecommendationRunLedgerCoverageCheck{}, fmt.Errorf("registry path is required")
	}
	if rollupPath == "" {
		return AtlasRecommendationRunLedgerCoverageCheck{}, fmt.Errorf("rollup path is required")
	}
	registry, err := LoadJSON[AtlasRecommendationEvidenceSchemaRegistry](registryPath)
	if err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	if err := ValidateAtlasRecommendationEvidenceSchemaRegistry(registry); err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	rollup, err := LoadJSON[AtlasRecommendationCommandRunLedgerRollup](rollupPath)
	if err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	if err := ValidateAtlasRecommendationCommandRunLedgerRollup(rollup); err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	registrySummary, err := BuildAtlasRecommendationArtifactSummary(registryPath)
	if err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	rollupSummary, err := BuildAtlasRecommendationArtifactSummary(rollupPath)
	if err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}

	requiredCommands := []string{"schema-registry"}
	excludedCommands := []string{}
	for _, entry := range registry.Schemas {
		if entry.Command == "run-ledger" {
			excludedCommands = append(excludedCommands, entry.Command)
			continue
		}
		if !containsValue(requiredCommands, entry.Command) {
			requiredCommands = append(requiredCommands, entry.Command)
		}
	}

	rollupCommands := map[string]bool{}
	for _, command := range rollup.Commands {
		rollupCommands[command] = true
	}
	coveredCommands := []string{}
	missingCommands := []string{}
	for _, command := range requiredCommands {
		if rollupCommands[command] {
			coveredCommands = append(coveredCommands, command)
			continue
		}
		missingCommands = append(missingCommands, command)
	}
	status := "passed"
	if len(missingCommands) > 0 || !rollup.AllOutputsNoPromotion || rollup.PromotionGranted || rollup.ClaimsAuthorityAdvance || !rollup.RSIRemainsDenied {
		status = "failed"
	}

	check := AtlasRecommendationRunLedgerCoverageCheck{
		Schema:                         AtlasRecommendationRunLedgerCoverageCheckContract,
		Status:                         status,
		SourceRegistryPath:             registrySummary.PublicPath,
		SourceRegistryDigest:           registrySummary.Digest,
		SourceRollupPath:               rollupSummary.PublicPath,
		SourceRollupDigest:             rollupSummary.Digest,
		RegistryCommandCount:           registry.SchemaCount,
		RequiredCommandCount:           len(requiredCommands),
		CoveredCommandCount:            len(coveredCommands),
		MissingCommandCount:            len(missingCommands),
		RequiredCommands:               requiredCommands,
		CoveredCommands:                coveredCommands,
		MissingCommands:                missingCommands,
		ExcludedCommands:               excludedCommands,
		AllControlPlaneCommandsCovered: len(missingCommands) == 0,
		AllOutputsNoPromotion:          rollup.AllOutputsNoPromotion && !rollup.PromotionGranted && !rollup.ClaimsAuthorityAdvance,
		PromotionGranted:               false,
		ClaimsAuthorityAdvance:         false,
		RSIRemainsDenied:               rollup.RSIRemainsDenied,
		SafeToExecute:                  false,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		MutatesRepositories:            false,
	}
	if err := ValidateAtlasRecommendationRunLedgerCoverageCheck(check); err != nil {
		return AtlasRecommendationRunLedgerCoverageCheck{}, err
	}
	return check, nil
}

func ValidateAtlasRecommendationRunLedgerCoverageCheck(check AtlasRecommendationRunLedgerCoverageCheck) error {
	var errs []string
	requireContract(&errs, "recommendation_run_ledger_coverage_check", check.Schema, AtlasRecommendationRunLedgerCoverageCheckContract)
	if !oneOf(check.Status, "passed", "failed") {
		errs = append(errs, "status must be passed or failed")
	}
	for field, value := range map[string]string{
		"source_registry_path": check.SourceRegistryPath,
		"source_rollup_path":   check.SourceRollupPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	if !digestPattern.MatchString(check.SourceRegistryDigest) {
		errs = append(errs, "source_registry_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(check.SourceRollupDigest) {
		errs = append(errs, "source_rollup_digest must be sha256 digest")
	}
	if check.RegistryCommandCount <= 0 {
		errs = append(errs, "registry_command_count must be positive")
	}
	if check.RequiredCommandCount != len(check.RequiredCommands) {
		errs = append(errs, "required_command_count must match required_commands")
	}
	if check.CoveredCommandCount != len(check.CoveredCommands) {
		errs = append(errs, "covered_command_count must match covered_commands")
	}
	if check.MissingCommandCount != len(check.MissingCommands) {
		errs = append(errs, "missing_command_count must match missing_commands")
	}
	if check.RequiredCommandCount != check.CoveredCommandCount+check.MissingCommandCount {
		errs = append(errs, "required_command_count must equal covered plus missing commands")
	}
	for _, command := range check.RequiredCommands {
		requireField(&errs, "required_commands", command)
		checkPublicPath(&errs, "required_commands", command, true)
		if containsValue(check.ExcludedCommands, command) {
			errs = append(errs, "required_commands must not include excluded commands")
		}
	}
	for _, command := range check.CoveredCommands {
		requireField(&errs, "covered_commands", command)
		checkPublicPath(&errs, "covered_commands", command, true)
		if !containsValue(check.RequiredCommands, command) {
			errs = append(errs, "covered_commands must be required commands")
		}
	}
	for _, command := range check.MissingCommands {
		requireField(&errs, "missing_commands", command)
		checkPublicPath(&errs, "missing_commands", command, true)
		if !containsValue(check.RequiredCommands, command) {
			errs = append(errs, "missing_commands must be required commands")
		}
	}
	for _, command := range check.ExcludedCommands {
		requireField(&errs, "excluded_commands", command)
		checkPublicPath(&errs, "excluded_commands", command, true)
	}
	expectedCovered := check.MissingCommandCount == 0 && check.CoveredCommandCount == check.RequiredCommandCount
	if check.AllControlPlaneCommandsCovered != expectedCovered {
		errs = append(errs, "all_control_plane_commands_covered must match coverage counts")
	}
	if check.Status == "passed" && !check.AllControlPlaneCommandsCovered {
		errs = append(errs, "passed status requires all control plane commands covered")
	}
	if !check.AllOutputsNoPromotion {
		errs = append(errs, "all_outputs_no_promotion must remain true")
	}
	if check.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if check.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !check.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must remain true")
	}
	if check.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if check.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if check.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if check.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if check.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func WriteAtlasRecommendationRunLedgerCoverageCheck(path string, check AtlasRecommendationRunLedgerCoverageCheck) error {
	return WriteJSON(path, check)
}
