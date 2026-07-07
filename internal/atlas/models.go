package atlas

type Instance struct {
	ContractVersion string            `json:"contract_version"`
	ID              string            `json:"id"`
	StateRoot       string            `json:"state_root"`
	ToolchainRoot   string            `json:"toolchain_root"`
	Roots           map[string]string `json:"roots"`
}

type AtlasRegistry struct {
	ContractVersion string            `json:"contract_version"`
	InstanceID      string            `json:"instance_id"`
	ToolchainRoot   string            `json:"toolchain_root"`
	Roots           map[string]string `json:"roots"`
	SchedulesWork   bool              `json:"schedules_work"`
	ExecutesWork    bool              `json:"executes_work"`
	ApprovesWork    bool              `json:"approves_work"`
}

type InstanceDoctorReport struct {
	ContractVersion        string            `json:"contract_version"`
	InstanceID             string            `json:"instance_id"`
	Status                 string            `json:"status"`
	Checks                 map[string]string `json:"checks"`
	FirstFailingCheck      string            `json:"first_failing_check"`
	BlockingNextActions    []string          `json:"blocking_next_actions"`
	MaintenanceSuggestions []string          `json:"maintenance_suggestions"`
	SchedulesWork          bool              `json:"schedules_work"`
	ExecutesWork           bool              `json:"executes_work"`
	ApprovesWork           bool              `json:"approves_work"`
}

type Intake struct {
	ContractVersion string   `json:"contract_version"`
	ID              string   `json:"id"`
	TargetInstance  string   `json:"target_instance"`
	BroadPrompt     string   `json:"broad_prompt"`
	InstructionRefs []string `json:"instruction_refs"`
	FolderRefs      []string `json:"folder_refs"`
	Constraints     []string `json:"constraints"`
}

type MissionStatus struct {
	ContractVersion          string                         `json:"contract_version"`
	IntakeID                 string                         `json:"intake_id"`
	WorkgraphID              string                         `json:"workgraph_id"`
	TargetInstance           string                         `json:"target_instance"`
	CompletionStatus         string                         `json:"completion_status"`
	NodeCounts               map[string]int                 `json:"node_counts"`
	RunLinks                 map[string]string              `json:"run_links"`
	MissingContextPacks      []string                       `json:"missing_context_packs"`
	MissingHandoffs          []string                       `json:"missing_handoffs"`
	NextRecommendedAction    string                         `json:"next_recommended_action"`
	NextActions              []string                       `json:"next_actions"`
	AuthorityLadder          *AuthorityLadderStatus         `json:"authority_ladder,omitempty"`
	FinalResponseAllowed     bool                           `json:"final_response_allowed"`
	FinalResponseReason      string                         `json:"final_response_reason"`
	FinalStateReconciliation *AtlasFinalStateReconciliation `json:"final_state_reconciliation,omitempty"`
	SchedulesWork            bool                           `json:"schedules_work"`
	ExecutesWork             bool                           `json:"executes_work"`
}

type AtlasFinalStateReconciliation struct {
	ContractVersion       string `json:"contract_version"`
	Status                string `json:"status"`
	WorkgraphStatus       string `json:"workgraph_status"`
	FoundryRollupStatus   string `json:"foundry_rollup_status"`
	PromoterVerdictStatus string `json:"promoter_verdict_status"`
	CommandReadbackStatus string `json:"command_readback_status"`
	ExactNextAction       string `json:"exact_next_action"`
	ContinuationReason    string `json:"continuation_contract_reason,omitempty"`
	ContinuationAgreement bool   `json:"continuation_reason_agreement,omitempty"`
	SchedulesWork         bool   `json:"schedules_work"`
	ExecutesWork          bool   `json:"executes_work"`
	ApprovesWork          bool   `json:"approves_work"`
}

type AOMissionImport struct {
	ContractVersion string                    `json:"contract_version"`
	MissionID       string                    `json:"mission_id"`
	Status          string                    `json:"status"`
	CurrentRoute    string                    `json:"current_route"`
	SourceArtifacts []AOMissionSourceArtifact `json:"source_artifacts"`
	NextAction      string                    `json:"next_action"`
	SafeToExecute   bool                      `json:"safe_to_execute"`
	SchedulesWork   bool                      `json:"schedules_work"`
	ExecutesWork    bool                      `json:"executes_work"`
	ApprovesWork    bool                      `json:"approves_work"`
}

type AOMissionWorkgraphMetadata struct {
	ContractVersion          string            `json:"contract_version"`
	MissionID                string            `json:"mission_id"`
	WorkgraphID              string            `json:"workgraph_id"`
	TargetInstance           string            `json:"target_instance"`
	CurrentRoute             string            `json:"current_route"`
	NodeCounts               map[string]int    `json:"node_counts"`
	MissionProvenance        map[string]int    `json:"mission_provenance"`
	ProvenanceNodes          []string          `json:"provenance_nodes,omitempty"`
	PrimaryMissionProvenance string            `json:"primary_mission_provenance"`
	ProvenanceDiagnostics    string            `json:"provenance_diagnostics"`
	SourceArtifacts          map[string]string `json:"source_artifacts"`
	NextAction               string            `json:"next_action"`
	SafeToExecute            bool              `json:"safe_to_execute"`
	SchedulesWork            bool              `json:"schedules_work"`
	ExecutesWork             bool              `json:"executes_work"`
	ApprovesWork             bool              `json:"approves_work"`
}

type AOMissionProvenanceRender struct {
	ContractVersion          string         `json:"contract_version"`
	Status                   string         `json:"status"`
	MissionID                string         `json:"mission_id"`
	WorkgraphID              string         `json:"workgraph_id"`
	PrimaryMissionProvenance string         `json:"primary_mission_provenance"`
	TotalProvenanceSources   int            `json:"total_provenance_sources"`
	ProvenanceSummary        string         `json:"provenance_summary"`
	ProvenanceNodes          []string       `json:"provenance_nodes"`
	MissionProvenance        map[string]int `json:"mission_provenance"`
	NextAction               string         `json:"next_action"`
	SafeToExecute            bool           `json:"safe_to_execute"`
	SchedulesWork            bool           `json:"schedules_work"`
	ExecutesWork             bool           `json:"executes_work"`
	ApprovesWork             bool           `json:"approves_work"`
}

type AOMissionFinalSynthesis struct {
	Schema                                string                                  `json:"schema"`
	Mission                               string                                  `json:"mission"`
	Status                                string                                  `json:"status"`
	CompletedNodes                        int                                     `json:"completed_nodes"`
	ReadyNodes                            int                                     `json:"ready_nodes"`
	BlockedNodes                          int                                     `json:"blocked_nodes"`
	MinimumNodes                          int                                     `json:"minimum_nodes"`
	TargetMinutes                         int                                     `json:"target_minutes"`
	MaxMinutes                            int                                     `json:"max_minutes"`
	FinalResponseAllowed                  bool                                    `json:"final_response_allowed"`
	AtlasWorkgraphStatus                  string                                  `json:"atlas_workgraph_status"`
	FoundryRollup                         string                                  `json:"foundry_rollup"`
	PromoterStatus                        string                                  `json:"promoter_status"`
	CommandReadback                       string                                  `json:"command_readback"`
	EventSearchBound                      bool                                    `json:"event_search_bound"`
	BranchCleanupBoundThroughPreviousNode bool                                    `json:"branch_cleanup_bound_through_previous_node"`
	MergedPRsFinal                        []int                                   `json:"merged_prs_final"`
	CurrentNodeBranch                     string                                  `json:"current_node_branch"`
	CurrentNodePRPending                  bool                                    `json:"current_node_pr_pending"`
	PromotionClaimed                      bool                                    `json:"promotion_claimed"`
	ClaimsAuthorityAdvance                bool                                    `json:"claims_authority_advance"`
	RSIRemainsDenied                      bool                                    `json:"rsi_remains_denied"`
	SafeToExecute                         bool                                    `json:"safe_to_execute"`
	ExecutesWork                          bool                                    `json:"executes_work"`
	ApprovesWork                          bool                                    `json:"approves_work"`
	MutatesRepositories                   bool                                    `json:"mutates_repositories"`
	FeatureDepthRecommendations           []AOMissionFinalSynthesisRecommendation `json:"feature_depth_recommendations"`
	ExactNextAction                       string                                  `json:"exact_next_action"`
}

type AOMissionFinalSynthesisRecommendation struct {
	ID              string `json:"id"`
	Owner           string `json:"owner"`
	Task            string `json:"task"`
	ExactNextAction string `json:"exact_next_action"`
}

type AOMissionFinalSynthesisReadback struct {
	ContractVersion        string   `json:"contract_version"`
	MissionID              string   `json:"mission_id"`
	Status                 string   `json:"status"`
	SourceDigest           string   `json:"source_digest"`
	TotalNodes             int      `json:"total_nodes"`
	CompletedNodes         int      `json:"completed_nodes"`
	ReadyNodes             int      `json:"ready_nodes"`
	BlockedNodes           int      `json:"blocked_nodes"`
	MinimumNodes           int      `json:"minimum_nodes"`
	TargetMinutes          int      `json:"target_minutes"`
	MaxMinutes             int      `json:"max_minutes"`
	ReturnGateStatus       string   `json:"return_gate_status"`
	FinalResponseAllowed   bool     `json:"final_response_allowed"`
	FinalResponseReason    string   `json:"final_response_reason"`
	AtlasWorkgraphStatus   string   `json:"atlas_workgraph_status"`
	FoundryRollup          string   `json:"foundry_rollup"`
	PromoterStatus         string   `json:"promoter_status"`
	CommandReadback        string   `json:"command_readback"`
	EventSearchBound       bool     `json:"event_search_bound"`
	BranchCleanupBound     bool     `json:"branch_cleanup_bound"`
	MergedPRsFinal         []int    `json:"merged_prs_final"`
	ExactNextAction        string   `json:"exact_next_action"`
	FeatureDepthNextTasks  []string `json:"feature_depth_next_tasks"`
	RSIRemainsDenied       bool     `json:"rsi_remains_denied"`
	PromotionClaimed       bool     `json:"promotion_claimed"`
	ClaimsAuthorityAdvance bool     `json:"claims_authority_advance"`
	SafeToExecute          bool     `json:"safe_to_execute"`
	SchedulesWork          bool     `json:"schedules_work"`
	ExecutesWork           bool     `json:"executes_work"`
	ApprovesWork           bool     `json:"approves_work"`
	MutatesRepositories    bool     `json:"mutates_repositories"`
}

