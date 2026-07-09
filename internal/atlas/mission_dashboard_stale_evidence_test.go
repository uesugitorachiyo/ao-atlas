package atlas

import (
	"path/filepath"
	"testing"
)

func TestMissionDashboardStaleEvidenceDetectionFlagsSupersededReadbacksAndOldExports(t *testing.T) {
	root := repoRoot(t)
	latestReadback := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-40", "recommendation-readback-after.json")
	dashboardPaths := []string{
		filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-33", "mission-dashboard-closure-binding.json"),
		filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-35", "mission-dashboard-freshness-checks.json"),
		filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-36", "mission-dashboard-compact-filters.json"),
	}
	oldExportRoots := []string{
		filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-followup-durability-v02"),
		filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04"),
	}

	fixture, err := BuildAtlasMissionDashboardStaleEvidenceDetection("refactoring-next-wave-34", latestReadback, dashboardPaths, oldExportRoots)
	if err != nil {
		t.Fatal(err)
	}
	if fixture.Status != "stale_dashboard_evidence_detected" ||
		fixture.LatestCompletedNodes != 40 ||
		fixture.DashboardEvidenceCount != 3 ||
		fixture.StaleDashboardEvidenceCount != 3 ||
		fixture.OldExportRootCount != 2 ||
		fixture.RecommendedAction != "regenerate_dashboard_evidence_from_latest_readback_before_operator_handoff" {
		t.Fatalf("stale dashboard detector did not summarize superseded evidence: %#v", fixture)
	}
	for _, item := range fixture.StaleDashboardEvidence {
		if !item.Stale || item.SourceReadbackDigest == fixture.LatestReadbackDigest || item.StaleReason == "" {
			t.Fatalf("stale dashboard item lost stale proof: %#v", item)
		}
	}
	if fixture.PromotionRequested || fixture.PromotionGranted || fixture.ClaimsAuthorityAdvance || !fixture.RSIRemainsDenied {
		t.Fatalf("stale dashboard detector must preserve no-promotion safety: %#v", fixture)
	}
}
