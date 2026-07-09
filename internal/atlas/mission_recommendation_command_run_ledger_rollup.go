package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildAtlasRecommendationCommandRunLedgerRollup(ledgerPaths []string) (AtlasRecommendationCommandRunLedgerRollup, error) {
	if len(ledgerPaths) == 0 {
		return AtlasRecommendationCommandRunLedgerRollup{}, fmt.Errorf("at least one ledger is required")
	}
	entries := []AtlasRecommendationCommandRunLedgerRollupEntry{}
	commands := []string{}
	outputStatusCounts := map[string]int{}
	failedCommands := []string{}
	allRecordInvocation := true
	allNoPromotion := true
	allRSIDenied := true

	for _, rawPath := range ledgerPaths {
		ledgerPath := filepath.ToSlash(strings.TrimSpace(rawPath))
		if ledgerPath == "" {
			return AtlasRecommendationCommandRunLedgerRollup{}, fmt.Errorf("ledger path is required")
		}
		ledger, err := LoadJSON[AtlasRecommendationCommandRunLedger](ledgerPath)
		if err != nil {
			return AtlasRecommendationCommandRunLedgerRollup{}, err
		}
		if err := ValidateAtlasRecommendationCommandRunLedger(ledger); err != nil {
			return AtlasRecommendationCommandRunLedgerRollup{}, err
		}
		ledgerSummary, err := BuildAtlasRecommendationArtifactSummary(ledgerPath)
		if err != nil {
			return AtlasRecommendationCommandRunLedgerRollup{}, err
		}
		artifactSummary := recommendationArtifactSummaryFromCommandRunLedger(ledger)
		entries = append(entries, AtlasRecommendationCommandRunLedgerRollupEntry{
			LedgerPath:             ledgerSummary.PublicPath,
			LedgerDigest:           ledgerSummary.Digest,
			Command:                ledger.Command,
			ArtifactSchema:         artifactSummary.Schema,
			TypedValidator:         artifactSummary.TypedValidator,
			OutputStatus:           artifactSummary.OutputStatus,
			ArtifactPath:           artifactSummary.PublicPath,
			ArtifactDigest:         artifactSummary.Digest,
			RecordsInvocation:      ledger.RecordsInvocation,
			NoPromotionRequested:   ledger.NoPromotionRequested,
			PromotionGranted:       ledger.PromotionGranted,
			ClaimsAuthorityAdvance: ledger.ClaimsAuthorityAdvance,
			RSIRemainsDenied:       ledger.RSIRemainsDenied,
		})
		commands = append(commands, ledger.Command)
		outputStatusCounts[ledger.OutputStatus]++
		if ClassifyAtlasRecommendationRunLedgerOutputStatus(ledger.OutputStatus).CountsAsFailedOutput {
			failedCommands = append(failedCommands, ledger.Command)
		}
		allRecordInvocation = allRecordInvocation && ledger.RecordsInvocation
		allNoPromotion = allNoPromotion && ledger.NoPromotionRequested && !ledger.PromotionGranted && !ledger.ClaimsAuthorityAdvance
		allRSIDenied = allRSIDenied && ledger.RSIRemainsDenied
	}

	rollup := AtlasRecommendationCommandRunLedgerRollup{
		Schema:                     AtlasRecommendationCommandRunLedgerRollupContract,
		Status:                     "rolled_up",
		LedgerCount:                len(entries),
		Ledgers:                    entries,
		Commands:                   commands,
		OutputStatusCounts:         outputStatusCounts,
		FailedOutputCount:          len(failedCommands),
		FailedCommands:             failedCommands,
		AllLedgersRecordInvocation: allRecordInvocation,
		AllOutputsNoPromotion:      allNoPromotion,
		PromotionGranted:           false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           allRSIDenied,
		SafeToExecute:              false,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		MutatesRepositories:        false,
	}
	if err := ValidateAtlasRecommendationCommandRunLedgerRollup(rollup); err != nil {
		return AtlasRecommendationCommandRunLedgerRollup{}, err
	}
	return rollup, nil
}

func recommendationArtifactSummaryFromCommandRunLedger(ledger AtlasRecommendationCommandRunLedger) AtlasRecommendationArtifactSummary {
	return AtlasRecommendationArtifactSummary{
		Path:           ledger.ArtifactPath,
		PublicPath:     publicArtifactRef(ledger.ArtifactPath),
		Digest:         ledger.ArtifactDigest,
		Schema:         ledger.ArtifactSchema,
		TypedValidator: ledger.TypedValidator,
		OutputStatus:   ledger.OutputStatus,
	}
}

func ClassifyAtlasRecommendationRunLedgerOutputStatus(status string) AtlasRecommendationRunLedgerOutputStatusClassification {
	status = strings.TrimSpace(status)
	category := "failed"
	countsAsFailed := true
	switch status {
	case "passed":
		category = "pass"
		countsAsFailed = false
	case "ready":
		category = "ready"
		countsAsFailed = false
	case "blocked", "blocked_hard_blocker":
		category = "blocked"
	default:
		category = "failed"
	}
	return AtlasRecommendationRunLedgerOutputStatusClassification{
		OutputStatus:         status,
		Category:             category,
		CountsAsFailedOutput: countsAsFailed,
	}
}

