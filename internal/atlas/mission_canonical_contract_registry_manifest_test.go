package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CanonicalContractRegistryManifestRecordsOwnersLifecycleDigestsAndConsumers(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-8")
	recordedPath := filepath.Join(nodeDir, "canonical-contract-registry-manifest.json")
	outPath := filepath.Join(t.TempDir(), "canonical-contract-registry-manifest.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "canonical-contract-registry-manifest",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("canonical-contract-registry-manifest command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=canonical_contract_registry_ready",
		"contract_count=4",
		"gate_critical_count=4",
		"consumer_count=9",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("canonical contract registry manifest output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("canonical contract registry manifest changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["contract_count"] != float64(4) ||
		generated["gate_critical_count"] != float64(4) ||
		generated["consumer_count"] != float64(9) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("canonical contract registry manifest lost safety state: %#v", generated)
	}
}

func TestMonth3CanonicalContractRegistryManifestUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-8", "canonical-contract-registry-manifest.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.canonical-contract-registry-manifest.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:canonical-contract-registry-manifest" {
		t.Fatalf("expected typed canonical contract registry manifest validator, got %s", validator)
	}
}
