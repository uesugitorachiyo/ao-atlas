package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasStaleRemoteBranchRepair(inputPath string) (AtlasStaleRemoteBranchRepair, error) {
	input, err := LoadJSON[AtlasStaleRemoteBranchRepairInput](inputPath)
	if err != nil {
		return AtlasStaleRemoteBranchRepair{}, err
	}
	if err := ValidateAtlasStaleRemoteBranchRepairInput(input); err != nil {
		return AtlasStaleRemoteBranchRepair{}, err
	}
	repair := summarizeStaleRemoteBranchRepairCases(input.Cases)
	repair.Schema = AtlasStaleRemoteBranchRepairContract
	repair.Status = "remote_branch_repair_matrix_recorded"
	repair.SourceInputPath = publicArtifactRef(inputPath)
	repair.SourceInputDigest = digestValue(input)
	repair.SourceBranchDeletionReadbackPath = input.SourceBranchDeletionReadbackPath
	repair.SourceBranchDeletionReadbackDigest = input.SourceBranchDeletionReadbackDigest
	repair.SchedulesWork = false
	repair.ExecutesWork = false
	repair.ApprovesWork = false
	repair.ClaimsAuthorityAdvance = false
	repair.RSIRemainsDenied = true
	if err := ValidateAtlasStaleRemoteBranchRepair(repair); err != nil {
		return AtlasStaleRemoteBranchRepair{}, err
	}
	return repair, nil
}

