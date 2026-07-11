package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3BlueprintCanonicalPreservationFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-11")
	recordedPath := filepath.Join(nodeDir, "blueprint-canonical-preservation-fixture.json")
	outPath := filepath.Join(t.TempDir(), "blueprint-canonical-preservation-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "blueprint-canonical-preservation-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("blueprint-canonical-preservation-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=preserved",
		"workspace_root_ref=examples/valid/blueprint-import-low-risk-code/blueprint-pack",
		"digest_preserved=true",
		"canonical_file_count=7",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("blueprint canonical preservation output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("blueprint canonical preservation fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["digest_preserved"] != true ||
		generated["canonical_bytes_preserved"] != true ||
		generated["schedules_work"] != false ||
		generated["executes_work"] != false ||
		generated["approves_work"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("blueprint canonical preservation fixture lost safety state: %#v", generated)
	}
}

func TestMonth3BlueprintCanonicalPreservationFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-11", "blueprint-canonical-preservation-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.blueprint-canonical-preservation-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:blueprint-canonical-preservation-fixture" {
		t.Fatalf("expected typed blueprint canonical preservation validator, got %s", validator)
	}
}
