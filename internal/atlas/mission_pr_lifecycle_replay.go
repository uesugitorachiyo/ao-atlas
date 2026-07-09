package atlas

import "fmt"

type AtlasPRLifecycleReplayFixture struct {
	Schema                 string                       `json:"schema"`
	Status                 string                       `json:"status"`
	SourceRecommendation   string                       `json:"source_recommendation"`
	CaseCount              int                          `json:"case_count"`
	Cases                  []AtlasPRLifecycleReplayCase `json:"cases"`
	PromotionRequested     bool                         `json:"promotion_requested"`
	PromotionGranted       bool                         `json:"promotion_granted"`
	ClaimsAuthorityAdvance bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                         `json:"rsi_remains_denied"`
}

type AtlasPRLifecycleReplayCase struct {
	CaseID                       string `json:"case_id"`
	PRNumber                     int    `json:"pr_number"`
	PRState                      string `json:"pr_state"`
	CIStatus                     string `json:"ci_status"`
	MergeCommit                  string `json:"merge_commit"`
	LocalMainSynced              bool   `json:"local_main_synced"`
	WorkingTreeClean             bool   `json:"working_tree_clean"`
	LocalCodexBranchesRemaining  int    `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int    `json:"remote_codex_branches_remaining"`
	OperatorAction               string `json:"operator_action"`
	FinalResponseAllowed         bool   `json:"final_response_allowed"`
	SafeToSelectNextNode         bool   `json:"safe_to_select_next_node"`
	ClaimsAuthorityAdvance       bool   `json:"claims_authority_advance"`
	RSIRemainsDenied             bool   `json:"rsi_remains_denied"`
}

func ValidateAtlasPRLifecycleReplayFixture(fixture AtlasPRLifecycleReplayFixture) error {
	var errs []string
	requireContract(&errs, "pr_lifecycle_replay_fixture", fixture.Schema, "ao.atlas.pr-lifecycle-replay-fixture.v0.1")
	if fixture.Status != "guarded" {
		errs = append(errs, "status must be guarded")
	}
	requireField(&errs, "source_recommendation", fixture.SourceRecommendation)
	checkPublicPath(&errs, "source_recommendation", fixture.SourceRecommendation, true)
	if fixture.CaseCount != len(fixture.Cases) {
		errs = append(errs, "case_count must match cases length")
	}
	if len(fixture.Cases) != 3 {
		errs = append(errs, "cases must cover interrupted merge, sync, and cleanup")
	}
	seen := map[string]bool{}
	for i, replayCase := range fixture.Cases {
		validatePRLifecycleReplayCase(&errs, fmt.Sprintf("cases[%d]", i), replayCase)
		if seen[replayCase.CaseID] {
			errs = append(errs, "case_id values must be unique")
		}
		seen[replayCase.CaseID] = true
	}
	for _, required := range []string{"interrupted_merge", "interrupted_sync", "interrupted_cleanup"} {
		if !seen[required] {
			errs = append(errs, "missing replay case "+required)
		}
	}
	if fixture.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if fixture.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if fixture.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !fixture.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func validatePRLifecycleReplayCase(errs *[]string, prefix string, replayCase AtlasPRLifecycleReplayCase) {
	requireField(errs, prefix+".case_id", replayCase.CaseID)
	checkPublicPath(errs, prefix+".case_id", replayCase.CaseID, true)
	if replayCase.PRNumber <= 0 {
		*errs = append(*errs, prefix+".pr_number must be greater than zero")
	}
	if !oneOf(replayCase.PRState, "open", "merged") {
		*errs = append(*errs, prefix+".pr_state must be open or merged")
	}
	if !oneOf(replayCase.CIStatus, "passed", "pending", "failed") {
		*errs = append(*errs, prefix+".ci_status must be passed, pending, or failed")
	}
	if replayCase.CaseID == "interrupted_merge" {
		if replayCase.MergeCommit != "" {
			*errs = append(*errs, prefix+".merge_commit must be empty before merge")
		}
	} else if len(replayCase.MergeCommit) != 40 {
		*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash after merge")
	}
	if replayCase.LocalCodexBranchesRemaining < 0 || replayCase.RemoteCodexBranchesRemaining < 0 {
		*errs = append(*errs, prefix+".remaining branch counts must be non-negative")
	}
	if !oneOf(replayCase.OperatorAction, "merge_after_checks_pass", "sync_main_before_cleanup", "delete_codex_branches_before_next_node") {
		*errs = append(*errs, prefix+".operator_action must be a known lifecycle repair action")
	}
	if replayCase.FinalResponseAllowed {
		*errs = append(*errs, prefix+".final_response_allowed must be false")
	}
	if replayCase.SafeToSelectNextNode {
		*errs = append(*errs, prefix+".safe_to_select_next_node must be false")
	}
	if replayCase.ClaimsAuthorityAdvance {
		*errs = append(*errs, prefix+".claims_authority_advance must be false")
	}
	if !replayCase.RSIRemainsDenied {
		*errs = append(*errs, prefix+".rsi_remains_denied must be true")
	}
}
