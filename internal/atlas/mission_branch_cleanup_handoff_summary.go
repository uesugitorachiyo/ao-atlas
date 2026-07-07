package atlas

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

func BuildAtlasBranchCleanupHandoffSummary(evidenceRoot, sourceReadbackPath string) (AtlasBranchCleanupHandoffSummary, error) {
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasBranchCleanupHandoffSummary{}, err
	}
	entries, err := collectBranchCleanupHandoffEntries(evidenceRoot)
	if err != nil {
		return AtlasBranchCleanupHandoffSummary{}, err
	}
	summary := summarizeBranchCleanupHandoffEntries(readback, entries)
	summary.Schema = AtlasBranchCleanupHandoffSummaryContract
	summary.Status = "branch_cleanup_handoff_summarized"
	summary.EvidenceRoot = publicArtifactRef(evidenceRoot)
	summary.SourceReadbackPath = publicArtifactRef(sourceReadbackPath)
	summary.SourceReadbackDigest = digestValue(readback)
	summary.SchedulesWork = false
	summary.ExecutesWork = false
	summary.ApprovesWork = false
	summary.ClaimsAuthorityAdvance = false
	summary.RSIRemainsDenied = true
	if err := ValidateAtlasBranchCleanupHandoffSummary(summary); err != nil {
		return AtlasBranchCleanupHandoffSummary{}, err
	}
	return summary, nil
}

