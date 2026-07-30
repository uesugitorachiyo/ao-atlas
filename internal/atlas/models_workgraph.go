package atlas

type Workgraph struct {
	ContractVersion      string          `json:"contract_version"`
	ID                   string          `json:"id"`
	TargetInstance       string          `json:"target_instance"`
	MissionID            string          `json:"mission_id,omitempty"`
	ObjectiveID          string          `json:"objective_id,omitempty"`
	ObjectiveDigest      string          `json:"objective_digest,omitempty"`
	CorrelationID        string          `json:"correlation_id,omitempty"`
	SoakID               string          `json:"soak_id,omitempty"`
	PlanID               string          `json:"plan_id,omitempty"`
	PolicyDigest         string          `json:"policy_digest,omitempty"`
	ActivationID         string          `json:"activation_id,omitempty"`
	EvidenceRootIdentity string          `json:"evidence_root_identity,omitempty"`
	MissionSourceHead    string          `json:"mission_source_head,omitempty"`
	AtlasSourceHead      string          `json:"atlas_source_head,omitempty"`
	MissionBinarySHA256  string          `json:"mission_binary_sha256,omitempty"`
	AtlasBinarySHA256    string          `json:"atlas_binary_sha256,omitempty"`
	Nodes                []WorkgraphNode `json:"nodes"`
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
