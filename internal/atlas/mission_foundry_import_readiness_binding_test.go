package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveFoundryImportReadinessBindingKeepsOneActiveNode(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceNodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-28")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-29")
	recordedPath := filepath.Join(nodeDir, "foundry-import-readiness-binding.json")
	outPath := filepath.Join(t.TempDir(), "foundry-import-readiness-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-import-readiness-binding",
		"--node-id", "mission-recommendation-feature-depth-next-wave-29",
		"--source-readback", filepath.Join(sourceNodeDir, "recommendation-readback-after.json"),
		"--source-workgraph", filepath.Join(sourceNodeDir, "workgraph-after.json"),
		"--foundry-import", filepath.Join(nodeDir, "foundry-import.json"),
		"--foundry-handoff", filepath.Join(nodeDir, "foundry-continuation-handoff.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-import-readiness-binding command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasFoundryImportReadinessBinding](t, recordedPath)
	generated := mustLoadJSON[AtlasFoundryImportReadinessBinding](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Foundry import readiness binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasFoundryImportReadinessBinding(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "single_active_foundry_import_ready" ||
		recorded.FoundryTaskCount != 1 ||
		recorded.ActiveNodeID != "mission-recommendation-feature-depth-next-wave-29" ||
		recorded.ActiveTaskID != "mission-recommendation-feature-depth-next-wave-29-task" ||
		recorded.WorkgraphNextReadyNode != recorded.ActiveNodeID ||
		recorded.ReadbackFirstExecutableNode != recorded.ActiveNodeID ||
		recorded.HandoffFirstSafeNode != recorded.ActiveNodeID ||
		!recorded.MatchesWorkgraph ||
		!recorded.MatchesReadbackNextNode ||
		!recorded.HandoffMatchesImport ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("Foundry import readiness binding lost single-active-node state: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.foundry-import-readiness-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-import-readiness-binding" {
		t.Fatalf("expected typed Foundry import readiness binding validator, got %s", validator)
	}
}
