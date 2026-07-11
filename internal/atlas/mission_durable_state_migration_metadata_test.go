package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3DurableStateMigrationMetadata(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-23")
	recordedPath := filepath.Join(nodeDir, "durable-state-migration-metadata.json")
	outPath := filepath.Join(t.TempDir(), "durable-state-migration-metadata.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "durable-state-migration-metadata",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("durable-state-migration-metadata command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=durable_state_migration_metadata_ready",
		"current_version=1",
		"unknown_version_handling=fail_closed",
		"migration_count=1",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("durable state migration metadata output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("durable state migration metadata changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["unknown_version_handling"] != "fail_closed" ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("durable state migration metadata lost authority state: %#v", generated)
	}
}

func TestMonth3DurableStateMigrationMetadataUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-23", "durable-state-migration-metadata.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.durable-state-migration-metadata.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:durable-state-migration-metadata" {
		t.Fatalf("expected typed durable state migration metadata validator, got %s", validator)
	}
}

func TestMonth3DurableStateMigrationMetadataRejectsUnknownVersionOpen(t *testing.T) {
	metadata, err := BuildAtlasDurableStateMigrationMetadata()
	if err != nil {
		t.Fatal(err)
	}
	metadata.UnknownVersionHandling = "accept"
	if err := ValidateAtlasDurableStateMigrationMetadata(metadata); err == nil || !strings.Contains(err.Error(), "unknown_version_handling must be fail_closed") {
		t.Fatalf("expected unknown migration version fail-closed rejection, got %v", err)
	}
}
