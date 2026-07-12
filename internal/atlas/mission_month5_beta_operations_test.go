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

type month5ControlPlaneTransactionalEvidenceFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	StorageAuthority       string   `json:"storage_authority"`
	RequiredAtomicSteps    []string `json:"required_atomic_steps"`
	RollbackReadback       []string `json:"rollback_readback"`
	RejectedStates         []string `json:"rejected_states"`
	ContentAddressedStore  bool     `json:"content_addressed_store"`
	IndexCommitAtomic      bool     `json:"index_commit_atomic"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5ControlPlaneMigrationMetadataFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	MigrationID            string   `json:"migration_id"`
	FromVersion            string   `json:"from_version"`
	ToVersion              string   `json:"to_version"`
	MetadataChecksum       string   `json:"metadata_checksum"`
	RequiredMetadata       []string `json:"required_metadata"`
	ReplayChecks           []string `json:"replay_checks"`
	NoDestructiveMigration bool     `json:"no_destructive_migration"`
	RollbackPlanBound      bool     `json:"rollback_plan_bound"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5LocalBackupRestoreDrillFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	BackupManifestDigest   string   `json:"backup_manifest_digest"`
	RestoreTarget          string   `json:"restore_target"`
	RequiredArtifacts      []string `json:"required_artifacts"`
	RestoreVerification    []string `json:"restore_verification"`
	NoDestructiveRestore   bool     `json:"no_destructive_restore"`
	DrillIsFixtureOnly     bool     `json:"drill_is_fixture_only"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5MissionRestartReplayFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	SourceCheckpoint       string   `json:"source_checkpoint"`
	ReplayInputs           []string `json:"replay_inputs"`
	ExpectedReadbacks      []string `json:"expected_readbacks"`
	ExactlyOnceAccounting  bool     `json:"exactly_once_accounting"`
	DuplicateNodeCompletion bool     `json:"duplicate_node_completion"`
	FinalResponseAllowed   bool     `json:"final_response_allowed"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5MissionKillRestartReplayFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	InterruptedNode        string   `json:"interrupted_node"`
	ResumeCheckpoint       string   `json:"resume_checkpoint"`
	InterruptionMarkers    []string `json:"interruption_markers"`
	ResumeAssertions       []string `json:"resume_assertions"`
	PartialNodePromoted    bool     `json:"partial_node_promoted"`
	DuplicateRunLink       bool     `json:"duplicate_run_link"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5GoldenPathDryRunReadinessFixture struct {
	Schema                 string                       `json:"schema"`
	NodeID                 string                       `json:"node_id"`
	MissionID              string                       `json:"mission_id"`
	Status                 string                       `json:"status"`
	MatrixRows             []month5GoldenPathReadinessRow `json:"matrix_rows"`
	DryRunOnly             bool                         `json:"dry_run_only"`
	NoProviderExecution    bool                         `json:"no_provider_execution"`
	NoPromotionRequested   bool                         `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                         `json:"rsi_remains_denied"`
	SafeToExecute          bool                         `json:"safe_to_execute"`
}

type month5GoldenPathReadinessRow struct {
	Component string `json:"component"`
	Handoff   string `json:"handoff"`
	Check     string `json:"check"`
	Status    string `json:"status"`
}

type month5CleanRoomNonAOReplayFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	TargetRepoClass        string   `json:"target_repo_class"`
	IsolatedWorktree       bool     `json:"isolated_worktree"`
	ReplayInputs           []string `json:"replay_inputs"`
	ExternalRepoBoundaries []string `json:"external_repo_boundaries"`
	NoExternalMutation     bool     `json:"no_external_mutation"`
	NoProviderExecution    bool     `json:"no_provider_execution"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5ArenaHostedCIWorkflowFixture struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	Repository             string   `json:"repository"`
	WorkflowPath           string   `json:"workflow_path"`
	RequiredJobs           []string `json:"required_jobs"`
	TriggerModes           []string `json:"trigger_modes"`
	FixtureOnly            bool     `json:"fixture_only"`
	NoWorkflowMutation     bool     `json:"no_workflow_mutation"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5CrucibleHostedCIWorkflowFixture struct {
	Schema                    string   `json:"schema"`
	NodeID                    string   `json:"node_id"`
	MissionID                 string   `json:"mission_id"`
	Status                    string   `json:"status"`
	Repository                string   `json:"repository"`
	WorkflowPath              string   `json:"workflow_path"`
	RequiredJobs              []string `json:"required_jobs"`
	TriggerModes              []string `json:"trigger_modes"`
	FailureInjectionReadbacks []string `json:"failure_injection_readbacks"`
	FixtureOnly               bool     `json:"fixture_only"`
	NoWorkflowMutation        bool     `json:"no_workflow_mutation"`
	NoPromotionRequested      bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance    bool     `json:"claims_authority_advance"`
	RSIRemainsDenied          bool     `json:"rsi_remains_denied"`
	SafeToExecute             bool     `json:"safe_to_execute"`
}

type month5SentinelHostedCINativeSignalBinding struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	Repository             string   `json:"repository"`
	FixtureRef             string   `json:"fixture_ref"`
	NativeSignalReadbacks  []string `json:"native_signal_readbacks"`
	FixtureOnly            bool     `json:"fixture_only"`
	NoWorkflowMutation     bool     `json:"no_workflow_mutation"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5PromoterHostedCINoActivationBinding struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	Repository             string   `json:"repository"`
	FixtureRef             string   `json:"fixture_ref"`
	WorkflowPath           string   `json:"workflow_path"`
	RequiredJobs           []string `json:"required_jobs"`
	TriggerModes           []string `json:"trigger_modes"`
	FixtureOnly            bool     `json:"fixture_only"`
	NoWorkflowMutation     bool     `json:"no_workflow_mutation"`
	ActivationAllowed      bool     `json:"activation_allowed"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5SentinelPromoterInputReadinessBinding struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	FixtureRef             string   `json:"fixture_ref"`
	PromoterInputs         []string `json:"promoter_inputs"`
	RequiredVerdicts       []string `json:"required_verdicts"`
	FixtureOnly            bool     `json:"fixture_only"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5PromoterBetaRollupNoActivationBinding struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	FixtureRef             string   `json:"fixture_ref"`
	RollupInputs           []string `json:"rollup_inputs"`
	AllowedDecisions       []string `json:"allowed_decisions"`
	ForbiddenActions       []string `json:"forbidden_actions"`
	NoPromotionRequested   bool     `json:"no_promotion_requested"`
	ActivationAllowed      bool     `json:"activation_allowed"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	SafeToExecute          bool     `json:"safe_to_execute"`
}

