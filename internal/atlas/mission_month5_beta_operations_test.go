package atlas

import (
	"path/filepath"
	"testing"
)

type month5AuthorityManifestFixture struct {
	Schema                 string                    `json:"schema"`
	NodeID                 string                    `json:"node_id"`
	MissionID              string                    `json:"mission_id"`
	Status                 string                    `json:"status"`
	StackLockfileVersion   string                    `json:"stack_lockfile_version"`
	GeneratedFromReadback  string                    `json:"generated_from_readback"`
	Repositories           []month5AuthorityRepo     `json:"repositories"`
	AuthorityBoundaries    []month5AuthorityBoundary `json:"authority_boundaries"`
	NoPromotionRequested   bool                      `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool                      `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                      `json:"rsi_remains_denied"`
	SafeToExecute          bool                      `json:"safe_to_execute"`
	ExecutesWork           bool                      `json:"executes_work"`
	ApprovesWork           bool                      `json:"approves_work"`
}

type month5AuthorityRepo struct {
	Name          string `json:"name"`
	Role          string `json:"role"`
	Authority     string `json:"authority"`
	SourceOfTruth string `json:"source_of_truth"`
}

type month5AuthorityBoundary struct {
	Owner      string `json:"owner"`
	Boundary   string `json:"boundary"`
	Constraint string `json:"constraint"`
}

type month5ArchitectureSourceTruthFixture struct {
	Schema                       string                    `json:"schema"`
	NodeID                       string                    `json:"node_id"`
	MissionID                    string                    `json:"mission_id"`
	Status                       string                    `json:"status"`
	ArchitectureReadinessSource  string                    `json:"architecture_readiness_source"`
	RepositoryBehaviorSource     string                    `json:"repository_behavior_source"`
	IncludedRepositories         []string                  `json:"included_repositories"`
	CurrentAuthorityStatements   []month5AuthorityBoundary `json:"current_authority_statements"`
	OutdatedDocumentationSignals []string                  `json:"outdated_documentation_signals"`
	NoPromotionRequested         bool                      `json:"no_promotion_requested"`
	RSIRemainsDenied             bool                      `json:"rsi_remains_denied"`
	ClaimsAuthorityAdvance       bool                      `json:"claims_authority_advance"`
	SafeToExecute                bool                      `json:"safe_to_execute"`
}

type month5SchemaRegistryHandoffFixture struct {
	Schema                 string                    `json:"schema"`
	NodeID                 string                    `json:"node_id"`
	MissionID              string                    `json:"mission_id"`
	Status                 string                    `json:"status"`
	RegistryOwner          string                    `json:"registry_owner"`
	ProducerConsumerRows   []month5ContractOwnerRow  `json:"producer_consumer_rows"`
	RequiredCompatibility  []string                  `json:"required_compatibility_checks"`
	AuthorityBoundaries    []month5AuthorityBoundary `json:"authority_boundaries"`
	NoPromotionRequested   bool                      `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool                      `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                      `json:"rsi_remains_denied"`
	SafeToExecute          bool                      `json:"safe_to_execute"`
}

type month5ContractOwnerRow struct {
	Contract string   `json:"contract"`
	Owner    string   `json:"owner"`
	Producer string   `json:"producer"`
	Consumers []string `json:"consumers"`
	Status   string   `json:"status"`
}

