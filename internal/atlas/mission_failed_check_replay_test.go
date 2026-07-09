package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveFailedCheckReplayFixtureBlocksMergeAndRoutesRetry(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			nodeDir := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-11")
			inputPath := filepath.Join(nodeDir, "failed-check-replay-input.json")
			recorded := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "failed-check-replay-fixture.json"))
			outPath := filepath.Join(t.TempDir(), "failed-check-replay-fixture.json")

			var out bytes.Buffer
			code := Run([]string{
				"mission", "recommendations", "failed-check-replay",
				"--input", inputPath,
				"--out", outPath,
			}, &out, &out)
			if code != 0 {
				t.Fatalf("failed-check-replay command failed: %s", out.String())
			}
			if !strings.Contains(out.String(), "status=replay_recorded") ||
				!strings.Contains(out.String(), "merge_denied_cases=3") ||
				!strings.Contains(out.String(), "retry_allowed_cases=1") {
				t.Fatalf("failed-check-replay output missing replay summary: %s", out.String())
			}
			generated := mustLoadJSON[map[string]any](t, outPath)
			if digestValue(generated) != digestValue(recorded) {
				t.Fatalf("failed check replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
			}
			if generated["safe_to_merge"] != false ||
				generated["merge_denied_cases"] != float64(3) ||
				generated["retry_allowed_cases"] != float64(1) ||
				generated["claims_authority_advance"] != false ||
				generated["rsi_remains_denied"] != true {
				t.Fatalf("failed check replay must block merge and preserve authority boundaries: %#v", generated)
			}
		})
	}
}

func TestFeatureDepthWaveFailedCheckReplayFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	for _, wave := range []string{"ao-atlas-feature-depth-wave-v01", "ao-atlas-feature-depth-wave-v02"} {
		t.Run(wave, func(t *testing.T) {
			path := filepath.Join(root, "docs", "evidence", wave, "nodes", "mission-recommendation-feature-depth-next-wave-11", "failed-check-replay-fixture.json")

			validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.failed-check-replay-fixture.v0.1")
			if err != nil {
				t.Fatal(err)
			}
			if validator != "typed:failed-check-replay-fixture" {
				t.Fatalf("expected typed failed check replay validator, got %s", validator)
			}
		})
	}
}
