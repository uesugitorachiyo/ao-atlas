package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveCompactionResumeRegressionPreservesReadyNodeAndExactAction(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18")
	sourcePromptFixture := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-17", "compaction-resume-prompt.json")
	sourcePromptMarkdown := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-17", "compaction-resume-prompt.md")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-17", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-regression.json")
	outPath := filepath.Join(t.TempDir(), "compaction-resume-regression.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-regression",
		"--source-prompt-fixture", sourcePromptFixture,
		"--source-prompt-markdown", sourcePromptMarkdown,
		"--source-readback", sourceReadback,
		"--node-id", "mission-recommendation-feature-depth-next-wave-18",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-19",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-regression command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=guarded") ||
		!strings.Contains(out.String(), "first_executable_node_before=mission-recommendation-feature-depth-next-wave-18") ||
		!strings.Contains(out.String(), "exact_next_action_preserved=true") {
		t.Fatalf("compaction-resume-regression output missing regression state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("compaction resume regression fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(sourcePromptMarkdown)
	if err != nil {
		t.Fatalf("read source prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Next executable node: `mission-recommendation-feature-depth-next-wave-17`",
		"Emit Foundry import for mission-recommendation-feature-depth-next-wave-17 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("source compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestFeatureDepthWaveCompactionResumeRegressionUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-18", "compaction-resume-regression.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.compaction-resume-regression.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:compaction-resume-regression" {
		t.Fatalf("expected typed compaction resume regression validator, got %s", validator)
	}
}

func TestFeatureDepthWaveCompactionResumeRegressionDigestNormalizesPromptLineEndings(t *testing.T) {
	lf := []byte("Next executable node: `mission-recommendation-feature-depth-next-wave-18`\nRSI remains denied.\n")
	crlf := []byte("Next executable node: `mission-recommendation-feature-depth-next-wave-18`\r\nRSI remains denied.\r\n")
	if digestBytes(lf) != digestBytes(crlf) {
		t.Fatalf("expected prompt digest to ignore checkout line endings")
	}
}
