package atlas

import (
	"fmt"
	"strings"
)

type AtlasP0BWindowsCIWaitTelemetry struct {
	Schema                      string                                 `json:"schema"`
	NodeID                      string                                 `json:"node_id"`
	Status                      string                                 `json:"status"`
	Source                      string                                 `json:"source"`
	LongRunningOS               string                                 `json:"long_running_os"`
	WaitThresholdSeconds        int                                    `json:"wait_threshold_seconds"`
	SourcePRs                   []int                                  `json:"source_prs"`
	WindowsCheckSampleCount     int                                    `json:"windows_check_sample_count"`
	CommandShardCount           int                                    `json:"command_shard_count"`
	CovenantShardCount          int                                    `json:"covenant_shard_count"`
	PendingStateObserved        bool                                   `json:"pending_state_observed"`
	CompletedPassStateObserved  bool                                   `json:"completed_pass_state_observed"`
	FailedStateObserved         bool                                   `json:"failed_state_observed"`
	MaxObservedDurationSeconds  int                                    `json:"max_observed_duration_seconds"`
	CompletedNodesBefore        int                                    `json:"completed_nodes_before"`
	ReadyNodesBefore            int                                    `json:"ready_nodes_before"`
	FinalResponseAllowed        bool                                   `json:"final_response_allowed"`
	ExpectedNextNode            string                                 `json:"expected_next_node"`
	PromotionRequested          bool                                   `json:"promotion_requested"`
	PromotionGranted            bool                                   `json:"promotion_granted"`
	ClaimsAuthorityAdvance      bool                                   `json:"claims_authority_advance"`
	RSIRemainsDenied            bool                                   `json:"rsi_remains_denied"`
	SchedulesWork               bool                                   `json:"schedules_work"`
	ExecutesWork                bool                                   `json:"executes_work"`
	ApprovesWork                bool                                   `json:"approves_work"`
	WindowsCheckDurationSamples []AtlasP0BWindowsCIWaitTelemetrySample `json:"windows_check_duration_samples"`
}

type AtlasP0BWindowsCIWaitTelemetrySample struct {
	Shard           string `json:"shard"`
	PRNumber        int    `json:"pr_number"`
	CheckName       string `json:"check_name"`
	FinalStatus     string `json:"final_status"`
	FinalConclusion string `json:"final_conclusion"`
	DurationSeconds int    `json:"duration_seconds"`
	WaitState       string `json:"wait_state"`
	OperatorAction  string `json:"operator_action"`
}

