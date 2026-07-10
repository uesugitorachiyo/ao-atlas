package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BRollbackAuditCoversEveryCompletedNodeBeforeClosure(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-23", "rollback-audit.json")
	audit := mustLoadJSON[AtlasP0BRollbackAudit](t, path)
	if err := ValidateAtlasP0BRollbackAudit(audit); err != nil {
		t.Fatal(err)
	}
	if audit.CoveredNodeStart != 1 ||
		audit.CoveredNodeEnd != 22 ||
		audit.EntryCount != 22 ||
		!audit.AllRollbackRecordsPresent ||
		!audit.AllRollbackRecordsReady ||
		audit.MissingRollbackRecordCount != 0 ||
		audit.ReleaseOrDeployRollbackCount != 0 ||
		audit.FinalResponseAllowed {
		t.Fatalf("P0-B rollback audit drifted from complete pre-closure coverage: %#v", audit)
	}
	for _, entry := range audit.Entries {
		if entry.RollbackRecordPath == "" ||
			entry.Status != "ready" ||
			entry.RollbackCommand == "" ||
			entry.RequiresReleaseOrDeploy ||
			entry.RollbackScopeCount == 0 {
			t.Fatalf("rollback audit entry must bind ready non-release rollback evidence: %#v", entry)
		}
	}
}

func TestP0BRollbackAuditUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-23", "rollback-audit.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0BRollbackAuditContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0b-rollback-audit" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
