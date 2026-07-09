package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveLocalMainSyncReadbackFixtureValidatesNextNodeSelectionGate(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			nodeDir := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-15")
			inputPath := filepath.Join(nodeDir, "local-main-sync-readback-input.json")
			recordedPath := filepath.Join(nodeDir, "local-main-sync-readback.json")
			recorded := mustLoadJSON[map[string]any](t, recordedPath)
			outPath := filepath.Join(t.TempDir(), "local-main-sync-readback.json")

			var out bytes.Buffer
			code := Run([]string{
				"mission", "recommendations", "local-main-sync-readback",
				"--input", inputPath,
				"--out", outPath,
			}, &out, &out)
			if code != 0 {
				t.Fatalf("local-main-sync-readback command failed: %s", out.String())
			}
			if !strings.Contains(out.String(), "status=local_main_sync_validated") ||
				!strings.Contains(out.String(), "local_main_synced=true") ||
				!strings.Contains(out.String(), "safe_to_select_next_node=true") ||
				!strings.Contains(out.String(), "denial_case_count=3") {
				t.Fatalf("local-main-sync-readback output missing sync summary: %s", out.String())
			}
			generated := mustLoadJSON[map[string]any](t, outPath)
			generated["source_input_path"] = recorded["source_input_path"]
			if digestValue(generated) != digestValue(recorded) {
				t.Fatalf("local main sync readback fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
			}
			if generated["status"] != "local_main_sync_validated" ||
				generated["local_main_synced"] != true ||
				generated["working_tree_clean"] != true ||
				generated["codex_branch_cleanup_confirmed"] != true ||
				generated["safe_to_select_next_node"] != true ||
				generated["claims_authority_advance"] != false ||
				generated["rsi_remains_denied"] != true {
				t.Fatalf("local main sync readback must allow next node only after clean synced main: %#v", generated)
			}
		})
	}
}

func TestFeatureDepthWaveLocalMainSyncReadbackFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			path := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-15", "local-main-sync-readback.json")

			validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.local-main-sync-readback.v0.1")
			if err != nil {
				t.Fatal(err)
			}
			if validator != "typed:local-main-sync-readback" {
				t.Fatalf("expected typed local main sync readback validator, got %s", validator)
			}
		})
	}
}
