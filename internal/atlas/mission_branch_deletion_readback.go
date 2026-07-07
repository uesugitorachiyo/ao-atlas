package atlas

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

type atlasPostMergeLifecycleEvidence struct {
	Schema                       string `json:"schema"`
	NodeID                       string `json:"node_id"`
	Status                       string `json:"status"`
	PRNumber                     int    `json:"pr_number"`
	MergeCommit                  string `json:"merge_commit"`
	CIStatus                     string `json:"ci_status"`
	LocalMainSynced              bool   `json:"local_main_synced"`
	LocalBranchDeleted           bool   `json:"local_branch_deleted"`
	RemoteBranchDeleted          bool   `json:"remote_branch_deleted"`
	LocalCodexBranchesRemaining  int    `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int    `json:"remote_codex_branches_remaining"`
}

func BuildAtlasPostMergeBranchDeletionReadback(evidenceRoot string) (AtlasPostMergeBranchDeletionReadback, error) {
	root := filepath.Clean(evidenceRoot)
	entries := []AtlasPostMergeBranchDeletionReadbackEntry{}
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
		entries = append(entries, AtlasPostMergeBranchDeletionReadbackEntry{
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
		return AtlasPostMergeBranchDeletionReadback{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	readback := summarizePostMergeBranchDeletionEntries(entries)
	readback.Schema = AtlasPostMergeBranchDeletionReadbackContract
	readback.Status = "branch_deletion_bound"
	readback.EvidenceRoot = publicArtifactRef(evidenceRoot)
	readback.SchedulesWork = false
	readback.ExecutesWork = false
	readback.ApprovesWork = false
	readback.ClaimsAuthorityAdvance = false
	readback.RSIRemainsDenied = true
	if err := ValidateAtlasPostMergeBranchDeletionReadback(readback); err != nil {
		return AtlasPostMergeBranchDeletionReadback{}, err
	}
	return readback, nil
}

func ValidateAtlasPostMergeBranchDeletionReadback(readback AtlasPostMergeBranchDeletionReadback) error {
	var errs []string
	requireContract(&errs, "post_merge_branch_deletion_readback", readback.Schema, AtlasPostMergeBranchDeletionReadbackContract)
	if readback.Status != "branch_deletion_bound" {
		errs = append(errs, "status must be branch_deletion_bound")
	}
	requireField(&errs, "evidence_root", readback.EvidenceRoot)
	checkPublicPath(&errs, "evidence_root", readback.EvidenceRoot, true)
	if len(readback.Entries) == 0 {
		errs = append(errs, "entries must not be empty")
	}
	expected := summarizePostMergeBranchDeletionEntries(readback.Entries)
	if readback.PostMergeLifecycleCount != len(readback.Entries) {
		errs = append(errs, "post_merge_lifecycle_count must match entries length")
	}
	if readback.LocalBranchDeletedCount != expected.LocalBranchDeletedCount {
		errs = append(errs, "local_branch_deleted_count must match entries")
	}
	if readback.RemoteBranchDeletedCount != expected.RemoteBranchDeletedCount {
		errs = append(errs, "remote_branch_deleted_count must match entries")
	}
	if readback.BranchesRemainingTotal != expected.BranchesRemainingTotal {
		errs = append(errs, "branches_remaining_total must match entries")
	}
	previousPath := ""
	seenPaths := map[string]bool{}
	for i, entry := range readback.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		requireField(&errs, prefix+".path", entry.Path)
		checkPublicPath(&errs, prefix+".path", entry.Path, true)
		if seenPaths[entry.Path] {
			errs = append(errs, "entries paths must be unique")
		}
		seenPaths[entry.Path] = true
		if previousPath != "" && entry.Path < previousPath {
			errs = append(errs, "entries must be sorted by path")
		}
		previousPath = entry.Path
		requireField(&errs, prefix+".node_id", entry.NodeID)
		checkPublicPath(&errs, prefix+".node_id", entry.NodeID, true)
		if entry.Status != "merged_and_cleaned" {
			errs = append(errs, prefix+".status must be merged_and_cleaned")
		}
		if entry.PRNumber <= 0 {
			errs = append(errs, prefix+".pr_number must be greater than zero")
		}
		requireField(&errs, prefix+".merge_commit", entry.MergeCommit)
		if len(entry.MergeCommit) != 40 {
			errs = append(errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if entry.CIStatus != "passed" {
			errs = append(errs, prefix+".ci_status must be passed")
		}
		if !entry.LocalBranchDeleted {
			errs = append(errs, prefix+".local_branch_deleted must be true")
		}
		if !entry.RemoteBranchDeleted {
			errs = append(errs, prefix+".remote_branch_deleted must be true")
		}
		if entry.LocalCodexBranchesRemaining < 0 || entry.RemoteCodexBranchesRemaining < 0 {
			errs = append(errs, prefix+".remaining branch counts must be non-negative")
		}
		if !digestPattern.MatchString(entry.Digest) {
			errs = append(errs, prefix+".digest must be sha256 digest")
		}
	}
	validateNoAuthorityEffects(&errs, readback.SchedulesWork, readback.ExecutesWork, readback.ApprovesWork, readback.ClaimsAuthorityAdvance, readback.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizePostMergeBranchDeletionEntries(entries []AtlasPostMergeBranchDeletionReadbackEntry) AtlasPostMergeBranchDeletionReadback {
	readback := AtlasPostMergeBranchDeletionReadback{
		PostMergeLifecycleCount: len(entries),
		Entries:                 append([]AtlasPostMergeBranchDeletionReadbackEntry(nil), entries...),
	}
	for _, entry := range entries {
		if entry.LocalBranchDeleted {
			readback.LocalBranchDeletedCount++
		}
		if entry.RemoteBranchDeleted {
			readback.RemoteBranchDeletedCount++
		}
		readback.BranchesRemainingTotal += entry.LocalCodexBranchesRemaining + entry.RemoteCodexBranchesRemaining
	}
	return readback
}
