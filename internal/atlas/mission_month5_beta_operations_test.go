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

type month5CompatibilityLedgerFixture struct {
	Schema                 string                        `json:"schema"`
	NodeID                 string                        `json:"node_id"`
	MissionID              string                        `json:"mission_id"`
	Status                 string                        `json:"status"`
	CompatibilityRows      []month5CompatibilityLedgerRow `json:"compatibility_rows"`
	NoPromotionRequested   bool                          `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool                          `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                          `json:"rsi_remains_denied"`
	SafeToExecute          bool                          `json:"safe_to_execute"`
}

type month5CompatibilityLedgerRow struct {
	Contract        string   `json:"contract"`
	Producer        string   `json:"producer"`
	Consumers       []string `json:"consumers"`
	RequiredFixture string   `json:"required_fixture"`
	DigestBinding   string   `json:"digest_binding"`
	Status          string   `json:"status"`
}

type month5BlueprintCanonicalBytesFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	CanonicalBytesRef      string   `json:"canonical_bytes_ref"`
	CanonicalBytesDigest   string   `json:"canonical_bytes_digest"`
	PreservedFields        []string `json:"preserved_fields"`
	DownstreamConsumers    []string `json:"downstream_consumers"`
	RejectedTransforms     []string `json:"rejected_transforms"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5AtlasCompatibilityMatrixFixture struct {
	Schema                 string                       `json:"schema"`
	NodeID                 string                       `json:"node_id"`
	MissionID              string                       `json:"mission_id"`
	Status                 string                       `json:"status"`
	MatrixRows             []month5AtlasCompatibilityRow `json:"matrix_rows"`
	CompatibilityFloorMet  bool                         `json:"compatibility_floor_met"`
	NoPromotionRequested   bool                         `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                         `json:"rsi_remains_denied"`
	SafeToExecute          bool                         `json:"safe_to_execute"`
}

type month5AtlasCompatibilityRow struct {
	Contract       string   `json:"contract"`
	Owner          string   `json:"owner"`
	Producer       string   `json:"producer"`
	Consumers      []string `json:"consumers"`
	AtlasArtifact  string   `json:"atlas_artifact"`
	Check          string   `json:"check"`
}

type month5FoundrySafeNextWorkFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	SourceWorkgraph        string   `json:"source_workgraph"`
	FoundryImport          string   `json:"foundry_import"`
	FirstExecutableNode    string   `json:"first_executable_node"`
	ExactlyOneActiveNode   bool     `json:"exactly_one_active_node"`
	ReadinessChecks        []string `json:"readiness_checks"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5ForgeGoalRunFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	GoalRunID              string   `json:"goal_run_id"`
	Mode                   string   `json:"mode"`
	LifecycleStates        []string `json:"lifecycle_states"`
	BoundaryChecks         []string `json:"boundary_checks"`
	NoProviderCalls        bool     `json:"no_provider_calls"`
	NoMutationApplication  bool     `json:"no_mutation_application"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5CommandThinClientBoundaryFixture struct {
	Schema                   string   `json:"schema"`
	NodeID                   string   `json:"node_id"`
	MissionID                string   `json:"mission_id"`
	Status                   string   `json:"status"`
	CommandRole              string   `json:"command_role"`
	SourceReadback           string   `json:"source_readback"`
	OwnedByMission           []string `json:"owned_by_mission"`
	CommandAllowed           []string `json:"command_allowed"`
	CommandForbidden         []string `json:"command_forbidden"`
	NoPromotionRequested     bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance   bool     `json:"claims_authority_advance"`
	RSIRemainsDenied         bool     `json:"rsi_remains_denied"`
	SafeToExecute            bool     `json:"safe_to_execute"`
	MissionStateMutation     bool     `json:"mission_state_mutation"`
	PolicyOverrideAllowed    bool     `json:"policy_override_allowed"`
	PromotionDecisionAllowed bool     `json:"promotion_decision_allowed"`
}

