package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsSchemaRegistryPublishesTypedArtifactCoverage(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	registryPath := filepath.Join(t.TempDir(), "recommendation-evidence-schema-registry.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-registry",
		"--out", registryPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("schema-registry failed: %s", out.String())
	}
	for _, want := range []string{
		"status=ready",
		"schema_count=9",
		"typed_validator_coverage_complete=true",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("schema-registry output missing %q: %s", want, out.String())
		}
	}

	registry := mustLoadJSON[map[string]any](t, registryPath)
	if registry["schema"] != "ao.atlas.recommendation-evidence-schema-registry.v0.1" ||
		registry["status"] != "ready" ||
		registry["registry_purpose"] != "recommendation_control_plane_typed_artifact_coverage" ||
		registry["schema_count"] != float64(9) ||
		registry["typed_validator_coverage_complete"] != true ||
		registry["no_promotion_requested"] != true ||
		registry["promotion_granted"] != false ||
		registry["claims_authority_advance"] != false ||
		registry["rsi_remains_denied"] != true ||
		registry["safe_to_execute"] != false ||
		registry["schedules_work"] != false ||
		registry["executes_work"] != false ||
		registry["approves_work"] != false ||
		registry["mutates_repositories"] != false {
		t.Fatalf("schema registry did not emit safe coverage metadata: %#v", registry)
	}
	schemas, ok := registry["schemas"].([]any)
	if !ok || len(schemas) != 9 {
		t.Fatalf("expected 9 schema registry entries, got %#v", registry["schemas"])
	}
	wantSchemas := []struct {
		schema         string
		artifact       string
		command        string
		typedValidator string
	}{
		{"ao.mission.refactoring-recommendations.v0.1", "recommendation-refactoring-recommendations", "export-refactoring-wave", "typed:recommendation-refactoring-recommendations"},
		{"ao.atlas.recommendation-next-track-decision.v0.1", "recommendation-next-track-decision", "next-track", "typed:recommendation-next-track-decision"},
		{"ao.atlas.consumed-recommendation-ledger.v0.1", "consumed-recommendation-ledger", "consumed-ledger", "typed:consumed-recommendation-ledger"},
		{"ao.atlas.recommendation-track-registry.v0.1", "recommendation-track-registry", "track-registry", "typed:recommendation-track-registry"},
		{"ao.atlas.recommendation-command-run-ledger.v0.1", "recommendation-command-run-ledger", "run-ledger", "typed:recommendation-command-run-ledger"},
		{"ao.atlas.recommendation-evidence-validation-report.v0.1", "recommendation-evidence-validation-report", "validate-evidence", "typed:recommendation-evidence-validation-report"},
		{"ao.atlas.recommendation-final-response-gates.v0.1", "recommendation-final-response-gates", "final-response-gates", "typed:recommendation-final-response-gates"},
		{"ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1", "recommendation-evidence-schema-registry-coverage", "schema-registry-coverage", "typed:recommendation-evidence-schema-registry-coverage"},
		{"ao.atlas.schema-health-repair-prompt.v0.1", "schema-health-repair-prompt", "schema-health-repair-prompt", "typed:schema-health-repair-prompt"},
	}
	for i, want := range wantSchemas {
		entry, ok := schemas[i].(map[string]any)
		if !ok {
			t.Fatalf("schema entry %d is malformed: %#v", i, schemas[i])
		}
		if entry["schema"] != want.schema ||
			entry["artifact"] != want.artifact ||
			entry["command"] != want.command ||
			entry["typed_validator"] != want.typedValidator ||
			entry["status_field"] != "status" ||
			entry["safety_class"] != "planning_readback_no_execution" ||
			entry["planning_only"] != true {
			t.Fatalf("schema entry %d is wrong: %#v", i, entry)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(registryPath, "ao.atlas.recommendation-evidence-schema-registry.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-evidence-schema-registry" {
		t.Fatalf("expected typed recommendation schema registry validator, got %s", validator)
	}
}

func TestMissionRecommendationsSchemaRegistryUsesTypedEntryConstructors(t *testing.T) {
	registry, err := DefaultAtlasRecommendationEvidenceSchemaRegistry()
	if err != nil {
		t.Fatal(err)
	}
	entries := defaultAtlasRecommendationEvidenceSchemaRegistryEntries()
	if len(entries) != 9 {
		t.Fatalf("typed registry constructors drifted: got %d entries", len(entries))
	}
	if strings.Join(schemaRegistryEntryKeys(registry.Schemas), ",") != strings.Join(schemaRegistryEntryKeys(entries), ",") {
		t.Fatalf("default registry entries are not constructor-backed\ngot  %#v\nwant %#v", registry.Schemas, entries)
	}
	for _, entry := range entries {
		if entry.StatusField != "status" ||
			entry.SafetyClass != "planning_readback_no_execution" ||
			!entry.PlanningOnly {
			t.Fatalf("typed registry constructor lost planning-only defaults: %#v", entry)
		}
	}
}

func TestMissionRecommendationsSchemaRegistryContractsUseRecommendationEvidenceGroup(t *testing.T) {
	contracts := defaultAtlasRecommendationEvidenceSchemaContracts()
	contractSet := map[string]bool{}
	for _, contract := range contracts.ControlPlane {
		contractSet[contract] = true
	}
	registry, err := DefaultAtlasRecommendationEvidenceSchemaRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range registry.Schemas {
		if !contractSet[entry.Schema] {
			t.Fatalf("schema registry entry %s is not covered by grouped recommendation evidence contracts: %#v", entry.Schema, contracts)
		}
	}
	if !contractSet[AtlasRecommendationEvidenceSchemaRegistryContract] ||
		!contractSet[AtlasRecommendationEvidenceSchemaRegistryCoverageContract] {
		t.Fatalf("grouped recommendation evidence contracts must include registry and coverage contracts: %#v", contracts)
	}
}

func TestMissionRecommendationsSchemaRegistryBacksTypedValidatorLookup(t *testing.T) {
	registry, err := DefaultAtlasRecommendationEvidenceSchemaRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range registry.Schemas {
		validator, ok := recommendationControlPlaneTypedValidator(entry.Schema)
		if !ok {
			t.Fatalf("schema %s missing registry-backed typed validator lookup", entry.Schema)
		}
		if validator != entry.TypedValidator {
			t.Fatalf("schema %s typed validator drifted: got %s want %s", entry.Schema, validator, entry.TypedValidator)
		}
	}
	if validator, ok := recommendationControlPlaneTypedValidator("ao.atlas.not-recommendation-control-plane.v0.1"); ok || validator != "" {
		t.Fatalf("unknown schema should not resolve through recommendation control-plane lookup: %q %t", validator, ok)
	}
}

func TestMissionRecommendationsSchemaRegistryGoldenFixtures(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	registryPath := filepath.Join(t.TempDir(), "recommendation-evidence-schema-registry.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-registry",
		"--out", registryPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("schema-registry failed: %s", out.String())
	}

	commandOutputGolden := filepath.Join("internal", "atlas", "testdata", "recommendation_schema_registry_command_output.golden")
	wantOutput, err := os.ReadFile(commandOutputGolden)
	if err != nil {
		t.Fatalf("read command output golden fixture: %v", err)
	}
	gotOutput := normalizeSchemaRegistryCommandOutput(out.String(), registryPath)
	expectedOutput := normalizeSchemaRegistryCommandOutput(string(wantOutput), registryPath)
	if gotOutput != expectedOutput {
		t.Fatalf("schema registry command output golden drifted\ngot:\n%s\nwant:\n%s", gotOutput, expectedOutput)
	}

	registry := mustLoadJSON[AtlasRecommendationEvidenceSchemaRegistry](t, registryPath)
	gotBindings := make([]schemaRegistryTypedValidatorBindingGolden, 0, len(registry.Schemas))
	for _, entry := range registry.Schemas {
		gotBindings = append(gotBindings, schemaRegistryTypedValidatorBindingGolden{
			Schema:         entry.Schema,
			TypedValidator: entry.TypedValidator,
		})
	}
	bindingGolden := filepath.Join("internal", "atlas", "testdata", "recommendation_schema_registry_typed_validator_bindings.json")
	wantBindings := mustLoadJSON[[]schemaRegistryTypedValidatorBindingGolden](t, bindingGolden)
	assertSchemaRegistryTypedValidatorBindings(t, gotBindings, wantBindings)
}

type schemaRegistryTypedValidatorBindingGolden struct {
	Schema         string `json:"schema"`
	TypedValidator string `json:"typed_validator"`
}

func normalizeSchemaRegistryCommandOutput(output, registryPath string) string {
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, registryPath, "<out>")
	return strings.ReplaceAll(output, filepath.ToSlash(registryPath), "<out>")
}

func TestNormalizeSchemaRegistryCommandOutputNormalizesWindowsLineEndings(t *testing.T) {
	got := normalizeSchemaRegistryCommandOutput("recommendation_evidence_schema_registry=C:\\tmp\\registry.json\r\n", "C:\\tmp\\registry.json")
	want := "recommendation_evidence_schema_registry=<out>\n"
	if got != want {
		t.Fatalf("normalized schema registry command output drifted: got %q want %q", got, want)
	}
}

func assertSchemaRegistryTypedValidatorBindings(t *testing.T, got, want []schemaRegistryTypedValidatorBindingGolden) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("typed validator binding count drifted: got %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("typed validator binding %d drifted: got %#v want %#v", i, got[i], want[i])
		}
	}
}

func schemaRegistryEntryKeys(entries []AtlasRecommendationEvidenceSchemaRegistryEntry) []string {
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Schema+"|"+entry.Artifact+"|"+entry.Command+"|"+entry.TypedValidator)
	}
	return keys
}
