package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveReadbackDiffFixturePreservesCompletedAndReadyTransitions(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-02")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-01", "recommendation-readback-after.json")
	targetReadback := filepath.Join(nodeDir, "recommendation-readback-after.json")
	deltaPath := filepath.Join(nodeDir, "mission-readback-delta.json")
	fixturePath := filepath.Join(nodeDir, "resumable-readback-diff-fixture.json")

	fixture, err := BuildAtlasMissionReadbackDiffFixture(sourceReadback, targetReadback, deltaPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasMissionReadbackDiffFixture(fixture); err != nil {
		t.Fatal(err)
	}
	recorded := mustLoadJSON[AtlasMissionReadbackDiffFixture](t, fixturePath)
	if err := ValidateAtlasMissionReadbackDiffFixture(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(fixture) != digestValue(recorded) {
		t.Fatalf("resumable readback diff fixture drifted\nwant %s\ngot  %s", digestValue(fixture), digestValue(recorded))
	}
	if fixture.Status != "resumable" ||
		fixture.CompletedNodeTransition.Before != 1 ||
		fixture.CompletedNodeTransition.After != 2 ||
		fixture.CompletedNodeTransition.Delta != 1 ||
		fixture.ReadyNodeTransition.Before != 39 ||
		fixture.ReadyNodeTransition.After != 38 ||
		fixture.ReadyNodeTransition.Delta != -1 ||
		fixture.CheckpointTransition.Before != 1 ||
		fixture.CheckpointTransition.After != 2 ||
		fixture.FirstExecutableNodeBefore != "mission-recommendation-feature-depth-next-wave-02" ||
		fixture.FirstExecutableNodeAfter != "mission-recommendation-feature-depth-next-wave-03" ||
		fixture.FinalResponseAllowedBefore ||
		fixture.FinalResponseAllowedAfter ||
		!fixture.ResumeRequired ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("fixture must preserve resumable completed/ready transitions without authority expansion: %#v", fixture)
	}
}

func TestMissionRecommendationsReadbackDiffFixtureCLIWritesDeterministicArtifact(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-02")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-01", "recommendation-readback-after.json")
	targetReadback := filepath.Join(nodeDir, "recommendation-readback-after.json")
	deltaPath := filepath.Join(nodeDir, "mission-readback-delta.json")
	recorded := mustLoadJSON[AtlasMissionReadbackDiffFixture](t, filepath.Join(nodeDir, "resumable-readback-diff-fixture.json"))
	outPath := filepath.Join(t.TempDir(), "resumable-readback-diff-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "readback-diff-fixture",
		"--source-readback", sourceReadback,
		"--target-readback", targetReadback,
		"--delta", deltaPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("readback-diff-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=resumable") ||
		!strings.Contains(out.String(), "completed_delta=1") ||
		!strings.Contains(out.String(), "ready_delta=-1") {
		t.Fatalf("readback-diff-fixture output missing transition summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasMissionReadbackDiffFixture](t, outPath)
	if err := ValidateAtlasMissionReadbackDiffFixture(generated); err != nil {
		t.Fatal(err)
	}
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("CLI diff fixture output drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
}
