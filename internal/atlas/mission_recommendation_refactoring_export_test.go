package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsExportRefactoringWaveUsesNextTrackDecision(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-wave-v01"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	sourceAssertion := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json"
	tempDir := t.TempDir()
	decisionPath := filepath.Join(tempDir, "next-track-decision.json")
	refactoringPath := filepath.Join(tempDir, "next-wave-refactoring-recommendations.json")

	if code := Run([]string{
		"mission", "recommendations", "next-track",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--out", decisionPath,
	}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("next-track failed")
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-refactoring-wave",
		"--mission-id", "ao-atlas-refactoring-wave-v01",
		"--source-evidence-root", sourceRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--next-track-decision", decisionPath,
		"--min-tasks", "40",
		"--out", refactoringPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("export-refactoring-wave failed: %s", out.String())
	}
	for _, want := range []string{
		"status=ready",
		"track=refactoring",
		"minimum_tasks=40",
		"recommendation_count=40",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("export-refactoring-wave output missing %q: %s", want, out.String())
		}
	}

	recommendations := mustLoadJSON[AOMissionRefactoringRecommendations](t, refactoringPath)
	if recommendations.Schema != "ao.mission.refactoring-recommendations.v0.1" ||
		recommendations.MissionID != "ao-atlas-refactoring-wave-v01" ||
		recommendations.Track != "refactoring" ||
		recommendations.MinimumTasks != 40 ||
		recommendations.RecommendationCount != 40 ||
		len(recommendations.Tasks) != 40 ||
		recommendations.Tasks[0].ID != "refactoring-next-wave-01" ||
		recommendations.Tasks[0].Owner != "ao-atlas" ||
		!strings.Contains(recommendations.Tasks[0].Task, "recommendation routing") ||
		!recommendations.NoPromotionRequested ||
		recommendations.PromotionGranted ||
		recommendations.ClaimsAuthorityAdvance ||
		!recommendations.RSIRemainsDenied ||
		recommendations.SafeToExecute ||
		recommendations.SchedulesWork ||
		recommendations.ExecutesWork ||
		recommendations.ApprovesWork ||
		recommendations.MutatesRepositories {
		t.Fatalf("refactoring recommendations lost planning-only safety state: %#v", recommendations)
	}
	if err := ValidateAtlasNextWaveRefactoringRecommendations(recommendations, 40); err != nil {
		t.Fatal(err)
	}
}

func TestMissionRecommendationsExportRefactoringWaveRejectsStaleNextTrackReadbackDigest(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-wave-v01"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	sourceAssertion := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json"
	tempDir := t.TempDir()
	decisionPath := filepath.Join(tempDir, "next-track-decision.json")
	refactoringPath := filepath.Join(tempDir, "next-wave-refactoring-recommendations.json")

	if code := Run([]string{
		"mission", "recommendations", "next-track",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--out", decisionPath,
	}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal("next-track failed")
	}
	decision := mustLoadJSON[AtlasRecommendationNextTrackDecision](t, decisionPath)
	decision.SourceReadbackDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	if err := WriteJSON(decisionPath, decision); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-refactoring-wave",
		"--mission-id", "ao-atlas-refactoring-wave-v01",
		"--source-evidence-root", sourceRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--next-track-decision", decisionPath,
		"--min-tasks", "40",
		"--out", refactoringPath,
	}, &out, &out)
	if code == 0 {
		t.Fatalf("export-refactoring-wave accepted stale next-track readback digest: %s", out.String())
	}
	if !strings.Contains(out.String(), "next-track decision source_readback_digest") ||
		!strings.Contains(out.String(), "does not match current source readback digest") {
		t.Fatalf("stale digest error missing actionable message: %s", out.String())
	}
}