func TestMonth5BetaOperationsRecommendationsImportAsLongRunWave(t *testing.T) {
	root := repoRoot(t)
	recommendationsPath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "month5-beta-operations-recommendations.json")
	recommendations := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, recommendationsPath)
	if err := ValidateAOMissionFeatureDepthRecommendations(recommendations, 40); err != nil {
		t.Fatalf("Month 5 recommendations are not importable: %v", err)
	}

	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath:  recommendationsPath,
		TargetInstance:       "ao-stack-month5-beta-operations-v01",
		MinTasks:             40,
		NodeBudget:           40,
		EstimatedMinutes:     150,
		MinMinutes:           120,
		MaxMinutes:           180,
		ContinueIfFastTarget: 40,
		ReturnOnlyWhen:       "all_40_month5_beta_operations_nodes_complete_or_true_hard_blocker",
		CheckpointPolicy:     "after_each_node_or_timed_interval",
		EvidencePolicy:       "implementation_or_contract_fixture_plus_tests_verification_public_safety_promoter_command",
		FinalReportContract:  "ao.atlas.month5-beta-operations-final-report.v0.1",
	})
	if err != nil {
		t.Fatalf("build Month 5 recommendation wave: %v", err)
	}

	if result.Wave.MissionID != "mission-4d91b0a9e4ab273e" ||
		result.Wave.TargetInstance != "ao-stack-month5-beta-operations-v01" ||
		result.Wave.MinimumTasks != 40 ||
		result.Wave.TotalTasks != 40 ||
		result.Wave.NodeBudget != 40 ||
		result.Wave.EstimatedMinutes != 150 ||
		result.Wave.Supervisor.MinMinutes != 120 ||
		result.Wave.Supervisor.MaxMinutes != 180 ||
		result.Wave.Supervisor.ContinueIfFastTarget != 40 ||
		result.Wave.FinalResponseAllowed ||
		result.Wave.SafeToExecute ||
		result.Wave.SchedulesWork ||
		result.Wave.ExecutesWork ||
		result.Wave.ApprovesWork {
		t.Fatalf("unexpected Month 5 wave contract: %#v", result.Wave)
	}
	if len(result.Workgraph.Nodes) != 40 {
		t.Fatalf("expected 40 generated nodes, got %d", len(result.Workgraph.Nodes))
	}
	if result.Workgraph.Nodes[0].ID != "mission-recommendation-month5-beta-operations-01" ||
		result.Workgraph.Nodes[39].ID != "mission-recommendation-month5-beta-operations-40" {
		t.Fatalf("unexpected Month 5 node range: first=%s last=%s", result.Workgraph.Nodes[0].ID, result.Workgraph.Nodes[39].ID)
	}
	for i, node := range result.Workgraph.Nodes {
		if node.Status != "ready" {
			t.Fatalf("node %d should start ready: %#v", i+1, node)
		}
		if node.FactoryTask.TargetFactoryRepo != "ao-atlas" ||
			node.FactoryTask.MutationClass != "low_risk_code" ||
			node.FactoryTask.AuthorityBoundary != "atlas_recommendation_planning_only" {
			t.Fatalf("node %d has unexpected bounded task contract: %#v", i+1, node.FactoryTask)
		}
	}
}