func ValidateAtlasRecommendationCommandRunLedgerRollup(rollup AtlasRecommendationCommandRunLedgerRollup) error {
	var errs []string
	requireContract(&errs, "recommendation_command_run_ledger_rollup", rollup.Schema, AtlasRecommendationCommandRunLedgerRollupContract)
	if rollup.Status != "rolled_up" {
		errs = append(errs, "status must be rolled_up")
	}
	if rollup.LedgerCount <= 0 {
		errs = append(errs, "ledger_count must be positive")
	}
	if rollup.LedgerCount != len(rollup.Ledgers) {
		errs = append(errs, "ledger_count must match ledgers")
	}
	if len(rollup.Commands) != len(rollup.Ledgers) {
		errs = append(errs, "commands must match ledgers")
	}
	expectedStatusCounts := map[string]int{}
	expectedFailedCommands := []string{}
	expectedRecordInvocation := true
	expectedNoPromotion := true
	expectedRSIDenied := true
	for index, entry := range rollup.Ledgers {
		prefix := fmt.Sprintf("ledgers[%d]", index)
		requireField(&errs, prefix+".ledger_path", entry.LedgerPath)
		checkPublicPath(&errs, prefix+".ledger_path", entry.LedgerPath, true)
		if !digestPattern.MatchString(entry.LedgerDigest) {
			errs = append(errs, prefix+".ledger_digest must be sha256 digest")
		}
		requireField(&errs, prefix+".command", entry.Command)
		requireField(&errs, prefix+".artifact_schema", entry.ArtifactSchema)
		requireField(&errs, prefix+".typed_validator", entry.TypedValidator)
		requireField(&errs, prefix+".output_status", entry.OutputStatus)
		requireField(&errs, prefix+".artifact_path", entry.ArtifactPath)
		checkPublicPath(&errs, prefix+".command", entry.Command, true)
		checkPublicPath(&errs, prefix+".artifact_schema", entry.ArtifactSchema, true)
		checkPublicPath(&errs, prefix+".typed_validator", entry.TypedValidator, true)
		checkPublicPath(&errs, prefix+".output_status", entry.OutputStatus, true)
		checkPublicPath(&errs, prefix+".artifact_path", entry.ArtifactPath, true)
		if !digestPattern.MatchString(entry.ArtifactDigest) {
			errs = append(errs, prefix+".artifact_digest must be sha256 digest")
		}
		if index < len(rollup.Commands) && rollup.Commands[index] != entry.Command {
			errs = append(errs, prefix+".command must match commands order")
		}
		expectedStatusCounts[entry.OutputStatus]++
		if ClassifyAtlasRecommendationRunLedgerOutputStatus(entry.OutputStatus).CountsAsFailedOutput {
			expectedFailedCommands = append(expectedFailedCommands, entry.Command)
		}
		expectedRecordInvocation = expectedRecordInvocation && entry.RecordsInvocation
		expectedNoPromotion = expectedNoPromotion && entry.NoPromotionRequested && !entry.PromotionGranted && !entry.ClaimsAuthorityAdvance
		expectedRSIDenied = expectedRSIDenied && entry.RSIRemainsDenied
	}
	if !intMapsEqual(rollup.OutputStatusCounts, expectedStatusCounts) {
		errs = append(errs, "output_status_counts must match ledgers")
	}
	if rollup.FailedOutputCount != len(rollup.FailedCommands) || rollup.FailedOutputCount != len(expectedFailedCommands) {
		errs = append(errs, "failed_output_count must match failed_commands")
	}
	if !stringSlicesEqual(rollup.FailedCommands, expectedFailedCommands) {
		errs = append(errs, "failed_commands must match non-passed ledger outputs")
	}
	if rollup.AllLedgersRecordInvocation != expectedRecordInvocation {
		errs = append(errs, "all_ledgers_record_invocation must match ledgers")
	}
	if rollup.AllOutputsNoPromotion != expectedNoPromotion {
		errs = append(errs, "all_outputs_no_promotion must match ledgers")
	}
	if !rollup.AllOutputsNoPromotion {
		errs = append(errs, "all_outputs_no_promotion must remain true")
	}
	if rollup.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if rollup.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if rollup.RSIRemainsDenied != expectedRSIDenied {
		errs = append(errs, "rsi_remains_denied must match ledgers")
	}
	if !rollup.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must remain true")
	}
	if rollup.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if rollup.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if rollup.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if rollup.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if rollup.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func WriteAtlasRecommendationCommandRunLedgerRollup(path string, rollup AtlasRecommendationCommandRunLedgerRollup) error {
	return WriteJSON(path, rollup)
}

