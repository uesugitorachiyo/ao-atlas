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
	ContractVersion       string                 `json:"contract_version"`
	IntakeID              string                 `json:"intake_id"`
	WorkgraphID           string                 `json:"workgraph_id"`
	TargetInstance        string                 `json:"target_instance"`
	CompletionStatus      string                 `json:"completion_status"`
	NodeCounts            map[string]int         `json:"node_counts"`
	RunLinks              map[string]string      `json:"run_links"`
	MissingContextPacks   []string               `json:"missing_context_packs"`
	MissingHandoffs       []string               `json:"missing_handoffs"`
	NextRecommendedAction string                 `json:"next_recommended_action"`
	NextActions           []string               `json:"next_actions"`
	AuthorityLadder       *AuthorityLadderStatus `json:"authority_ladder,omitempty"`
	SchedulesWork         bool                   `json:"schedules_work"`
	ExecutesWork          bool                   `json:"executes_work"`
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
	PrimaryMissionProvenance string            `json:"primary_mission_provenance"`
	ProvenanceDiagnostics    string            `json:"provenance_diagnostics"`
	SourceArtifacts          map[string]string `json:"source_artifacts"`
	NextAction               string            `json:"next_action"`
	SafeToExecute            bool              `json:"safe_to_execute"`
	SchedulesWork            bool              `json:"schedules_work"`
	ExecutesWork             bool              `json:"executes_work"`
	ApprovesWork             bool              `json:"approves_work"`
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
