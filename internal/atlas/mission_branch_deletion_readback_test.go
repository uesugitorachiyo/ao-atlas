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
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
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
}

func TestFeatureDepthWaveBranchDeletionReadbackFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-13", "post-merge-branch-deletion-readback.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.post-merge-branch-deletion-readback.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:post-merge-branch-deletion-readback" {
		t.Fatalf("expected typed post-merge branch deletion readback validator, got %s", validator)
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
