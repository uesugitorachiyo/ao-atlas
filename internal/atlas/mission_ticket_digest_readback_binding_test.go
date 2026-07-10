package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BTicketDigestReadbackBindingFixtureMatchesCommandAndCovenant(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-07")
	inputPath := filepath.Join(nodeDir, "ticket-digest-readback-binding-input.json")
	recordedPath := filepath.Join(nodeDir, "ticket-digest-readback-binding-fixture.json")
	outPath := filepath.Join(t.TempDir(), "ticket-digest-readback-binding-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "ticket-digest-readback-binding-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("ticket-digest-readback-binding-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=ticket_digest_readbacks_bound") ||
		!strings.Contains(out.String(), "digest_binding_passed=true") ||
		!strings.Contains(out.String(), "case_count=2") {
		t.Fatalf("ticket digest binding output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("ticket digest readback binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["digest_binding_passed"] != true ||
		generated["mismatched_cases"] != float64(0) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("ticket digest binding fixture lost safety state: %#v", generated)
	}
}

func TestP0BTicketDigestReadbackBindingFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-07", "ticket-digest-readback-binding-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.ticket-digest-readback-binding-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:ticket-digest-readback-binding-fixture" {
		t.Fatalf("expected typed ticket digest readback binding fixture validator, got %s", validator)
	}
}
