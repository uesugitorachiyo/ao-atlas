package atlas

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinalClosureConsolidationBindsPublicSafetyAndNodeOneLifecycle(t *testing.T) {
	longRunRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	finalNodeDir := filepath.Join(longRunRoot, "nodes", "mission-recommendation-hardening-40")
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(finalNodeDir, "recommendation-readback-after.json"))
	sentinel := mustLoadJSON[struct {
		Status                 string `json:"status"`
		ForbiddenClaimsPresent bool   `json:"forbidden_claims_present"`
		RSIClaimPresent        bool   `json:"rsi_claim_present"`
	}](t, filepath.Join(finalNodeDir, "sentinel_public_safety.json"))
	verification := mustLoadJSON[struct {
		Status                  string `json:"status"`
		PublicSafetyScanPassed  bool   `json:"public_safety_scan_passed"`
		LocalVerificationPassed bool   `json:"local_verification_passed"`
		RSIRemainsDenied        bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(finalNodeDir, "final-verification-summary.json"))

	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeOneLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-01", "post-merge-lifecycle.json"))
	nodeTwoFixture := mustLoadJSON[struct {
		Schema                         string `json:"schema"`
		NodeID                         string `json:"node_id"`
		Status                         string `json:"status"`
		SourceReadbackPath             string `json:"source_readback_path"`
		SentinelEvidencePath           string `json:"sentinel_evidence_path"`
		VerificationSummaryPath        string `json:"verification_summary_path"`
		BoundPublicSafetyScanStatus    string `json:"bound_public_safety_scan_status"`
		PreviousPublicSafetyScanStatus string `json:"previous_public_safety_scan_status"`
		ReadyNodesAfterBinding         int    `json:"ready_nodes_after_binding"`
		FinalResponseAllowedAfter      bool   `json:"final_response_allowed_after_binding"`
		RSIRemainsDenied               bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-02", "public-safety-readback-binding.json"))
	nodeOneReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-01", "recommendation-readback-after.json"))

	if finalReadback.PublicSafetyScanStatus != "passed" ||
		sentinel.Status != "passed" ||
		sentinel.ForbiddenClaimsPresent ||
		sentinel.RSIClaimPresent ||
		verification.Status != "passed" ||
		!verification.PublicSafetyScanPassed ||
		!verification.LocalVerificationPassed ||
		!verification.RSIRemainsDenied {
		t.Fatalf("final readback must bind to passed Sentinel and production verification: readback=%#v sentinel=%#v verification=%#v", finalReadback.PublicSafetyScanStatus, sentinel, verification)
	}
	if nodeOneLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeOneLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-01" ||
		nodeOneLifecycle.Status != "merged_and_cleaned" ||
		nodeOneLifecycle.PRNumber != 304 ||
		nodeOneLifecycle.MergeCommit != "4ae4e22db082e4633d04003be7f0b576e2ccc0c5" ||
		nodeOneLifecycle.CIStatus != "passed" ||
		!nodeOneLifecycle.LocalMainSynced ||
		!nodeOneLifecycle.LocalBranchDeleted ||
		!nodeOneLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 1 lifecycle evidence must record PR, CI, merge, sync, and cleanup: %#v", nodeOneLifecycle)
	}
	if nodeTwoFixture.Schema != "ao.atlas.public-safety-readback-binding.v0.1" ||
		nodeTwoFixture.NodeID != "mission-recommendation-final-closure-consolidation-02" ||
		nodeTwoFixture.Status != "bound" ||
		nodeTwoFixture.SourceReadbackPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/recommendation-readback-after.json" ||
		nodeTwoFixture.SentinelEvidencePath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/sentinel_public_safety.json" ||
		nodeTwoFixture.VerificationSummaryPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/final-verification-summary.json" ||
		nodeTwoFixture.BoundPublicSafetyScanStatus != finalReadback.PublicSafetyScanStatus ||
		nodeTwoFixture.PreviousPublicSafetyScanStatus != "required_pending_verification" ||
		nodeTwoFixture.ReadyNodesAfterBinding != nodeOneReadback.ReadyNodes ||
		nodeTwoFixture.FinalResponseAllowedAfter ||
		!nodeTwoFixture.RSIRemainsDenied {
		t.Fatalf("node 2 public-safety binding fixture must preserve the final readback and continuation state: %#v", nodeTwoFixture)
	}
}

func TestFinalClosureConsolidationPublicSafetyGuardRejectsPendingClosure(t *testing.T) {
	longRunRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(longRunRoot, "nodes", "mission-recommendation-hardening-40", "recommendation-readback-after.json"))
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTwoLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-02", "post-merge-lifecycle.json"))
	fixture := mustLoadJSON[struct {
		Schema                                      string   `json:"schema"`
		NodeID                                      string   `json:"node_id"`
		Status                                      string   `json:"status"`
		SourceReadbackPath                          string   `json:"source_readback_path"`
		FinalResponseAllowed                        bool     `json:"final_response_allowed"`
		BoundPublicSafetyScanStatus                 string   `json:"bound_public_safety_scan_status"`
		RejectedPublicSafetyScanStatuses            []string `json:"rejected_public_safety_scan_statuses"`
		RequiresNonPendingStatusBeforeFinalResponse bool     `json:"requires_non_pending_status_before_final_response"`
		RSIRemainsDenied                            bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-03", "public-safety-pending-closure-guard.json"))

	if !finalReadback.FinalResponseAllowed ||
		finalReadback.PublicSafetyScanStatus == "required_pending_verification" ||
		finalReadback.PublicSafetyScanStatus != fixture.BoundPublicSafetyScanStatus {
		t.Fatalf("final readback must not allow closure with pending public-safety status: %#v", finalReadback)
	}
	if nodeTwoLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTwoLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-02" ||
		nodeTwoLifecycle.Status != "merged_and_cleaned" ||
		nodeTwoLifecycle.PRNumber != 305 ||
		nodeTwoLifecycle.MergeCommit != "b291edcb9089206b77f153a053bf7cc035495cbd" ||
		nodeTwoLifecycle.CIStatus != "passed" ||
		!nodeTwoLifecycle.LocalMainSynced ||
		!nodeTwoLifecycle.LocalBranchDeleted ||
		!nodeTwoLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 2 lifecycle evidence must record PR, CI, merge, sync, and cleanup: %#v", nodeTwoLifecycle)
	}
	if fixture.Schema != "ao.atlas.public-safety-pending-closure-guard.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-03" ||
		fixture.Status != "guarded" ||
		fixture.SourceReadbackPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/recommendation-readback-after.json" ||
		!fixture.FinalResponseAllowed ||
		!containsString(fixture.RejectedPublicSafetyScanStatuses, "required_pending_verification") ||
		!fixture.RequiresNonPendingStatusBeforeFinalResponse ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("pending closure guard fixture must reject pending public-safety final closure: %#v", fixture)
	}
}

