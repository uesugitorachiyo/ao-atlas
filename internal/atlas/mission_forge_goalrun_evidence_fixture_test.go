package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ForgeGoalRunEvidenceFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-21")
	recordedPath := filepath.Join(nodeDir, "forge-goalrun-evidence-fixture.json")
	outPath := filepath.Join(t.TempDir(), "forge-goalrun-evidence-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "forge-goalrun-evidence-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("forge-goalrun-evidence-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=forge_goalrun_evidence_ready",
		"goalrun_start_required=true",
		"provider_execution_allowed=false",
		"terminal_receipt_required=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("Forge GoalRun evidence output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Forge GoalRun evidence fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["provider_execution_allowed"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("Forge GoalRun evidence fixture lost authority state: %#v", generated)
	}
}

func TestMonth3ForgeGoalRunEvidenceFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-21", "forge-goalrun-evidence-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.forge-goalrun-evidence-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:forge-goalrun-evidence-fixture" {
		t.Fatalf("expected typed Forge GoalRun evidence validator, got %s", validator)
	}
}

func TestMonth3ForgeGoalRunEvidenceFixtureRejectsProviderExecution(t *testing.T) {
	fixture, err := BuildAtlasForgeGoalRunEvidenceFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.ProviderExecutionAllowed = true
	if err := ValidateAtlasForgeGoalRunEvidenceFixture(fixture); err == nil || !strings.Contains(err.Error(), "provider_execution_allowed must be false") {
		t.Fatalf("expected provider execution rejection, got %v", err)
	}
}