type AOMissionFeatureDepthRecommendations struct {
	Schema              string                      `json:"schema"`
	MissionID           string                      `json:"mission_id"`
	Status              string                      `json:"status"`
	MinimumTasks        int                         `json:"minimum_tasks"`
	RecommendationCount int                         `json:"recommendation_count"`
	SourceEvidenceRoot  string                      `json:"source_evidence_root,omitempty"`
	SourceReadbackPath  string                      `json:"source_readback_path,omitempty"`
	SourceAssertionPath string                      `json:"source_assertion_path,omitempty"`
	Tasks               []AOMissionFeatureDepthTask `json:"tasks"`
	SafeToExecute       bool                        `json:"safe_to_execute"`
	SchedulesWork       bool                        `json:"schedules_work,omitempty"`
	ExecutesWork        bool                        `json:"executes_work"`
	ApprovesWork        bool                        `json:"approves_work"`
	MutatesRepositories bool                        `json:"mutates_repositories,omitempty"`
}

type AOMissionFeatureDepthTask struct {
	Rank         int      `json:"rank,omitempty"`
	ID           string   `json:"id"`
	Owner        string   `json:"owner"`
	Theme        string   `json:"theme,omitempty"`
	Task         string   `json:"task"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

type AtlasNextWaveRecommendationExport struct {
	Schema                 string                               `json:"schema"`
	NodeID                 string                               `json:"node_id"`
	Status                 string                               `json:"status"`
	SourceEvidenceRoot     string                               `json:"source_evidence_root"`
	SourceReadbackPath     string                               `json:"source_readback_path"`
	SourceAssertionPath    string                               `json:"source_assertion_path"`
	CompletedNodesBefore   int                                  `json:"completed_nodes_before_export"`
	ReadyNodesBefore       int                                  `json:"ready_nodes_before_export"`
	ExpectedNextNode       string                               `json:"expected_next_node_after_completion"`
	MinimumRankedTasks     int                                  `json:"minimum_ranked_tasks"`
	RecommendationCount    int                                  `json:"recommendation_count"`
	RankedTaskFloorMet     bool                                 `json:"ranked_task_floor_met"`
	NoPromotionRequested   bool                                 `json:"no_promotion_requested"`
	PromotionGranted       bool                                 `json:"promotion_granted"`
	ClaimsAuthorityAdvance bool                                 `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                                 `json:"rsi_remains_denied"`
	FeatureDepthExport     AOMissionFeatureDepthRecommendations `json:"feature_depth_export"`
}

type AtlasRecommendationWave struct {
	ContractVersion        string                    `json:"contract_version"`
	MissionID              string                    `json:"mission_id"`
	TargetInstance         string                    `json:"target_instance"`
	Status                 string                    `json:"status"`
	SourceDigest           string                    `json:"source_digest"`
	MinimumTasks           int                       `json:"minimum_tasks"`
	TotalTasks             int                       `json:"total_tasks"`
	NodeBudget             int                       `json:"node_budget"`
	EstimatedMinutes       int                       `json:"estimated_minutes"`
	Supervisor             *AtlasLongRunSupervisor   `json:"supervisor,omitempty"`
	Tasks                  []AtlasRecommendationTask `json:"tasks"`
	NextRecommendedPrompt  string                    `json:"next_recommended_prompt"`
	FinalResponseAllowed   bool                      `json:"final_response_allowed"`
	FinalResponseReason    string                    `json:"final_response_reason"`
	PromoterReadbackStatus string                    `json:"promoter_readback_status"`
	CommandReadbackStatus  string                    `json:"command_readback_status"`
	PublicSafetyScanStatus string                    `json:"public_safety_scan_status"`
	SafeToExecute          bool                      `json:"safe_to_execute"`
	SchedulesWork          bool                      `json:"schedules_work"`
	ExecutesWork           bool                      `json:"executes_work"`
	ApprovesWork           bool                      `json:"approves_work"`
}

type AtlasLongRunSupervisor struct {
	ContractVersion      string `json:"contract_version"`
	MinNodes             int    `json:"min_nodes"`
	MinMinutes           int    `json:"min_minutes"`
	MaxMinutes           int    `json:"max_minutes"`
	ContinueIfFastTarget int    `json:"continue_if_fast_target"`
	ReturnOnlyWhen       string `json:"return_only_when"`
	CheckpointPolicy     string `json:"checkpoint_policy"`
	EvidencePolicy       string `json:"evidence_policy"`
	FinalReportContract  string `json:"final_report_contract"`
}

type AtlasRecommendationTask struct {
	ID                string   `json:"id"`
	Owner             string   `json:"owner"`
	Task              string   `json:"task"`
	NodeID            string   `json:"node_id"`
	TaskID            string   `json:"task_id"`
	MutationClass     string   `json:"mutation_class"`
	SourceTaskDigest  string   `json:"source_task_digest"`
	TargetFactoryRepo string   `json:"target_factory_repo"`
	FactoryFolder     string   `json:"factory_folder"`
	RequiredGates     []string `json:"required_gates"`
	Verification      []string `json:"verification_commands"`
	SafetyLimits      []string `json:"safety_limits"`
}

type AtlasRecommendationReadback struct {
	ContractVersion                 string                                `json:"contract_version"`
	MissionID                       string                                `json:"mission_id"`
	TargetInstance                  string                                `json:"target_instance"`
	Status                          string                                `json:"status"`
	SourceDigest                    string                                `json:"source_digest"`
	WaveDigest                      string                                `json:"wave_digest,omitempty"`
	WorkgraphDigest                 string                                `json:"workgraph_digest,omitempty"`
	EvidenceRoot                    string                                `json:"evidence_root,omitempty"`
	Supervisor                      *AtlasLongRunSupervisor               `json:"supervisor,omitempty"`
	StartedAt                       string                                `json:"started_at,omitempty"`
	CompletedAt                     string                                `json:"completed_at,omitempty"`
	ElapsedMinutes                  int                                   `json:"elapsed_minutes"`
	MinMinutesMet                   bool                                  `json:"min_minutes_met"`
	LeaseTimeStatus                 string                                `json:"lease_time_status"`
	TotalNodes                      int                                   `json:"total_nodes"`
	MinimumNodes                    int                                   `json:"minimum_nodes"`
	CompletedNodes                  int                                   `json:"completed_nodes"`
	ReadyNodes                      int                                   `json:"ready_nodes"`
	BlockedNodes                    int                                   `json:"blocked_nodes"`
	FailedNodes                     int                                   `json:"failed_nodes"`
	ExecutableReadyNodes            int                                   `json:"executable_ready_nodes"`
	FirstExecutableNode             string                                `json:"first_executable_node,omitempty"`
	LeaseHealthStatus               string                                `json:"lease_health_status"`
	CheckpointFreshnessStatus       string                                `json:"checkpoint_freshness_status"`
	StaleRouteDecisionStatus        string                                `json:"stale_route_decision_status"`
	EarlyReturnRiskStatus           string                                `json:"early_return_risk_status"`
	FoundryRollupStatus             string                                `json:"foundry_rollup_status"`
	FoundryTerminalStatusReadback   map[string]string                     `json:"foundry_terminal_status_readback"`
	FoundryTerminalStatusExamples   []AtlasFoundryTerminalStatusExample   `json:"foundry_terminal_status_examples"`
	FoundryDeniedTerminalExamples   []AtlasFoundryDeniedTerminalExample   `json:"foundry_denied_terminal_examples"`
	PromoterReadbackStatus          string                                `json:"promoter_readback_status"`
	PromoterNoPromotionStatus       string                                `json:"promoter_no_promotion_status"`
	PromoterNoPromotionPlaceholders []AtlasPromoterNoPromotionPlaceholder `json:"promoter_no_promotion_placeholders"`
	CommandReadbackStatus           string                                `json:"command_readback_status"`
	CommandTimelineStatus           string                                `json:"command_timeline_status"`
	CommandTimelinePlaceholders     []AtlasCommandTimelinePlaceholder     `json:"command_timeline_placeholders"`
	PublicSafetyScanStatus          string                                `json:"public_safety_scan_status"`
	ReturnGateStatus                string                                `json:"return_gate_status,omitempty"`
	CheckpointCount                 int                                   `json:"checkpoint_count"`
	FinalResponseAllowed            bool                                  `json:"final_response_allowed"`
	FinalResponseDenialGate         string                                `json:"final_response_denial_gate"`
	FinalResponseReason             string                                `json:"final_response_reason"`
	ExactNextAction                 string                                `json:"exact_next_action"`
	ContinuationContract            AtlasContinuationContract             `json:"continuation_contract"`
	ExactNextActionReadback         AtlasExactNextActionReadback          `json:"exact_next_action_readback"`
	NodeEvidence                    []AtlasRecommendationNodeEvidence     `json:"node_evidence"`
	FeatureDepthRecommendations     []string                              `json:"feature_depth_recommendations"`
	SafetyBoundaries                map[string]bool                       `json:"safety_boundaries"`
	SchedulesWork                   bool                                  `json:"schedules_work"`
	ExecutesWork                    bool                                  `json:"executes_work"`
	ApprovesWork                    bool                                  `json:"approves_work"`
}

