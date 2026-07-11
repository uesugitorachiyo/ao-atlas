package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3NonAODryRunReplayBindingConsumesFixtureAndTerminalBinding(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-04-non-ao-dry-run-replay")
	recordedPath := filepath.Join(nodeDir, "month3-non-ao-dry-run-replay.json")
	outPath := filepath.Join(t.TempDir(), "month3-non-ao-dry-run-replay.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-non-ao-dry-run-replay",
		"--node-id", "mission-recommendation-month3-final-closure-04-non-ao-dry-run-replay",
		"--source-fixture", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-37", "non-ao-replay-binding-fixture.json"),
		"--terminal-binding", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-03-terminal-digest-binding", "month3-terminal-digest-binding.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-non-ao-dry-run-replay command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3NonAODryRunReplayBinding](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3NonAODryRunReplayBinding](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 non-AO dry-run replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3NonAODryRunReplayBinding(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "non_ao_dry_run_replay_bound" ||
		!recorded.FixtureOnlyExecutionEvidence ||
		!recorded.TinyNonAORepo ||
		!recorded.ReviewedPREvidence ||
		!recorded.ObserverReadbackBound ||
		!recorded.NoPromotionBoundary ||
		recorded.PromotionRequested ||
		recorded.LiveProviderCalls ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied ||
		!recorded.TerminalDigestBindingBound ||
		!recorded.TerminalFinalResponseAllowed {
		t.Fatalf("non-AO dry-run replay binding lost safety state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-non-ao-dry-run-replay.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-non-ao-dry-run-replay" {
		t.Fatalf("expected typed Month 3 non-AO dry-run replay validator, got %s", validator)
	}
}
