package atlas

import "fmt"

type AtlasP0BPRCILedger struct {
	Schema                  string                  `json:"schema"`
	NodeID                  string                  `json:"node_id"`
	Status                  string                  `json:"status"`
	Source                  string                  `json:"source"`
	CoveredNodeStart        int                     `json:"covered_node_start"`
	CoveredNodeEnd          int                     `json:"covered_node_end"`
	EntryCount              int                     `json:"entry_count"`
	AllPRsMerged            bool                    `json:"all_prs_merged"`
	AllCIStatusesPassed     bool                    `json:"all_ci_statuses_passed"`
	AllMergeCommitsRecorded bool                    `json:"all_merge_commits_recorded"`
	AllBranchesDeleted      bool                    `json:"all_branches_deleted"`
	BranchesRemainingTotal  int                     `json:"branches_remaining_total"`
	CompletedNodesBefore    int                     `json:"completed_nodes_before"`
	ReadyNodesBefore        int                     `json:"ready_nodes_before"`
	FinalResponseAllowed    bool                    `json:"final_response_allowed"`
	Entries                 []AtlasP0BPRCILedgerRow `json:"entries"`
	PromotionRequested      bool                    `json:"promotion_requested"`
	PromotionGranted        bool                    `json:"promotion_granted"`
	ClaimsAuthorityAdvance  bool                    `json:"claims_authority_advance"`
	RSIRemainsDenied        bool                    `json:"rsi_remains_denied"`
	SchedulesWork           bool                    `json:"schedules_work"`
	ExecutesWork            bool                    `json:"executes_work"`
	ApprovesWork            bool                    `json:"approves_work"`
}

type AtlasP0BPRCILedgerRow struct {
	NodeNumber                   int    `json:"node_number"`
	NodeID                       string `json:"node_id"`
	PRNumber                     int    `json:"pr_number"`
	PRURL                        string `json:"pr_url"`
	Title                        string `json:"title"`
	HeadRef                      string `json:"head_ref"`
	State                        string `json:"state"`
	MergedAt                     string `json:"merged_at"`
	MergeCommit                  string `json:"merge_commit"`
	CIStatus                     string `json:"ci_status"`
	LocalMainSynced              bool   `json:"local_main_synced"`
	LocalBranchDeleted           bool   `json:"local_branch_deleted"`
	RemoteBranchDeleted          bool   `json:"remote_branch_deleted"`
	LocalCodexBranchesRemaining  int    `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int    `json:"remote_codex_branches_remaining"`
}

func ValidateAtlasP0BPRCILedger(ledger AtlasP0BPRCILedger) error {
	var errs []string
	requireContract(&errs, "p0b_pr_ci_ledger", ledger.Schema, AtlasP0BPRCILedgerContract)
	if ledger.Status != "complete" {
		errs = append(errs, "status must be complete")
	}
	requireField(&errs, "node_id", ledger.NodeID)
	checkPublicPath(&errs, "node_id", ledger.NodeID, true)
	requireField(&errs, "source", ledger.Source)
	checkPublicPath(&errs, "source", ledger.Source, true)
	if ledger.CoveredNodeStart <= 0 || ledger.CoveredNodeEnd < ledger.CoveredNodeStart {
		errs = append(errs, "covered node range must be positive and ordered")
	}
	expectedCount := ledger.CoveredNodeEnd - ledger.CoveredNodeStart + 1
	if ledger.EntryCount != len(ledger.Entries) || ledger.EntryCount != expectedCount {
		errs = append(errs, "entry_count must match covered node range and entries length")
	}
	if ledger.BranchesRemainingTotal < 0 || ledger.CompletedNodesBefore < 0 || ledger.ReadyNodesBefore < 0 {
		errs = append(errs, "summary counts must be non-negative")
	}
	if ledger.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while P0-B work remains")
	}
	seenNodes := map[int]bool{}
	branchesRemaining := 0
	allMerged := len(ledger.Entries) > 0
	allCIPassed := len(ledger.Entries) > 0
	allMergeHeads := len(ledger.Entries) > 0
	allBranchesDeleted := len(ledger.Entries) > 0
	for i, entry := range ledger.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		validateP0BPRCILedgerRow(&errs, prefix, entry)
		if entry.NodeNumber < ledger.CoveredNodeStart || entry.NodeNumber > ledger.CoveredNodeEnd {
			errs = append(errs, prefix+".node_number must be inside covered range")
		}
		if seenNodes[entry.NodeNumber] {
			errs = append(errs, prefix+".node_number must be unique")
		}
		seenNodes[entry.NodeNumber] = true
		allMerged = allMerged && entry.State == "MERGED"
		allCIPassed = allCIPassed && entry.CIStatus == "passed"
		allMergeHeads = allMergeHeads && len(entry.MergeCommit) == 40
		allBranchesDeleted = allBranchesDeleted && entry.LocalBranchDeleted && entry.RemoteBranchDeleted
		branchesRemaining += entry.LocalCodexBranchesRemaining + entry.RemoteCodexBranchesRemaining
	}
	if ledger.AllPRsMerged != allMerged {
		errs = append(errs, "all_prs_merged must match entries")
	}
	if ledger.AllCIStatusesPassed != allCIPassed {
		errs = append(errs, "all_ci_statuses_passed must match entries")
	}
	if ledger.AllMergeCommitsRecorded != allMergeHeads {
		errs = append(errs, "all_merge_commits_recorded must match entries")
	}
	if ledger.AllBranchesDeleted != allBranchesDeleted {
		errs = append(errs, "all_branches_deleted must match entries")
	}
	if ledger.BranchesRemainingTotal != branchesRemaining {
		errs = append(errs, "branches_remaining_total must match entries")
	}
	if ledger.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if ledger.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, ledger.SchedulesWork, ledger.ExecutesWork, ledger.ApprovesWork, ledger.ClaimsAuthorityAdvance, ledger.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateP0BPRCILedgerRow(errs *[]string, prefix string, entry AtlasP0BPRCILedgerRow) {
	requireField(errs, prefix+".node_id", entry.NodeID)
	checkPublicPath(errs, prefix+".node_id", entry.NodeID, true)
	if entry.PRNumber <= 0 {
		*errs = append(*errs, prefix+".pr_number must be greater than zero")
	}
	requireField(errs, prefix+".pr_url", entry.PRURL)
	checkPublicPath(errs, prefix+".pr_url", entry.PRURL, false)
	requireField(errs, prefix+".title", entry.Title)
	checkPublicPath(errs, prefix+".title", entry.Title, false)
	requireField(errs, prefix+".head_ref", entry.HeadRef)
	checkPublicPath(errs, prefix+".head_ref", entry.HeadRef, true)
	if entry.State != "MERGED" {
		*errs = append(*errs, prefix+".state must be MERGED")
	}
	requireField(errs, prefix+".merged_at", entry.MergedAt)
	if len(entry.MergeCommit) != 40 {
		*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
	}
	if entry.CIStatus != "passed" {
		*errs = append(*errs, prefix+".ci_status must be passed")
	}
	if !entry.LocalMainSynced {
		*errs = append(*errs, prefix+".local_main_synced must be true")
	}
	if !entry.LocalBranchDeleted {
		*errs = append(*errs, prefix+".local_branch_deleted must be true")
	}
	if !entry.RemoteBranchDeleted {
		*errs = append(*errs, prefix+".remote_branch_deleted must be true")
	}
	if entry.LocalCodexBranchesRemaining < 0 || entry.RemoteCodexBranchesRemaining < 0 {
		*errs = append(*errs, prefix+".remaining branch counts must be non-negative")
	}
}