type AtlasMissionReadbackDelta struct {
	Schema                  string                                           `json:"schema"`
	Status                  string                                           `json:"status"`
	SourceReadbackPath      string                                           `json:"source_readback_path"`
	TargetReadbackPath      string                                           `json:"target_readback_path"`
	SourceReadbackDigest    string                                           `json:"source_readback_digest"`
	TargetReadbackDigest    string                                           `json:"target_readback_digest"`
	DeterministicComparison bool                                             `json:"deterministic_comparison"`
	ChangedFields           []string                                         `json:"changed_fields"`
	NumericDeltas           map[string]int                                   `json:"numeric_deltas"`
	BooleanTransitions      map[string]AtlasMissionReadbackBooleanTransition `json:"boolean_transitions"`
	StringTransitions       map[string]AtlasMissionReadbackStringTransition  `json:"string_transitions"`
	SafetyBoundaries        map[string]bool                                  `json:"safety_boundaries"`
	SchedulesWork           bool                                             `json:"schedules_work"`
	ExecutesWork            bool                                             `json:"executes_work"`
	ApprovesWork            bool                                             `json:"approves_work"`
	ClaimsAuthorityAdvance  bool                                             `json:"claims_authority_advance"`
	RSIRemainsDenied        bool                                             `json:"rsi_remains_denied"`
}

type AtlasMissionReadbackBooleanTransition struct {
	Before bool `json:"before"`
	After  bool `json:"after"`
}

type AtlasMissionReadbackStringTransition struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type AtlasMissionReadbackDiffFixture struct {
	Schema                           string                                `json:"schema"`
	Status                           string                                `json:"status"`
	MissionID                        string                                `json:"mission_id"`
	TargetInstance                   string                                `json:"target_instance"`
	SourceReadbackPath               string                                `json:"source_readback_path"`
	TargetReadbackPath               string                                `json:"target_readback_path"`
	DeltaPath                        string                                `json:"delta_path"`
	SourceReadbackDigest             string                                `json:"source_readback_digest"`
	TargetReadbackDigest             string                                `json:"target_readback_digest"`
	DeltaDigest                      string                                `json:"delta_digest"`
	CompletedNodeTransition          AtlasMissionReadbackNumericTransition `json:"completed_node_transition"`
	ReadyNodeTransition              AtlasMissionReadbackNumericTransition `json:"ready_node_transition"`
	CheckpointTransition             AtlasMissionReadbackNumericTransition `json:"checkpoint_transition"`
	FirstExecutableNodeBefore        string                                `json:"first_executable_node_before"`
	FirstExecutableNodeAfter         string                                `json:"first_executable_node_after"`
	ExactNextActionBefore            string                                `json:"exact_next_action_before"`
	ExactNextActionAfter             string                                `json:"exact_next_action_after"`
	ReturnGateStatusBefore           string                                `json:"return_gate_status_before"`
	ReturnGateStatusAfter            string                                `json:"return_gate_status_after"`
	ContinuationContractReasonBefore string                                `json:"continuation_contract_reason_before"`
	ContinuationContractReasonAfter  string                                `json:"continuation_contract_reason_after"`
	FinalResponseAllowedBefore       bool                                  `json:"final_response_allowed_before"`
	FinalResponseAllowedAfter        bool                                  `json:"final_response_allowed_after"`
	ResumeRequired                   bool                                  `json:"resume_required"`
	ResumeReason                     string                                `json:"resume_reason"`
	ExpectedNextAction               string                                `json:"expected_next_action"`
	SafetyBoundaries                 map[string]bool                       `json:"safety_boundaries"`
	SchedulesWork                    bool                                  `json:"schedules_work"`
	ExecutesWork                     bool                                  `json:"executes_work"`
	ApprovesWork                     bool                                  `json:"approves_work"`
	ClaimsAuthorityAdvance           bool                                  `json:"claims_authority_advance"`
	RSIRemainsDenied                 bool                                  `json:"rsi_remains_denied"`
}

type AtlasMissionStaleCheckpointRejection struct {
	Schema                            string                         `json:"schema"`
	Status                            string                         `json:"status"`
	MissionID                         string                         `json:"mission_id"`
	TargetInstance                    string                         `json:"target_instance"`
	StaleReadbackPath                 string                         `json:"stale_readback_path"`
	LatestReadbackPath                string                         `json:"latest_readback_path"`
	PromptReadbackPath                string                         `json:"prompt_readback_path"`
	StaleReadbackDigest               string                         `json:"stale_readback_digest"`
	LatestReadbackDigest              string                         `json:"latest_readback_digest"`
	PromptReadbackDigest              string                         `json:"prompt_readback_digest"`
	StaleCheckpoint                   AtlasMissionCheckpointSnapshot `json:"stale_checkpoint"`
	LatestCheckpoint                  AtlasMissionCheckpointSnapshot `json:"latest_checkpoint"`
	PromptNextExecutableNode          string                         `json:"prompt_next_executable_node"`
	ExpectedCurrentNextExecutableNode string                         `json:"expected_current_next_executable_node"`
	PromptExactNextAction             string                         `json:"prompt_exact_next_action"`
	ExpectedCurrentExactNextAction    string                         `json:"expected_current_exact_next_action"`
	RejectionReason                   string                         `json:"rejection_reason"`
	ContinuationContractReason        string                         `json:"continuation_contract_reason"`
	FinalResponseAllowed              bool                           `json:"final_response_allowed"`
	SafetyBoundaries                  map[string]bool                `json:"safety_boundaries"`
	SchedulesWork                     bool                           `json:"schedules_work"`
	ExecutesWork                      bool                           `json:"executes_work"`
	ApprovesWork                      bool                           `json:"approves_work"`
	ClaimsAuthorityAdvance            bool                           `json:"claims_authority_advance"`
	RSIRemainsDenied                  bool                           `json:"rsi_remains_denied"`
}

type AtlasMissionCheckpointSnapshot struct {
	CompletedNodes             int    `json:"completed_nodes"`
	ReadyNodes                 int    `json:"ready_nodes"`
	CheckpointCount            int    `json:"checkpoint_count"`
	FirstExecutableNode        string `json:"first_executable_node"`
	ExactNextAction            string `json:"exact_next_action"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
}

type AtlasMissionOperatorSummaryCheck struct {
	Schema                            string   `json:"schema"`
	Status                            string   `json:"status"`
	MissionID                         string   `json:"mission_id"`
	TargetInstance                    string   `json:"target_instance"`
	SourceReadbackPath                string   `json:"source_readback_path"`
	SummaryMarkdownPath               string   `json:"summary_markdown_path"`
	SourceReadbackDigest              string   `json:"source_readback_digest"`
	CompletedNodes                    int      `json:"completed_nodes"`
	TotalNodes                        int      `json:"total_nodes"`
	ReadyNodes                        int      `json:"ready_nodes"`
	BlockedNodes                      int      `json:"blocked_nodes"`
	FailedNodes                       int      `json:"failed_nodes"`
	FirstExecutableNode               string   `json:"first_executable_node"`
	ExactNextAction                   string   `json:"exact_next_action"`
	ExactNextActionOccurrences        int      `json:"exact_next_action_occurrences"`
	ExactNextActionWordingPresent     bool     `json:"exact_next_action_wording_present"`
	NextExecutableNodeWordingPresent  bool     `json:"next_executable_node_wording_present"`
	FinalResponseDeniedWordingPresent bool     `json:"final_response_denied_wording_present"`
	ReturnGateStatus                  string   `json:"return_gate_status"`
	ContinuationContractReason        string   `json:"continuation_contract_reason"`
	EarlyReturnRiskStatus             string   `json:"early_return_risk_status"`
	FinalResponseAllowed              bool     `json:"final_response_allowed"`
	RefusesFinalResponse              bool     `json:"refuses_final_response"`
	SummaryAssertions                 []string `json:"summary_assertions"`
	SchedulesWork                     bool     `json:"schedules_work"`
	ExecutesWork                      bool     `json:"executes_work"`
	ApprovesWork                      bool     `json:"approves_work"`
	ClaimsAuthorityAdvance            bool     `json:"claims_authority_advance"`
	RSIRemainsDenied                  bool     `json:"rsi_remains_denied"`
}

type AtlasNodeCommandReadbackEvidence struct {
	Schema                          string `json:"schema"`
	NodeID                          string `json:"node_id"`
	TaskID                          string `json:"task_id"`
	Status                          string `json:"status"`
	CompletedNodesBefore            int    `json:"completed_nodes_before"`
	ReadyNodesBefore                int    `json:"ready_nodes_before"`
	ExpectedCompletedNodesAfter     int    `json:"expected_completed_nodes_after"`
	ExpectedReadyNodesAfter         int    `json:"expected_ready_nodes_after"`
	ExpectedNextExecutableNodeAfter string `json:"expected_next_executable_node_after"`
	FinalResponseAllowed            bool   `json:"final_response_allowed"`
	RSIRemainsDenied                bool   `json:"rsi_remains_denied"`
}

type AtlasNodePromoterNoPromotionEvidence struct {
	Schema                 string `json:"schema"`
	NodeID                 string `json:"node_id"`
	TaskID                 string `json:"task_id"`
	Status                 string `json:"status"`
	PromotionClaimed       bool   `json:"promotion_claimed"`
	ClaimsAuthorityAdvance bool   `json:"claims_authority_advance"`
	RSIRemainsDenied       bool   `json:"rsi_remains_denied"`
}

type AtlasNodeSentinelPublicSafetyEvidence struct {
	Schema                    string   `json:"schema"`
	NodeID                    string   `json:"node_id"`
	TaskID                    string   `json:"task_id"`
	Status                    string   `json:"status"`
	ScanScope                 []string `json:"scan_scope"`
	UnsafePublicClaimDetected bool     `json:"unsafe_public_claim_detected"`
	PromotionClaimDetected    bool     `json:"promotion_claim_detected"`
	RSIClaimDetected          bool     `json:"rsi_claim_detected"`
	RSIRemainsDenied          bool     `json:"rsi_remains_denied"`
}

type AtlasRunLinkSchemaCoverage struct {
	Schema                 string                           `json:"schema"`
	Status                 string                           `json:"status"`
	EvidenceRoot           string                           `json:"evidence_root"`
	RunLinkCount           int                              `json:"run_link_count"`
	CompletedRunLinks      int                              `json:"completed_run_links"`
	SchemaCounts           map[string]int                   `json:"schema_counts"`
	ValidatorCounts        map[string]int                   `json:"validator_counts"`
	Entries                []AtlasRunLinkSchemaCoverageItem `json:"entries"`
	SchedulesWork          bool                             `json:"schedules_work"`
	ExecutesWork           bool                             `json:"executes_work"`
	ApprovesWork           bool                             `json:"approves_work"`
	ClaimsAuthorityAdvance bool                             `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                             `json:"rsi_remains_denied"`
}

