package atlas

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestFeatureDepthWaveV02RunLinkDigestCheckVerifiesCompletedEvidencePacket(t *testing.T) {
	root := repoRoot(t)
	featureRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
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
		t.Fatalf("v02 run-link digest check fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
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
		t.Fatalf("v02 run-link digest check lost completed evidence packet binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.run-link-digest-check.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:run-link-digest-check" {
		t.Fatalf("expected typed run-link digest check validator, got %s", validator)
	}
}

func TestRunLinkDigestCheckVerifiesCompactFoundryEvidenceRecords(t *testing.T) {
	dir := t.TempDir()
	manifestPath := writeCompactFoundryEvidenceFixture(t, dir, false)
	link, err := BuildRunLink("compact-evidence-consumer-task", "completed", map[string]string{
		"foundry_compact": "compact/manifest.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	runLinkPath := filepath.Join(dir, "run-link.json")
	if err := WriteJSON(runLinkPath, link); err != nil {
		t.Fatal(err)
	}

	check, err := BuildAtlasRunLinkDigestCheck("compact-evidence-consumer-node", runLinkPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if manifestPath == "" || check.EvidenceCount != 2 || check.SchemaBoundEvidenceCount != 2 || len(check.EvidenceEntries) != 2 {
		t.Fatalf("compact manifest should verify as two logical evidence records, manifest=%s check=%#v", manifestPath, check)
	}
	if check.EvidenceEntries[0].Key != "foundry_compact:decision:covenant-local-allow" ||
		check.EvidenceEntries[1].Key != "foundry_compact:evidence:verification-output" {
		t.Fatalf("compact entries should preserve deterministic record IDs: %#v", check.EvidenceEntries)
	}
}

func TestRunLinkDigestCheckRejectsCorruptCompactFoundryEvidence(t *testing.T) {
	dir := t.TempDir()
	writeCompactFoundryEvidenceFixture(t, dir, true)
	link, err := BuildRunLink("compact-evidence-corrupt-task", "completed", map[string]string{
		"foundry_compact": "compact/manifest.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	runLinkPath := filepath.Join(dir, "run-link.json")
	if err := WriteJSON(runLinkPath, link); err != nil {
		t.Fatal(err)
	}

	_, err = BuildAtlasRunLinkDigestCheck("compact-evidence-corrupt-node", runLinkPath, dir)
	if err == nil || !strings.Contains(err.Error(), "duplicate record_id evidence:verification-output") {
		t.Fatalf("expected duplicate compact record rejection, got %v", err)
	}
}

func writeCompactFoundryEvidenceFixture(t *testing.T, root string, duplicate bool) string {
	t.Helper()
	compactDir := filepath.Join(root, "compact")
	chunkDir := filepath.Join(compactDir, "chunks")
	if err := os.MkdirAll(chunkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	recordID := "decision:covenant-local-allow"
	if duplicate {
		recordID = "evidence:verification-output"
	}
	chunkBody := []byte(
		`{"kind":"decision","payload":{"decision":"allow"},"record_id":"` + recordID + "\"}\n" +
			`{"kind":"evidence","payload":{"status":"passed"},"record_id":"evidence:verification-output"}` + "\n",
	)
	chunkPath := filepath.Join(chunkDir, "000001.jsonl")
	if err := os.WriteFile(chunkPath, chunkBody, 0o644); err != nil {
		t.Fatal(err)
	}
	chunkSum := sha256.Sum256(chunkBody)
	manifest := map[string]any{
		"schema":             "ao.foundry.compact-evidence-manifest.v0.1",
		"format_version":     "v0.1",
		"source_run":         "runs/example.foundry-run.json",
		"chunk_order":        []string{"chunks/000001.jsonl"},
		"total_record_count": 2,
		"chunks": []map[string]any{{
			"path":            "chunks/000001.jsonl",
			"record_count":    2,
			"first_record_id": recordID,
			"last_record_id":  "evidence:verification-output",
			"sha256":          fmt.Sprintf("%x", chunkSum[:]),
		}},
		"lookup": map[string]any{
			"strategy": "ordered_chunk_ranges",
			"ranges": []map[string]any{{
				"chunk":           "chunks/000001.jsonl",
				"first_record_id": recordID,
				"last_record_id":  "evidence:verification-output",
			}},
		},
	}
	digestable := map[string]any{}
	for key, value := range manifest {
		digestable[key] = value
	}
	digestable["manifest_digest"] = ""
	body, err := json.Marshal(digestable)
	if err != nil {
		t.Fatal(err)
	}
	manifestSum := sha256.Sum256(body)
	manifest["manifest_digest"] = fmt.Sprintf("%x", manifestSum[:])
	manifestPath := filepath.Join(compactDir, "manifest.json")
	if err := WriteJSON(manifestPath, manifest); err != nil {
		t.Fatal(err)
	}
	return manifestPath
}
