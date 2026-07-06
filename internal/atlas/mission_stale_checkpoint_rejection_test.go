package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveStaleCheckpointRejectionRejectsOutdatedContinuationPrompt(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-03")
	staleReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-01", "recommendation-readback-after.json")
	latestReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-02", "recommendation-readback-after.json")
	fixturePath := filepath.Join(nodeDir, "stale-checkpoint-rejection.json")

	fixture, err := BuildAtlasMissionStaleCheckpointRejection(staleReadback, latestReadback, staleReadback)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasMissionStaleCheckpointRejection(fixture); err != nil {
		t.Fatal(err)
	}
	recorded := mustLoadJSON[AtlasMissionStaleCheckpointRejection](t, fixturePath)
	if err := ValidateAtlasMissionStaleCheckpointRejection(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(fixture) != digestValue(recorded) {
		t.Fatalf("stale checkpoint rejection fixture drifted\nwant %s\ngot  %s", digestValue(fixture), digestValue(recorded))
	}
	if fixture.Status != "rejected" ||
		fixture.RejectionReason != "stale_checkpoint" ||
		fixture.StaleCheckpoint.CompletedNodes != 1 ||
		fixture.LatestCheckpoint.CompletedNodes != 2 ||
		fixture.StaleCheckpoint.ReadyNodes != 39 ||
		fixture.LatestCheckpoint.ReadyNodes != 38 ||
		fixture.StaleCheckpoint.CheckpointCount != 1 ||
		fixture.LatestCheckpoint.CheckpointCount != 2 ||
		fixture.PromptNextExecutableNode != "mission-recommendation-feature-depth-next-wave-02" ||
		fixture.ExpectedCurrentNextExecutableNode != "mission-recommendation-feature-depth-next-wave-03" ||
		fixture.FinalResponseAllowed ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("fixture must reject stale continuation prompts without scheduling work or expanding authority: %#v", fixture)
	}
}

func TestMissionRecommendationsStaleCheckpointRejectionCLIWritesDeterministicArtifact(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-03")
	staleReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-01", "recommendation-readback-after.json")
	latestReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-02", "recommendation-readback-after.json")
	recorded := mustLoadJSON[AtlasMissionStaleCheckpointRejection](t, filepath.Join(nodeDir, "stale-checkpoint-rejection.json"))
	outPath := filepath.Join(t.TempDir(), "stale-checkpoint-rejection.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "stale-checkpoint-rejection",
		"--stale-readback", staleReadback,
		"--latest-readback", latestReadback,
		"--prompt-readback", staleReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("stale-checkpoint-rejection command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=rejected") ||
		!strings.Contains(out.String(), "rejection_reason=stale_checkpoint") ||
		!strings.Contains(out.String(), "prompt_next_node=mission-recommendation-feature-depth-next-wave-02") ||
		!strings.Contains(out.String(), "expected_current_next_node=mission-recommendation-feature-depth-next-wave-03") {
		t.Fatalf("stale-checkpoint-rejection output missing rejection summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasMissionStaleCheckpointRejection](t, outPath)
	if err := ValidateAtlasMissionStaleCheckpointRejection(generated); err != nil {
		t.Fatal(err)
	}
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("CLI stale checkpoint rejection output drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
}
