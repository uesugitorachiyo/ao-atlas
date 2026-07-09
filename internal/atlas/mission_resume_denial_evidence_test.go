package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveResumeDenialEvidenceBindsReadyWorkBlocker(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-19", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "resume-denial-evidence.json")
	outPath := filepath.Join(t.TempDir(), "resume-denial-evidence.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "resume-denial-evidence",
		"--readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("resume-denial-evidence command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=denied_ready_work_remains") ||
		!strings.Contains(out.String(), "ready_nodes=21") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("resume-denial-evidence output missing denial state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("resume denial evidence fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["status"] != "denied_ready_work_remains" ||
		recorded["final_response_allowed"] != false ||
		recorded["refuses_final_response"] != true ||
		recorded["ready_nodes"].(float64) != 21 ||
		recorded["current_next_executable_node"] != "mission-recommendation-feature-depth-next-wave-20" ||
		recorded["final_response_denial_gate"] != "deny_ready_nodes_or_exact_next_action_remain" ||
		recorded["continuation_contract_reason"] != "ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("resume denial fixture lost ready-work denial state: %#v", recorded)
	}
	assertions, ok := recorded["denial_assertions"].([]any)
	if !ok || len(assertions) < 5 {
		t.Fatalf("resume denial fixture missing denial assertions: %#v", recorded["denial_assertions"])
	}
}

func TestFeatureDepthWaveV02ResumeDenialEvidenceBindsReadyWorkBlocker(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-19", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "resume-denial-evidence.json")
	outPath := filepath.Join(t.TempDir(), "resume-denial-evidence.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "resume-denial-evidence",
		"--readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("resume-denial-evidence command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=denied_ready_work_remains") ||
		!strings.Contains(out.String(), "ready_nodes=21") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("v02 resume-denial-evidence output missing denial state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 resume denial evidence fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["status"] != "denied_ready_work_remains" ||
		recorded["final_response_allowed"] != false ||
		recorded["refuses_final_response"] != true ||
		recorded["ready_nodes"].(float64) != 21 ||
		recorded["current_next_executable_node"] != "mission-recommendation-feature-depth-next-wave-20" ||
		recorded["final_response_denial_gate"] != "deny_ready_nodes_or_exact_next_action_remain" ||
		recorded["continuation_contract_reason"] != "ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("v02 resume denial fixture lost ready-work denial state: %#v", recorded)
	}
}

func TestFeatureDepthWaveResumeDenialEvidenceUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-20", "resume-denial-evidence.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.resume-denial-evidence.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:resume-denial-evidence" {
		t.Fatalf("expected typed resume denial evidence validator, got %s", validator)
	}
}
