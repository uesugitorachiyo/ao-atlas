package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3BoundedExecutionPacketFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-20")
	recordedPath := filepath.Join(nodeDir, "bounded-execution-packet-fixture.json")
	outPath := filepath.Join(t.TempDir(), "bounded-execution-packet-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "bounded-execution-packet-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("bounded-execution-packet-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=bounded_execution_packet_ready",
		"isolated_worktree_required=true",
		"exact_digest_approval_required=true",
		"rollback_receipt_required=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("bounded execution packet output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("bounded execution packet fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("bounded execution packet fixture lost authority state: %#v", generated)
	}
}

func TestMonth3BoundedExecutionPacketFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-20", "bounded-execution-packet-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.bounded-execution-packet-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:bounded-execution-packet-fixture" {
		t.Fatalf("expected typed bounded execution packet validator, got %s", validator)
	}
}

func TestMonth3BoundedExecutionPacketFixtureRejectsMissingRollbackReceipt(t *testing.T) {
	fixture, err := BuildAtlasBoundedExecutionPacketFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.RollbackReceiptRequired = false
	if err := ValidateAtlasBoundedExecutionPacketFixture(fixture); err == nil || !strings.Contains(err.Error(), "rollback_receipt_required must be true") {
		t.Fatalf("expected missing rollback receipt rejection, got %v", err)
	}
}
