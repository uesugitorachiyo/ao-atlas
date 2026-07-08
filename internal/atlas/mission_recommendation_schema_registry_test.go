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
		"schema_count=6",
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
		registry["schema_count"] != float64(6) ||
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
		t.Fatalf("schema registry did not publish safe coverage metadata: %#v", registry)
	}
	schemas, ok := registry["schemas"].([]any)
	if !ok || len(schemas) != 6 {
		t.Fatalf("expected 6 schema registry entries, got %#v", registry["schemas"])
	}
	wantSchemas := []struct {
		schema         string
		artifact       string
		command        string
		typedValidator string
	}{
		{"ao.atlas.recommendation-next-track-decision.v0.1", "recommendation-next-track-decision", "next-track", "typed:recommendation-next-track-decision"},
		{"ao.atlas.consumed-recommendation-ledger.v0.1", "consumed-recommendation-ledger", "consumed-ledger", "typed:consumed-recommendation-ledger"},
		{"ao.atlas.recommendation-track-registry.v0.1", "recommendation-track-registry", "track-registry", "typed:recommendation-track-registry"},
		{"ao.atlas.recommendation-command-run-ledger.v0.1", "recommendation-command-run-ledger", "run-ledger", "typed:recommendation-command-run-ledger"},
		{"ao.atlas.recommendation-final-response-gates.v0.1", "recommendation-final-response-gates", "final-response-gates", "typed:recommendation-final-response-gates"},
		{"ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1", "recommendation-evidence-schema-registry-coverage", "schema-registry-coverage", "typed:recommendation-evidence-schema-registry-coverage"},
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
