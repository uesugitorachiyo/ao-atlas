package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3BoundedSignerContractFixtureRecordsRotationAndRevocationWithoutAuthorityEffects(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-7")
	recordedPath := filepath.Join(nodeDir, "bounded-signer-contract-fixture.json")
	outPath := filepath.Join(t.TempDir(), "bounded-signer-contract-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "bounded-signer-contract-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("bounded-signer-contract-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=bounded_signer_contract_ready",
		"signer_count=2",
		"rotation_boundary=overlap_required",
		"revocation_boundary=deny_on_or_after_revoked_at",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("bounded signer contract output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("bounded signer contract fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["signer_count"] != float64(2) ||
		generated["rotation_boundary"] != "overlap_required" ||
		generated["revocation_boundary"] != "deny_on_or_after_revoked_at" ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("bounded signer contract fixture lost safety state: %#v", generated)
	}
}

func TestMonth3BoundedSignerContractFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-7", "bounded-signer-contract-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.bounded-signer-contract-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:bounded-signer-contract-fixture" {
		t.Fatalf("expected typed bounded signer contract fixture validator, got %s", validator)
	}
}
