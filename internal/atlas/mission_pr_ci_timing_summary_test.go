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

func TestPRCINormalizedRowsCoverFeatureDepthClosureAndRefactoringWaves(t *testing.T) {
	root := repoRoot(t)
	featureDepthLedger := mustLoadJSON[AtlasPRCITimingLedger](t, filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-09", "pr-ci-timing-ledger.json"))
	closureLedger := mustLoadJSON[struct {
		Entries []AtlasPRCILedgerEntry `json:"entries"`
	}](t, filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01", "nodes", "mission-recommendation-final-closure-consolidation-08", "hardening-nodes-28-40-pr-ci-ledger.json"))

	rows, err := NormalizeAtlasPRCILedgerRows([]AtlasPRCINormalizationInput{
		{
			SourceWave: "feature_depth",
			Rows:       FeatureDepthPRCITimingRowsToNormalizedInputs(featureDepthLedger.Rows[:2]),
		},
		{
			SourceWave: "final_closure",
			Rows:       ClosurePRCILedgerEntriesToNormalizedInputs(closureLedger.Entries[:2]),
		},
		{
			SourceWave: "refactoring",
			Rows: []AtlasPRCINormalizedRow{
				{
					NodeID:          "refactoring-next-wave-20",
					PRNumber:        433,
					MergeCommit:     "cc20aae78888f4bfbdbf57b5f2b822e1323dc431",
					CIStatus:        "passed",
					CheckCount:      9,
					SuccessCount:    9,
					UbuntuPassed:    true,
					MacosPassed:     true,
					WindowsPassed:   true,
					WindowsSeconds:  1948,
					MaxCheckSeconds: 1948,
					SlowestCheck:    "Production readiness (windows-latest)",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasPRCINormalizedRows(rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("expected five normalized rows, got %d: %#v", len(rows), rows)
	}
	waves := map[string]int{}
	for _, row := range rows {
		waves[row.SourceWave]++
		if row.NormalizedSchema != AtlasPRCINormalizedRowContract ||
			row.CIStatus != "passed" ||
			!row.Merged ||
			!row.AllRequiredChecksPassed ||
			!row.UbuntuPassed ||
			!row.MacosPassed ||
			!row.WindowsPassed ||
			row.PromotionGranted ||
			row.ClaimsAuthorityAdvance ||
			!row.RSIRemainsDenied ||
			row.SafeToExecute ||
			row.SchedulesWork ||
			row.ExecutesWork ||
			row.ApprovesWork ||
			row.MutatesRepositories {
			t.Fatalf("normalized row lost PR/CI or safety state: %#v", row)
		}
	}
	if waves["feature_depth"] != 2 || waves["final_closure"] != 2 || waves["refactoring"] != 1 {
		t.Fatalf("normalized rows did not cover all source waves: %#v", waves)
	}
	if rows[0].PRNumber != 291 || rows[4].PRNumber != 433 {
		t.Fatalf("normalized rows must be sorted by PR number: %#v", rows)
	}
	if rows[4].SourceWave != "refactoring" || rows[4].NodeID != "refactoring-next-wave-20" || rows[4].WindowsSeconds != 1948 {
		t.Fatalf("refactoring row did not preserve current wave lifecycle evidence: %#v", rows[4])
	}
}
