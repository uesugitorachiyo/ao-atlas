package atlas

import (
	"path/filepath"
	"testing"
)

func TestConsolidationRepositoryBaselineSeparatesPreExistingStateFromWaveChanges(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "evidence", "ao-stack-consolidation-month1-wave-v01", "nodes", "mission-recommendation-consolidation-month1-01", "repository-baseline.json")
	baseline := mustLoadJSON[ConsolidationRepositoryBaseline](t, path)
	if err := ValidateConsolidationRepositoryBaseline(baseline); err != nil {
		t.Fatal(err)
	}

	states := map[string]string{}
	for _, repo := range baseline.Repositories {
		states[repo.Repository] = repo.StateClass
		if len(repo.WaveOwnedFiles) != 0 {
			t.Fatalf("baseline must not attribute pre-existing files to this wave: %#v", repo)
		}
	}
	if states["ao-mission"] != "pre_existing_dirty_and_behind" {
		t.Fatalf("Mission dirty state was not classified as pre-existing and out of sync: %#v", states)
	}
	if states["ao-foundry"] != "pre_existing_dirty" {
		t.Fatalf("Foundry local state was not classified as pre-existing dirty: %#v", states)
	}
	if states["ao-covenant"] != "pre_existing_codex_branch_and_out_of_sync" {
		t.Fatalf("Covenant branch state was not preserved: %#v", states)
	}
}
