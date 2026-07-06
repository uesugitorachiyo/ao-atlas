package atlas

import (
	"bytes"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMissionReadbackDeltaEvidenceBindsCheckpointComparison(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeID := "mission-recommendation-feature-depth-next-wave-01"
	nodeDir := filepath.Join(waveRoot, "nodes", nodeID)
	beforePath := filepath.Join(waveRoot, "recommendation-readback.json")
	afterPath := filepath.Join(nodeDir, "recommendation-readback-after.json")
	fixturePath := filepath.Join(nodeDir, "mission-readback-delta.json")

	delta, err := BuildAtlasMissionReadbackDelta(beforePath, afterPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasMissionReadbackDelta(delta); err != nil {
		t.Fatal(err)
	}
	fixture := mustLoadJSON[AtlasMissionReadbackDelta](t, fixturePath)
	if err := ValidateAtlasMissionReadbackDelta(fixture); err != nil {
		t.Fatal(err)
	}
	if digestValue(delta) != digestValue(fixture) {
		t.Fatalf("mission readback delta fixture drifted\nwant %s\ngot  %s", digestValue(delta), digestValue(fixture))
	}

	requiredChangedFields := []string{
		"checkpoint_count",
		"completed_nodes",
		"exact_next_action",
		"first_executable_node",
		"ready_nodes",
		"status",
	}
	for _, field := range requiredChangedFields {
		if !containsString(delta.ChangedFields, field) {
			t.Fatalf("delta missing required changed field %q: %#v", field, delta.ChangedFields)
		}
	}
	if !sort.StringsAreSorted(delta.ChangedFields) {
		t.Fatalf("changed fields must be deterministic and sorted: %#v", delta.ChangedFields)
	}
	if delta.Status != "changed" ||
		delta.SourceReadbackPath != "docs/evidence/ao-atlas-feature-depth-wave-v01/recommendation-readback.json" ||
		delta.TargetReadbackPath != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-01/recommendation-readback-after.json" ||
		delta.NumericDeltas["completed_nodes"] != 1 ||
		delta.NumericDeltas["ready_nodes"] != -1 ||
		delta.NumericDeltas["checkpoint_count"] != 1 ||
		delta.StringTransitions["first_executable_node"].Before != nodeID ||
		delta.StringTransitions["first_executable_node"].After != "mission-recommendation-feature-depth-next-wave-02" ||
		delta.BooleanTransitions["final_response_allowed"].Before ||
		delta.BooleanTransitions["final_response_allowed"].After ||
		delta.SchedulesWork ||
		delta.ExecutesWork ||
		delta.ApprovesWork ||
		delta.ClaimsAuthorityAdvance ||
		!delta.RSIRemainsDenied {
		t.Fatalf("mission readback delta must bind node-count, continuation, and safety evidence: %#v", delta)
	}
}

func TestMissionRecommendationsReadbackDeltaCLIWritesDeterministicArtifact(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-01")
	beforePath := filepath.Join(waveRoot, "recommendation-readback.json")
	afterPath := filepath.Join(nodeDir, "recommendation-readback-after.json")
	fixture := mustLoadJSON[AtlasMissionReadbackDelta](t, filepath.Join(nodeDir, "mission-readback-delta.json"))
	outPath := filepath.Join(t.TempDir(), "mission-readback-delta.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "readback-delta",
		"--source-readback", beforePath,
		"--target-readback", afterPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("readback-delta command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "changed_fields=10") ||
		!strings.Contains(out.String(), "mission_readback_delta=") {
		t.Fatalf("readback-delta output missing deterministic summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasMissionReadbackDelta](t, outPath)
	if err := ValidateAtlasMissionReadbackDelta(generated); err != nil {
		t.Fatal(err)
	}
	if digestValue(generated) != digestValue(fixture) {
		t.Fatalf("CLI delta output drifted from fixture\nwant %s\ngot  %s", digestValue(fixture), digestValue(generated))
	}
}
