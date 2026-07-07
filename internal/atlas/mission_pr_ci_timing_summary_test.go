package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWavePRCITimingSummaryFixtureAggregatesMergedNodeChecks(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-09")
	ledgerPath := filepath.Join(nodeDir, "pr-ci-timing-ledger.json")
	recorded := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "pr-ci-timing-summary.json"))
	outPath := filepath.Join(t.TempDir(), "pr-ci-timing-summary.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "pr-ci-timing-summary",
		"--ledger", ledgerPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("pr-ci-timing-summary command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=summarized") ||
		!strings.Contains(out.String(), "row_count=3") ||
		!strings.Contains(out.String(), "max_windows_seconds=812") {
		t.Fatalf("pr-ci-timing-summary output missing aggregate summary: %s", out.String())
	}
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("PR/CI timing summary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["status"] != "summarized" ||
		generated["row_count"] != float64(3) ||
		generated["merged_prs"] != float64(3) ||
		generated["ci_passed_prs"] != float64(3) ||
		generated["max_check_seconds"] != float64(812) ||
		generated["slowest_pr_number"] != float64(334) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("summary must bind merged PR/CI timing without authority effects: %#v", generated)
	}
}

func TestFeatureDepthWavePRCITimingSummaryUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-09", "pr-ci-timing-summary.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.pr-ci-timing-summary.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:pr-ci-timing-summary" {
		t.Fatalf("expected typed PR/CI timing summary validator, got %s", validator)
	}
}
