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

func TestMissionDashboardCompactFiltersCarrySchemaHealthStatusWhenReadbackHasIt(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-35")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	workgraphPath := filepath.Join(sourceNodeDir, "workgraph-after.json")
	tempDir := t.TempDir()
	syntheticReadbackPath := filepath.Join(tempDir, "recommendation-readback-after.json")
	outPath := filepath.Join(tempDir, "mission-dashboard-compact-filters.json")

	readback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)
	readback.SchemaHealthStatus = "failed_missing_registry_artifacts"
	if err := WriteJSON(syntheticReadbackPath, readback); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-compact-filters",
		"--node-id", "mission-recommendation-schema-health-compact-filters",
		"--source-readback", syntheticReadbackPath,
		"--source-workgraph", workgraphPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-compact-filters command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "schema_health_status=failed_missing_registry_artifacts") {
		t.Fatalf("dashboard compact filters output missing schema health status: %s", out.String())
	}

	fixture := mustLoadJSON[AtlasMissionDashboardCompactFilters](t, outPath)
	if err := ValidateAtlasMissionDashboardCompactFilters(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaHealthStatus != "failed_missing_registry_artifacts" ||
		fixture.SchemaHealthFilterKey != "schema_health" ||
		fixture.SchemaHealthFilterStatus != "failed_missing_registry_artifacts" ||
		!fixture.SchemaHealthFilterActionable {
		t.Fatalf("compact dashboard filters lost schema-health status: %#v", fixture)
	}
	if fixture.FilterCount != 5 || len(fixture.Filters) != 5 {
		t.Fatalf("schema-health status should add a compact filter row: count=%d filters=%#v", fixture.FilterCount, fixture.Filters)
	}
	schemaHealthFilter := fixture.Filters[4]
	if schemaHealthFilter.Key != "schema_health" ||
		schemaHealthFilter.Label != "Schema Health" ||
		schemaHealthFilter.Count != 1 ||
		schemaHealthFilter.DashboardStatus != "failed_missing_registry_artifacts" ||
		!schemaHealthFilter.Actionable ||
		schemaHealthFilter.Empty {
		t.Fatalf("schema-health compact filter row is not actionable: %#v", schemaHealthFilter)
	}

	validator, err := validateRecommendationEvidenceTypedFile(outPath, "ao.atlas.mission-dashboard-compact-filters.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-compact-filters" {
		t.Fatalf("expected typed Mission dashboard compact filters validator, got %s", validator)
	}
}

func TestMissionDashboardCompactFiltersClassifySchemaHealthFilterStates(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-35")
	sourceReadbackPath := filepath.Join(sourceNodeDir, "recommendation-readback-after.json")
	workgraphPath := filepath.Join(sourceNodeDir, "workgraph-after.json")

	cases := []struct {
		name       string
		status     string
		wantState  string
		actionable bool
	}{
		{
			name:       "failed",
			status:     "failed_missing_registry_artifacts",
			wantState:  "failed",
			actionable: true,
		},
		{
			name:       "pending",
			status:     "pending_schema_health_repair",
			wantState:  "pending",
			actionable: true,
		},
		{
			name:       "ready",
			status:     "ready_schema_registry_health",
			wantState:  "ready",
			actionable: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			syntheticReadbackPath := filepath.Join(tempDir, "recommendation-readback-after.json")
			readback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)
			readback.SchemaHealthStatus = tt.status
			if err := WriteJSON(syntheticReadbackPath, readback); err != nil {
				t.Fatal(err)
			}

			fixture, err := BuildAtlasMissionDashboardCompactFilters("mission-recommendation-schema-health-filter-states", syntheticReadbackPath, workgraphPath)
			if err != nil {
				t.Fatal(err)
			}
			if fixture.SchemaHealthFilterState != tt.wantState {
				t.Fatalf("schema health status %q classified as %q, want %q", tt.status, fixture.SchemaHealthFilterState, tt.wantState)
			}
			if fixture.SchemaHealthFilterActionable != tt.actionable {
				t.Fatalf("schema health status %q actionable=%t, want %t", tt.status, fixture.SchemaHealthFilterActionable, tt.actionable)
			}
			if err := ValidateAtlasMissionDashboardCompactFilters(fixture); err != nil {
				t.Fatal(err)
			}
		})
	}
}
