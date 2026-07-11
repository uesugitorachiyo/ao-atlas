package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3ControlPlaneObserverBindingConnectsAdapterToMissionTimeline(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-06-control-plane-observer")
	recordedPath := filepath.Join(nodeDir, "month3-control-plane-observer-binding.json")
	outPath := filepath.Join(t.TempDir(), "month3-control-plane-observer-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-control-plane-observer",
		"--node-id", "mission-recommendation-month3-final-closure-06-control-plane-observer",
		"--adapter-fixture", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-29", "command-readback-adapter-boundary-fixture.json"),
		"--source-readback", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-05-real-run-acceptance", "recommendation-readback-after.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-control-plane-observer command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3ControlPlaneObserverBinding](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3ControlPlaneObserverBinding](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 control-plane observer fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3ControlPlaneObserverBinding(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "control_plane_observer_bound" ||
		!recorded.AdapterDelegatesToControlPlane ||
		!recorded.MissionTimelineReadbackBound ||
		!recorded.ObserverReadOnly ||
		recorded.DuplicatesDomainDecisions ||
		recorded.CompletedNodes != 5 ||
		recorded.ReadyNodes != 25 ||
		recorded.NextExecutableNode != "mission-recommendation-month3-final-closure-06-control-plane-observer" ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("control-plane observer binding lost safety state: %#v", recorded)
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-control-plane-observer-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-control-plane-observer-binding" {
		t.Fatalf("expected typed Month 3 control-plane observer validator, got %s", validator)
	}
}
