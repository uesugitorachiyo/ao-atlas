package atlas

import "strings"

type AtlasMergeReadinessGuardInput struct {
	NodeID      string `json:"node_id"`
	PRNumber    int    `json:"pr_number"`
	MergeCommit string `json:"merge_commit"`
	CIStatus    string `json:"ci_status"`
}

type AtlasMergeReadinessGuard struct {
	Status                            string `json:"status"`
	Reason                            string `json:"reason"`
	NodeID                            string `json:"node_id"`
	PRNumber                          int    `json:"pr_number"`
	MergeCommitBound                  bool   `json:"merge_commit_bound"`
	PassedChecksRequiredBeforeCleanup bool   `json:"passed_checks_required_before_cleanup"`
	RequiredChecksPassed              bool   `json:"required_checks_passed"`
	BranchCleanupEvidenceAllowed      bool   `json:"branch_cleanup_evidence_allowed"`
	SchedulesWork                     bool   `json:"schedules_work"`
	ExecutesWork                      bool   `json:"executes_work"`
	ApprovesWork                      bool   `json:"approves_work"`
	ClaimsAuthorityAdvance            bool   `json:"claims_authority_advance"`
	RSIRemainsDenied                  bool   `json:"rsi_remains_denied"`
}

func EvaluateAtlasMergeReadinessGuard(input AtlasMergeReadinessGuardInput) AtlasMergeReadinessGuard {
	guard := AtlasMergeReadinessGuard{
		Status:                            "blocked",
		NodeID:                            strings.TrimSpace(input.NodeID),
		PRNumber:                          input.PRNumber,
		MergeCommitBound:                  len(strings.TrimSpace(input.MergeCommit)) == 40,
		PassedChecksRequiredBeforeCleanup: true,
		RequiredChecksPassed:              strings.TrimSpace(input.CIStatus) == "passed",
		SchedulesWork:                     false,
		ExecutesWork:                      false,
		ApprovesWork:                      false,
		ClaimsAuthorityAdvance:            false,
		RSIRemainsDenied:                  true,
	}
	switch {
	case guard.NodeID == "":
		guard.Reason = "blocked_node_id_missing"
	case input.PRNumber <= 0:
		guard.Reason = "blocked_pr_number_missing"
	case !guard.MergeCommitBound:
		guard.Reason = "blocked_merge_commit_unbound"
	case !guard.RequiredChecksPassed:
		guard.Reason = "blocked_ci_not_passed"
	default:
		guard.Status = "ready_for_branch_cleanup_evidence"
		guard.Reason = "passed_checks_bound_to_merge_commit"
		guard.BranchCleanupEvidenceAllowed = true
	}
	return guard
}
