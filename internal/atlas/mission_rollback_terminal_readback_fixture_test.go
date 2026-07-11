package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3RollbackTerminalReadbackFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-39")
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

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("rollback terminal readback fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["rollback_receipt_replayed"] != true ||
		generated["terminal_state"] != "rolled_back" ||
		generated["readbacks_agree"] != true ||
		generated["promotion_requested"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("rollback terminal fixture lost safety state: %#v", generated)
	}
}

func TestMonth3RollbackTerminalReadbackFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-39", "rollback-terminal-readback-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.rollback-terminal-readback-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:rollback-terminal-readback-fixture" {
		t.Fatalf("expected typed rollback terminal validator, got %s", validator)
	}
}

func TestMonth3RollbackTerminalReadbackFixtureRejectsDisagreement(t *testing.T) {
	fixture, err := BuildAtlasRollbackTerminalReadbackFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.ReadbacksAgree = false
	if err := ValidateAtlasRollbackTerminalReadbackFixture(fixture); err == nil || !strings.Contains(err.Error(), "readbacks_agree must be true") {
		t.Fatalf("expected readback disagreement rejection, got %v", err)
	}
}
