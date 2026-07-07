package atlas

import "fmt"

func BuildAtlasPRCIWindowsThresholdEvidence(summaryPath string, thresholdSeconds int) (AtlasPRCIWindowsThresholdEvidence, error) {
	summary, err := LoadJSON[AtlasPRCITimingSummary](summaryPath)
	if err != nil {
		return AtlasPRCIWindowsThresholdEvidence{}, err
	}
	if err := ValidateAtlasPRCITimingSummary(summary); err != nil {
		return AtlasPRCIWindowsThresholdEvidence{}, err
	}
	evidence := summarizeWindowsThresholdRows(summary.Rows, thresholdSeconds)
	evidence.Schema = AtlasPRCIWindowsThresholdEvidenceContract
	evidence.Status = "windows_thresholds_recorded"
	evidence.SourceSummaryPath = publicArtifactRef(summaryPath)
	evidence.SourceSummaryDigest = digestValue(summary)
	evidence.SchedulesWork = false
	evidence.ExecutesWork = false
	evidence.ApprovesWork = false
	evidence.ClaimsAuthorityAdvance = false
	evidence.RSIRemainsDenied = true
	if err := ValidateAtlasPRCIWindowsThresholdEvidence(evidence); err != nil {
		return AtlasPRCIWindowsThresholdEvidence{}, err
	}
	return evidence, nil
}

func ValidateAtlasPRCIWindowsThresholdEvidence(evidence AtlasPRCIWindowsThresholdEvidence) error {
	var errs []string
	requireContract(&errs, "pr_ci_windows_threshold_evidence", evidence.Schema, AtlasPRCIWindowsThresholdEvidenceContract)
	if evidence.Status != "windows_thresholds_recorded" {
		errs = append(errs, "status must be windows_thresholds_recorded")
	}
	requireField(&errs, "source_summary_path", evidence.SourceSummaryPath)
	checkPublicPath(&errs, "source_summary_path", evidence.SourceSummaryPath, true)
	if !digestPattern.MatchString(evidence.SourceSummaryDigest) {
		errs = append(errs, "source_summary_digest must be sha256 digest")
	}
	if evidence.ThresholdSeconds <= 0 {
		errs = append(errs, "threshold_seconds must be greater than zero")
	}
	if len(evidence.Rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	validateWindowsThresholdRows(&errs, evidence.Rows, evidence.ThresholdSeconds)
	expected := summarizeThresholdEvidenceRows(evidence.Rows, evidence.ThresholdSeconds)
	if evidence.RowCount != expected.RowCount {
		errs = append(errs, "row_count must match rows length")
	}
	if evidence.LongRunningWindowsChecks != expected.LongRunningWindowsChecks {
		errs = append(errs, "long_running_windows_checks must match rows")
	}
	if evidence.MaxWindowsSeconds != expected.MaxWindowsSeconds {
		errs = append(errs, "max_windows_seconds must match rows")
	}
	if evidence.MaxOverThresholdSeconds != expected.MaxOverThresholdSeconds {
		errs = append(errs, "max_over_threshold_seconds must match rows")
	}
	if evidence.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if evidence.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if evidence.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if evidence.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !evidence.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func summarizeWindowsThresholdRows(rows []AtlasPRCITimingRow, thresholdSeconds int) AtlasPRCIWindowsThresholdEvidence {
	thresholdRows := make([]AtlasPRCIWindowsThresholdRow, 0, len(rows))
	for _, row := range rows {
		over := row.WindowsSeconds - thresholdSeconds
		if over < 0 {
			over = 0
		}
		thresholdRows = append(thresholdRows, AtlasPRCIWindowsThresholdRow{
			NodeID:               row.NodeID,
			PRNumber:             row.PRNumber,
			CIStatus:             row.CIStatus,
			MergeCommit:          row.MergeCommit,
			WindowsSeconds:       row.WindowsSeconds,
			ThresholdSeconds:     thresholdSeconds,
			ExceedsThreshold:     row.WindowsSeconds > thresholdSeconds,
			OverThresholdSeconds: over,
		})
	}
	return summarizeThresholdEvidenceRows(thresholdRows, thresholdSeconds)
}

func summarizeThresholdEvidenceRows(rows []AtlasPRCIWindowsThresholdRow, thresholdSeconds int) AtlasPRCIWindowsThresholdEvidence {
	evidence := AtlasPRCIWindowsThresholdEvidence{
		ThresholdSeconds: thresholdSeconds,
		RowCount:         len(rows),
		Rows:             append([]AtlasPRCIWindowsThresholdRow(nil), rows...),
	}
	for _, row := range rows {
		if row.ExceedsThreshold {
			evidence.LongRunningWindowsChecks++
		}
		if row.WindowsSeconds > evidence.MaxWindowsSeconds {
			evidence.MaxWindowsSeconds = row.WindowsSeconds
		}
		if row.OverThresholdSeconds > evidence.MaxOverThresholdSeconds {
			evidence.MaxOverThresholdSeconds = row.OverThresholdSeconds
		}
	}
	return evidence
}

func validateWindowsThresholdRows(errs *[]string, rows []AtlasPRCIWindowsThresholdRow, thresholdSeconds int) {
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
		if !oneOf(row.CIStatus, "passed", "failed", "pending") {
			*errs = append(*errs, prefix+".ci_status must be passed, failed, or pending")
		}
		requireField(errs, prefix+".merge_commit", row.MergeCommit)
		if len(row.MergeCommit) != 40 {
			*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if row.WindowsSeconds < 0 || row.ThresholdSeconds <= 0 || row.OverThresholdSeconds < 0 {
			*errs = append(*errs, prefix+".seconds fields must be valid")
		}
		if row.ThresholdSeconds != thresholdSeconds {
			*errs = append(*errs, prefix+".threshold_seconds must match evidence threshold")
		}
		expectedOver := row.WindowsSeconds - thresholdSeconds
		if expectedOver < 0 {
			expectedOver = 0
		}
		if row.OverThresholdSeconds != expectedOver {
			*errs = append(*errs, prefix+".over_threshold_seconds must match windows seconds over threshold")
		}
		if row.ExceedsThreshold != (row.WindowsSeconds > thresholdSeconds) {
			*errs = append(*errs, prefix+".exceeds_threshold must match windows seconds over threshold")
		}
	}
}
