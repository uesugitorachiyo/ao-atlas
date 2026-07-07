package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveRunLinkDigestCheckVerifiesCompletedEvidencePacket(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceRunLinkPath := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-29", "run-link.json")
	nodeDir := filepath.Join(featureRoot, "nodes", "mission-recommendation-feature-depth-next-wave-30")
	recordedPath := filepath.Join(nodeDir, "run-link-digest-check.json")
	outPath := filepath.Join(t.TempDir(), "run-link-digest-check.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-link-digest-check",
		"--node-id", "mission-recommendation-feature-depth-next-wave-30",
		"--run-link", sourceRunLinkPath,
		"--evidence-root", root,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-link-digest-check command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasRunLinkDigestCheck](t, recordedPath)
	generated := mustLoadJSON[AtlasRunLinkDigestCheck](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("run-link digest check fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasRunLinkDigestCheck(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "run_link_digest_verified" ||
		recorded.TaskID != "mission-recommendation-feature-depth-next-wave-29-task" ||
		recorded.RunLinkStatus != "completed" ||
		!recorded.DigestMatches ||
		recorded.RecordedDigest != recorded.RecomputedDigest ||
		recorded.EvidenceCount == 0 ||
		recorded.SchemaBoundEvidenceCount != recorded.EvidenceCount ||
		len(recorded.MissingEvidence) != 0 ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("run-link digest check lost completed evidence packet binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.run-link-digest-check.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:run-link-digest-check" {
		t.Fatalf("expected typed run-link digest check validator, got %s", validator)
	}
}