func ValidateAtlasStaleRemoteBranchRepairInput(input AtlasStaleRemoteBranchRepairInput) error {
	var errs []string
	requireContract(&errs, "stale_remote_branch_repair_input", input.Schema, AtlasStaleRemoteBranchRepairInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	requireField(&errs, "source_branch_deletion_readback_path", input.SourceBranchDeletionReadbackPath)
	checkPublicPath(&errs, "source_branch_deletion_readback_path", input.SourceBranchDeletionReadbackPath, true)
	if !digestPattern.MatchString(input.SourceBranchDeletionReadbackDigest) {
		errs = append(errs, "source_branch_deletion_readback_digest must be sha256 digest")
	}
	if len(input.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	validateStaleRemoteBranchRepairCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasStaleRemoteBranchRepair(repair AtlasStaleRemoteBranchRepair) error {
	var errs []string
	requireContract(&errs, "stale_remote_branch_repair", repair.Schema, AtlasStaleRemoteBranchRepairContract)
	if repair.Status != "remote_branch_repair_matrix_recorded" {
		errs = append(errs, "status must be remote_branch_repair_matrix_recorded")
	}
	requireField(&errs, "source_input_path", repair.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", repair.SourceInputPath, true)
	if !digestPattern.MatchString(repair.SourceInputDigest) {
		errs = append(errs, "source_input_digest must be sha256 digest")
	}
	requireField(&errs, "source_branch_deletion_readback_path", repair.SourceBranchDeletionReadbackPath)
	checkPublicPath(&errs, "source_branch_deletion_readback_path", repair.SourceBranchDeletionReadbackPath, true)
	if !digestPattern.MatchString(repair.SourceBranchDeletionReadbackDigest) {
		errs = append(errs, "source_branch_deletion_readback_digest must be sha256 digest")
	}
	if len(repair.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	expected := summarizeStaleRemoteBranchRepairDecisions(repair.Cases)
	if repair.CaseCount != len(repair.Cases) {
		errs = append(errs, "case_count must match cases length")
	}
	if repair.RepairRequiredCases != expected.RepairRequiredCases {
		errs = append(errs, "repair_required_cases must match cases")
	}
	if repair.CleanupSafeCases != expected.CleanupSafeCases {
		errs = append(errs, "cleanup_safe_cases must match cases")
	}
	if repair.BlockedCases != expected.BlockedCases {
		errs = append(errs, "blocked_cases must match cases")
	}
	previousID := ""
	seenIDs := map[string]bool{}
	for i, decision := range repair.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		validateStaleRemoteBranchRepairDecision(&errs, prefix, decision)
		if seenIDs[decision.ID] {
			errs = append(errs, "cases ids must be unique")
		}
		seenIDs[decision.ID] = true
		if previousID != "" && decision.ID < previousID {
			errs = append(errs, "cases must be sorted by id")
		}
		previousID = decision.ID
	}
	validateNoAuthorityEffects(&errs, repair.SchedulesWork, repair.ExecutesWork, repair.ApprovesWork, repair.ClaimsAuthorityAdvance, repair.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeStaleRemoteBranchRepairCases(cases []AtlasStaleRemoteBranchRepairCase) AtlasStaleRemoteBranchRepair {
	decisions := make([]AtlasStaleRemoteBranchRepairDecision, 0, len(cases))
	for _, item := range cases {
		decision := AtlasStaleRemoteBranchRepairDecision{
			ID:                           item.ID,
			NodeID:                       item.NodeID,
			PRNumber:                     item.PRNumber,
			MergeCommit:                  item.MergeCommit,
			HeadBranch:                   item.HeadBranch,
			HandoffStatus:                item.HandoffStatus,
			RemoteBranchDeleted:          item.RemoteBranchDeleted,
			RemoteCodexBranchesRemaining: item.RemoteCodexBranchesRemaining,
		}
		staleRemote := !item.RemoteBranchDeleted || item.RemoteCodexBranchesRemaining > 0
		if staleRemote {
			decision.RepairRequired = true
			decision.SafeToRepair = true
			decision.RepairAction = "delete_remote_codex_branch"
			decision.RepairCommand = "git push origin --delete " + item.HeadBranch
			decision.BlocksNextNode = true
			decision.Reason = "remote_codex_branch_remains_after_merge_cleanup_handoff"
		} else {
			decision.RepairAction = "no_repair_required"
			decision.Reason = "remote_branch_cleanup_already_bound"
		}
		decisions = append(decisions, decision)
	}
	return summarizeStaleRemoteBranchRepairDecisions(decisions)
}

func summarizeStaleRemoteBranchRepairDecisions(decisions []AtlasStaleRemoteBranchRepairDecision) AtlasStaleRemoteBranchRepair {
	repair := AtlasStaleRemoteBranchRepair{
		CaseCount: len(decisions),
		Cases:     append([]AtlasStaleRemoteBranchRepairDecision(nil), decisions...),
	}
	for _, decision := range decisions {
		if decision.RepairRequired {
			repair.RepairRequiredCases++
		}
		if decision.SafeToRepair {
			repair.CleanupSafeCases++
		}
		if decision.RepairRequired && !decision.SafeToRepair {
			repair.BlockedCases++
		}
	}
	return repair
}

func validateStaleRemoteBranchRepairCases(errs *[]string, cases []AtlasStaleRemoteBranchRepairCase) {
	previousID := ""
	seenIDs := map[string]bool{}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		checkPublicPath(errs, prefix+".id", item.ID, true)
		if seenIDs[item.ID] {
			*errs = append(*errs, "cases ids must be unique")
		}
		seenIDs[item.ID] = true
		if previousID != "" && item.ID < previousID {
			*errs = append(*errs, "cases must be sorted by id")
		}
		previousID = item.ID
		requireField(errs, prefix+".node_id", item.NodeID)
		checkPublicPath(errs, prefix+".node_id", item.NodeID, true)
		if item.PRNumber <= 0 {
			*errs = append(*errs, prefix+".pr_number must be greater than zero")
		}
		requireField(errs, prefix+".merge_commit", item.MergeCommit)
		if len(item.MergeCommit) != 40 {
			*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		requireField(errs, prefix+".head_branch", item.HeadBranch)
		checkPublicPath(errs, prefix+".head_branch", item.HeadBranch, true)
		if !strings.HasPrefix(item.HeadBranch, "codex/") {
			*errs = append(*errs, prefix+".head_branch must be a codex branch")
		}
		if !oneOf(item.HandoffStatus, "completed_clean", "interrupted_after_merge", "remote_delete_failed") {
			*errs = append(*errs, prefix+".handoff_status must be completed_clean, interrupted_after_merge, or remote_delete_failed")
		}
		if item.RemoteCodexBranchesRemaining < 0 {
			*errs = append(*errs, prefix+".remote_codex_branches_remaining must be non-negative")
		}
		if item.HandoffStatus == "completed_clean" && (!item.RemoteBranchDeleted || item.RemoteCodexBranchesRemaining != 0) {
			*errs = append(*errs, prefix+".completed_clean handoff must have remote branch deleted and no remaining remote codex branches")
		}
	}
}

func validateStaleRemoteBranchRepairDecision(errs *[]string, prefix string, decision AtlasStaleRemoteBranchRepairDecision) {
	requireField(errs, prefix+".id", decision.ID)
	checkPublicPath(errs, prefix+".id", decision.ID, true)
	requireField(errs, prefix+".node_id", decision.NodeID)
	checkPublicPath(errs, prefix+".node_id", decision.NodeID, true)
	if decision.PRNumber <= 0 {
		*errs = append(*errs, prefix+".pr_number must be greater than zero")
	}
	requireField(errs, prefix+".merge_commit", decision.MergeCommit)
	if len(decision.MergeCommit) != 40 {
		*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
	}
	requireField(errs, prefix+".head_branch", decision.HeadBranch)
	checkPublicPath(errs, prefix+".head_branch", decision.HeadBranch, true)
	if !strings.HasPrefix(decision.HeadBranch, "codex/") {
		*errs = append(*errs, prefix+".head_branch must be a codex branch")
	}
	if !oneOf(decision.HandoffStatus, "completed_clean", "interrupted_after_merge", "remote_delete_failed") {
		*errs = append(*errs, prefix+".handoff_status must be completed_clean, interrupted_after_merge, or remote_delete_failed")
	}
	if decision.RemoteCodexBranchesRemaining < 0 {
		*errs = append(*errs, prefix+".remote_codex_branches_remaining must be non-negative")
	}
	staleRemote := !decision.RemoteBranchDeleted || decision.RemoteCodexBranchesRemaining > 0
	if decision.RepairRequired != staleRemote {
		*errs = append(*errs, prefix+".repair_required must match remote branch state")
	}
	if staleRemote {
		if !decision.SafeToRepair {
			*errs = append(*errs, prefix+".safe_to_repair must be true for codex remote branch cleanup")
		}
		if decision.RepairAction != "delete_remote_codex_branch" {
			*errs = append(*errs, prefix+".repair_action must be delete_remote_codex_branch")
		}
		wantCommand := "git push origin --delete " + decision.HeadBranch
		if decision.RepairCommand != wantCommand {
			*errs = append(*errs, prefix+".repair_command must name remote codex branch deletion")
		}
		if !decision.BlocksNextNode {
			*errs = append(*errs, prefix+".blocks_next_node must be true until cleanup is complete")
		}
		if decision.Reason != "remote_codex_branch_remains_after_merge_cleanup_handoff" {
			*errs = append(*errs, prefix+".reason must explain stale remote branch cleanup")
		}
		return
	}
	if decision.SafeToRepair {
		*errs = append(*errs, prefix+".safe_to_repair must be false when no repair is required")
	}
	if decision.RepairAction != "no_repair_required" {
		*errs = append(*errs, prefix+".repair_action must be no_repair_required")
	}
	if strings.TrimSpace(decision.RepairCommand) != "" {
		*errs = append(*errs, prefix+".repair_command must be empty when no repair is required")
	}
	if decision.BlocksNextNode {
		*errs = append(*errs, prefix+".blocks_next_node must be false when no repair is required")
	}
	if decision.Reason != "remote_branch_cleanup_already_bound" {
		*errs = append(*errs, prefix+".reason must explain clean remote branch state")
	}
}
