package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveFoundryHandoffReplayFixtureBindsResumedNode(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-31")
	sourceReadbackPath := filepath.Join(nodeDir, "resume-readback.json")
	sourceWorkgraphPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-30", "workgraph-after.json")
	foundryImportPath := filepath.Join(nodeDir, "foundry-import.json")
	foundryHandoffPath := filepath.Join(nodeDir, "foundry-continuation-handoff.json")
	recordedPath := filepath.Join(nodeDir, "foundry-handoff-replay-fixture.json")
	outPath := filepath.Join(t.TempDir(), "foundry-handoff-replay-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-handoff-replay-fixture",
		"--node-id", "mission-recommendation-feature-depth-next-wave-31",
		"--source-readback", sourceReadbackPath,
		"--source-workgraph", sourceWorkgraphPath,
		"--foundry-import", foundryImportPath,
		"--foundry-handoff", foundryHandoffPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-handoff-replay-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=replay_guarded") ||
		!strings.Contains(out.String(), "active_node_id=mission-recommendation-feature-depth-next-wave-31") ||
		!strings.Contains(out.String(), "single_active_import_task=true") {
		t.Fatalf("foundry handoff replay output missing replay state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasFoundryHandoffReplayFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasFoundryHandoffReplayFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("foundry handoff replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasFoundryHandoffReplayFixture(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "replay_guarded" ||
		recorded.ResumedFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-31" ||
		recorded.ActiveNodeID != recorded.ResumedFirstExecutableNode ||
		recorded.HandoffFirstSafeNode != recorded.ResumedFirstExecutableNode ||
		recorded.FoundryTaskCount != 1 ||
		recorded.MutationClass != "low_risk_code" ||
		!recorded.SingleActiveImportTask ||
		!recorded.HandoffMatchesResumedReadback ||
		!recorded.ImportMatchesResumedReadback ||
		!recorded.HandoffMatchesWorkgraph ||
		!recorded.BoundedMutationClass ||
		!recorded.PromptPreservesActiveNode ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("foundry handoff replay fixture lost resumed-node binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.foundry-handoff-replay-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-handoff-replay-fixture" {
		t.Fatalf("expected typed foundry handoff replay fixture validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02FoundryHandoffReplayFixtureBindsResumedNode(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-31")
	sourceReadbackPath := filepath.Join(nodeDir, "resume-readback.json")
	sourceWorkgraphPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-30", "workgraph-after.json")
	foundryImportPath := filepath.Join(nodeDir, "foundry-import.json")
	foundryHandoffPath := filepath.Join(nodeDir, "foundry-continuation-handoff.json")
	recordedPath := filepath.Join(nodeDir, "foundry-handoff-replay-fixture.json")
	outPath := filepath.Join(t.TempDir(), "foundry-handoff-replay-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-handoff-replay-fixture",
		"--node-id", "mission-recommendation-feature-depth-next-wave-31",
		"--source-readback", sourceReadbackPath,
		"--source-workgraph", sourceWorkgraphPath,
		"--foundry-import", foundryImportPath,
		"--foundry-handoff", foundryHandoffPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-handoff-replay-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=replay_guarded") ||
		!strings.Contains(out.String(), "active_node_id=mission-recommendation-feature-depth-next-wave-31") ||
		!strings.Contains(out.String(), "single_active_import_task=true") {
		t.Fatalf("foundry handoff replay output missing replay state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasFoundryHandoffReplayFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasFoundryHandoffReplayFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 foundry handoff replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasFoundryHandoffReplayFixture(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "replay_guarded" ||
		recorded.ResumedFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-31" ||
		recorded.ActiveNodeID != recorded.ResumedFirstExecutableNode ||
		recorded.HandoffFirstSafeNode != recorded.ResumedFirstExecutableNode ||
		recorded.FoundryTaskCount != 1 ||
		recorded.MutationClass != "low_risk_code" ||
		!recorded.SingleActiveImportTask ||
		!recorded.HandoffMatchesResumedReadback ||
		!recorded.ImportMatchesResumedReadback ||
		!recorded.HandoffMatchesWorkgraph ||
		!recorded.BoundedMutationClass ||
		!recorded.PromptPreservesActiveNode ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 foundry handoff replay fixture lost resumed-node binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.foundry-handoff-replay-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-handoff-replay-fixture" {
		t.Fatalf("expected typed foundry handoff replay fixture validator, got %s", validator)
	}
}
