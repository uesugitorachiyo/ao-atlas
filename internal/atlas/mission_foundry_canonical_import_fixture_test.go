package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FoundryCanonicalImportFixtureRejectsSchemaAliases(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-12")
	recordedPath := filepath.Join(nodeDir, "foundry-canonical-import-fixture.json")
	outPath := filepath.Join(t.TempDir(), "foundry-canonical-import-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-canonical-import-fixture",
		"--workgraph", filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-11", "workgraph-after.json"),
		"--expected-node", "mission-recommendation-month3-golden-path-12",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-canonical-import-fixture command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=canonical_import_ready",
		"accepted_canonical_import=true",
		"rejected_alias_count=3",
		"expected_node=mission-recommendation-month3-golden-path-12",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("foundry canonical import output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("foundry canonical import fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["accepted_canonical_import"] != true ||
		generated["canonical_workgraph_fields_consumed"] != true ||
		generated["schedules_work"] != false ||
		generated["executes_work"] != false ||
		generated["approves_work"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("foundry canonical import fixture lost safety state: %#v", generated)
	}
}

func TestMonth3FoundryCanonicalImportRejectsAmbiguousEnvelope(t *testing.T) {
	raw := []byte(`{"contract_version":"ao.atlas.foundry-import.v0.1","schema_version":"ao.atlas.foundry-import.v0.0","id":"x","workgraph_id":"wg","target_instance":"target","status":"ready_for_foundry_fixture_import","source_artifacts":[{"ref":"generated","digest":"sha256:1111111111111111111111111111111111111111111111111111111111111111"}],"tasks":[],"schedules_work":false,"executes_work":false,"approves_work":false}`)
	if err := ValidateFoundryImportCanonicalEnvelope(raw); err == nil || !strings.Contains(err.Error(), "schema_version alias") {
		t.Fatalf("expected schema_version alias rejection, got %v", err)
	}
}

func TestMonth3FoundryCanonicalImportFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-12", "foundry-canonical-import-fixture.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.foundry-canonical-import-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-canonical-import-fixture" {
		t.Fatalf("expected typed foundry canonical import validator, got %s", validator)
	}
}
