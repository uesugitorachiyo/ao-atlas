package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasLocalMainSyncReadback(inputPath string) (AtlasLocalMainSyncReadback, error) {
	input, err := LoadJSON[AtlasLocalMainSyncReadbackInput](inputPath)
	if err != nil {
		return AtlasLocalMainSyncReadback{}, err
	}
	if err := ValidateAtlasLocalMainSyncReadbackInput(input); err != nil {
		return AtlasLocalMainSyncReadback{}, err
	}
	readback := summarizeLocalMainSyncReadback(input)
	readback.Schema = AtlasLocalMainSyncReadbackContract
	readback.Status = "local_main_sync_validated"
	readback.SourceInputPath = publicArtifactRef(inputPath)
	readback.SourceInputDigest = digestValue(input)
	readback.SchedulesWork = false
	readback.ExecutesWork = false
	readback.ApprovesWork = false
	readback.ClaimsAuthorityAdvance = false
	readback.RSIRemainsDenied = true
	if err := ValidateAtlasLocalMainSyncReadback(readback); err != nil {
		return AtlasLocalMainSyncReadback{}, err
	}
	return readback, nil
}

func ValidateAtlasLocalMainSyncReadbackInput(input AtlasLocalMainSyncReadbackInput) error {
	var errs []string
	requireContract(&errs, "local_main_sync_readback_input", input.Schema, AtlasLocalMainSyncReadbackInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	requireField(&errs, "source_readback_path", input.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", input.SourceReadbackPath, true)
	if !digestPattern.MatchString(input.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	validateMainSyncFields(&errs, input.CurrentBranch, input.LocalMainHead, input.OriginMainHead, input.LocalCodexBranches, input.RemoteCodexBranches)
	if input.CompletedNodes <= 0 || input.ReadyNodes < 0 {
		errs = append(errs, "completed_nodes must be positive and ready_nodes must be non-negative")
	}
	requireField(&errs, "first_executable_node", input.FirstExecutableNode)
	checkPublicPath(&errs, "first_executable_node", input.FirstExecutableNode, true)
	if input.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasLocalMainSyncReadback(readback AtlasLocalMainSyncReadback) error {
	var errs []string
	requireContract(&errs, "local_main_sync_readback", readback.Schema, AtlasLocalMainSyncReadbackContract)
	if readback.Status != "local_main_sync_validated" {
		errs = append(errs, "status must be local_main_sync_validated")
	}
	requireField(&errs, "source_input_path", readback.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", readback.SourceInputPath, true)
	if !digestPattern.MatchString(readback.SourceInputDigest) {
		errs = append(errs, "source_input_digest must be sha256 digest")
	}
	requireField(&errs, "source_readback_path", readback.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", readback.SourceReadbackPath, true)
	if !digestPattern.MatchString(readback.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	validateMainSyncFields(&errs, readback.CurrentBranch, readback.LocalMainHead, readback.OriginMainHead, nil, nil)
	if readback.LocalMainSynced != (readback.CurrentBranch == "main" && readback.LocalMainHead == readback.OriginMainHead) {
		errs = append(errs, "local_main_synced must reflect main branch and matching heads")
	}
	if readback.CodexBranchCleanupConfirmed != true {
		errs = append(errs, "codex_branch_cleanup_confirmed must be true")
	}
	expectedSafe := readback.LocalMainSynced && readback.WorkingTreeClean && readback.CodexBranchCleanupConfirmed && !readback.FinalResponseAllowed
	if readback.SafeToSelectNextNode != expectedSafe {
		errs = append(errs, "safe_to_select_next_node must match sync, cleanliness, cleanup, and return gate state")
	}
	if readback.CompletedNodes <= 0 || readback.ReadyNodes < 0 {
		errs = append(errs, "completed_nodes must be positive and ready_nodes must be non-negative")
	}
	requireField(&errs, "first_executable_node", readback.FirstExecutableNode)
	checkPublicPath(&errs, "first_executable_node", readback.FirstExecutableNode, true)
	if readback.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	if readback.DenialCaseCount != len(readback.DenialCases) {
		errs = append(errs, "denial_case_count must match denial_cases length")
	}
	validateLocalMainSyncDenialCases(&errs, readback.DenialCases)
	validateNoAuthorityEffects(&errs, readback.SchedulesWork, readback.ExecutesWork, readback.ApprovesWork, readback.ClaimsAuthorityAdvance, readback.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeLocalMainSyncReadback(input AtlasLocalMainSyncReadbackInput) AtlasLocalMainSyncReadback {
	localMainSynced := input.CurrentBranch == "main" && input.LocalMainHead == input.OriginMainHead
	cleanupConfirmed := len(input.LocalCodexBranches) == 0 && len(input.RemoteCodexBranches) == 0
	return AtlasLocalMainSyncReadback{
		SourceReadbackPath:          input.SourceReadbackPath,
		SourceReadbackDigest:        input.SourceReadbackDigest,
		CurrentBranch:               input.CurrentBranch,
		LocalMainHead:               input.LocalMainHead,
		OriginMainHead:              input.OriginMainHead,
		LocalMainSynced:             localMainSynced,
		WorkingTreeClean:            input.WorkingTreeClean,
		CodexBranchCleanupConfirmed: cleanupConfirmed,
		SafeToSelectNextNode:        localMainSynced && input.WorkingTreeClean && cleanupConfirmed && !input.FinalResponseAllowed,
		CompletedNodes:              input.CompletedNodes,
		ReadyNodes:                  input.ReadyNodes,
		FirstExecutableNode:         input.FirstExecutableNode,
		FinalResponseAllowed:        input.FinalResponseAllowed,
		DenialCaseCount:             len(localMainSyncDenialCases()),
		DenialCases:                 localMainSyncDenialCases(),
	}
}

func localMainSyncDenialCases() []AtlasLocalMainSyncDenialCase {
	return []AtlasLocalMainSyncDenialCase{
		{
			Name:                 "stale_main",
			LocalMainStale:       true,
			SafeToSelectNextNode: false,
			Reason:               "blocked_local_main_not_synced",
		},
		{
			Name:                 "dirty_worktree",
			WorkingTreeDirty:     true,
			SafeToSelectNextNode: false,
			Reason:               "blocked_worktree_not_clean",
		},
		{
			Name:                 "codex_branch_remaining",
			CodexBranchRemaining: true,
			SafeToSelectNextNode: false,
			Reason:               "blocked_codex_branch_cleanup_incomplete",
		},
	}
}

func validateMainSyncFields(errs *[]string, currentBranch, localHead, originHead string, localCodexBranches, remoteCodexBranches []string) {
	requireField(errs, "current_branch", currentBranch)
	checkPublicPath(errs, "current_branch", currentBranch, true)
	if currentBranch != "main" {
		*errs = append(*errs, "current_branch must be main")
	}
	requireField(errs, "local_main_head", localHead)
	requireField(errs, "origin_main_head", originHead)
	if len(localHead) != 40 || len(originHead) != 40 {
		*errs = append(*errs, "main head fields must be 40 character commit hashes")
	}
	for _, branch := range localCodexBranches {
		if strings.TrimSpace(branch) == "" || !strings.HasPrefix(branch, "codex/") {
			*errs = append(*errs, "local_codex_branches entries must be codex branches")
		}
		checkPublicPath(errs, "local_codex_branches", branch, true)
	}
	for _, branch := range remoteCodexBranches {
		if strings.TrimSpace(branch) == "" || !strings.HasPrefix(branch, "codex/") {
			*errs = append(*errs, "remote_codex_branches entries must be codex branches")
		}
		checkPublicPath(errs, "remote_codex_branches", branch, true)
	}
}

func validateLocalMainSyncDenialCases(errs *[]string, cases []AtlasLocalMainSyncDenialCase) {
	want := map[string]string{
		"stale_main":             "blocked_local_main_not_synced",
		"dirty_worktree":         "blocked_worktree_not_clean",
		"codex_branch_remaining": "blocked_codex_branch_cleanup_incomplete",
	}
	seen := map[string]bool{}
	for i, item := range cases {
		prefix := fmt.Sprintf("denial_cases[%d]", i)
		requireField(errs, prefix+".name", item.Name)
		if seen[item.Name] {
			*errs = append(*errs, "denial_cases names must be unique")
		}
		seen[item.Name] = true
		reason, ok := want[item.Name]
		if !ok {
			*errs = append(*errs, prefix+".name is not an expected denial case")
		}
		if item.Reason != reason {
			*errs = append(*errs, prefix+".reason must match expected denial reason")
		}
		if item.SafeToSelectNextNode {
			*errs = append(*errs, prefix+".safe_to_select_next_node must be false")
		}
		switch item.Name {
		case "stale_main":
			if !item.LocalMainStale || item.WorkingTreeDirty || item.CodexBranchRemaining {
				*errs = append(*errs, prefix+" must mark only local_main_stale")
			}
		case "dirty_worktree":
			if item.LocalMainStale || !item.WorkingTreeDirty || item.CodexBranchRemaining {
				*errs = append(*errs, prefix+" must mark only working_tree_dirty")
			}
		case "codex_branch_remaining":
			if item.LocalMainStale || item.WorkingTreeDirty || !item.CodexBranchRemaining {
				*errs = append(*errs, prefix+" must mark only codex_branch_remaining")
			}
		}
	}
	for name := range want {
		if !seen[name] {
			*errs = append(*errs, "denial_cases missing "+name)
		}
	}
}
