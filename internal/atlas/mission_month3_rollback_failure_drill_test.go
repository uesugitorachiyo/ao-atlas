package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosureRollbackFailureDrillFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-28-rollback-failure-drill")
	recordedPath := filepath.Join(nodeDir, "rollback-terminal-readback-fixture.json")
	outPath := filepath.Join(t.TempDir(), "rollback-terminal-readback-fixture.json")

	var out bytes.Buffer
	code := Run([]string{"mission", "recommendations", "rollback-terminal-readback-fixture", "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("rollback-terminal-readback-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=rollback_terminal_readbacks_agree",
		"rollback_receipt_replayed=true",
		"readback_agreement_count=4",
		"terminal_state=rolled_back",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("rollback terminal output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[AtlasRollbackTerminalReadbackFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasRollbackTerminalReadbackFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("rollback failure drill fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded.TerminalState != "rolled_back" ||
		!recorded.RollbackReceiptReplayed ||
		!recorded.ReadbacksAgree ||
		recorded.PromotionRequested ||
		recorded.LiveProviderCalls ||
		recorded.ExecutesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("rollback failure drill fixture lost terminal readback or safety state: %#v", recorded)
	}
}

func TestMonth3FinalClosureRollbackFailureDrillFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-28-rollback-failure-drill", "rollback-terminal-readback-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.rollback-terminal-readback-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:rollback-terminal-readback-fixture" {
		t.Fatalf("expected typed rollback terminal validator, got %s", validator)
	}
}