type month5CommandTimelineApprovalInboxBinding struct {
	Schema                 string   `json:"schema"`
	NodeID                 string   `json:"node_id"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	TimelineFixtureRef     string   `json:"timeline_fixture_ref"`
	TimelineSegments       []string `json:"timeline_segments"`
	ApprovalInboxStates    []string `json:"approval_inbox_states"`
	DisplayOnly            bool     `json:"display_only"`
	ApprovesWork           bool     `json:"approves_work"`
	MutatesMissionState    bool     `json:"mutates_mission_state"`
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

func TestMonth5ControlPlaneTransactionalEvidenceFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-13", "control-plane-transactional-evidence.json")
	fixture := mustLoadJSON[month5ControlPlaneTransactionalEvidenceFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.control-plane-transactional-evidence.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-13" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "transactional_transition_bound" ||
		fixture.StorageAuthority != "ao2-control-plane" {
		t.Fatalf("unexpected control-plane transactional fixture header: %#v", fixture)
	}
	for _, required := range []string{"prepare_transition", "write_content_addressed_blob", "update_index", "commit_manifest", "emit_readback"} {
		if !containsValue(fixture.RequiredAtomicSteps, required) {
			t.Fatalf("control-plane fixture missing atomic step %s: %#v", required, fixture.RequiredAtomicSteps)
		}
	}
	for _, required := range []string{"rollback_receipt", "previous_index_digest", "post_rollback_readback"} {
		if !containsValue(fixture.RollbackReadback, required) {
			t.Fatalf("control-plane fixture missing rollback readback %s: %#v", required, fixture.RollbackReadback)
		}
	}
	for _, rejected := range []string{"partial_write_visible", "unsigned_acceptance", "gc_race", "index_without_blob"} {
		if !containsValue(fixture.RejectedStates, rejected) {
			t.Fatalf("control-plane fixture missing rejected state %s: %#v", rejected, fixture.RejectedStates)
		}
	}
	if !fixture.ContentAddressedStore ||
		!fixture.IndexCommitAtomic ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("control-plane transactional fixture changed safety posture: %#v", fixture)
	}
}

func TestMonth5ControlPlaneMigrationMetadataFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-14", "control-plane-migration-metadata.json")
	fixture := mustLoadJSON[month5ControlPlaneMigrationMetadataFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.control-plane-migration-metadata.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-14" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "durable_beta_storage_metadata_bound" ||
		fixture.MigrationID == "" ||
		fixture.FromVersion == "" ||
		fixture.ToVersion == "" {
		t.Fatalf("unexpected control-plane migration metadata header: %#v", fixture)
	}
	if !digestPattern.MatchString(fixture.MetadataChecksum) {
		t.Fatalf("migration metadata checksum must be sha256-bound: %s", fixture.MetadataChecksum)
	}
	for _, required := range []string{"migration_id", "applied_at", "from_version", "to_version", "checksum", "rollback_plan"} {
		if !containsValue(fixture.RequiredMetadata, required) {
			t.Fatalf("control-plane migration fixture missing metadata %s: %#v", required, fixture.RequiredMetadata)
		}
	}
	for _, required := range []string{"forward_replay", "rollback_replay", "idempotent_reapply", "pre_migration_backup_readback"} {
		if !containsValue(fixture.ReplayChecks, required) {
			t.Fatalf("control-plane migration fixture missing replay check %s: %#v", required, fixture.ReplayChecks)
		}
	}
	if !fixture.NoDestructiveMigration ||
		!fixture.RollbackPlanBound ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("control-plane migration metadata fixture changed safety posture: %#v", fixture)
	}
}

func TestMonth5LocalBackupRestoreDrillFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-15", "local-backup-restore-drill.json")
	fixture := mustLoadJSON[month5LocalBackupRestoreDrillFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.local-backup-restore-drill.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-15" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "backup_restore_drill_bound" ||
		fixture.RestoreTarget != "isolated_local_readback_copy" {
		t.Fatalf("unexpected local backup restore drill header: %#v", fixture)
	}
	if !digestPattern.MatchString(fixture.BackupManifestDigest) {
		t.Fatalf("backup manifest digest must be sha256-bound: %s", fixture.BackupManifestDigest)
	}
	for _, required := range []string{"mission_ledger", "atlas_workgraph", "run_links", "checkpoint_readbacks"} {
		if !containsValue(fixture.RequiredArtifacts, required) {
			t.Fatalf("backup restore drill missing artifact %s: %#v", required, fixture.RequiredArtifacts)
		}
	}
	for _, required := range []string{"manifest_digest_match", "restored_readback_matches_source", "rollback_receipt_present", "no_source_state_mutation"} {
		if !containsValue(fixture.RestoreVerification, required) {
			t.Fatalf("backup restore drill missing verification %s: %#v", required, fixture.RestoreVerification)
		}
	}
	if !fixture.NoDestructiveRestore ||
		!fixture.DrillIsFixtureOnly ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("local backup restore drill changed safety posture: %#v", fixture)
	}
}

func TestMonth5MissionRestartReplayFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-16", "mission-restart-replay.json")
	fixture := mustLoadJSON[month5MissionRestartReplayFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.mission-restart-replay.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-16" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "restart_replay_bound" ||
		fixture.SourceCheckpoint == "" {
		t.Fatalf("unexpected Mission restart replay header: %#v", fixture)
	}
	for _, required := range []string{"mission_ledger", "recommendation_workgraph", "run_links", "checkpoint_readback"} {
		if !containsValue(fixture.ReplayInputs, required) {
			t.Fatalf("Mission restart replay missing input %s: %#v", required, fixture.ReplayInputs)
		}
	}
	for _, required := range []string{"completed_nodes_unchanged", "ready_nodes_unchanged", "exact_next_action_preserved", "final_response_remains_denied"} {
		if !containsValue(fixture.ExpectedReadbacks, required) {
			t.Fatalf("Mission restart replay missing readback %s: %#v", required, fixture.ExpectedReadbacks)
		}
	}
	if !fixture.ExactlyOnceAccounting ||
		fixture.DuplicateNodeCompletion ||
		fixture.FinalResponseAllowed ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Mission restart replay changed safety posture: %#v", fixture)
	}
}

func TestMonth5MissionKillRestartReplayFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-17", "mission-kill-restart-replay.json")
	fixture := mustLoadJSON[month5MissionKillRestartReplayFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.mission-kill-restart-replay.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-17" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "interrupted_node_resume_bound" ||
		fixture.InterruptedNode == "" ||
		fixture.ResumeCheckpoint == "" {
		t.Fatalf("unexpected Mission kill restart replay header: %#v", fixture)
	}
	for _, required := range []string{"process_exit_before_run_link", "dirty_node_evidence_detected", "resume_from_last_completed_checkpoint"} {
		if !containsValue(fixture.InterruptionMarkers, required) {
			t.Fatalf("Mission kill restart replay missing marker %s: %#v", required, fixture.InterruptionMarkers)
		}
	}
	for _, required := range []string{"partial_node_not_completed", "same_node_reselected", "run_link_required_before_advance", "final_response_remains_denied"} {
		if !containsValue(fixture.ResumeAssertions, required) {
			t.Fatalf("Mission kill restart replay missing assertion %s: %#v", required, fixture.ResumeAssertions)
		}
	}
	if fixture.PartialNodePromoted ||
		fixture.DuplicateRunLink ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Mission kill restart replay changed safety posture: %#v", fixture)
	}
}

func TestMonth5GoldenPathDryRunReadinessFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-18", "golden-path-dry-run-readiness.json")
	fixture := mustLoadJSON[month5GoldenPathDryRunReadinessFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.golden-path-dry-run-readiness.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-18" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "dry_run_readiness_matrix_bound" {
		t.Fatalf("unexpected golden path readiness header: %#v", fixture)
	}
	rows := map[string]month5GoldenPathReadinessRow{}
	for _, row := range fixture.MatrixRows {
		rows[row.Component] = row
		if row.Handoff == "" || row.Check == "" || row.Status != "ready_for_dry_run" {
			t.Fatalf("golden path row must include handoff, check, and ready status: %#v", row)
		}
	}
	for _, required := range []string{"ao-mission", "ao-blueprint", "ao-atlas", "ao-foundry", "ao-forge", "ao-covenant", "ao2", "ao2-control-plane", "ao-command"} {
		if _, ok := rows[required]; !ok {
			t.Fatalf("golden path readiness missing component %s", required)
		}
	}
	if !fixture.DryRunOnly ||
		!fixture.NoProviderExecution ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("golden path readiness changed safety posture: %#v", fixture)
	}
}

func TestMonth5CleanRoomNonAOReplayFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-19", "clean-room-non-ao-replay.json")
	fixture := mustLoadJSON[month5CleanRoomNonAOReplayFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.clean-room-non-ao-replay.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-19" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "external_repository_preflight_bound" ||
		fixture.TargetRepoClass != "non_ao_clean_room" ||
		!fixture.IsolatedWorktree {
		t.Fatalf("unexpected clean-room replay header: %#v", fixture)
	}
	for _, required := range []string{"objective_packet", "dry_run_workgraph", "rollback_plan", "operator_review_gate"} {
		if !containsValue(fixture.ReplayInputs, required) {
			t.Fatalf("clean-room replay missing input %s: %#v", required, fixture.ReplayInputs)
		}
	}
	for _, required := range []string{"no_repo_write_without_approval", "no_credentials", "no_provider_calls", "no_release_or_publish"} {
		if !containsValue(fixture.ExternalRepoBoundaries, required) {
			t.Fatalf("clean-room replay missing boundary %s: %#v", required, fixture.ExternalRepoBoundaries)
		}
	}
	if !fixture.NoExternalMutation ||
		!fixture.NoProviderExecution ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("clean-room replay changed safety posture: %#v", fixture)
	}
}

func TestMonth5ArenaHostedCIWorkflowFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-20", "arena-hosted-ci-workflow.json")
	fixture := mustLoadJSON[month5ArenaHostedCIWorkflowFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.arena-hosted-ci-workflow.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-20" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "hosted_ci_workflow_fixture_bound" ||
		fixture.Repository != "ao-arena" ||
		fixture.WorkflowPath != ".github/workflows/production-readiness.yml" {
		t.Fatalf("unexpected Arena hosted CI workflow header: %#v", fixture)
	}
	for _, required := range []string{"go_test", "go_vet", "fixture_validation", "public_safety_scan"} {
		if !containsValue(fixture.RequiredJobs, required) {
			t.Fatalf("Arena hosted CI workflow missing job %s: %#v", required, fixture.RequiredJobs)
		}
	}
	for _, required := range []string{"pull_request", "push_main", "workflow_dispatch"} {
		if !containsValue(fixture.TriggerModes, required) {
			t.Fatalf("Arena hosted CI workflow missing trigger %s: %#v", required, fixture.TriggerModes)
		}
	}
	if !fixture.FixtureOnly ||
		!fixture.NoWorkflowMutation ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Arena hosted CI workflow changed safety posture: %#v", fixture)
	}
}

func TestMonth5CrucibleHostedCIWorkflowFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-21", "crucible-hosted-ci-workflow.json")
	fixture := mustLoadJSON[month5CrucibleHostedCIWorkflowFixture](t, fixturePath)

	if fixture.Schema != "ao.atlas.month5.crucible-hosted-ci-workflow.v0.1" ||
		fixture.NodeID != "mission-recommendation-month5-beta-operations-21" ||
		fixture.MissionID != "mission-4d91b0a9e4ab273e" ||
		fixture.Status != "hosted_ci_failure_injection_fixture_bound" ||
		fixture.Repository != "ao-crucible" ||
		fixture.WorkflowPath != ".github/workflows/production-readiness.yml" {
		t.Fatalf("unexpected Crucible hosted CI workflow header: %#v", fixture)
	}
	for _, required := range []string{"go_test", "go_vet", "fixture_validation", "failure_injection_fixture_validation", "public_safety_scan"} {
		if !containsValue(fixture.RequiredJobs, required) {
			t.Fatalf("Crucible hosted CI workflow missing job %s: %#v", required, fixture.RequiredJobs)
		}
	}
	for _, required := range []string{"pull_request", "push_main", "workflow_dispatch"} {
		if !containsValue(fixture.TriggerModes, required) {
			t.Fatalf("Crucible hosted CI workflow missing trigger %s: %#v", required, fixture.TriggerModes)
		}
	}
	for _, required := range []string{"adversarial_probe_fixture", "failure_injection_replay", "crash_only_recovery_readback"} {
		if !containsValue(fixture.FailureInjectionReadbacks, required) {
			t.Fatalf("Crucible hosted CI workflow missing failure readback %s: %#v", required, fixture.FailureInjectionReadbacks)
		}
	}
	if !fixture.FixtureOnly ||
		!fixture.NoWorkflowMutation ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute {
		t.Fatalf("Crucible hosted CI workflow changed safety posture: %#v", fixture)
	}
}

func TestMonth5SentinelHostedCIWorkflowFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-22")
	fixturePath := filepath.Join(nodeDir, "sentinel-hosted-ci-workflow-fixture.json")
	fixture := mustLoadJSON[AtlasSentinelHostedCIWorkflowFixture](t, fixturePath)
	if err := ValidateAtlasSentinelHostedCIWorkflowFixture(fixture); err != nil {
		t.Fatalf("Sentinel hosted CI fixture is invalid: %v", err)
	}
	if fixture.WorkflowPath != ".github/workflows/sentinel-fixture-verification.yml" ||
		fixture.Permissions != "contents:read" ||
		fixture.UsesProviderCredentials ||
		fixture.UsesSecrets ||
		fixture.TriggersRelease ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("Sentinel hosted CI fixture changed safety posture: %#v", fixture)
	}

	bindingPath := filepath.Join(nodeDir, "sentinel-native-signal-readback-binding.json")
	binding := mustLoadJSON[month5SentinelHostedCINativeSignalBinding](t, bindingPath)
	if binding.Schema != "ao.atlas.month5.sentinel-hosted-ci-native-signal-binding.v0.1" ||
		binding.NodeID != "mission-recommendation-month5-beta-operations-22" ||
		binding.MissionID != "mission-4d91b0a9e4ab273e" ||
		binding.Status != "hosted_ci_native_signal_fixture_bound" ||
		binding.Repository != "ao-sentinel" ||
		binding.FixtureRef != "sentinel-hosted-ci-workflow-fixture.json" {
		t.Fatalf("unexpected Sentinel hosted CI binding header: %#v", binding)
	}
	for _, required := range []string{"ci_signal", "public_safety_signal", "evidence_freshness_signal", "policy_signal"} {
		if !containsValue(binding.NativeSignalReadbacks, required) {
			t.Fatalf("Sentinel hosted CI binding missing signal %s: %#v", required, binding.NativeSignalReadbacks)
		}
	}
	if !binding.FixtureOnly ||
		!binding.NoWorkflowMutation ||
		!binding.NoPromotionRequested ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute {
		t.Fatalf("Sentinel hosted CI binding changed safety posture: %#v", binding)
	}
}

func TestMonth5PromoterHostedCIWorkflowFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-23")
	fixturePath := filepath.Join(nodeDir, "promoter-no-activation-boundary-fixture.json")
	fixture := mustLoadJSON[AtlasPromoterNoActivationBoundaryFixture](t, fixturePath)
	if err := ValidateAtlasPromoterNoActivationBoundaryFixture(fixture); err != nil {
		t.Fatalf("Promoter no-activation fixture is invalid: %v", err)
	}
	if fixture.Decision != "no_promotion" ||
		fixture.ActivationExecutionOwned ||
		fixture.ReleaseExecutionOwned ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("Promoter no-activation fixture changed safety posture: %#v", fixture)
	}

	bindingPath := filepath.Join(nodeDir, "promoter-hosted-ci-no-activation-binding.json")
	binding := mustLoadJSON[month5PromoterHostedCINoActivationBinding](t, bindingPath)
	if binding.Schema != "ao.atlas.month5.promoter-hosted-ci-no-activation-binding.v0.1" ||
		binding.NodeID != "mission-recommendation-month5-beta-operations-23" ||
		binding.MissionID != "mission-4d91b0a9e4ab273e" ||
		binding.Status != "hosted_ci_no_activation_fixture_bound" ||
		binding.Repository != "ao-promoter" ||
		binding.FixtureRef != "promoter-no-activation-boundary-fixture.json" ||
		binding.WorkflowPath != ".github/workflows/production-readiness.yml" {
		t.Fatalf("unexpected Promoter hosted CI binding header: %#v", binding)
	}
	for _, required := range []string{"go_test", "go_vet", "no_activation_fixture_validation", "public_safety_scan"} {
		if !containsValue(binding.RequiredJobs, required) {
			t.Fatalf("Promoter hosted CI binding missing job %s: %#v", required, binding.RequiredJobs)
		}
	}
	for _, required := range []string{"pull_request", "push_main", "workflow_dispatch"} {
		if !containsValue(binding.TriggerModes, required) {
			t.Fatalf("Promoter hosted CI binding missing trigger %s: %#v", required, binding.TriggerModes)
		}
	}
	if !binding.FixtureOnly ||
		!binding.NoWorkflowMutation ||
		binding.ActivationAllowed ||
		!binding.NoPromotionRequested ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute {
		t.Fatalf("Promoter hosted CI binding changed safety posture: %#v", binding)
	}
}

func TestMonth5SentinelSignalStatePromoterInputReadinessFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-24")
	fixturePath := filepath.Join(nodeDir, "sentinel-signal-state-fixture.json")
	fixture := mustLoadJSON[AtlasSentinelSignalStateFixture](t, fixturePath)
	if err := ValidateAtlasSentinelSignalStateFixture(fixture); err != nil {
		t.Fatalf("Sentinel signal state fixture is invalid: %v", err)
	}
	if fixture.SignalCount != 4 ||
		fixture.StateCount != 4 ||
		fixture.MatrixCount != 16 ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("Sentinel signal state fixture changed safety posture: %#v", fixture)
	}

	bindingPath := filepath.Join(nodeDir, "sentinel-promoter-input-readiness-binding.json")
	binding := mustLoadJSON[month5SentinelPromoterInputReadinessBinding](t, bindingPath)
	if binding.Schema != "ao.atlas.month5.sentinel-promoter-input-readiness-binding.v0.1" ||
		binding.NodeID != "mission-recommendation-month5-beta-operations-24" ||
		binding.MissionID != "mission-4d91b0a9e4ab273e" ||
		binding.Status != "promoter_input_readiness_bound" ||
		binding.FixtureRef != "sentinel-signal-state-fixture.json" {
		t.Fatalf("unexpected Sentinel/Promoter input readiness binding header: %#v", binding)
	}
	for _, required := range []string{"ci", "runtime", "policy", "evidence_freshness"} {
		if !containsValue(binding.PromoterInputs, required) {
			t.Fatalf("Sentinel/Promoter binding missing input %s: %#v", required, binding.PromoterInputs)
		}
	}
	for _, required := range []string{"allow_continue", "wait", "refresh_required", "block"} {
		if !containsValue(binding.RequiredVerdicts, required) {
			t.Fatalf("Sentinel/Promoter binding missing verdict %s: %#v", required, binding.RequiredVerdicts)
		}
	}
	if !binding.FixtureOnly ||
		!binding.NoPromotionRequested ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute {
		t.Fatalf("Sentinel/Promoter input readiness binding changed safety posture: %#v", binding)
	}
}

func TestMonth5PromoterNoActivationBoundaryRollupFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-25")
	fixturePath := filepath.Join(nodeDir, "promoter-no-activation-boundary-fixture.json")
	fixture := mustLoadJSON[AtlasPromoterNoActivationBoundaryFixture](t, fixturePath)
	if err := ValidateAtlasPromoterNoActivationBoundaryFixture(fixture); err != nil {
		t.Fatalf("Promoter no-activation rollup fixture is invalid: %v", err)
	}
	if fixture.Decision != "no_promotion" ||
		fixture.ActivationExecutionOwned ||
		fixture.ReleaseExecutionOwned ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("Promoter no-activation rollup fixture changed safety posture: %#v", fixture)
	}

	bindingPath := filepath.Join(nodeDir, "promoter-beta-rollup-no-activation-binding.json")
	binding := mustLoadJSON[month5PromoterBetaRollupNoActivationBinding](t, bindingPath)
	if binding.Schema != "ao.atlas.month5.promoter-beta-rollup-no-activation-binding.v0.1" ||
		binding.NodeID != "mission-recommendation-month5-beta-operations-25" ||
		binding.MissionID != "mission-4d91b0a9e4ab273e" ||
		binding.Status != "beta_rollup_no_activation_bound" ||
		binding.FixtureRef != "promoter-no-activation-boundary-fixture.json" {
		t.Fatalf("unexpected Promoter beta rollup no-activation binding header: %#v", binding)
	}
	for _, required := range []string{"sentinel_signal_state", "command_readback", "foundry_rollup", "public_safety_scan"} {
		if !containsValue(binding.RollupInputs, required) {
			t.Fatalf("Promoter beta rollup binding missing input %s: %#v", required, binding.RollupInputs)
		}
	}
	for _, required := range []string{"no_promotion", "blocked", "insufficient_evidence"} {
		if !containsValue(binding.AllowedDecisions, required) {
			t.Fatalf("Promoter beta rollup binding missing decision %s: %#v", required, binding.AllowedDecisions)
		}
	}
	for _, forbidden := range []string{"activate", "release", "deploy", "publish", "tag"} {
		if !containsValue(binding.ForbiddenActions, forbidden) {
			t.Fatalf("Promoter beta rollup binding missing forbidden action %s: %#v", forbidden, binding.ForbiddenActions)
		}
	}
	if !binding.NoPromotionRequested ||
		binding.ActivationAllowed ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute {
		t.Fatalf("Promoter beta rollup binding changed safety posture: %#v", binding)
	}
}

func TestMonth5CommandTimelineApprovalInboxFixture(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-26")
	fixturePath := filepath.Join(nodeDir, "compact-timeline-filter-fixture.json")
	fixture := mustLoadJSON[AtlasCompactTimelineFilterFixture](t, fixturePath)
	if err := ValidateAtlasCompactTimelineFilterFixture(fixture); err != nil {
		t.Fatalf("compact timeline fixture is invalid: %v", err)
	}
	if fixture.FilterCount != 5 ||
		!fixture.StaleRecordsDistinguished ||
		!fixture.DuplicateRecordsDistinguished ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("compact timeline fixture changed safety posture: %#v", fixture)
	}

	bindingPath := filepath.Join(nodeDir, "command-timeline-approval-inbox-binding.json")
	binding := mustLoadJSON[month5CommandTimelineApprovalInboxBinding](t, bindingPath)
	if binding.Schema != "ao.atlas.month5.command-timeline-approval-inbox-binding.v0.1" ||
		binding.NodeID != "mission-recommendation-month5-beta-operations-26" ||
		binding.MissionID != "mission-4d91b0a9e4ab273e" ||
		binding.Status != "compact_timeline_approval_inbox_bound" ||
		binding.TimelineFixtureRef != "compact-timeline-filter-fixture.json" {
		t.Fatalf("unexpected Command timeline approval inbox binding header: %#v", binding)
	}
	for _, required := range []string{"checkpoint", "exact_next_action", "return_gate", "lease_health", "node_counts"} {
		if !containsValue(binding.TimelineSegments, required) {
			t.Fatalf("Command timeline binding missing segment %s: %#v", required, binding.TimelineSegments)
		}
	}
	for _, required := range []string{"pending_review", "approved_elsewhere", "blocked", "not_requested"} {
		if !containsValue(binding.ApprovalInboxStates, required) {
			t.Fatalf("Command timeline binding missing inbox state %s: %#v", required, binding.ApprovalInboxStates)
		}
	}
	if !binding.DisplayOnly ||
		binding.ApprovesWork ||
		binding.MutatesMissionState ||
		!binding.NoPromotionRequested ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied ||
		binding.SafeToExecute {
		t.Fatalf("Command timeline approval inbox binding changed safety posture: %#v", binding)
	}
}

func TestMonth5DeterministicRunProvenanceProviderModelFixture(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month5-beta-operations-v01", "nodes", "mission-recommendation-month5-beta-operations-27", "deterministic-run-provenance.json")
	fixture := mustLoadJSON[AtlasMonth3ProviderModelProvenance](t, fixturePath)
	if err := ValidateAtlasMonth3ProviderModelProvenance(fixture); err != nil {
		t.Fatalf("provider/model provenance fixture is invalid: %v", err)
	}
	if fixture.NodeID != "mission-recommendation-month5-beta-operations-27" ||
		fixture.Status != "provider_model_provenance_ready" ||
		fixture.RunRecordCount != 4 ||
		!fixture.EveryRunHasProvider ||
		!fixture.EveryRunHasModel ||
		!fixture.EveryRunHasModelClass ||
		fixture.LiveProviderCallCount != 0 ||
		fixture.FinalResponseAllowed ||
		fixture.SchedulesWork ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("provider/model provenance fixture changed safety posture: %#v", fixture)
	}
	for _, record := range fixture.RunRecords {
		if record.Provider == "" || record.Model == "" || record.ModelClass == "" || record.LiveProviderCall {
			t.Fatalf("run record missing provenance or records a live provider call: %#v", record)
		}
	}
}