type AtlasRunLinkSchemaCoverageItem struct {
	Path             string `json:"path"`
	TaskID           string `json:"task_id"`
	Status           string `json:"status"`
	Schema           string `json:"schema"`
	Validator        string `json:"validator"`
	EvidenceKeyCount int    `json:"evidence_key_count"`
	Digest           string `json:"digest"`
}

type AtlasSchemaValidatorDriftEvidence struct {
	Schema                 string         `json:"schema"`
	Status                 string         `json:"status"`
	SourceReportPath       string         `json:"source_report_path"`
	TargetReportPath       string         `json:"target_report_path"`
	SourceReportDigest     string         `json:"source_report_digest"`
	TargetReportDigest     string         `json:"target_report_digest"`
	SourceNodeCount        int            `json:"source_node_count"`
	TargetNodeCount        int            `json:"target_node_count"`
	JSONFileDelta          int            `json:"json_file_delta"`
	TypedValidatorDelta    int            `json:"typed_validator_delta"`
	GenericSchemaDelta     int            `json:"generic_schema_delta"`
	SchemaCountDeltas      map[string]int `json:"schema_count_deltas"`
	ValidatorCountDeltas   map[string]int `json:"validator_count_deltas"`
	AddedSchemas           []string       `json:"added_schemas"`
	LostSchemas            []string       `json:"lost_schemas"`
	AddedValidators        []string       `json:"added_validators"`
	LostValidators         []string       `json:"lost_validators"`
	UnexpectedLossDetected bool           `json:"unexpected_loss_detected"`
	SchedulesWork          bool           `json:"schedules_work"`
	ExecutesWork           bool           `json:"executes_work"`
	ApprovesWork           bool           `json:"approves_work"`
	ClaimsAuthorityAdvance bool           `json:"claims_authority_advance"`
	RSIRemainsDenied       bool           `json:"rsi_remains_denied"`
}

type AtlasPRCITimingLedger struct {
	Schema                 string               `json:"schema"`
	Status                 string               `json:"status"`
	EvidenceRoot           string               `json:"evidence_root"`
	Rows                   []AtlasPRCITimingRow `json:"rows"`
	SchedulesWork          bool                 `json:"schedules_work"`
	ExecutesWork           bool                 `json:"executes_work"`
	ApprovesWork           bool                 `json:"approves_work"`
	ClaimsAuthorityAdvance bool                 `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                 `json:"rsi_remains_denied"`
}

type AtlasPRCITimingSummary struct {
	Schema                 string               `json:"schema"`
	Status                 string               `json:"status"`
	SourceLedgerPath       string               `json:"source_ledger_path"`
	SourceLedgerDigest     string               `json:"source_ledger_digest"`
	RowCount               int                  `json:"row_count"`
	MergedPRs              int                  `json:"merged_prs"`
	CIPassedPRs            int                  `json:"ci_passed_prs"`
	CIFailedPRs            int                  `json:"ci_failed_prs"`
	CIPendingPRs           int                  `json:"ci_pending_prs"`
	PRNumbers              []int                `json:"pr_numbers"`
	NodeIDs                []string             `json:"node_ids"`
	TotalUbuntuSeconds     int                  `json:"total_ubuntu_seconds"`
	TotalMacosSeconds      int                  `json:"total_macos_seconds"`
	TotalWindowsSeconds    int                  `json:"total_windows_seconds"`
	MeanUbuntuSeconds      int                  `json:"mean_ubuntu_seconds"`
	MeanMacosSeconds       int                  `json:"mean_macos_seconds"`
	MeanWindowsSeconds     int                  `json:"mean_windows_seconds"`
	MaxWindowsSeconds      int                  `json:"max_windows_seconds"`
	MaxCheckSeconds        int                  `json:"max_check_seconds"`
	SlowestPRNumber        int                  `json:"slowest_pr_number"`
	SlowestNodeID          string               `json:"slowest_node_id"`
	SlowestCheck           string               `json:"slowest_check"`
	Rows                   []AtlasPRCITimingRow `json:"rows"`
	SchedulesWork          bool                 `json:"schedules_work"`
	ExecutesWork           bool                 `json:"executes_work"`
	ApprovesWork           bool                 `json:"approves_work"`
	ClaimsAuthorityAdvance bool                 `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                 `json:"rsi_remains_denied"`
}

type AtlasPRCITimingRow struct {
	NodeID          string `json:"node_id"`
	PRNumber        int    `json:"pr_number"`
	MergeCommit     string `json:"merge_commit"`
	CIStatus        string `json:"ci_status"`
	UbuntuSeconds   int    `json:"ubuntu_seconds"`
	MacosSeconds    int    `json:"macos_seconds"`
	WindowsSeconds  int    `json:"windows_seconds"`
	MaxCheckSeconds int    `json:"max_check_seconds"`
	SlowestCheck    string `json:"slowest_check"`
}

type AtlasPRCIWindowsThresholdEvidence struct {
	Schema                   string                         `json:"schema"`
	Status                   string                         `json:"status"`
	SourceSummaryPath        string                         `json:"source_summary_path"`
	SourceSummaryDigest      string                         `json:"source_summary_digest"`
	ThresholdSeconds         int                            `json:"threshold_seconds"`
	RowCount                 int                            `json:"row_count"`
	LongRunningWindowsChecks int                            `json:"long_running_windows_checks"`
	MaxWindowsSeconds        int                            `json:"max_windows_seconds"`
	MaxOverThresholdSeconds  int                            `json:"max_over_threshold_seconds"`
	Rows                     []AtlasPRCIWindowsThresholdRow `json:"rows"`
	SchedulesWork            bool                           `json:"schedules_work"`
	ExecutesWork             bool                           `json:"executes_work"`
	ApprovesWork             bool                           `json:"approves_work"`
	ClaimsAuthorityAdvance   bool                           `json:"claims_authority_advance"`
	RSIRemainsDenied         bool                           `json:"rsi_remains_denied"`
}

type AtlasPRCIWindowsThresholdRow struct {
	NodeID               string `json:"node_id"`
	PRNumber             int    `json:"pr_number"`
	CIStatus             string `json:"ci_status"`
	MergeCommit          string `json:"merge_commit"`
	WindowsSeconds       int    `json:"windows_seconds"`
	ThresholdSeconds     int    `json:"threshold_seconds"`
	ExceedsThreshold     bool   `json:"exceeds_threshold"`
	OverThresholdSeconds int    `json:"over_threshold_seconds"`
}

type AtlasFailedCheckReplayInput struct {
	Schema                 string                       `json:"schema"`
	Status                 string                       `json:"status"`
	Cases                  []AtlasFailedCheckReplayCase `json:"cases"`
	SchedulesWork          bool                         `json:"schedules_work"`
	ExecutesWork           bool                         `json:"executes_work"`
	ApprovesWork           bool                         `json:"approves_work"`
	ClaimsAuthorityAdvance bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                         `json:"rsi_remains_denied"`
}