type month5AO2ExactApprovalBytesFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	BaseCommit             string   `json:"base_commit"`
	ProposedBytesDigest    string   `json:"proposed_bytes_digest"`
	PatchDigest            string   `json:"patch_digest"`
	ApprovalTicketDigest   string   `json:"approval_ticket_digest"`
	RequiredBindings       []string `json:"required_bindings"`
	RejectedApprovalModes  []string `json:"rejected_approval_modes"`
	NoAutoApproval         bool     `json:"no_auto_approval"`
	NoProviderCalls        bool     `json:"no_provider_calls"`
	NoMutationApplication  bool     `json:"no_mutation_application"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5AO2AutoApprovalDenialFixture struct {
	Schema                   string   `json:"schema"`
	NodeID                   string   `json:"node_id"`
	MissionID                string   `json:"mission_id"`
	Status                   string   `json:"status"`
	DeniedPaths              []string `json:"denied_paths"`
	RequiredOperatorControls []string `json:"required_operator_controls"`
	DenialReasons            []string `json:"denial_reasons"`
	AutoApprovalAllowed      bool     `json:"auto_approval_allowed"`
	HardcodedIdentityAllowed bool     `json:"hardcoded_identity_allowed"`
	NoProviderCalls          bool     `json:"no_provider_calls"`
	NoMutationApplication    bool     `json:"no_mutation_application"`
	NoPromotionRequested     bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance   bool     `json:"claims_authority_advance"`
	RSIRemainsDenied         bool     `json:"rsi_remains_denied"`
	SafeToExecute            bool     `json:"safe_to_execute"`
}

type month5CovenantPolicyHashFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	PolicyHash             string   `json:"policy_hash"`
	TicketDigest           string   `json:"ticket_digest"`
	BoundPolicyFields      []string `json:"bound_policy_fields"`
	RejectedOmissions      []string `json:"rejected_omissions"`
	CovenantAuthority       string   `json:"covenant_authority"`
	CommandCompatibility    string   `json:"command_compatibility"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
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

func TestMonth5MissionBlueprintAtlasCompatibilityLedgerFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-04", "mission-blueprint-atlas-compatibility-ledger.json")
	fixture := mustLoadJSON[month5CompatibilityLedgerFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.mission-blueprint-atlas-compatibility-ledger.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-04" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "compatibility_rows_ready" {
		t.Fatalf("unexpected compatibility ledger header: %#v", fixture)
	}
	if len(fixture.CompatibilityRows) < 6 {
		t.Fatalf("expected at least six compatibility rows: %#v", fixture.CompatibilityRows)
	}
	rows := map[string]month5CompatibilityLedgerRow{}
	for _, row := range fixture.CompatibilityRows {
		rows[row.Contract] = row
		if row.Producer == "" || len(row.Consumers) == 0 || row.RequiredFixture == "" || row.DigestBinding == "" || row.Status == "" {
			t.Fatalf("compatibility row must include producer, consumers, fixture, digest binding, and status: %#v", row)
		}
	}
	for _, required := range []string{"mission.intake", "blueprint.canonical-pack", "atlas.workgraph", "atlas.context-pack", "atlas.readback", "foundry.import"} {
		if _, ok := rows[required]; !ok {
			t.Fatalf("missing compatibility row for %s", required)
		}
	}
	if rows["blueprint.canonical-pack"].DigestBinding != "canonical_bytes_sha256" ||
		rows["atlas.workgraph"].DigestBinding != "workgraph_sha256" ||
		rows["atlas.readback"].DigestBinding != "readback_sha256" {
		t.Fatalf("digest bindings drifted: blueprint=%#v workgraph=%#v readback=%#v", rows["blueprint.canonical-pack"], rows["atlas.workgraph"], rows["atlas.readback"])
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("compatibility ledger changed safety posture: %#v", fixture)
	}
}

