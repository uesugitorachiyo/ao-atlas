package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveBranchCleanupHandoffSummaryFixtureBindsOperatorHandoff(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-15", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "branch-cleanup-handoff-summary.json")
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	checkpointRoot := branchDeletionCheckpointRoot(t, waveRoot, recorded)
	outPath := filepath.Join(t.TempDir(), "branch-cleanup-handoff-summary.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "branch-cleanup-handoff-summary",
		"--evidence-root", checkpointRoot,
		"--source-readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("branch-cleanup-handoff-summary command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=branch_cleanup_handoff_summarized") ||
		!strings.Contains(out.String(), "post_merge_lifecycle_count=15") ||
		!strings.Contains(out.String(), "operator_handoff_status=cleanup_ledger_ready") {
		t.Fatalf("branch-cleanup-handoff-summary output missing handoff summary: %s", out.String())
	}
	generated := mustLoadJSON[map[string]any](t, outPath)
	generated["evidence_root"] = recorded["evidence_root"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("branch cleanup handoff summary fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["status"] != "branch_cleanup_handoff_summarized" ||
		generated["post_merge_lifecycle_count"] != float64(15) ||
		generated["merged_and_cleaned_count"] != float64(15) ||
		generated["passed_ci_count"] != float64(15) ||
		generated["cleanup_complete"] != true ||
		generated["final_response_allowed"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("branch cleanup handoff must summarize cleanup without authority effects: %#v", generated)
	}
}

func TestFeatureDepthWaveBranchCleanupHandoffSummaryUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-16", "branch-cleanup-handoff-summary.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.branch-cleanup-handoff-summary.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:branch-cleanup-handoff-summary" {
		t.Fatalf("expected typed branch cleanup handoff summary validator, got %s", validator)
	}
}

func TestMergeReadinessGuardRequiresPassedChecksBeforeBranchCleanupEvidence(t *testing.T) {
	root := repoRoot(t)
	summary := mustLoadJSON[AtlasBranchCleanupHandoffSummary](t, filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-16", "branch-cleanup-handoff-summary.json"))
	if len(summary.Entries) == 0 {
		t.Fatal("branch cleanup handoff fixture has no entries")
	}
	entry := summary.Entries[0]
	ready := EvaluateAtlasMergeReadinessGuard(AtlasMergeReadinessGuardInput{
		NodeID:      entry.NodeID,
		PRNumber:    entry.PRNumber,
		MergeCommit: entry.MergeCommit,
		CIStatus:    entry.CIStatus,
	})
	if ready.Status != "ready_for_branch_cleanup_evidence" ||
		ready.Reason != "passed_checks_bound_to_merge_commit" ||
		!ready.PassedChecksRequiredBeforeCleanup ||
		!ready.BranchCleanupEvidenceAllowed ||
		ready.ClaimsAuthorityAdvance ||
		!ready.RSIRemainsDenied {
		t.Fatalf("merge readiness guard should allow clean passed entry without authority effects: %#v", ready)
	}

	blocked := EvaluateAtlasMergeReadinessGuard(AtlasMergeReadinessGuardInput{
		NodeID:      entry.NodeID,
		PRNumber:    entry.PRNumber,
		MergeCommit: entry.MergeCommit,
		CIStatus:    "pending",
	})
	if blocked.Status != "blocked" ||
		blocked.Reason != "blocked_ci_not_passed" ||
		blocked.BranchCleanupEvidenceAllowed ||
		!blocked.PassedChecksRequiredBeforeCleanup ||
		blocked.ClaimsAuthorityAdvance ||
		!blocked.RSIRemainsDenied {
		t.Fatalf("merge readiness guard should block cleanup evidence before passed checks: %#v", blocked)
	}

	summary.Entries[0].CIStatus = "pending"
	summary.PassedCICount--
	summary.CleanupComplete = false
	summary.OperatorHandoffStatus = "cleanup_ledger_blocked"
	err := ValidateAtlasBranchCleanupHandoffSummary(summary)
	if err == nil || !strings.Contains(err.Error(), "merge readiness guard blocks branch cleanup evidence: blocked_ci_not_passed") {
		t.Fatalf("branch cleanup handoff validation must use merge readiness guard, got %v", err)
	}
}
