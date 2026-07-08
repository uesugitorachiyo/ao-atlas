package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsConsumedLedgerRecordsCompletedFeatureDepthSource(t *testing.T) {
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
	tempDir := t.TempDir()
	nextTrackPath := filepath.Join(tempDir, "next-track-decision.json")
	ledgerPath := filepath.Join(tempDir, "consumed-recommendation-ledger.json")

	var nextTrackOut bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "next-track",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--out", nextTrackPath,
	}, &nextTrackOut, &nextTrackOut)
	if code != 0 {
		t.Fatalf("next-track failed: %s", nextTrackOut.String())
	}

	var out bytes.Buffer
	code = Run([]string{
		"mission", "recommendations", "consumed-ledger",
		"--source-evidence-root", sourceRoot,
		"--readback", sourceReadback,
		"--next-track-decision", nextTrackPath,
		"--out", ledgerPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("consumed-ledger failed: %s", out.String())
	}
	for _, want := range []string{
		"status=consumed_recorded",
		"consumed_track=feature_depth",
		"recommended_track=refactoring",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("consumed-ledger output missing %q: %s", want, out.String())
		}
	}

	ledger := mustLoadJSON[map[string]any](t, ledgerPath)
	if ledger["schema"] != "ao.atlas.consumed-recommendation-ledger.v0.1" ||
		ledger["status"] != "consumed_recorded" ||
		ledger["source_evidence_root"] != sourceRoot ||
		ledger["source_readback_path"] != sourceReadback ||
		ledger["next_track_decision_path"] != filepath.ToSlash(nextTrackPath) ||
		ledger["current_track"] != "feature_depth" ||
		ledger["current_track_status"] != "completed_saturated" ||
		ledger["consumed_track"] != "feature_depth" ||
		ledger["consumed_reason"] != "completed_saturated_feature_depth_routed_to_refactoring" ||
		ledger["recommended_track"] != "refactoring" ||
		ledger["duplicate_export_blocked"] != true ||
		ledger["import_bypass_blocked"] != true ||
		ledger["no_promotion_requested"] != true ||
		ledger["promotion_granted"] != false ||
		ledger["claims_authority_advance"] != false ||
		ledger["rsi_remains_denied"] != true ||
		ledger["safe_to_execute"] != false ||
		ledger["schedules_work"] != false ||
		ledger["executes_work"] != false ||
		ledger["approves_work"] != false ||
		ledger["mutates_repositories"] != false {
		t.Fatalf("consumed ledger did not record completed Feature Depth source safely: %#v", ledger)
	}
	validator, err := validateRecommendationEvidenceTypedFile(ledgerPath, "ao.atlas.consumed-recommendation-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:consumed-recommendation-ledger" {
		t.Fatalf("expected typed consumed recommendation ledger validator, got %s", validator)
	}
}
