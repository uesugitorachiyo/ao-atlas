package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3SentinelHostedCIWorkflowFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-15")
	recordedPath := filepath.Join(nodeDir, "sentinel-hosted-ci-workflow-fixture.json")
	outPath := filepath.Join(t.TempDir(), "sentinel-hosted-ci-workflow-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "sentinel-hosted-ci-workflow-fixture",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("sentinel-hosted-ci-workflow-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=least_privilege_workflow_ready",
		"permissions=contents:read",
		"deterministic_fixture_commands=3",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("sentinel CI output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Sentinel hosted CI workflow fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["permissions_read_only"] != true ||
		generated["uses_provider_credentials"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("Sentinel CI fixture lost least-privilege or authority state: %#v", generated)
	}
}

func TestMonth3SentinelHostedCIWorkflowFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-15", "sentinel-hosted-ci-workflow-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.sentinel-hosted-ci-workflow-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:sentinel-hosted-ci-workflow-fixture" {
		t.Fatalf("expected typed Sentinel hosted CI workflow validator, got %s", validator)
	}
}

func TestMonth3SentinelHostedCIWorkflowFixtureRejectsWritePermission(t *testing.T) {
	fixture, err := BuildAtlasSentinelHostedCIWorkflowFixture()
	if err != nil {
		t.Fatal(err)
	}
	fixture.Permissions = "contents:write"
	fixture.PermissionsReadOnly = false
	if err := ValidateAtlasSentinelHostedCIWorkflowFixture(fixture); err == nil || !strings.Contains(err.Error(), "permissions must be contents:read") {
		t.Fatalf("expected write permission rejection, got %v", err)
	}
}
