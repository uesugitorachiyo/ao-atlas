package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveSchemaValidatorDriftFixtureRecordsRegeneratedDirectoryChanges(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-08")
	sourceReport := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-06", "feature-depth-evidence-validation-report.json")
	targetReport := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-07", "feature-depth-evidence-validation-report.json")
	fixturePath := filepath.Join(nodeDir, "schema-validator-drift.json")

	drift, err := BuildAtlasSchemaValidatorDriftEvidence(sourceReport, targetReport)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasSchemaValidatorDriftEvidence(drift); err != nil {
		t.Fatal(err)
	}
	recorded := mustLoadJSON[AtlasSchemaValidatorDriftEvidence](t, fixturePath)
	if err := ValidateAtlasSchemaValidatorDriftEvidence(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(drift) != digestValue(recorded) {
		t.Fatalf("schema validator drift fixture changed\nwant %s\ngot  %s", digestValue(drift), digestValue(recorded))
	}
	if drift.Status != "recorded_no_unexpected_loss" ||
		drift.SourceNodeCount != 6 ||
		drift.TargetNodeCount != 7 ||
		drift.JSONFileDelta != 21 ||
		drift.TypedValidatorDelta != 12 ||
		drift.GenericSchemaDelta != 9 ||
		len(drift.LostSchemas) != 0 ||
		len(drift.LostValidators) != 0 ||
		drift.SchemaCountDeltas[RunLinkContract] != 1 ||
		drift.ValidatorCountDeltas["typed:run-link"] != 1 ||
		drift.ValidatorCountDeltas["generic:schema-marker"] != 9 ||
		drift.ClaimsAuthorityAdvance ||
		!drift.RSIRemainsDenied {
		t.Fatalf("schema validator drift must record regenerated fixture deltas without authority effects: %#v", drift)
	}
}

func TestMissionRecommendationsSchemaValidatorDriftCLIWritesDeterministicArtifact(t *testing.T) {
	root := repoRoot(t)
	sourceReport := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-06", "feature-depth-evidence-validation-report.json")
	targetReport := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-07", "feature-depth-evidence-validation-report.json")
	recorded := mustLoadJSON[AtlasSchemaValidatorDriftEvidence](t, filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-08", "schema-validator-drift.json"))
	outPath := filepath.Join(t.TempDir(), "schema-validator-drift.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "schema-validator-drift",
		"--source-report", sourceReport,
		"--target-report", targetReport,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("schema-validator-drift command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=recorded_no_unexpected_loss") ||
		!strings.Contains(out.String(), "typed_validator_delta=12") ||
		!strings.Contains(out.String(), "lost_validators=0") {
		t.Fatalf("schema-validator-drift output missing drift summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasSchemaValidatorDriftEvidence](t, outPath)
	if err := ValidateAtlasSchemaValidatorDriftEvidence(generated); err != nil {
		t.Fatal(err)
	}
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("CLI schema validator drift output changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
}
