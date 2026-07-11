package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3EvidenceExternalizationPlanBindsContentAddressedManifest(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-08-evidence-externalization")
	recordedPath := filepath.Join(nodeDir, "month3-evidence-externalization-plan.json")
	outPath := filepath.Join(t.TempDir(), "month3-evidence-externalization-plan.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-evidence-externalization",
		"--node-id", "mission-recommendation-month3-final-closure-08-evidence-externalization",
		"--content-manifest", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-32", "content-addressed-evidence-manifest-fixture.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-evidence-externalization command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3EvidenceExternalizationPlan](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3EvidenceExternalizationPlan](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 evidence externalization plan changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3EvidenceExternalizationPlan(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "evidence_externalization_plan_ready" ||
		recorded.ExternalizedClassCount != 3 ||
		recorded.RetainedFixtureClassCount != 3 ||
		!recorded.ContentManifestBound ||
		!recorded.BulkEvidenceExternalized ||
		!recorded.SmallReplayableFixturesRetained ||
		!recorded.ContentAddressingRequired ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("evidence externalization plan lost safety state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-evidence-externalization-plan.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-evidence-externalization-plan" {
		t.Fatalf("expected typed Month 3 evidence externalization plan validator, got %s", validator)
	}
}
