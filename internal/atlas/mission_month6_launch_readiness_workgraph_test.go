package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth6LaunchReadinessWorkgraphBindsFortyRecommendations(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month6-beta-launch-v01", "nodes", "month6-recommendation-40-launch-readiness-workgraph", "launch-readiness-workgraph.json")

	fixture := mustLoadJSON[AtlasMonth6LaunchReadinessWorkgraph](t, fixturePath)
	if err := ValidateAtlasMonth6LaunchReadinessWorkgraph(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Status != "launch_readiness_workgraph_bound" ||
		fixture.SourceRecommendationRank != 40 ||
		fixture.CompletedRecommendationCount != 40 ||
		len(fixture.Nodes) != 40 ||
		fixture.ReadyNodes != 0 ||
		fixture.BlockedNodes != 0 ||
		fixture.FailedNodes != 0 ||
		!fixture.FinalResponseAllowed ||
		!fixture.AllRecommendationsCompleted ||
		!fixture.AllNodesHavePRCIMergeEvidence ||
		!fixture.AllNodesHaveEvidencePath ||
		!fixture.AllNodesHaveOperatorReadback ||
		!fixture.AllNodesHavePublicSafety ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories ||
		fixture.ProviderCallsAllowed ||
		fixture.CredentialUseAllowed ||
		fixture.LiveMutationAllowed ||
		fixture.ReleaseOrPublishAllowed ||
		fixture.ApprovalGranted {
		t.Fatalf("Month 6 launch-readiness workgraph lost final closure or safety state: %#v", fixture)
	}
	for rank := 1; rank <= 40; rank++ {
		if !fixture.CompletedRecommendationsBound[rank] {
			t.Fatalf("completed recommendation rank %d is not bound: %#v", rank, fixture.CompletedRecommendationsBound)
		}
	}
	seen := map[int]bool{}
	for _, node := range fixture.Nodes {
		if seen[node.Rank] {
			t.Fatalf("duplicate rank %d in launch-readiness workgraph", node.Rank)
		}
		seen[node.Rank] = true
		if node.Status != "completed" ||
			node.CIStatus != "passed" ||
			node.MergeStatus != "merged" ||
			node.EvidenceStatus != "available" ||
			node.OperatorReadbackStatus != "ready" ||
			node.PublicSafetyStatus != "passed" ||
			node.PromotionStatus != "no_promotion_requested" ||
			!node.RSIRemainsDenied ||
			node.SafeToExecute ||
			node.ExecutesWork ||
			node.MutatesRepository {
			t.Fatalf("unsafe or incomplete launch-readiness node: %#v", node)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(fixturePath, AtlasMonth6LaunchReadinessWorkgraphContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month6-launch-readiness-workgraph" {
		t.Fatalf("expected typed Month 6 launch-readiness workgraph validator, got %s", validator)
	}
}

func TestMonth6LaunchReadinessWorkgraphRejectsReadyNodes(t *testing.T) {
	fixture := validMonth6LaunchReadinessWorkgraphForTest()
	fixture.ReadyNodes = 1
	fixture.FinalResponseAllowed = false
	fixture.Nodes[0].Status = "ready"
	if err := ValidateAtlasMonth6LaunchReadinessWorkgraph(fixture); err == nil {
		t.Fatal("expected non-terminal launch-readiness workgraph to be rejected")
	}
}

func validMonth6LaunchReadinessWorkgraphForTest() AtlasMonth6LaunchReadinessWorkgraph {
	nodes := make([]AtlasMonth6LaunchReadinessNode, 0, 40)
	bound := make(map[int]bool, 40)
	for rank := 1; rank <= 40; rank++ {
		nodes = append(nodes, validMonth6LaunchReadinessNodeForTest(rank))
		bound[rank] = true
	}
	return AtlasMonth6LaunchReadinessWorkgraph{
		Schema:                        AtlasMonth6LaunchReadinessWorkgraphContract,
		NodeID:                        "month6-recommendation-40-launch-readiness-workgraph",
		Status:                        "launch_readiness_workgraph_bound",
		SourceRecommendationRank:      40,
		SourceRecommendationTask:      "Generate Month 6 launch readiness workgraph",
		SafetyGate:                    "planning_only_no_provider_no_release",
		CompletedRecommendationCount:  40,
		ReadyNodes:                    0,
		BlockedNodes:                  0,
		FailedNodes:                   0,
		FinalResponseAllowed:          true,
		CompletedRecommendationsBound: bound,
		Nodes:                         nodes,
		AllRecommendationsCompleted:   true,
		AllNodesHavePRCIMergeEvidence: true,
		AllNodesHaveEvidencePath:      true,
		AllNodesHaveOperatorReadback:  true,
		AllNodesHavePublicSafety:      true,
		FixtureOnly:                   true,
		NoPromotionRequested:          true,
		ClaimsAuthorityAdvance:        false,
		RSIRemainsDenied:              true,
		SafeToExecute:                 false,
		ExecutesWork:                  false,
		ApprovesWork:                  false,
		MutatesRepositories:           false,
		ProviderCallsAllowed:          false,
		CredentialUseAllowed:          false,
		LiveMutationAllowed:           false,
		ReleaseOrPublishAllowed:       false,
		ApprovalGranted:               false,
	}
}

func validMonth6LaunchReadinessNodeForTest(rank int) AtlasMonth6LaunchReadinessNode {
	repo := "ao-atlas"
	if rank%5 == 1 {
		repo = "ao-mission"
	}
	return AtlasMonth6LaunchReadinessNode{
		Rank:                   rank,
		NodeID:                 month6LaunchReadinessNodeID(rank),
		Repository:             repo,
		Task:                   "task",
		Category:               "contract_registry",
		Status:                 "completed",
		PR:                     "https://github.com/uesugitorachiyo/" + repo + "/pull/1",
		MergeCommit:            "0000000000000000000000000000000000000000",
		CIStatus:               "passed",
		MergeStatus:            "merged",
		EvidencePath:           "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-40-launch-readiness-workgraph/launch-readiness-workgraph.json",
		EvidenceStatus:         "available",
		OperatorReadbackStatus: "ready",
		PublicSafetyStatus:     "passed",
		PromotionStatus:        "no_promotion_requested",
		RSIRemainsDenied:       true,
		SafeToExecute:          false,
		ExecutesWork:           false,
		MutatesRepository:      false,
	}
}
