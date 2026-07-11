package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3FoundrySafeNextWorkFixtureBindsTerminalReadiness(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-19-foundry-safe-next-work")
	sourceReadback := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-18-command-thin-client", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-18-command-thin-client", "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "foundry-safe-next-work-fixture.json")
	outPath := filepath.Join(t.TempDir(), "foundry-safe-next-work-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-foundry-safe-next-work",
		"--node-id", "mission-recommendation-month3-final-closure-19-foundry-safe-next-work",
		"--source-readback", sourceReadback,
		"--source-workgraph", sourceWorkgraph,
		"--expected-selected-node", "mission-recommendation-month3-final-closure-19-foundry-safe-next-work",
		"--expected-next-node-after-completion", "mission-recommendation-month3-final-closure-20-mission-recovery-invariant",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-foundry-safe-next-work command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=safe_next_work_selected",
		"selected_node=mission-recommendation-month3-final-closure-19-foundry-safe-next-work",
		"single_active_task=true",
		"executes_work=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("safe-next-work output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[AtlasMonth3FoundrySafeNextWorkFixture](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3FoundrySafeNextWorkFixture](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 Foundry safe-next-work fixture drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3FoundrySafeNextWorkFixture(recorded); err != nil {
		t.Fatalf("recorded Foundry safe-next-work fixture invalid: %v", err)
	}
	if recorded.CompletedNodesBefore != 18 ||
		recorded.ReadyNodesBefore != 12 ||
		recorded.SelectedNode != "mission-recommendation-month3-final-closure-19-foundry-safe-next-work" ||
		recorded.ExpectedNextNodeAfterCompletion != "mission-recommendation-month3-final-closure-20-mission-recovery-invariant" ||
		!recorded.SingleActiveTask ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("safe-next-work fixture lost terminal readiness or authority boundary: %#v", recorded)
	}
}
