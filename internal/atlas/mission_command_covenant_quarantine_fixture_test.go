package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCommandCovenantQuarantineFixtureBlocksCommandAcceptanceOfRejectedTickets(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-05")
	inputPath := filepath.Join(nodeDir, "command-covenant-quarantine-input.json")
	recordedPath := filepath.Join(nodeDir, "command-covenant-quarantine-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-covenant-quarantine-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-covenant-quarantine-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-covenant-quarantine-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=rejected_paths_quarantined") ||
		!strings.Contains(out.String(), "quarantined_paths=2") ||
		!strings.Contains(out.String(), "safe_to_accept=false") {
		t.Fatalf("quarantine fixture output missing summary: %s", out.String())
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("quarantine fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["safe_to_accept"] != false ||
		generated["all_rejected_acceptance_paths_quarantined"] != true ||
		generated["quarantined_paths"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("quarantine fixture lost safety state: %#v", generated)
	}
}

func TestP0BCommandCovenantQuarantineFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-05", "command-covenant-quarantine-fixture.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-covenant-quarantine-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-covenant-quarantine-fixture" {
		t.Fatalf("expected typed command/covenant quarantine fixture validator, got %s", validator)
	}
}
