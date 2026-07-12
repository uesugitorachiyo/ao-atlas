package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth6OperatorEvidenceDashboardPacketBindsCompletedRecommendations(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month6-beta-launch-v01", "nodes", "month6-recommendation-06-operator-evidence-dashboard-packet", "operator-evidence-dashboard-packet.json")

	fixture := mustLoadJSON[AtlasMonth6OperatorEvidenceDashboardPacket](t, fixturePath)
	if err := ValidateAtlasMonth6OperatorEvidenceDashboardPacket(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Status != "operator_evidence_dashboard_packet_bound" ||
		fixture.SourceRecommendationRank != 6 ||
		fixture.CompletedRecommendationCount != 5 ||
		len(fixture.CompletedRecommendations) != 5 ||
		fixture.DashboardRowCount != 5 ||
		len(fixture.DashboardRows) != 5 ||
		!fixture.AllCompletedRecommendationsBound ||
		!fixture.AllDashboardRowsHaveMergeEvidence ||
		!fixture.FixtureOnly ||
		fixture.ProviderCallsAllowed ||
		fixture.CredentialUseAllowed ||
		fixture.ReleaseOrPublishAllowed ||
		fixture.PromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Month 6 operator evidence dashboard packet lost safety or binding state: %#v", fixture)
	}
	for _, rank := range []int{1, 2, 3, 4, 5} {
		if !fixture.CompletedRecommendationsBound[rank] {
			t.Fatalf("completed recommendation rank %d is not bound: %#v", rank, fixture.CompletedRecommendationsBound)
		}
	}
	for _, row := range fixture.DashboardRows {
		if row.MergeCommit == "" ||
			row.PR == "" ||
			row.CIStatus != "passed" ||
			row.EvidenceStatus != "available" ||
			row.OperatorReadbackStatus != "ready" ||
			row.PublicSafetyStatus != "passed" ||
			row.PromotionStatus != "no_promotion_requested" ||
			!row.RSIRemainsDenied {
			t.Fatalf("dashboard row missing operator verification state: %#v", row)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(fixturePath, AtlasMonth6OperatorEvidenceDashboardPacketContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month6-operator-evidence-dashboard-packet" {
		t.Fatalf("expected typed Month 6 operator evidence dashboard packet validator, got %s", validator)
	}
}

func TestMonth6OperatorEvidenceDashboardPacketRejectsMissingMergeEvidence(t *testing.T) {
	fixture := validMonth6OperatorEvidenceDashboardPacketForTest()
	fixture.DashboardRows[0].MergeCommit = ""
	fixture.AllDashboardRowsHaveMergeEvidence = false
	if err := ValidateAtlasMonth6OperatorEvidenceDashboardPacket(fixture); err == nil {
		t.Fatal("expected missing merge evidence to be rejected")
	}
}

func validMonth6OperatorEvidenceDashboardPacketForTest() AtlasMonth6OperatorEvidenceDashboardPacket {
	return AtlasMonth6OperatorEvidenceDashboardPacket{
		Schema:                               AtlasMonth6OperatorEvidenceDashboardPacketContract,
		NodeID:                               "month6-recommendation-06-operator-evidence-dashboard-packet",
		Status:                               "operator_evidence_dashboard_packet_bound",
		SourceRecommendationRank:             6,
		SourceRecommendationTask:             "Generate operator evidence verification dashboard packet",
		SafetyGate:                           "planning_only_no_provider_no_release",
		DashboardScope:                       "month6_completed_recommendations_01_05",
		CompletedRecommendationCount:         5,
		CompletedRecommendationsBound:        map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true},
		CompletedRecommendations:             validMonth6OperatorEvidenceDashboardRowsForTest(),
		DashboardRowCount:                    5,
		DashboardRows:                        validMonth6OperatorEvidenceDashboardRowsForTest(),
		AllCompletedRecommendationsBound:     true,
		AllDashboardRowsHaveMergeEvidence:    true,
		AllDashboardRowsHaveEvidencePath:     true,
		AllDashboardRowsHaveOperatorReadback: true,
		FixtureOnly:                          true,
		ProviderCallsAllowed:                 false,
		CredentialUseAllowed:                 false,
		ReleaseOrPublishAllowed:              false,
		PromotionRequested:                   false,
		ClaimsAuthorityAdvance:               false,
		RSIRemainsDenied:                     true,
		SafeToExecute:                        false,
	}
}

func validMonth6OperatorEvidenceDashboardRowsForTest() []AtlasMonth6OperatorEvidenceDashboardRow {
	return []AtlasMonth6OperatorEvidenceDashboardRow{
		validMonth6OperatorEvidenceDashboardRowForTest(1, "ao-mission"),
		validMonth6OperatorEvidenceDashboardRowForTest(2, "ao-command"),
		validMonth6OperatorEvidenceDashboardRowForTest(3, "ao-atlas"),
		validMonth6OperatorEvidenceDashboardRowForTest(4, "ao-covenant"),
		validMonth6OperatorEvidenceDashboardRowForTest(5, "ao2-control-plane"),
	}
}

func validMonth6OperatorEvidenceDashboardRowForTest(rank int, repo string) AtlasMonth6OperatorEvidenceDashboardRow {
	return AtlasMonth6OperatorEvidenceDashboardRow{
		Rank:                   rank,
		Repository:             repo,
		Task:                   "task",
		Category:               "beta_operability",
		PR:                     "https://github.com/uesugitorachiyo/" + repo + "/pull/1",
		MergeCommit:            "0000000000000000000000000000000000000000",
		CIStatus:               "passed",
		EvidencePath:           "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-06-operator-evidence-dashboard-packet/operator-evidence-dashboard-packet.json",
		EvidenceStatus:         "available",
		OperatorReadbackStatus: "ready",
		PublicSafetyStatus:     "passed",
		PromotionStatus:        "no_promotion_requested",
		RSIRemainsDenied:       true,
	}
}