func collectBranchCleanupHandoffEntries(evidenceRoot string) ([]AtlasBranchCleanupHandoffEntry, error) {
	root := filepath.Clean(evidenceRoot)
	entries := []AtlasBranchCleanupHandoffEntry{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Base(path) != "post-merge-lifecycle.json" {
			return nil
		}
		lifecycle, err := LoadJSON[atlasPostMergeLifecycleEvidence](path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries = append(entries, AtlasBranchCleanupHandoffEntry{
			Path:                         filepath.ToSlash(rel),
			NodeID:                       lifecycle.NodeID,
			Status:                       lifecycle.Status,
			PRNumber:                     lifecycle.PRNumber,
			MergeCommit:                  lifecycle.MergeCommit,
			CIStatus:                     lifecycle.CIStatus,
			LocalBranchDeleted:           lifecycle.LocalBranchDeleted,
			RemoteBranchDeleted:          lifecycle.RemoteBranchDeleted,
			LocalCodexBranchesRemaining:  lifecycle.LocalCodexBranchesRemaining,
			RemoteCodexBranchesRemaining: lifecycle.RemoteCodexBranchesRemaining,
			Digest:                       digestValue(lifecycle),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func ValidateAtlasBranchCleanupHandoffSummary(summary AtlasBranchCleanupHandoffSummary) error {
	var errs []string
	requireContract(&errs, "branch_cleanup_handoff_summary", summary.Schema, AtlasBranchCleanupHandoffSummaryContract)
	if summary.Status != "branch_cleanup_handoff_summarized" {
		errs = append(errs, "status must be branch_cleanup_handoff_summarized")
	}
	requireField(&errs, "evidence_root", summary.EvidenceRoot)
	checkPublicPath(&errs, "evidence_root", summary.EvidenceRoot, true)
	requireField(&errs, "source_readback_path", summary.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", summary.SourceReadbackPath, true)
	if !digestPattern.MatchString(summary.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if summary.CompletedNodes <= 0 || summary.ReadyNodes < 0 || summary.TotalNodes < summary.CompletedNodes {
		errs = append(errs, "node counts must be positive and internally consistent")
	}
	requireField(&errs, "first_executable_node", summary.FirstExecutableNode)
	checkPublicPath(&errs, "first_executable_node", summary.FirstExecutableNode, true)
	requireField(&errs, "exact_next_action", summary.ExactNextAction)
	checkPublicStrings(&errs, "exact_next_action", []string{summary.ExactNextAction}, true)
	if summary.ReadyNodes > 0 && summary.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready nodes remain")
	}
	expected := summarizeBranchCleanupEntriesOnly(summary.Entries)
	if summary.PostMergeLifecycleCount != len(summary.Entries) {
		errs = append(errs, "post_merge_lifecycle_count must match entries length")
	}
	if summary.PostMergeLifecycleCount != summary.CompletedNodes {
		errs = append(errs, "post_merge_lifecycle_count must match completed_nodes")
	}
	if summary.MergedAndCleanedCount != expected.MergedAndCleanedCount {
		errs = append(errs, "merged_and_cleaned_count must match entries")
	}
	if summary.PassedCICount != expected.PassedCICount {
		errs = append(errs, "passed_ci_count must match entries")
	}
	if summary.LocalBranchDeletedCount != expected.LocalBranchDeletedCount {
		errs = append(errs, "local_branch_deleted_count must match entries")
	}
	if summary.RemoteBranchDeletedCount != expected.RemoteBranchDeletedCount {
		errs = append(errs, "remote_branch_deleted_count must match entries")
	}
	if summary.BranchesRemainingTotal != expected.BranchesRemainingTotal {
		errs = append(errs, "branches_remaining_total must match entries")
	}
	expectedComplete := summary.PostMergeLifecycleCount == summary.CompletedNodes &&
		summary.MergedAndCleanedCount == summary.CompletedNodes &&
		summary.PassedCICount == summary.CompletedNodes &&
		summary.LocalBranchDeletedCount == summary.CompletedNodes &&
		summary.RemoteBranchDeletedCount == summary.CompletedNodes &&
		summary.BranchesRemainingTotal == 0
	if summary.CleanupComplete != expectedComplete {
		errs = append(errs, "cleanup_complete must reflect completed node cleanup counts")
	}
	expectedStatus := "cleanup_ledger_blocked"
	if expectedComplete {
		expectedStatus = "cleanup_ledger_ready"
	}
	if summary.OperatorHandoffStatus != expectedStatus {
		errs = append(errs, "operator_handoff_status must reflect cleanup completeness")
	}
	validateBranchCleanupHandoffEntries(&errs, summary.Entries)
	validateNoAuthorityEffects(&errs, summary.SchedulesWork, summary.ExecutesWork, summary.ApprovesWork, summary.ClaimsAuthorityAdvance, summary.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeBranchCleanupHandoffEntries(readback AtlasRecommendationReadback, entries []AtlasBranchCleanupHandoffEntry) AtlasBranchCleanupHandoffSummary {
	summary := summarizeBranchCleanupEntriesOnly(entries)
	summary.CompletedNodes = readback.CompletedNodes
	summary.ReadyNodes = readback.ReadyNodes
	summary.TotalNodes = readback.TotalNodes
	summary.FirstExecutableNode = readback.FirstExecutableNode
	summary.FinalResponseAllowed = readback.FinalResponseAllowed
	summary.ExactNextAction = readback.ExactNextAction
	summary.CleanupComplete = summary.PostMergeLifecycleCount == summary.CompletedNodes &&
		summary.MergedAndCleanedCount == summary.CompletedNodes &&
		summary.PassedCICount == summary.CompletedNodes &&
		summary.LocalBranchDeletedCount == summary.CompletedNodes &&
		summary.RemoteBranchDeletedCount == summary.CompletedNodes &&
		summary.BranchesRemainingTotal == 0
	summary.OperatorHandoffStatus = "cleanup_ledger_blocked"
	if summary.CleanupComplete {
		summary.OperatorHandoffStatus = "cleanup_ledger_ready"
	}
	return summary
}

func summarizeBranchCleanupEntriesOnly(entries []AtlasBranchCleanupHandoffEntry) AtlasBranchCleanupHandoffSummary {
	summary := AtlasBranchCleanupHandoffSummary{
		PostMergeLifecycleCount: len(entries),
		Entries:                 append([]AtlasBranchCleanupHandoffEntry(nil), entries...),
	}
	for _, entry := range entries {
		if entry.Status == "merged_and_cleaned" {
			summary.MergedAndCleanedCount++
		}
		if entry.CIStatus == "passed" {
			summary.PassedCICount++
		}
		if entry.LocalBranchDeleted {
			summary.LocalBranchDeletedCount++
		}
		if entry.RemoteBranchDeleted {
			summary.RemoteBranchDeletedCount++
		}
		summary.BranchesRemainingTotal += entry.LocalCodexBranchesRemaining + entry.RemoteCodexBranchesRemaining
	}
	return summary
}

func validateBranchCleanupHandoffEntries(errs *[]string, entries []AtlasBranchCleanupHandoffEntry) {
	if len(entries) == 0 {
		*errs = append(*errs, "entries must not be empty")
	}
	previousPath := ""
	seenPaths := map[string]bool{}
	for i, entry := range entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		requireField(errs, prefix+".path", entry.Path)
		checkPublicPath(errs, prefix+".path", entry.Path, true)
		if seenPaths[entry.Path] {
			*errs = append(*errs, "entries paths must be unique")
		}
		seenPaths[entry.Path] = true
		if previousPath != "" && entry.Path < previousPath {
			*errs = append(*errs, "entries must be sorted by path")
		}
		previousPath = entry.Path
		requireField(errs, prefix+".node_id", entry.NodeID)
		checkPublicPath(errs, prefix+".node_id", entry.NodeID, true)
		if entry.Status != "merged_and_cleaned" {
			*errs = append(*errs, prefix+".status must be merged_and_cleaned")
		}
		if entry.PRNumber <= 0 {
			*errs = append(*errs, prefix+".pr_number must be greater than zero")
		}
		requireField(errs, prefix+".merge_commit", entry.MergeCommit)
		if len(entry.MergeCommit) != 40 {
			*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if entry.CIStatus != "passed" {
			*errs = append(*errs, prefix+".ci_status must be passed")
		}
		if !entry.LocalBranchDeleted {
			*errs = append(*errs, prefix+".local_branch_deleted must be true")
		}
		if !entry.RemoteBranchDeleted {
			*errs = append(*errs, prefix+".remote_branch_deleted must be true")
		}
		if entry.LocalCodexBranchesRemaining != 0 || entry.RemoteCodexBranchesRemaining != 0 {
			*errs = append(*errs, prefix+".remaining branch counts must be zero")
		}
		if !digestPattern.MatchString(entry.Digest) {
			*errs = append(*errs, prefix+".digest must be sha256 digest")
		}
	}
}
