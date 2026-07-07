package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMissionDashboardCompactFiltersSummarizeReadyBlockedFailedStates(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-36")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-35")
	readbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	workgraphPath := filepath.Join(sourceNodeDir, "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "mission-dashboard-compact-filters.json")
	outPath := filepath.Join(t.TempDir(), "mission-dashboard-compact-filters.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-compact-filters",
		"--node-id", "mission-recommendation-feature-depth-next-wave-36",
		"--source-readback", readbackPath,
		"--source-workgraph", workgraphPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-compact-filters command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=compact_dashboard_filters_bound") ||
		!strings.Contains(out.String(), "ready_nodes=5") ||
		!strings.Contains(out.String(), "blocked_nodes=0") ||
		!strings.Contains(out.String(), "active_filter=ready") {
		t.Fatalf("dashboard compact filters output missing expected state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMissionDashboardCompactFilters](t, recordedPath)
	generated := mustLoadJSON[AtlasMissionDashboardCompactFilters](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Mission dashboard compact filters fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMissionDashboardCompactFilters(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "compact_dashboard_filters_bound" ||
		recorded.CompletedNodes != 35 ||
		recorded.ReadyNodes != 5 ||
		recorded.BlockedNodes != 0 ||
		recorded.FailedNodes != 0 ||
		recorded.ActiveFilterKey != "ready" ||
		recorded.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-36" ||
		recorded.FinalResponseAllowed ||
		!recorded.ReadyFilterActionable ||
		!recorded.BlockedFilterEmpty ||
		!recorded.FailedFilterEmpty ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("compact dashboard filters lost ready/blocked status: %#v", recorded)
	}
	if len(recorded.Filters) != 4 ||
		recorded.Filters[0].Key != "ready" ||
		recorded.Filters[0].Count != 5 ||
		recorded.Filters[0].PreviewNodeIDs[0] != "mission-recommendation-feature-depth-next-wave-36" ||
		!recorded.Filters[0].Actionable ||
		recorded.Filters[1].Key != "blocked" ||
		recorded.Filters[1].Count != 0 ||
		!recorded.Filters[1].Empty {
		t.Fatalf("compact dashboard filters did not preserve ready versus blocked rows: %#v", recorded.Filters)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.mission-dashboard-compact-filters.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-compact-filters" {
		t.Fatalf("expected typed Mission dashboard compact filters validator, got %s", validator)
	}
}