func TestMonth5StackLockfileAuthorityManifestFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-01", "stack-lockfile-authority-manifest.json")
	fixture := mustLoadJSON[month5AuthorityManifestFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.stack-lockfile-authority-manifest.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-01" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "ready_for_beta_operations_preflight" ||
		fixture.StackLockfileVersion != "ao-stack-month5-beta-operations-v0.1" {
		t.Fatalf("unexpected Month 5 authority manifest header: %#v", fixture)
	}
	if fixture.GeneratedFromReadback != "docs/evidence/ao-stack-month4-consolidation-v01/final-closure/recommendation-readback-after-node-36.json" {
		t.Fatalf("manifest must bind to Month 4 parent closure readback: %s", fixture.GeneratedFromReadback)
	}
	if len(fixture.Repositories) != 14 {
		t.Fatalf("expected 14 active AO repositories in stack lockfile, got %d", len(fixture.Repositories))
	}
	repos := map[string]month5AuthorityRepo{}
	for _, repo := range fixture.Repositories {
		repos[repo.Name] = repo
		if repo.Role == "" || repo.Authority == "" || repo.SourceOfTruth == "" {
			t.Fatalf("repo row must include role, authority, and source of truth: %#v", repo)
		}
	}
	for _, required := range []string{"ao-mission", "ao-blueprint", "ao-atlas", "ao-foundry", "ao-forge", "ao-covenant", "ao2", "ao2-control-plane", "ao-command", "ao-arena", "ao-crucible", "ao-sentinel", "ao-promoter", "ao-architecture"} {
		if _, ok := repos[required]; !ok {
			t.Fatalf("missing repository from Month 5 stack lockfile: %s", required)
		}
	}
	if repos["ao-covenant"].Authority != "policy_and_contract_authority" ||
		repos["ao2"].Authority != "execution_runtime_authority" ||
		repos["ao-mission"].Authority != "mission_state_authority" ||
		repos["ao-atlas"].Authority != "workgraph_context_authority" {
		t.Fatalf("core authority rows drifted: covenant=%#v ao2=%#v mission=%#v atlas=%#v", repos["ao-covenant"], repos["ao2"], repos["ao-mission"], repos["ao-atlas"])
	}
	if len(fixture.AuthorityBoundaries) < 8 {
		t.Fatalf("authority manifest must record concrete cross-component boundaries: %#v", fixture.AuthorityBoundaries)
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork {
		t.Fatalf("authority manifest changed promotion or execution boundaries: %#v", fixture)
	}
}

func TestMonth5ArchitectureSourceTruthReadbackFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-02", "architecture-source-truth-readback.json")
	fixture := mustLoadJSON[month5ArchitectureSourceTruthFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.architecture-source-truth-readback.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-02" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "current_behavior_inventory_bound" {
		t.Fatalf("unexpected architecture source-truth header: %#v", fixture)
	}
	if fixture.ArchitectureReadinessSource == "" || fixture.RepositoryBehaviorSource == "" {
		t.Fatalf("architecture readback must bind both documentation and behavior sources: %#v", fixture)
	}
	if !containsValue(fixture.IncludedRepositories, "ao-mission") ||
		!containsValue(fixture.IncludedRepositories, "ao-blueprint") ||
		!containsValue(fixture.IncludedRepositories, "ao-atlas") ||
		len(fixture.IncludedRepositories) != 14 {
		t.Fatalf("architecture source-truth inventory must include all active repositories: %#v", fixture.IncludedRepositories)
	}
	foundMissionBoundary := false
	foundBlueprintBoundary := false
	foundRSIBoundary := false
	for _, statement := range fixture.CurrentAuthorityStatements {
		switch statement.Owner {
		case "ao-mission":
			foundMissionBoundary = true
		case "ao-blueprint":
			foundBlueprintBoundary = true
		case "rsi":
			foundRSIBoundary = true
		}
	}
	if !foundMissionBoundary || !foundBlueprintBoundary || !foundRSIBoundary {
		t.Fatalf("source-truth readback must preserve Mission, Blueprint, and RSI boundaries: %#v", fixture.CurrentAuthorityStatements)
	}
	if len(fixture.OutdatedDocumentationSignals) < 2 {
		t.Fatalf("expected concrete stale documentation signals: %#v", fixture.OutdatedDocumentationSignals)
	}
	if !fixture.NoPromotionRequested ||
		!fixture.RSIRemainsDenied ||
		fixture.ClaimsAuthorityAdvance ||
		fixture.SafeToExecute {
		t.Fatalf("source-truth readback changed safety posture: %#v", fixture)
	}
}

func TestMonth5CovenantSchemaRegistryHandoffFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-03", "covenant-schema-registry-handoff.json")
	fixture := mustLoadJSON[month5SchemaRegistryHandoffFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.covenant-schema-registry-handoff.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-03" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "registry_handoff_ready" ||
		fixture.RegistryOwner != "ao-covenant" {
		t.Fatalf("unexpected Covenant schema registry handoff header: %#v", fixture)
	}
	if len(fixture.ProducerConsumerRows) < 8 {
		t.Fatalf("expected at least eight producer/consumer contract rows: %#v", fixture.ProducerConsumerRows)
	}
	rows := map[string]month5ContractOwnerRow{}
	for _, row := range fixture.ProducerConsumerRows {
		rows[row.Contract] = row
		if row.Owner == "" || row.Producer == "" || len(row.Consumers) == 0 || row.Status == "" {
			t.Fatalf("contract row must include owner, producer, consumers, and status: %#v", row)
		}
	}
	for _, required := range []string{"covenant.approval-ticket", "mission.blueprint-pack", "atlas.workgraph", "foundry.run-link", "ao2.approval-digest"} {
		if _, ok := rows[required]; !ok {
			t.Fatalf("missing schema registry row for %s", required)
		}
	}
	if rows["covenant.approval-ticket"].Owner != "ao-covenant" ||
		rows["ao2.approval-digest"].Owner != "ao-covenant" {
		t.Fatalf("Covenant must own gate-critical approval contracts: %#v %#v", rows["covenant.approval-ticket"], rows["ao2.approval-digest"])
	}
	if !containsValue(fixture.RequiredCompatibility, "producer_consumer_fixture_roundtrip") ||
		!containsValue(fixture.RequiredCompatibility, "canonical_json_digest_vectors") {
		t.Fatalf("handoff must require executable compatibility checks: %#v", fixture.RequiredCompatibility)
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("schema registry handoff changed safety posture: %#v", fixture)
	}
}
