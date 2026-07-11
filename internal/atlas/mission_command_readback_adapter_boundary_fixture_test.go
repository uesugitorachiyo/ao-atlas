package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CommandReadbackAdapterBoundaryFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-29")
	recordedPath := filepath.Join(nodeDir, "command-readback-adapter-boundary-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-readback-adapter-boundary-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-readback-adapter-boundary-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-readback-adapter-boundary-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=command_readback_adapter_boundary_ready",
		"adapter_count=2",
		"duplicates_domain_decisions=false",
		"delegates_domain_decisions=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("command readback adapter boundary output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("command readback adapter boundary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["delegates_domain_decisions"] != true ||
		generated["duplicates_domain_decisions"] != false ||
		generated["executes_work"] != false ||
		generated["approves_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("command readback adapter boundary fixture lost authority state: %#v", generated)
	}
}

func TestMonth3CommandReadbackAdapterBoundaryFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-29", "command-readback-adapter-boundary-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-readback-adapter-boundary-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-readback-adapter-boundary-fixture" {
		t.Fatalf("expected typed command readback adapter boundary validator, got %s", validator)
	}
}

func TestMonth3CommandReadbackAdapterBoundaryFixtureRejectsDuplicatedDomainDecision(t *testing.T) {
	fixture, err := BuildAtlasCommandReadbackAdapterBoundaryFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.DuplicatesDomainDecisions = true
	if err := ValidateAtlasCommandReadbackAdapterBoundaryFixture(fixture); err == nil || !strings.Contains(err.Error(), "duplicates_domain_decisions must be false") {
		t.Fatalf("expected duplicated domain decision rejection, got %v", err)
	}
}
