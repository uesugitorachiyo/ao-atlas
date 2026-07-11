package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CompactTimelineFilterFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-30")
	recordedPath := filepath.Join(nodeDir, "compact-timeline-filter-fixture.json")
	outPath := filepath.Join(t.TempDir(), "compact-timeline-filter-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compact-timeline-filter-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compact-timeline-filter-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=compact_timeline_filter_ready",
		"filter_count=5",
		"stale_records_distinguished=true",
		"duplicate_records_distinguished=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("compact timeline filter output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("compact timeline filter fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["stale_records_distinguished"] != true ||
		generated["duplicate_records_distinguished"] != true ||
		generated["pending_records_distinguished"] != true ||
		generated["denied_records_distinguished"] != true ||
		generated["completed_records_distinguished"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("compact timeline filter fixture lost safety state: %#v", generated)
	}
}

func TestMonth3CompactTimelineFilterFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-30", "compact-timeline-filter-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.compact-timeline-filter-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:compact-timeline-filter-fixture" {
		t.Fatalf("expected typed compact timeline filter validator, got %s", validator)
	}
}

func TestMonth3CompactTimelineFilterFixtureRejectsMissingDeniedFilter(t *testing.T) {
	fixture, err := BuildAtlasCompactTimelineFilterFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.DeniedRecordsDistinguished = false
	if err := ValidateAtlasCompactTimelineFilterFixture(fixture); err == nil || !strings.Contains(err.Error(), "denied_records_distinguished must be true") {
		t.Fatalf("expected missing denied filter rejection, got %v", err)
	}
}
