package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsTrackRegistryPublishesSafeRoutingPolicy(t *testing.T) {
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

	registryPath := filepath.Join(t.TempDir(), "recommendation-track-registry.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "track-registry",
		"--out", registryPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("track-registry failed: %s", out.String())
	}
	for _, want := range []string{
		"status=ready",
		"default_track=feature_depth",
		"saturated_feature_depth_next_track=refactoring",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("track-registry output missing %q: %s", want, out.String())
		}
	}

	registry := mustLoadJSON[map[string]any](t, registryPath)
	if registry["schema"] != "ao.atlas.recommendation-track-registry.v0.1" ||
		registry["status"] != "ready" ||
		registry["default_track"] != "feature_depth" ||
		registry["saturated_feature_depth_next_track"] != "refactoring" ||
		registry["rsi_track_status"] != "boundary_hardening_only_denied" ||
		registry["no_promotion_requested"] != true ||
		registry["promotion_granted"] != false ||
		registry["claims_authority_advance"] != false ||
		registry["rsi_remains_denied"] != true ||
		registry["safe_to_execute"] != false ||
		registry["schedules_work"] != false ||
		registry["executes_work"] != false ||
		registry["approves_work"] != false ||
		registry["mutates_repositories"] != false {
		t.Fatalf("track registry did not publish safe routing policy: %#v", registry)
	}
	priority, ok := registry["priority_order"].([]any)
	if !ok || len(priority) != 3 ||
		priority[0] != "refactoring" ||
		priority[1] != "feature_depth" ||
		priority[2] != "rsi_boundary_hardening" {
		t.Fatalf("track registry priority order is wrong: %#v", registry["priority_order"])
	}
	tracks, ok := registry["tracks"].([]any)
	if !ok || len(tracks) != 3 {
		t.Fatalf("track registry should include 3 tracks: %#v", registry["tracks"])
	}
	validator, err := validateRecommendationEvidenceTypedFile(registryPath, "ao.atlas.recommendation-track-registry.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-track-registry" {
		t.Fatalf("expected typed recommendation track registry validator, got %s", validator)
	}

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-followup-durability-v04"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	decision, err := BuildAtlasRecommendationNextTrackDecision(sourceRoot, sourceReadback)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(decision.PriorityOrder, ",") != "refactoring,feature_depth,rsi_boundary_hardening" {
		t.Fatalf("next-track did not use registry priority order: %#v", decision.PriorityOrder)
	}
}
