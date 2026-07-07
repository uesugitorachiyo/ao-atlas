package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveBranchCleanupHandoffSummaryFixtureBindsOperatorHandoff(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-15", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "branch-cleanup-handoff-summary.json")
	outPath := filepath.Join(t.TempDir(), "branch-cleanup-handoff-summary.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "branch-cleanup-handoff-summary",
		"--evidence-root", waveRoot,
		"--source-readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("branch-cleanup-handoff-summary command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=branch_cleanup_handoff_summarized") ||
		!strings.Contains(out.String(), "post_merge_lifecycle_count=15") ||
		!strings.Contains(out.String(), "operator_handoff_status=cleanup_ledger_ready") {
		t.Fatalf("branch-cleanup-handoff-summary output missing handoff summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("branch cleanup handoff summary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["status"] != "branch_cleanup_handoff_summarized" ||
		generated["post_merge_lifecycle_count"] != float64(15) ||
		generated["merged_and_cleaned_count"] != float64(15) ||
		generated["passed_ci_count"] != float64(15) ||
		generated["cleanup_complete"] != true ||
		generated["final_response_allowed"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("branch cleanup handoff must summarize cleanup without authority effects: %#v", generated)
	}
}

func TestFeatureDepthWaveBranchCleanupHandoffSummaryUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-16", "branch-cleanup-handoff-summary.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.branch-cleanup-handoff-summary.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:branch-cleanup-handoff-summary" {
		t.Fatalf("expected typed branch cleanup handoff summary validator, got %s", validator)
	}
}
