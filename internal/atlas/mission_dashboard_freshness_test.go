package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMissionDashboardFreshnessChecksMergedPRAndSyncedMain(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-35")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-34")
	provenancePath := filepath.Join(sourceNodeDir, "mission-dashboard-provenance-links.json")
	readbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	lifecyclePath := filepath.Join(sourceNodeDir, "post-merge-lifecycle.json")
	recordedPath := filepath.Join(nodeDir, "mission-dashboard-freshness-checks.json")
	outPath := filepath.Join(t.TempDir(), "mission-dashboard-freshness-checks.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-freshness-checks",
		"--node-id", "mission-recommendation-feature-depth-next-wave-35",
		"--provenance-links", provenancePath,
		"--source-readback", readbackPath,
		"--post-merge-lifecycle", lifecyclePath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-freshness-checks command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=dashboard_freshness_verified") ||
		!strings.Contains(out.String(), "freshness_check_count=6") ||
		!strings.Contains(out.String(), "all_freshness_checks_passed=true") {
		t.Fatalf("dashboard freshness output missing expected state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMissionDashboardFreshnessChecks](t, recordedPath)
	generated := mustLoadJSON[AtlasMissionDashboardFreshnessChecks](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Mission dashboard freshness fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMissionDashboardFreshnessChecks(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "dashboard_freshness_verified" ||
		recorded.SourceCompletedNodes != 34 ||
		recorded.SourceReadyNodes != 6 ||
		recorded.PRNumber != 361 ||
		recorded.MergeCommit != "ce9161a84769afb5d36bb8c9e9fab0e599277c93" ||
		!recorded.PRMergedAndCleaned ||
		!recorded.MainSyncedToMergeCommit ||
		!recorded.DashboardSourceStillFresh ||
		!recorded.AllFreshnessChecksPassed ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("dashboard freshness checks lost merge/main freshness state: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.mission-dashboard-freshness-checks.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-freshness-checks" {
		t.Fatalf("expected typed Mission dashboard freshness checks validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02MissionDashboardFreshnessChecksMergedPRAndSyncedMain(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-35")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-34")
	provenancePath := filepath.Join(sourceNodeDir, "mission-dashboard-provenance-links.json")
	readbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	lifecyclePath := filepath.Join(sourceNodeDir, "post-merge-lifecycle.json")
	recordedPath := filepath.Join(nodeDir, "mission-dashboard-freshness-checks.json")
	outPath := filepath.Join(t.TempDir(), "mission-dashboard-freshness-checks.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-freshness-checks",
		"--node-id", "mission-recommendation-feature-depth-next-wave-35",
		"--provenance-links", provenancePath,
		"--source-readback", readbackPath,
		"--post-merge-lifecycle", lifecyclePath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-freshness-checks command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=dashboard_freshness_verified") ||
		!strings.Contains(out.String(), "freshness_check_count=6") ||
		!strings.Contains(out.String(), "all_freshness_checks_passed=true") {
		t.Fatalf("dashboard freshness output missing expected state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMissionDashboardFreshnessChecks](t, recordedPath)
	generated := mustLoadJSON[AtlasMissionDashboardFreshnessChecks](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 Mission dashboard freshness fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMissionDashboardFreshnessChecks(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "dashboard_freshness_verified" ||
		recorded.SourceCompletedNodes != 34 ||
		recorded.SourceReadyNodes != 6 ||
		recorded.PRNumber != 489 ||
		recorded.MergeCommit != "348ce2309166cd86cb79521c6c19c6e072adaf1f" ||
		!recorded.PRMergedAndCleaned ||
		!recorded.MainSyncedToMergeCommit ||
		!recorded.DashboardSourceStillFresh ||
		!recorded.AllFreshnessChecksPassed ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 dashboard freshness checks lost merge/main freshness state: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.mission-dashboard-freshness-checks.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-freshness-checks" {
		t.Fatalf("expected typed Mission dashboard freshness checks validator, got %s", validator)
	}
}
