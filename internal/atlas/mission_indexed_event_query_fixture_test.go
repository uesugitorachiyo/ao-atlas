package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3IndexedEventQueryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-26")
	recordedPath := filepath.Join(nodeDir, "indexed-event-query-fixture.json")
	outPath := filepath.Join(t.TempDir(), "indexed-event-query-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "indexed-event-query-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("indexed-event-query-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=indexed_event_query_ready",
		"event_type_count=5",
		"migration_required=true",
		"query_index_required=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("indexed event query output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("indexed event query fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["migration_required"] != true ||
		generated["query_index_required"] != true ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("indexed event query fixture lost authority state: %#v", generated)
	}
}

func TestMonth3IndexedEventQueryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-26", "indexed-event-query-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.indexed-event-query-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:indexed-event-query-fixture" {
		t.Fatalf("expected typed indexed event query validator, got %s", validator)
	}
}

func TestMonth3IndexedEventQueryFixtureRejectsMissingRollbackEvent(t *testing.T) {
	fixture, err := BuildAtlasIndexedEventQueryFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.EventTypes = []string{"mission", "policy", "approval", "readback"}
	fixture.EventTypeCount = len(fixture.EventTypes)
	if err := ValidateAtlasIndexedEventQueryFixture(fixture); err == nil || !strings.Contains(err.Error(), "event_types must include rollback") {
		t.Fatalf("expected missing rollback event rejection, got %v", err)
	}
}
