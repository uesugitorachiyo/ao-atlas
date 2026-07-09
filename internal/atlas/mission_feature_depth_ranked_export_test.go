package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveRankedExportBindsFinalClosureReadbackEvidence(t *testing.T) {
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
	finalClosureRoot := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01"
	sourceReadback := finalClosureRoot + "/nodes/mission-recommendation-final-closure-consolidation-22/recommendation-readback-after.json"
	sourceAssertion := finalClosureRoot + "/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json"
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-37")
	recordedExportPath := filepath.Join(nodeDir, "next-wave-feature-depth-recommendations.json")
	recordedFixturePath := filepath.Join(nodeDir, "next-wave-recommendation-export.json")
	tmpDir := t.TempDir()
	tmpExportPath := filepath.Join(tmpDir, "next-wave-feature-depth-recommendations.json")
	tmpFixturePath := filepath.Join(tmpDir, "next-wave-recommendation-export.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-next-wave",
		"--mission-id", "ao-atlas-next-feature-depth-wave-v01",
		"--source-evidence-root", finalClosureRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--min-tasks", "40",
		"--out", tmpExportPath,
		"--fixture-out", tmpFixturePath,
		"--node-id", "mission-recommendation-feature-depth-next-wave-37",
		"--expected-next-node", "mission-recommendation-feature-depth-next-wave-38",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("export-next-wave failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "recommendation_count=40") ||
		!strings.Contains(out.String(), "ranked_tasks=40") ||
		!strings.Contains(out.String(), "safe_to_execute=false") {
		t.Fatalf("export-next-wave output missing ranked planning-only state: %s", out.String())
	}
	recordedExport := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, recordedExportPath)
	generatedExport := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, tmpExportPath)
	if digestValue(generatedExport) != digestValue(recordedExport) {
		t.Fatalf("Feature Depth recommendations export changed\nwant %s\ngot  %s", digestValue(recordedExport), digestValue(generatedExport))
	}
	recordedFixture := mustLoadJSON[AtlasNextWaveRecommendationExport](t, recordedFixturePath)
	generatedFixture := mustLoadJSON[AtlasNextWaveRecommendationExport](t, tmpFixturePath)
	if digestValue(generatedFixture) != digestValue(recordedFixture) {
		t.Fatalf("next-wave recommendation export fixture changed\nwant %s\ngot  %s", digestValue(recordedFixture), digestValue(generatedFixture))
	}
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(recordedExport, 40); err != nil {
		t.Fatal(err)
	}
	if recordedFixture.Schema != "ao.atlas.next-wave-recommendation-export.v0.1" ||
		recordedFixture.NodeID != "mission-recommendation-feature-depth-next-wave-37" ||
		recordedFixture.Status != "exported" ||
		recordedFixture.SourceEvidenceRoot != finalClosureRoot ||
		recordedFixture.SourceReadbackPath != sourceReadback ||
		recordedFixture.SourceAssertionPath != sourceAssertion ||
		recordedFixture.CompletedNodesBefore != 22 ||
		recordedFixture.ReadyNodesBefore != 2 ||
		recordedFixture.ExpectedNextNode != "mission-recommendation-feature-depth-next-wave-38" ||
		recordedFixture.MinimumRankedTasks != 40 ||
		recordedFixture.RecommendationCount != 40 ||
		!recordedFixture.RankedTaskFloorMet ||
		!recordedFixture.NoPromotionRequested ||
		recordedFixture.PromotionGranted ||
		recordedFixture.ClaimsAuthorityAdvance ||
		!recordedFixture.RSIRemainsDenied {
		t.Fatalf("ranked export fixture lost final closure evidence binding: %#v", recordedFixture)
	}
	if len(recordedExport.Tasks) != 40 ||
		recordedExport.Tasks[0].ID != "feature-depth-next-wave-01" ||
		recordedExport.Tasks[39].ID != "feature-depth-next-wave-40" ||
		recordedExport.Tasks[0].EvidenceRefs[0] != sourceReadback ||
		recordedExport.Tasks[0].EvidenceRefs[1] != sourceAssertion ||
		recordedExport.SafeToExecute ||
		recordedExport.SchedulesWork ||
		recordedExport.ExecutesWork ||
		recordedExport.ApprovesWork ||
		recordedExport.MutatesRepositories {
		t.Fatalf("ranked Feature Depth export must remain a 40-task intent-only planning artifact: %#v", recordedExport)
	}
}

