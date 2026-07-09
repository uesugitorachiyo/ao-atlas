package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveStaleRemoteBranchRepairFixtureRecordsInterruptedCleanupHandoffs(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			nodeDir := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-14")
			inputPath := filepath.Join(nodeDir, "stale-remote-branch-repair-input.json")
			recordedPath := filepath.Join(nodeDir, "stale-remote-branch-repair.json")
			recorded := mustLoadJSON[map[string]any](t, recordedPath)
			outPath := filepath.Join(t.TempDir(), "stale-remote-branch-repair.json")

			var out bytes.Buffer
			code := Run([]string{
				"mission", "recommendations", "stale-remote-branch-repair",
				"--input", inputPath,
				"--out", outPath,
			}, &out, &out)
			if code != 0 {
				t.Fatalf("stale-remote-branch-repair command failed: %s", out.String())
			}
			if !strings.Contains(out.String(), "status=remote_branch_repair_matrix_recorded") ||
				!strings.Contains(out.String(), "case_count=3") ||
				!strings.Contains(out.String(), "repair_required_cases=2") ||
				!strings.Contains(out.String(), "cleanup_safe_cases=2") {
				t.Fatalf("stale-remote-branch-repair output missing repair summary: %s", out.String())
			}
			generated := mustLoadJSON[map[string]any](t, outPath)
			generated["source_input_path"] = recorded["source_input_path"]
			if digestValue(generated) != digestValue(recorded) {
				t.Fatalf("stale remote branch repair fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
			}
			if generated["status"] != "remote_branch_repair_matrix_recorded" ||
				generated["case_count"] != float64(3) ||
				generated["repair_required_cases"] != float64(2) ||
				generated["cleanup_safe_cases"] != float64(2) ||
				generated["blocked_cases"] != float64(0) ||
				generated["claims_authority_advance"] != false ||
				generated["rsi_remains_denied"] != true {
				t.Fatalf("stale remote branch repair evidence must stay intent-only without authority effects: %#v", generated)
			}
			cases, _ := generated["cases"].([]any)
			for _, item := range cases {
				entry, _ := item.(map[string]any)
				if entry["repair_required"] == true {
					command, _ := entry["repair_command"].(string)
					if !strings.Contains(command, "git push origin --delete codex/") {
						t.Fatalf("repair case must name remote branch deletion command: %#v", entry)
					}
					if entry["safe_to_repair"] != true || entry["blocks_next_node"] != true {
						t.Fatalf("repair case must be safe but block continuation until cleanup: %#v", entry)
					}
				}
			}
		})
	}
}

func TestFeatureDepthWaveStaleRemoteBranchRepairFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			path := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-14", "stale-remote-branch-repair.json")

			validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.stale-remote-branch-repair.v0.1")
			if err != nil {
				t.Fatal(err)
			}
			if validator != "typed:stale-remote-branch-repair" {
				t.Fatalf("expected typed stale remote branch repair validator, got %s", validator)
			}
		})
	}
}
