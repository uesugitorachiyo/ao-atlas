package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3ProviderModelProvenanceCoversEveryModelBackedRunRecord(t *testing.T) {
	root := repoRoot(t)
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-12-provider-model-provenance")
	recordedPath := filepath.Join(nodeDir, "month3-provider-model-provenance.json")
	outPath := filepath.Join(t.TempDir(), "month3-provider-model-provenance.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-provider-model-provenance",
		"--node-id", "mission-recommendation-month3-final-closure-12-provider-model-provenance",
		"--source-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-11-restart-resume-soak", "recommendation-readback-after.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-provider-model-provenance command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3ProviderModelProvenance](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3ProviderModelProvenance](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 provider model provenance fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3ProviderModelProvenance(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "provider_model_provenance_ready" ||
		recorded.RunRecordCount != 4 ||
		!recorded.EveryRunHasProvider ||
		!recorded.EveryRunHasModel ||
		!recorded.EveryRunHasModelClass ||
		recorded.LiveProviderCallCount != 0 ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("provider/model provenance lost safe fixture state: %#v", recorded)
	}
	for _, run := range recorded.RunRecords {
		if run.Provider == "" || run.Model == "" || run.ModelClass == "" || run.LiveProviderCall {
			t.Fatalf("run record missing provenance or records live provider call: %#v", run)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-provider-model-provenance.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-provider-model-provenance" {
		t.Fatalf("expected typed Month 3 provider model provenance validator, got %s", validator)
	}
}
