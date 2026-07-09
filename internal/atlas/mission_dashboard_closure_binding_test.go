package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMissionDashboardClosureBindingRowsBindAtlasClosureEvidence(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-33")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-32")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "mission-dashboard-closure-binding.json")
	outPath := filepath.Join(t.TempDir(), "mission-dashboard-closure-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-closure-binding",
		"--node-id", "mission-recommendation-feature-depth-next-wave-33",
		"--source-readback", sourceReadbackPath,
		"--source-node-dir", sourceNodeDir,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-closure-binding command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=mission_dashboard_closure_bound") ||
		!strings.Contains(out.String(), "row_count=6") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("dashboard closure binding output missing expected state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMissionDashboardClosureBinding](t, recordedPath)
	generated := mustLoadJSON[AtlasMissionDashboardClosureBinding](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Mission dashboard closure binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMissionDashboardClosureBinding(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "mission_dashboard_closure_bound" ||
		recorded.CompletedNodesBefore != 32 ||
		recorded.ReadyNodesBefore != 8 ||
		recorded.FinalResponseAllowed ||
		recorded.RowCount != 6 ||
		!recorded.AtlasClosureEvidenceBound ||
		!recorded.EveryRowHasClosureEvidence ||
		!recorded.EveryRowPreservesSafety ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("dashboard closure binding lost Mission closure guards: %#v", recorded)
	}
	rows := map[string]AtlasMissionDashboardClosureBindingRow{}
	for _, row := range recorded.Rows {
		rows[row.Repo] = row
	}
	for _, repo := range []string{"ao-atlas", "ao-foundry", "ao-command", "ao-promoter", "ao-sentinel", "ao-mission"} {
		row, ok := rows[repo]
		if !ok {
			t.Fatalf("dashboard row missing %s", repo)
		}
		if row.ClosureEvidencePath == "" || row.EvidenceStatus != "bound" || !row.RSIRemainsDenied || row.AuthorityAdvanceClaimed {
			t.Fatalf("dashboard row %s lost closure evidence or safety state: %#v", repo, row)
		}
	}
	if rows["ao-atlas"].ClosureEvidencePath != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-32/recommendation-readback-after.json" {
		t.Fatalf("ao-atlas row must bind source recommendation readback, got %s", rows["ao-atlas"].ClosureEvidencePath)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.mission-dashboard-closure-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-closure-binding" {
		t.Fatalf("expected typed Mission dashboard closure binding validator, got %s", validator)
	}
}

func TestMissionDashboardClosureBindingCarriesSchemaHealthStatusWhenReadbackHasIt(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-32")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	readback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)
	readback.SchemaHealthStatus = "failed_missing_registry_artifacts"

	dir := t.TempDir()
	syntheticReadbackPath := filepath.Join(dir, "recommendation-readback-after.json")
	if err := WriteJSON(syntheticReadbackPath, readback); err != nil {
		t.Fatal(err)
	}
	binding, err := BuildAtlasMissionDashboardClosureBinding(
		"mission-recommendation-feature-depth-next-wave-33",
		syntheticReadbackPath,
		sourceNodeDir,
	)
	if err != nil {
		t.Fatal(err)
	}
	if binding.SchemaHealthStatus != "failed_missing_registry_artifacts" {
		t.Fatalf("dashboard binding lost schema health status: %#v", binding.SchemaHealthStatus)
	}
	assertSchemaHasProperty(t, filepath.Join(root, "schemas", "mission-dashboard-closure-binding.schema.json"), "schema_health_status")
}

func TestMissionDashboardClosureBindingSubsystemRowsUseRollupEvidence(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-32")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")

	binding, err := BuildAtlasMissionDashboardClosureBinding(
		"mission-recommendation-feature-depth-next-wave-33",
		sourceReadbackPath,
		sourceNodeDir,
	)
	if err != nil {
		t.Fatal(err)
	}

	rows := map[string]AtlasMissionDashboardClosureBindingRow{}
	for _, row := range binding.Rows {
		rows[row.Repo] = row
	}
	expected := map[string]struct {
		pathSuffix      string
		readinessStatus string
	}{
		"ao-command":  {"command_readback.json", "readback_agrees"},
		"ao-foundry":  {"foundry-rollup.json", "foundry_rollup_bound"},
		"ao-promoter": {"promoter_no_promotion.json", "no_promotion_requested"},
		"ao-sentinel": {"sentinel_public_safety.json", "public_safety_passed"},
	}
	for repo, want := range expected {
		row, ok := rows[repo]
		if !ok {
			t.Fatalf("dashboard closure binding missing %s row", repo)
		}
		if !strings.HasSuffix(row.ClosureEvidencePath, want.pathSuffix) || row.ReadinessStatus != want.readinessStatus {
			t.Fatalf("%s row did not bind expected evidence: %#v", repo, row)
		}
		if row.ClosureEvidenceDigest == "" || row.AuthorityAdvanceClaimed || !row.RSIRemainsDenied {
			t.Fatalf("%s row lost digest or safety state: %#v", repo, row)
		}
	}
}
