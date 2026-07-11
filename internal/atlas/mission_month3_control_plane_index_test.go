package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosureControlPlaneIndexFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-27-control-plane-index")
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

	recorded := mustLoadJSON[AtlasIndexedEventQueryFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasIndexedEventQueryFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("control-plane index fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if !recorded.MigrationRequired ||
		!recorded.QueryIndexRequired ||
		!containsAll(recorded.EventTypes, []string{"mission", "policy", "approval", "rollback", "readback"}) ||
		!containsAll(recorded.QueryFields, []string{"mission_id", "event_type", "created_at", "source_digest"}) ||
		recorded.ExecutesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("control-plane index fixture lost durable search or safety state: %#v", recorded)
	}
}

func TestMonth3FinalClosureControlPlaneIndexFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-27-control-plane-index", "indexed-event-query-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.indexed-event-query-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:indexed-event-query-fixture" {
		t.Fatalf("expected typed indexed event query validator, got %s", validator)
	}
}
