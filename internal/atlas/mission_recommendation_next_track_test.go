package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsNextTrackRoutesCompletedFeatureDepthToRefactoring(t *testing.T) {
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

	sourceRoot := "docs/evidence/ao-atlas-feature-depth-followup-durability-v04"
	sourceReadback := sourceRoot + "/nodes/mission-recommendation-feature-depth-next-wave-40/recommendation-readback-after.json"
	outPath := filepath.Join(t.TempDir(), "next-track-decision.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "next-track",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("next-track failed: %s", out.String())
	}
	for _, want := range []string{
		"status=routed",
		"current_track=feature_depth",
		"recommended_track=refactoring",
		"rsi_track_status=boundary_hardening_only_denied",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("next-track output missing %q: %s", want, out.String())
		}
	}

	decision := mustLoadJSON[map[string]any](t, outPath)
	if decision["schema"] != "ao.atlas.recommendation-next-track-decision.v0.1" ||
		decision["status"] != "routed" ||
		decision["source_evidence_root"] != sourceRoot ||
		decision["source_readback_path"] != sourceReadback ||
		decision["current_track"] != "feature_depth" ||
		decision["current_track_status"] != "completed_saturated" ||
		decision["recommended_track"] != "refactoring" ||
		decision["feature_depth_status"] != "saturated_completed" ||
		decision["refactoring_status"] != "recommended_next" ||
		decision["rsi_track_status"] != "boundary_hardening_only_denied" ||
		decision["exact_next_action"] != "Start AO Atlas refactoring wave for recommendation routing, consumed-task ledger, final-response gates, and non-self-referential handoffs." ||
		decision["final_response_allowed_observed"] != true ||
		decision["no_promotion_requested"] != true ||
		decision["promotion_granted"] != false ||
		decision["claims_authority_advance"] != false ||
		decision["rsi_remains_denied"] != true ||
		decision["safe_to_execute"] != false ||
		decision["schedules_work"] != false ||
		decision["executes_work"] != false ||
		decision["approves_work"] != false ||
		decision["mutates_repositories"] != false {
		t.Fatalf("next-track decision did not route completed Feature Depth to refactoring safely: %#v", decision)
	}
	priority, ok := decision["priority_order"].([]any)
	if !ok || len(priority) != 3 ||
		priority[0] != "refactoring" ||
		priority[1] != "feature_depth" ||
		priority[2] != "rsi_boundary_hardening" {
		t.Fatalf("next-track priority order is wrong: %#v", decision["priority_order"])
	}
	validator, err := validateRecommendationEvidenceTypedFile(outPath, "ao.atlas.recommendation-next-track-decision.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-next-track-decision" {
		t.Fatalf("expected typed next-track decision validator, got %s", validator)
	}
}