type AtlasFailedCheckReplayFixture struct {
	Schema                 string                              `json:"schema"`
	Status                 string                              `json:"status"`
	SourceInputPath        string                              `json:"source_input_path"`
	SourceInputDigest      string                              `json:"source_input_digest"`
	CaseCount              int                                 `json:"case_count"`
	MergeDeniedCases       int                                 `json:"merge_denied_cases"`
	RetryAllowedCases      int                                 `json:"retry_allowed_cases"`
	SafeToMerge            bool                                `json:"safe_to_merge"`
	Cases                  []AtlasFailedCheckReplayFixtureCase `json:"cases"`
	SchedulesWork          bool                                `json:"schedules_work"`
	ExecutesWork           bool                                `json:"executes_work"`
	ApprovesWork           bool                                `json:"approves_work"`
	ClaimsAuthorityAdvance bool                                `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                                `json:"rsi_remains_denied"`
}

type AtlasFailedCheckReplayCase struct {
	ID          string `json:"id"`
	CheckName   string `json:"check_name"`
	Platform    string `json:"platform"`
	CheckStatus string `json:"check_status"`
	FailureKind string `json:"failure_kind"`
	Retryable   bool   `json:"retryable"`
}

type AtlasFailedCheckReplayFixtureCase struct {
	ID            string `json:"id"`
	CheckName     string `json:"check_name"`
	Platform      string `json:"platform"`
	CheckStatus   string `json:"check_status"`
	FailureKind   string `json:"failure_kind"`
	RetryDecision string `json:"retry_decision"`
	MergeDecision string `json:"merge_decision"`
	Reason        string `json:"reason"`
}

type AtlasMergeCheckBindingInput struct {
	Schema                 string                      `json:"schema"`
	Status                 string                      `json:"status"`
	Rows                   []AtlasMergeCheckBindingRow `json:"rows"`
	SchedulesWork          bool                        `json:"schedules_work"`
	ExecutesWork           bool                        `json:"executes_work"`
	ApprovesWork           bool                        `json:"approves_work"`
	ClaimsAuthorityAdvance bool                        `json:"claims_authority_advance"`
	RSIRemainsDenied       bool                        `json:"rsi_remains_denied"`
}

type AtlasMergeCheckBinding struct {
	Schema                  string                      `json:"schema"`
	Status                  string                      `json:"status"`
	SourceInputPath         string                      `json:"source_input_path"`
	SourceInputDigest       string                      `json:"source_input_digest"`
	RowCount                int                         `json:"row_count"`
	PassedRequiredCheckRows int                         `json:"passed_required_check_rows"`
	UnboundMergeCommits     int                         `json:"unbound_merge_commits"`
	Rows                    []AtlasMergeCheckBindingRow `json:"rows"`
	SchedulesWork           bool                        `json:"schedules_work"`
	ExecutesWork            bool                        `json:"executes_work"`
	ApprovesWork            bool                        `json:"approves_work"`
	ClaimsAuthorityAdvance  bool                        `json:"claims_authority_advance"`
	RSIRemainsDenied        bool                        `json:"rsi_remains_denied"`
}

type AtlasMergeCheckBindingRow struct {
	NodeID                   string `json:"node_id"`
	PRNumber                 int    `json:"pr_number"`
	MergeCommit              string `json:"merge_commit"`
	RequiredCheckCount       int    `json:"required_check_count"`
	PassedRequiredCheckCount int    `json:"passed_required_check_count"`
	RequiredChecksStatus     string `json:"required_checks_status,omitempty"`
	MergeCommitBound         bool   `json:"merge_commit_bound,omitempty"`
}

type AtlasPostMergeBranchDeletionReadback struct {
	Schema                   string                                      `json:"schema"`
	Status                   string                                      `json:"status"`
	EvidenceRoot             string                                      `json:"evidence_root"`
	PostMergeLifecycleCount  int                                         `json:"post_merge_lifecycle_count"`
	LocalBranchDeletedCount  int                                         `json:"local_branch_deleted_count"`
	RemoteBranchDeletedCount int                                         `json:"remote_branch_deleted_count"`
	BranchesRemainingTotal   int                                         `json:"branches_remaining_total"`
	Entries                  []AtlasPostMergeBranchDeletionReadbackEntry `json:"entries"`
	SchedulesWork            bool                                        `json:"schedules_work"`
	ExecutesWork             bool                                        `json:"executes_work"`
	ApprovesWork             bool                                        `json:"approves_work"`
	ClaimsAuthorityAdvance   bool                                        `json:"claims_authority_advance"`
	RSIRemainsDenied         bool                                        `json:"rsi_remains_denied"`
}

type AtlasPostMergeBranchDeletionReadbackEntry struct {
	Path                         string `json:"path"`
	NodeID                       string `json:"node_id"`
	Status                       string `json:"status"`
	PRNumber                     int    `json:"pr_number"`
	MergeCommit                  string `json:"merge_commit"`
	CIStatus                     string `json:"ci_status"`
	LocalBranchDeleted           bool   `json:"local_branch_deleted"`
	RemoteBranchDeleted          bool   `json:"remote_branch_deleted"`
	LocalCodexBranchesRemaining  int    `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int    `json:"remote_codex_branches_remaining"`
	Digest                       string `json:"digest"`
}

type AtlasMissionReadbackNumericTransition struct {
	Before int `json:"before"`
	After  int `json:"after"`
	Delta  int `json:"delta"`
}

type AtlasFoundryTerminalStatusExample struct {
	SourceStatus     string `json:"source_status"`
	NormalizedStatus string `json:"normalized_status"`
	Terminal         bool   `json:"terminal"`
	CanCloseMission  bool   `json:"can_close_mission"`
	RequiredReadback string `json:"required_readback"`
}

type AtlasFoundryDeniedTerminalExample struct {
	DenialReason                 string `json:"denial_reason"`
	NormalizedStatus             string `json:"normalized_status"`
	Terminal                     bool   `json:"terminal"`
	CanCloseMission              bool   `json:"can_close_mission"`
	RequiresExactMissingEvidence bool   `json:"requires_exact_missing_evidence"`
	RequiredReadback             string `json:"required_readback"`
	RSIRemainsDenied             bool   `json:"rsi_remains_denied"`
	AuthorityAdvanceClaimed      bool   `json:"authority_advance_claimed"`
}

type AtlasContinuationContract struct {
	ContractVersion      string `json:"contract_version"`
	Status               string `json:"status"`
	ReadyNodes           int    `json:"ready_nodes"`
	ExactNextAction      string `json:"exact_next_action"`
	ReturnGateStatus     string `json:"return_gate_status"`
	FinalResponseAllowed bool   `json:"final_response_allowed"`
	RefusesFinalResponse bool   `json:"refuses_final_response"`
	Reason               string `json:"reason"`
	Source               string `json:"source"`
}

type AtlasExactNextActionReadback struct {
	Status               string `json:"status"`
	Action               string `json:"action"`
	NextExecutableNode   string `json:"next_executable_node"`
	ReturnGateStatus     string `json:"return_gate_status"`
	FinalResponseAllowed bool   `json:"final_response_allowed"`
	Source               string `json:"source"`
}

type AtlasCommandTimelinePlaceholder struct {
	Slot                        string `json:"slot"`
	Source                      string `json:"source"`
	Status                      string `json:"status"`
	Summary                     string `json:"summary"`
	RequiredBeforeFinalResponse bool   `json:"required_before_final_response"`
}

type AtlasPromoterNoPromotionPlaceholder struct {
	Slot                        string `json:"slot"`
	Source                      string `json:"source"`
	Status                      string `json:"status"`
	Summary                     string `json:"summary"`
	RequiredBeforeFinalResponse bool   `json:"required_before_final_response"`
}

type AtlasRecommendationNodeEvidence struct {
	NodeID                 string   `json:"node_id"`
	TaskID                 string   `json:"task_id"`
	Status                 string   `json:"status"`
	NodeGate               string   `json:"node_gate"`
	CandidateRecord        string   `json:"candidate_record"`
	RollbackRecord         string   `json:"rollback_record"`
	ImplementationEvidence string   `json:"implementation_evidence"`
	Tests                  string   `json:"tests"`
	Verification           string   `json:"verification"`
	PublicSafetyWording    string   `json:"public_safety_wording"`
	PromoterReadback       string   `json:"promoter_readback"`
	CommandReadback        string   `json:"command_readback"`
	RequiredGates          []string `json:"required_gates"`
	VerificationCommands   []string `json:"verification_commands"`
}

type AtlasRecommendationExecutionReadback struct {
	Schema                         string                                            `json:"schema"`
	Status                         string                                            `json:"status"`
	MissionID                      string                                            `json:"mission_id"`
	EvidenceRoot                   string                                            `json:"evidence_root,omitempty"`
	LeaseHealthStatus              string                                            `json:"lease_health_status"`
	CheckpointFreshnessStatus      string                                            `json:"checkpoint_freshness_status"`
	ReturnGateStatus               string                                            `json:"return_gate_status,omitempty"`
	ContinuationContractReason     string                                            `json:"continuation_contract_reason"`
	ExactNextAction                string                                            `json:"exact_next_action"`
	FinalResponseAllowed           bool                                              `json:"final_response_allowed"`
	RefusesFinalResponse           bool                                              `json:"refuses_final_response"`
	CompletedRecommendationNodes   int                                               `json:"completed_recommendation_nodes"`
	TotalRecommendationNodes       int                                               `json:"total_recommendation_nodes"`
	GeneratedWorkgraph             AtlasRecommendationGeneratedWorkgraphReadback     `json:"generated_workgraph"`
	FoundryRunLinkReadinessSummary AtlasRecommendationFoundryRunLinkReadinessSummary `json:"foundry_run_link_readiness_summary"`
	ContinuationReasonCoverage     AtlasRecommendationContinuationReasonCoverage     `json:"continuation_reason_coverage"`
	ReasonArtifactAgreementSummary AtlasRecommendationReasonArtifactAgreementSummary `json:"reason_artifact_agreement_summary"`
	SourceArtifacts                []SourceRef                                       `json:"source_artifacts"`
}

type AtlasRecommendationContinuationReasonCoverage struct {
	Status                    string   `json:"status"`
	ExpectedReason            string   `json:"expected_reason"`
	IndexedSources            []string `json:"indexed_sources"`
	SourceCount               int      `json:"source_count"`
	FinalResponseAllowed      bool     `json:"final_response_allowed"`
	RefusesFinalResponse      bool     `json:"refuses_final_response"`
	ExactNextAction           string   `json:"exact_next_action"`
	ReturnGateStatus          string   `json:"return_gate_status,omitempty"`
	LeaseHealthStatus         string   `json:"lease_health_status"`
	CheckpointFreshnessStatus string   `json:"checkpoint_freshness_status"`
	ClaimsAuthorityAdvance    bool     `json:"claims_authority_advance"`
	RSIRemainsDenied          bool     `json:"rsi_remains_denied"`
}

type AtlasRecommendationReasonArtifactAgreementSummary struct {
	Status                           string   `json:"status"`
	ExpectedReason                   string   `json:"expected_reason"`
	IndexedSources                   []string `json:"indexed_sources"`
	SourceCount                      int      `json:"source_count"`
	AllRequiredSourcesIndexed        bool     `json:"all_required_sources_indexed"`
	SourceArtifactRefs               []string `json:"source_artifact_refs"`
	SourceArtifactCount              int      `json:"source_artifact_count"`
	SourceArtifactsAgree             bool     `json:"source_artifacts_agree"`
	FoundryRunLinkReadinessDigest    string   `json:"foundry_run_link_readiness_digest"`
	ContinuationReasonCoverageDigest string   `json:"continuation_reason_coverage_digest"`
	FinalResponseAllowed             bool     `json:"final_response_allowed"`
	RefusesFinalResponse             bool     `json:"refuses_final_response"`
	ExactNextAction                  string   `json:"exact_next_action"`
	ReturnGateStatus                 string   `json:"return_gate_status,omitempty"`
	ClaimsAuthorityAdvance           bool     `json:"claims_authority_advance"`
	RSIRemainsDenied                 bool     `json:"rsi_remains_denied"`
}

