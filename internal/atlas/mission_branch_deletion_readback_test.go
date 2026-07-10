package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveBranchDeletionReadbackFixtureBindsPostMergeCleanup(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			waveRoot := filepath.Join(root, "docs", "evidence", wave)
			nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-13")
			recorded := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "post-merge-branch-deletion-readback.json"))
			checkpointRoot := branchDeletionCheckpointRoot(t, waveRoot, recorded)
			outPath := filepath.Join(t.TempDir(), "post-merge-branch-deletion-readback.json")

			var out bytes.Buffer
			code := Run([]string{
				"mission", "recommendations", "post-merge-branch-deletion-readback",
				"--evidence-root", checkpointRoot,
				"--out", outPath,
			}, &out, &out)
			if code != 0 {
				t.Fatalf("post-merge-branch-deletion-readback command failed: %s", out.String())
			}
			if !strings.Contains(out.String(), "status=branch_deletion_bound") ||
				!strings.Contains(out.String(), "post_merge_lifecycle_count=12") ||
				!strings.Contains(out.String(), "remote_branch_deleted_count=12") {
				t.Fatalf("post-merge-branch-deletion-readback output missing cleanup summary: %s", out.String())
			}
			generated := mustLoadJSON[map[string]any](t, outPath)
			generated["evidence_root"] = recorded["evidence_root"]
			if digestValue(generated) != digestValue(recorded) {
				t.Fatalf("post-merge branch deletion readback fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
			}
			if generated["status"] != "branch_deletion_bound" ||
				generated["post_merge_lifecycle_count"] != float64(12) ||
				generated["local_branch_deleted_count"] != float64(12) ||
				generated["remote_branch_deleted_count"] != float64(12) ||
				generated["branches_remaining_total"] != float64(0) ||
				generated["claims_authority_advance"] != false ||
				generated["rsi_remains_denied"] != true {
				t.Fatalf("branch deletion readback must bind cleanup without authority effects: %#v", generated)
			}
		})
	}
}

func TestFeatureDepthWaveBranchDeletionReadbackFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			path := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-13", "post-merge-branch-deletion-readback.json")

			validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.post-merge-branch-deletion-readback.v0.1")
			if err != nil {
				t.Fatal(err)
			}
			if validator != "typed:post-merge-branch-deletion-readback" {
				t.Fatalf("expected typed post-merge branch deletion readback validator, got %s", validator)
			}
		})
	}
}

func TestPostMergeBranchDeletionUsesReusableLocalAndRemoteCleanupRecords(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			readback := mustLoadJSON[AtlasPostMergeBranchDeletionReadback](t, filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-13", "post-merge-branch-deletion-readback.json"))
			if len(readback.Entries) == 0 {
				t.Fatal("branch deletion readback fixture has no entries")
			}
			entry := readback.Entries[0]
			records := BuildAtlasBranchCleanupRecords(AtlasBranchCleanupRecordInput{
				LocalBranchDeleted:           entry.LocalBranchDeleted,
				RemoteBranchDeleted:          entry.RemoteBranchDeleted,
				LocalCodexBranchesRemaining:  entry.LocalCodexBranchesRemaining,
				RemoteCodexBranchesRemaining: entry.RemoteCodexBranchesRemaining,
			})
			summary := SummarizeAtlasBranchCleanupRecords(records)
			if len(records) != 2 ||
				records[0].Scope != "local" ||
				records[1].Scope != "remote" ||
				!records[0].CleanupComplete ||
				!records[1].CleanupComplete ||
				summary.LocalBranchDeletedCount != 1 ||
				summary.RemoteBranchDeletedCount != 1 ||
				summary.BranchesRemainingTotal != 0 ||
				!summary.CleanupComplete {
				t.Fatalf("cleanup records must split local and remote branch deletion facts: records=%#v summary=%#v", records, summary)
			}

			readback.Entries[0].RemoteBranchDeleted = false
			readback.Entries[0].RemoteCodexBranchesRemaining = 1
			readback.RemoteBranchDeletedCount--
			readback.BranchesRemainingTotal = 1
			err := ValidateAtlasPostMergeBranchDeletionReadback(readback)
			if err == nil || !strings.Contains(err.Error(), "remote cleanup record incomplete") {
				t.Fatalf("branch deletion validation must use reusable remote cleanup record, got %v", err)
			}
		})
	}
}

func TestP0BContractConvergencePostMergeCleanupReadbackCoversRecentMergedNodes(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-27")
	recordedPath := filepath.Join(nodeDir, "post-merge-branch-deletion-readback.json")
	outPath := filepath.Join(t.TempDir(), "post-merge-branch-deletion-readback.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "post-merge-branch-deletion-readback",
		"--evidence-root", waveRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("post-merge-branch-deletion-readback command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=branch_deletion_bound",
		"post_merge_lifecycle_count=6",
		"local_branch_deleted_count=6",
		"remote_branch_deleted_count=6",
		"branches_remaining_total=0",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("post-merge-branch-deletion-readback output missing %q:\n%s", want, out.String())
		}
	}
	recorded := mustLoadJSON[AtlasPostMergeBranchDeletionReadback](t, recordedPath)
	if err := ValidateAtlasPostMergeBranchDeletionReadback(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.PostMergeLifecycleCount != 6 ||
		recorded.LocalBranchDeletedCount != 6 ||
		recorded.RemoteBranchDeletedCount != 6 ||
		recorded.BranchesRemainingTotal != 0 ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("P0-B post-merge cleanup readback must bind six clean node lifecycles without authority effects: %#v", recorded)
	}
	wantNodes := map[string]bool{
		"mission-recommendation-p0b-contract-convergence-21": false,
		"mission-recommendation-p0b-contract-convergence-22": false,
		"mission-recommendation-p0b-contract-convergence-23": false,
		"mission-recommendation-p0b-contract-convergence-24": false,
		"mission-recommendation-p0b-contract-convergence-25": false,
		"mission-recommendation-p0b-contract-convergence-26": false,
	}
	for _, entry := range recorded.Entries {
		if _, ok := wantNodes[entry.NodeID]; ok {
			wantNodes[entry.NodeID] = true
		}
		if entry.CIStatus != "passed" ||
			!entry.LocalBranchDeleted ||
			!entry.RemoteBranchDeleted ||
			entry.LocalCodexBranchesRemaining != 0 ||
			entry.RemoteCodexBranchesRemaining != 0 {
			t.Fatalf("P0-B cleanup entry must prove passed CI and local/remote branch deletion: %#v", entry)
		}
	}
	for nodeID, seen := range wantNodes {
		if !seen {
			t.Fatalf("P0-B cleanup readback missing node %s", nodeID)
		}
	}
}

func branchDeletionCheckpointRoot(t *testing.T, waveRoot string, recorded map[string]any) string {
	t.Helper()
	checkpointRoot := filepath.Join(t.TempDir(), "evidence")
	entries, _ := recorded["entries"].([]any)
	for _, item := range entries {
		entry, _ := item.(map[string]any)
		path, _ := entry["path"].(string)
		if strings.TrimSpace(path) == "" {
			t.Fatalf("recorded branch deletion entry missing path: %#v", item)
		}
		src := filepath.Join(waveRoot, filepath.FromSlash(path))
		dst := filepath.Join(checkpointRoot, filepath.FromSlash(path))
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return checkpointRoot
}
