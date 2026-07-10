package atlas

const MissionLifecycleMetricsEvidenceContract = "ao.mission.lifecycle-metrics.v0.1"

type MissionLifecycleMetricsEvidence struct {
	Schema                            string `json:"schema"`
	MissionID                         string `json:"mission_id"`
	Status                            string `json:"status"`
	HandoffSteps                      int    `json:"handoff_steps"`
	EvidenceCompletedNodes            int    `json:"evidence_completed_nodes"`
	CompletedNodes                    int    `json:"completed_nodes"`
	TotalNodes                        int    `json:"total_nodes"`
	ReadyNodes                        int    `json:"ready_nodes"`
	BlockedNodes                      int    `json:"blocked_nodes"`
	FailedNodes                       int    `json:"failed_nodes"`
	CompletionBasis                   string `json:"completion_basis"`
	HandoffStepsCountAsCompletedNodes bool   `json:"handoff_steps_count_as_completed_nodes"`
	FinalResponseAllowed              bool   `json:"final_response_allowed"`
	ReturnGateStatus                  string `json:"return_gate_status"`
	ExactNextAction                   string `json:"exact_next_action"`
	SafeToExecute                     bool   `json:"safe_to_execute"`
	ExecutesWork                      bool   `json:"executes_work"`
	ApprovesWork                      bool   `json:"approves_work"`
	MutatesRepositories               bool   `json:"mutates_repositories"`
	RSIRemainsDenied                  bool   `json:"rsi_remains_denied"`
}

func ValidateMissionLifecycleMetricsEvidence(metrics MissionLifecycleMetricsEvidence) error {
	var errs []string
	requireContract(&errs, "mission_lifecycle_metrics", metrics.Schema, MissionLifecycleMetricsEvidenceContract)
	requireField(&errs, "mission_id", metrics.MissionID)
	if metrics.Status != "audited" {
		errs = append(errs, "status must be audited")
	}
	for name, value := range map[string]int{
		"handoff_steps":            metrics.HandoffSteps,
		"evidence_completed_nodes": metrics.EvidenceCompletedNodes,
		"completed_nodes":          metrics.CompletedNodes,
		"total_nodes":              metrics.TotalNodes,
		"ready_nodes":              metrics.ReadyNodes,
		"blocked_nodes":            metrics.BlockedNodes,
		"failed_nodes":             metrics.FailedNodes,
	} {
		if value < 0 {
			errs = append(errs, name+" must be non-negative")
		}
	}
	if metrics.CompletedNodes != metrics.EvidenceCompletedNodes {
		errs = append(errs, "completed_nodes must equal evidence_completed_nodes")
	}
	if metrics.HandoffStepsCountAsCompletedNodes {
		errs = append(errs, "handoff_steps_count_as_completed_nodes must be false")
	}
	if metrics.CompletionBasis != "downstream_evidence_not_handoff_steps" {
		errs = append(errs, "completion_basis must describe downstream evidence")
	}
	if metrics.TotalNodes > 0 && metrics.CompletedNodes > metrics.TotalNodes {
		errs = append(errs, "completed_nodes must not exceed total_nodes")
	}
	if metrics.FinalResponseAllowed && (metrics.ReadyNodes > 0 || metrics.ExactNextAction != "") {
		errs = append(errs, "final response cannot be allowed with ready nodes or an exact next action")
	}
	if metrics.SafeToExecute || metrics.ExecutesWork || metrics.ApprovesWork || metrics.MutatesRepositories {
		errs = append(errs, "lifecycle metrics must not claim execution or approval authority")
	}
	if !metrics.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}