func TestMonth5BlueprintCanonicalBytesPreservationFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-05", "blueprint-canonical-bytes-fixture.json")
	fixture := mustLoadJSON[month5BlueprintCanonicalBytesFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.blueprint-canonical-bytes-fixture.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-05" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "canonical_bytes_preserved" {
		t.Fatalf("unexpected Blueprint canonical bytes fixture header: %#v", fixture)
	}
	if fixture.CanonicalBytesRef == "" || !digestPattern.MatchString(fixture.CanonicalBytesDigest) {
		t.Fatalf("canonical bytes fixture must bind a portable ref and sha256 digest: %#v", fixture)
	}
	for _, required := range []string{"schema", "requirements", "authorization_scope", "operator_prompt_digest", "canonical_bytes_digest"} {
		if !containsValue(fixture.PreservedFields, required) {
			t.Fatalf("canonical bytes fixture missing preserved field %s: %#v", required, fixture.PreservedFields)
		}
	}
	for _, consumer := range []string{"ao-mission", "ao-atlas", "ao-foundry"} {
		if !containsValue(fixture.DownstreamConsumers, consumer) {
			t.Fatalf("canonical bytes fixture missing consumer %s: %#v", consumer, fixture.DownstreamConsumers)
		}
	}
	if !containsValue(fixture.RejectedTransforms, "schema_alias_rewrite") ||
		!containsValue(fixture.RejectedTransforms, "field_renaming_without_digest_change") {
		t.Fatalf("canonical bytes fixture must reject lossy transforms: %#v", fixture.RejectedTransforms)
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("canonical bytes fixture changed safety posture: %#v", fixture)
	}
}

func TestMonth5AtlasCompatibilityMatrixFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-06", "atlas-compatibility-matrix.json")
	fixture := mustLoadJSON[month5AtlasCompatibilityMatrixFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.compatibility-matrix.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-06" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "matrix_rows_ready" ||
		!fixture.CompatibilityFloorMet {
		t.Fatalf("unexpected Atlas compatibility matrix header: %#v", fixture)
	}
	if len(fixture.MatrixRows) < 10 {
		t.Fatalf("expected at least ten Atlas compatibility matrix rows: %#v", fixture.MatrixRows)
	}
	rows := map[string]month5AtlasCompatibilityRow{}
	for _, row := range fixture.MatrixRows {
		rows[row.Contract] = row
		if row.Owner == "" || row.Producer == "" || len(row.Consumers) == 0 || row.AtlasArtifact == "" || row.Check == "" {
			t.Fatalf("compatibility matrix row must be complete: %#v", row)
		}
	}
	for _, required := range []string{"recommendation-wave", "recommendation-workgraph", "recommendation-readback", "workgraph-readiness-packet", "foundry-import", "run-link"} {
		if _, ok := rows[required]; !ok {
			t.Fatalf("missing Atlas compatibility matrix row for %s", required)
		}
	}
	if rows["recommendation-workgraph"].Check != "workgraph validate" ||
		rows["foundry-import"].Check != "foundry import replay" ||
		rows["run-link"].Check != "run-link validate" {
		t.Fatalf("matrix executable checks drifted: %#v %#v %#v", rows["recommendation-workgraph"], rows["foundry-import"], rows["run-link"])
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Atlas compatibility matrix changed safety posture: %#v", fixture)
	}
}

func TestMonth5FoundrySafeNextWorkReadinessFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-07", "foundry-safe-next-work-readiness.json")
	fixture := mustLoadJSON[month5FoundrySafeNextWorkFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.foundry-safe-next-work-readiness.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-07" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "safe_next_work_bound" {
		t.Fatalf("unexpected Foundry safe-next-work fixture header: %#v", fixture)
	}
	if fixture.SourceWorkgraph == "" || fixture.FoundryImport == "" ||
		fixture.FirstExecutableNode != "mission-recommendation-month5-beta-operations-07" ||
		!fixture.ExactlyOneActiveNode {
		t.Fatalf("Foundry readiness must bind source workgraph, import, and exactly one active node: %#v", fixture)
	}
	for _, required := range []string{"workgraph_first_executable_matches_import", "no_concurrent_mutation", "run_link_required_before_advance", "final_response_denied_while_ready_nodes_remain"} {
		if !containsValue(fixture.ReadinessChecks, required) {
			t.Fatalf("Foundry readiness missing check %s: %#v", required, fixture.ReadinessChecks)
		}
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Foundry readiness changed safety posture: %#v", fixture)
	}
}

