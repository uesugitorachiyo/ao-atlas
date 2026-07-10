package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCommandCompactRejectionReasonFixtureRendersCovenantNativeReasons(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-11")
	inputPath := filepath.Join(nodeDir, "command-compact-rejection-reason-input.json")
	recordedPath := filepath.Join(nodeDir, "command-compact-rejection-reason-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-compact-rejection-reason-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-compact-rejection-reason-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-compact-rejection-reason-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=command_compact_rejection_reasons_rendered") ||
		!strings.Contains(out.String(), "reasons_rendered=true") ||
		!strings.Contains(out.String(), "case_count=2") {
		t.Fatalf("command compact rejection reason output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("command compact rejection reason fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["reasons_rendered"] != true ||
		generated["case_count"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("command compact rejection reason fixture lost safety state: %#v", generated)
	}
}

func TestP0BCommandCompactRejectionReasonFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-11", "command-compact-rejection-reason-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-compact-rejection-reason-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-compact-rejection-reason-fixture" {
		t.Fatalf("expected typed command compact rejection reason fixture validator, got %s", validator)
	}
}