func BuildAtlasRecommendationRunLedgerOperatorSummaryBinding(rollupPath, summaryCheckPath string) (AtlasRecommendationRunLedgerOperatorSummaryBinding, error) {
	rollupPath = filepath.ToSlash(strings.TrimSpace(rollupPath))
	summaryCheckPath = filepath.ToSlash(strings.TrimSpace(summaryCheckPath))
	if rollupPath == "" {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, fmt.Errorf("rollup path is required")
	}
	if summaryCheckPath == "" {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, fmt.Errorf("operator summary check path is required")
	}
	rollup, err := LoadJSON[AtlasRecommendationCommandRunLedgerRollup](rollupPath)
	if err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	if err := ValidateAtlasRecommendationCommandRunLedgerRollup(rollup); err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	summaryCheck, err := LoadJSON[AtlasMissionOperatorSummaryCheck](summaryCheckPath)
	if err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	if err := ValidateAtlasMissionOperatorSummaryCheck(summaryCheck); err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	rollupDigest, err := digestFile(rollupPath)
	if err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	summaryCheckDigest, err := digestFile(summaryCheckPath)
	if err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	binding := AtlasRecommendationRunLedgerOperatorSummaryBinding{
		Schema:                           AtlasRecommendationRunLedgerOperatorSummaryBindingContract,
		Status:                           "bound",
		SourceRollupPath:                 publicArtifactRef(rollupPath),
		SourceRollupDigest:               rollupDigest,
		SourceOperatorSummaryCheckPath:   publicArtifactRef(summaryCheckPath),
		SourceOperatorSummaryCheckDigest: summaryCheckDigest,
		RollupLedgerCount:                rollup.LedgerCount,
		RollupFailedOutputCount:          rollup.FailedOutputCount,
		OperatorSummaryStatus:            summaryCheck.Status,
		OperatorSummaryExactNextAction:   summaryCheck.ExactNextAction,
		SummaryRequiresOwnRunLedger:      false,
		RollupRequiresSummaryRunLedger:   false,
		SelfReferentialLedgerRequirement: false,
		AllOutputsNoPromotion:            rollup.AllOutputsNoPromotion && !rollup.PromotionGranted && !rollup.ClaimsAuthorityAdvance,
		PromotionGranted:                 false,
		ClaimsAuthorityAdvance:           false,
		RSIRemainsDenied:                 rollup.RSIRemainsDenied && summaryCheck.RSIRemainsDenied,
		SafeToExecute:                    false,
		SchedulesWork:                    false,
		ExecutesWork:                     false,
		ApprovesWork:                     false,
		MutatesRepositories:              false,
	}
	if err := ValidateAtlasRecommendationRunLedgerOperatorSummaryBinding(binding); err != nil {
		return AtlasRecommendationRunLedgerOperatorSummaryBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasRecommendationRunLedgerOperatorSummaryBinding(binding AtlasRecommendationRunLedgerOperatorSummaryBinding) error {
	var errs []string
	requireContract(&errs, "recommendation_run_ledger_operator_summary_binding", binding.Schema, AtlasRecommendationRunLedgerOperatorSummaryBindingContract)
	if binding.Status != "bound" {
		errs = append(errs, "status must be bound")
	}
	for field, value := range map[string]string{
		"source_rollup_path":                 binding.SourceRollupPath,
		"source_operator_summary_check_path": binding.SourceOperatorSummaryCheckPath,
		"operator_summary_status":            binding.OperatorSummaryStatus,
		"operator_summary_exact_next_action": binding.OperatorSummaryExactNextAction,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	if !digestPattern.MatchString(binding.SourceRollupDigest) {
		errs = append(errs, "source_rollup_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(binding.SourceOperatorSummaryCheckDigest) {
		errs = append(errs, "source_operator_summary_check_digest must be sha256 digest")
	}
	if binding.RollupLedgerCount <= 0 {
		errs = append(errs, "rollup_ledger_count must be positive")
	}
	if binding.RollupFailedOutputCount < 0 {
		errs = append(errs, "rollup_failed_output_count must not be negative")
	}
	if binding.SummaryRequiresOwnRunLedger {
		errs = append(errs, "summary_requires_own_run_ledger must be false")
	}
	if binding.RollupRequiresSummaryRunLedger {
		errs = append(errs, "rollup_requires_summary_run_ledger must be false")
	}
	if binding.SelfReferentialLedgerRequirement {
		errs = append(errs, "self_referential_ledger_requirement must be false")
	}
	if !binding.AllOutputsNoPromotion {
		errs = append(errs, "all_outputs_no_promotion must be true")
	}
	if binding.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if binding.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !binding.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if binding.SafeToExecute || binding.SchedulesWork || binding.ExecutesWork || binding.ApprovesWork || binding.MutatesRepositories {
		errs = append(errs, "binding must not execute, schedule, approve, mutate repositories, or mark safe_to_execute")
	}
	return joinErrors(errs)
}

func intMapsEqual(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