func ValidateAtlasP0BWindowsCIWaitTelemetry(telemetry AtlasP0BWindowsCIWaitTelemetry) error {
	var errs []string
	requireContract(&errs, "p0b_windows_ci_wait_telemetry", telemetry.Schema, AtlasP0BWindowsCIWaitTelemetryContract)
	if telemetry.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	requireField(&errs, "node_id", telemetry.NodeID)
	checkPublicPath(&errs, "node_id", telemetry.NodeID, true)
	requireField(&errs, "source", telemetry.Source)
	checkPublicPath(&errs, "source", telemetry.Source, true)
	if telemetry.LongRunningOS != "windows-latest" {
		errs = append(errs, "long_running_os must be windows-latest")
	}
	if telemetry.WaitThresholdSeconds <= 0 {
		errs = append(errs, "wait_threshold_seconds must be greater than zero")
	}
	if len(telemetry.SourcePRs) == 0 {
		errs = append(errs, "source_prs must not be empty")
	}
	if telemetry.WindowsCheckSampleCount != len(telemetry.WindowsCheckDurationSamples) {
		errs = append(errs, "windows_check_sample_count must match samples length")
	}
	if telemetry.CompletedNodesBefore < 0 || telemetry.ReadyNodesBefore < 0 {
		errs = append(errs, "node counts must be non-negative")
	}
	if telemetry.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while P0-B work remains")
	}
	requireField(&errs, "expected_next_node", telemetry.ExpectedNextNode)
	checkPublicPath(&errs, "expected_next_node", telemetry.ExpectedNextNode, true)

	commandCount := 0
	covenantCount := 0
	pendingObserved := false
	successObserved := false
	failedObserved := false
	maxDuration := 0
	for i, sample := range telemetry.WindowsCheckDurationSamples {
		prefix := fmt.Sprintf("windows_check_duration_samples[%d]", i)
		validateP0BWindowsCIWaitTelemetrySample(&errs, prefix, sample, telemetry.WaitThresholdSeconds)
		switch sample.Shard {
		case "ao-command":
			commandCount++
		case "ao-covenant":
			covenantCount++
		default:
			errs = append(errs, prefix+".shard must be ao-command or ao-covenant")
		}
		status := strings.ToUpper(strings.TrimSpace(sample.FinalStatus))
		conclusion := strings.ToUpper(strings.TrimSpace(sample.FinalConclusion))
		pendingObserved = pendingObserved || oneOf(status, "QUEUED", "PENDING", "IN_PROGRESS")
		successObserved = successObserved || (status == "COMPLETED" && conclusion == "SUCCESS")
		failedObserved = failedObserved || (status == "COMPLETED" && oneOf(conclusion, "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE"))
		if sample.DurationSeconds > maxDuration {
			maxDuration = sample.DurationSeconds
		}
	}
	if telemetry.CommandShardCount != commandCount {
		errs = append(errs, "command_shard_count must match samples")
	}
	if telemetry.CovenantShardCount != covenantCount {
		errs = append(errs, "covenant_shard_count must match samples")
	}
	if telemetry.PendingStateObserved != pendingObserved {
		errs = append(errs, "pending_state_observed must match samples")
	}
	if telemetry.CompletedPassStateObserved != successObserved {
		errs = append(errs, "completed_pass_state_observed must match samples")
	}
	if telemetry.FailedStateObserved != failedObserved {
		errs = append(errs, "failed_state_observed must match samples")
	}
	if telemetry.MaxObservedDurationSeconds != maxDuration {
		errs = append(errs, "max_observed_duration_seconds must match samples")
	}
	if telemetry.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if telemetry.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, telemetry.SchedulesWork, telemetry.ExecutesWork, telemetry.ApprovesWork, telemetry.ClaimsAuthorityAdvance, telemetry.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateP0BWindowsCIWaitTelemetrySample(errs *[]string, prefix string, sample AtlasP0BWindowsCIWaitTelemetrySample, thresholdSeconds int) {
	requireField(errs, prefix+".shard", sample.Shard)
	checkPublicPath(errs, prefix+".shard", sample.Shard, true)
	if sample.PRNumber <= 0 {
		*errs = append(*errs, prefix+".pr_number must be greater than zero")
	}
	requireField(errs, prefix+".check_name", sample.CheckName)
	checkPublicPath(errs, prefix+".check_name", sample.CheckName, false)
	if !strings.Contains(strings.ToLower(sample.CheckName), "windows-latest") {
		*errs = append(*errs, prefix+".check_name must reference windows-latest")
	}
	state, err := ClassifyAtlasWindowsCIWaitState(AtlasWindowsCIWaitStateInput{
		CheckName:        sample.CheckName,
		GitHubStatus:     sample.FinalStatus,
		GitHubConclusion: sample.FinalConclusion,
		DurationSeconds:  sample.DurationSeconds,
		ThresholdSeconds: thresholdSeconds,
	})
	if err != nil {
		*errs = append(*errs, prefix+": "+err.Error())
		return
	}
	if state.WaitState != sample.WaitState {
		*errs = append(*errs, prefix+".wait_state must match classifier")
	}
	if state.OperatorAction != sample.OperatorAction {
		*errs = append(*errs, prefix+".operator_action must match classifier")
	}
	if state.FinalResponseAllowed {
		*errs = append(*errs, prefix+".classifier final_response_allowed must be false")
	}
	if state.ClaimsAuthorityAdvance {
		*errs = append(*errs, prefix+".classifier claims_authority_advance must be false")
	}
	if !state.RSIRemainsDenied {
		*errs = append(*errs, prefix+".classifier rsi_remains_denied must be true")
	}
}
