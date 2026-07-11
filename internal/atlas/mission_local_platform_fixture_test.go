package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3LocalPlatformFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-36")
	recordedPath := filepath.Join(nodeDir, "local-platform-fixture.json")
	outPath := filepath.Join(t.TempDir(), "local-platform-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "local-platform-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("local-platform-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=local_platform_fixture_ready",
		"platform_count=3",
		"line_ending_modes=2",
		"live_provider_calls=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("local platform output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("local platform fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["deterministic_install"] != true ||
		generated["path_normalization_checked"] != true ||
		generated["rollback_receipt_required"] != true ||
		generated["live_provider_calls"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("local platform fixture lost safety state: %#v", generated)
	}
}

func TestMonth3LocalPlatformFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-36", "local-platform-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.local-platform-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:local-platform-fixture" {
		t.Fatalf("expected typed local platform validator, got %s", validator)
	}
}

func TestMonth3LocalPlatformFixtureRejectsMissingWindowsPathMode(t *testing.T) {
	fixture, err := BuildAtlasLocalPlatformFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.PathModes = []string{"posix"}
	if err := ValidateAtlasLocalPlatformFixture(fixture); err == nil || !strings.Contains(err.Error(), "path_modes must include windows") {
		t.Fatalf("expected windows path mode rejection, got %v", err)
	}
}
