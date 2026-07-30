package atlas

type Workgraph struct {
	ContractVersion      string                       `json:"contract_version"`
	ID                   string                       `json:"id"`
	TargetInstance       string                       `json:"target_instance"`
	MissionID            string                       `json:"mission_id,omitempty"`
	ObjectiveID          string                       `json:"objective_id,omitempty"`
	ObjectiveDigest      string                       `json:"objective_digest,omitempty"`
	CorrelationID        string                       `json:"correlation_id,omitempty"`
	SoakID               string                       `json:"soak_id,omitempty"`
	PlanID               string                       `json:"plan_id,omitempty"`
	PolicyDigest         string                       `json:"policy_digest,omitempty"`
	ActivationID         string                       `json:"activation_id,omitempty"`
	EvidenceRootIdentity string                       `json:"evidence_root_identity,omitempty"`
	MissionSourceHead    string                       `json:"mission_source_head,omitempty"`
	AtlasSourceHead      string                       `json:"atlas_source_head,omitempty"`
	MissionBinarySHA256  string                       `json:"mission_binary_sha256,omitempty"`
	AtlasBinarySHA256    string                       `json:"atlas_binary_sha256,omitempty"`
	OperationalBinding   *OperationalWorkgraphBinding `json:"operational_binding,omitempty"`
	Nodes                []WorkgraphNode              `json:"nodes"`
}

type WorkgraphNode struct {
	ID                 string                           `json:"id"`
	Status             string                           `json:"status"`
	FactoryTask        FactoryTask                      `json:"factory_task"`
	Dependencies       []string                         `json:"dependencies"`
	Blockers           []string                         `json:"blockers"`
	StitchTask         bool                             `json:"stitch_task"`
	OperationalBinding *WorkgraphNodeOperationalBinding `json:"operational_binding,omitempty"`
	presentFields      map[string]bool
	nullFields         map[string]bool
}

type OperationalWorkgraphBinding struct {
	ExecutionProfileDigest string                      `json:"execution_profile_digest"`
	CommandCatalogDigest   string                      `json:"command_catalog_digest"`
	DurationHistoryDigest  string                      `json:"duration_history_digest"`
	PlannerInputDigest     string                      `json:"planner_input_digest"`
	PlannerReadbackDigest  string                      `json:"planner_readback_digest"`
	GraphBindingDigest     string                      `json:"graph_binding_digest"`
	RetryPolicy            WorkgraphRetryPolicy        `json:"retry_policy"`
	PlannerPartitions      []WorkgraphPlannerPartition `json:"planner_partitions"`
	SafetyBoundaries       WorkgraphSafetyBoundaries   `json:"safety_boundaries"`
}

type WorkgraphRetryPolicy struct {
	MaximumAttempts     int `json:"maximum_attempts"`
	MaximumTotalRetries int `json:"maximum_total_retries"`
	presentFields       map[string]bool
}

type WorkgraphPlannerPartition struct {
	PartitionID          string `json:"partition_id"`
	NodeID               string `json:"node_id"`
	TestID               string `json:"test_id"`
	Classification       string `json:"classification"`
	RequestedRepeatCount int    `json:"requested_repeat_count"`
	EffectiveRepeatCount int    `json:"effective_repeat_count"`
	ScaleDimension       string `json:"scale_dimension,omitempty"`
	EstimatedDurationMS  int64  `json:"estimated_duration_ms"`
	RetryAllowanceMS     int64  `json:"retry_allowance_ms"`
	PerAttemptTimeoutMS  int64  `json:"per_attempt_timeout_ms"`
	TotalNodeTimeoutMS   int64  `json:"total_node_timeout_ms"`
	NodeBudgetMS         int64  `json:"node_budget_ms"`
	presentFields        map[string]bool
}

type WorkgraphSafetyBoundaries struct {
	MutatesRepositories                bool `json:"mutates_repositories"`
	ReleasesOrPublishes                bool `json:"releases_or_publishes"`
	CallsProviders                     bool `json:"calls_providers"`
	ExpandsAuthority                   bool `json:"expands_authority"`
	AllowsUnplannedCommands            bool `json:"allows_unplanned_commands"`
	AllowsUnboundedRetries             bool `json:"allows_unbounded_retries"`
	ActivatesChildProcesses            bool `json:"activates_child_processes"`
	SchedulesWork                      bool `json:"schedules_work"`
	ExecutesWork                       bool `json:"executes_work"`
	ApprovesWork                       bool `json:"approves_work"`
	ActivationRequiresValidatedBinding bool `json:"activation_requires_validated_binding"`
	PreservesPhaseClockOnRetry         bool `json:"preserves_phase_clock_on_retry"`
	presentFields                      map[string]bool
}

type WorkgraphNodeOperationalBinding struct {
	PlannerPartitionID string                    `json:"planner_partition_id"`
	TestID             string                    `json:"test_id"`
	Classification     string                    `json:"classification"`
	SafetyBoundaries   WorkgraphSafetyBoundaries `json:"safety_boundaries"`
}

type OperationalWorkgraphBindingDocument struct {
	ContractVersion      string `json:"contract_version"`
	WorkgraphID          string `json:"workgraph_id"`
	TargetInstance       string `json:"target_instance"`
	MissionID            string `json:"mission_id"`
	ObjectiveID          string `json:"objective_id"`
	ObjectiveDigest      string `json:"objective_digest"`
	CorrelationID        string `json:"correlation_id"`
	SoakID               string `json:"soak_id"`
	PlanID               string `json:"plan_id"`
	PolicyDigest         string `json:"policy_digest"`
	ActivationID         string `json:"activation_id"`
	EvidenceRootIdentity string `json:"evidence_root_identity"`
	MissionSourceHead    string `json:"mission_source_head"`
	AtlasSourceHead      string `json:"atlas_source_head"`
	MissionBinarySHA256  string `json:"mission_binary_sha256"`
	AtlasBinarySHA256    string `json:"atlas_binary_sha256"`
	OperationalWorkgraphBinding
}

type OperationalWorkgraphBindingReadback struct {
	ContractVersion      string   `json:"contract_version"`
	WorkgraphID          string   `json:"workgraph_id,omitempty"`
	GraphBindingDigest   string   `json:"graph_binding_digest,omitempty"`
	ActivationAllowed    bool     `json:"activation_allowed"`
	ConflictCodes        []string `json:"conflict_codes"`
	ChildProcessLaunches int      `json:"child_process_launches"`
	ExecutesWork         bool     `json:"executes_work"`
	SafeToExecute        bool     `json:"safe_to_execute"`
	ExactNextAction      string   `json:"exact_next_action"`
}

type OperationalWorkgraphBindingDigestReadback struct {
	ContractVersion      string `json:"contract_version"`
	WorkgraphID          string `json:"workgraph_id"`
	GraphBindingDigest   string `json:"graph_binding_digest"`
	ChildProcessLaunches int    `json:"child_process_launches"`
	ExecutesWork         bool   `json:"executes_work"`
	SafeToExecute        bool   `json:"safe_to_execute"`
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
