package atlas

import (
	"fmt"
	"strings"
)

func ValidateAtlasMonth6LaunchReadinessWorkgraph(fixture AtlasMonth6LaunchReadinessWorkgraph) error {
	var errs []string
	requireContract(&errs, "month6_launch_readiness_workgraph", fixture.Schema, AtlasMonth6LaunchReadinessWorkgraphContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "launch_readiness_workgraph_bound" {
		errs = append(errs, "status must be launch_readiness_workgraph_bound")
	}
	if fixture.SourceRecommendationRank != 40 {
		errs = append(errs, "source_recommendation_rank must be 40")
	}
	if strings.TrimSpace(fixture.SourceRecommendationTask) != "Generate Month 6 launch readiness workgraph" {
		errs = append(errs, "source_recommendation_task must match Month 6 recommendation 40")
	}
	if fixture.SafetyGate != "planning_only_no_provider_no_release" {
		errs = append(errs, "safety_gate must be planning_only_no_provider_no_release")
	}
	if fixture.CompletedRecommendationCount != 40 || len(fixture.Nodes) != 40 {
		errs = append(errs, "completed_recommendation_count must equal forty nodes")
	}
	if fixture.ReadyNodes != 0 {
		errs = append(errs, "ready_nodes must be zero")
	}
	if fixture.BlockedNodes != 0 {
		errs = append(errs, "blocked_nodes must be zero")
	}
	if fixture.FailedNodes != 0 {
		errs = append(errs, "failed_nodes must be zero")
	}
	if !fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be true for terminal launch readiness")
	}
	for rank := 1; rank <= 40; rank++ {
		if !fixture.CompletedRecommendationsBound[rank] {
			errs = append(errs, fmt.Sprintf("completed_recommendations_bound[%d] must be true", rank))
		}
	}
	if !fixture.AllRecommendationsCompleted {
		errs = append(errs, "all_recommendations_completed must be true")
	}
	if !fixture.AllNodesHavePRCIMergeEvidence {
		errs = append(errs, "all_nodes_have_pr_ci_merge_evidence must be true")
	}
	if !fixture.AllNodesHaveEvidencePath {
		errs = append(errs, "all_nodes_have_evidence_path must be true")
	}
	if !fixture.AllNodesHaveOperatorReadback {
		errs = append(errs, "all_nodes_have_operator_readback must be true")
	}
	if !fixture.AllNodesHavePublicSafety {
		errs = append(errs, "all_nodes_have_public_safety must be true")
	}
	if !fixture.FixtureOnly {
		errs = append(errs, "fixture_only must be true")
	}
	if !fixture.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if fixture.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false for launch-readiness fixture")
	}
	if fixture.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if fixture.ProviderCallsAllowed {
		errs = append(errs, "provider_calls_allowed must be false")
	}
	if fixture.CredentialUseAllowed {
		errs = append(errs, "credential_use_allowed must be false")
	}
	if fixture.LiveMutationAllowed {
		errs = append(errs, "live_mutation_allowed must be false")
	}
	if fixture.ReleaseOrPublishAllowed {
		errs = append(errs, "release_or_publish_allowed must be false")
	}
	if fixture.ApprovalGranted {
		errs = append(errs, "approval_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)

	seen := map[int]bool{}
	for i, node := range fixture.Nodes {
		validateAtlasMonth6LaunchReadinessNode(&errs, fmt.Sprintf("nodes[%d]", i), node, seen)
	}
	for rank := 1; rank <= 40; rank++ {
		if !seen[rank] {
			errs = append(errs, fmt.Sprintf("nodes missing rank %d", rank))
		}
	}
	return joinErrors(errs)
}

func validateAtlasMonth6LaunchReadinessNode(errs *[]string, prefix string, node AtlasMonth6LaunchReadinessNode, seen map[int]bool) {
	if node.Rank < 1 || node.Rank > 40 {
		*errs = append(*errs, prefix+".rank must be between 1 and 40")
	}
	if seen[node.Rank] {
		*errs = append(*errs, prefix+".rank must be unique")
	}
	seen[node.Rank] = true
	if node.NodeID != month6LaunchReadinessNodeID(node.Rank) {
		*errs = append(*errs, prefix+".node_id must match month6 launch readiness rank")
	}
	requireField(errs, prefix+".repository", node.Repository)
	checkPublicPath(errs, prefix+".repository", node.Repository, true)
	requireField(errs, prefix+".task", node.Task)
	requireField(errs, prefix+".category", node.Category)
	if !oneOf(node.Category, "beta_operability", "evidence_ui", "golden_path", "contract_registry") {
		*errs = append(*errs, prefix+".category must be a Month 6 recommendation category")
	}
	if node.Status != "completed" {
		*errs = append(*errs, prefix+".status must be completed")
	}
	requireField(errs, prefix+".pr", node.PR)
	if !strings.HasPrefix(node.PR, "https://github.com/uesugitorachiyo/") {
		*errs = append(*errs, prefix+".pr must be a GitHub PR URL under uesugitorachiyo")
	}
	if !lowerHex40(node.MergeCommit) || node.MergeCommit == "0000000000000000000000000000000000000000" {
		*errs = append(*errs, prefix+".merge_commit must be 40 lowercase hex characters")
	}
	if node.CIStatus != "passed" {
		*errs = append(*errs, prefix+".ci_status must be passed")
	}
	if node.MergeStatus != "merged" {
		*errs = append(*errs, prefix+".merge_status must be merged")
	}
	requireField(errs, prefix+".evidence_path", node.EvidencePath)
	checkPublicPath(errs, prefix+".evidence_path", node.EvidencePath, true)
	if node.EvidenceStatus != "available" {
		*errs = append(*errs, prefix+".evidence_status must be available")
	}
	if node.OperatorReadbackStatus != "ready" {
		*errs = append(*errs, prefix+".operator_readback_status must be ready")
	}
	if node.PublicSafetyStatus != "passed" {
		*errs = append(*errs, prefix+".public_safety_status must be passed")
	}
	if node.PromotionStatus != "no_promotion_requested" {
		*errs = append(*errs, prefix+".promotion_status must be no_promotion_requested")
	}
	if !node.RSIRemainsDenied {
		*errs = append(*errs, prefix+".rsi_remains_denied must be true")
	}
	if node.SafeToExecute {
		*errs = append(*errs, prefix+".safe_to_execute must be false")
	}
	if node.ExecutesWork {
		*errs = append(*errs, prefix+".executes_work must be false")
	}
	if node.MutatesRepository {
		*errs = append(*errs, prefix+".mutates_repository must be false")
	}
}

func month6LaunchReadinessNodeID(rank int) string {
	return fmt.Sprintf("month6-recommendation-%02d", rank)
}
