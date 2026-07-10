package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BPolicyHashMismatchRejectionFixtureRejectsMismatchedTickets(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-08")
	inputPath := filepath.Join(nodeDir, "policy-hash-mismatch-rejection-input.json")
	recordedPath := filepath.Join(nodeDir, "policy-hash-mismatch-rejection-fixture.json")
	outPath := filepath.Join(t.TempDir(), "policy-hash-mismatch-rejection-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "policy-hash-mismatch-rejection-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("policy-hash-mismatch-rejection-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=policy_hash_mismatches_rejected") ||
		!strings.Contains(out.String(), "rejected_cases=2") ||
		!strings.Contains(out.String(), "safe_to_accept=false") {
		t.Fatalf("policy hash mismatch output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("policy hash mismatch fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["safe_to_accept"] != false ||
		generated["all_mismatches_rejected"] != true ||
		generated["rejected_cases"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("policy hash mismatch fixture lost safety state: %#v", generated)
	}
}

func TestP0BPolicyHashMismatchRejectionFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-08", "policy-hash-mismatch-rejection-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.policy-hash-mismatch-rejection-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:policy-hash-mismatch-rejection-fixture" {
		t.Fatalf("expected typed policy hash mismatch rejection fixture validator, got %s", validator)
	}
}
