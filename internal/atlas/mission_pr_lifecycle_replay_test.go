package atlas

import (
	"path/filepath"
	"testing"
)

func TestPRLifecycleReplayFixtureCoversInterruptedMergeSyncAndCleanupHandoffs(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01", "nodes", "refactoring-next-wave-25", "pr-lifecycle-replay-fixture.json")
	fixture := mustLoadJSON[AtlasPRLifecycleReplayFixture](t, path)
	if err := ValidateAtlasPRLifecycleReplayFixture(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "ao.atlas.pr-lifecycle-replay-fixture.v0.1" ||
		fixture.Status != "guarded" ||
		fixture.CaseCount != 3 ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("PR lifecycle replay fixture must cover guarded no-promotion handoffs: %#v", fixture)
	}

	cases := map[string]AtlasPRLifecycleReplayCase{}
	for _, replayCase := range fixture.Cases {
		cases[replayCase.CaseID] = replayCase
		if replayCase.FinalResponseAllowed ||
			replayCase.SafeToSelectNextNode ||
			replayCase.ClaimsAuthorityAdvance ||
			!replayCase.RSIRemainsDenied {
			t.Fatalf("replay case must block final response and next node selection without authority effects: %#v", replayCase)
		}
	}
	expectedActions := map[string]string{
		"interrupted_merge":   "merge_after_checks_pass",
		"interrupted_sync":    "sync_main_before_cleanup",
		"interrupted_cleanup": "delete_codex_branches_before_next_node",
	}
	for caseID, action := range expectedActions {
		replayCase, ok := cases[caseID]
		if !ok {
			t.Fatalf("missing replay case %q: %#v", caseID, cases)
		}
		if replayCase.OperatorAction != action {
			t.Fatalf("replay case %q action mismatch: %#v", caseID, replayCase)
		}
	}
	if cases["interrupted_merge"].MergeCommit != "" ||
		cases["interrupted_merge"].PRState != "open" ||
		cases["interrupted_merge"].CIStatus != "passed" {
		t.Fatalf("interrupted merge case must represent passed checks before merge: %#v", cases["interrupted_merge"])
	}
	if cases["interrupted_sync"].MergeCommit == "" ||
		cases["interrupted_sync"].LocalMainSynced ||
		cases["interrupted_sync"].LocalCodexBranchesRemaining != 1 {
		t.Fatalf("interrupted sync case must require main sync before cleanup: %#v", cases["interrupted_sync"])
	}
	if cases["interrupted_cleanup"].MergeCommit == "" ||
		!cases["interrupted_cleanup"].LocalMainSynced ||
		cases["interrupted_cleanup"].LocalCodexBranchesRemaining == 0 ||
		cases["interrupted_cleanup"].RemoteCodexBranchesRemaining == 0 {
		t.Fatalf("interrupted cleanup case must require local and remote codex branch deletion: %#v", cases["interrupted_cleanup"])
	}
}