func TestMonth5ForgeGoalRunDryRunLifecycleFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-08", "forge-goalrun-dry-run-lifecycle.json")
	fixture := mustLoadJSON[month5ForgeGoalRunFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.forge-goalrun-dry-run-lifecycle.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-08" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "dry_run_lifecycle_bound" ||
		fixture.Mode != "fixture_only_no_execution" ||
		fixture.GoalRunID == "" {
		t.Fatalf("unexpected Forge GoalRun fixture header: %#v", fixture)
	}
	for _, required := range []string{"created", "planned", "policy_checked", "readback_recorded", "closed_without_execution"} {
		if !containsValue(fixture.LifecycleStates, required) {
			t.Fatalf("GoalRun lifecycle missing state %s: %#v", required, fixture.LifecycleStates)
		}
	}
	for _, required := range []string{"no_provider_calls", "no_mutation_application", "no_release_or_deploy", "rollback_receipt_required_before_live_use"} {
		if !containsValue(fixture.BoundaryChecks, required) {
			t.Fatalf("GoalRun boundary missing check %s: %#v", required, fixture.BoundaryChecks)
		}
	}
	if !fixture.NoProviderCalls ||
		!fixture.NoMutationApplication ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Forge GoalRun fixture changed safety posture: %#v", fixture)
	}
}

func TestMonth5CommandThinClientBoundaryFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-09", "command-thin-client-boundary.json")
	fixture := mustLoadJSON[month5CommandThinClientBoundaryFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.command-thin-client-boundary.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-09" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "mission_readback_ownership_bound" ||
		fixture.CommandRole != "thin_client_presentation_only" ||
		fixture.SourceReadback == "" {
		t.Fatalf("unexpected Command thin-client fixture header: %#v", fixture)
	}
	for _, required := range []string{"durable_state", "routing_state", "final_response_gate", "exact_next_action"} {
		if !containsValue(fixture.OwnedByMission, required) {
			t.Fatalf("Command boundary missing Mission-owned field %s: %#v", required, fixture.OwnedByMission)
		}
	}
	for _, required := range []string{"compact_timeline", "status_projection", "approval_inbox_projection"} {
		if !containsValue(fixture.CommandAllowed, required) {
			t.Fatalf("Command boundary missing allowed projection %s: %#v", required, fixture.CommandAllowed)
		}
	}
	for _, forbidden := range []string{"mission_state_mutation", "policy_override", "promotion_decision"} {
		if !containsValue(fixture.CommandForbidden, forbidden) {
			t.Fatalf("Command boundary missing forbidden authority %s: %#v", forbidden, fixture.CommandForbidden)
		}
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.MissionStateMutation ||
		fixture.PolicyOverrideAllowed ||
		fixture.PromotionDecisionAllowed {
		t.Fatalf("Command thin-client fixture changed authority posture: %#v", fixture)
	}
}

func TestMonth5AO2ExactApprovalBytesFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-10", "ao2-exact-approval-bytes.json")
	fixture := mustLoadJSON[month5AO2ExactApprovalBytesFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.ao2-exact-approval-bytes.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-10" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "approval_bytes_bound" ||
		fixture.BaseCommit == "" {
		t.Fatalf("unexpected AO2 approval fixture header: %#v", fixture)
	}
	for name, digest := range map[string]string{
		"proposed_bytes_digest":  fixture.ProposedBytesDigest,
		"patch_digest":           fixture.PatchDigest,
		"approval_ticket_digest": fixture.ApprovalTicketDigest,
	} {
		if !digestPattern.MatchString(digest) {
			t.Fatalf("%s must be sha256-bound: %s", name, digest)
		}
	}
	for _, required := range []string{"base_commit", "proposed_bytes_digest", "patch_digest", "approval_ticket_digest", "operator_identity"} {
		if !containsValue(fixture.RequiredBindings, required) {
			t.Fatalf("AO2 approval fixture missing binding %s: %#v", required, fixture.RequiredBindings)
		}
	}
	for _, rejected := range []string{"hardcoded_identity_auto_approval", "patchless_approval", "base_commit_omitted", "digest_substitution"} {
		if !containsValue(fixture.RejectedApprovalModes, rejected) {
			t.Fatalf("AO2 approval fixture missing rejected mode %s: %#v", rejected, fixture.RejectedApprovalModes)
		}
	}
	if !fixture.NoAutoApproval ||
		!fixture.NoProviderCalls ||
		!fixture.NoMutationApplication ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("AO2 approval fixture changed safety posture: %#v", fixture)
	}
}

