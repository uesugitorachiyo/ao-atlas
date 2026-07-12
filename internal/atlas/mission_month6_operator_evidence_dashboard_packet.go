package atlas

import (
	"fmt"
	"strings"
)

func ValidateAtlasMonth6OperatorEvidenceDashboardPacket(fixture AtlasMonth6OperatorEvidenceDashboardPacket) error {
	var errs []string
	requireContract(&errs, "month6_operator_evidence_dashboard_packet", fixture.Schema, AtlasMonth6OperatorEvidenceDashboardPacketContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "operator_evidence_dashboard_packet_bound" {
		errs = append(errs, "status must be operator_evidence_dashboard_packet_bound")
	}
	if fixture.SourceRecommendationRank != 6 {
		errs = append(errs, "source_recommendation_rank must be 6")
	}
	if strings.TrimSpace(fixture.SourceRecommendationTask) != "Generate operator evidence verification dashboard packet" {
		errs = append(errs, "source_recommendation_task must match Month 6 recommendation 6")
	}
	if fixture.SafetyGate != "planning_only_no_provider_no_release" {
		errs = append(errs, "safety_gate must be planning_only_no_provider_no_release")
	}
	if fixture.DashboardScope != "month6_completed_recommendations_01_05" {
		errs = append(errs, "dashboard_scope must bind completed recommendations 01 through 05")
	}
	if fixture.CompletedRecommendationCount != 5 || len(fixture.CompletedRecommendations) != 5 {
		errs = append(errs, "completed_recommendation_count must equal five completed recommendations")
	}
	if fixture.DashboardRowCount != 5 || len(fixture.DashboardRows) != 5 {
		errs = append(errs, "dashboard_row_count must equal five dashboard rows")
	}
	for rank := 1; rank <= 5; rank++ {
		if !fixture.CompletedRecommendationsBound[rank] {
			errs = append(errs, fmt.Sprintf("completed_recommendations_bound[%d] must be true", rank))
		}
	}
	if !fixture.AllCompletedRecommendationsBound {
		errs = append(errs, "all_completed_recommendations_bound must be true")
	}
	if !fixture.AllDashboardRowsHaveMergeEvidence {
		errs = append(errs, "all_dashboard_rows_have_merge_evidence must be true")
	}
	if !fixture.AllDashboardRowsHaveEvidencePath {
		errs = append(errs, "all_dashboard_rows_have_evidence_path must be true")
	}
	if !fixture.AllDashboardRowsHaveOperatorReadback {
		errs = append(errs, "all_dashboard_rows_have_operator_readback must be true")
	}
	if !fixture.FixtureOnly {
		errs = append(errs, "fixture_only must be true")
	}
	if fixture.ProviderCallsAllowed {
		errs = append(errs, "provider_calls_allowed must be false")
	}
	if fixture.CredentialUseAllowed {
		errs = append(errs, "credential_use_allowed must be false")
	}
	if fixture.ReleaseOrPublishAllowed {
		errs = append(errs, "release_or_publish_allowed must be false")
	}
	if fixture.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if fixture.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false for fixture-only dashboard packets")
	}
	validateNoAuthorityEffects(&errs, false, false, false, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)

	completedRanks := map[int]bool{}
	for i, row := range fixture.CompletedRecommendations {
		validateMonth6OperatorEvidenceDashboardRow(&errs, fmt.Sprintf("completed_recommendations[%d]", i), row)
		completedRanks[row.Rank] = true
	}
	dashboardRanks := map[int]bool{}
	for i, row := range fixture.DashboardRows {
		validateMonth6OperatorEvidenceDashboardRow(&errs, fmt.Sprintf("dashboard_rows[%d]", i), row)
		dashboardRanks[row.Rank] = true
	}
	for rank := 1; rank <= 5; rank++ {
		if !completedRanks[rank] {
			errs = append(errs, fmt.Sprintf("completed_recommendations missing rank %d", rank))
		}
		if !dashboardRanks[rank] {
			errs = append(errs, fmt.Sprintf("dashboard_rows missing rank %d", rank))
		}
	}
	return joinErrors(errs)
}

func validateMonth6OperatorEvidenceDashboardRow(errs *[]string, prefix string, row AtlasMonth6OperatorEvidenceDashboardRow) {
	if row.Rank < 1 || row.Rank > 5 {
		*errs = append(*errs, prefix+".rank must be between 1 and 5")
	}
	requireField(errs, prefix+".repository", row.Repository)
	checkPublicPath(errs, prefix+".repository", row.Repository, true)
	requireField(errs, prefix+".task", row.Task)
	requireField(errs, prefix+".category", row.Category)
	if !oneOf(row.Category, "beta_operability", "evidence_ui", "golden_path", "contract_registry") {
		*errs = append(*errs, prefix+".category must be a Month 6 recommendation category")
	}
	requireField(errs, prefix+".pr", row.PR)
	if !strings.HasPrefix(row.PR, "https://github.com/uesugitorachiyo/") {
		*errs = append(*errs, prefix+".pr must be a GitHub PR URL under uesugitorachiyo")
	}
	if !lowerHex40(row.MergeCommit) {
		*errs = append(*errs, prefix+".merge_commit must be 40 lowercase hex characters")
	}
	if row.CIStatus != "passed" {
		*errs = append(*errs, prefix+".ci_status must be passed")
	}
	requireField(errs, prefix+".evidence_path", row.EvidencePath)
	checkPublicPath(errs, prefix+".evidence_path", row.EvidencePath, true)
	if row.EvidenceStatus != "available" {
		*errs = append(*errs, prefix+".evidence_status must be available")
	}
	if row.OperatorReadbackStatus != "ready" {
		*errs = append(*errs, prefix+".operator_readback_status must be ready")
	}
	if row.PublicSafetyStatus != "passed" {
		*errs = append(*errs, prefix+".public_safety_status must be passed")
	}
	if row.PromotionStatus != "no_promotion_requested" {
		*errs = append(*errs, prefix+".promotion_status must be no_promotion_requested")
	}
	if !row.RSIRemainsDenied {
		*errs = append(*errs, prefix+".rsi_remains_denied must be true")
	}
}
