package atlas

import (
	"fmt"
	"sort"
	"strings"
)

func BuildAtlasPRCITimingSummary(ledgerPath string) (AtlasPRCITimingSummary, error) {
	ledger, err := LoadJSON[AtlasPRCITimingLedger](ledgerPath)
	if err != nil {
		return AtlasPRCITimingSummary{}, err
	}
	if err := ValidateAtlasPRCITimingLedger(ledger); err != nil {
		return AtlasPRCITimingSummary{}, err
	}
	summary := summarizePRCITimingRows(ledger.Rows)
	summary.Schema = AtlasPRCITimingSummaryContract
	summary.Status = "summarized"
	summary.SourceLedgerPath = publicArtifactRef(ledgerPath)
	summary.SourceLedgerDigest = digestValue(ledger)
	summary.SchedulesWork = false
	summary.ExecutesWork = false
	summary.ApprovesWork = false
	summary.ClaimsAuthorityAdvance = false
	summary.RSIRemainsDenied = true
	if err := ValidateAtlasPRCITimingSummary(summary); err != nil {
		return AtlasPRCITimingSummary{}, err
	}
	return summary, nil
}

func ValidateAtlasPRCITimingLedger(ledger AtlasPRCITimingLedger) error {
	var errs []string
	requireContract(&errs, "pr_ci_timing_ledger", ledger.Schema, AtlasPRCITimingLedgerContract)
	if ledger.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	requireField(&errs, "evidence_root", ledger.EvidenceRoot)
	checkPublicPath(&errs, "evidence_root", ledger.EvidenceRoot, true)
	if len(ledger.Rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	validatePRCITimingRows(&errs, ledger.Rows)
	if ledger.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if ledger.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if ledger.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if ledger.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !ledger.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func ValidateAtlasPRCITimingSummary(summary AtlasPRCITimingSummary) error {
	var errs []string
	requireContract(&errs, "pr_ci_timing_summary", summary.Schema, AtlasPRCITimingSummaryContract)
	if summary.Status != "summarized" {
		errs = append(errs, "status must be summarized")
	}
	requireField(&errs, "source_ledger_path", summary.SourceLedgerPath)
	checkPublicPath(&errs, "source_ledger_path", summary.SourceLedgerPath, true)
	if !digestPattern.MatchString(summary.SourceLedgerDigest) {
		errs = append(errs, "source_ledger_digest must be sha256 digest")
	}
	if len(summary.Rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	validatePRCITimingRows(&errs, summary.Rows)
	expected := summarizePRCITimingRows(summary.Rows)
	if summary.RowCount != expected.RowCount {
		errs = append(errs, "row_count must match rows length")
	}
	if summary.MergedPRs != expected.MergedPRs {
		errs = append(errs, "merged_prs must match rows with merge commits")
	}
	if summary.CIPassedPRs != expected.CIPassedPRs || summary.CIFailedPRs != expected.CIFailedPRs || summary.CIPendingPRs != expected.CIPendingPRs {
		errs = append(errs, "ci status counts must match rows")
	}
	if !equalInts(summary.PRNumbers, expected.PRNumbers) {
		errs = append(errs, "pr_numbers must match sorted row PR numbers")
	}
	if !equalStrings(summary.NodeIDs, expected.NodeIDs) {
		errs = append(errs, "node_ids must match sorted row node IDs")
	}
	if summary.TotalUbuntuSeconds != expected.TotalUbuntuSeconds ||
		summary.TotalMacosSeconds != expected.TotalMacosSeconds ||
		summary.TotalWindowsSeconds != expected.TotalWindowsSeconds {
		errs = append(errs, "total platform seconds must match rows")
	}
	if summary.MeanUbuntuSeconds != expected.MeanUbuntuSeconds ||
		summary.MeanMacosSeconds != expected.MeanMacosSeconds ||
		summary.MeanWindowsSeconds != expected.MeanWindowsSeconds {
		errs = append(errs, "mean platform seconds must match totals")
	}
	if summary.MaxWindowsSeconds != expected.MaxWindowsSeconds ||
		summary.MaxCheckSeconds != expected.MaxCheckSeconds ||
		summary.SlowestPRNumber != expected.SlowestPRNumber ||
		summary.SlowestNodeID != expected.SlowestNodeID ||
		summary.SlowestCheck != expected.SlowestCheck {
		errs = append(errs, "slowest check summary must match rows")
	}
	if summary.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if summary.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if summary.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if summary.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !summary.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func summarizePRCITimingRows(rows []AtlasPRCITimingRow) AtlasPRCITimingSummary {
	normalized := append([]AtlasPRCITimingRow(nil), rows...)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].PRNumber == normalized[j].PRNumber {
			return normalized[i].NodeID < normalized[j].NodeID
		}
		return normalized[i].PRNumber < normalized[j].PRNumber
	})
	summary := AtlasPRCITimingSummary{
		RowCount: len(normalized),
		Rows:     normalized,
	}
	for _, row := range normalized {
		summary.PRNumbers = append(summary.PRNumbers, row.PRNumber)
		summary.NodeIDs = append(summary.NodeIDs, row.NodeID)
		if strings.TrimSpace(row.MergeCommit) != "" {
			summary.MergedPRs++
		}
		switch row.CIStatus {
		case "passed":
			summary.CIPassedPRs++
		case "failed":
			summary.CIFailedPRs++
		case "pending":
			summary.CIPendingPRs++
		}
		summary.TotalUbuntuSeconds += row.UbuntuSeconds
		summary.TotalMacosSeconds += row.MacosSeconds
		summary.TotalWindowsSeconds += row.WindowsSeconds
		if row.WindowsSeconds > summary.MaxWindowsSeconds {
			summary.MaxWindowsSeconds = row.WindowsSeconds
		}
		if row.MaxCheckSeconds > summary.MaxCheckSeconds {
			summary.MaxCheckSeconds = row.MaxCheckSeconds
			summary.SlowestPRNumber = row.PRNumber
			summary.SlowestNodeID = row.NodeID
			summary.SlowestCheck = row.SlowestCheck
		}
	}
	if summary.RowCount > 0 {
		summary.MeanUbuntuSeconds = summary.TotalUbuntuSeconds / summary.RowCount
		summary.MeanMacosSeconds = summary.TotalMacosSeconds / summary.RowCount
		summary.MeanWindowsSeconds = summary.TotalWindowsSeconds / summary.RowCount
	}
	return summary
}

func validatePRCITimingRows(errs *[]string, rows []AtlasPRCITimingRow) {
	seenPRs := map[int]bool{}
	previousPR := 0
	for i, row := range rows {
		prefix := fmt.Sprintf("rows[%d]", i)
		requireField(errs, prefix+".node_id", row.NodeID)
		checkPublicPath(errs, prefix+".node_id", row.NodeID, true)
		if row.PRNumber <= 0 {
			*errs = append(*errs, prefix+".pr_number must be greater than zero")
		}
		if seenPRs[row.PRNumber] {
			*errs = append(*errs, "rows pr_number values must be unique")
		}
		seenPRs[row.PRNumber] = true
		if previousPR != 0 && row.PRNumber < previousPR {
			*errs = append(*errs, "rows must be sorted by pr_number")
		}
		previousPR = row.PRNumber
		requireField(errs, prefix+".merge_commit", row.MergeCommit)
		if len(row.MergeCommit) != 40 {
			*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if !oneOf(row.CIStatus, "passed", "failed", "pending") {
			*errs = append(*errs, prefix+".ci_status must be passed, failed, or pending")
		}
		if row.UbuntuSeconds < 0 || row.MacosSeconds < 0 || row.WindowsSeconds < 0 || row.MaxCheckSeconds < 0 {
			*errs = append(*errs, prefix+".seconds fields must be non-negative")
		}
		if row.MaxCheckSeconds != maxInt(row.UbuntuSeconds, row.MacosSeconds, row.WindowsSeconds) {
			*errs = append(*errs, prefix+".max_check_seconds must equal the row platform maximum")
		}
		requireField(errs, prefix+".slowest_check", row.SlowestCheck)
		checkPublicPath(errs, prefix+".slowest_check", row.SlowestCheck, true)
	}
}

func FeatureDepthPRCITimingRowsToNormalizedInputs(rows []AtlasPRCITimingRow) []AtlasPRCINormalizedRow {
	normalized := make([]AtlasPRCINormalizedRow, 0, len(rows))
	for _, row := range rows {
		ciPassed := row.CIStatus == "passed"
		normalized = append(normalized, AtlasPRCINormalizedRow{
			NodeID:          row.NodeID,
			PRNumber:        row.PRNumber,
			MergeCommit:     row.MergeCommit,
			CIStatus:        row.CIStatus,
			CheckCount:      3,
			SuccessCount:    boolToInt(ciPassed) * 3,
			UbuntuPassed:    ciPassed,
			MacosPassed:     ciPassed,
			WindowsPassed:   ciPassed,
			UbuntuSeconds:   row.UbuntuSeconds,
			MacosSeconds:    row.MacosSeconds,
			WindowsSeconds:  row.WindowsSeconds,
			MaxCheckSeconds: row.MaxCheckSeconds,
			SlowestCheck:    row.SlowestCheck,
		})
	}
	return normalized
}

func ClosurePRCILedgerEntriesToNormalizedInputs(entries []AtlasPRCILedgerEntry) []AtlasPRCINormalizedRow {
	normalized := make([]AtlasPRCINormalizedRow, 0, len(entries))
	for _, entry := range entries {
		normalized = append(normalized, AtlasPRCINormalizedRow{
			NodeID:        entry.NodeID,
			PRNumber:      entry.PRNumber,
			MergeCommit:   entry.MergeCommit,
			CIStatus:      entry.CIStatus,
			CheckCount:    entry.CheckCount,
			SuccessCount:  entry.SuccessCount,
			UbuntuPassed:  entry.UbuntuSuccessCount > 0,
			MacosPassed:   entry.MacOSSuccessCount > 0,
			WindowsPassed: entry.WindowsSuccessCount > 0,
		})
	}
	return normalized
}

func NormalizeAtlasPRCILedgerRows(inputs []AtlasPRCINormalizationInput) ([]AtlasPRCINormalizedRow, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one PR/CI normalization input is required")
	}
	rows := []AtlasPRCINormalizedRow{}
	for inputIndex, input := range inputs {
		sourceWave := strings.TrimSpace(input.SourceWave)
		if sourceWave == "" {
			return nil, fmt.Errorf("inputs[%d].source_wave is required", inputIndex)
		}
		if len(input.Rows) == 0 {
			return nil, fmt.Errorf("inputs[%d].rows must not be empty", inputIndex)
		}
		for _, rawRow := range input.Rows {
			row := normalizePRCIRow(sourceWave, rawRow)
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PRNumber == rows[j].PRNumber {
			return rows[i].SourceWave < rows[j].SourceWave
		}
		return rows[i].PRNumber < rows[j].PRNumber
	})
	if err := ValidateAtlasPRCINormalizedRows(rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func normalizePRCIRow(sourceWave string, row AtlasPRCINormalizedRow) AtlasPRCINormalizedRow {
	row.NormalizedSchema = AtlasPRCINormalizedRowContract
	row.SourceWave = sourceWave
	row.NodeID = strings.TrimSpace(row.NodeID)
	row.MergeCommit = strings.TrimSpace(row.MergeCommit)
	row.CIStatus = strings.TrimSpace(row.CIStatus)
	row.SlowestCheck = strings.TrimSpace(row.SlowestCheck)
	row.Merged = len(row.MergeCommit) == 40
	row.AllRequiredChecksPassed = row.CIStatus == "passed" &&
		row.CheckCount > 0 &&
		row.SuccessCount == row.CheckCount &&
		row.UbuntuPassed &&
		row.MacosPassed &&
		row.WindowsPassed
	row.PromotionGranted = false
	row.ClaimsAuthorityAdvance = false
	row.RSIRemainsDenied = true
	row.SafeToExecute = false
	row.SchedulesWork = false
	row.ExecutesWork = false
	row.ApprovesWork = false
	row.MutatesRepositories = false
	return row
}

func ValidateAtlasPRCINormalizedRows(rows []AtlasPRCINormalizedRow) error {
	var errs []string
	if len(rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	seenPRs := map[int]bool{}
	previousPR := 0
	for index, row := range rows {
		prefix := fmt.Sprintf("rows[%d]", index)
		requireContract(&errs, prefix+".normalized_schema", row.NormalizedSchema, AtlasPRCINormalizedRowContract)
		requireField(&errs, prefix+".source_wave", row.SourceWave)
		checkPublicPath(&errs, prefix+".source_wave", row.SourceWave, true)
		requireField(&errs, prefix+".node_id", row.NodeID)
		checkPublicPath(&errs, prefix+".node_id", row.NodeID, true)
		if row.PRNumber <= 0 {
			errs = append(errs, prefix+".pr_number must be greater than zero")
		}
		if seenPRs[row.PRNumber] {
			errs = append(errs, "rows pr_number values must be unique")
		}
		seenPRs[row.PRNumber] = true
		if previousPR != 0 && row.PRNumber < previousPR {
			errs = append(errs, "rows must be sorted by pr_number")
		}
		previousPR = row.PRNumber
		requireField(&errs, prefix+".merge_commit", row.MergeCommit)
		if len(row.MergeCommit) != 40 {
			errs = append(errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if !oneOf(row.CIStatus, "passed", "failed", "pending") {
			errs = append(errs, prefix+".ci_status must be passed, failed, or pending")
		}
		if row.Merged != (len(row.MergeCommit) == 40) {
			errs = append(errs, prefix+".merged must match merge_commit")
		}
		if row.CheckCount <= 0 {
			errs = append(errs, prefix+".check_count must be positive")
		}
		if row.SuccessCount < 0 || row.SuccessCount > row.CheckCount {
			errs = append(errs, prefix+".success_count must be between zero and check_count")
		}
		expectedPassed := row.CIStatus == "passed" &&
			row.CheckCount > 0 &&
			row.SuccessCount == row.CheckCount &&
			row.UbuntuPassed &&
			row.MacosPassed &&
			row.WindowsPassed
		if row.AllRequiredChecksPassed != expectedPassed {
			errs = append(errs, prefix+".all_required_checks_passed must match CI status and platform checks")
		}
		if row.UbuntuSeconds < 0 || row.MacosSeconds < 0 || row.WindowsSeconds < 0 || row.MaxCheckSeconds < 0 {
			errs = append(errs, prefix+".seconds fields must be non-negative")
		}
		if row.MaxCheckSeconds > 0 && row.MaxCheckSeconds != maxInt(row.UbuntuSeconds, row.MacosSeconds, row.WindowsSeconds) {
			errs = append(errs, prefix+".max_check_seconds must equal the row platform maximum")
		}
		if row.PromotionGranted {
			errs = append(errs, prefix+".promotion_granted must be false")
		}
		if row.ClaimsAuthorityAdvance {
			errs = append(errs, prefix+".claims_authority_advance must be false")
		}
		if !row.RSIRemainsDenied {
			errs = append(errs, prefix+".rsi_remains_denied must be true")
		}
		if row.SafeToExecute || row.SchedulesWork || row.ExecutesWork || row.ApprovesWork || row.MutatesRepositories {
			errs = append(errs, prefix+" must not execute, schedule, approve, mutate repositories, or mark safe_to_execute")
		}
	}
	return joinErrors(errs)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func maxInt(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
