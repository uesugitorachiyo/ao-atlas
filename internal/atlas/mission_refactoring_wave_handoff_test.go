package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactoringWaveHandoffDocumentsRankedTasksAndVerificationGates(t *testing.T) {
	root := repoRoot(t)
	handoffPath := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01", "refactoring-wave-handoff.md")
	summaryPath := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01", "refactoring-wave-handoff-summary.json")
	handoffBytes, err := os.ReadFile(handoffPath)
	if err != nil {
		t.Fatal(err)
	}
	handoff := string(handoffBytes)
	summary := mustLoadJSON[map[string]any](t, summaryPath)

	for _, required := range []string{
		"## Verification Gates",
		"## Merged PR Ledger",
		"## Ranked Tasks",
		"## Next Recommendations",
		"scripts/recommendation-targeted-regressions.sh validator-boundaries",
		"scripts/production-readiness.sh",
		"RSI remains denied",
	} {
		if !strings.Contains(handoff, required) {
			t.Fatalf("refactoring handoff missing %q", required)
		}
	}
	for i := 1; i <= 40; i++ {
		id := "refactoring-next-wave-" + twoDigit(i)
		if !strings.Contains(handoff, id) {
			t.Fatalf("refactoring handoff missing ranked task %s", id)
		}
	}
	if summary["total_nodes"].(float64) != 40 ||
		summary["completed_nodes_before_node_40_pr"].(float64) != 39 ||
		summary["ready_nodes_before_node_40_pr"].(float64) != 1 ||
		summary["promoter_status"].(string) != "no_promotion_requested" ||
		summary["command_readback_status"].(string) != "readback_agrees_no_promotion" ||
		summary["promotion_requested"].(bool) ||
		summary["promotion_granted"].(bool) ||
		summary["claims_authority_advance"].(bool) ||
		!summary["rsi_remains_denied"].(bool) {
		t.Fatalf("refactoring handoff summary lost closure boundaries: %#v", summary)
	}
}
