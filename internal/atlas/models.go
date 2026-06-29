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
}

type InstanceDoctorReport struct {
	ContractVersion string            `json:"contract_version"`
	InstanceID      string            `json:"instance_id"`
	Status          string            `json:"status"`
	Checks          map[string]string `json:"checks"`
	SchedulesWork   bool              `json:"schedules_work"`
	ExecutesWork    bool              `json:"executes_work"`
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
	ContractVersion  string            `json:"contract_version"`
	IntakeID         string            `json:"intake_id"`
	WorkgraphID      string            `json:"workgraph_id"`
	TargetInstance   string            `json:"target_instance"`
	CompletionStatus string            `json:"completion_status"`
	NodeCounts       map[string]int    `json:"node_counts"`
	RunLinks         map[string]string `json:"run_links"`
	NextActions      []string          `json:"next_actions"`
	SchedulesWork    bool              `json:"schedules_work"`
	ExecutesWork     bool              `json:"executes_work"`
}

type BlueprintRequest struct {
	ContractVersion string   `json:"contract_version"`
	IntakeID        string   `json:"intake_id"`
	Status          string   `json:"status"`
	Missing         []string `json:"missing"`
	Reason          string   `json:"reason"`
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

type FactoryTask struct {
	ContractVersion   string   `json:"contract_version"`
	ID                string   `json:"id"`
	Objective         string   `json:"objective"`
	TargetFactoryRepo string   `json:"target_factory_repo"`
	FactoryFolder     string   `json:"factory_folder"`
	Acceptance        []string `json:"acceptance_criteria"`
	NonGoals          []string `json:"non_goals"`
	WriteScope        []string `json:"write_scope"`
	Verification      []string `json:"verification_commands"`
	RequiredEvidence  []string `json:"required_evidence"`
	SafetyLimits      []string `json:"safety_limits"`
	DependencyRefs    []string `json:"dependency_refs"`
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
	ContractVersion string      `json:"contract_version"`
	ID              string      `json:"id"`
	TaskID          string      `json:"task_id"`
	BudgetBytes     int         `json:"budget_bytes"`
	SourceRefs      []SourceRef `json:"source_refs"`
	Summaries       []string    `json:"summaries"`
	Assumptions     []string    `json:"assumptions"`
	Exclusions      []string    `json:"exclusions"`
	MissingProtocol string      `json:"missing_context_protocol"`
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
	Tasks           []FoundryImportTaskFixture `json:"tasks"`
	SchedulesWork   bool                       `json:"schedules_work"`
	ExecutesWork    bool                       `json:"executes_work"`
	ApprovesWork    bool                       `json:"approves_work"`
}

type FoundryImportTaskFixture struct {
	NodeID   string      `json:"node_id"`
	TaskID   string      `json:"task_id"`
	Path     string      `json:"path"`
	Task     FactoryTask `json:"task"`
	TaskHash string      `json:"task_digest"`
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
