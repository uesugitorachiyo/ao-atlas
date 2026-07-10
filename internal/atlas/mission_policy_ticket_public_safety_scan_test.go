package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BPolicyTicketPublicSafetyScanRecordsAllowedClaims(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-17")
	inputPath := filepath.Join(nodeDir, "policy-ticket-public-safety-scan-input.json")
	recordedPath := filepath.Join(nodeDir, "policy-ticket-public-safety-scan.json")
	outPath := filepath.Join(t.TempDir(), "policy-ticket-public-safety-scan.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "policy-ticket-public-safety-scan",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("policy-ticket-public-safety-scan command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=passed_policy_ticket_public_safety_scan") ||
		!strings.Contains(out.String(), "unsafe_claims_found=0") ||
		!strings.Contains(out.String(), "claim_count=4") {
		t.Fatalf("policy ticket public-safety output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("policy ticket public-safety scan changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["unsafe_claims_found"] != float64(0) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("policy ticket public-safety scan lost safety state: %#v", generated)
	}
}

func TestP0BPolicyTicketPublicSafetyScanUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-17", "policy-ticket-public-safety-scan.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.policy-ticket-public-safety-scan.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:policy-ticket-public-safety-scan" {
		t.Fatalf("expected typed policy ticket public-safety scan validator, got %s", validator)
	}
}
