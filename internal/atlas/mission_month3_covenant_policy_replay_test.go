package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FinalClosureCovenantPolicyReplayFixtureRejectsStaleTickets(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-23-covenant-policy-replay")
	inputPath := filepath.Join(nodeDir, "policy-version-replay-rejection-input.json")
	recordedPath := filepath.Join(nodeDir, "policy-version-replay-rejection-fixture.json")
	outPath := filepath.Join(t.TempDir(), "policy-version-replay-rejection-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "policy-version-replay-rejection-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("policy-version-replay-rejection-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=stale_policy_versions_rejected") ||
		!strings.Contains(out.String(), "rejected_cases=3") ||
		!strings.Contains(out.String(), "safe_to_accept=false") {
		t.Fatalf("policy replay output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("policy replay fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["safe_to_accept"] != false ||
		generated["all_stale_versions_rejected"] != true ||
		generated["rejected_cases"] != float64(3) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("policy replay fixture lost safety state: %#v", generated)
	}
}

func TestMonth3FinalClosureCovenantPolicyReplayFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-23-covenant-policy-replay", "policy-version-replay-rejection-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.policy-version-replay-rejection-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:policy-version-replay-rejection-fixture" {
		t.Fatalf("expected typed policy version replay rejection fixture validator, got %s", validator)
	}
}
