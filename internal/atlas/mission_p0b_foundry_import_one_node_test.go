package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BFoundryImportEmitsExactlyOneReadyNode(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-20")
	path := filepath.Join(nodeDir, "foundry-import", "foundry-import.json")
	recorded := mustLoadJSON[FoundryImport](t, path)
	if err := ValidateFoundryImport(recorded); err != nil {
		t.Fatal(err)
	}
	if len(recorded.Tasks) != 1 {
		t.Fatalf("expected exactly one Foundry import task, got %d", len(recorded.Tasks))
	}
	task := recorded.Tasks[0]
	if task.NodeID != "mission-recommendation-p0b-contract-convergence-20" ||
		task.TaskID != "mission-recommendation-p0b-contract-convergence-20-task" ||
		task.AuthorityBoundary == "" ||
		task.AuthorityBoundary != task.Task.AuthorityBoundary {
		t.Fatalf("Foundry import task drifted from one-node safe handoff: %#v", task)
	}
	if recorded.ExecutesWork || recorded.ApprovesWork || recorded.SchedulesWork {
		t.Fatalf("Foundry import widened authority: %#v", recorded)
	}
}
