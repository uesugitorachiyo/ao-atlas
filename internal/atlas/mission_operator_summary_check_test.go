package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveOperatorSummaryCheckPreservesExactNextActionWording(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-04")
	readbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-03", "recommendation-readback-after.json")
	summaryPath := filepath.Join(nodeDir, "operator-summary.md")
	fixturePath := filepath.Join(nodeDir, "operator-summary-check.json")
	readback := mustLoadJSON[AtlasRecommendationReadback](t, readbackPath)

	fixture, err := BuildAtlasMissionOperatorSummaryCheck(readbackPath, summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasMissionOperatorSummaryCheck(fixture); err != nil {
		t.Fatal(err)
	}
	recorded := mustLoadJSON[AtlasMissionOperatorSummaryCheck](t, fixturePath)
	if err := ValidateAtlasMissionOperatorSummaryCheck(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(fixture) != digestValue(recorded) {
		t.Fatalf("operator summary check fixture drifted\nwant %s\ngot  %s", digestValue(fixture), digestValue(recorded))
	}
	summaryBytes, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read operator summary: %v", err)
	}
	summary := string(summaryBytes)
	if fixture.Status != "passed" ||
		fixture.CompletedNodes != 3 ||
		fixture.ReadyNodes != 37 ||
		fixture.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-04" ||
		fixture.ExactNextAction != readback.ExactNextAction ||
		fixture.ExactNextActionOccurrences != 1 ||
		!fixture.ExactNextActionWordingPresent ||
		!fixture.NextExecutableNodeWordingPresent ||
		!fixture.FinalResponseDeniedWordingPresent ||
		fixture.FinalResponseAllowed ||
		!fixture.RefusesFinalResponse ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("operator summary check must preserve next action wording without authority effects: %#v", fixture)
	}
	for _, want := range []string{
		"Next executable node: `mission-recommendation-feature-depth-next-wave-04`",
		"Final response allowed: `false`",
		readback.ExactNextAction,
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("operator summary missing %q:\n%s", want, summary)
		}
	}
}

func TestMissionRecommendationsOperatorSummaryCheckCLIWritesDeterministicArtifacts(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-04")
	readbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-03", "recommendation-readback-after.json")
	recorded := mustLoadJSON[AtlasMissionOperatorSummaryCheck](t, filepath.Join(nodeDir, "operator-summary-check.json"))
	recordedSummary, err := os.ReadFile(filepath.Join(nodeDir, "operator-summary.md"))
	if err != nil {
		t.Fatalf("read recorded operator summary: %v", err)
	}
	outDir := t.TempDir()
	summaryOut := filepath.Join(outDir, "operator-summary.md")
	checkOut := filepath.Join(outDir, "operator-summary-check.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "operator-summary-check",
		"--readback", readbackPath,
		"--summary-out", summaryOut,
		"--out", checkOut,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("operator-summary-check command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=passed") ||
		!strings.Contains(out.String(), "exact_next_action_occurrences=1") ||
		!strings.Contains(out.String(), "first_executable_node=mission-recommendation-feature-depth-next-wave-04") {
		t.Fatalf("operator-summary-check output missing summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasMissionOperatorSummaryCheck](t, checkOut)
	if err := ValidateAtlasMissionOperatorSummaryCheck(generated); err != nil {
		t.Fatal(err)
	}
	generated.SummaryMarkdownPath = recorded.SummaryMarkdownPath
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("CLI operator summary check output drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	generatedSummary, err := os.ReadFile(summaryOut)
	if err != nil {
		t.Fatalf("read generated operator summary: %v", err)
	}
	recordedSummaryText := strings.ReplaceAll(string(recordedSummary), "\r\n", "\n")
	generatedSummaryText := strings.ReplaceAll(string(generatedSummary), "\r\n", "\n")
	if generatedSummaryText != recordedSummaryText {
		t.Fatalf("CLI operator summary markdown drifted\nwant:\n%s\ngot:\n%s", recordedSummaryText, generatedSummaryText)
	}
}
