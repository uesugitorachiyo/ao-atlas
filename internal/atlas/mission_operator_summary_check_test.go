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

func TestFeatureDepthWaveV02OperatorSummaryCheckPreservesExactNextActionWording(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
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
		t.Fatalf("v02 operator summary check fixture drifted\nwant %s\ngot  %s", digestValue(fixture), digestValue(recorded))
	}
	summaryBytes, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read v02 operator summary: %v", err)
	}
	summary := string(summaryBytes)
	if fixture.CompletedNodes != 3 ||
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
		t.Fatalf("v02 operator summary check must preserve next action wording without authority effects: %#v", fixture)
	}
	for _, want := range []string{
		"Next executable node: `mission-recommendation-feature-depth-next-wave-04`",
		"Final response allowed: `false`",
		readback.ExactNextAction,
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("v02 operator summary missing %q:\n%s", want, summary)
		}
	}
}

func TestP0BContractConvergenceOperatorSummaryPreservesContinuationDenial(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-28")
	readbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-27", "recommendation-readback-after.json")
	summaryPath := filepath.Join(nodeDir, "operator-summary.md")
	fixturePath := filepath.Join(nodeDir, "operator-summary-check.json")

	fixture, err := BuildAtlasMissionOperatorSummaryCheck(readbackPath, summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	recorded := mustLoadJSON[AtlasMissionOperatorSummaryCheck](t, fixturePath)
	if err := ValidateAtlasMissionOperatorSummaryCheck(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(fixture) != digestValue(recorded) {
		t.Fatalf("P0-B operator summary check fixture drifted\nwant %s\ngot  %s", digestValue(fixture), digestValue(recorded))
	}
	summaryBytes, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read P0-B operator summary: %v", err)
	}
	summary := string(summaryBytes)
	for _, want := range []string{
		"Completed nodes: 27 / 30",
		"Ready nodes: 3",
		"Next executable node: `mission-recommendation-p0b-contract-convergence-28`",
		"Emit Foundry import for mission-recommendation-p0b-contract-convergence-28 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("P0-B operator summary missing %q:\n%s", want, summary)
		}
	}
	if recorded.FinalResponseAllowed ||
		!recorded.RefusesFinalResponse ||
		!recorded.FinalResponseDeniedWordingPresent ||
		recorded.ExactNextActionOccurrences != 1 ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("P0-B operator summary must preserve continuation denial without authority effects: %#v", recorded)
	}
}
