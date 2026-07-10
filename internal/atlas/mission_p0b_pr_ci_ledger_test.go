package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BPRCILedgerRecordsMergeHeadsAndBranchCleanup(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-21", "p0b-pr-ci-ledger.json")
	ledger := mustLoadJSON[AtlasP0BPRCILedger](t, path)
	if err := ValidateAtlasP0BPRCILedger(ledger); err != nil {
		t.Fatal(err)
	}
	if ledger.CoveredNodeStart != 1 ||
		ledger.CoveredNodeEnd != 20 ||
		ledger.EntryCount != 20 ||
		!ledger.AllPRsMerged ||
		!ledger.AllCIStatusesPassed ||
		!ledger.AllMergeCommitsRecorded ||
		!ledger.AllBranchesDeleted ||
		ledger.BranchesRemainingTotal != 0 ||
		ledger.FinalResponseAllowed {
		t.Fatalf("P0-B PR/CI ledger summary drifted from complete-clean state: %#v", ledger)
	}
	for _, entry := range ledger.Entries {
		if entry.PRNumber < 499 ||
			entry.PRNumber > 518 ||
			entry.State != "MERGED" ||
			entry.CIStatus != "passed" ||
			entry.MergeCommit == "" ||
			!entry.LocalBranchDeleted ||
			!entry.RemoteBranchDeleted ||
			entry.LocalCodexBranchesRemaining != 0 ||
			entry.RemoteCodexBranchesRemaining != 0 {
			t.Fatalf("P0-B PR/CI entry must bind PR, CI, merge head, and cleanup: %#v", entry)
		}
	}
}

func TestP0BPRCILedgerUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-21", "p0b-pr-ci-ledger.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0BPRCILedgerContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0b-pr-ci-ledger" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
