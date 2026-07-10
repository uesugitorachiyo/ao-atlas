package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCovenantEvidenceDigestReadbackFixturePreservesRequiredDigests(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-10")
	inputPath := filepath.Join(nodeDir, "covenant-evidence-digest-readback-input.json")
	recordedPath := filepath.Join(nodeDir, "covenant-evidence-digest-readback-fixture.json")
	outPath := filepath.Join(t.TempDir(), "covenant-evidence-digest-readback-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "covenant-evidence-digest-readback-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("covenant-evidence-digest-readback-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=covenant_digest_readback_recorded") ||
		!strings.Contains(out.String(), "digest_readback_complete=true") ||
		!strings.Contains(out.String(), "case_count=2") {
		t.Fatalf("covenant digest readback output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("covenant digest readback fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["digest_readback_complete"] != true ||
		generated["case_count"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("covenant digest readback fixture lost safety state: %#v", generated)
	}
}

func TestP0BCovenantEvidenceDigestReadbackFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-10", "covenant-evidence-digest-readback-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.covenant-evidence-digest-readback-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:covenant-evidence-digest-readback-fixture" {
		t.Fatalf("expected typed covenant evidence digest readback fixture validator, got %s", validator)
	}
}
