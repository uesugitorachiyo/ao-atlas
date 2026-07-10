package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCommandTicketBytePreservationFixtureBindsCanonicalTicketBytes(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-06")
	inputPath := filepath.Join(nodeDir, "command-ticket-byte-preservation-input.json")
	recordedPath := filepath.Join(nodeDir, "command-ticket-byte-preservation-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-ticket-byte-preservation-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-ticket-byte-preservation-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-ticket-byte-preservation-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=ticket_bytes_preserved") ||
		!strings.Contains(out.String(), "byte_preservation_passed=true") ||
		!strings.Contains(out.String(), "case_count=2") {
		t.Fatalf("ticket byte preservation output missing summary: %s", out.String())
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("ticket byte preservation fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["byte_preservation_passed"] != true ||
		generated["case_count"] != float64(2) ||
		generated["mismatched_cases"] != float64(0) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("ticket byte preservation fixture lost safety state: %#v", generated)
	}
}

func TestP0BCommandTicketBytePreservationFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-06", "command-ticket-byte-preservation-fixture.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-ticket-byte-preservation-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-ticket-byte-preservation-fixture" {
		t.Fatalf("expected typed command ticket byte preservation fixture validator, got %s", validator)
	}
}