func TestMonth5AO2AutoApprovalDenialFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-11", "ao2-auto-approval-denial.json")
	fixture := mustLoadJSON[month5AO2AutoApprovalDenialFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.ao2-auto-approval-denial.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-11" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "hardcoded_identity_paths_denied" {
		t.Fatalf("unexpected AO2 auto-approval denial header: %#v", fixture)
	}
	for _, denied := range []string{"hardcoded_identity_auto_approval", "operator_identity_substitution", "approval_without_digest_binding"} {
		if !containsValue(fixture.DeniedPaths, denied) {
			t.Fatalf("AO2 denial fixture missing denied path %s: %#v", denied, fixture.DeniedPaths)
		}
	}
	for _, required := range []string{"explicit_operator_approval", "exact_digest_match", "base_commit_match", "rollback_receipt_before_apply"} {
		if !containsValue(fixture.RequiredOperatorControls, required) {
			t.Fatalf("AO2 denial fixture missing operator control %s: %#v", required, fixture.RequiredOperatorControls)
		}
	}
	if len(fixture.DenialReasons) < 3 {
		t.Fatalf("AO2 denial fixture must include concrete denial reasons: %#v", fixture.DenialReasons)
	}
	if fixture.AutoApprovalAllowed ||
		fixture.HardcodedIdentityAllowed ||
		!fixture.NoProviderCalls ||
		!fixture.NoMutationApplication ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("AO2 auto-approval denial changed safety posture: %#v", fixture)
	}
}

func TestMonth5CovenantPolicyHashFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-12", "covenant-policy-hash-binding.json")
	fixture := mustLoadJSON[month5CovenantPolicyHashFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.covenant-policy-hash-binding.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-12" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "policy_fields_bound" ||
		fixture.CovenantAuthority != "policy_and_contract_authority" ||
		fixture.CommandCompatibility != "must_reject_ticket_covenant_rejects" {
		t.Fatalf("unexpected Covenant policy hash fixture header: %#v", fixture)
	}
	for name, digest := range map[string]string{
		"policy_hash":   fixture.PolicyHash,
		"ticket_digest": fixture.TicketDigest,
	} {
		if !digestPattern.MatchString(digest) {
			t.Fatalf("%s must be sha256-bound: %s", name, digest)
		}
	}
	for _, required := range []string{"policy_id", "policy_version", "decision", "constraints", "scope", "expires_at"} {
		if !containsValue(fixture.BoundPolicyFields, required) {
			t.Fatalf("Covenant policy hash missing bound field %s: %#v", required, fixture.BoundPolicyFields)
		}
	}
	for _, rejected := range []string{"policyless_ticket_digest", "constraints_omitted", "decision_omitted", "command_accepts_covenant_reject"} {
		if !containsValue(fixture.RejectedOmissions, rejected) {
			t.Fatalf("Covenant policy hash missing rejected omission %s: %#v", rejected, fixture.RejectedOmissions)
		}
	}
	if !fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Covenant policy hash fixture changed safety posture: %#v", fixture)
	}
}
