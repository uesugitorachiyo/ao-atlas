package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3TerminalDigestBindingMatchesReadbackAndMatrix(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-03-terminal-digest-binding")
	recordedPath := filepath.Join(nodeDir, "month3-terminal-digest-binding.json")
	outPath := filepath.Join(t.TempDir(), "month3-terminal-digest-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-terminal-digest-binding",
		"--node-id", "mission-recommendation-month3-final-closure-03-terminal-digest-binding",
		"--source-readback", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "recommendation-readback-after.json"),
		"--readiness-matrix", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "golden-path-readiness-matrix.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-terminal-digest-binding command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3TerminalDigestBinding](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3TerminalDigestBinding](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 terminal digest binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3TerminalDigestBinding(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "terminal_digest_binding_ready" ||
		recorded.ReadbackCompletedNodes != 40 ||
		recorded.MatrixCompletedNodes != 40 ||
		recorded.ReadbackReadyNodes != 0 ||
		recorded.MatrixReadyNodes != 0 ||
		!recorded.NodeCountsMatch ||
		!recorded.FinalResponseAllowed ||
		recorded.MatrixRecommendationCount != 40 ||
		recorded.PromotionRequested ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork {
		t.Fatalf("terminal digest binding lost terminal state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-terminal-digest-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-terminal-digest-binding" {
		t.Fatalf("expected typed Month 3 terminal digest binding validator, got %s", validator)
	}
}
