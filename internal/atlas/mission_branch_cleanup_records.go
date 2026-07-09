package atlas

type AtlasBranchCleanupRecordInput struct {
	LocalBranchDeleted           bool `json:"local_branch_deleted"`
	RemoteBranchDeleted          bool `json:"remote_branch_deleted"`
	LocalCodexBranchesRemaining  int  `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int  `json:"remote_codex_branches_remaining"`
}

type AtlasBranchCleanupRecord struct {
	Scope             string `json:"scope"`
	BranchDeleted     bool   `json:"branch_deleted"`
	BranchesRemaining int    `json:"branches_remaining"`
	CleanupComplete   bool   `json:"cleanup_complete"`
}

type AtlasBranchCleanupRecordSummary struct {
	LocalBranchDeletedCount  int  `json:"local_branch_deleted_count"`
	RemoteBranchDeletedCount int  `json:"remote_branch_deleted_count"`
	BranchesRemainingTotal   int  `json:"branches_remaining_total"`
	CleanupComplete          bool `json:"cleanup_complete"`
}

func BuildAtlasBranchCleanupRecords(input AtlasBranchCleanupRecordInput) []AtlasBranchCleanupRecord {
	return []AtlasBranchCleanupRecord{
		buildAtlasBranchCleanupRecord("local", input.LocalBranchDeleted, input.LocalCodexBranchesRemaining),
		buildAtlasBranchCleanupRecord("remote", input.RemoteBranchDeleted, input.RemoteCodexBranchesRemaining),
	}
}

func SummarizeAtlasBranchCleanupRecords(records []AtlasBranchCleanupRecord) AtlasBranchCleanupRecordSummary {
	summary := AtlasBranchCleanupRecordSummary{CleanupComplete: len(records) > 0}
	for _, record := range records {
		if record.Scope == "local" && record.BranchDeleted {
			summary.LocalBranchDeletedCount++
		}
		if record.Scope == "remote" && record.BranchDeleted {
			summary.RemoteBranchDeletedCount++
		}
		summary.BranchesRemainingTotal += record.BranchesRemaining
		if !record.CleanupComplete {
			summary.CleanupComplete = false
		}
	}
	return summary
}

func buildAtlasBranchCleanupRecord(scope string, deleted bool, remaining int) AtlasBranchCleanupRecord {
	return AtlasBranchCleanupRecord{
		Scope:             scope,
		BranchDeleted:     deleted,
		BranchesRemaining: remaining,
		CleanupComplete:   deleted && remaining == 0,
	}
}

func validateAtlasBranchCleanupRecords(errs *[]string, prefix string, records []AtlasBranchCleanupRecord) {
	if len(records) != 2 {
		*errs = append(*errs, prefix+".cleanup_records must include local and remote records")
		return
	}
	for _, record := range records {
		if !oneOf(record.Scope, "local", "remote") {
			*errs = append(*errs, prefix+".cleanup_records scope must be local or remote")
		}
		if record.BranchesRemaining < 0 {
			*errs = append(*errs, prefix+"."+record.Scope+" cleanup record remaining count must be non-negative")
		}
		expectedComplete := record.BranchDeleted && record.BranchesRemaining == 0
		if record.CleanupComplete != expectedComplete {
			*errs = append(*errs, prefix+"."+record.Scope+" cleanup record completeness mismatch")
		}
		if !record.CleanupComplete {
			*errs = append(*errs, prefix+"."+record.Scope+" cleanup record incomplete")
		}
	}
}