type AtlasRecommendationFoundryRunLinkReadinessSummary struct {
	Status                     string `json:"status"`
	Summary                    string `json:"summary"`
	CompletedRunLinks          int    `json:"completed_run_links"`
	RequiredRunLinks           int    `json:"required_run_links"`
	MissingRunLinks            int    `json:"missing_run_links"`
	ReadyNodes                 int    `json:"ready_nodes"`
	NextExecutableNode         string `json:"next_executable_node,omitempty"`
	LeaseHealthStatus          string `json:"lease_health_status"`
	CheckpointFreshnessStatus  string `json:"checkpoint_freshness_status"`
	ReturnGateStatus           string `json:"return_gate_status,omitempty"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	ExactNextAction            string `json:"exact_next_action"`
	RefusesFinalResponse       bool   `json:"refuses_final_response"`
	CheckpointCount            int    `json:"checkpoint_count"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
}

type AtlasRecommendationLeaseStart struct {
	Schema                 string `json:"schema"`
	Status                 string `json:"status"`
	MissionID              string `json:"mission_id"`
	TargetInstance         string `json:"target_instance"`
	EvidenceRoot           string `json:"evidence_root,omitempty"`
	StartedAt              string `json:"started_at"`
	MinMinutes             int    `json:"min_minutes"`
	MaxMinutes             int    `json:"max_minutes"`
	ContinueIfFastTarget   int    `json:"continue_if_fast_target"`
	CheckpointPolicy       string `json:"checkpoint_policy"`
	WaveDigest             string `json:"wave_digest"`
	WorkgraphDigest        string `json:"workgraph_digest"`
	FinalResponseAllowed   bool   `json:"final_response_allowed"`
	FinalResponseReason    string `json:"final_response_reason"`
	SchedulesWork          bool   `json:"schedules_work"`
	ExecutesWork           bool   `json:"executes_work"`
	ApprovesWork           bool   `json:"approves_work"`
	MutatesRepositories    bool   `json:"mutates_repositories"`
	CallsProviders         bool   `json:"calls_providers"`
	ClaimsAuthorityAdvance bool   `json:"claims_authority_advance"`
}

type AtlasRecommendationCheckpointReadback struct {
	Schema                     string `json:"schema"`
	Status                     string `json:"status"`
	MissionID                  string `json:"mission_id"`
	EvidenceRoot               string `json:"evidence_root,omitempty"`
	StartedAt                  string `json:"started_at,omitempty"`
	CompletedAt                string `json:"completed_at,omitempty"`
	ElapsedMinutes             int    `json:"elapsed_minutes"`
	MinMinutes                 int    `json:"min_minutes"`
	MaxMinutes                 int    `json:"max_minutes"`
	MinMinutesMet              bool   `json:"min_minutes_met"`
	LeaseTimeStatus            string `json:"lease_time_status"`
	LeaseHealthStatus          string `json:"lease_health_status"`
	CheckpointFreshnessStatus  string `json:"checkpoint_freshness_status"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	CompletedNodes             int    `json:"completed_nodes"`
	ReadyNodes                 int    `json:"ready_nodes"`
	BlockedNodes               int    `json:"blocked_nodes"`
	FailedNodes                int    `json:"failed_nodes"`
	TotalNodes                 int    `json:"total_nodes"`
	FirstExecutableNode        string `json:"first_executable_node,omitempty"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
	FinalResponseReason        string `json:"final_response_reason"`
	ExactNextAction            string `json:"exact_next_action"`
	SchedulesWork              bool   `json:"schedules_work"`
	ExecutesWork               bool   `json:"executes_work"`
	ApprovesWork               bool   `json:"approves_work"`
	ClaimsAuthorityAdvance     bool   `json:"claims_authority_advance"`
}

type AtlasRecommendationGeneratedWorkgraphReadback struct {
	TotalNodes                int    `json:"total_nodes"`
	ReadyNodes                int    `json:"ready_nodes"`
	ExecutableReadyNodes      int    `json:"executable_ready_nodes"`
	FirstExecutableNode       string `json:"first_executable_node,omitempty"`
	LeaseHealthStatus         string `json:"lease_health_status"`
	CheckpointFreshnessStatus string `json:"checkpoint_freshness_status"`
	ReturnGateStatus          string `json:"return_gate_status,omitempty"`
	CheckpointCount           int    `json:"checkpoint_count"`
	FinalResponseAllowed      bool   `json:"final_response_allowed"`
	FinalResponseReason       string `json:"final_response_reason,omitempty"`
}

type AtlasRecommendationWorkgraphReadinessPacket struct {
	Schema                          string `json:"schema"`
	Status                          string `json:"status"`
	MissionID                       string `json:"mission_id"`
	TargetInstance                  string `json:"target_instance"`
	EvidenceRoot                    string `json:"evidence_root,omitempty"`
	WaveDigest                      string `json:"wave_digest"`
	WorkgraphDigest                 string `json:"workgraph_digest"`
	ReadbackDigest                  string `json:"readback_digest"`
	TotalNodes                      int    `json:"total_nodes"`
	MinimumNodes                    int    `json:"minimum_nodes"`
	NodeBudget                      int    `json:"node_budget"`
	ContinueIfFastTarget            int    `json:"continue_if_fast_target"`
	CompletedNodes                  int    `json:"completed_nodes"`
	ReadyNodes                      int    `json:"ready_nodes"`
	BlockedNodes                    int    `json:"blocked_nodes"`
	FailedNodes                     int    `json:"failed_nodes"`
	ExecutableReadyNodes            int    `json:"executable_ready_nodes"`
	FirstExecutableNode             string `json:"first_executable_node,omitempty"`
	LeaseHealthStatus               string `json:"lease_health_status"`
	CheckpointFreshnessStatus       string `json:"checkpoint_freshness_status"`
	ReturnGateStatus                string `json:"return_gate_status"`
	CheckpointCount                 int    `json:"checkpoint_count"`
	EarlyReturnRiskStatus           string `json:"early_return_risk_status"`
	ContinuationBudgetStatus        string `json:"continuation_budget_status"`
	FinalResponseAllowed            bool   `json:"final_response_allowed"`
	FinalResponseReason             string `json:"final_response_reason"`
	ExactNextAction                 string `json:"exact_next_action"`
	ContinuationContractReason      string `json:"continuation_contract_reason"`
	OneExecutableMutationNodeActive bool   `json:"one_executable_mutation_node_active"`
	RefusesFinalResponse            bool   `json:"refuses_final_response"`
	SchedulesWork                   bool   `json:"schedules_work"`
	ExecutesWork                    bool   `json:"executes_work"`
	ApprovesWork                    bool   `json:"approves_work"`
	ClaimsAuthorityAdvance          bool   `json:"claims_authority_advance"`
	RSIRemainsDenied                bool   `json:"rsi_remains_denied"`
}

type AtlasRecommendationCommandReadback struct {
	Schema                     string                                    `json:"schema"`
	Status                     string                                    `json:"status"`
	MissionID                  string                                    `json:"mission_id"`
	EvidenceRoot               string                                    `json:"evidence_root,omitempty"`
	CompletedNodes             int                                       `json:"completed_nodes"`
	ReadyNodes                 int                                       `json:"ready_nodes"`
	BlockedNodes               int                                       `json:"blocked_nodes"`
	FailedNodes                int                                       `json:"failed_nodes"`
	TotalNodes                 int                                       `json:"total_nodes"`
	StartedAt                  string                                    `json:"started_at,omitempty"`
	CompletedAt                string                                    `json:"completed_at,omitempty"`
	ElapsedMinutes             int                                       `json:"elapsed_minutes"`
	MinMinutes                 int                                       `json:"min_minutes"`
	MinMinutesMet              bool                                      `json:"min_minutes_met"`
	LeaseTimeStatus            string                                    `json:"lease_time_status"`
	LeaseHealthStatus          string                                    `json:"lease_health_status"`
	CheckpointFreshnessStatus  string                                    `json:"checkpoint_freshness_status"`
	NodeCompletionStatus       string                                    `json:"node_completion_status"`
	ReturnGateStatus           string                                    `json:"return_gate_status,omitempty"`
	CheckpointCount            int                                       `json:"checkpoint_count"`
	FinalResponseAllowed       bool                                      `json:"final_response_allowed"`
	FinalResponseReason        string                                    `json:"final_response_reason"`
	ExactNextAction            string                                    `json:"exact_next_action"`
	ContinuationContractReason string                                    `json:"continuation_contract_reason"`
	CompactTimeline            string                                    `json:"compact_timeline"`
	CommandTimelineBinding     AtlasRecommendationCommandTimelineBinding `json:"command_timeline_binding"`
	SchedulesWork              bool                                      `json:"schedules_work"`
	ExecutesWork               bool                                      `json:"executes_work"`
	ApprovesWork               bool                                      `json:"approves_work"`
	ClaimsAuthorityAdvance     bool                                      `json:"claims_authority_advance"`
}

type AtlasRecommendationCommandTimelineBinding struct {
	Summary                    string `json:"summary"`
	FirstExecutableNode        string `json:"first_executable_node,omitempty"`
	ExactNextAction            string `json:"exact_next_action"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	ReturnGateStatus           string `json:"return_gate_status"`
	NodeCompletionStatus       string `json:"node_completion_status"`
	LeaseTimeStatus            string `json:"lease_time_status"`
	LeaseHealthStatus          string `json:"lease_health_status"`
	CheckpointFreshnessStatus  string `json:"checkpoint_freshness_status"`
	CheckpointCount            int    `json:"checkpoint_count"`
	CompletedNodes             int    `json:"completed_nodes"`
	ReadyNodes                 int    `json:"ready_nodes"`
	TotalNodes                 int    `json:"total_nodes"`
	ElapsedMinutes             int    `json:"elapsed_minutes"`
	MinMinutes                 int    `json:"min_minutes"`
	MinMinutesMet              bool   `json:"min_minutes_met"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
}