func TestFinalClosureConsolidationPostMergeCleanupRollupCoversMergedNodes(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	type lifecycleEvidence struct {
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
		LocalCodexBranches  int    `json:"local_codex_branches"`
		RemoteCodexBranches int    `json:"remote_codex_branches"`
	}
	expected := map[string]struct {
		pr    int
		merge string
	}{
		"mission-recommendation-final-closure-consolidation-01": {304, "4ae4e22db082e4633d04003be7f0b576e2ccc0c5"},
		"mission-recommendation-final-closure-consolidation-02": {305, "b291edcb9089206b77f153a053bf7cc035495cbd"},
		"mission-recommendation-final-closure-consolidation-03": {306, "94a84a836bb39a45de2601c509128cf2f6fa388f"},
	}
	for nodeID, want := range expected {
		lifecycle := mustLoadJSON[lifecycleEvidence](t, filepath.Join(consolidationRoot, "nodes", nodeID, "post-merge-lifecycle.json"))
		if lifecycle.NodeID != nodeID ||
			lifecycle.Status != "merged_and_cleaned" ||
			lifecycle.PRNumber != want.pr ||
			lifecycle.MergeCommit != want.merge ||
			lifecycle.CIStatus != "passed" ||
			!lifecycle.LocalMainSynced ||
			!lifecycle.LocalBranchDeleted ||
			!lifecycle.RemoteBranchDeleted ||
			lifecycle.LocalCodexBranches != 0 ||
			lifecycle.RemoteCodexBranches != 0 {
			t.Fatalf("post-merge lifecycle evidence must be clean for %s: %#v", nodeID, lifecycle)
		}
	}

	rollup := mustLoadJSON[struct {
		Schema                         string `json:"schema"`
		NodeID                         string `json:"node_id"`
		Status                         string `json:"status"`
		CoveredNodes                   int    `json:"covered_nodes"`
		LocalCodexBranches             int    `json:"local_codex_branches"`
		RemoteCodexBranches            int    `json:"remote_codex_branches"`
		AllBranchesDeleted             bool   `json:"all_branches_deleted"`
		AllCIStatusesPassed            bool   `json:"all_ci_statuses_passed"`
		AllMergeCommitsRecorded        bool   `json:"all_merge_commits_recorded"`
		GeneratedAfterBranchDeletion   bool   `json:"generated_after_branch_deletion"`
		ReadyNodesAfterCleanupEvidence int    `json:"ready_nodes_after_cleanup_evidence"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-04", "post-merge-cleanup-rollup.json"))
	nodeThreeReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-03", "recommendation-readback-after.json"))
	if rollup.Schema != "ao.atlas.post-merge-cleanup-rollup.v0.1" ||
		rollup.NodeID != "mission-recommendation-final-closure-consolidation-04" ||
		rollup.Status != "clean" ||
		rollup.CoveredNodes != 3 ||
		rollup.LocalCodexBranches != 0 ||
		rollup.RemoteCodexBranches != 0 ||
		!rollup.AllBranchesDeleted ||
		!rollup.AllCIStatusesPassed ||
		!rollup.AllMergeCommitsRecorded ||
		!rollup.GeneratedAfterBranchDeletion ||
		rollup.ReadyNodesAfterCleanupEvidence != nodeThreeReadback.ReadyNodes {
		t.Fatalf("cleanup rollup must summarize clean merged node lifecycle evidence: %#v", rollup)
	}
}

func TestFinalClosureConsolidationBranchCleanupHandoffRegressionCoversNodeFive(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeFourLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
		LocalCodexBranches  int    `json:"local_codex_branches"`
		RemoteCodexBranches int    `json:"remote_codex_branches"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-04", "post-merge-lifecycle.json"))
	nodeFourReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-04", "recommendation-readback-after.json"))
	fixture := mustLoadJSON[struct {
		Schema                           string `json:"schema"`
		NodeID                           string `json:"node_id"`
		Status                           string `json:"status"`
		SourceReadbackPath               string `json:"source_readback_path"`
		CompletedNodesBefore             int    `json:"completed_nodes_before"`
		ReadyNodesBefore                 int    `json:"ready_nodes_before"`
		FirstExecutableNodeBefore        string `json:"first_executable_node_before"`
		LocalCodexBranches               int    `json:"local_codex_branches"`
		RemoteCodexBranches              int    `json:"remote_codex_branches"`
		AllPriorCleanupRollupsClean      bool   `json:"all_prior_cleanup_rollups_clean"`
		NextNodeRequiresCleanBranchState bool   `json:"next_node_requires_clean_branch_state"`
		FinalResponseAllowedBefore       bool   `json:"final_response_allowed_before"`
		RSIRemainsDenied                 bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-05", "branch-cleanup-handoff-regression.json"))

	if nodeFourLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeFourLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-04" ||
		nodeFourLifecycle.Status != "merged_and_cleaned" ||
		nodeFourLifecycle.PRNumber != 307 ||
		nodeFourLifecycle.MergeCommit != "7b21bd14e42714352203eb7790c9372da5eb4e3f" ||
		nodeFourLifecycle.CIStatus != "passed" ||
		!nodeFourLifecycle.LocalMainSynced ||
		!nodeFourLifecycle.LocalBranchDeleted ||
		!nodeFourLifecycle.RemoteBranchDeleted ||
		nodeFourLifecycle.LocalCodexBranches != 0 ||
		nodeFourLifecycle.RemoteCodexBranches != 0 {
		t.Fatalf("node 4 lifecycle evidence must prove clean branch handoff: %#v", nodeFourLifecycle)
	}
	if fixture.Schema != "ao.atlas.branch-cleanup-handoff-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-05" ||
		fixture.Status != "guarded" ||
		fixture.SourceReadbackPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-04/recommendation-readback-after.json" ||
		fixture.CompletedNodesBefore != nodeFourReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeFourReadback.ReadyNodes ||
		fixture.FirstExecutableNodeBefore != nodeFourReadback.FirstExecutableNode ||
		fixture.LocalCodexBranches != 0 ||
		fixture.RemoteCodexBranches != 0 ||
		!fixture.AllPriorCleanupRollupsClean ||
		!fixture.NextNodeRequiresCleanBranchState ||
		fixture.FinalResponseAllowedBefore ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("branch cleanup handoff fixture must bind node 4 readback to clean node 5 start: %#v", fixture)
	}
}

func TestFinalClosureConsolidationAggregatePromoterCommandRollupBindsFinalWave(t *testing.T) {
	longRunRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	finalNodeDir := filepath.Join(longRunRoot, "nodes", "mission-recommendation-hardening-40")
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(finalNodeDir, "recommendation-readback-after.json"))
	promoter := mustLoadJSON[struct {
		Status                          string `json:"status"`
		HighestProvenLiveClassUnchanged bool   `json:"highest_proven_live_class_unchanged"`
		RSIRemainsDenied                bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(finalNodeDir, "promoter_no_promotion.json"))
	command := mustLoadJSON[struct {
		Status string `json:"status"`
	}](t, filepath.Join(finalNodeDir, "command_readback.json"))
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeFiveLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-05", "post-merge-lifecycle.json"))
	rollup := mustLoadJSON[struct {
		Schema                       string `json:"schema"`
		NodeID                       string `json:"node_id"`
		Status                       string `json:"status"`
		SourcePromoterPath           string `json:"source_promoter_path"`
		SourceCommandPath            string `json:"source_command_path"`
		SourceFinalReadbackPath      string `json:"source_final_readback_path"`
		FinalWaveCompletedNodes      int    `json:"final_wave_completed_nodes"`
		FinalWaveReadyNodes          int    `json:"final_wave_ready_nodes"`
		FinalResponseAllowed         bool   `json:"final_response_allowed"`
		PromoterStatus               string `json:"promoter_status"`
		CommandStatus                string `json:"command_status"`
		AggregatePromotionStatus     string `json:"aggregate_promotion_status"`
		PromotionRequested           bool   `json:"promotion_requested"`
		PromotionGranted             bool   `json:"promotion_granted"`
		CommandAgreesNoPromotion     bool   `json:"command_agrees_no_promotion"`
		RSIRemainsDenied             bool   `json:"rsi_remains_denied"`
		ConsolidationCompletedBefore int    `json:"consolidation_completed_before"`
		ConsolidationReadyBefore     int    `json:"consolidation_ready_before"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-06", "aggregate-promoter-command-rollup.json"))
	nodeFiveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-05", "recommendation-readback-after.json"))

	if promoter.Status != "no_promotion_requested" ||
		!promoter.HighestProvenLiveClassUnchanged ||
		!promoter.RSIRemainsDenied ||
		command.Status != "readback_agrees_no_promotion" {
		t.Fatalf("source Promoter and Command evidence must agree no promotion: promoter=%#v command=%#v", promoter, command)
	}
	if nodeFiveLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeFiveLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-05" ||
		nodeFiveLifecycle.Status != "merged_and_cleaned" ||
		nodeFiveLifecycle.PRNumber != 308 ||
		nodeFiveLifecycle.MergeCommit != "fe6d5536e8bf8f5951cc2cb230d32202a176482b" ||
		nodeFiveLifecycle.CIStatus != "passed" ||
		!nodeFiveLifecycle.LocalBranchDeleted ||
		!nodeFiveLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 5 lifecycle evidence must be recorded before node 6 rollup: %#v", nodeFiveLifecycle)
	}
	if rollup.Schema != "ao.atlas.aggregate-promoter-command-rollup.v0.1" ||
		rollup.NodeID != "mission-recommendation-final-closure-consolidation-06" ||
		rollup.Status != "no_promotion_rollup_bound" ||
		rollup.SourcePromoterPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/promoter_no_promotion.json" ||
		rollup.SourceCommandPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/command_readback.json" ||
		rollup.SourceFinalReadbackPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/recommendation-readback-after.json" ||
		rollup.FinalWaveCompletedNodes != finalReadback.CompletedNodes ||
		rollup.FinalWaveReadyNodes != finalReadback.ReadyNodes ||
		!rollup.FinalResponseAllowed ||
		rollup.PromoterStatus != promoter.Status ||
		rollup.CommandStatus != command.Status ||
		rollup.AggregatePromotionStatus != "no_promotion_requested" ||
		rollup.PromotionRequested ||
		rollup.PromotionGranted ||
		!rollup.CommandAgreesNoPromotion ||
		!rollup.RSIRemainsDenied ||
		rollup.ConsolidationCompletedBefore != nodeFiveReadback.CompletedNodes ||
		rollup.ConsolidationReadyBefore != nodeFiveReadback.ReadyNodes {
		t.Fatalf("aggregate Promoter/Command rollup must bind final wave no-promotion evidence: %#v", rollup)
	}
}

func TestFinalClosureConsolidationAggregateRollupRegressionPreservesNoPromotionAgreement(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeSixLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
		LocalCodexBranches  int    `json:"local_codex_branches"`
		RemoteCodexBranches int    `json:"remote_codex_branches"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-06", "post-merge-lifecycle.json"))
	nodeSixRollup := mustLoadJSON[struct {
		PromoterStatus           string `json:"promoter_status"`
		CommandStatus            string `json:"command_status"`
		AggregatePromotionStatus string `json:"aggregate_promotion_status"`
		PromotionRequested       bool   `json:"promotion_requested"`
		PromotionGranted         bool   `json:"promotion_granted"`
		CommandAgreesNoPromotion bool   `json:"command_agrees_no_promotion"`
		RSIRemainsDenied         bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-06", "aggregate-promoter-command-rollup.json"))
	nodeSixReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-06", "recommendation-readback-after.json"))
	fixture := mustLoadJSON[struct {
		Schema                         string `json:"schema"`
		NodeID                         string `json:"node_id"`
		Status                         string `json:"status"`
		SourceRollupPath               string `json:"source_rollup_path"`
		SourceReadbackPath             string `json:"source_readback_path"`
		PromoterStatus                 string `json:"promoter_status"`
		CommandStatus                  string `json:"command_status"`
		AggregatePromotionStatus       string `json:"aggregate_promotion_status"`
		PromotionRequested             bool   `json:"promotion_requested"`
		PromotionGranted               bool   `json:"promotion_granted"`
		CommandAgreesNoPromotion       bool   `json:"command_agrees_no_promotion"`
		CompletedNodesBefore           int    `json:"completed_nodes_before"`
		ReadyNodesBefore               int    `json:"ready_nodes_before"`
		FirstExecutableNodeBefore      string `json:"first_executable_node_before"`
		FinalResponseAllowedBefore     bool   `json:"final_response_allowed_before"`
		RegressionBlocksPromotionDrift bool   `json:"regression_blocks_promotion_drift"`
		RSIRemainsDenied               bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-07", "aggregate-rollup-regression.json"))

	if nodeSixLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeSixLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-06" ||
		nodeSixLifecycle.Status != "merged_and_cleaned" ||
		nodeSixLifecycle.PRNumber != 309 ||
		nodeSixLifecycle.MergeCommit != "6140ee3ac7c167d6ff3f6f0ea0c9a927f2020c3e" ||
		nodeSixLifecycle.CIStatus != "passed" ||
		!nodeSixLifecycle.LocalMainSynced ||
		!nodeSixLifecycle.LocalBranchDeleted ||
		!nodeSixLifecycle.RemoteBranchDeleted ||
		nodeSixLifecycle.LocalCodexBranches != 0 ||
		nodeSixLifecycle.RemoteCodexBranches != 0 {
		t.Fatalf("node 6 lifecycle evidence must prove clean branch handoff: %#v", nodeSixLifecycle)
	}
	if fixture.Schema != "ao.atlas.aggregate-rollup-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-07" ||
		fixture.Status != "guarded" ||
		fixture.SourceRollupPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-06/aggregate-promoter-command-rollup.json" ||
		fixture.SourceReadbackPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-06/recommendation-readback-after.json" ||
		fixture.PromoterStatus != nodeSixRollup.PromoterStatus ||
		fixture.CommandStatus != nodeSixRollup.CommandStatus ||
		fixture.AggregatePromotionStatus != nodeSixRollup.AggregatePromotionStatus ||
		fixture.PromotionRequested != nodeSixRollup.PromotionRequested ||
		fixture.PromotionGranted != nodeSixRollup.PromotionGranted ||
		fixture.CommandAgreesNoPromotion != nodeSixRollup.CommandAgreesNoPromotion ||
		fixture.CompletedNodesBefore != nodeSixReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeSixReadback.ReadyNodes ||
		fixture.FirstExecutableNodeBefore != nodeSixReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowedBefore != nodeSixReadback.FinalResponseAllowed ||
		!fixture.RegressionBlocksPromotionDrift ||
		!fixture.RSIRemainsDenied ||
		!nodeSixRollup.RSIRemainsDenied {
		t.Fatalf("aggregate rollup regression must preserve no-promotion, Command agreement, and RSI denial: %#v", fixture)
	}
}

func TestFinalClosureConsolidationPRCILedgerCoversHardeningNodesTwentyEightThroughForty(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeSevenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
		LocalCodexBranches  int    `json:"local_codex_branches"`
		RemoteCodexBranches int    `json:"remote_codex_branches"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-07", "post-merge-lifecycle.json"))
	nodeSevenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-07", "recommendation-readback-after.json"))
	ledger := mustLoadJSON[struct {
		Schema                     string `json:"schema"`
		NodeID                     string `json:"node_id"`
		Status                     string `json:"status"`
		Source                     string `json:"source"`
		CoveredHardeningNodeStart  int    `json:"covered_hardening_node_start"`
		CoveredHardeningNodeEnd    int    `json:"covered_hardening_node_end"`
		EntryCount                 int    `json:"entry_count"`
		AllPRsMerged               bool   `json:"all_prs_merged"`
		AllCIStatusesPassed        bool   `json:"all_ci_statuses_passed"`
		AllMergeCommitsRecorded    bool   `json:"all_merge_commits_recorded"`
		ReadyNodesBefore           int    `json:"ready_nodes_before"`
		CompletedNodesBefore       int    `json:"completed_nodes_before"`
		FirstExecutableNodeBefore  string `json:"first_executable_node_before"`
		FinalResponseAllowedBefore bool   `json:"final_response_allowed_before"`
		Entries                    []struct {
			HardeningNode       int    `json:"hardening_node"`
			NodeID              string `json:"node_id"`
			PRNumber            int    `json:"pr_number"`
			PRURL               string `json:"pr_url"`
			Title               string `json:"title"`
			HeadRef             string `json:"head_ref"`
			State               string `json:"state"`
			MergedAt            string `json:"merged_at"`
			MergeCommit         string `json:"merge_commit"`
			CheckCount          int    `json:"check_count"`
			SuccessCount        int    `json:"success_count"`
			UbuntuSuccessCount  int    `json:"ubuntu_success_count"`
			MacOSSuccessCount   int    `json:"macos_success_count"`
			WindowsSuccessCount int    `json:"windows_success_count"`
			CIStatus            string `json:"ci_status"`
		} `json:"entries"`
		RSIRemainsDenied bool `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-08", "hardening-nodes-28-40-pr-ci-ledger.json"))

	if nodeSevenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeSevenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-07" ||
		nodeSevenLifecycle.Status != "merged_and_cleaned" ||
		nodeSevenLifecycle.PRNumber != 310 ||
		nodeSevenLifecycle.MergeCommit != "64c8cbac11e7ec752cda5e160a6fd4e8bfd59a65" ||
		nodeSevenLifecycle.CIStatus != "passed" ||
		!nodeSevenLifecycle.LocalMainSynced ||
		!nodeSevenLifecycle.LocalBranchDeleted ||
		!nodeSevenLifecycle.RemoteBranchDeleted ||
		nodeSevenLifecycle.LocalCodexBranches != 0 ||
		nodeSevenLifecycle.RemoteCodexBranches != 0 {
		t.Fatalf("node 7 lifecycle evidence must prove clean branch handoff: %#v", nodeSevenLifecycle)
	}
	if ledger.Schema != "ao.atlas.pr-ci-ledger.v0.1" ||
		ledger.NodeID != "mission-recommendation-final-closure-consolidation-08" ||
		ledger.Status != "complete" ||
		ledger.Source != "gh_pr_view_291_303" ||
		ledger.CoveredHardeningNodeStart != 28 ||
		ledger.CoveredHardeningNodeEnd != 40 ||
		ledger.EntryCount != 13 ||
		len(ledger.Entries) != 13 ||
		!ledger.AllPRsMerged ||
		!ledger.AllCIStatusesPassed ||
		!ledger.AllMergeCommitsRecorded ||
		ledger.CompletedNodesBefore != nodeSevenReadback.CompletedNodes ||
		ledger.ReadyNodesBefore != nodeSevenReadback.ReadyNodes ||
		ledger.FirstExecutableNodeBefore != nodeSevenReadback.FirstExecutableNode ||
		ledger.FinalResponseAllowedBefore != nodeSevenReadback.FinalResponseAllowed ||
		!ledger.RSIRemainsDenied {
		t.Fatalf("node 8 ledger must summarize hardening nodes 28-40 from node 7 readback: %#v", ledger)
	}
	for i, entry := range ledger.Entries {
		wantNode := 28 + i
		wantPR := 291 + i
		if entry.HardeningNode != wantNode ||
			entry.NodeID != fmt.Sprintf("mission-recommendation-hardening-%02d", wantNode) ||
			entry.PRNumber != wantPR ||
			entry.PRURL != fmt.Sprintf("https://github.com/uesugitorachiyo/ao-atlas/pull/%d", wantPR) ||
			entry.State != "MERGED" ||
			entry.MergedAt == "" ||
			entry.MergeCommit == "" ||
			entry.CheckCount != 9 ||
			entry.SuccessCount != 9 ||
			entry.UbuntuSuccessCount != 3 ||
			entry.MacOSSuccessCount != 3 ||
			entry.WindowsSuccessCount != 3 ||
			entry.CIStatus != "passed" {
			t.Fatalf("ledger entry %d must bind hardening node %d to merged PR %d and 9 passing checks: %#v", i, wantNode, wantPR, entry)
		}
	}
	if first, last := ledger.Entries[0], ledger.Entries[len(ledger.Entries)-1]; first.MergeCommit != "a805a44bed4ecc5b7cde5ed38d9e5e0131ad8cae" ||
		last.MergeCommit != "1201070f2f68decab5c9156babaef506d0b67945" {
		t.Fatalf("ledger must preserve first and final hardening merge commits: first=%#v last=%#v", first, last)
	}
}

func TestFinalClosureConsolidationPRCILedgerRegressionLocksPRNumbersMergeHeadsAndCheckStates(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeEightLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
		LocalCodexBranches  int    `json:"local_codex_branches"`
		RemoteCodexBranches int    `json:"remote_codex_branches"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-08", "post-merge-lifecycle.json"))
	nodeEightReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-08", "recommendation-readback-after.json"))
	ledger := mustLoadJSON[struct {
		Entries []struct {
			HardeningNode       int    `json:"hardening_node"`
			PRNumber            int    `json:"pr_number"`
			MergeCommit         string `json:"merge_commit"`
			CheckCount          int    `json:"check_count"`
			SuccessCount        int    `json:"success_count"`
			UbuntuSuccessCount  int    `json:"ubuntu_success_count"`
			MacOSSuccessCount   int    `json:"macos_success_count"`
			WindowsSuccessCount int    `json:"windows_success_count"`
			CIStatus            string `json:"ci_status"`
		} `json:"entries"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-08", "hardening-nodes-28-40-pr-ci-ledger.json"))
	fixture := mustLoadJSON[struct {
		Schema                         string `json:"schema"`
		NodeID                         string `json:"node_id"`
		Status                         string `json:"status"`
		SourceLedgerPath               string `json:"source_ledger_path"`
		CoveredEntryCount              int    `json:"covered_entry_count"`
		FirstHardeningNode             int    `json:"first_hardening_node"`
		LastHardeningNode              int    `json:"last_hardening_node"`
		FirstPRNumber                  int    `json:"first_pr_number"`
		LastPRNumber                   int    `json:"last_pr_number"`
		FirstMergeCommit               string `json:"first_merge_commit"`
		FinalMergeCommit               string `json:"final_merge_commit"`
		ExpectedCheckCountPerEntry     int    `json:"expected_check_count_per_entry"`
		ExpectedSuccessCountPerEntry   int    `json:"expected_success_count_per_entry"`
		ExpectedOSSuccessCountPerEntry int    `json:"expected_os_success_count_per_entry"`
		PRNumbersConsecutive           bool   `json:"pr_numbers_consecutive"`
		MergeHeadsUnique               bool   `json:"merge_heads_unique"`
		CIStatesAllPassed              bool   `json:"ci_states_all_passed"`
		CompletedNodesBefore           int    `json:"completed_nodes_before"`
		ReadyNodesBefore               int    `json:"ready_nodes_before"`
		FirstExecutableNodeBefore      string `json:"first_executable_node_before"`
		FinalResponseAllowedBefore     bool   `json:"final_response_allowed_before"`
		RSIRemainsDenied               bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-09", "pr-ci-ledger-regression.json"))

	if nodeEightLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeEightLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-08" ||
		nodeEightLifecycle.Status != "merged_and_cleaned" ||
		nodeEightLifecycle.PRNumber != 311 ||
		nodeEightLifecycle.MergeCommit != "89ce8cd3c727f5b5025058c6accafec6b645a63b" ||
		nodeEightLifecycle.CIStatus != "passed" ||
		!nodeEightLifecycle.LocalMainSynced ||
		!nodeEightLifecycle.LocalBranchDeleted ||
		!nodeEightLifecycle.RemoteBranchDeleted ||
		nodeEightLifecycle.LocalCodexBranches != 0 ||
		nodeEightLifecycle.RemoteCodexBranches != 0 {
		t.Fatalf("node 8 lifecycle evidence must prove clean branch handoff: %#v", nodeEightLifecycle)
	}
	if fixture.Schema != "ao.atlas.pr-ci-ledger-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-09" ||
		fixture.Status != "guarded" ||
		fixture.SourceLedgerPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-08/hardening-nodes-28-40-pr-ci-ledger.json" ||
		fixture.CoveredEntryCount != len(ledger.Entries) ||
		fixture.FirstHardeningNode != 28 ||
		fixture.LastHardeningNode != 40 ||
		fixture.FirstPRNumber != 291 ||
		fixture.LastPRNumber != 303 ||
		fixture.FirstMergeCommit != "a805a44bed4ecc5b7cde5ed38d9e5e0131ad8cae" ||
		fixture.FinalMergeCommit != "1201070f2f68decab5c9156babaef506d0b67945" ||
		fixture.ExpectedCheckCountPerEntry != 9 ||
		fixture.ExpectedSuccessCountPerEntry != 9 ||
		fixture.ExpectedOSSuccessCountPerEntry != 3 ||
		!fixture.PRNumbersConsecutive ||
		!fixture.MergeHeadsUnique ||
		!fixture.CIStatesAllPassed ||
		fixture.CompletedNodesBefore != nodeEightReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeEightReadback.ReadyNodes ||
		fixture.FirstExecutableNodeBefore != nodeEightReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowedBefore != nodeEightReadback.FinalResponseAllowed ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("ledger regression fixture must lock PR, merge-head, CI state, and readback continuity: %#v", fixture)
	}
	seenMergeHeads := map[string]bool{}
	for i, entry := range ledger.Entries {
		if entry.HardeningNode != 28+i ||
			entry.PRNumber != 291+i ||
			entry.MergeCommit == "" ||
			seenMergeHeads[entry.MergeCommit] ||
			entry.CheckCount != fixture.ExpectedCheckCountPerEntry ||
			entry.SuccessCount != fixture.ExpectedSuccessCountPerEntry ||
			entry.UbuntuSuccessCount != fixture.ExpectedOSSuccessCountPerEntry ||
			entry.MacOSSuccessCount != fixture.ExpectedOSSuccessCountPerEntry ||
			entry.WindowsSuccessCount != fixture.ExpectedOSSuccessCountPerEntry ||
			entry.CIStatus != "passed" {
			t.Fatalf("ledger regression entry %d drifted from expected PR/CI continuity: %#v", i, entry)
		}
		seenMergeHeads[entry.MergeCommit] = true
	}
}

func TestFinalClosureConsolidationFinalOperatorSummaryBindsReadbackAndClosureFixture(t *testing.T) {
	repoRoot := repoRoot(t)
	consolidationRoot := filepath.Join(repoRoot, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	longRunFinalDir := filepath.Join(repoRoot, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01", "nodes", "mission-recommendation-hardening-40")
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(longRunFinalDir, "recommendation-readback-after.json"))
	closure := mustLoadJSON[struct {
		Schema                    string `json:"schema"`
		Status                    string `json:"status"`
		PromoterNoPromotionStatus string `json:"promoter_no_promotion_status"`
		CommandReadbackStatus     string `json:"command_readback_status"`
		ClaimsAuthorityAdvance    bool   `json:"claims_authority_advance"`
		RSIRemainsDenied          bool   `json:"rsi_remains_denied"`
		FinalResponseAllowedAfter bool   `json:"final_response_allowed_after_node"`
		ReadyNodesAfterNode       int    `json:"ready_nodes_after_node"`
		BlockedNodesAfterNode     int    `json:"blocked_nodes_after_node"`
		CleanRepoStatusPath       string `json:"clean_repo_status_path"`
		VerificationSummaryPath   string `json:"verification_summary_path"`
	}](t, filepath.Join(longRunFinalDir, "final-closure-artifacts-fixture.json"))
	nodeNineLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-09", "post-merge-lifecycle.json"))
	nodeNineReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-09", "recommendation-readback-after.json"))
	summary := mustLoadJSON[struct {
		Schema                          string   `json:"schema"`
		NodeID                          string   `json:"node_id"`
		Status                          string   `json:"status"`
		SourceFinalReadbackPath         string   `json:"source_final_readback_path"`
		SourceClosureFixturePath        string   `json:"source_closure_fixture_path"`
		SummaryMarkdownPath             string   `json:"summary_markdown_path"`
		FinalWaveCompletedNodes         int      `json:"final_wave_completed_nodes"`
		FinalWaveReadyNodes             int      `json:"final_wave_ready_nodes"`
		FinalWaveBlockedNodes           int      `json:"final_wave_blocked_nodes"`
		FinalWaveFailedNodes            int      `json:"final_wave_failed_nodes"`
		FinalWaveFinalResponseAllowed   bool     `json:"final_wave_final_response_allowed"`
		PublicSafetyScanStatus          string   `json:"public_safety_scan_status"`
		PromoterNoPromotionStatus       string   `json:"promoter_no_promotion_status"`
		CommandReadbackStatus           string   `json:"command_readback_status"`
		ClosureArtifactPaths            []string `json:"closure_artifact_paths"`
		ConsolidationCompletedBefore    int      `json:"consolidation_completed_before"`
		ConsolidationReadyBefore        int      `json:"consolidation_ready_before"`
		ConsolidationFirstNodeBefore    string   `json:"consolidation_first_executable_node_before"`
		ConsolidationFinalAllowedBefore bool     `json:"consolidation_final_response_allowed_before"`
		NextConsolidationNode           string   `json:"next_consolidation_node"`
		ClaimsAuthorityAdvance          bool     `json:"claims_authority_advance"`
		RSIRemainsDenied                bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-10", "final-operator-summary.json"))

	if nodeNineLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeNineLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-09" ||
		nodeNineLifecycle.Status != "merged_and_cleaned" ||
		nodeNineLifecycle.PRNumber != 312 ||
		nodeNineLifecycle.MergeCommit != "391d921fdf02c4ab0bd10f15001759dbb1f551ed" ||
		nodeNineLifecycle.CIStatus != "passed" ||
		!nodeNineLifecycle.LocalMainSynced ||
		!nodeNineLifecycle.LocalBranchDeleted ||
		!nodeNineLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 9 lifecycle evidence must prove clean branch handoff: %#v", nodeNineLifecycle)
	}
	if closure.Schema != "ao.atlas.final-closure-artifacts-fixture.v0.1" ||
		closure.Status != "final_closure_artifacts_recorded" ||
		closure.PromoterNoPromotionStatus != "no_promotion_requested" ||
		closure.CommandReadbackStatus != "readback_agrees_no_promotion" ||
		closure.ClaimsAuthorityAdvance ||
		!closure.RSIRemainsDenied ||
		!closure.FinalResponseAllowedAfter ||
		closure.ReadyNodesAfterNode != 0 ||
		closure.BlockedNodesAfterNode != 0 {
		t.Fatalf("source closure fixture must preserve final no-promotion closure: %#v", closure)
	}
	if summary.Schema != "ao.atlas.final-operator-summary.v0.1" ||
		summary.NodeID != "mission-recommendation-final-closure-consolidation-10" ||
		summary.Status != "generated" ||
		summary.SourceFinalReadbackPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/recommendation-readback-after.json" ||
		summary.SourceClosureFixturePath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/final-closure-artifacts-fixture.json" ||
		summary.SummaryMarkdownPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-10/final-operator-summary.md" ||
		summary.FinalWaveCompletedNodes != finalReadback.CompletedNodes ||
		summary.FinalWaveReadyNodes != finalReadback.ReadyNodes ||
		summary.FinalWaveBlockedNodes != finalReadback.BlockedNodes ||
		summary.FinalWaveFailedNodes != finalReadback.FailedNodes ||
		summary.FinalWaveFinalResponseAllowed != finalReadback.FinalResponseAllowed ||
		summary.PublicSafetyScanStatus != finalReadback.PublicSafetyScanStatus ||
		summary.PromoterNoPromotionStatus != closure.PromoterNoPromotionStatus ||
		summary.CommandReadbackStatus != closure.CommandReadbackStatus ||
		len(summary.ClosureArtifactPaths) != 5 ||
		summary.ConsolidationCompletedBefore != nodeNineReadback.CompletedNodes ||
		summary.ConsolidationReadyBefore != nodeNineReadback.ReadyNodes ||
		summary.ConsolidationFirstNodeBefore != nodeNineReadback.FirstExecutableNode ||
		summary.ConsolidationFinalAllowedBefore != nodeNineReadback.FinalResponseAllowed ||
		summary.NextConsolidationNode != "mission-recommendation-final-closure-consolidation-11" ||
		summary.ClaimsAuthorityAdvance ||
		!summary.RSIRemainsDenied {
		t.Fatalf("final operator summary must bind final readback, closure fixture, and consolidation continuation: %#v", summary)
	}
}

func TestFinalClosureConsolidationOperatorSummaryRegressionPreservesCountsNextActionAndNoPromotionWording(t *testing.T) {
	repoRoot := repoRoot(t)
	consolidationRoot := filepath.Join(repoRoot, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-10", "post-merge-lifecycle.json"))
	summary := mustLoadJSON[struct {
		FinalWaveCompletedNodes       int    `json:"final_wave_completed_nodes"`
		FinalWaveReadyNodes           int    `json:"final_wave_ready_nodes"`
		FinalWaveFinalResponseAllowed bool   `json:"final_wave_final_response_allowed"`
		PromoterNoPromotionStatus     string `json:"promoter_no_promotion_status"`
		CommandReadbackStatus         string `json:"command_readback_status"`
		NextConsolidationNode         string `json:"next_consolidation_node"`
		ClaimsAuthorityAdvance        bool   `json:"claims_authority_advance"`
		RSIRemainsDenied              bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-10", "final-operator-summary.json"))
	markdownBytes, err := os.ReadFile(filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-10", "final-operator-summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(markdownBytes)
	fixture := mustLoadJSON[struct {
		Schema                          string `json:"schema"`
		NodeID                          string `json:"node_id"`
		Status                          string `json:"status"`
		SourceSummaryPath               string `json:"source_summary_path"`
		SourceSummaryMarkdownPath       string `json:"source_summary_markdown_path"`
		CompletedNodesWordingPresent    bool   `json:"completed_nodes_wording_present"`
		NextActionWordingPresent        bool   `json:"next_action_wording_present"`
		NoPromotionWordingPresent       bool   `json:"no_promotion_wording_present"`
		RSIDenialWordingPresent         bool   `json:"rsi_denial_wording_present"`
		ClaimsAuthorityAdvance          bool   `json:"claims_authority_advance"`
		ExpectedCompletedNodes          int    `json:"expected_completed_nodes"`
		ExpectedNextConsolidationNode   string `json:"expected_next_consolidation_node"`
		ExpectedPromoterStatus          string `json:"expected_promoter_status"`
		ExpectedCommandStatus           string `json:"expected_command_status"`
		ConsolidationCompletedBefore    int    `json:"consolidation_completed_before"`
		ConsolidationReadyBefore        int    `json:"consolidation_ready_before"`
		ConsolidationFinalAllowedBefore bool   `json:"consolidation_final_response_allowed_before"`
		RSIRemainsDenied                bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-11", "operator-summary-regression.json"))
	nodeTenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-10", "recommendation-readback-after.json"))

	if nodeTenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-10" ||
		nodeTenLifecycle.Status != "merged_and_cleaned" ||
		nodeTenLifecycle.PRNumber != 313 ||
		nodeTenLifecycle.MergeCommit != "78f17411cc08214db3d2842f28939036b805a712" ||
		nodeTenLifecycle.CIStatus != "passed" ||
		!nodeTenLifecycle.LocalMainSynced ||
		!nodeTenLifecycle.LocalBranchDeleted ||
		!nodeTenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 10 lifecycle evidence must prove clean branch handoff: %#v", nodeTenLifecycle)
	}
	if summary.FinalWaveCompletedNodes != 40 ||
		summary.FinalWaveReadyNodes != 0 ||
		!summary.FinalWaveFinalResponseAllowed ||
		summary.PromoterNoPromotionStatus != "no_promotion_requested" ||
		summary.CommandReadbackStatus != "readback_agrees_no_promotion" ||
		summary.NextConsolidationNode != "mission-recommendation-final-closure-consolidation-11" ||
		summary.ClaimsAuthorityAdvance ||
		!summary.RSIRemainsDenied {
		t.Fatalf("source operator summary drifted before regression fixture: %#v", summary)
	}
	for _, want := range []string{
		"40 completed nodes",
		"Promoter status is `no_promotion_requested`",
		"Command status is `readback_agrees_no_promotion`",
		"next executable node is `mission-recommendation-final-closure-consolidation-11`",
		"keeps RSI denied",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("operator summary markdown missing %q:\n%s", want, markdown)
		}
	}
	if fixture.Schema != "ao.atlas.operator-summary-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-11" ||
		fixture.Status != "guarded" ||
		fixture.SourceSummaryPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-10/final-operator-summary.json" ||
		fixture.SourceSummaryMarkdownPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-10/final-operator-summary.md" ||
		!fixture.CompletedNodesWordingPresent ||
		!fixture.NextActionWordingPresent ||
		!fixture.NoPromotionWordingPresent ||
		!fixture.RSIDenialWordingPresent ||
		fixture.ClaimsAuthorityAdvance ||
		fixture.ExpectedCompletedNodes != summary.FinalWaveCompletedNodes ||
		fixture.ExpectedNextConsolidationNode != summary.NextConsolidationNode ||
		fixture.ExpectedPromoterStatus != summary.PromoterNoPromotionStatus ||
		fixture.ExpectedCommandStatus != summary.CommandReadbackStatus ||
		fixture.ConsolidationCompletedBefore != nodeTenReadback.CompletedNodes ||
		fixture.ConsolidationReadyBefore != nodeTenReadback.ReadyNodes ||
		fixture.ConsolidationFinalAllowedBefore != nodeTenReadback.FinalResponseAllowed ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("operator summary regression fixture must preserve wording and readback continuity: %#v", fixture)
	}
}

func TestFinalClosureConsolidationEvidenceValidationCommandCoversCompletedHardeningWave(t *testing.T) {
	repoRoot := repoRoot(t)
	hardeningRoot := filepath.Join(repoRoot, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	outPath := filepath.Join(t.TempDir(), "evidence-validation-report.json")
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{
		"mission", "recommendations", "validate-evidence",
		"--evidence-root", hardeningRoot,
		"--out", outPath,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("validate-evidence failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	report := mustLoadJSON[struct {
		Schema                   string         `json:"schema"`
		Status                   string         `json:"status"`
		EvidenceRoot             string         `json:"evidence_root"`
		NodeCount                int            `json:"node_count"`
		JSONFileCount            int            `json:"json_file_count"`
		ValidatedJSONFiles       int            `json:"validated_json_files"`
		SchemaBoundFiles         int            `json:"schema_bound_files"`
		TypedValidatorFiles      int            `json:"typed_validator_files"`
		GenericSchemaFiles       int            `json:"generic_schema_files"`
		MissingSchemaFiles       []string       `json:"missing_schema_files"`
		FailedFiles              []string       `json:"failed_files"`
		RequiredFilenames        []string       `json:"required_filenames"`
		RequiredFilenamesCovered bool           `json:"required_filenames_covered"`
		SchemaCounts             map[string]int `json:"schema_counts"`
		Validators               map[string]int `json:"validators"`
	}](t, outPath)
	if report.Schema != "ao.atlas.recommendation-evidence-validation-report.v0.1" ||
		report.Status != "passed" ||
		report.EvidenceRoot != filepath.ToSlash(hardeningRoot) ||
		report.NodeCount != 40 ||
		report.JSONFileCount != 786 ||
		report.ValidatedJSONFiles != report.JSONFileCount ||
		report.SchemaBoundFiles != report.JSONFileCount ||
		report.TypedValidatorFiles < 240 ||
		report.GenericSchemaFiles <= 0 ||
		len(report.MissingSchemaFiles) != 0 ||
		len(report.FailedFiles) != 0 ||
		!report.RequiredFilenamesCovered ||
		len(report.RequiredFilenames) < 8 ||
		report.SchemaCounts["ao.atlas.run-link.v0.1"] != 40 ||
		report.SchemaCounts["ao.atlas.workgraph.v0.1"] != 40 ||
		report.SchemaCounts["ao.atlas.recommendation-readback.v0.1"] != 44 ||
		report.SchemaCounts["ao.atlas.long-recommendation-wave-execution.v0.3"] != 44 ||
		report.Validators["typed:run-link"] != 40 ||
		report.Validators["typed:workgraph"] != 40 ||
		report.Validators["typed:recommendation-readback"] != 44 {
		t.Fatalf("evidence validation report must cover every completed hardening node evidence JSON: %#v", report)
	}
	if !strings.Contains(stdout.String(), "status=passed") ||
		!strings.Contains(stdout.String(), "json_files=786") ||
		!strings.Contains(stdout.String(), "node_count=40") {
		t.Fatalf("validate-evidence stdout missing summary: %s", stdout.String())
	}
}

func TestFinalClosureConsolidationSchemaValidationCoverageRegressionBindsGatesRunLinksAndReadbacks(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTwelveLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-12", "post-merge-lifecycle.json"))
	report := mustLoadJSON[struct {
		Status              string `json:"status"`
		NodeCount           int    `json:"node_count"`
		JSONFileCount       int    `json:"json_file_count"`
		SchemaBoundFiles    int    `json:"schema_bound_files"`
		TypedValidatorFiles int    `json:"typed_validator_files"`
		Entries             []struct {
			Path      string `json:"path"`
			NodeID    string `json:"node_id"`
			Filename  string `json:"filename"`
			Schema    string `json:"schema"`
			Validator string `json:"validator"`
			Status    string `json:"status"`
		} `json:"entries"`
		Validators map[string]int `json:"validators"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-12", "completed-wave-evidence-validation-report.json"))
	fixture := mustLoadJSON[struct {
		Schema                                  string         `json:"schema"`
		NodeID                                  string         `json:"node_id"`
		Status                                  string         `json:"status"`
		SourceReportPath                        string         `json:"source_report_path"`
		CompletedWaveNodes                      int            `json:"completed_wave_nodes"`
		JSONFileCount                           int            `json:"json_file_count"`
		SchemaBoundFiles                        int            `json:"schema_bound_files"`
		NodeGateFiles                           int            `json:"node_gate_files"`
		RunLinkFiles                            int            `json:"run_link_files"`
		RecommendationReadbackAfterFiles        int            `json:"recommendation_readback_after_files"`
		CheckpointReadbackAfterFiles            int            `json:"checkpoint_readback_after_files"`
		ExecutionReadbackAfterFiles             int            `json:"execution_readback_after_files"`
		WorkgraphAfterFiles                     int            `json:"workgraph_after_files"`
		AllNodeGateRunLinkReadbackCoverage      bool           `json:"all_node_gate_run_link_readback_coverage"`
		TypedValidatorCounts                    map[string]int `json:"typed_validator_counts"`
		ConsolidationCompletedBefore            int            `json:"consolidation_completed_before"`
		ConsolidationReadyBefore                int            `json:"consolidation_ready_before"`
		ConsolidationFirstExecutableNodeBefore  string         `json:"consolidation_first_executable_node_before"`
		ConsolidationFinalResponseAllowedBefore bool           `json:"consolidation_final_response_allowed_before"`
		RSIRemainsDenied                        bool           `json:"rsi_remains_denied"`
	}](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-13", "schema-validation-coverage-regression.json"))
	nodeTwelveReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-12", "recommendation-readback-after.json"))

	if nodeTwelveLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTwelveLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-12" ||
		nodeTwelveLifecycle.Status != "merged_and_cleaned" ||
		nodeTwelveLifecycle.PRNumber != 315 ||
		nodeTwelveLifecycle.MergeCommit != "65d3bd0eb65c4b16d35f1bdb7da44ac22f3313f2" ||
		nodeTwelveLifecycle.CIStatus != "passed" ||
		!nodeTwelveLifecycle.LocalMainSynced ||
		!nodeTwelveLifecycle.LocalBranchDeleted ||
		!nodeTwelveLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 12 lifecycle evidence must prove clean branch handoff: %#v", nodeTwelveLifecycle)
	}
	filenameCounts := map[string]int{}
	for _, entry := range report.Entries {
		if entry.Status != "passed" || entry.Schema == "" || entry.Path == "" || entry.NodeID == "" {
			t.Fatalf("validation report entry must be passed and schema-bound: %#v", entry)
		}
		filenameCounts[entry.Filename]++
	}
	for _, filename := range []string{
		"node_gate.json",
		"run-link.json",
		"recommendation-readback-after.json",
		"checkpoint-readback-after.json",
		"execution-readback-after.json",
		"workgraph-after.json",
	} {
		if filenameCounts[filename] != 40 {
			t.Fatalf("validation report must cover %s for all 40 nodes, got %d", filename, filenameCounts[filename])
		}
	}
	if fixture.Schema != "ao.atlas.schema-validation-coverage-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-13" ||
		fixture.Status != "guarded" ||
		fixture.SourceReportPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-12/completed-wave-evidence-validation-report.json" ||
		fixture.CompletedWaveNodes != report.NodeCount ||
		fixture.JSONFileCount != report.JSONFileCount ||
		fixture.SchemaBoundFiles != report.SchemaBoundFiles ||
		fixture.NodeGateFiles != filenameCounts["node_gate.json"] ||
		fixture.RunLinkFiles != filenameCounts["run-link.json"] ||
		fixture.RecommendationReadbackAfterFiles != filenameCounts["recommendation-readback-after.json"] ||
		fixture.CheckpointReadbackAfterFiles != filenameCounts["checkpoint-readback-after.json"] ||
		fixture.ExecutionReadbackAfterFiles != filenameCounts["execution-readback-after.json"] ||
		fixture.WorkgraphAfterFiles != filenameCounts["workgraph-after.json"] ||
		!fixture.AllNodeGateRunLinkReadbackCoverage ||
		fixture.TypedValidatorCounts["typed:run-link"] != report.Validators["typed:run-link"] ||
		fixture.TypedValidatorCounts["typed:recommendation-readback"] != report.Validators["typed:recommendation-readback"] ||
		fixture.TypedValidatorCounts["typed:recommendation-checkpoint-readback"] != report.Validators["typed:recommendation-checkpoint-readback"] ||
		fixture.TypedValidatorCounts["typed:recommendation-execution-readback"] != report.Validators["typed:recommendation-execution-readback"] ||
		fixture.TypedValidatorCounts["typed:workgraph"] != report.Validators["typed:workgraph"] ||
		fixture.ConsolidationCompletedBefore != nodeTwelveReadback.CompletedNodes ||
		fixture.ConsolidationReadyBefore != nodeTwelveReadback.ReadyNodes ||
		fixture.ConsolidationFirstExecutableNodeBefore != nodeTwelveReadback.FirstExecutableNode ||
		fixture.ConsolidationFinalResponseAllowedBefore != nodeTwelveReadback.FinalResponseAllowed ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("schema validation coverage fixture must bind gates, run-links, readbacks, typed validators, and readback continuity: %#v", fixture)
	}
}

func TestFinalClosureConsolidationWaveSeedsTwentyFourSerializedAuditNodes(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "ao-atlas-final-closure-consolidation-wave-v01" ||
		wave.TotalTasks != 24 ||
		wave.MinimumTasks != 24 ||
		wave.NodeBudget != 24 ||
		wave.EstimatedMinutes != 150 ||
		wave.FinalResponseAllowed ||
		wave.PublicSafetyScanStatus != "required_pending_verification" ||
		wave.PromoterReadbackStatus != "required_not_bound" ||
		wave.CommandReadbackStatus != "required_not_bound" {
		t.Fatalf("consolidation wave must preserve the 24-node closure budget and blocked final response: %#v", wave)
	}
	if wave.Supervisor == nil ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 24 ||
		wave.Supervisor.ReturnOnlyWhen != "all_24_nodes_complete_or_true_hard_blocker" {
		t.Fatalf("consolidation supervisor must preserve the long-run lease contract: %#v", wave.Supervisor)
	}
	for _, want := range []string{
		"Bind final readback public safety scan status",
		"Generate post merge cleanup evidence",
		"machine readable pull request and continuous integration ledger",
		"compaction resume prompt",
		"at least forty ranked Feature Depth tasks",
	} {
		found := false
		for _, task := range wave.Tasks {
			if strings.Contains(task.Task, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("consolidation wave missing task theme %q", want)
		}
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(root, "recommendation-workgraph.json"))
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 24 ||
		workgraph.Nodes[0].ID != "mission-recommendation-final-closure-consolidation-01" ||
		workgraph.Nodes[23].ID != "mission-recommendation-final-closure-consolidation-24" {
		t.Fatalf("consolidation workgraph must contain the expected serialized node ids: %#v", workgraph.Nodes)
	}
	for i, node := range workgraph.Nodes {
		if node.Status != "ready" {
			t.Fatalf("node %d must start ready: %#v", i, node)
		}
		if i == 0 && len(node.Dependencies) != 0 {
			t.Fatalf("first node must have no dependency: %#v", node)
		}
		if i > 0 {
			wantDependency := workgraph.Nodes[i-1].ID
			if len(node.Dependencies) != 1 || node.Dependencies[0] != wantDependency {
				t.Fatalf("node %s must depend on %s: %#v", node.ID, wantDependency, node.Dependencies)
			}
		}
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "recommendation-readback.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 0 ||
		readback.ReadyNodes != 24 ||
		readback.BlockedNodes != 0 ||
		readback.FailedNodes != 0 ||
		readback.FirstExecutableNode != "mission-recommendation-final-closure-consolidation-01" ||
		readback.FinalResponseAllowed ||
		readback.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		len(readback.FeatureDepthRecommendations) != 24 {
		t.Fatalf("initial consolidation readback must point at node 1 and deny final response: %#v", readback)
	}
}

func digestFileWithNormalizedLineEndings(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	return DigestBytes(data), nil
}

func TestMissionRecommendationsRejectMixedOwnerDefaultWaveWithExactReadback(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	var bundle AOMissionFeatureDepthRecommendations
	if err := readJSONIfPossible(recommendationsPath, &bundle); err != nil {
		t.Fatal(err)
	}
	bundle.Tasks[39].Owner = "ao-foundry"
	if err := WriteJSON(recommendationsPath, bundle); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", filepath.Join(dir, "out"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("mixed-owner default wave was accepted")
	}
	if !strings.Contains(out.String(), "requires at least 30 AO Atlas-owned tasks and 40 tasks for continue-if-fast target") {
		t.Fatalf("mixed-owner error did not report exact readback: %s", out.String())
	}
}

func TestProductionReadinessExercisesMissionRecommendationsImport(t *testing.T) {
	root := repoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"mission recommendations import",
		"--recommendations examples/valid/ao-mission/feature-depth-recommendations.json",
		"--min-tasks 30",
		"--min-minutes 120",
		"--max-minutes 180",
		"--continue-if-fast-target 40",
		"recommendation-workgraph.json",
		"lease-start.json",
		"recommendation-readback.json",
		"mission recommendations readback",
		"mission recommendations complete-node",
		"mission recommendations resume",
		"--lease-start",
		"--elapsed-minutes",
		"--started-at",
		"--completed-at",
		"--lease-timing-mode",
		"--out-checkpoint-readback",
		"checkpoint-readback-after-node-01.json",
		"command-readback-resumed.json",
		"promoter-readback-resumed.json",
		"foundry-rollup-resumed.json",
		"reconciliation-packet-resumed.json",
		"--out-reconciliation-packet",
		"recommendation-reconciliation-packet.schema.json",
		"recommendation-lease-start.schema.json",
		"recommendation-checkpoint-readback.schema.json",
		"recommendation-command-readback.schema.json",
		"recommendation-promoter-readback.schema.json",
		"recommendation-foundry-rollup.schema.json",
		"minimum_minutes_unmet",
		"lease_timing_missing",
		"minimum_minutes_met",
		"--out-execution-readback",
		"completed_recommendation_nodes",
		"checkpoint_count",
		"return_gate_status",
		"blocked_ready_nodes_remain",
		"blocked_minimum_minutes_unmet",
		"blocked_lease_timing_missing",
		"final_response_allowed",
		"min_minutes_met=true",
		"recommendation-ledger-consistency",
		"next-recommended-prompt.md",
		"reject_generated_recommendation_prompt_public_safety",
		"recommendation-prompt-public-safety-scan",
		"execution-readback-regenerated.json",
		"reason_artifact_agreement_summary",
		"generated-recommendation-prompt-continuation-reason-negative-scan",
		"lease-resume-wave-public-safety-readback",
		"lease_resume_root=\"docs/evidence/ao-atlas-lease-resume-wave-v01\"",
		"final-synthesis.json",
		"Current workgraph:",
		"early_return_risk_status",
		"Early-return risk:",
		"do not produce a final response",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing recommendation coverage %q", want)
		}
	}
}

func TestProductionReadinessGuardsCommittedBuildArtifacts(t *testing.T) {
	root := repoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"reject_local_build_artifacts",
		"build_artifacts=(",
		"atlas.exe",
		"cmd/atlas/atlas",
		"git ls-files --error-unmatch",
		"tracked build artifact present",
		"local build artifact present",
		"build-artifact-guard",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing build artifact guard coverage %q", want)
		}
	}
}

func TestFinalClosureConsolidationBuildArtifactGuardRegressionReportsBeforePromotionClosure(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeFourteenDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-final-closure-consolidation-14")
	nodeFifteenDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-final-closure-consolidation-15")

	guard := mustLoadJSON[struct {
		Schema             string   `json:"schema"`
		NodeID             string   `json:"node_id"`
		Status             string   `json:"status"`
		ReadinessCheckName string   `json:"readiness_check_name"`
		GuardedArtifacts   []string `json:"guarded_artifacts"`
		FailureMessages    []string `json:"failure_messages"`
		RunsBeforeGoBuild  bool     `json:"runs_before_go_build"`
		RSIRemainsDenied   bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeFourteenDir, "build-artifact-guard.json"))

	fixture := mustLoadJSON[struct {
		Schema                        string   `json:"schema"`
		NodeID                        string   `json:"node_id"`
		Status                        string   `json:"status"`
		SourceGuardNodeID             string   `json:"source_guard_node_id"`
		SourceGuardArtifact           string   `json:"source_guard_artifact"`
		ReadinessCheckName            string   `json:"readiness_check_name"`
		ReportedBeforePromotionClose  bool     `json:"reported_before_promotion_closure"`
		PromotionClosureDependsOn     []string `json:"promotion_closure_depends_on"`
		GuardedArtifacts              []string `json:"guarded_artifacts"`
		FailureMessages               []string `json:"failure_messages"`
		PromotionRequested            bool     `json:"promotion_requested"`
		PromotionGranted              bool     `json:"promotion_granted"`
		FinalResponseAllowed          bool     `json:"final_response_allowed"`
		ExpectedNextNodeAfterComplete string   `json:"expected_next_node_after_completion"`
		RSIRemainsDenied              bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeFifteenDir, "build-artifact-promotion-closure-regression.json"))

	if guard.Schema != "ao.atlas.build-artifact-guard.v0.1" ||
		guard.NodeID != "mission-recommendation-final-closure-consolidation-14" ||
		guard.Status != "guarded" ||
		guard.ReadinessCheckName != "build-artifact-guard" ||
		!guard.RunsBeforeGoBuild ||
		!guard.RSIRemainsDenied {
		t.Fatalf("node 14 guard artifact is not promotion-closure ready: %#v", guard)
	}
	if !containsString(guard.GuardedArtifacts, "atlas") ||
		!containsString(guard.GuardedArtifacts, "atlas.exe") ||
		!containsString(guard.FailureMessages, "tracked build artifact present") ||
		!containsString(guard.FailureMessages, "local build artifact present") {
		t.Fatalf("node 14 guard must report local and tracked build artifact failures: %#v", guard)
	}
	if fixture.Schema != "ao.atlas.build-artifact-promotion-closure-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-15" ||
		fixture.Status != "guarded" ||
		fixture.SourceGuardNodeID != guard.NodeID ||
		fixture.SourceGuardArtifact != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-14/build-artifact-guard.json" ||
		fixture.ReadinessCheckName != guard.ReadinessCheckName ||
		!fixture.ReportedBeforePromotionClose ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		fixture.FinalResponseAllowed ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-16" ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 15 build-artifact promotion closure regression is not guarded: %#v", fixture)
	}
	for _, requiredDependency := range []string{
		"build-artifact-guard",
		"promoter_no_promotion",
		"command_readback",
	} {
		if !containsString(fixture.PromotionClosureDependsOn, requiredDependency) {
			t.Fatalf("node 15 fixture must bind promotion closure dependency %q: %#v", requiredDependency, fixture.PromotionClosureDependsOn)
		}
	}
	for _, artifact := range guard.GuardedArtifacts {
		if !containsString(fixture.GuardedArtifacts, artifact) {
			t.Fatalf("node 15 fixture lost guarded artifact %q: %#v", artifact, fixture.GuardedArtifacts)
		}
	}
	for _, message := range guard.FailureMessages {
		if !containsString(fixture.FailureMessages, message) {
			t.Fatalf("node 15 fixture lost guard failure message %q: %#v", message, fixture.FailureMessages)
		}
	}
}

func TestFinalClosureConsolidationWindowsCIWaitStateTelemetryRecordsLongRunningChecks(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeFifteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-15")
	nodeSixteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-16")

	nodeFifteenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeFifteenDir, "post-merge-lifecycle.json"))
	nodeFifteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeFifteenDir, "recommendation-readback-after.json"))
	telemetry := mustLoadJSON[struct {
		Schema                        string `json:"schema"`
		NodeID                        string `json:"node_id"`
		Status                        string `json:"status"`
		Source                        string `json:"source"`
		LongRunningOS                 string `json:"long_running_os"`
		WaitThresholdSeconds          int    `json:"wait_threshold_seconds"`
		SourcePRs                     []int  `json:"source_prs"`
		WindowsCheckSampleCount       int    `json:"windows_check_sample_count"`
		PendingStateObserved          bool   `json:"pending_state_observed"`
		CompletedPassStateObserved    bool   `json:"completed_pass_state_observed"`
		FailedStateObserved           bool   `json:"failed_state_observed"`
		MaxObservedDurationSeconds    int    `json:"max_observed_duration_seconds"`
		CompletedNodesBefore          int    `json:"completed_nodes_before"`
		ReadyNodesBefore              int    `json:"ready_nodes_before"`
		FirstExecutableNodeBefore     string `json:"first_executable_node_before"`
		FinalResponseAllowedBefore    bool   `json:"final_response_allowed_before"`
		ExpectedNextNodeAfterComplete string `json:"expected_next_node_after_completion"`
		PromotionRequested            bool   `json:"promotion_requested"`
		PromotionGranted              bool   `json:"promotion_granted"`
		RSIRemainsDenied              bool   `json:"rsi_remains_denied"`
		WindowsCheckDurationSamples   []struct {
			PRNumber        int    `json:"pr_number"`
			CheckName       string `json:"check_name"`
			FinalStatus     string `json:"final_status"`
			FinalConclusion string `json:"final_conclusion"`
			DurationSeconds int    `json:"duration_seconds"`
			WaitState       string `json:"wait_state"`
		} `json:"windows_check_duration_samples"`
	}](t, filepath.Join(nodeSixteenDir, "windows-ci-wait-state-telemetry.json"))

	if nodeFifteenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeFifteenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-15" ||
		nodeFifteenLifecycle.Status != "merged_and_cleaned" ||
		nodeFifteenLifecycle.PRNumber != 318 ||
		nodeFifteenLifecycle.MergeCommit != "8c708461a563550e0e2df5fc8eed5774d23ca7fd" ||
		nodeFifteenLifecycle.CIStatus != "passed" ||
		!nodeFifteenLifecycle.LocalMainSynced ||
		!nodeFifteenLifecycle.LocalBranchDeleted ||
		!nodeFifteenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 15 lifecycle evidence must prove clean branch handoff before node 16 telemetry: %#v", nodeFifteenLifecycle)
	}
	if telemetry.Schema != "ao.atlas.windows-ci-wait-state-telemetry.v0.1" ||
		telemetry.NodeID != "mission-recommendation-final-closure-consolidation-16" ||
		telemetry.Status != "recorded" ||
		telemetry.Source != "gh_pr_checks_317_318" ||
		telemetry.LongRunningOS != "windows-latest" ||
		telemetry.WaitThresholdSeconds != 600 ||
		telemetry.WindowsCheckSampleCount != len(telemetry.WindowsCheckDurationSamples) ||
		telemetry.WindowsCheckSampleCount < 6 ||
		!containsInt(telemetry.SourcePRs, 317) ||
		!containsInt(telemetry.SourcePRs, 318) ||
		!telemetry.PendingStateObserved ||
		!telemetry.CompletedPassStateObserved ||
		telemetry.FailedStateObserved ||
		telemetry.MaxObservedDurationSeconds < telemetry.WaitThresholdSeconds ||
		telemetry.CompletedNodesBefore != nodeFifteenReadback.CompletedNodes ||
		telemetry.ReadyNodesBefore != nodeFifteenReadback.ReadyNodes ||
		telemetry.FirstExecutableNodeBefore != nodeFifteenReadback.FirstExecutableNode ||
		telemetry.FinalResponseAllowedBefore != nodeFifteenReadback.FinalResponseAllowed ||
		telemetry.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-17" ||
		telemetry.PromotionRequested ||
		telemetry.PromotionGranted ||
		!telemetry.RSIRemainsDenied {
		t.Fatalf("node 16 telemetry fixture must bind long-running Windows CI state without promotion: %#v", telemetry)
	}
	for _, sample := range telemetry.WindowsCheckDurationSamples {
		if sample.PRNumber != 317 && sample.PRNumber != 318 {
			t.Fatalf("unexpected telemetry PR sample: %#v", sample)
		}
		if !strings.Contains(sample.CheckName, "windows-latest") ||
			sample.FinalStatus != "COMPLETED" ||
			sample.FinalConclusion != "SUCCESS" ||
			sample.DurationSeconds < telemetry.WaitThresholdSeconds ||
			sample.WaitState != "long_running_pending_before_success" {
			t.Fatalf("Windows telemetry sample must report long-running pending-before-success state: %#v", sample)
		}
	}
}

func TestFinalClosureConsolidationWindowsCIStateRegressionCoversPendingPassingAndFailingStates(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeSixteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-16")
	nodeSeventeenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-17")

	nodeSixteenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeSixteenDir, "post-merge-lifecycle.json"))
	nodeSixteenReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeSixteenDir, "recommendation-readback-after.json"))
	telemetry := mustLoadJSON[struct {
		Schema                     string `json:"schema"`
		NodeID                     string `json:"node_id"`
		Status                     string `json:"status"`
		WindowsCheckSampleCount    int    `json:"windows_check_sample_count"`
		PendingStateObserved       bool   `json:"pending_state_observed"`
		CompletedPassStateObserved bool   `json:"completed_pass_state_observed"`
		FailedStateObserved        bool   `json:"failed_state_observed"`
	}](t, filepath.Join(nodeSixteenDir, "windows-ci-wait-state-telemetry.json"))
	fixture := mustLoadJSON[struct {
		Schema                        string   `json:"schema"`
		NodeID                        string   `json:"node_id"`
		Status                        string   `json:"status"`
		SourceTelemetryPath           string   `json:"source_telemetry_path"`
		SourceTelemetrySampleCount    int      `json:"source_telemetry_sample_count"`
		CoveredWindowsStates          []string `json:"covered_windows_states"`
		PendingStateCovered           bool     `json:"pending_state_covered"`
		PassingStateCovered           bool     `json:"passing_state_covered"`
		FailingStateCovered           bool     `json:"failing_state_covered"`
		SyntheticFailureFixture       bool     `json:"synthetic_failure_fixture"`
		CompletedNodesBefore          int      `json:"completed_nodes_before"`
		ReadyNodesBefore              int      `json:"ready_nodes_before"`
		FirstExecutableNodeBefore     string   `json:"first_executable_node_before"`
		FinalResponseAllowedBefore    bool     `json:"final_response_allowed_before"`
		ExpectedNextNodeAfterComplete string   `json:"expected_next_node_after_completion"`
		PromotionRequested            bool     `json:"promotion_requested"`
		PromotionGranted              bool     `json:"promotion_granted"`
		RSIRemainsDenied              bool     `json:"rsi_remains_denied"`
		StateSamples                  []struct {
			State                string `json:"state"`
			GitHubStatus         string `json:"github_status"`
			GitHubConclusion     string `json:"github_conclusion"`
			OperatorAction       string `json:"operator_action"`
			FinalResponseAllowed bool   `json:"final_response_allowed"`
		} `json:"state_samples"`
	}](t, filepath.Join(nodeSeventeenDir, "windows-ci-state-regression.json"))

	if nodeSixteenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeSixteenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-16" ||
		nodeSixteenLifecycle.Status != "merged_and_cleaned" ||
		nodeSixteenLifecycle.PRNumber != 319 ||
		nodeSixteenLifecycle.MergeCommit != "169baa5990fa9726fe42526968af5c1892a52446" ||
		nodeSixteenLifecycle.CIStatus != "passed" ||
		!nodeSixteenLifecycle.LocalMainSynced ||
		!nodeSixteenLifecycle.LocalBranchDeleted ||
		!nodeSixteenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 16 lifecycle evidence must prove clean branch handoff before node 17 state regression: %#v", nodeSixteenLifecycle)
	}
	if telemetry.Schema != "ao.atlas.windows-ci-wait-state-telemetry.v0.1" ||
		telemetry.NodeID != "mission-recommendation-final-closure-consolidation-16" ||
		telemetry.Status != "recorded" ||
		telemetry.WindowsCheckSampleCount < 6 ||
		!telemetry.PendingStateObserved ||
		!telemetry.CompletedPassStateObserved ||
		telemetry.FailedStateObserved {
		t.Fatalf("node 16 telemetry must provide pending and passing source states without real failure: %#v", telemetry)
	}
	if fixture.Schema != "ao.atlas.windows-ci-state-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-17" ||
		fixture.Status != "guarded" ||
		fixture.SourceTelemetryPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-16/windows-ci-wait-state-telemetry.json" ||
		fixture.SourceTelemetrySampleCount != telemetry.WindowsCheckSampleCount ||
		!fixture.PendingStateCovered ||
		!fixture.PassingStateCovered ||
		!fixture.FailingStateCovered ||
		!fixture.SyntheticFailureFixture ||
		fixture.CompletedNodesBefore != nodeSixteenReadback.CompletedNodes ||
		fixture.ReadyNodesBefore != nodeSixteenReadback.ReadyNodes ||
		fixture.FirstExecutableNodeBefore != nodeSixteenReadback.FirstExecutableNode ||
		fixture.FinalResponseAllowedBefore != nodeSixteenReadback.FinalResponseAllowed ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-18" ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 17 Windows CI state regression must cover pending/passing/failing states without promotion: %#v", fixture)
	}
	for _, want := range []string{"pending", "passing", "failing"} {
		if !containsString(fixture.CoveredWindowsStates, want) {
			t.Fatalf("node 17 fixture missing covered Windows state %q: %#v", want, fixture.CoveredWindowsStates)
		}
	}
	seen := map[string]bool{}
	for _, sample := range fixture.StateSamples {
		seen[sample.State] = true
		switch sample.State {
		case "pending":
			if sample.GitHubStatus != "IN_PROGRESS" || sample.GitHubConclusion != "" || sample.OperatorAction != "wait_for_ci" || sample.FinalResponseAllowed {
				t.Fatalf("pending sample must force CI wait: %#v", sample)
			}
		case "passing":
			if sample.GitHubStatus != "COMPLETED" || sample.GitHubConclusion != "SUCCESS" || sample.OperatorAction != "merge_after_all_required_checks_pass" || sample.FinalResponseAllowed {
				t.Fatalf("passing sample must allow merge but not final response while ready work remains: %#v", sample)
			}
		case "failing":
			if sample.GitHubStatus != "COMPLETED" || sample.GitHubConclusion != "FAILURE" || sample.OperatorAction != "repair_before_merge" || sample.FinalResponseAllowed {
				t.Fatalf("failing sample must force repair before merge: %#v", sample)
			}
		default:
			t.Fatalf("unexpected Windows CI state sample: %#v", sample)
		}
	}
	for _, want := range []string{"pending", "passing", "failing"} {
		if !seen[want] {
			t.Fatalf("node 17 fixture missing state sample %q: %#v", want, fixture.StateSamples)
		}
	}
}

func TestFinalClosureConsolidationCompactionResumePromptBindsLatestReadback(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeSeventeenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-17")
	nodeEighteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-18")

	nodeSeventeenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeSeventeenDir, "post-merge-lifecycle.json"))
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeSeventeenDir, "recommendation-readback-after.json"))
	fixture := mustLoadJSON[struct {
		Schema                        string `json:"schema"`
		NodeID                        string `json:"node_id"`
		Status                        string `json:"status"`
		SourceReadbackPath            string `json:"source_readback_path"`
		PromptPath                    string `json:"prompt_path"`
		CompletedNodes                int    `json:"completed_nodes"`
		TotalNodes                    int    `json:"total_nodes"`
		ReadyNodes                    int    `json:"ready_nodes"`
		BlockedNodes                  int    `json:"blocked_nodes"`
		FailedNodes                   int    `json:"failed_nodes"`
		FirstExecutableNode           string `json:"first_executable_node"`
		ExactNextAction               string `json:"exact_next_action"`
		ReturnGateStatus              string `json:"return_gate_status"`
		ContinuationContractReason    string `json:"continuation_contract_reason"`
		EarlyReturnRiskStatus         string `json:"early_return_risk_status"`
		FinalResponseAllowed          bool   `json:"final_response_allowed"`
		RefusesFinalResponse          bool   `json:"refuses_final_response"`
		ExpectedNextNodeAfterComplete string `json:"expected_next_node_after_completion"`
		PromotionRequested            bool   `json:"promotion_requested"`
		PromotionGranted              bool   `json:"promotion_granted"`
		RSIRemainsDenied              bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeEighteenDir, "compaction-resume-prompt.json"))
	promptBytes, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(fixture.PromptPath)))
	if err != nil {
		t.Fatalf("read compaction resume prompt: %v", err)
	}
	prompt := string(promptBytes)

	if nodeSeventeenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeSeventeenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-17" ||
		nodeSeventeenLifecycle.Status != "merged_and_cleaned" ||
		nodeSeventeenLifecycle.PRNumber != 320 ||
		nodeSeventeenLifecycle.MergeCommit != "f12cb80dd960055fe51764d54570b5aa6affefee" ||
		nodeSeventeenLifecycle.CIStatus != "passed" ||
		!nodeSeventeenLifecycle.LocalMainSynced ||
		!nodeSeventeenLifecycle.LocalBranchDeleted ||
		!nodeSeventeenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 17 lifecycle evidence must prove clean branch handoff before node 18 resume prompt: %#v", nodeSeventeenLifecycle)
	}
	if fixture.Schema != "ao.atlas.compaction-resume-prompt.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-18" ||
		fixture.Status != "generated" ||
		fixture.SourceReadbackPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-17/recommendation-readback-after.json" ||
		fixture.PromptPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-18/compaction-resume-prompt.md" ||
		fixture.CompletedNodes != readback.CompletedNodes ||
		fixture.TotalNodes != readback.TotalNodes ||
		fixture.ReadyNodes != readback.ReadyNodes ||
		fixture.BlockedNodes != readback.BlockedNodes ||
		fixture.FailedNodes != readback.FailedNodes ||
		fixture.FirstExecutableNode != readback.FirstExecutableNode ||
		fixture.ExactNextAction != readback.ExactNextAction ||
		fixture.ReturnGateStatus != readback.ReturnGateStatus ||
		fixture.ContinuationContractReason != readback.ContinuationContract.Reason ||
		fixture.EarlyReturnRiskStatus != readback.EarlyReturnRiskStatus ||
		fixture.FinalResponseAllowed != readback.FinalResponseAllowed ||
		!fixture.RefusesFinalResponse ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-19" ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 18 compaction resume prompt fixture must bind latest readback without promotion: %#v", fixture)
	}
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas final-closure consolidation wave after context compaction.",
		"Current readback: `docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-17/recommendation-readback-after.json`",
		"Completed nodes: 17 / 24",
		"Ready nodes: 7",
		"Next executable node: `mission-recommendation-final-closure-consolidation-18`",
		"Final response allowed: `false`",
		"Early-return risk: `blocked_final_response_ready_nodes_remain`",
		"Emit Foundry import for mission-recommendation-final-closure-consolidation-18 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestFinalClosureConsolidationCompactionResumeRegressionPreservesNextNodeAndReturnGate(t *testing.T) {
	consolidationRoot := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeEighteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-18")
	nodeNineteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-19")

	nodeEighteenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeEighteenDir, "post-merge-lifecycle.json"))
	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeEighteenDir, "recommendation-readback-after.json"))
	promptFixture := mustLoadJSON[struct {
		Schema                string `json:"schema"`
		NodeID                string `json:"node_id"`
		FirstExecutableNode   string `json:"first_executable_node"`
		ExactNextAction       string `json:"exact_next_action"`
		ReturnGateStatus      string `json:"return_gate_status"`
		FinalResponseAllowed  bool   `json:"final_response_allowed"`
		RefusesFinalResponse  bool   `json:"refuses_final_response"`
		ExpectedNextNodeAfter string `json:"expected_next_node_after_completion"`
		ContinuationReason    string `json:"continuation_contract_reason"`
		EarlyReturnRiskStatus string `json:"early_return_risk_status"`
		RSIRemainsDenied      bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeEighteenDir, "compaction-resume-prompt.json"))
	fixture := mustLoadJSON[struct {
		Schema                           string   `json:"schema"`
		NodeID                           string   `json:"node_id"`
		Status                           string   `json:"status"`
		SourcePromptFixturePath          string   `json:"source_prompt_fixture_path"`
		SourcePromptMarkdownPath         string   `json:"source_prompt_markdown_path"`
		SourceReadbackPath               string   `json:"source_readback_path"`
		CompletedNodesBefore             int      `json:"completed_nodes_before"`
		TotalNodes                       int      `json:"total_nodes"`
		ReadyNodesBefore                 int      `json:"ready_nodes_before"`
		BlockedNodesBefore               int      `json:"blocked_nodes_before"`
		FailedNodesBefore                int      `json:"failed_nodes_before"`
		FirstExecutableNodeBefore        string   `json:"first_executable_node_before"`
		ExactNextActionBefore            string   `json:"exact_next_action_before"`
		ReturnGateStatusBefore           string   `json:"return_gate_status_before"`
		ContinuationContractReasonBefore string   `json:"continuation_contract_reason_before"`
		EarlyReturnRiskStatusBefore      string   `json:"early_return_risk_status_before"`
		FinalResponseAllowedBefore       bool     `json:"final_response_allowed_before"`
		RefusesFinalResponseBefore       bool     `json:"refuses_final_response_before"`
		RegressionAssertions             []string `json:"regression_assertions"`
		ExpectedNextNodeAfterComplete    string   `json:"expected_next_node_after_completion"`
		PromotionRequested               bool     `json:"promotion_requested"`
		PromotionGranted                 bool     `json:"promotion_granted"`
		RSIRemainsDenied                 bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeNineteenDir, "compaction-resume-regression.json"))
	promptBytes, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(fixture.SourcePromptMarkdownPath)))
	if err != nil {
		t.Fatalf("read compaction resume markdown: %v", err)
	}
	prompt := string(promptBytes)

	if nodeEighteenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeEighteenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-18" ||
		nodeEighteenLifecycle.Status != "merged_and_cleaned" ||
		nodeEighteenLifecycle.PRNumber != 321 ||
		nodeEighteenLifecycle.MergeCommit != "73928367bbc4c456564738e35d35e95fa04f7086" ||
		nodeEighteenLifecycle.CIStatus != "passed" ||
		!nodeEighteenLifecycle.LocalMainSynced ||
		!nodeEighteenLifecycle.LocalBranchDeleted ||
		!nodeEighteenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 18 lifecycle evidence must prove clean branch handoff before node 19 regression: %#v", nodeEighteenLifecycle)
	}
	if promptFixture.Schema != "ao.atlas.compaction-resume-prompt.v0.1" ||
		promptFixture.NodeID != "mission-recommendation-final-closure-consolidation-18" ||
		promptFixture.FirstExecutableNode != "mission-recommendation-final-closure-consolidation-18" ||
		promptFixture.ExactNextAction != "Emit Foundry import for mission-recommendation-final-closure-consolidation-18 and execute exactly one active node." ||
		promptFixture.ReturnGateStatus != readback.ReturnGateStatus ||
		promptFixture.ContinuationReason != readback.ContinuationContract.Reason ||
		promptFixture.EarlyReturnRiskStatus != readback.EarlyReturnRiskStatus ||
		promptFixture.FinalResponseAllowed ||
		!promptFixture.RefusesFinalResponse ||
		promptFixture.ExpectedNextNodeAfter != "mission-recommendation-final-closure-consolidation-19" ||
		!promptFixture.RSIRemainsDenied {
		t.Fatalf("source compaction resume prompt must preserve node 19 continuation fields: %#v", promptFixture)
	}
	if fixture.Schema != "ao.atlas.compaction-resume-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-19" ||
		fixture.Status != "guarded" ||
		fixture.SourcePromptFixturePath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-18/compaction-resume-prompt.json" ||
		fixture.SourcePromptMarkdownPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-18/compaction-resume-prompt.md" ||
		fixture.SourceReadbackPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-18/recommendation-readback-after.json" ||
		fixture.CompletedNodesBefore != readback.CompletedNodes ||
		fixture.TotalNodes != readback.TotalNodes ||
		fixture.ReadyNodesBefore != readback.ReadyNodes ||
		fixture.BlockedNodesBefore != readback.BlockedNodes ||
		fixture.FailedNodesBefore != readback.FailedNodes ||
		fixture.FirstExecutableNodeBefore != readback.FirstExecutableNode ||
		fixture.ExactNextActionBefore != readback.ExactNextAction ||
		fixture.ReturnGateStatusBefore != readback.ReturnGateStatus ||
		fixture.ContinuationContractReasonBefore != readback.ContinuationContract.Reason ||
		fixture.EarlyReturnRiskStatusBefore != readback.EarlyReturnRiskStatus ||
		fixture.FinalResponseAllowedBefore != readback.FinalResponseAllowed ||
		!fixture.RefusesFinalResponseBefore ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-20" ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 19 compaction resume regression must preserve exact next-node and return-gate state: %#v", fixture)
	}
	for _, want := range []string{
		"first_executable_node_preserved",
		"exact_next_action_preserved",
		"return_gate_status_preserved",
		"final_response_denied_until_ready_work_consumed",
		"rsi_denial_preserved",
	} {
		if !containsString(fixture.RegressionAssertions, want) {
			t.Fatalf("node 19 regression missing assertion %q: %#v", want, fixture.RegressionAssertions)
		}
	}
	for _, want := range []string{
		"Next executable node: `mission-recommendation-final-closure-consolidation-18`",
		"Final response allowed: `false`",
		"Emit Foundry import for mission-recommendation-final-closure-consolidation-18 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("node 19 regression source prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestFinalClosureConsolidationMissionDashboardBindsMultiRepoEvidence(t *testing.T) {
	root := repoRoot(t)
	consolidationRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	completedWaveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeNineteenDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-19")
	nodeTwentyDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-20")

	nodeNineteenLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeNineteenDir, "post-merge-lifecycle.json"))
	consolidationReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeNineteenDir, "recommendation-readback-after.json"))
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(completedWaveRoot, "nodes", "mission-recommendation-hardening-40", "recommendation-readback-after.json"))
	closureFixture := mustLoadJSON[struct {
		Schema                    string   `json:"schema"`
		NodeID                    string   `json:"node_id"`
		Status                    string   `json:"status"`
		ClosureArtifactPaths      []string `json:"closure_artifact_paths"`
		PromoterNoPromotionStatus string   `json:"promoter_no_promotion_status"`
		CommandReadbackStatus     string   `json:"command_readback_status"`
		FinalResponseAllowedAfter bool     `json:"final_response_allowed_after_node"`
		ReadyNodesAfter           int      `json:"ready_nodes_after_node"`
		BlockedNodesAfter         int      `json:"blocked_nodes_after_node"`
		ClaimsAuthorityAdvance    bool     `json:"claims_authority_advance"`
		RSIRemainsDenied          bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(completedWaveRoot, "nodes", "mission-recommendation-hardening-40", "final-closure-artifacts-fixture.json"))
	dashboard := mustLoadJSON[struct {
		Schema                        string `json:"schema"`
		NodeID                        string `json:"node_id"`
		Status                        string `json:"status"`
		DashboardMarkdownPath         string `json:"dashboard_markdown_path"`
		SourceFinalReadbackPath       string `json:"source_final_readback_path"`
		SourceClosureFixturePath      string `json:"source_closure_fixture_path"`
		SourceOperatorSummaryPath     string `json:"source_operator_summary_path"`
		SourcePromoterCommandRollup   string `json:"source_promoter_command_rollup_path"`
		SourceSchemaValidationReport  string `json:"source_schema_validation_report_path"`
		SourcePostMergeCleanupRollup  string `json:"source_post_merge_cleanup_rollup_path"`
		SourceConsolidationReadback   string `json:"source_consolidation_readback_path"`
		FinalWaveCompletedNodes       int    `json:"final_wave_completed_nodes"`
		FinalWaveReadyNodes           int    `json:"final_wave_ready_nodes"`
		FinalWaveBlockedNodes         int    `json:"final_wave_blocked_nodes"`
		FinalWaveFailedNodes          int    `json:"final_wave_failed_nodes"`
		FinalWaveFinalResponseAllowed bool   `json:"final_wave_final_response_allowed"`
		ConsolidationCompletedBefore  int    `json:"consolidation_completed_before"`
		ConsolidationReadyBefore      int    `json:"consolidation_ready_before"`
		ConsolidationFirstExecutable  string `json:"consolidation_first_executable_node_before"`
		ConsolidationFinalAllowed     bool   `json:"consolidation_final_response_allowed_before"`
		NextConsolidationNode         string `json:"next_consolidation_node"`
		ComponentCount                int    `json:"component_count"`
		AllSourcePathsExist           bool   `json:"all_source_paths_exist"`
		PublicSafetyScanStatus        string `json:"public_safety_scan_status"`
		PromoterNoPromotionStatus     string `json:"promoter_no_promotion_status"`
		CommandReadbackStatus         string `json:"command_readback_status"`
		PromotionRequested            bool   `json:"promotion_requested"`
		PromotionGranted              bool   `json:"promotion_granted"`
		ClaimsAuthorityAdvance        bool   `json:"claims_authority_advance"`
		RSIRemainsDenied              bool   `json:"rsi_remains_denied"`
		RepoBindings                  []struct {
			Repo         string `json:"repo"`
			Owner        string `json:"owner"`
			EvidenceRole string `json:"evidence_role"`
			SourcePath   string `json:"source_path"`
			Status       string `json:"status"`
		} `json:"repo_bindings"`
	}](t, filepath.Join(nodeTwentyDir, "mission-dashboard-binding.json"))
	markdownBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(dashboard.DashboardMarkdownPath)))
	if err != nil {
		t.Fatalf("read mission dashboard markdown: %v", err)
	}
	markdown := string(markdownBytes)

	if nodeNineteenLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeNineteenLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-19" ||
		nodeNineteenLifecycle.Status != "merged_and_cleaned" ||
		nodeNineteenLifecycle.PRNumber != 322 ||
		nodeNineteenLifecycle.MergeCommit != "ca45cca5ff434a3b92f39f029b2fcb78b1b543ff" ||
		nodeNineteenLifecycle.CIStatus != "passed" ||
		!nodeNineteenLifecycle.LocalMainSynced ||
		!nodeNineteenLifecycle.LocalBranchDeleted ||
		!nodeNineteenLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 19 lifecycle evidence must prove clean branch handoff before node 20 dashboard: %#v", nodeNineteenLifecycle)
	}
	if dashboard.Schema != "ao.atlas.mission-dashboard-binding.v0.1" ||
		dashboard.NodeID != "mission-recommendation-final-closure-consolidation-20" ||
		dashboard.Status != "bound" ||
		dashboard.DashboardMarkdownPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-20/mission-dashboard.md" ||
		dashboard.SourceFinalReadbackPath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/recommendation-readback-after.json" ||
		dashboard.SourceClosureFixturePath != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/final-closure-artifacts-fixture.json" ||
		dashboard.SourceOperatorSummaryPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-10/final-operator-summary.json" ||
		dashboard.SourcePromoterCommandRollup != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-06/aggregate-promoter-command-rollup.json" ||
		dashboard.SourceSchemaValidationReport != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-12/completed-wave-evidence-validation-report.json" ||
		dashboard.SourcePostMergeCleanupRollup != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-04/post-merge-cleanup-rollup.json" ||
		dashboard.SourceConsolidationReadback != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-19/recommendation-readback-after.json" ||
		dashboard.FinalWaveCompletedNodes != finalReadback.CompletedNodes ||
		dashboard.FinalWaveReadyNodes != finalReadback.ReadyNodes ||
		dashboard.FinalWaveBlockedNodes != finalReadback.BlockedNodes ||
		dashboard.FinalWaveFailedNodes != finalReadback.FailedNodes ||
		dashboard.FinalWaveFinalResponseAllowed != finalReadback.FinalResponseAllowed ||
		dashboard.ConsolidationCompletedBefore != consolidationReadback.CompletedNodes ||
		dashboard.ConsolidationReadyBefore != consolidationReadback.ReadyNodes ||
		dashboard.ConsolidationFirstExecutable != consolidationReadback.FirstExecutableNode ||
		dashboard.ConsolidationFinalAllowed != consolidationReadback.FinalResponseAllowed ||
		dashboard.NextConsolidationNode != "mission-recommendation-final-closure-consolidation-21" ||
		dashboard.ComponentCount != 6 ||
		len(dashboard.RepoBindings) != dashboard.ComponentCount ||
		!dashboard.AllSourcePathsExist ||
		dashboard.PublicSafetyScanStatus != finalReadback.PublicSafetyScanStatus ||
		dashboard.PromoterNoPromotionStatus != closureFixture.PromoterNoPromotionStatus ||
		dashboard.CommandReadbackStatus != closureFixture.CommandReadbackStatus ||
		dashboard.PromotionRequested ||
		dashboard.PromotionGranted ||
		dashboard.ClaimsAuthorityAdvance ||
		!dashboard.RSIRemainsDenied ||
		!closureFixture.FinalResponseAllowedAfter ||
		closureFixture.ReadyNodesAfter != 0 ||
		closureFixture.BlockedNodesAfter != 0 ||
		closureFixture.ClaimsAuthorityAdvance ||
		!closureFixture.RSIRemainsDenied {
		t.Fatalf("node 20 dashboard must bind final closure and current consolidation state without promotion: %#v", dashboard)
	}
	wantRepos := map[string]string{
		"ao-atlas":    "final_readback_and_closure",
		"ao-foundry":  "bounded_import_run_link_coverage",
		"ao-promoter": "no_promotion_rollup",
		"ao-command":  "class_decision_readback",
		"ao-sentinel": "public_safety_scan",
		"ao-mission":  "supervised_continuation_state",
	}
	for _, binding := range dashboard.RepoBindings {
		if wantRepos[binding.Repo] != binding.EvidenceRole {
			t.Fatalf("unexpected dashboard binding for repo %s: %#v", binding.Repo, binding)
		}
		if binding.SourcePath == "" || binding.Status == "" || binding.Owner == "" {
			t.Fatalf("dashboard binding must include owner, source path, and status: %#v", binding)
		}
	}
	for repo := range wantRepos {
		if !strings.Contains(markdown, repo) {
			t.Fatalf("mission dashboard markdown missing repo %q:\n%s", repo, markdown)
		}
	}
	for _, want := range []string{
		"Final wave: 40/40 completed",
		"Consolidation wave: 19/24 completed",
		"Next consolidation node: `mission-recommendation-final-closure-consolidation-21`",
		"Promoter: no_promotion_requested",
		"Command: readback_agrees_no_promotion",
		"RSI remains denied.",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("mission dashboard markdown missing %q:\n%s", want, markdown)
		}
	}
}

func TestFinalClosureConsolidationMissionDashboardRegressionCoversCoreBindings(t *testing.T) {
	root := repoRoot(t)
	consolidationRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTwentyDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-20")
	nodeTwentyOneDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-21")

	nodeTwentyLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeTwentyDir, "post-merge-lifecycle.json"))
	dashboard := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		DashboardMarkdown   string `json:"dashboard_markdown_path"`
		ComponentCount      int    `json:"component_count"`
		AllSourcePathsExist bool   `json:"all_source_paths_exist"`
		PromotionRequested  bool   `json:"promotion_requested"`
		PromotionGranted    bool   `json:"promotion_granted"`
		RSIRemainsDenied    bool   `json:"rsi_remains_denied"`
		RepoBindings        []struct {
			Repo         string `json:"repo"`
			Owner        string `json:"owner"`
			EvidenceRole string `json:"evidence_role"`
			SourcePath   string `json:"source_path"`
			Status       string `json:"status"`
		} `json:"repo_bindings"`
	}](t, filepath.Join(nodeTwentyDir, "mission-dashboard-binding.json"))
	fixture := mustLoadJSON[struct {
		Schema                        string            `json:"schema"`
		NodeID                        string            `json:"node_id"`
		Status                        string            `json:"status"`
		SourceDashboardPath           string            `json:"source_dashboard_path"`
		SourceDashboardMarkdownPath   string            `json:"source_dashboard_markdown_path"`
		SourceComponentCount          int               `json:"source_component_count"`
		RequiredBindingRepos          []string          `json:"required_binding_repos"`
		RequiredEvidenceRoles         map[string]string `json:"required_evidence_roles"`
		RequiredStatuses              map[string]string `json:"required_statuses"`
		RequiredBindingsPresent       bool              `json:"required_bindings_present"`
		RequiredBindingSourcePathsSet bool              `json:"required_binding_source_paths_set"`
		MissionContinuationRetained   bool              `json:"mission_continuation_retained"`
		AllSourcePathsExist           bool              `json:"all_source_paths_exist"`
		ExpectedNextNodeAfterComplete string            `json:"expected_next_node_after_completion"`
		PromotionRequested            bool              `json:"promotion_requested"`
		PromotionGranted              bool              `json:"promotion_granted"`
		RSIRemainsDenied              bool              `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeTwentyOneDir, "mission-dashboard-regression.json"))
	markdownBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(fixture.SourceDashboardMarkdownPath)))
	if err != nil {
		t.Fatalf("read mission dashboard markdown: %v", err)
	}
	markdown := string(markdownBytes)

	if nodeTwentyLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTwentyLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-20" ||
		nodeTwentyLifecycle.Status != "merged_and_cleaned" ||
		nodeTwentyLifecycle.PRNumber != 323 ||
		nodeTwentyLifecycle.MergeCommit != "c60d7fb535373c93191e1d5883eef9fca713b249" ||
		nodeTwentyLifecycle.CIStatus != "passed" ||
		!nodeTwentyLifecycle.LocalMainSynced ||
		!nodeTwentyLifecycle.LocalBranchDeleted ||
		!nodeTwentyLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 20 lifecycle evidence must prove clean branch handoff before node 21 regression: %#v", nodeTwentyLifecycle)
	}
	if fixture.Schema != "ao.atlas.mission-dashboard-regression.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-21" ||
		fixture.Status != "guarded" ||
		fixture.SourceDashboardPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-20/mission-dashboard-binding.json" ||
		fixture.SourceDashboardMarkdownPath != dashboard.DashboardMarkdown ||
		fixture.SourceComponentCount != dashboard.ComponentCount ||
		len(fixture.RequiredBindingRepos) != 5 ||
		!fixture.RequiredBindingsPresent ||
		!fixture.RequiredBindingSourcePathsSet ||
		!fixture.MissionContinuationRetained ||
		fixture.AllSourcePathsExist != dashboard.AllSourcePathsExist ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-22" ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		!fixture.RSIRemainsDenied ||
		dashboard.Schema != "ao.atlas.mission-dashboard-binding.v0.1" ||
		dashboard.NodeID != "mission-recommendation-final-closure-consolidation-20" ||
		dashboard.Status != "bound" ||
		dashboard.ComponentCount != 6 ||
		dashboard.PromotionRequested ||
		dashboard.PromotionGranted ||
		!dashboard.RSIRemainsDenied {
		t.Fatalf("node 21 dashboard regression must bind dashboard state without promotion: fixture=%#v dashboard=%#v", fixture, dashboard)
	}
	bindings := map[string]struct {
		role       string
		status     string
		sourcePath string
	}{}
	for _, binding := range dashboard.RepoBindings {
		bindings[binding.Repo] = struct {
			role       string
			status     string
			sourcePath string
		}{role: binding.EvidenceRole, status: binding.Status, sourcePath: binding.SourcePath}
	}
	for _, repo := range fixture.RequiredBindingRepos {
		binding, ok := bindings[repo]
		if !ok {
			t.Fatalf("dashboard missing required repo binding %q: %#v", repo, dashboard.RepoBindings)
		}
		if binding.role != fixture.RequiredEvidenceRoles[repo] || binding.status != fixture.RequiredStatuses[repo] || binding.sourcePath == "" {
			t.Fatalf("dashboard binding mismatch for %s: got %#v roles=%#v statuses=%#v", repo, binding, fixture.RequiredEvidenceRoles, fixture.RequiredStatuses)
		}
		if !strings.Contains(markdown, repo) || !strings.Contains(markdown, binding.status) {
			t.Fatalf("dashboard markdown must include repo %s and status %s:\n%s", repo, binding.status, markdown)
		}
	}
	if bindings["ao-mission"].status != "continuation_required" || !strings.Contains(markdown, "ao-mission") {
		t.Fatalf("dashboard regression must preserve Mission continuation context: %#v\n%s", bindings["ao-mission"], markdown)
	}
}

func TestFinalClosureConsolidationNoPromotionNoRSIAssertionCoversCompletedNodes(t *testing.T) {
	root := repoRoot(t)
	consolidationRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	hardeningRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-long-run-hardening-wave-v01")
	nodeTwentyOneDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-21")
	nodeTwentyTwoDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-22")

	nodeTwentyOneLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeTwentyOneDir, "post-merge-lifecycle.json"))
	consolidationWorkgraph := mustLoadJSON[Workgraph](t, filepath.Join(nodeTwentyOneDir, "workgraph-after.json"))
	hardeningWorkgraph := mustLoadJSON[Workgraph](t, filepath.Join(hardeningRoot, "nodes", "mission-recommendation-hardening-40", "workgraph-after.json"))
	fixture := mustLoadJSON[struct {
		Schema                        string   `json:"schema"`
		NodeID                        string   `json:"node_id"`
		Status                        string   `json:"status"`
		SourceConsolidationWorkgraph  string   `json:"source_consolidation_workgraph_path"`
		SourceHardeningWorkgraph      string   `json:"source_hardening_workgraph_path"`
		CompletedHardeningNodes       int      `json:"completed_hardening_nodes"`
		CompletedConsolidationBefore  int      `json:"completed_consolidation_nodes_before"`
		CoveredCompletedNodesTotal    int      `json:"covered_completed_nodes_total"`
		PromoterNoPromotionFiles      int      `json:"promoter_no_promotion_files"`
		CommandReadbackFiles          int      `json:"command_readback_files"`
		SentinelPublicSafetyFiles     int      `json:"sentinel_public_safety_files"`
		PromotionRequestedFalseCount  int      `json:"promotion_requested_false_count"`
		PromotionGrantedFalseCount    int      `json:"promotion_granted_false_count"`
		SentinelRSIDeniedCount        int      `json:"sentinel_rsi_denied_count"`
		AllowedPromoterStatuses       []string `json:"allowed_promoter_statuses"`
		AllowedCommandStatusPrefixes  []string `json:"allowed_command_status_prefixes"`
		NoPromotionInvariantHolds     bool     `json:"no_promotion_invariant_holds"`
		RSIDenialInvariantHolds       bool     `json:"rsi_denial_invariant_holds"`
		ExpectedNextNodeAfterComplete string   `json:"expected_next_node_after_completion"`
		PromotionRequested            bool     `json:"promotion_requested"`
		PromotionGranted              bool     `json:"promotion_granted"`
		ClaimsAuthorityAdvance        bool     `json:"claims_authority_advance"`
		RSIRemainsDenied              bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeTwentyTwoDir, "no-promotion-no-rsi-assertion.json"))

	completedNodeIDs := func(workgraph Workgraph) map[string]bool {
		completed := map[string]bool{}
		for _, node := range workgraph.Nodes {
			if node.Status == "completed" {
				completed[node.ID] = true
			}
		}
		return completed
	}
	countNamedFiles := func(rootDir, name string, coveredNodes map[string]bool) int {
		t.Helper()
		count := 0
		if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == name {
				rel, relErr := filepath.Rel(rootDir, path)
				if relErr != nil {
					return relErr
				}
				nodeID := strings.Split(filepath.ToSlash(rel), "/")[0]
				if coveredNodes[nodeID] {
					count++
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("walk %s: %v", rootDir, err)
		}
		return count
	}
	countCompleted := func(workgraph Workgraph) int {
		count := 0
		for _, node := range workgraph.Nodes {
			if node.Status == "completed" {
				count++
			}
		}
		return count
	}

	completedHardening := countCompleted(hardeningWorkgraph)
	completedConsolidation := countCompleted(consolidationWorkgraph)
	completedHardeningIDs := completedNodeIDs(hardeningWorkgraph)
	completedConsolidationIDs := completedNodeIDs(consolidationWorkgraph)
	promoterFiles := countNamedFiles(filepath.Join(hardeningRoot, "nodes"), "promoter_no_promotion.json", completedHardeningIDs) + countNamedFiles(filepath.Join(consolidationRoot, "nodes"), "promoter_no_promotion.json", completedConsolidationIDs)
	commandFiles := countNamedFiles(filepath.Join(hardeningRoot, "nodes"), "command_readback.json", completedHardeningIDs) + countNamedFiles(filepath.Join(consolidationRoot, "nodes"), "command_readback.json", completedConsolidationIDs)
	sentinelFiles := countNamedFiles(filepath.Join(hardeningRoot, "nodes"), "sentinel_public_safety.json", completedHardeningIDs) + countNamedFiles(filepath.Join(consolidationRoot, "nodes"), "sentinel_public_safety.json", completedConsolidationIDs)

	if nodeTwentyOneLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTwentyOneLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-21" ||
		nodeTwentyOneLifecycle.Status != "merged_and_cleaned" ||
		nodeTwentyOneLifecycle.PRNumber != 324 ||
		nodeTwentyOneLifecycle.MergeCommit != "b7984846c9faeab349b8546646c10a9e050f2951" ||
		nodeTwentyOneLifecycle.CIStatus != "passed" ||
		!nodeTwentyOneLifecycle.LocalMainSynced ||
		!nodeTwentyOneLifecycle.LocalBranchDeleted ||
		!nodeTwentyOneLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 21 lifecycle evidence must prove clean branch handoff before node 22 assertion: %#v", nodeTwentyOneLifecycle)
	}
	if fixture.Schema != "ao.atlas.no-promotion-no-rsi-assertion.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-22" ||
		fixture.Status != "asserted" ||
		fixture.SourceConsolidationWorkgraph != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-21/workgraph-after.json" ||
		fixture.SourceHardeningWorkgraph != "docs/evidence/ao-atlas-long-run-hardening-wave-v01/nodes/mission-recommendation-hardening-40/workgraph-after.json" ||
		fixture.CompletedHardeningNodes != completedHardening ||
		fixture.CompletedConsolidationBefore != completedConsolidation ||
		fixture.CoveredCompletedNodesTotal != completedHardening+completedConsolidation ||
		fixture.PromoterNoPromotionFiles != promoterFiles ||
		fixture.CommandReadbackFiles != commandFiles ||
		fixture.SentinelPublicSafetyFiles != sentinelFiles ||
		fixture.PromotionRequestedFalseCount != promoterFiles ||
		fixture.PromotionGrantedFalseCount != promoterFiles ||
		fixture.SentinelRSIDeniedCount != sentinelFiles ||
		!containsString(fixture.AllowedPromoterStatuses, "no_promotion_requested") ||
		!containsString(fixture.AllowedPromoterStatuses, "no_promotion") ||
		!containsString(fixture.AllowedPromoterStatuses, "recorded") ||
		!containsString(fixture.AllowedCommandStatusPrefixes, "readback_agrees_no_promotion") ||
		!containsString(fixture.AllowedCommandStatusPrefixes, "recorded") ||
		!fixture.NoPromotionInvariantHolds ||
		!fixture.RSIDenialInvariantHolds ||
		fixture.ExpectedNextNodeAfterComplete != "mission-recommendation-final-closure-consolidation-23" ||
		fixture.PromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 22 no-promotion/no-RSI assertion must cover completed node evidence: %#v", fixture)
	}
}

func TestFinalClosureConsolidationNextWaveExporterRanksFortyFeatureDepthTasks(t *testing.T) {
	root := repoRoot(t)
	consolidationRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTwentyTwoDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-22")
	nodeTwentyThreeDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-23")
	sourceEvidenceRoot := "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01"
	sourceReadback := filepath.ToSlash(filepath.Join(sourceEvidenceRoot, "nodes", "mission-recommendation-final-closure-consolidation-22", "recommendation-readback-after.json"))
	sourceAssertion := filepath.ToSlash(filepath.Join(sourceEvidenceRoot, "nodes", "mission-recommendation-final-closure-consolidation-22", "no-promotion-no-rsi-assertion.json"))

	type rankedTask struct {
		Rank         int      `json:"rank"`
		ID           string   `json:"id"`
		Owner        string   `json:"owner"`
		Theme        string   `json:"theme"`
		Task         string   `json:"task"`
		EvidenceRefs []string `json:"evidence_refs"`
	}
	type featureDepthExport struct {
		Schema              string       `json:"schema"`
		MissionID           string       `json:"mission_id"`
		Status              string       `json:"status"`
		MinimumTasks        int          `json:"minimum_tasks"`
		RecommendationCount int          `json:"recommendation_count"`
		SourceEvidenceRoot  string       `json:"source_evidence_root"`
		SourceReadbackPath  string       `json:"source_readback_path"`
		SourceAssertionPath string       `json:"source_assertion_path"`
		Tasks               []rankedTask `json:"tasks"`
		SafeToExecute       bool         `json:"safe_to_execute"`
		SchedulesWork       bool         `json:"schedules_work"`
		ExecutesWork        bool         `json:"executes_work"`
		ApprovesWork        bool         `json:"approves_work"`
		MutatesRepositories bool         `json:"mutates_repositories"`
	}
	type exporterFixture struct {
		Schema                 string             `json:"schema"`
		NodeID                 string             `json:"node_id"`
		Status                 string             `json:"status"`
		SourceEvidenceRoot     string             `json:"source_evidence_root"`
		SourceReadbackPath     string             `json:"source_readback_path"`
		SourceAssertionPath    string             `json:"source_assertion_path"`
		CompletedNodesBefore   int                `json:"completed_nodes_before_export"`
		ReadyNodesBefore       int                `json:"ready_nodes_before_export"`
		ExpectedNextNode       string             `json:"expected_next_node_after_completion"`
		MinimumRankedTasks     int                `json:"minimum_ranked_tasks"`
		RecommendationCount    int                `json:"recommendation_count"`
		RankedTaskFloorMet     bool               `json:"ranked_task_floor_met"`
		NoPromotionRequested   bool               `json:"no_promotion_requested"`
		PromotionGranted       bool               `json:"promotion_granted"`
		ClaimsAuthorityAdvance bool               `json:"claims_authority_advance"`
		RSIRemainsDenied       bool               `json:"rsi_remains_denied"`
		FeatureDepthExport     featureDepthExport `json:"feature_depth_export"`
	}
	assertExport := func(t *testing.T, export featureDepthExport) {
		t.Helper()
		if export.Schema != "ao.mission.feature-depth-recommendations.v0.3" ||
			export.MissionID != "ao-atlas-next-feature-depth-wave-v01" ||
			export.Status != "ready" ||
			export.MinimumTasks != 40 ||
			export.RecommendationCount != 40 ||
			len(export.Tasks) != 40 ||
			export.SourceEvidenceRoot != sourceEvidenceRoot ||
			export.SourceReadbackPath != sourceReadback ||
			export.SourceAssertionPath != sourceAssertion ||
			export.SafeToExecute ||
			export.SchedulesWork ||
			export.ExecutesWork ||
			export.ApprovesWork ||
			export.MutatesRepositories {
			t.Fatalf("next-wave export has invalid contract or unsafe authority flags: %#v", export)
		}
		seenThemes := map[string]bool{}
		for i, task := range export.Tasks {
			wantRank := i + 1
			if task.Rank != wantRank ||
				task.ID != "feature-depth-next-wave-"+twoDigit(wantRank) ||
				task.Owner != "ao-atlas" ||
				task.Theme == "" ||
				len(strings.Fields(task.Task)) < 6 ||
				len(task.EvidenceRefs) == 0 ||
				!containsString(task.EvidenceRefs, sourceReadback) {
				t.Fatalf("task %d is not a ranked, evidence-bound AO Atlas Feature Depth task: %#v", wantRank, task)
			}
			seenThemes[task.Theme] = true
		}
		if len(seenThemes) < 10 {
			t.Fatalf("next-wave export must span at least 10 themes, got %d: %#v", len(seenThemes), seenThemes)
		}
	}

	sourceReadbackValue := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, sourceReadback))
	sourceAssertionValue := mustLoadJSON[map[string]any](t, filepath.Join(nodeTwentyTwoDir, "no-promotion-no-rsi-assertion.json"))
	if sourceReadbackValue.CompletedNodes != 22 ||
		sourceReadbackValue.ReadyNodes != 2 ||
		sourceReadbackValue.FirstExecutableNode != "mission-recommendation-final-closure-consolidation-23" ||
		sourceReadbackValue.FinalResponseAllowed {
		t.Fatalf("node 23 exporter must bind the node 22 continuation checkpoint: %#v", sourceReadbackValue)
	}
	if sourceAssertionValue["rsi_remains_denied"] != true || sourceAssertionValue["promotion_granted"] != false {
		t.Fatalf("node 23 exporter must bind no-promotion/no-RSI assertion: %#v", sourceAssertionValue)
	}

	tmpExportPath := filepath.Join(t.TempDir(), "next-wave-feature-depth-recommendations.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "export-next-wave",
		"--mission-id", "ao-atlas-next-feature-depth-wave-v01",
		"--source-evidence-root", sourceEvidenceRoot,
		"--source-readback", sourceReadback,
		"--source-assertion", sourceAssertion,
		"--min-tasks", "40",
		"--out", tmpExportPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("export-next-wave failed: %s", out.String())
	}
	generatedExport := mustLoadJSON[featureDepthExport](t, tmpExportPath)
	assertExport(t, generatedExport)

	fixture := mustLoadJSON[exporterFixture](t, filepath.Join(nodeTwentyThreeDir, "next-wave-recommendation-export.json"))
	if fixture.Schema != "ao.atlas.next-wave-recommendation-export.v0.1" ||
		fixture.NodeID != "mission-recommendation-final-closure-consolidation-23" ||
		fixture.Status != "exported" ||
		fixture.SourceEvidenceRoot != sourceEvidenceRoot ||
		fixture.SourceReadbackPath != sourceReadback ||
		fixture.SourceAssertionPath != sourceAssertion ||
		fixture.CompletedNodesBefore != sourceReadbackValue.CompletedNodes ||
		fixture.ReadyNodesBefore != sourceReadbackValue.ReadyNodes ||
		fixture.ExpectedNextNode != "mission-recommendation-final-closure-consolidation-24" ||
		fixture.MinimumRankedTasks != 40 ||
		fixture.RecommendationCount != 40 ||
		!fixture.RankedTaskFloorMet ||
		!fixture.NoPromotionRequested ||
		fixture.PromotionGranted ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("node 23 exporter fixture does not prove a safe 40-task next wave: %#v", fixture)
	}
	assertExport(t, fixture.FeatureDepthExport)
}

func TestFinalClosureConsolidationFinalReadbackClosesAllNodes(t *testing.T) {
	root := repoRoot(t)
	consolidationRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-final-closure-consolidation-wave-v01")
	nodeTwentyThreeDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-23")
	nodeTwentyFourDir := filepath.Join(consolidationRoot, "nodes", "mission-recommendation-final-closure-consolidation-24")

	nodeTwentyThreeLifecycle := mustLoadJSON[struct {
		Schema              string `json:"schema"`
		NodeID              string `json:"node_id"`
		Status              string `json:"status"`
		PRNumber            int    `json:"pr_number"`
		MergeCommit         string `json:"merge_commit"`
		CIStatus            string `json:"ci_status"`
		LocalMainSynced     bool   `json:"local_main_synced"`
		LocalBranchDeleted  bool   `json:"local_branch_deleted"`
		RemoteBranchDeleted bool   `json:"remote_branch_deleted"`
	}](t, filepath.Join(nodeTwentyThreeDir, "post-merge-lifecycle.json"))
	exporter := mustLoadJSON[AtlasNextWaveRecommendationExport](t, filepath.Join(nodeTwentyThreeDir, "next-wave-recommendation-export.json"))
	exporterRegression := mustLoadJSON[struct {
		Schema                 string `json:"schema"`
		NodeID                 string `json:"node_id"`
		Status                 string `json:"status"`
		SourceExporterPath     string `json:"source_exporter_path"`
		RecommendationCount    int    `json:"recommendation_count"`
		FirstRank              int    `json:"first_rank"`
		LastRank               int    `json:"last_rank"`
		RankedTaskFloorMet     bool   `json:"ranked_task_floor_met"`
		ExpectedFinalCompleted int    `json:"expected_final_completed_nodes"`
		ExpectedFinalReady     int    `json:"expected_final_ready_nodes"`
		ExpectedFinalResponse  bool   `json:"expected_final_response_allowed"`
		PromotionGranted       bool   `json:"promotion_granted"`
		ClaimsAuthorityAdvance bool   `json:"claims_authority_advance"`
		RSIRemainsDenied       bool   `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeTwentyFourDir, "exporter-regression-evidence.json"))
	finalClosure := mustLoadJSON[struct {
		Schema                  string   `json:"schema"`
		NodeID                  string   `json:"node_id"`
		Status                  string   `json:"status"`
		CompletedNodes          int      `json:"completed_nodes"`
		TotalNodes              int      `json:"total_nodes"`
		ReadyNodes              int      `json:"ready_nodes"`
		BlockedNodes            int      `json:"blocked_nodes"`
		FailedNodes             int      `json:"failed_nodes"`
		FinalResponseAllowed    bool     `json:"final_response_allowed"`
		ReturnGateStatus        string   `json:"return_gate_status"`
		PromoterStatus          string   `json:"promoter_status"`
		CommandStatus           string   `json:"command_status"`
		FoundryRollupStatus     string   `json:"foundry_rollup_status"`
		NextWaveRecommendation  string   `json:"next_wave_recommendation_path"`
		FeatureDepthTaskCount   int      `json:"feature_depth_task_count"`
		FeatureDepthSampleTasks []string `json:"feature_depth_sample_tasks"`
		PromotionGranted        bool     `json:"promotion_granted"`
		ClaimsAuthorityAdvance  bool     `json:"claims_authority_advance"`
		RSIRemainsDenied        bool     `json:"rsi_remains_denied"`
	}](t, filepath.Join(nodeTwentyFourDir, "final-consolidation-closure-readback.json"))
	finalReadback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(nodeTwentyFourDir, "recommendation-readback-after.json"))
	command := mustLoadJSON[AtlasRecommendationCommandReadback](t, filepath.Join(nodeTwentyFourDir, "final-command-readback.json"))
	promoter := mustLoadJSON[AtlasRecommendationPromoterReadback](t, filepath.Join(nodeTwentyFourDir, "final-promoter-readback.json"))
	foundry := mustLoadJSON[AtlasRecommendationFoundryRollup](t, filepath.Join(nodeTwentyFourDir, "final-foundry-rollup.json"))

	if nodeTwentyThreeLifecycle.Schema != "ao.atlas.post-merge-lifecycle.v0.1" ||
		nodeTwentyThreeLifecycle.NodeID != "mission-recommendation-final-closure-consolidation-23" ||
		nodeTwentyThreeLifecycle.Status != "merged_and_cleaned" ||
		nodeTwentyThreeLifecycle.PRNumber != 326 ||
		nodeTwentyThreeLifecycle.MergeCommit != "28f4fb009d32e01529f8a009014c40bd5f4f2229" ||
		nodeTwentyThreeLifecycle.CIStatus != "passed" ||
		!nodeTwentyThreeLifecycle.LocalMainSynced ||
		!nodeTwentyThreeLifecycle.LocalBranchDeleted ||
		!nodeTwentyThreeLifecycle.RemoteBranchDeleted {
		t.Fatalf("node 23 lifecycle evidence must prove clean branch handoff before final closure: %#v", nodeTwentyThreeLifecycle)
	}
	if exporter.Status != "exported" ||
		exporter.RecommendationCount != 40 ||
		!exporter.RankedTaskFloorMet ||
		len(exporter.FeatureDepthExport.Tasks) != 40 {
		t.Fatalf("node 24 final closure must bind the node 23 exporter: %#v", exporter)
	}
	if exporterRegression.Schema != "ao.atlas.exporter-regression-evidence.v0.1" ||
		exporterRegression.NodeID != "mission-recommendation-final-closure-consolidation-24" ||
		exporterRegression.Status != "passed" ||
		exporterRegression.SourceExporterPath != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-23/next-wave-recommendation-export.json" ||
		exporterRegression.RecommendationCount != 40 ||
		exporterRegression.FirstRank != 1 ||
		exporterRegression.LastRank != 40 ||
		!exporterRegression.RankedTaskFloorMet ||
		exporterRegression.ExpectedFinalCompleted != 24 ||
		exporterRegression.ExpectedFinalReady != 0 ||
		!exporterRegression.ExpectedFinalResponse ||
		exporterRegression.PromotionGranted ||
		exporterRegression.ClaimsAuthorityAdvance ||
		!exporterRegression.RSIRemainsDenied {
		t.Fatalf("node 24 exporter regression evidence must prove ranked exporter closure: %#v", exporterRegression)
	}
	if err := ValidateAtlasRecommendationReadback(finalReadback); err != nil {
		t.Fatal(err)
	}
	if finalReadback.Status != "completed" ||
		finalReadback.CompletedNodes != 24 ||
		finalReadback.TotalNodes != 24 ||
		finalReadback.ReadyNodes != 0 ||
		finalReadback.BlockedNodes != 0 ||
		finalReadback.FailedNodes != 0 ||
		finalReadback.FirstExecutableNode != "" ||
		!finalReadback.FinalResponseAllowed ||
		finalReadback.ReturnGateStatus != "final_response_allowed" ||
		finalReadback.ExactNextAction != "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks." ||
		finalReadback.ContinuationContract.Status != "ready_for_final_response" ||
		finalReadback.ContinuationContract.RefusesFinalResponse {
		t.Fatalf("final node 24 readback must close all nodes: %#v", finalReadback)
	}
	if command.Status != "completed" ||
		command.NodeCompletionStatus != "all_nodes_complete" ||
		!command.FinalResponseAllowed ||
		command.ReturnGateStatus != "final_response_allowed" ||
		command.ReadyNodes != 0 {
		t.Fatalf("final Command readback must agree with closure: %#v", command)
	}
	if promoter.Status != "no_promotion" ||
		promoter.PromotionClaimed ||
		promoter.ClaimsAuthorityAdvance ||
		!promoter.RSIRemainsDenied ||
		!promoter.FinalResponseAllowed {
		t.Fatalf("final Promoter readback must preserve no-promotion/no-RSI boundary: %#v", promoter)
	}
	if foundry.Status != "completed" ||
		foundry.NodeCompletionStatus != "all_nodes_complete" ||
		!foundry.FinalResponseAllowed ||
		foundry.ReadyNodes != 0 {
		t.Fatalf("final Foundry rollup must report completed closure: %#v", foundry)
	}
	if finalClosure.Schema != "ao.atlas.final-consolidation-closure.v0.1" ||
		finalClosure.NodeID != "mission-recommendation-final-closure-consolidation-24" ||
		finalClosure.Status != "closed" ||
		finalClosure.CompletedNodes != finalReadback.CompletedNodes ||
		finalClosure.TotalNodes != finalReadback.TotalNodes ||
		finalClosure.ReadyNodes != 0 ||
		finalClosure.BlockedNodes != 0 ||
		finalClosure.FailedNodes != 0 ||
		!finalClosure.FinalResponseAllowed ||
		finalClosure.ReturnGateStatus != finalReadback.ReturnGateStatus ||
		finalClosure.PromoterStatus != promoter.Status ||
		finalClosure.CommandStatus != command.Status ||
		finalClosure.FoundryRollupStatus != foundry.Status ||
		finalClosure.NextWaveRecommendation != "docs/evidence/ao-atlas-final-closure-consolidation-wave-v01/nodes/mission-recommendation-final-closure-consolidation-23/next-wave-feature-depth-recommendations.json" ||
		finalClosure.FeatureDepthTaskCount != 40 ||
		len(finalClosure.FeatureDepthSampleTasks) < 10 ||
		finalClosure.PromotionGranted ||
		finalClosure.ClaimsAuthorityAdvance ||
		!finalClosure.RSIRemainsDenied {
		t.Fatalf("final closure artifact must bind readback, rollups, and next-wave recommendations: %#v", finalClosure)
	}
}

func TestProductionReadinessRejectsUnsafeRecommendationPromptContinuationReasonFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "examples", "invalid", "recommendation-prompt-unsafe-continuation-reason.md")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read unsafe continuation reason prompt fixture: %v", err)
	}
	if !strings.Contains(string(fixture), "Continuation contract reason: `fully_unsupervised_complex_mutation is proven`") {
		t.Fatalf("unsafe prompt fixture must poison the continuation reason line:\n%s", string(fixture))
	}

	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"examples/invalid/recommendation-prompt-unsafe-continuation-reason.md",
		"unsafe_recommendation_reason_scan",
		"unsafe generated recommendation prompt continuation reason was accepted",
		"generated recommendation prompt contains unsafe wording",
		"generated-recommendation-prompt-continuation-reason-negative-scan",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing unsafe recommendation prompt fixture coverage %q", want)
		}
	}
}
