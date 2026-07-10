package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BMissionImportReadbackRecordsNoFinalResponse(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-19", "mission-import-readback.json")
	recorded := mustLoadJSON[map[string]any](t, path)

	if recorded["schema"] != "ao.mission.import-readback.v0.1" ||
		recorded["mission_id"] != "mission-710327df54728420" ||
		recorded["kind"] != "atlas-recommendation-readback" ||
		recorded["status"] != "recorded" ||
		recorded["exact_next_action"] != "Emit Foundry import for mission-recommendation-p0b-contract-convergence-19 and execute exactly one active node." ||
		recorded["safe_to_execute"] != false ||
		recorded["executes_work"] != false ||
		recorded["approves_work"] != false {
		t.Fatalf("mission import readback lost no-final-response state: %#v", recorded)
	}
	artifact, ok := recorded["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("mission import readback missing artifact: %#v", recorded)
	}
	if artifact["kind"] != "atlas-recommendation-readback" ||
		artifact["digest"] != "sha256:214c94790970bd1cee68dc07ed7197e06e108c53bca6f34052d4d5c97874fafb" ||
		artifact["ref"] != "docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-18/recommendation-readback-after.json" {
		t.Fatalf("mission import readback artifact drifted: %#v", artifact)
	}
}