func TestFeatureDepthWaveV02RankedExportBindsFinalClosureReadbackEvidence(t *testing.T) {
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
	finalClosureRoot := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01"
	sourceReadback := finalClosureRoot + "/nodes/mission-recommendation-final-closure-consolidation-22/recommendation-readback-after.json"
	sourceAssertion := finalClosureRoot + "/nodes/mission-recommendation-final-closure-consolidation-22/no-promotion-no-rsi-assertion.json"
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02", "nodes", "mission-recommendation-feature-depth-next-wave-37")
	recordedExportPath := filepath.Join(nodeDir, "next-wave-feature-depth-recommendations.json")
	recordedFixturePath := filepath.Join(nodeDir, "next-wave-recommendation-export.json")
	tmpDir := t.TempDir()
	tmpExportPath := filepath.Join(tmpDir, "next-wave-feature-depth-recommendations.json")
	tmpFixturePath := filepath.Join(tmpDir, "next-wave-recommendation-export.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-next-wave",
		"--mission-id", "ao-atlas-next-feature-depth-wave-v02",
		"--source-evidence-root", finalClosureRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--min-tasks", "40",
		"--out", tmpExportPath,
		"--fixture-out", tmpFixturePath,
		"--node-id", "mission-recommendation-feature-depth-next-wave-37",
		"--expected-next-node", "mission-recommendation-feature-depth-next-wave-38",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("export-next-wave failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "recommendation_count=40") ||
		!strings.Contains(out.String(), "ranked_tasks=40") ||
		!strings.Contains(out.String(), "safe_to_execute=false") {
		t.Fatalf("export-next-wave output missing ranked planning-only state: %s", out.String())
	}
	recordedExport := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, recordedExportPath)
	generatedExport := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, tmpExportPath)
	if digestValue(generatedExport) != digestValue(recordedExport) {
		t.Fatalf("v02 Feature Depth recommendations export changed\nwant %s\ngot  %s", digestValue(recordedExport), digestValue(generatedExport))
	}
	recordedFixture := mustLoadJSON[AtlasNextWaveRecommendationExport](t, recordedFixturePath)
	generatedFixture := mustLoadJSON[AtlasNextWaveRecommendationExport](t, tmpFixturePath)
	if digestValue(generatedFixture) != digestValue(recordedFixture) {
		t.Fatalf("v02 next-wave recommendation export fixture changed\nwant %s\ngot  %s", digestValue(recordedFixture), digestValue(generatedFixture))
	}
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(recordedExport, 40); err != nil {
		t.Fatal(err)
	}
	if recordedFixture.Schema != "ao.atlas.next-wave-recommendation-export.v0.1" ||
		recordedFixture.NodeID != "mission-recommendation-feature-depth-next-wave-37" ||
		recordedFixture.Status != "exported" ||
		recordedFixture.SourceEvidenceRoot != finalClosureRoot ||
		recordedFixture.SourceReadbackPath != sourceReadback ||
		recordedFixture.SourceAssertionPath != sourceAssertion ||
		recordedFixture.CompletedNodesBefore != 22 ||
		recordedFixture.ReadyNodesBefore != 2 ||
		recordedFixture.ExpectedNextNode != "mission-recommendation-feature-depth-next-wave-38" ||
		recordedFixture.MinimumRankedTasks != 40 ||
		recordedFixture.RecommendationCount != 40 ||
		!recordedFixture.RankedTaskFloorMet ||
		!recordedFixture.NoPromotionRequested ||
		recordedFixture.PromotionGranted ||
		recordedFixture.ClaimsAuthorityAdvance ||
		!recordedFixture.RSIRemainsDenied {
		t.Fatalf("v02 ranked export fixture lost final closure evidence binding: %#v", recordedFixture)
	}
	if len(recordedExport.Tasks) != 40 ||
		recordedExport.Tasks[0].ID != "feature-depth-next-wave-01" ||
		recordedExport.Tasks[39].ID != "feature-depth-next-wave-40" ||
		recordedExport.Tasks[0].EvidenceRefs[0] != sourceReadback ||
		recordedExport.Tasks[0].EvidenceRefs[1] != sourceAssertion ||
		recordedExport.SafeToExecute ||
		recordedExport.SchedulesWork ||
		recordedExport.ExecutesWork ||
		recordedExport.ApprovesWork ||
		recordedExport.MutatesRepositories {
		t.Fatalf("v02 ranked Feature Depth export must remain a 40-task intent-only planning artifact: %#v", recordedExport)
	}
}

func TestFeatureDepthWaveRankedExportRejectsCompletedFollowupSource(t *testing.T) {
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

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-followup-durability-v04"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	sourceAssertion := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/promoter_no_promotion.json"
	tmpExportPath := filepath.Join(t.TempDir(), "next-wave-feature-depth-recommendations.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-next-wave",
		"--mission-id", "ao-atlas-next-feature-depth-followup-durability-v05",
		"--source-evidence-root", sourceRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--min-tasks", "40",
		"--out", tmpExportPath,
	}, &out, &out)
	if code == 0 {
		t.Fatalf("completed Feature Depth follow-up source was exported again: %s", out.String())
	}
	for _, want := range []string{
		"feature depth recommendations saturated",
		"completed 40/40",
		"route to AO Atlas refactoring/strategy review",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("saturation refusal missing %q: %s", want, out.String())
		}
	}
	if _, err := os.Stat(tmpExportPath); !os.IsNotExist(err) {
		t.Fatalf("saturated export should not write a next-wave artifact, stat err=%v", err)
	}
}