type AtlasRecommendationPromoterReadback struct {
	Schema                     string `json:"schema"`
	Status                     string `json:"status"`
	MissionID                  string `json:"mission_id"`
	EvidenceRoot               string `json:"evidence_root,omitempty"`
	PromotionClaimed           bool   `json:"promotion_claimed"`
	RSIRemainsDenied           bool   `json:"rsi_remains_denied"`
	NoPromotionSummary         string `json:"no_promotion_summary"`
	NoPromotionReasonSummary   string `json:"no_promotion_reason_summary"`
	NextDeniedClass            string `json:"next_denied_class"`
	Reason                     string `json:"reason"`
	ElapsedMinutes             int    `json:"elapsed_minutes"`
	MinMinutesMet              bool   `json:"min_minutes_met"`
	LeaseHealthStatus          string `json:"lease_health_status"`
	CheckpointFreshnessStatus  string `json:"checkpoint_freshness_status"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
	SchedulesWork              bool   `json:"schedules_work"`
	ExecutesWork               bool   `json:"executes_work"`
	ApprovesWork               bool   `json:"approves_work"`
	ClaimsAuthorityAdvance     bool   `json:"claims_authority_advance"`
}

type AtlasRecommendationFoundryRollup struct {
	Schema                     string `json:"schema"`
	Status                     string `json:"status"`
	MissionID                  string `json:"mission_id"`
	EvidenceRoot               string `json:"evidence_root,omitempty"`
	CompletedNodes             int    `json:"completed_nodes"`
	ReadyNodes                 int    `json:"ready_nodes"`
	BlockedNodes               int    `json:"blocked_nodes"`
	FailedNodes                int    `json:"failed_nodes"`
	TotalNodes                 int    `json:"total_nodes"`
	NodeCompletionStatus       string `json:"node_completion_status"`
	LeaseCompletionStatus      string `json:"lease_completion_status"`
	LeaseHealthStatus          string `json:"lease_health_status"`
	CheckpointFreshnessStatus  string `json:"checkpoint_freshness_status"`
	ReturnGateStatus           string `json:"return_gate_status,omitempty"`
	CheckpointCount            int    `json:"checkpoint_count"`
	FinalResponseAllowed       bool   `json:"final_response_allowed"`
	ExactNextAction            string `json:"exact_next_action"`
	ContinuationContractReason string `json:"continuation_contract_reason"`
	SchedulesWork              bool   `json:"schedules_work"`
	ExecutesWork               bool   `json:"executes_work"`
	ApprovesWork               bool   `json:"approves_work"`
	ClaimsAuthorityAdvance     bool   `json:"claims_authority_advance"`
}

type AtlasRecommendationReconciliationPacket struct {
	Schema                       string                        `json:"schema"`
	Status                       string                        `json:"status"`
	MissionID                    string                        `json:"mission_id"`
	EvidenceRoot                 string                        `json:"evidence_root,omitempty"`
	FinalStateReconciliation     AtlasFinalStateReconciliation `json:"final_state_reconciliation"`
	CompletedNodes               int                           `json:"completed_nodes"`
	ReadyNodes                   int                           `json:"ready_nodes"`
	BlockedNodes                 int                           `json:"blocked_nodes"`
	FailedNodes                  int                           `json:"failed_nodes"`
	TotalNodes                   int                           `json:"total_nodes"`
	CheckpointCount              int                           `json:"checkpoint_count"`
	ReturnGateStatus             string                        `json:"return_gate_status"`
	LeaseTimeStatus              string                        `json:"lease_time_status"`
	LeaseHealthStatus            string                        `json:"lease_health_status"`
	CheckpointFreshnessStatus    string                        `json:"checkpoint_freshness_status"`
	StaleRouteDecisionStatus     string                        `json:"stale_route_decision_status"`
	FinalResponseAllowed         bool                          `json:"final_response_allowed"`
	FinalResponseReason          string                        `json:"final_response_reason"`
	ExactNextAction              string                        `json:"exact_next_action"`
	ContinuationContractReason   string                        `json:"continuation_contract_reason"`
	CommandReturnGateStatus      string                        `json:"command_return_gate_status"`
	CommandContinuationReason    string                        `json:"command_continuation_contract_reason"`
	CommandFinalResponseAllowed  bool                          `json:"command_final_response_allowed"`
	PromoterStatus               string                        `json:"promoter_status"`
	PromoterContinuationReason   string                        `json:"promoter_continuation_contract_reason"`
	PromotionClaimed             bool                          `json:"promotion_claimed"`
	RSIRemainsDenied             bool                          `json:"rsi_remains_denied"`
	FoundryStatus                string                        `json:"foundry_status"`
	FoundryReturnGateStatus      string                        `json:"foundry_return_gate_status"`
	FoundryContinuationReason    string                        `json:"foundry_continuation_contract_reason"`
	FoundryNodeCompletionStatus  string                        `json:"foundry_node_completion_status"`
	FoundryLeaseCompletionStatus string                        `json:"foundry_lease_completion_status"`
	FoundryFinalResponseAllowed  bool                          `json:"foundry_final_response_allowed"`
	ContinuationReasonAgreement  bool                          `json:"continuation_reason_agreement"`
	ArtifactsAgree               bool                          `json:"artifacts_agree"`
	SchedulesWork                bool                          `json:"schedules_work"`
	ExecutesWork                 bool                          `json:"executes_work"`
	ApprovesWork                 bool                          `json:"approves_work"`
	ClaimsAuthorityAdvance       bool                          `json:"claims_authority_advance"`
}

type AOMissionSourceArtifact struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type AuthorityLadderStatus struct {
	CurrentClass        string            `json:"current_class"`
	NextClass           string            `json:"next_class"`
	ProvenLiveClasses   []string          `json:"proven_live_classes"`
	DryRunReadyClasses  []string          `json:"dry_run_ready_classes"`
	Blockers            []string          `json:"blockers"`
	RequiredEvidence    []string          `json:"required_evidence"`
	DeniedHigherClasses map[string]string `json:"denied_higher_classes"`
	DoNotAdvanceGates   []string          `json:"do_not_advance_gates"`
}

type BlueprintRequest struct {
	ContractVersion string   `json:"contract_version"`
	IntakeID        string   `json:"intake_id"`
	Status          string   `json:"status"`
	Missing         []string `json:"missing"`
	Reason          string   `json:"reason"`
}

type BlueprintBuildAuthorization struct {
	SchemaVersion       string   `json:"schema"`
	ProjectID           string   `json:"project_id"`
	Status              string   `json:"status"`
	Score               int      `json:"score"`
	ApprovedByUser      bool     `json:"approved_by_user"`
	BlockingAssumptions []string `json:"blocking_assumptions"`
	NextAllowedAction   string   `json:"next_allowed_action"`
	Scope               string   `json:"scope,omitempty"`
	MutationClass       string   `json:"mutation_class,omitempty"`
	BlueprintPackDigest string   `json:"blueprint_pack_digest,omitempty"`
	ExpiresAtUTC        string   `json:"expires_at_utc,omitempty"`
}

type BlueprintCandidateRules struct {
	SchemaVersion     string   `json:"schema_version"`
	ProjectID         string   `json:"project_id"`
	TargetInstance    string   `json:"target_instance"`
	WorkgraphID       string   `json:"workgraph_id"`
	CandidateID       string   `json:"candidate_id"`
	MutationClass     string   `json:"mutation_class"`
	TargetFactoryRepo string   `json:"target_factory_repo"`
	FactoryFolder     string   `json:"factory_folder"`
	Objective         string   `json:"objective"`
	Acceptance        []string `json:"acceptance_criteria"`
	NonGoals          []string `json:"non_goals"`
	WriteScope        []string `json:"write_scope"`
	RollbackScope     []string `json:"rollback_scope"`
	RequiredGates     []string `json:"required_gates"`
	Verification      []string `json:"verification_commands"`
	RequiredEvidence  []string `json:"required_evidence"`
	SafetyLimits      []string `json:"safety_limits"`
	AuthorityBoundary string   `json:"authority_boundary"`
	ContextRefs       []string `json:"context_refs"`
	DependencyRefs    []string `json:"dependency_refs,omitempty"`
}

type BlueprintCandidateSelection struct {
	ContractVersion     string            `json:"contract_version"`
	ID                  string            `json:"id"`
	ProjectID           string            `json:"project_id"`
	Status              string            `json:"status"`
	SelectedCandidateID string            `json:"selected_candidate_id"`
	MutationClass       string            `json:"mutation_class"`
	TargetFactoryRepo   string            `json:"target_factory_repo"`
	WorkgraphID         string            `json:"workgraph_id"`
	NodeID              string            `json:"node_id"`
	TaskID              string            `json:"task_id"`
	RequiredGates       []string          `json:"required_gates"`
	RequiredEvidence    []string          `json:"required_evidence"`
	SafetyLimits        []string          `json:"safety_limits"`
	Digests             map[string]string `json:"digests"`
	SchedulesWork       bool              `json:"schedules_work"`
	ExecutesWork        bool              `json:"executes_work"`
	ApprovesWork        bool              `json:"approves_work"`
	MutatesRepositories bool              `json:"mutates_repositories"`
	SafeToExecute       bool              `json:"safe_to_execute"`
	LiveExecutionProven bool              `json:"live_execution_proven"`
}

type BlueprintImport struct {
	ContractVersion                      string                      `json:"contract_version"`
	ID                                   string                      `json:"id"`
	ProjectID                            string                      `json:"project_id"`
	Status                               string                      `json:"status"`
	Reason                               string                      `json:"reason"`
	BlueprintPack                        SourceRef                   `json:"blueprint_pack"`
	BuildAuthorization                   SourceRef                   `json:"build_authorization,omitempty"`
	TargetInstance                       string                      `json:"target_instance"`
	WorkgraphID                          string                      `json:"workgraph_id,omitempty"`
	MutationClass                        string                      `json:"mutation_class,omitempty"`
	CandidateSelection                   BlueprintCandidateSelection `json:"candidate_selection,omitempty"`
	DownstreamFoundryImport              SourceRef                   `json:"downstream_foundry_import,omitempty"`
	DownstreamFoundryContinuationHandoff SourceRef                   `json:"downstream_foundry_continuation_handoff,omitempty"`
	Digests                              map[string]string           `json:"digests"`
	SafetyLimits                         []string                    `json:"safety_limits"`
	BlockingNextActions                  []string                    `json:"blocking_next_actions,omitempty"`
	ReadyForFoundry                      bool                        `json:"ready_for_foundry"`
	SafeToExecute                        bool                        `json:"safe_to_execute"`
	LiveExecutionProven                  bool                        `json:"live_execution_proven"`
	SchedulesWork                        bool                        `json:"schedules_work"`
	ExecutesWork                         bool                        `json:"executes_work"`
	ApprovesWork                         bool                        `json:"approves_work"`
	MutatesRepositories                  bool                        `json:"mutates_repositories"`
	CallsProviders                       bool                        `json:"calls_providers"`
	ReleaseOrPublishAllowed              bool                        `json:"release_or_publish_allowed"`
}

type BlueprintImportPaths struct {
	PackPath            string
	CandidateRulesPath  string
	AuthorizationPath   string
	InstancePath        string
	MutationClassesPath string
	OutDir              string
}

type BlueprintImportResult struct {
	Record        BlueprintImport
	Request       BlueprintRequest
	Intake        Intake
	Candidate     BlueprintCandidateSelection
	ContextPacks  []ContextPack
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Handoff       FoundryContinuationHandoff
}

type Workgraph struct {
	ContractVersion string          `json:"contract_version"`
	ID              string          `json:"id"`
	TargetInstance  string          `json:"target_instance"`
	Nodes           []WorkgraphNode `json:"nodes"`
}

type WorkgraphNode struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	FactoryTask  FactoryTask `json:"factory_task"`
	Dependencies []string    `json:"dependencies"`
	Blockers     []string    `json:"blockers"`
	StitchTask   bool        `json:"stitch_task"`
}

type WorkgraphRepairPlan struct {
	ContractVersion     string        `json:"contract_version"`
	ID                  string        `json:"id"`
	TaskID              string        `json:"task_id"`
	Status              string        `json:"status"`
	SourceRunLinkStatus string        `json:"source_run_link_status"`
	Reason              string        `json:"reason"`
	RepairTasks         []FactoryTask `json:"repair_tasks"`
	SchedulesWork       bool          `json:"schedules_work"`
	ExecutesWork        bool          `json:"executes_work"`
	ApprovesWork        bool          `json:"approves_work"`
}

type MutationClassModel struct {
	ContractVersion string                    `json:"contract_version"`
	ID              string                    `json:"id"`
	Classes         []MutationClassDefinition `json:"classes"`
	SchedulesWork   bool                      `json:"schedules_work"`
	ExecutesWork    bool                      `json:"executes_work"`
	ApprovesWork    bool                      `json:"approves_work"`
}

type MutationClassDefinition struct {
	Name                  string   `json:"name"`
	AllowedPaths          []string `json:"allowed_paths"`
	ForbiddenPaths        []string `json:"forbidden_paths"`
	MaxFiles              int      `json:"max_files"`
	RequiredGates         []string `json:"required_gates"`
	RollbackRequirements  []string `json:"rollback_requirements"`
	CIRequirements        []string `json:"ci_requirements"`
	PromotionRequirements []string `json:"promotion_requirements"`
}

type LowRiskCodeDenialAudit struct {
	SchemaVersion                    string   `json:"schema_version"`
	Status                           string   `json:"status"`
	MutationClass                    string   `json:"mutation_class"`
	CurrentProvenLiveClass           string   `json:"current_proven_live_class"`
	NextDeniedClass                  string   `json:"next_denied_class"`
	SafeToRequest                    bool     `json:"safe_to_request"`
	SafeToExecute                    bool     `json:"safe_to_execute"`
	MissingPolicyEvidence            []string `json:"missing_policy_evidence"`
	MissingRollbackEvidence          []string `json:"missing_rollback_evidence"`
	MissingSentinelPromoterEvidence  []string `json:"missing_sentinel_promoter_evidence"`
	SentinelState                    string   `json:"sentinel_state"`
	PromoterState                    string   `json:"promoter_state"`
	CIRequirements                   []string `json:"ci_requirements"`
	ExactNextAction                  string   `json:"exact_next_action"`
	DenialReason                     string   `json:"denial_reason"`
	SchedulesWork                    bool     `json:"schedules_work"`
	ExecutesWork                     bool     `json:"executes_work"`
	ApprovesWork                     bool     `json:"approves_work"`
	MutatesRepositories              bool     `json:"mutates_repositories"`
	CallsProviders                   bool     `json:"calls_providers"`
	ReleaseOrPublishAllowed          bool     `json:"release_or_publish_allowed"`
	FullyUnsupervisedMutationClaimed bool     `json:"fully_unsupervised_mutation_claimed"`
}

type FactoryTask struct {
	ContractVersion   string   `json:"contract_version"`
	ID                string   `json:"id"`
	Objective         string   `json:"objective"`
	TargetFactoryRepo string   `json:"target_factory_repo"`
	FactoryFolder     string   `json:"factory_folder"`
	MutationClass     string   `json:"mutation_class,omitempty"`
	Acceptance        []string `json:"acceptance_criteria"`
	NonGoals          []string `json:"non_goals"`
	WriteScope        []string `json:"write_scope"`
	RequiredGates     []string `json:"required_gates,omitempty"`
	RollbackScope     []string `json:"rollback_scope,omitempty"`
	Verification      []string `json:"verification_commands"`
	RequiredEvidence  []string `json:"required_evidence"`
	SafetyLimits      []string `json:"safety_limits"`
	AuthorityBoundary string   `json:"authority_boundary,omitempty"`
	DependencyRefs    []string `json:"dependency_refs"`
	ContextPackRefs   []string `json:"context_pack_refs"`
}

type FactoryMaterialization struct {
	ContractVersion string   `json:"contract_version"`
	TaskID          string   `json:"task_id"`
	Mode            string   `json:"mode"`
	OutputRoot      string   `json:"output_root"`
	Files           []string `json:"files"`
	ExecutesWork    bool     `json:"executes_work"`
	SchedulesWork   bool     `json:"schedules_work"`
	TaskDigest      string   `json:"task_digest"`
}

type ContextPack struct {
	ContractVersion      string      `json:"contract_version"`
	ID                   string      `json:"id"`
	TaskID               string      `json:"task_id"`
	BudgetBytes          int         `json:"budget_bytes"`
	SourceRefs           []SourceRef `json:"source_refs"`
	Summaries            []string    `json:"summaries"`
	Assumptions          []string    `json:"assumptions"`
	Exclusions           []string    `json:"exclusions"`
	MissingContextReason string      `json:"missing_context_reason,omitempty"`
	MissingProtocol      string      `json:"missing_context_protocol"`
}

type SourceRef struct {
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

type FoundryHandoff struct {
	ContractVersion string             `json:"contract_version"`
	ID              string             `json:"id"`
	TargetInstance  string             `json:"target_instance"`
	Status          string             `json:"status"`
	Tasks           []FoundryTaskEntry `json:"tasks"`
}

type FoundryImport struct {
	ContractVersion string                     `json:"contract_version"`
	ID              string                     `json:"id"`
	WorkgraphID     string                     `json:"workgraph_id"`
	TargetInstance  string                     `json:"target_instance"`
	Status          string                     `json:"status"`
	SourceArtifacts []SourceRef                `json:"source_artifacts"`
	Tasks           []FoundryImportTaskFixture `json:"tasks"`
	SchedulesWork   bool                       `json:"schedules_work"`
	ExecutesWork    bool                       `json:"executes_work"`
	ApprovesWork    bool                       `json:"approves_work"`
}

type FoundryContinuationHandoff struct {
	ContractVersion                 string   `json:"contract_version"`
	ID                              string   `json:"id"`
	TargetFolder                    string   `json:"target_folder"`
	Command                         string   `json:"command"`
	NextRecommendedAction           string   `json:"next_recommended_action"`
	Prompt                          string   `json:"prompt"`
	BlueprintPackPath               string   `json:"blueprint_pack_path"`
	AtlasImportPath                 string   `json:"atlas_import_path"`
	WorkgraphPath                   string   `json:"workgraph_path"`
	FoundryImportPath               string   `json:"foundry_import_path"`
	MissionContinuationEvidencePath string   `json:"mission_continuation_evidence_path,omitempty"`
	FirstSafeNode                   string   `json:"first_safe_node"`
	TotalNodeCount                  int      `json:"total_node_count"`
	CompletedNodeCount              int      `json:"completed_node_count"`
	BlockedNodeCount                int      `json:"blocked_node_count"`
	ReadyNodeCount                  int      `json:"ready_node_count"`
	ClassBoundary                   string   `json:"class_boundary"`
	StopConditions                  []string `json:"stop_conditions"`
	SafetyProhibitions              []string `json:"safety_prohibitions"`
	SchedulesWork                   bool     `json:"schedules_work"`
	ExecutesWork                    bool     `json:"executes_work"`
	ApprovesWork                    bool     `json:"approves_work"`
}

type FoundryImportTaskFixture struct {
	NodeID            string      `json:"node_id"`
	TaskID            string      `json:"task_id"`
	Path              string      `json:"path"`
	MutationClass     string      `json:"mutation_class"`
	WriteScope        []string    `json:"write_scope"`
	RollbackScope     []string    `json:"rollback_scope"`
	RequiredGates     []string    `json:"required_gates"`
	RequiredEvidence  []string    `json:"required_evidence"`
	AuthorityBoundary string      `json:"authority_boundary"`
	Task              FactoryTask `json:"task"`
	TaskHash          string      `json:"task_digest"`
}

type FoundryTaskEntry struct {
	ID                string   `json:"id"`
	Objective         string   `json:"objective"`
	TargetFactoryRepo string   `json:"target_factory_repo"`
	FactoryFolder     string   `json:"factory_folder"`
	Verification      []string `json:"verification_commands"`
	RequiredEvidence  []string `json:"required_evidence"`
}

type RunLink struct {
	ContractVersion string            `json:"contract_version"`
	TaskID          string            `json:"task_id"`
	Status          string            `json:"status"`
	Evidence        map[string]string `json:"evidence"`
	Digest          string            `json:"digest"`
}
