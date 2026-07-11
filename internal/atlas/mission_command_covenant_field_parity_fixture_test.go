package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3CommandCovenantFieldParityFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-13")
	recordedPath := filepath.Join(nodeDir, "command-covenant-field-parity-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-covenant-field-parity-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-covenant-field-parity-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-covenant-field-parity-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=field_parity_verified",
		"policy_field_count=4",
		"approval_field_count=4",
		"rejected_extra_field_count=3",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("field parity output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("command/covenant field parity fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["command_accepts_only_covenant_fields"] != true ||
		generated["covenant_validates_same_fields"] != true ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("field parity fixture lost parity or authority state: %#v", generated)
	}
}

func TestMonth3CommandCovenantFieldParityFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-13", "command-covenant-field-parity-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-covenant-field-parity-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-covenant-field-parity-fixture" {
		t.Fatalf("expected typed command/covenant field parity validator, got %s", validator)
	}
}
