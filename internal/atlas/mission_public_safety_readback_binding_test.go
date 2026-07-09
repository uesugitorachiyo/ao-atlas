package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWavePublicSafetyReadbackBindingBindsSentinelPassedStatus(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-21")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "recommendation-readback-after.json")
	sentinelPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "sentinel_public_safety.json")
	verificationPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "verification.json")
	recordedPath := filepath.Join(nodeDir, "public-safety-readback-binding.json")
	outPath := filepath.Join(t.TempDir(), "public-safety-readback-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "public-safety-readback-binding",
		"--readback", sourceReadback,
		"--sentinel", sentinelPath,
		"--verification", verificationPath,
		"--node-id", "mission-recommendation-feature-depth-next-wave-21",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("public-safety-readback-binding command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=bound") ||
		!strings.Contains(out.String(), "bound_public_safety_scan_status=passed") ||
		!strings.Contains(out.String(), "previous_public_safety_scan_status=required_pending_verification") {
		t.Fatalf("public-safety-readback-binding output missing binding state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("public safety readback binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["status"] != "bound" ||
		recorded["bound_public_safety_scan_status"] != "passed" ||
		recorded["previous_public_safety_scan_status"] != "required_pending_verification" ||
		recorded["ready_nodes_after_binding"].(float64) != 20 ||
		recorded["final_response_allowed_after_binding"] != false ||
		recorded["rsi_remains_denied"] != true {
		t.Fatalf("public safety binding fixture lost readback state: %#v", recorded)
	}
}

func TestFeatureDepthWaveV02PublicSafetyReadbackBindingPreservesAlreadyBoundStatus(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-21")
	sourceReadbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "recommendation-readback-after.json")
	sourceBindingPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "public-safety-readback-binding.json")
	recordedPath := filepath.Join(nodeDir, "public-safety-readback-binding-preservation.json")

	readback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)
	binding := mustLoadJSON[AtlasPublicSafetyReadbackBinding](t, sourceBindingPath)
	recorded := mustLoadJSON[map[string]any](t, recordedPath)

	if readback.PublicSafetyScanStatus != "passed" ||
		readback.CompletedNodes != 20 ||
		readback.ReadyNodes != 20 ||
		readback.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-21" ||
		readback.FinalResponseAllowed {
		t.Fatalf("v02 source readback must carry already-bound public safety continuation state: %#v", readback)
	}
	if binding.Status != "bound" ||
		binding.BoundPublicSafetyScanStatus != "passed" ||
		binding.PreviousPublicSafetyScanStatus != "required_pending_verification" ||
		binding.ReadyNodesAfterBinding != 20 ||
		binding.FinalResponseAllowedAfter ||
		!binding.RSIRemainsDenied {
		t.Fatalf("v02 source binding must prove passed public safety without final response: %#v", binding)
	}
	if recorded["status"] != "preserved" ||
		recorded["source_readback_public_safety_scan_status"] != "passed" ||
		recorded["source_binding_status"] != "bound" ||
		recorded["bound_public_safety_scan_status"] != "passed" ||
		recorded["ready_nodes_after_binding"].(float64) != 20 ||
		recorded["final_response_allowed"] != false ||
		recorded["rsi_remains_denied"] != true {
		t.Fatalf("v02 public safety preservation evidence lost already-bound state: %#v", recorded)
	}
}

func TestFeatureDepthWavePublicSafetyReadbackBindingUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-21", "public-safety-readback-binding.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.public-safety-readback-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:public-safety-readback-binding" {
		t.Fatalf("expected typed public safety readback binding validator, got %s", validator)
	}
}

func TestFeatureDepthWavePublicSafetyReadbackStatusOverridePreservesContinuation(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	wavePath := filepath.Join(waveRoot, "recommendation-wave.json")
	workgraphPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-20", "workgraph-after.json")
	wave := mustLoadJSON[AtlasRecommendationWave](t, wavePath)
	workgraph := mustLoadJSON[Workgraph](t, workgraphPath)

	readback, err := BuildAtlasRecommendationReadback(wave, workgraph, AtlasRecommendationReadbackOptions{
		WavePath:               wavePath,
		WorkgraphPath:          workgraphPath,
		EvidenceRoot:           "docs/evidence/ao-atlas-feature-depth-wave-v01",
		PublicSafetyScanStatus: "passed",
		CompletedAt:            "2026-07-07T05:40:00Z",
		ElapsedMinutes:         516,
		LeaseTimingMode:        "supervised",
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.PublicSafetyScanStatus != "passed" ||
		readback.CompletedNodes != 20 ||
		readback.ReadyNodes != 20 ||
		readback.FirstExecutableNode != "mission-recommendation-feature-depth-next-wave-21" ||
		readback.FinalResponseAllowed {
		t.Fatalf("public safety status override must preserve continuation state: %#v", readback)
	}
}
