package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWaveScopedPublicSafetyScanCoversEvidenceAndPromptArtifacts(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-22")
	sourceScope := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-21")
	recordedPath := filepath.Join(nodeDir, "scoped-public-safety-scan.json")
	outPath := filepath.Join(t.TempDir(), "scoped-public-safety-scan.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "scoped-public-safety-scan",
		"--node-id", "mission-recommendation-feature-depth-next-wave-22",
		"--scope", sourceScope,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("scoped-public-safety-scan command failed: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("scoped public safety scan fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["status"] != "passed" ||
		recorded["public_safety_scan_passed"] != true ||
		recorded["unsafe_match_count"].(float64) != 0 ||
		recorded["changed_evidence_files"].(float64) < 1 ||
		recorded["changed_prompt_artifacts"].(float64) < 1 ||
		recorded["rsi_remains_denied"] != true {
		t.Fatalf("scoped public safety scan lost coverage or safety state: %#v", recorded)
	}
}

func TestFeatureDepthWaveV02ScopedPublicSafetyScanCoversEvidenceAndPromptArtifacts(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-22")
	sourceScope := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-21")
	recordedPath := filepath.Join(nodeDir, "scoped-public-safety-scan.json")
	outPath := filepath.Join(t.TempDir(), "scoped-public-safety-scan.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "scoped-public-safety-scan",
		"--node-id", "mission-recommendation-feature-depth-next-wave-22",
		"--scope", sourceScope,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("scoped-public-safety-scan command failed: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 scoped public safety scan fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["status"] != "passed" ||
		recorded["public_safety_scan_passed"] != true ||
		recorded["unsafe_match_count"].(float64) != 0 ||
		recorded["changed_evidence_files"].(float64) < 1 ||
		recorded["changed_prompt_artifacts"].(float64) < 1 ||
		recorded["rsi_remains_denied"] != true {
		t.Fatalf("v02 scoped public safety scan lost coverage or safety state: %#v", recorded)
	}
}

func TestFeatureDepthWaveScopedPublicSafetyScanUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-22", "scoped-public-safety-scan.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.scoped-public-safety-scan.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:scoped-public-safety-scan" {
		t.Fatalf("expected typed scoped public safety scan validator, got %s", validator)
	}
}
