package atlas

type Instance struct {
	ContractVersion string            `json:"contract_version"`
	ID              string            `json:"id"`
	StateRoot       string            `json:"state_root"`
	ToolchainRoot   string            `json:"toolchain_root"`
	Roots           map[string]string `json:"roots"`
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
